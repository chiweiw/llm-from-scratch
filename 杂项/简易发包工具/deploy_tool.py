import os
import sys
import logging
import shutil
import subprocess
import socket
import posixpath

logging.basicConfig(level=logging.INFO, format="%(asctime)s - %(levelname)s - %(message)s", handlers=[logging.StreamHandler(sys.stdout)])
logger = logging.getLogger(__name__)

CONFIG = {
    "local": {
        "project_root": r"D:\your_actual_project",
        "mvn_params": ["clean", "package", "-DskipTests", "-s", r"D:\java_tools\apache-maven-3.9.12\conf\settings_sgt0903.xml", "-Dmaven.repo.local=D:\m2\repository"]
    },
    "server": {
        "host": "192.168.8.26",
        "port": 22,
        "username": "ch",
        "password": "your_password",
        "restart_sh_cmd": "/opt/app/restart.sh"
    },
    "targets": [
        {
            "jar_name": "trust-fund-1.0.jar",
            "remote_dir": "/opt/app/",
            "remote_name": "trust-fund.jar"
        }
    ],
    "timeout": 600,
    "dry_run": False
}

def resolve_mvn_cmd():
    """
    方法: 解析 Maven 可执行路径
    参数: 无
    返回: str | None - mvn 命令路径
    说明: 优先查找 mvn.cmd，其次 mvn
    """
    cmd = shutil.which("mvn.cmd")
    if cmd:
        return cmd
    cmd = shutil.which("mvn")
    return cmd

def parse_mvn_settings(params):
    """
    方法: 解析 Maven 参数中的 settings 与本地仓库路径
    参数: params(list[str]) - Maven 参数数组
    返回: (settings_path, repo_local)
    说明: 提取 -s 与 -Dmaven.repo.local 的值
    """
    settings_path = None
    repo_local = None
    for i, p in enumerate(params):
        if p == "-s" and i + 1 < len(params):
            settings_path = params[i + 1]
        if isinstance(p, str) and p.startswith("-Dmaven.repo.local="):
            repo_local = p.split("=", 1)[1]
    return settings_path, repo_local

def ensure_local_paths(settings_path, repo_local):
    """
    方法: 校验本地 Maven 路径有效性
    参数: settings_path(str|None), repo_local(str|None)
    返回: bool - 是否有效
    说明: 校验 settings 文件存在并保证本地仓库路径可用
    """
    if settings_path and not os.path.exists(settings_path):
        logger.error(f"Maven settings 文件不存在: {settings_path}")
        return False
    if repo_local:
        try:
            os.makedirs(repo_local, exist_ok=True)
        except Exception as e:
            logger.error(f"本地仓库路径不可用: {repo_local}, {e}")
            return False
    return True

def check_project_root(project_root):
    """
    方法: 校验项目根目录与 pom.xml
    参数: project_root(str)
    返回: bool - 是否有效
    说明: 要求项目根目录存在且包含 pom.xml
    """
    if not os.path.isdir(project_root):
        logger.error(f"项目根目录不存在: {project_root}")
        return False
    pom = os.path.join(project_root, "pom.xml")
    if not os.path.isfile(pom):
        logger.error(f"pom.xml 不存在: {pom}")
        return False
    return True

def ping_host(host):
    """
    方法: 检查服务器网络连通性 (ping)
    参数: host(str)
    返回: bool - 是否可达
    说明: 使用系统 ping 测试单次连通性
    """
    try:
        if os.name == "nt":
            cmd = ["ping", "-n", "1", "-w", "1000", host]
        else:
            cmd = ["ping", "-c", "1", "-W", "1", host]
        r = subprocess.run(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
        return r.returncode == 0
    except Exception:
        return False

def check_port_open(host, port, timeout):
    """
    方法: 检查服务器端口可达性
    参数: host(str), port(int), timeout(float)
    返回: bool - 端口是否可访问
    说明: 使用 socket 进行 TCP 连接测试
    """
    try:
        with socket.create_connection((host, port), timeout=timeout):
            return True
    except Exception:
        return False

def get_password(cfg):
    """
    方法: 获取服务器密码（明文）
    参数: cfg(dict) - 配置
    返回: str | None - 密码
    说明: 必须使用内置明文密码；不支持环境变量或密钥登录
    """
    pwd = cfg.get("server", {}).get("password")
    if pwd and pwd != "your_password":
        return pwd
    logger.error("未提供服务器明文密码，请在 CONFIG.server.password 中填写。")
    return None

def connect_ssh(host, port, username, password, timeout):
    """
    方法: 建立 SSH 连接
    参数: host(str), port(int), username(str), password(str), timeout(int)
    返回: paramiko.SSHClient | None
    说明: 使用用户名密码方式进行连接
    """
    try:
        import paramiko
    except ImportError:
        logger.error("缺少依赖 paramiko，请先安装: pip install paramiko")
        sys.exit(1)
    client = paramiko.SSHClient()
    client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    try:
        client.connect(hostname=host, port=port, username=username, password=password, timeout=timeout, allow_agent=False, look_for_keys=False)
        return client
    except Exception as e:
        logger.error(f"SSH 登录失败: {e}")
        return None

def ensure_remote_dir(sftp, remote_dir):
    """
    方法: 确保远程目录存在且可写
    参数: sftp(paramiko.SFTPClient), remote_dir(str)
    返回: bool - 是否可写
    说明: 逐层创建目录并进行写权限测试
    """
    path = remote_dir.replace("\\", "/")
    if not path.endswith("/"):
        path += "/"
    parts = [p for p in path.split("/") if p]
    cur = ""
    for p in parts:
        cur = cur + "/" + p if cur else "/" + p
        try:
            sftp.stat(cur)
        except IOError:
            try:
                sftp.mkdir(cur)
            except Exception as e:
                logger.error(f"远程目录创建失败: {cur}, {e}")
                return False
    test_name = posixpath.join(path, "._deploy_write_test")
    try:
        f = sftp.open(test_name, "w")
        f.write("test")
        f.flush()
        f.close()
        sftp.remove(test_name)
    except Exception as e:
        logger.error(f"远程目录不可写: {path}, {e}")
        return False
    return True

def check_remote_script_executable(sftp, script_path):
    """
    方法: 检查远程脚本可执行性
    参数: sftp(paramiko.SFTPClient), script_path(str)
    返回: bool - 是否可执行
    说明: 通过 stat 校验执行位
    """
    try:
        st = sftp.stat(script_path)
        if (st.st_mode & 0o111) == 0:
            logger.error(f"远程脚本不可执行: {script_path}")
            return False
        return True
    except Exception as e:
        logger.error(f"远程脚本不存在或不可访问: {script_path}, {e}")
        return False

def run_maven_build(mvn_cmd, mvn_params, project_root, timeout):
    """
    方法: 执行 Maven 打包
    参数: mvn_cmd(str), mvn_params(list[str]), project_root(str), timeout(int)
    返回: bool - 是否打包成功
    说明: 在项目根目录执行 Maven 命令并校验返回码
    """
    if not mvn_cmd:
        logger.error("未找到 mvn.cmd 或 mvn 可执行文件")
        return False
    cmd = [mvn_cmd] + mvn_params
    logger.info("开始执行 Maven 打包")
    logger.info(" ".join(cmd))
    try:
        r = subprocess.run(cmd, cwd=project_root, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True, timeout=timeout)
    except subprocess.TimeoutExpired:
        logger.error("Maven 打包超时")
        return False
    except Exception as e:
        logger.error(f"Maven 打包执行失败: {e}")
        return False
    logger.info(r.stdout)
    if r.returncode != 0:
        logger.error(r.stderr)
        logger.error(f"Maven 打包失败，返回码: {r.returncode}")
        return False
    return True

def check_jar_file(project_root, jar_name):
    """
    方法: 校验本地 Jar 文件
    参数: project_root(str), jar_name(str)
    返回: str | None - Jar 文件本地路径
    说明: 检查 target 下对应 Jar 是否存在且大小>0
    """
    p = os.path.join(project_root, "target", jar_name)
    if not os.path.isfile(p):
        logger.error(f"Jar 文件不存在: {p}")
        return None
    if os.path.getsize(p) <= 0:
        logger.error(f"Jar 文件大小异常: {p}")
        return None
    return p

def upload_jar(ssh_client, local_path, remote_dir, jar_name):
    """
    方法: 上传单个 Jar 文件
    参数: ssh_client(paramiko.SSHClient), local_path(str), remote_dir(str), jar_name(str)
    返回: bool - 上传是否成功
    说明: 使用 SFTP 将本地 Jar 发送到远程目录
    """
    try:
        sftp = ssh_client.open_sftp()
    except Exception as e:
        logger.error(f"打开 SFTP 失败: {e}")
        return False
    try:
        path = remote_dir.replace("\\", "/")
        if not path.endswith("/"):
            path += "/"
        remote_path = posixpath.join(path, jar_name)
        sftp.put(local_path, remote_path)
        sftp.close()
        logger.info(f"上传完成: {remote_path}")
        return True
    except Exception as e:
        logger.error(f"上传失败: {e}")
        return False

def exec_restart(ssh_client, cmd, timeout):
    """
    方法: 执行远程重启脚本
    参数: ssh_client(paramiko.SSHClient), cmd(str), timeout(int)
    返回: bool - 是否执行成功
    说明: 通过 SSH 执行重启脚本并校验退出码
    """
    try:
        stdin, stdout, stderr = ssh_client.exec_command(cmd, timeout=timeout)
        out = stdout.read().decode(errors="ignore")
        err = stderr.read().decode(errors="ignore")
        code = stdout.channel.recv_exit_status()
        logger.info(out)
        if err.strip():
            logger.error(err)
        if code != 0:
            logger.error(f"重启脚本执行失败，退出码: {code}")
            return False
        return True
    except Exception as e:
        logger.error(f"重启脚本执行异常: {e}")
        return False

def make_rename_mapping(cfg):
    """
    方法: 更名策略映射（预留）
    参数: cfg(dict)
    返回: list[dict] - [{jar_name, remote_dir, remote_name}]
    说明: 使用映射配置定义本地 Jar 与远端文件名关系，默认按 targets 输出
    """
    return cfg.get("targets", [])

def backup_remote_file(sftp, remote_dir, remote_name):
    """
    方法: 备份远程文件
    参数: sftp(paramiko.SFTPClient), remote_dir(str), remote_name(str)
    返回: bool - 是否成功或无需备份
    说明: 若目标文件存在，则重命名为 .<timestamp>.bak
    """
    path = remote_dir.replace("\\", "/")
    if not path.endswith("/"):
        path += "/"
    target = posixpath.join(path, remote_name)
    try:
        sftp.stat(target)
    except IOError:
        return True
    ts = __import__("time").strftime("%Y%m%d%H%M%S", __import__("time").localtime())
    bak = f"{target}.{ts}.bak"
    try:
        sftp.rename(target, bak)
        logger.info(f"已备份远程文件: {target} -> {bak}")
        return True
    except Exception as e:
        logger.error(f"远程备份失败: {target}, {e}")
        return False

def verify_remote_file(sftp, remote_dir, remote_name):
    """
    方法: 校验远程文件存在且非空
    参数: sftp(paramiko.SFTPClient), remote_dir(str), remote_name(str)
    返回: bool - 是否通过校验
    说明: 使用 stat 校验文件大小
    """
    path = remote_dir.replace("\\", "/")
    if not path.endswith("/"):
        path += "/"
    target = posixpath.join(path, remote_name)
    try:
        st = sftp.stat(target)
        ok = st.st_size > 0
        if not ok:
            logger.error(f"远程文件为空: {target}")
        return ok
    except Exception as e:
        logger.error(f"远程文件不存在: {target}, {e}")
        return False



def precheck(cfg):
    """
    方法: 前置校验阶段
    步骤: 本地 Maven、项目目录、网络与端口、SSH 登录、远程目录与脚本
    返回: paramiko.SSHClient - 已连接的 SSH 客户端
    """
    timeout = int(cfg.get("timeout", 600))
    mvn_cmd = resolve_mvn_cmd()
    settings_path, repo_local = parse_mvn_settings(cfg["local"]["mvn_params"])
    if not ensure_local_paths(settings_path, repo_local):
        sys.exit(2)
    if not check_project_root(cfg["local"]["project_root"]):
        sys.exit(2)
    host = cfg["server"]["host"]
    port = int(cfg["server"]["port"])
    username = cfg["server"]["username"]
    restart_cmd = cfg["server"]["restart_sh_cmd"]
    if not ping_host(host):
        logger.error(f"服务器网络不可达: {host}")
        sys.exit(2)
    if not check_port_open(host, port, timeout=3):
        logger.error(f"服务器端口不可访问: {host}:{port}")
        sys.exit(2)
    password = get_password(cfg)
    if not password:
        sys.exit(2)
    ssh = connect_ssh(host, port, username, password, timeout)
    if not ssh:
        sys.exit(2)
    try:
        sftp = ssh.open_sftp()
        mappings = make_rename_mapping(cfg)
        for t in mappings:
            if not ensure_remote_dir(sftp, t["remote_dir"]):
                sftp.close()
                ssh.close()
                sys.exit(2)
        if not check_remote_script_executable(sftp, restart_cmd):
            sftp.close()
            ssh.close()
            sys.exit(2)
        sftp.close()
    except Exception as e:
        logger.error(f"远程资源校验失败: {e}")
        ssh.close()
        sys.exit(2)
    return ssh

def build_stage(cfg, mvn_cmd):
    """
    方法: 打包阶段
    返回: bool - 打包是否成功
    """
    timeout = int(cfg.get("timeout", 600))
    return run_maven_build(mvn_cmd, cfg["local"]["mvn_params"], cfg["local"]["project_root"], timeout)

def rename_stage(cfg):
    """
    方法: 更名映射阶段（预留）
    返回: list[dict] - [{jar_name, remote_dir, remote_name}]
    """
    return make_rename_mapping(cfg)

def backup_stage(ssh, mappings):
    """
    方法: 备份阶段
    返回: bool - 是否成功
    """
    try:
        sftp = ssh.open_sftp()
        for t in mappings:
            if not backup_remote_file(sftp, t["remote_dir"], t.get("remote_name", t["jar_name"])):
                sftp.close()
                return False
        sftp.close()
        return True
    except Exception as e:
        logger.error(f"远程备份异常: {e}")
        return False

def send_stage(ssh, cfg, mappings):
    """
    方法: 发送阶段
    返回: bool - 是否成功
    """
    for t in mappings:
        local_path = check_jar_file(cfg["local"]["project_root"], t["jar_name"])
        if not local_path:
            return False
        remote_name = t.get("remote_name", t["jar_name"])
        if not upload_jar(ssh, local_path, t["remote_dir"], remote_name):
            return False
    return True

def verify_stage(ssh, mappings):
    """
    方法: 校验阶段
    返回: bool - 是否成功
    """
    try:
        sftp = ssh.open_sftp()
        for t in mappings:
            remote_name = t.get("remote_name", t["jar_name"])
            if not verify_remote_file(sftp, t["remote_dir"], remote_name):
                sftp.close()
                return False
        sftp.close()
        return True
    except Exception as e:
        logger.error(f"远程校验异常: {e}")
        return False

def start_stage(ssh, cfg):
    """
    方法: 启动阶段
    返回: bool - 是否成功
    """
    restart_cmd = cfg["server"]["restart_sh_cmd"]
    timeout = int(cfg.get("timeout", 600))
    return exec_restart(ssh, restart_cmd, timeout)

def main():
    """
    方法: 主流程调度（零参数运行）
    阶段: 前置校验 -> 打包 -> 更名映射 -> 备份 -> 发送 -> 校验 -> 启动
    """
    cfg = CONFIG
    # 前置校验
    ssh = precheck(cfg)
    if cfg.get("dry_run"):
        logger.info("前置校验通过，干跑结束")
        ssh.close()
        sys.exit(0)
    # 打包
    mvn_cmd = resolve_mvn_cmd()
    ok_build = build_stage(cfg, mvn_cmd)
    if not ok_build:
        ssh.close()
        sys.exit(2)
    # 更名映射
    mappings = rename_stage(cfg)
    # 备份
    if not backup_stage(ssh, mappings):
        ssh.close()
        sys.exit(2)
    # 发送
    if not send_stage(ssh, cfg, mappings):
        ssh.close()
        sys.exit(2)
    # 校验
    if not verify_stage(ssh, mappings):
        ssh.close()
        sys.exit(2)
    # 启动
    if not start_stage(ssh, cfg):
        ssh.close()
        sys.exit(2)
    logger.info("打包 - 备份 - 发送 - 校验 - 启动 全流程完成")
    ssh.close()

    
if __name__ == "__main__":
    main()
