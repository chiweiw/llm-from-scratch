import os

BASE_DIR = os.path.dirname(os.path.abspath(__file__))

USER_DATA_DIR = os.path.join(BASE_DIR, 'browser_data')
OUTPUT_DIR = os.path.join(BASE_DIR, 'output')

os.makedirs(USER_DATA_DIR, exist_ok=True)
os.makedirs(OUTPUT_DIR, exist_ok=True)

TARGET_USER_ID = 'kaifulee'

MAX_ANSWER_COUNT = 0

LOGIN_WAIT_TIME = 60

RANDOM_SLEEP_MIN = 3
RANDOM_SLEEP_MAX = 6

MAX_RETRY = 3
RETRY_INTERVAL = 5

USER_AGENT = 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36'

ANSWERS_API_PATTERN = 'api/v4/members/*/answers'

PROGRESS_FILE = os.path.join(BASE_DIR, 'progress.json')
