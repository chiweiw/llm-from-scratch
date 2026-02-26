"""
合瑞产品成立专项提醒 - 测试执行脚本
基于 合瑞专项提醒.sql 执行测试，验证合瑞产品成立当日提醒结果
"""

import pymysql
import pandas as pd
import sys
import os
from datetime import datetime

# 添加项目根目录到路径，以便导入 db_envs
project_root = os.path.dirname(
    os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
)
sys.path.append(project_root)

try:
    from db_envs import get_db
except ImportError:
    sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
    try:
        from db_envs import get_db
    except ImportError:

        def get_db(env):
            return None


ALERT_SQL_FILE = os.path.join(
    os.path.dirname(os.path.abspath(__file__)), "合瑞专项提醒.sql"
)
with open(ALERT_SQL_FILE, "r", encoding="utf-8") as f:
    ALERT_SQL = f.read()


def run_tests():
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
            print("正在执行合瑞专项提醒SQL...")
            cursor.execute(ALERT_SQL)
            columns = [d[0] for d in cursor.description]
            rows = cursor.fetchall()

        df = pd.DataFrame(rows, columns=columns)

        print("\n" + "=" * 60)
        print("合瑞产品成立专项提醒 - 测试结果")
        print("=" * 60)
        print(f"执行时间: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
        print(f"总记录数: {len(df)}")

        if not df.empty:
            print("\n前20条记录预览:")
            if "CONTENT" in df.columns:
                df_display = df.copy()
                df_display["CONTENT"] = (
                    df_display["CONTENT"].astype(str).str.slice(0, 80) + "..."
                )
                cols_to_show = [
                    "account",
                    "CONTENT",
                    "PRD_CODE",
                    "PRD_NAME",
                    "SETUP_DATE",
                    "RECEIVER_NAME",
                ]
                valid_cols = [c for c in cols_to_show if c in df.columns]
                print(df_display[valid_cols].head(20).to_string())
            else:
                print(df.head(20).to_string())
        else:
            print("\n未查询到任何合瑞专项提醒记录")

        output_dir = os.path.dirname(os.path.abspath(__file__))
        timestamp = datetime.now().strftime("%Y%m%d_%H%M")
        output_file = os.path.join(output_dir, f"合瑞专项提醒测试结果_{timestamp}.xlsx")
        df.to_excel(output_file, index=False)
        print(f"\n结果已输出到: {output_file}")

    except Exception as e:
        print(f"执行出错: {e}")
        import traceback

        traceback.print_exc()
    finally:
        conn.close()


if __name__ == "__main__":
    run_tests()

