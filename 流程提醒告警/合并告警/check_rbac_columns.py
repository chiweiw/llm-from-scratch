"""
检查 sys_rbac_user 和 sys_rbac_role 的表结构
"""

import pymysql
import sys
import os

# 添加项目根目录到路径
project_root = os.path.dirname(
    os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
)
sys.path.append(project_root)

try:
    from db_envs import get_db
except ImportError:
    sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
    from db_envs import get_db


def check_columns():
    db_config = get_db("dev")
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
            tables = ["ODS_PRDYZ_BASE_INFO"]
            for table in tables:
                print(f"\nScanning table: {table}")
                try:
                    cursor.execute(f"DESC {table}")
                    columns = cursor.fetchall()
                    print(f"{'Field':<20} {'Type':<15} {'Null':<5} {'Key':<5}")
                    print("-" * 50)
                    for col in columns:
                        # Field, Type, Null, Key, Default, Extra
                        print(f"{col[0]:<20} {col[1]:<15} {col[2]:<5} {col[3]:<5}")
                except Exception as e:
                    print(f"Error describing {table}: {e}")

            # List all tables matching 'role'
            print("\nSearching for tables with 'role' in name:")
            cursor.execute("SHOW TABLES LIKE '%role%'")
            tables = cursor.fetchall()
            for t in tables:
                print(t[0])
    finally:
        conn.close()


if __name__ == "__main__":
    check_columns()
