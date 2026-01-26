import argparse
import ctypes
import json
import locale
import os
import subprocess
import sys
import tempfile
import time
from dataclasses import dataclass
from typing import Dict, Iterable, List, Optional, Sequence, TextIO, Tuple

import psutil


def _default_encoding() -> str:
    try:
        enc = locale.getpreferredencoding(False) or "utf-8"
        return enc
    except Exception:
        return "utf-8"


def _is_admin() -> bool:
    try:
        return bool(ctypes.windll.shell32.IsUserAnAdmin())
    except Exception:
        return False


class _ReportWriter:
    def __init__(self, file: TextIO, echo_to_stdout: bool) -> None:
        self._file = file
        self._echo = echo_to_stdout

    def write_line(self, text: str = "") -> None:
        self._file.write(text + "\n")
        self._file.flush()
        if self._echo:
            sys.stdout.write(text + "\n")
            sys.stdout.flush()


def _run_cmd(
    cmd: Sequence[str],
    timeout_s: int = 60,
    encoding: Optional[str] = None,
) -> Tuple[int, str, str]:
    enc = encoding or _default_encoding()
    try:
        p = subprocess.run(
            list(cmd),
            capture_output=True,
            text=True,
            encoding=enc,
            errors="replace",
            timeout=timeout_s,
            creationflags=getattr(subprocess, "CREATE_NO_WINDOW", 0),
        )
        return p.returncode, p.stdout.strip(), p.stderr.strip()
    except subprocess.TimeoutExpired:
        return 124, "", f"超时（{timeout_s} 秒）"
    except FileNotFoundError:
        return 127, "", "命令未找到"
    except Exception as e:
        return 1, "", str(e)


def _format_bytes(value: Optional[int]) -> str:
    if value is None:
        return "-"
    if value < 0:
        return f"-{_format_bytes(-value)}"
    units = ["B", "KB", "MB", "GB", "TB", "PB"]
    v = float(value)
    idx = 0
    while v >= 1024.0 and idx < len(units) - 1:
        v /= 1024.0
        idx += 1
    if idx == 0:
        return f"{int(v)} {units[idx]}"
    return f"{v:.2f} {units[idx]}"


def _print_kv_table(writer: _ReportWriter, title: str, rows: List[Tuple[str, str]]) -> None:
    if title:
        writer.write_line("=" * 70)
        writer.write_line(title)
        writer.write_line("=" * 70)
    key_width = max((len(k) for k, _ in rows), default=10)
    for k, v in rows:
        writer.write_line(f"{k:<{key_width}}  {v}")
    writer.write_line()


def _print_table(writer: _ReportWriter, title: str, headers: List[str], rows: List[List[str]]) -> None:
    if title:
        writer.write_line("=" * 70)
        writer.write_line(title)
        writer.write_line("=" * 70)
    all_rows = [headers] + rows
    widths = [max(len(r[i]) for r in all_rows) for i in range(len(headers))]
    fmt = "  ".join([f"{{:<{w}}}" for w in widths])
    writer.write_line(fmt.format(*headers))
    writer.write_line(fmt.format(*["-" * w for w in widths]))
    for r in rows:
        writer.write_line(fmt.format(*r))
    writer.write_line()


def _powershell_json(cmd: str, timeout_s: int = 30) -> Tuple[bool, Dict[str, object], str]:
    rc, out, err = _run_cmd(
        [
            "powershell",
            "-NoProfile",
            "-ExecutionPolicy",
            "Bypass",
            "-Command",
            cmd,
        ],
        timeout_s=timeout_s,
        encoding="utf-8",
    )
    if rc != 0:
        return False, {}, err or out
    try:
        data = json.loads(out)
        if isinstance(data, dict):
            return True, data, ""
        return True, {"value": data}, ""
    except Exception as e:
        return False, {}, f"解析 JSON 失败: {e}"


def _collect_perf_os_memory() -> Tuple[bool, Dict[str, int], str]:
    try:
        import pythoncom  # type: ignore
        import wmi  # type: ignore

        pythoncom.CoInitialize()
        c = wmi.WMI(namespace="root\\cimv2")
        items = c.Win32_PerfFormattedData_PerfOS_Memory()
        if not items:
            return False, {}, "WMI 返回为空"
        it = items[0]
        out = {}
        for k in ["PoolNonpagedBytes", "PoolPagedBytes", "CacheBytes", "CommitLimit", "CommittedBytes", "AvailableBytes"]:
            v = getattr(it, k, None)
            if v is None:
                continue
            try:
                out[k] = int(v)
            except Exception:
                continue
        if not out:
            return False, {}, "WMI 未返回可用字段"
        return True, out, ""
    except Exception:
        pass

    wmi_cmd = (
        "Get-CimInstance Win32_PerfFormattedData_PerfOS_Memory | "
        "Select-Object -First 1 "
        "PoolNonpagedBytes,PoolPagedBytes,CacheBytes,CommitLimit,CommittedBytes,AvailableBytes | "
        "ConvertTo-Json -Compress"
    )
    ok, data, err = _powershell_json(wmi_cmd, timeout_s=15)
    if not ok:
        return False, {}, err
    out2: Dict[str, int] = {}
    for k in ["PoolNonpagedBytes", "PoolPagedBytes", "CacheBytes", "CommitLimit", "CommittedBytes", "AvailableBytes"]:
        if k not in data or data.get(k) is None:
            continue
        try:
            out2[k] = int(data.get(k))
        except Exception:
            continue
    if not out2:
        return False, {}, "CIM 未返回可用字段"
    return True, out2, ""


@dataclass(frozen=True)
class MemorySnapshot:
    total_physical_bytes: int
    available_physical_bytes: int
    used_physical_bytes: int
    process_rss_sum_bytes: int
    black_hole_bytes: int
    pool_nonpaged_bytes: Optional[int]
    pool_paged_bytes: Optional[int]
    system_cache_bytes: Optional[int]
    commit_limit_bytes: Optional[int]
    committed_bytes: Optional[int]


def collect_memory_summary() -> MemorySnapshot:
    vm = psutil.virtual_memory()
    process_rss_sum = 0
    for p in psutil.process_iter(["pid"]):
        try:
            process_rss_sum += int(p.memory_info().rss)
        except Exception:
            continue
    used = int(vm.total - vm.available)
    black_hole = used - process_rss_sum
    if black_hole < 0:
        black_hole = 0

    ok, data, _err = _collect_perf_os_memory()
    pool_nonpaged = None
    pool_paged = None
    cache = None
    commit_limit = None
    committed = None
    if ok:
        try:
            pool_nonpaged = int(data.get("PoolNonpagedBytes")) if data.get("PoolNonpagedBytes") is not None else None
            pool_paged = int(data.get("PoolPagedBytes")) if data.get("PoolPagedBytes") is not None else None
            cache = int(data.get("CacheBytes")) if data.get("CacheBytes") is not None else None
            commit_limit = int(data.get("CommitLimit")) if data.get("CommitLimit") is not None else None
            committed = int(data.get("CommittedBytes")) if data.get("CommittedBytes") is not None else None
        except Exception:
            pool_nonpaged = None
            pool_paged = None
            cache = None
            commit_limit = None
            committed = None

    return MemorySnapshot(
        total_physical_bytes=int(vm.total),
        available_physical_bytes=int(vm.available),
        used_physical_bytes=used,
        process_rss_sum_bytes=process_rss_sum,
        black_hole_bytes=black_hole,
        pool_nonpaged_bytes=pool_nonpaged,
        pool_paged_bytes=pool_paged,
        system_cache_bytes=cache,
        commit_limit_bytes=commit_limit,
        committed_bytes=committed,
    )


def print_memory_summary(writer: _ReportWriter, snapshot: MemorySnapshot) -> None:
    rows = [
        ("物理内存总量", _format_bytes(snapshot.total_physical_bytes)),
        ("物理内存已用", _format_bytes(snapshot.used_physical_bytes)),
        ("物理内存可用", _format_bytes(snapshot.available_physical_bytes)),
        ("进程 RSS 总和", _format_bytes(snapshot.process_rss_sum_bytes)),
        ("黑洞大小（已用 - RSS）", _format_bytes(snapshot.black_hole_bytes)),
        ("非分页池（Non-paged Pool）", _format_bytes(snapshot.pool_nonpaged_bytes)),
        ("分页池（Paged Pool）", _format_bytes(snapshot.pool_paged_bytes)),
        ("系统缓存（System Cache）", _format_bytes(snapshot.system_cache_bytes)),
        ("已提交（Committed）", _format_bytes(snapshot.committed_bytes)),
        ("提交上限（Commit Limit）", _format_bytes(snapshot.commit_limit_bytes)),
    ]
    _print_kv_table(writer, "1) 系统内存全局概览", rows)

    suspicious = []
    if snapshot.pool_nonpaged_bytes is not None and snapshot.pool_nonpaged_bytes >= 1 * 1024**3:
        suspicious.append("非分页池 >= 1GB（极大概率为内核驱动泄漏）")
    if snapshot.black_hole_bytes >= 2 * 1024**3:
        suspicious.append("黑洞 >= 2GB（倾向内核/驱动/系统缓存压力异常）")
    if suspicious:
        _print_kv_table(writer, "可疑信号", [(str(i + 1) + ".", s) for i, s in enumerate(suspicious)])


def _which(executable: str) -> Optional[str]:
    if os.path.isabs(executable) and os.path.exists(executable):
        return executable
    paths = os.environ.get("PATH", "").split(os.pathsep)
    for p in paths:
        full = os.path.join(p, executable)
        if os.path.exists(full):
            return full
        if not executable.lower().endswith(".exe"):
            full_exe = full + ".exe"
            if os.path.exists(full_exe):
                return full_exe
    return None


def _find_poolmon() -> Optional[str]:
    candidates = [
        "poolmon.exe",
        os.path.join(os.environ.get("ProgramFiles(x86)", r"C:\Program Files (x86)"), "Windows Kits", "10", "Tools", "x64", "poolmon.exe"),
        os.path.join(os.environ.get("ProgramFiles(x86)", r"C:\Program Files (x86)"), "Windows Kits", "10", "Tools", "x86", "poolmon.exe"),
    ]
    for c in candidates:
        path = _which(c) if c.endswith(".exe") and not os.path.isabs(c) else (c if os.path.exists(c) else None)
        if path and os.path.exists(path):
            return path
    return None


def _parse_poolmon_snapshot(snapshot_text: str) -> List[Dict[str, object]]:
    lines = [ln.rstrip("\r\n") for ln in snapshot_text.splitlines()]
    parsed: List[Dict[str, object]] = []
    for ln in lines:
        if not ln:
            continue
        if ln.lstrip().startswith(("Memory:", "Commit:", "System", "Tag")):
            continue
        parts = ln.split()
        if len(parts) < 6:
            continue
        tag = parts[0]
        pool_type = parts[1]
        bytes_str = parts[5].replace(",", "")
        try:
            bytes_used = int(bytes_str)
        except Exception:
            continue
        parsed.append({"tag": tag, "type": pool_type, "bytes": bytes_used, "raw": ln})
    parsed.sort(key=lambda x: int(x.get("bytes", 0)), reverse=True)
    return parsed


def _find_drivers_by_tag(tag: str, limit: int = 8) -> List[str]:
    system_root = os.environ.get("SystemRoot", r"C:\Windows")
    drivers_glob = os.path.join(system_root, "System32", "drivers", "*.sys")
    rc, out, err = _run_cmd(
        ["cmd", "/c", "findstr", "/m", "/s", "/l", tag, drivers_glob],
        timeout_s=45,
    )
    if rc != 0 and not out:
        _ = err
        return []
    paths = [p.strip() for p in out.splitlines() if p.strip()]
    return paths[:limit]


def pooltag_analysis(writer: _ReportWriter, top_n: int = 5) -> None:
    poolmon_path = _find_poolmon()
    if not poolmon_path:
        _print_kv_table(
            writer,
            "2) Non-paged PoolTag Analysis",
            [
                ("状态", "未找到 poolmon.exe（需要安装 Windows WDK / Support Tools）"),
                ("建议", "若非分页池很高，安装 WDK 后重新运行本脚本以定位 PoolTag"),
            ],
        )
        return

    with tempfile.TemporaryDirectory(prefix="poolmon_") as td:
        snap_path = os.path.join(td, "poolsnap.log")
        rc, out, err = _run_cmd(
            [poolmon_path, "/p", "/b", "/n", snap_path],
            timeout_s=20,
        )
        if rc != 0:
            _print_kv_table(
                writer,
                "2) Non-paged PoolTag Analysis",
                [("状态", f"poolmon 执行失败（rc={rc}）"), ("详情", err or out)],
            )
            return

        try:
            with open(snap_path, "r", encoding="utf-8", errors="replace") as f:
                snap_text = f.read()
        except Exception:
            with open(snap_path, "r", encoding=_default_encoding(), errors="replace") as f:
                snap_text = f.read()

    tags = _parse_poolmon_snapshot(snap_text)
    top = tags[:top_n]
    if not top:
        _print_kv_table(
            writer,
            "2) Non-paged PoolTag Analysis",
            [("状态", "poolmon 快照解析为空（输出格式可能变化）")],
        )
        return

    keyword_hits = {"SANG": "深信服相关", "TENC": "腾讯相关", "NDIS": "网络栈相关"}
    rows = []
    for i, item in enumerate(top, start=1):
        tag = str(item["tag"])
        size = _format_bytes(int(item["bytes"]))
        hint = ""
        for k, v in keyword_hits.items():
            if tag.upper().startswith(k):
                hint = v
                break
        rows.append([str(i), tag, item["type"], size, hint or "-"])
    _print_table(writer, "2) 非分页池 PoolTag 分析（占用 Top）", ["序号", "Tag", "类型", "占用", "提示"], rows)

    mapped_rows = []
    for item in top:
        tag = str(item["tag"])
        drivers = _find_drivers_by_tag(tag, limit=6)
        mapped_rows.append([tag, ", ".join([os.path.basename(d) for d in drivers]) if drivers else "-"])
    _print_table(writer, "PoolTag -> 可能对应的驱动文件（findstr 扫描）", ["Tag", "命中的 .sys（文件名）"], mapped_rows)


def zombie_process_hunter(writer: _ReportWriter, top_n: int = 15) -> None:
    rows = []
    candidates: List[Tuple[int, int, int, str]] = []
    for p in psutil.process_iter(["pid", "name"]):
        try:
            handles = p.num_handles()
            rss = int(p.memory_info().rss)
            name = p.info.get("name") or ""
            candidates.append((handles, rss, int(p.pid), name))
        except Exception:
            continue
    candidates.sort(key=lambda x: (x[0], x[1]), reverse=True)
    for handles, rss, pid, name in candidates[:top_n]:
        suspicious = "是" if handles >= 10_000 else "-"
        rows.append([str(pid), name, f"{handles:,}", _format_bytes(rss), suspicious])
    _print_table(writer, "3) 僵尸/隐藏进程探测（句柄压力）", ["PID", "进程名", "句柄数", "RSS", "可疑"], rows)


def network_buffer_check(writer: _ReportWriter) -> None:
    suspicious_keywords = ["Sangfor", "EasyConnect", "Tencent", "NDIS"]
    rc, out, err = _run_cmd(["netsh", "wfp", "show", "filters"], timeout_s=45)
    if rc != 0:
        _print_kv_table(
            writer,
            "4) 网络缓冲区 / WFP 检查",
            [("状态", f"netsh wfp show filters 执行失败（rc={rc}）"), ("详情", err or out)],
        )
        return

    lines = out.splitlines()
    hit_lines = []
    for ln in lines:
        if any(k.lower() in ln.lower() for k in suspicious_keywords):
            hit_lines.append(ln.strip())

    providers = {}
    for ln in lines:
        if ":" not in ln:
            continue
        k, v = ln.split(":", 1)
        k = k.strip().lower()
        v = v.strip()
        if k in {"provider key", "provider name", "filter name", "layer name", "callout name"}:
            providers[k] = providers.get(k, 0) + 1

    summary_rows = [
        ("输出行数", f"{len(lines):,}"),
        ("可疑关键字命中行数", f"{len(hit_lines):,}"),
        ("Provider Name 条目数", f"{providers.get('provider name', 0):,}"),
        ("Filter Name 条目数", f"{providers.get('filter name', 0):,}"),
        ("Layer Name 条目数", f"{providers.get('layer name', 0):,}"),
        ("Callout Name 条目数", f"{providers.get('callout name', 0):,}"),
    ]
    _print_kv_table(writer, "4) 网络缓冲区 / WFP 检查（摘要）", summary_rows)

    sample = hit_lines[:40]
    if sample:
        sample_rows = [[str(i + 1), ln] for i, ln in enumerate(sample)]
        _print_table(writer, "WFP 可疑关键字命中样例（前 40 行）", ["序号", "内容"], sample_rows)
    else:
        _print_kv_table(writer, "WFP 可疑关键字命中样例", [("状态", "未发现明显 Sangfor/Tencent/NDIS 关键字命中")])


def _default_report_path() -> str:
    base_dir = os.path.dirname(os.path.abspath(__file__))
    ts = time.strftime("%Y%m%d_%H%M%S")
    return os.path.join(base_dir, f"windows_memory_blackhole_report_{ts}.txt")


def main(argv: Optional[Sequence[str]] = None) -> int:
    parser = argparse.ArgumentParser(prog="windows_memory_blackhole_probe.py")
    parser.add_argument("--out", default=None, help="报告输出路径（默认脚本目录下的 .txt）")
    parser.add_argument("--quiet", action="store_true", help="不在控制台输出，仅写入报告文件")
    parser.add_argument("--no-pooltag", action="store_true", help="跳过 PoolTag 分析（poolmon）")
    parser.add_argument("--no-wfp", action="store_true", help="跳过 WFP filters 检查")
    parser.add_argument("--top-tags", type=int, default=5)
    parser.add_argument("--top-procs", type=int, default=15)
    args = parser.parse_args(list(argv) if argv is not None else None)

    report_path = os.path.abspath(args.out) if args.out else _default_report_path()
    os.makedirs(os.path.dirname(report_path), exist_ok=True)
    with open(report_path, "w", encoding="utf-8", errors="replace") as f:
        writer = _ReportWriter(f, echo_to_stdout=not args.quiet)

        writer.write_line("Windows 内存黑洞（内核/驱动泄漏）深度排查报告")
        writer.write_line(f"管理员权限：{'是' if _is_admin() else '否'}")
        writer.write_line(f"时间：{time.strftime('%Y-%m-%d %H:%M:%S')}")
        writer.write_line(f"报告文件：{report_path}")
        writer.write_line()

        snapshot = collect_memory_summary()
        print_memory_summary(writer, snapshot)
        if not args.no_pooltag:
            pooltag_analysis(writer, top_n=max(1, args.top_tags))
        zombie_process_hunter(writer, top_n=max(1, args.top_procs))
        if not args.no_wfp:
            network_buffer_check(writer)

    if args.quiet:
        sys.stdout.write(f"已生成报告：{report_path}\n")
        sys.stdout.flush()
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

