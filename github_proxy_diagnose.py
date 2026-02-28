import os
import socket
import ssl
import subprocess
import sys
import urllib.request
from typing import Optional, Tuple


def run_cmd(cmd: list[str]) -> Tuple[int, str, str]:
    try:
        proc = subprocess.Popen(
            cmd,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
            shell=False,
        )
        out, err = proc.communicate(timeout=20)
        return proc.returncode, out.strip(), err.strip()
    except Exception as e:
        return 1, "", str(e)


def print_section(title: str) -> None:
    print()
    print("=" * 60)
    print(title)
    print("=" * 60)


def check_env_proxies() -> None:
    print_section("1. 环境变量中的代理设置")
    keys = ["HTTP_PROXY", "http_proxy", "HTTPS_PROXY", "https_proxy", "ALL_PROXY", "all_proxy", "NO_PROXY", "no_proxy"]
    for key in keys:
        value = os.environ.get(key)
        if value:
            print(f"{key} = {value}")
    if not any(os.environ.get(k) for k in ["HTTP_PROXY", "http_proxy", "HTTPS_PROXY", "https_proxy", "ALL_PROXY", "all_proxy"]):
        print("未检测到环境变量层面的代理设置")


def check_git_proxy() -> None:
    print_section("2. git 配置中的代理设置")
    for scope in ["system", "global", "local"]:
        code, out, err = run_cmd(["git", "config", f"--{scope}", "--get-regexp", "proxy"])
        label = f"git {scope} 配置"
        if code == 0 and out:
            print(f"{label}:")
            print(out)
        else:
            print(f"{label}: 未检测到 proxy 相关配置")
            if err:
                print(f"  备注: {err}")


def check_dns(host: str) -> None:
    print_section("3. DNS 解析检测")
    try:
        addr_infos = socket.getaddrinfo(host, 443)
        addrs = sorted({ai[4][0] for ai in addr_infos})
        print(f"{host} 解析结果:")
        for a in addrs:
            print(f"  {a}")
    except Exception as e:
        print(f"{host} 解析失败: {e}")


def check_tcp_connect(host: str, port: int, use_proxy: bool) -> None:
    label = "通过系统代理" if use_proxy else "直连"
    print_section(f"4. TCP 连接检测 ({label})")
    target = f"{host}:{port}"
    try:
        sock = socket.create_connection((host, port), timeout=10)
        try:
            context = ssl.create_default_context()
            context.check_hostname = False
            context.verify_mode = ssl.CERT_NONE
            context.wrap_socket(sock, server_hostname=host)
            print(f"到 {target} 的 TCP+TLS 连接成功")
        finally:
            sock.close()
    except Exception as e:
        print(f"到 {target} 的 TCP 连接失败: {e}")


def check_http_request(url: str, use_env_proxy: bool) -> None:
    mode = "使用环境变量代理" if use_env_proxy else "忽略环境变量代理"
    print_section(f"5. HTTP 请求检测 ({mode})")
    if not use_env_proxy:
        opener = urllib.request.build_opener(urllib.request.ProxyHandler({}))
    else:
        opener = urllib.request.build_opener()
    req = urllib.request.Request(url, method="GET")
    try:
        with opener.open(req, timeout=15) as resp:
            print(f"请求 {url} 成功, 状态码: {resp.status}")
    except Exception as e:
        print(f"请求 {url} 失败: {e}")


def check_git_ls_remote(use_env_proxy: bool) -> None:
    mode = "继承当前环境变量中的代理设置" if use_env_proxy else "在清空代理环境变量下执行"
    print_section(f"6. git ls-remote 检测 ({mode})")
    env = os.environ.copy()
    if not use_env_proxy:
        for key in ["HTTP_PROXY", "http_proxy", "HTTPS_PROXY", "https_proxy", "ALL_PROXY", "all_proxy"]:
            env.pop(key, None)
    cmd = ["git", "ls-remote", "https://github.com/"]
    try:
        proc = subprocess.Popen(
            cmd,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
            env=env,
            shell=False,
        )
        out, err = proc.communicate(timeout=30)
        print(f"退出码: {proc.returncode}")
        if out:
            print("stdout:")
            print(out.strip())
        if err:
            print("stderr:")
            print(err.strip())
    except Exception as e:
        print(f"执行 git ls-remote 失败: {e}")


def summarize() -> None:
    print_section("7. 初步诊断建议")
    print("1) 如果环境变量中设置了 HTTP_PROXY/HTTPS_PROXY/ALL_PROXY, 而 git push 报错为无法连接或超时,")
    print("   很可能是 git 也走了这个代理, 但代理不支持或无法访问 GitHub。可以尝试:")
    print("   - 临时清空代理后再执行 git 命令:")
    print("       在当前终端执行:")
    print("         set HTTP_PROXY=")
    print("         set HTTPS_PROXY=")
    print("         set ALL_PROXY=")
    print("       然后再执行 git push")
    print("   - 或在 git 中取消代理配置:")
    print("         git config --global --unset http.proxy")
    print("         git config --global --unset https.proxy")
    print()
    print("2) 如果在直连模式下 (忽略环境变量代理) 能访问 GitHub, 而在使用代理时失败,")
    print("   请检查系统代理软件是否允许 git 进程访问, 以及是否配置了正确的代理端口和协议。")
    print()
    print("3) 如果 DNS 解析失败, 说明系统本身无法解析 github.com, 需要检查系统 DNS 设置或代理的 DNS 规则。")
    print()
    print("4) 请对比本脚本在“使用环境变量代理”和“忽略环境变量代理”两组测试的差异,")
    print("   重点关注: HTTP 请求状态码、git ls-remote 的 stderr 输出, 这些信息能指示出具体问题所在。")


def main() -> None:
    print("GitHub 代理诊断工具")
    print("当前 Python 可执行文件:", sys.executable)
    check_env_proxies()
    check_git_proxy()
    check_dns("github.com")
    check_tcp_connect("github.com", 443, use_proxy=True)
    check_http_request("https://github.com/", use_env_proxy=True)
    check_http_request("https://github.com/", use_env_proxy=False)
    check_git_ls_remote(use_env_proxy=True)
    check_git_ls_remote(use_env_proxy=False)
    summarize()


if __name__ == "__main__":
    main()
