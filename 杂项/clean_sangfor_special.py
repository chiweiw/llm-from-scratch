import os
import sys
import ctypes
import subprocess
import logging
import time

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[logging.StreamHandler(sys.stdout)]
)

class SangforCleaner:
    def __init__(self):
        self.is_admin = self._check_admin()
        
    def _check_admin(self) -> bool:
        """检查管理员权限"""
        try:
            return ctypes.windll.shell32.IsUserAnAdmin()
        except:
            return False

    def _run_cmd(self, cmd_list):
        """执行命令并返回结果"""
        try:
            logging.info(f"执行命令: {' '.join(cmd_list)}")
            result = subprocess.run(
                cmd_list,
                capture_output=True,
                text=True,
                encoding='gbk',
                errors='replace',
                creationflags=subprocess.CREATE_NO_WINDOW
            )
            if result.returncode == 0:
                logging.info("  -> 成功")
                return True, result.stdout
            else:
                logging.warning(f"  -> 失败: {result.stderr.strip() or result.stdout.strip()}")
                return False, result.stderr
        except Exception as e:
            logging.error(f"  -> 执行异常: {e}")
            return False, str(e)

    def disable_services(self):
        """1. 服务强制禁用"""
        logging.info("="*50)
        logging.info("任务 1: 强制禁用 Sangfor 相关服务")
        logging.info("="*50)
        
        target_services = ['SangforPWEx', 'SangforSPVDI']
        
        for svc in target_services:
            logging.info(f"处理服务: {svc}")
            # 1. 停止服务
            self._run_cmd(['sc', 'stop', svc])
            # 2. 禁用服务 (start= disabled) 注意空格
            success, _ = self._run_cmd(['sc', 'config', svc, 'start=', 'disabled'])
            if success:
                logging.info(f"  服务 {svc} 已被彻底禁用。")
            else:
                logging.warning(f"  无法禁用服务 {svc} (可能服务不存在或权限不足)。")

    def delete_drivers(self):
        """2. 驱动强力注销"""
        logging.info("="*50)
        logging.info("任务 2: 强力注销 Sangfor 内核驱动")
        logging.info("="*50)
        
        target_drivers = ['SangforVnic', 'sfusbhub', 'sfvusb']
        
        for drv in target_drivers:
            logging.info(f"处理驱动: {drv}")
            # 1. 停止驱动
            self._run_cmd(['sc', 'stop', drv])
            # 2. 删除驱动
            success, _ = self._run_cmd(['sc', 'delete', drv])
            if success:
                logging.info(f"  驱动 {drv} 已被删除。")
            else:
                logging.warning(f"  无法删除驱动 {drv} (可能驱动不存在)。")

    def file_placeholder(self):
        """3. 文件占位 (SangforTcpX64.dll)"""
        logging.info("="*50)
        logging.info("任务 3: SangforTcpX64.dll 文件占位处理")
        logging.info("="*50)
        
        target_filename = "SangforTcpX64.dll"
        
        # 搜索路径：ProgramFiles 和 AppData 下的 Sangfor 目录
        search_roots = [
            os.environ.get("ProgramFiles"),
            os.environ.get("ProgramFiles(x86)"),
            os.environ.get("AppData"),
            os.environ.get("LocalAppData")
        ]
        
        found_files = []
        logging.info("正在搜索目标文件 (仅限 Sangfor 相关目录)...")
        
        for root_dir in search_roots:
            if not root_dir or not os.path.exists(root_dir):
                continue
            
            # 优化：只在包含 'Sangfor' 的子目录中搜索，提高效率并减少误判
            for root, dirs, files in os.walk(root_dir):
                # 检查当前目录是否可能是 Sangfor 的安装目录
                if 'sangfor' in root.lower() or 'easyconnect' in root.lower():
                    if target_filename in files:
                        full_path = os.path.join(root, target_filename)
                        found_files.append(full_path)
        
        if not found_files:
            logging.info(f"未在常用路径找到 {target_filename}，跳过占位操作。")
            return

        for file_path in found_files:
            logging.info(f"发现文件: {file_path}")
            dead_path = file_path + ".dead"
            
            # 1. 尝试重命名
            try:
                if os.path.exists(dead_path):
                    try:
                        os.remove(dead_path) # 如果 .dead 已存在，先删除
                    except:
                        pass
                
                os.rename(file_path, dead_path)
                logging.info(f"  -> 重命名成功: {os.path.basename(dead_path)}")
            except OSError as e:
                logging.error(f"  -> 重命名失败: {e}")
                logging.info("  尝试使用 move 命令强制移动...")
                # 尝试 cmd move
                self._run_cmd(['cmd', '/c', 'move', '/Y', file_path, dead_path])
                
                if os.path.exists(file_path):
                    logging.error("  -> 文件仍然存在，可能被进程锁定。")
                    continue

            # 2. 创建同名空文件夹
            try:
                os.makedirs(file_path, exist_ok=True)
                logging.info(f"  -> 占位文件夹创建成功: {file_path}")
                
                # 3. (可选) 设置文件夹权限为只读，防止被修改? 暂不实现以免过于复杂
            except Exception as e:
                logging.error(f"  -> 创建占位文件夹失败: {e}")

    def run(self):
        print("Sangfor 残留清理专用脚本 v1.0")
        if not self.is_admin:
            logging.error("错误：请以管理员身份运行此脚本！")
            logging.error("右键 -> 以管理员身份运行")
            input("按 Enter 退出...")
            return

        self.disable_services()
        self.delete_drivers()
        self.file_placeholder()
        
        logging.info("\n清理完成！建议重启计算机以生效。")
        input("按 Enter 退出...")

if __name__ == "__main__":
    cleaner = SangforCleaner()
    cleaner.run()
