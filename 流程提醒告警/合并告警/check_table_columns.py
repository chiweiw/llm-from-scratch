import pymysql
import sys

sys.path.append("e:/pyProject/lm-from-scratch")
from db_envs import get_db

db = get_db("dev")
conn = pymysql.connect(
    host=db[0], port=db[1], user=db[2], password=db[3], database=db[4]
)

cur = conn.cursor()
cur.execute("DESC ODS_PRDYZ_ADD_ISSUE_SON")
print("ODS_PRDYZ_ADD_ISSUE_SON columns:")
for row in cur.fetchall():
    print(f"  {row[0]}")

conn.close()
