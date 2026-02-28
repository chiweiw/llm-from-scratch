"""
简易发包工具：本地 Maven 打包 + 多服务器上传 + 远程重启
"""

# ==============================================================================
# 1. 全局配置与日志初始化
# ==============================================================================

import os
import sys
import logging
import shutil
import subprocess
import socket
import posixpath

logging.basicConfig(
    level=logging.DEBUG,
    format="%(asctime)s - %(levelname)s - %(message)s",
    handlers=[],
)
logger = logging.getLogger(__name__)
GLOBAL_LOG_PATH = None

# ------------------------------------------------------------------------------
# 本地机器相关参数（共享脚本时，通常只需要修改这一段）
# ------------------------------------------------------------------------------

# 本地项目根目录：包含 pom.xml 的目录
LOCAL_PROJECT_ROOT = r"D:\javaproject\backcode"
# 本地 JDK 安装路径：用于 Maven 编译环境配置
LOCAL_JAVA_HOME = r"C:\Program Files\Java\jdk1.8.0_202\bin"
# 本地 Maven 可执行：如需使用 IDEA 自带 Maven，请设置为其 mvn.cmd 路径
LOCAL_MAVEN_CMD = r"D:\Program Files\JetBrains\IntelliJ IDEA 2025.3.3\plugins\maven\lib\maven3\bin\mvn.cmd"
# Maven 全局配置文件路径：包含本地仓库、镜像等设置
MAVEN_SETTINGS_PATH = r"D:\java_tools\apache-maven-3.9.12\conf\settings_sgt0903.xml"
# 本地 Maven 仓库路径：用于缓存依赖与本地构建结果
MAVEN_REPO_LOCAL = r"D:\m2\repository"
# 本地 Maven 命令参数数组, 无需修改
LOCAL_MVN_PARAMS = [
    "clean",
    "package",
    "-DskipTests",
    "-s",
    MAVEN_SETTINGS_PATH,
    f"-Dmaven.repo.local={MAVEN_REPO_LOCAL}",
]

CONFIG = {
    # 本地环境相关配置：用于 Maven 打包与本地路径解析
    "local": {
        # 项目根目录：必须包含 pom.xml
        "project_root": LOCAL_PROJECT_ROOT,
        # Maven 命令参数数组：可在此设置 settings 与本地仓库路径
        "mvn_params": list(LOCAL_MVN_PARAMS),
        # 指定 Maven 可执行文件路径（优先于系统 PATH）
        "mvn_cmd": LOCAL_MAVEN_CMD,
        # JDK 安装路径：可填 JDK 根或其 bin 目录；优先使用此路径进行编译环境配置
        "java_home": LOCAL_JAVA_HOME,
        # 是否使用安静模式（-q）：仅在 compact_mvn_log=True 时启用
        "mvn_quiet": True,
        # 控制台是否精简 Maven 日志：True 打印摘要；False 打印完整 STDOUT
        "compact_mvn_log": True,
        # 是否显式指定 pom 文件（-f <path>/pom.xml）：对齐 IDEA 行为
        "mvn_use_f_pom": True,
        # 是否离线构建（-o/--offline）：仅使用本地仓库，不访问远端
        "mvn_offline": True,
    },
    # 多服务器配置：每个节点独立配置上传目录与重启脚本
    "servers": [
        {
            # 主机地址
            "host": "192.168.8.26",
            # SSH 端口
            "port": 22221,
            # 登录用户
            "username": "omp",
            # 登录口令（明文）
            "password": "cB7JzLsk",
            # 远端部署目录（需具备写权限）
            "remote_dir": "/home/omp/shanguotou/jar/",
            # 重启脚本路径（需具备执行权限）
            "restart_sh_cmd": "/home/omp/shanguotou/jar/restart_jar_dev.sh",
            # 是否在部署后执行重启脚本
            "enable_restart": True,
            # 是否使用 sudo 执行重启脚本（需免密 sudo）
            "use_sudo": False
        }
    ],
    # 目标文件配置：本地 Jar 与远端部署文件名的映射
    "targets": [
        {
            # 本地打包后生成的 Jar 相对路径（相对于 project_root）
            "jar_path": r"startup\platform-startup-project\target\platform-startup-project.jar",
            # 部署到远端后的文件名（如不设置，则默认与本地 Jar 同名）
            # "remote_name": "platform-startup-project.jar",
        },
        {
            "jar_path": r"startup\platform-startup-system\target\platform-startup-system.jar",
        },
        {
            "jar_path": r"startup\platform-startup-customer\target\platform-startup-customer.jar",
        }
    ],
    # 全流程超时时间（秒）
    "timeout": 600,
    # 干跑模式：仅做自检，不执行打包与后续步骤
    "dry_run": False,
    # 备份前是否清理旧备份文件（*.bak）
    "backup_cleanup": True,
}


def setup_logging(cfg):
    """
    方法: 初始化全局日志（控制台精简 + 文件全量）
    参数: cfg(dict) - CONFIG 配置
    返回: None
    说明:
      - 控制台日志级别由 local.compact_mvn_log 控制（True -> INFO；False -> DEBUG）
      - 文件日志恒为 DEBUG，全量记录脚本过程与 Maven 输出
    """
    global GLOBAL_LOG_PATH
    ts = __import__("time").strftime("%Y%m%d_%H%M%S", __import__("time").localtime())
    script_dir = os.path.dirname(os.path.abspath(__file__))
    GLOBAL_LOG_PATH = os.path.join(script_dir, f"deploy_dev_{ts}.log")
    root = logging.getLogger()
    root.handlers = []
    root.setLevel(logging.DEBUG)
    ch = logging.StreamHandler(sys.stdout)
    compact = bool(cfg.get("local", {}).get("compact_mvn_log"))
    ch.setLevel(logging.INFO if compact else logging.DEBUG)
    fh = logging.FileHandler(GLOBAL_LOG_PATH, encoding="utf-8")
    fh.setLevel(logging.DEBUG)
    fmt = logging.Formatter("%(asctime)s - %(levelname)s - %(message)s")
    ch.setFormatter(fmt)
    fh.setFormatter(fmt)
    root.addHandler(ch)
    root.addHandler(fh)
    logger.info(f"日志文件: {GLOBAL_LOG_PATH}")

# ==============================================================================
# 2. 基础工具与本地环境检测 (Maven, JDK, 路径校验)
# ==============================================================================

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

def ensure_jdk(cfg):
    """
    方法: 检测并设置 JDK 环境
    参数: cfg(dict)
    返回: bool - 是否可用
    说明: 校验 javac 存在，输出 Java 与 Javac 版本，并设置 JAVA_HOME 与 PATH
    """
    is_win = os.name == "nt"
    javac_name = "javac.exe" if is_win else "javac"
    java_name = "java.exe" if is_win else "java"
    java_home_cfg = cfg.get("local", {}).get("java_home")
    javac_path = shutil.which("javac")
    java_path = shutil.which("java")
    resolved_home = None
    if not javac_path and java_home_cfg and os.path.isdir(java_home_cfg):
        base = os.path.basename(java_home_cfg).lower()
        if base == "bin":
            bin_dir = java_home_cfg
            candidate = os.path.join(bin_dir, javac_name)
            if os.path.isfile(candidate):
                javac_path = candidate
                resolved_home = os.path.dirname(bin_dir)
        else:
            candidate = os.path.join(java_home_cfg, "bin", javac_name)
            if os.path.isfile(candidate):
                javac_path = candidate
                resolved_home = java_home_cfg
    if javac_path and not resolved_home:
        bin_dir = os.path.dirname(javac_path)
        parent = os.path.dirname(bin_dir)
        if os.path.isdir(parent):
            resolved_home = parent
    if not javac_path:
        logger.error("未检测到 JDK 编译器(javac)，请安装 JDK 或配置 CONFIG.local.java_home / 环境变量 JAVA_HOME")
        return False
    if resolved_home:
        os.environ["JAVA_HOME"] = resolved_home
        bin_path = os.path.join(resolved_home, "bin")
        current_path = os.environ.get("PATH", "")
        paths = current_path.split(os.pathsep) if current_path else []
        if bin_path not in paths:
            os.environ["PATH"] = bin_path + os.pathsep + current_path
    java_cmd = java_path or (os.path.join(os.environ.get("JAVA_HOME", ""), "bin", java_name))
    def run_ver(cmd):
        try:
            r = subprocess.run(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
            out = (r.stdout.strip() + "\n" + r.stderr.strip()).strip()
            return r.returncode, out
        except Exception as e:
            return -1, str(e)
    code_java, ver_java = run_ver([java_cmd, "-version"]) if java_cmd else (-1, "")
    code_javac, ver_javac = run_ver([javac_path, "-version"])
    logger.info(f"JAVA_HOME: {os.environ.get('JAVA_HOME') or '(未设置)'}")
    logger.info(f"java 路径: {java_cmd or '(未知)'}")
    logger.info(f"javac 路径: {javac_path}")
    if ver_java:
        logger.info(f"Java 版本: {ver_java}")
    if ver_javac:
        logger.info(f"Javac 版本: {ver_javac}")
    if code_javac != 0:
        logger.error("JDK 编译器不可用")
        return False
    return True

# ==============================================================================
# 3. 网络与 SSH/SFTP 工具 (Ping, Port, SSH, 上传/执行)
# ==============================================================================

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
        client.connect(
            hostname=host,
            port=port,
            username=username,
            password=password,
            timeout=timeout,
            allow_agent=False,
            look_for_keys=False,
        )
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

def check_sudo_permission(ssh_client):
    """
    方法: 检查当前用户是否具备免密 sudo 权限
    参数: ssh_client(paramiko.SSHClient)
    返回: bool - 是否具备免密 sudo 权限
    说明: 通过 sudo -n -l 校验；不触发交互式输入密码
    """
    try:
        stdin, stdout, stderr = ssh_client.exec_command("sudo -n -l", timeout=10)
        out = stdout.read().decode(errors="ignore").strip()
        err = stderr.read().decode(errors="ignore").strip()
        code = stdout.channel.recv_exit_status()
        if code == 0:
            logger.info("检测到免密 sudo 权限")
            return True
        if err:
            logger.error(f"sudo 权限校验失败: {err}")
        else:
            logger.error("sudo 权限校验失败")
        return False
    except Exception as e:
        logger.error(f"sudo 校验异常: {e}")
        return False


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


# ==============================================================================
# 4. Maven 构建相关 (打包命令, Jar 查找)
# ==============================================================================

def run_maven_build(mvn_cmd, mvn_params, project_root, timeout, compact_log=False):
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
    logger.info(f"工作目录: {project_root}")
    try:
        full_cmd = subprocess.list2cmdline(cmd)
    except Exception:
        full_cmd = " ".join(f'"{x}"' if " " in x else x for x in cmd)
    logger.info(f"完整命令行: {full_cmd}")
    try:
        r = subprocess.run(
            cmd,
            cwd=project_root,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
            timeout=timeout,
        )
    except subprocess.TimeoutExpired:
        logger.error("Maven 打包超时")
        return False
    except Exception as e:
        logger.error(f"Maven 打包执行失败: {e}")
        return False
    logger.debug(f"Maven 工作目录: {project_root}")
    logger.debug(f"Maven 完整命令行: {full_cmd}")
    if r.stdout:
        logger.debug("=== Maven STDOUT ===")
        logger.debug(r.stdout)
    if r.stderr:
        logger.debug("=== Maven STDERR ===")
        logger.debug(r.stderr)
    if GLOBAL_LOG_PATH:
        logger.info(f"Maven 全量日志已写入: {GLOBAL_LOG_PATH}")
    if not compact_log:
        logger.info(r.stdout)
    else:
        out = r.stdout or ""
        lines = out.splitlines()
        errors = [l for l in lines if l.startswith("[ERROR]")]
        status = [l for l in lines if "BUILD SUCCESS" in l or "BUILD FAILURE" in l]
        summary = []
        idx = -1
        for i, l in enumerate(lines):
            if l.startswith("[INFO] Reactor Summary"):
                idx = i
                break
        if idx != -1:
            for j in range(idx, len(lines)):
                summary.append(lines[j])
                if lines[j].strip().startswith("[INFO] ------------------------------------------------------------------------") and j > idx:
                    break
        phase = [l for l in lines if l.startswith("[INFO] --- ")]
        building = [l for l in lines if l.startswith("[INFO] Building ")]
        for l in building[:10]:
            logger.info(l)
        for l in phase[:15]:
            logger.info(l)
        if summary:
            for l in summary:
                logger.info(l)
        if status:
            for l in status:
                logger.info(l)
        if errors:
            for l in errors:
                logger.error(l)
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
    if os.path.isfile(p) and os.path.getsize(p) > 0:
        logger.info(f"检测到 Jar: {p}")
        return p
    candidates = []
    for root, dirs, files in os.walk(project_root):
        if os.path.basename(root) == "target":
            candidate = os.path.join(root, jar_name)
            if os.path.isfile(candidate) and os.path.getsize(candidate) > 0:
                candidates.append(candidate)
    if candidates:
        candidates.sort(key=lambda x: os.path.getmtime(x), reverse=True)
        select = candidates[0]
        logger.info(f"检测到多模块 Jar，使用最新文件: {select}")
        return select
    logger.error(f"Jar 文件不存在: {p}")
    return None


# ==============================================================================
# 5. 核心阶段 (Stage) 实现 (前置, 打包, 备份, 发送, 校验, 启动)
# ==============================================================================

def make_rename_mapping(cfg):
    """
    方法: 更名策略映射（预留）
    参数: cfg(dict)
    返回: list[dict] - [{jar_name, remote_dir, remote_name}]
    说明: 使用映射配置定义本地 Jar 与远端文件名关系，默认按 targets 输出
    """
    return cfg.get("targets", [])


def backup_remote_file(sftp, remote_dir, remote_name, cleanup=False):
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
    if cleanup:
        try:
            dir_list = sftp.listdir(path)
            for name in dir_list:
                if name.startswith(remote_name + ".") and name.endswith(".bak"):
                    bak_path = posixpath.join(path, name)
                    try:
                        sftp.remove(bak_path)
                        logger.info(f"已清理旧备份: {bak_path}")
                    except Exception as e:
                        logger.error(f"旧备份删除失败: {bak_path}, {e}")
        except Exception as e:
            logger.error(f"遍历备份文件失败: {path}, {e}")
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


def get_server_list(cfg):
    """
    方法: 获取 servers 列表
    参数: cfg(dict) - CONFIG 配置
    返回: list[dict] - servers 列表（未配置时返回空数组）
    """
    s = cfg.get("servers")
    return s if isinstance(s, list) else []


def get_password_for_server(server):
    """
    方法: 获取单个 server 配置中的明文密码
    参数: server(dict) - 单个服务器配置
    返回: str | None - 密码
    """
    pwd = server.get("password")
    if pwd and pwd != "your_password":
        return pwd
    logger.error("未提供服务器明文密码")
    return None


def precheck_server(server, mappings, timeout):
    """
    方法: 单服务器前置校验并建立 SSH 连接
    参数: server(dict), mappings(list[dict]), timeout(int)
    返回: paramiko.SSHClient | None - 成功返回已连接对象，失败返回 None
    说明: 校验网络、端口、SSH 登录、远程目录可写、重启脚本可执行、sudo 权限（可选）
    """
    host = server["host"]
    port = int(server["port"])
    username = server["username"]
    restart_cmd = server.get("restart_sh_cmd")
    enable_restart = bool(server.get("enable_restart", True))
    use_sudo = bool(server.get("use_sudo", False))
    if not ping_host(host):
        logger.error(f"服务器网络不可达: {host}")
        return None
    if not check_port_open(host, port, timeout=3):
        logger.error(f"服务器端口不可访问: {host}:{port}")
        return None
    password = get_password_for_server(server)
    if not password:
        return None
    ssh = connect_ssh(host, port, username, password, timeout)
    if not ssh:
        return None
    try:
        sftp = ssh.open_sftp()
        remote_dir = server.get("remote_dir")
        if not remote_dir:
            logger.error("服务器未配置 remote_dir")
            sftp.close()
            ssh.close()
            return None
        if not ensure_remote_dir(sftp, remote_dir):
            sftp.close()
            ssh.close()
            return None
        if enable_restart and restart_cmd:
            if not check_remote_script_executable(sftp, restart_cmd):
                sftp.close()
                ssh.close()
                return None
        if enable_restart and use_sudo:
            if not check_sudo_permission(ssh):
                sftp.close()
                ssh.close()
                return None
        sftp.close()
        logger.info(f"服务器就绪: {host}:{port} 用户:{username} 重启:{enable_restart} sudo:{use_sudo}")
        return ssh
    except Exception as e:
        logger.error(f"远程资源校验失败: {e}")
        ssh.close()
        return None


def start_stage_for_server(ssh, server, timeout):
    """
    方法: 单服务器启动阶段（可选 sudo）
    参数: ssh(paramiko.SSHClient), server(dict), timeout(int)
    返回: bool - 是否成功
    """
    restart_cmd = server.get("restart_sh_cmd")
    enable_restart = bool(server.get("enable_restart", True))
    use_sudo = bool(server.get("use_sudo", False))
    if not enable_restart or not restart_cmd:
        logger.info("已跳过启动阶段（按配置）")
        return True
    logger.info("开始启动阶段")
    restart_dir = posixpath.dirname(restart_cmd)
    restart_base = posixpath.basename(restart_cmd)
    if use_sudo:
        cmd = f"cd {restart_dir} && sudo -n ./{restart_base}"
    else:
        cmd = f"cd {restart_dir} && ./{restart_base}"
    if use_sudo:
        logger.info("使用 sudo 执行重启脚本")
    ok = exec_restart(ssh, cmd, timeout)
    if ok:
        logger.info("启动阶段完成")
    return ok

def build_stage(cfg, mvn_cmd):
    """
    方法: 打包阶段
    返回: bool - 打包是否成功
    """
    timeout = int(cfg.get("timeout", 600))
    params = list(cfg["local"]["mvn_params"])
    quiet = bool(cfg["local"].get("mvn_quiet"))
    compact = bool(cfg["local"].get("compact_mvn_log"))
    if quiet and compact and "-q" not in params:
        params.insert(0, "-q")
    if cfg["local"].get("mvn_use_f_pom"):
        pom = os.path.join(cfg["local"]["project_root"], "pom.xml")
        params += ["-f", pom]
    if cfg["local"].get("mvn_offline"):
        if all(x not in params for x in ("-o", "--offline")):
            params.insert(0, "--offline")
    return run_maven_build(
        mvn_cmd, params, cfg["local"]["project_root"], timeout, compact_log=cfg["local"].get("compact_mvn_log", False)
    )


def rename_stage(cfg):
    """
    方法: 更名映射阶段（预留）
    返回: list[dict] - [{jar_name, remote_dir, remote_name}]
    """
    return make_rename_mapping(cfg)


def backup_stage(ssh, remote_dir, mappings):
    """
    方法: 备份阶段
    返回: bool - 是否成功
    """
    try:
        sftp = ssh.open_sftp()
        logger.info("开始备份阶段")
        for t in mappings:
            rn = t.get("remote_name")
            if not rn:
                rn = os.path.basename(t.get("jar_path", t.get("jar_name", "")))
            cleanup = bool(CONFIG.get("backup_cleanup", False))
            if not backup_remote_file(
                sftp, remote_dir, rn, cleanup=cleanup
            ):
                sftp.close()
                return False
        sftp.close()
        logger.info("备份阶段完成")
        return True
    except Exception as e:
        logger.error(f"远程备份异常: {e}")
        return False


def send_stage(ssh, cfg, remote_dir, mappings):
    """
    方法: 发送阶段
    返回: bool - 是否成功
    """
    logger.info("开始发送阶段")
    for t in mappings:
        local_path = None
        if "jar_path" in t:
            jp = t["jar_path"]
            p = jp if os.path.isabs(jp) else os.path.join(cfg["local"]["project_root"], jp)
            if os.path.isfile(p) and os.path.getsize(p) > 0:
                local_path = p
            else:
                logger.error(f"Jar 文件不存在: {p}")
                return False
        else:
            local_path = check_jar_file(cfg["local"]["project_root"], t["jar_name"])
        if not local_path:
            return False
        remote_name = t.get("remote_name") or os.path.basename(local_path)
        logger.info(f"准备上传: {local_path} -> {remote_dir}{remote_name}")
        if not upload_jar(ssh, local_path, remote_dir, remote_name):
            return False
    logger.info("发送阶段完成")
    return True


def verify_stage(ssh, remote_dir, mappings):
    """
    方法: 校验阶段
    返回: bool - 是否成功
    """
    try:
        sftp = ssh.open_sftp()
        logger.info("开始校验阶段")
        for t in mappings:
            remote_name = t.get("remote_name")
            if not remote_name:
                remote_name = os.path.basename(t.get("jar_path", t.get("jar_name", "")))
            if not verify_remote_file(sftp, remote_dir, remote_name):
                sftp.close()
                return False
        sftp.close()
        logger.info("校验阶段完成")
        return True
    except Exception as e:
        logger.error(f"远程校验异常: {e}")
        return False


# ==============================================================================
# 6. 主流程入口 (Main)
# ==============================================================================

def main():
    """
    方法: 主流程调度（零参数运行）
    阶段: 前置校验 -> 打包 -> 更名映射 -> 备份 -> 发送 -> 校验 -> 启动
    """
    # 1. 初始配置与日志初始化
    cfg = CONFIG
    setup_logging(cfg)

    # 2. 基础配置检查（必须配置 servers）
    servers = get_server_list(cfg)
    if not servers:
        logger.error("未配置 servers")
        sys.exit(2)

    # 3. 本地环境检测（JDK / Maven / 项目目录 / settings / 本地仓库）
    timeout = int(cfg.get("timeout", 600))
    mvn_cmd = cfg["local"].get("mvn_cmd") or resolve_mvn_cmd()
    settings_path, repo_local = parse_mvn_settings(cfg["local"]["mvn_params"])
    if not ensure_jdk(cfg):
        sys.exit(2)
    if not ensure_local_paths(settings_path, repo_local):
        sys.exit(2)
    if not check_project_root(cfg["local"]["project_root"]):
        sys.exit(2)

    # 4. 目标映射生成（Jar -> remote_name）
    mappings = rename_stage(cfg)

    # 5. 多服务器前置自检与连接建立
    ssh_clients = []
    for server in servers:
        ssh = precheck_server(server, mappings, timeout)
        if not ssh:
            for c in ssh_clients:
                c.close()
            sys.exit(2)
        ssh_clients.append(ssh)

    # 6. 干跑模式：仅自检，不执行打包与部署
    if cfg.get("dry_run"):
        logger.info("前置校验通过，干跑结束")
        for c in ssh_clients:
            c.close()
        sys.exit(0)

    # 7. Maven 打包阶段（仅执行一次）
    ok_build = build_stage(cfg, mvn_cmd)
    if not ok_build:
        for c in ssh_clients:
            c.close()
        sys.exit(2)
    logger.info("打包阶段完成")
    logger.info(f"更名映射阶段完成: {len(mappings)} 项")

    # 8. 多服务器部署（备份 -> 发送 -> 校验）
    for server, ssh in zip(servers, ssh_clients):
        remote_dir = server.get("remote_dir")
        if not backup_stage(ssh, remote_dir, mappings):
            for c in ssh_clients:
                c.close()
            sys.exit(2)
        logger.info("备份阶段完成")
        if not send_stage(ssh, cfg, remote_dir, mappings):
            for c in ssh_clients:
                c.close()
            sys.exit(2)
        logger.info("发送阶段完成")
        if not verify_stage(ssh, remote_dir, mappings):
            for c in ssh_clients:
                c.close()
            sys.exit(2)
        logger.info("校验阶段完成")

    # 9. 多服务器启动（可选执行）
    ok_all = True
    for server, ssh in zip(servers, ssh_clients):
        if not start_stage_for_server(ssh, server, timeout):
            ok_all = False
    for c in ssh_clients:
        c.close()
    if not ok_all:
        sys.exit(2)
    logger.info("打包 - 备份 - 发送 - 校验 - 启动 全流程完成")


if __name__ == "__main__":
    main()
