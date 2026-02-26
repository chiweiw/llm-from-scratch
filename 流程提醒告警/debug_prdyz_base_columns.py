import os
import sys

import pymysql


sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
try:
    from db_envs import get_db
except ImportError:
    def get_db(env):
        return None


def main():
    db_config = get_db("dev")
    if not db_config:
        print("Error: DB Config not found")
        return
    conn = pymysql.connect(
        host=db_config[0],
        port=db_config[1],
        user=db_config[2],
        password=db_config[3],
        database=db_config[4],
        charset="utf8mb4",
    )
    try:
        with conn.cursor() as cursor:
            cursor.execute("DESCRIBE ODS_PRDYZ_BASE_INFO")
            rows = cursor.fetchall()
        for col in rows:
            print(col[0])
    finally:
        conn.close()


if __name__ == "__main__":
    main()

