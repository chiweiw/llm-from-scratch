import os
import sys
import ctypes
import psutil
import subprocess
import logging
import time
from typing import List, Dict, Optional

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.StreamHandler(sys.stdout)
    ]
)

class SystemCleaner:
    """
    系统残留服务与驱动深度排查工具
    用于深度检测、停止并清理系统中残留的“流氓”级办公软件
    """
    
    KEYWORDS = ['Sangfor', 'SvcHost', 'EasyConnect', 'Tencent', 'iOA']
    
    def __init__(self):
        self.is_admin = self._check_admin()

    def _check_admin(self) -> bool:
        """检查是否有管理员权限"""
        try:
            return ctypes.windll.shell32.IsUserAnAdmin()
        except:
            return False

    def _run_cmd(self, cmd: List[str], encoding='gbk') -> str:
        """执行系统命令并返回输出"""
        try:
            result = subprocess.run(
                cmd,
                capture_output=True,
                text=True,
                encoding=encoding,
                errors='replace',  # 防止解码错误
                creationflags=subprocess.CREATE_NO_WINDOW
            )
            return result.stdout
        except Exception as e:
            logging.error(f"命令执行失败 {' '.join(cmd)}: {e}")
            return ""

    def process_killer(self):
        """
        1. 进程巡检与强杀
        扫描系统当前所有运行进程，匹配关键词并递归终止
        """
        logging.info("="*50)
        logging.info("开始执行：进程巡检与强杀 (Process Killer)")
        logging.info("="*50)
        
        found_processes = []
        for proc in psutil.process_iter(['pid', 'name', 'exe', 'cmdline']):
            try:
                pinfo = proc.info
                name = pinfo['name'] or ""
                exe = pinfo['exe'] or ""
                cmdline = " ".join(pinfo['cmdline'] or [])
                
                # 检查是否包含关键字
                matched = False
                for kw in self.KEYWORDS:
                    if (kw.lower() in name.lower() or 
                        kw.lower() in exe.lower() or 
                        kw.lower() in cmdline.lower()):
                        matched = True
                        break
                
                if matched:
                    found_processes.append(proc)
            except (psutil.NoSuchProcess, psutil.AccessDenied, psutil.ZombieProcess):
                continue

        if not found_processes:
            logging.info("未发现相关残留进程。")
            return

        logging.info(f"发现 {len(found_processes)} 个可疑进程：")
        for p in found_processes:
            try:
                logging.info(f"  [PID: {p.pid}] {p.name()} - {p.exe()}")
            except:
                pass

        confirm = input("\n是否确认终止上述所有进程？(y/n): ").strip().lower()
        if confirm == 'y':
            for p in found_processes:
                try:
                    # 递归终止进程树
                    children = p.children(recursive=True)
                    for child in children:
                        try:
                            child.kill()
                            logging.info(f"已终止子进程 PID: {child.pid}")
                        except psutil.NoSuchProcess:
                            pass
                        except Exception as e:
                            logging.error(f"终止子进程失败 PID: {child.pid}: {e}")
                    
                    p.kill()
                    logging.info(f"已终止主进程 PID: {p.pid} ({p.name()})")
                except psutil.NoSuchProcess:
                    logging.info(f"进程已不存在 PID: {p.pid}")
                except Exception as e:
                    logging.error(f"终止进程失败 PID: {p.pid}: {e}")
        else:
            logging.info("用户取消操作。")

    def driver_auditor(self):
        """
        2. 内核驱动审计
        获取当前加载的驱动，筛选关键字并标记位置
        """
        logging.info("="*50)
        logging.info("开始执行：内核驱动审计 (Driver Auditor)")
        logging.info("="*50)

        # 使用 driverquery /V /FO CSV 获取详细信息
        output = self._run_cmd(['driverquery', '/V', '/FO', 'CSV'])
        if not output:
            logging.error("无法获取驱动列表")
            return

        import csv
        from io import StringIO
        
        reader = csv.DictReader(StringIO(output))
        found_drivers = []
        
        system_drivers_path = os.path.join(os.environ.get('SystemRoot', 'C:\\Windows'), 'System32', 'drivers')

        for row in reader:
            display_name = row.get('Display Name', '')
            description = row.get('Description', '')
            module_name = row.get('Module Name', '')
            path = row.get('Path', '') # 注意：driverquery 的 path 往往不完整或是逻辑路径

            matched = False
            for kw in self.KEYWORDS:
                if (kw.lower() in display_name.lower() or 
                    kw.lower() in description.lower() or
                    kw.lower() in module_name.lower()):
                    matched = True
                    break
            
            if matched:
                # 尝试定位实际文件
                sys_path = os.path.join(system_drivers_path, f"{module_name}.sys")
                if os.path.exists(sys_path):
                    actual_path = sys_path
                else:
                    actual_path = "未知位置 (建议在 C:\\Windows\\System32\\drivers 中搜索)"
                
                found_drivers.append({
                    'name': module_name,
                    'display': display_name,
                    'desc': description,
                    'path': actual_path
                })

        if not found_drivers:
            logging.info("未发现相关残留驱动。")
        else:
            logging.info(f"发现 {len(found_drivers)} 个可疑驱动：")
            for drv in found_drivers:
                logging.info(f"  [驱动名: {drv['name']}] {drv['display']}")
                logging.info(f"  描述: {drv['desc']}")
                logging.info(f"  推测路径: {drv['path']}")
                logging.info("-" * 30)

    def service_manager(self):
        """
        3. 服务状态扫描
        查找包含关键字的服务，列出状态并提供禁用选项
        """
        logging.info("="*50)
        logging.info("开始执行：服务状态扫描 (Service Manager)")
        logging.info("="*50)

        found_services = []
        for service in psutil.win_service_iter():
            try:
                info = service.as_dict()
                name = info['name']
                display_name = info['display_name']
                
                matched = False
                for kw in self.KEYWORDS:
                    if kw.lower() in name.lower() or kw.lower() in display_name.lower():
                        matched = True
                        break
                
                if matched:
                    found_services.append(service)
            except Exception as e:
                continue

        if not found_services:
            logging.info("未发现相关残留服务。")
            return

        logging.info(f"发现 {len(found_services)} 个可疑服务：")
        for svc in found_services:
            try:
                logging.info(f"  [服务名: {svc.name()}] {svc.display_name()} - 状态: {svc.status()}")
            except:
                pass

        confirm = input("\n是否尝试禁用并停止上述所有服务？(y/n): ").strip().lower()
        if confirm == 'y':
            for svc in found_services:
                svc_name = svc.name()
                logging.info(f"正在处理服务: {svc_name}")
                
                # 尝试停止服务
                try:
                    if svc.status() != 'stopped':
                        svc.stop()
                        logging.info("  已发送停止指令")
                except Exception as e:
                    logging.error(f"  停止服务失败: {e}")
                
                # 尝试禁用服务 (使用 sc config)
                cmd = ['sc', 'config', svc_name, 'start=', 'disabled']
                result = self._run_cmd(cmd)
                if "SUCCESS" in result or "成功" in result:
                    logging.info("  已修改启动类型为：禁用 (Disabled)")
                else:
                    logging.error(f"  禁用服务失败: {result.strip()}")
        else:
            logging.info("用户取消操作。")

    def file_lock_hunter(self, target_file: str = ""):
        """
        4. 文件占用分析与解锁
        """
        logging.info("="*50)
        logging.info("开始执行：文件占用分析与解锁 (File Lock Hunter)")
        logging.info("="*50)
        
        if not target_file:
            target_file = input("请输入要检测的残留文件完整路径 (例如 C:\\Program Files\\Sangfor\\...): ").strip()
        
        if not target_file:
            logging.info("未输入路径，跳过。")
            return

        if not os.path.exists(target_file):
            logging.warning(f"文件不存在: {target_file}")
            # 即使文件不存在，也可能是因为无法访问，继续尝试重命名逻辑可能不合适，但用户可能输错
            return

        logging.info(f"正在分析文件: {target_file}")

        # 尝试检测占用 (依赖 handle.exe)
        handle_exe = "handle.exe"
        # 简单检查 handle.exe 是否在当前目录或 PATH 中
        has_handle = False
        try:
            subprocess.run([handle_exe, '/?'], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
            has_handle = True
        except FileNotFoundError:
            logging.warning("未检测到 handle.exe 工具，无法精确识别占用进程。")
            logging.warning("建议下载 Sysinternals Suite 并将 handle.exe 放入系统路径。")
        
        if has_handle:
            output = self._run_cmd([handle_exe, target_file])
            if target_file in output:
                logging.info("发现以下进程可能正在占用文件：")
                print(output)
            else:
                logging.info("handle.exe 未报告明确的占用信息。")

        # 尝试重命名
        confirm = input(f"是否尝试强制重命名文件 '{os.path.basename(target_file)}' 为 .bak 以便重启删除？(y/n): ").strip().lower()
        if confirm == 'y':
            new_name = target_file + ".bak"
            try:
                if os.path.exists(new_name):
                    os.remove(new_name) # 如果 .bak 已存在，先删除
                os.rename(target_file, new_name)
                logging.info(f"成功重命名为: {new_name}")
                logging.info("请重启计算机后删除该 .bak 文件。")
            except Exception as e:
                logging.error(f"重命名失败: {e}")
                logging.info("文件可能被内核驱动或受保护进程强力锁定。")

    def network_adapter_check(self):
        """
        5. 网络适配器清理
        """
        logging.info("="*50)
        logging.info("开始执行：网络适配器清理 (Network Adapter Check)")
        logging.info("="*50)

        found_adapters = []
        try:
            adapters = psutil.net_if_addrs()
            stats = psutil.net_if_stats()
            
            for name, snics in adapters.items():
                matched = False
                # 检查网卡名称
                if 'Sangfor' in name or 'VPN' in name:
                    matched = True
                
                if matched:
                    is_up = False
                    if name in stats:
                        is_up = stats[name].isup
                    found_adapters.append((name, is_up))
        except Exception as e:
            logging.error(f"获取网络适配器失败: {e}")
            return

        if not found_adapters:
            logging.info("未发现相关虚拟网卡设备。")
        else:
            logging.info(f"发现 {len(found_adapters)} 个可疑网络适配器：")
            for name, is_up in found_adapters:
                status = "启用" if is_up else "禁用"
                logging.info(f"  [网卡名] {name} - 状态: {status}")
            
            logging.info("\n提示：请前往【控制面板 -> 网络和共享中心 -> 更改适配器设置】手动禁用或卸载这些设备。")

    def run(self):
        print(r"""
   _____            _                    _____ _                            
  / ____|          | |                  / ____| |                           
 | (___  _   _  ___| |_ ___ _ __ ___   | |    | | ___  __ _ _ __   ___ _ __ 
  \___ \| | | |/ __| __/ _ \ '_ ` _ \  | |    | |/ _ \/ _` | '_ \ / _ \ '__|
  ____) | |_| | (__| ||  __/ | | | | | | |____| |  __/ (_| | | | |  __/ |   
 |_____/ \__, |\___|\__\___|_| |_| |_|  \_____|_|\___|\__,_|_| |_|\___|_|   
          __/ |                                                             
         |___/                                                              
        """)
        logging.info("系统残留清理工具 v1.0")
        
        if not self.is_admin:
            logging.warning("警告：当前未检测到管理员权限！")
            logging.warning("部分功能（如终止进程、停止服务）可能无法正常工作。")
            logging.warning("请右键脚本 -> 以管理员身份运行。")
            input("按 Enter 键继续运行，或 Ctrl+C 退出...")

        self.process_killer()
        self.driver_auditor()
        self.service_manager()
        self.network_adapter_check()
        
        # 文件锁检查是交互式的，最后执行
        check_file = input("\n是否需要检测特定文件的占用情况？(y/n): ").strip().lower()
        if check_file == 'y':
            self.file_lock_hunter()
        
        logging.info("\n所有扫描任务完成。")
        input("按 Enter 键退出...")

if __name__ == "__main__":
    try:
        cleaner = SystemCleaner()
        cleaner.run()
    except KeyboardInterrupt:
        print("\n程序已终止。")
    except Exception as e:
        logging.error(f"发生未预期的错误: {e}", exc_info=True)
