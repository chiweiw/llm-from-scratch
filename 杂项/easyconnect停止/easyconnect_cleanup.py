import os
import sys
import ctypes
import subprocess
import logging
import time
from typing import List, Tuple
try:
    import psutil
except Exception:
    psutil = None

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[logging.StreamHandler(sys.stdout)]
)

class EasyConnectCleaner:
    def __init__(self):
        self.is_admin = self._check_admin()
        self.services_stopped = 0
        self.services_disabled = 0
        self.processes_killed = 0
        self.adapters_reset = 0
        self.target_service_display = {
            'Sangfor Promotion Service',
            'Sangfor TCP Proxy Service',
            'Sangfor Auto-Updater'
        }
        self.target_proc_names = {
            'EasyConnect.exe',
            'ECAgent.exe',
            'SangforPromoteService.exe',
            'SangforProxyService.exe',
            'SangforDaemon.exe',
            'SangforUpdateServer.exe'
        }

    def _check_admin(self) -> bool:
        try:
            return bool(ctypes.windll.shell32.IsUserAnAdmin())
        except Exception:
            return False

    def _run_cmd(self, cmd: List[str]) -> Tuple[int, str, str]:
        try:
            r = subprocess.run(
                cmd,
                capture_output=True,
                text=True,
                encoding='gbk',
                errors='replace',
                creationflags=subprocess.CREATE_NO_WINDOW
            )
            return r.returncode, r.stdout.strip(), r.stderr.strip()
        except Exception as e:
            return 1, '', str(e)

    def _ensure_admin(self):
        if self.is_admin:
            return
        logging.info('正在请求管理员权限...')
        try:
            script = os.path.abspath(__file__)
            args = ' '.join([f'"{a}"' if ' ' in a else a for a in sys.argv[1:]])
            ctypes.windll.shell32.ShellExecuteW(None, 'runas', sys.executable, f'"{script}" {args}', None, 1)
            sys.exit(0)
        except Exception as e:
            logging.error(f'管理员权限请求失败: {e}')
            sys.exit(1)

    def _detect_components(self) -> bool:
        found = False
        if psutil:
            try:
                for svc in psutil.win_service_iter():
                    info = svc.as_dict()
                    name = (info.get('name') or '').lower()
                    display = (info.get('display_name') or '').lower()
                    if 'sangfor' in name or 'sangfor' in display or 'easyconnect' in name or 'easyconnect' in display:
                        found = True
                        break
            except Exception:
                pass
            if not found:
                try:
                    for p in psutil.process_iter(['name', 'exe', 'cmdline']):
                        ni = (p.info.get('name') or '').lower()
                        ei = (p.info.get('exe') or '').lower()
                        ci = ' '.join(p.info.get('cmdline') or []).lower()
                        if 'sangfor' in ni or 'sangfor' in ei or 'sangfor' in ci or 'easyconnect' in ni or 'easyconnect' in ei or 'easyconnect' in ci:
                            found = True
                            break
                except Exception:
                    pass
        else:
            code, out, _ = self._run_cmd(['sc', 'query', 'type=', 'service'])
            if code == 0 and any(k in out.lower() for k in ['sangfor', 'easyconnect']):
                found = True
            if not found:
                code, out, _ = self._run_cmd(['tasklist'])
                if code == 0 and any(k in out.lower() for k in ['sangfor', 'easyconnect', 'easyconnect.exe']):
                    found = True
        if not found:
            logging.info('未检测到相关组件，不执行后续操作。')
        return found

    def stop_and_disable_services(self):
        targets = []
        if psutil:
            try:
                for svc in psutil.win_service_iter():
                    info = svc.as_dict()
                    name = info.get('name') or ''
                    display = info.get('display_name') or ''
                    if display in self.target_service_display:
                        targets.append(name)
                        continue
                    dn = name.lower()
                    dd = display.lower()
                    if dn.startswith('sangfor') or 'easyconnect' in dn or dd.startswith('sangfor') or 'easyconnect' in dd:
                        targets.append(name)
            except Exception:
                pass
        else:
            code, out, _ = self._run_cmd(['sc', 'query', 'type=', 'service'])
            if code == 0:
                lines = out.splitlines()
                cur = ''
                for ln in lines:
                    if ln.strip().startswith('SERVICE_NAME:'):
                        cur = ln.split(':', 1)[1].strip()
                    if ln.strip().startswith('DISPLAY_NAME:'):
                        disp = ln.split(':', 1)[1].strip()
                        dd = disp.lower()
                        if disp in self.target_service_display or dd.startswith('sangfor') or 'easyconnect' in dd:
                            if cur:
                                targets.append(cur)
        if not targets:
            logging.info('未发现需处理的服务')
            return
        unique = list(dict.fromkeys(targets))
        for svc_name in unique:
            try:
                logging.info(f'处理服务: {svc_name}')
                if psutil:
                    try:
                        svc = psutil.win_service_get(svc_name)
                        if svc and svc.status() != 'stopped':
                            try:
                                svc.stop()
                                self.services_stopped += 1
                                time.sleep(0.5)
                            except Exception:
                                pass
                    except Exception:
                        pass
                else:
                    self._run_cmd(['sc', 'stop', svc_name])
                code, out, err = self._run_cmd(['sc', 'config', svc_name, 'start=', 'disabled'])
                if code == 0 and ('SUCCESS' in out or '成功' in out):
                    self.services_disabled += 1
            except Exception:
                continue

    def kill_processes(self):
        names_lower = {n.lower() for n in self.target_proc_names}
        if psutil:
            found = []
            for proc in psutil.process_iter(['pid', 'name', 'exe', 'cmdline']):
                try:
                    name = (proc.info.get('name') or '')
                    exe = (proc.info.get('exe') or '')
                    cmd = ' '.join(proc.info.get('cmdline') or [])
                    nl = name.lower()
                    el = exe.lower()
                    cl = cmd.lower()
                    if nl in names_lower or nl.startswith('sangfor') or 'easyconnect' in nl or 'sangfor' in el or 'easyconnect' in el or 'sangfor' in cl or 'easyconnect' in cl:
                        found.append(proc)
                except Exception:
                    continue
            for p in found:
                try:
                    children = p.children(recursive=True)
                    for c in children:
                        try:
                            c.kill()
                            self.processes_killed += 1
                        except Exception:
                            pass
                    p.kill()
                    self.processes_killed += 1
                except Exception:
                    pass
        else:
            code, out, _ = self._run_cmd(['tasklist'])
            if code == 0:
                lines = out.splitlines()
                to_kill = []
                for ln in lines[3:]:
                    parts = ln.split()
                    if not parts:
                        continue
                    pname = parts[0]
                    pl = pname.lower()
                    if pl in names_lower or pl.startswith('sangfor') or 'easyconnect' in pl:
                        to_kill.append(pname)
                for pname in dict.fromkeys(to_kill):
                    self._run_cmd(['taskkill', '/F', '/IM', pname])

    def reset_adapter(self):
        names = []
        if psutil:
            try:
                names = list(psutil.net_if_addrs().keys())
            except Exception:
                names = []
        else:
            names = []
        candidates = []
        target = 'EasyConnect Virtual Adapter'
        if target in names:
            candidates = [target]
        else:
            for n in names:
                if 'easyconnect' in n.lower():
                    candidates.append(n)
        if not candidates:
            logging.info('未发现 EasyConnect 虚拟网卡')
            return
        for name in dict.fromkeys(candidates):
            logging.info(f'重置网卡: {name}')
            self._run_cmd(['netsh', 'interface', 'set', 'interface', name, 'admin=disabled'])
            time.sleep(1.0)
            self._run_cmd(['netsh', 'interface', 'set', 'interface', name, 'admin=enabled'])
            self.adapters_reset += 1

    def run(self):
        self._ensure_admin()
        if not self._detect_components():
            return
        logging.info('开始清理 EasyConnect 与 Sangfor 相关组件')
        self.stop_and_disable_services()
        self.kill_processes()
        self.reset_adapter()
        logging.info(f'服务已停止:{self.services_stopped} 已禁用:{self.services_disabled}')
        logging.info(f'进程已结束:{self.processes_killed}')
        logging.info(f'网卡已重置:{self.adapters_reset}')
        logging.info('清理完成，建议重启计算机')

if __name__ == '__main__':
    try:
        EasyConnectCleaner().run()
    except Exception as e:
        logging.error(f'执行失败: {e}')
