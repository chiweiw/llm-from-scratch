import os

# ==================== 基础目录配置 ====================
# 获取当前脚本所在目录的绝对路径
BASE_DIR = os.path.dirname(os.path.abspath(__file__))

# 浏览器数据目录：用于存储浏览器的登录状态、cookies等信息
# 首次运行需要手动扫码登录，之后会自动保存登录状态，无需重复登录
USER_DATA_DIR = os.path.join(BASE_DIR, "browser_data")

# 输出目录：爬取的数据将保存为CSV文件到此目录
OUTPUT_DIR = os.path.join(BASE_DIR, "output")

# 自动创建必要的目录
os.makedirs(USER_DATA_DIR, exist_ok=True)
os.makedirs(OUTPUT_DIR, exist_ok=True)


# ==================== 目标用户配置 ====================
# 目标用户ID：从知乎用户主页URL中提取
# 例如：https://www.zhihu.com/people/dong-bu-dong-95-73
# 则 TARGET_USER_ID = 'dong-bu-dong-95-73'
TARGET_USER_ID = "dong-bu-dong-95-73"


# ==================== 爬取数量限制 ====================
# 最大爬取回答数量
# 0 表示不限制，爬取该用户的所有回答
# 设置为正整数（如 100）则只爬取指定数量的回答
MAX_ANSWER_COUNT = 100


# ==================== 登录配置 ====================
# 登录等待时间（秒）
# 首次运行时，程序会打开浏览器并等待您手动扫码登录
# 如果在此时间内未完成登录，程序将退出
# 登录成功后，登录状态会保存在 USER_DATA_DIR 中，下次运行无需重复登录
LOGIN_WAIT_TIME = 60


# ==================== 反爬虫配置 ====================
# 随机等待时间范围（秒）
# 每次翻页前会随机等待 RANDOM_SLEEP_MIN 到 RANDOM_SLEEP_MAX 秒
# 模拟人工操作，降低被知乎检测为爬虫的风险
RANDOM_SLEEP_MIN = 3
RANDOM_SLEEP_MAX = 6


# ==================== 重试配置 ====================
# 最大重试次数：当遇到网络错误或其他异常时的重试次数
MAX_RETRY = 3

# 重试间隔（秒）：每次重试前等待的时间
RETRY_INTERVAL = 5


# ==================== 浏览器配置 ====================
# User-Agent：模拟真实浏览器访问
# 可以根据需要修改为其他浏览器的 User-Agent
USER_AGENT = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"


# ==================== API监听配置 ====================
# 知乎回答列表API的URL模式
# 程序通过监听此API来获取回答数据
# 一般情况下无需修改
ANSWERS_API_PATTERN = "api/v4/members/*/answers"


# ==================== 断点续传配置 ====================
# 进度文件路径：用于保存爬取进度，支持断点续传
# 如果程序中断，下次运行时会从上次的位置继续爬取
PROGRESS_FILE = os.path.join(BASE_DIR, "progress.json")


# ==================== 使用说明 ====================
# 1. 修改 TARGET_USER_ID 为您要爬取的知乎用户ID
# 2. 运行 spider.py
# 3. 首次运行时，浏览器会自动打开，请在 LOGIN_WAIT_TIME 秒内完成扫码登录
# 4. 登录成功后，程序会自动开始爬取数据
# 5. 爬取的数据会保存到 output 目录下的CSV文件中
# 6. 如果程序中断，再次运行会从上次的位置继续爬取（断点续传）
