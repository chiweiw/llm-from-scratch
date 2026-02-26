"""
产品发行审批流程提醒告警 - 测试执行脚本
基于 发行审批提醒.sql 执行测试
"""

import pymysql
import pandas as pd
import sys
import os
from datetime import datetime

sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
try:
    from db_envs import get_db
except ImportError:

    def get_db(env):
        return None


# 读取发行审批SQL
ALERT_SQL_FILE = os.path.join(
    os.path.dirname(os.path.abspath(__file__)),
    "流程模板单个SQL",
    "发行审批提醒.sql",
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
            print("正在执行发行审批提醒SQL...")
            cursor.execute(ALERT_SQL)
            columns = [d[0] for d in cursor.description]
            rows = cursor.fetchall()

        df = pd.DataFrame(rows, columns=columns)

        print("\n" + "=" * 60)
        print("产品发行审批流程提醒告警 - 测试结果")
        print("=" * 60)
        print(f"执行时间: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
        print(f"总记录数: {len(df)}")

        if not df.empty:
            print("\n前20条记录预览:")
            if "CONTENT" in df.columns:
                df_display = df.copy()
                df_display["CONTENT"] = (
                    df_display["CONTENT"].astype(str).str.slice(0, 80)
                )
                print(df_display.head(20).to_string())
            else:
                print(df.head(20).to_string())

            # 按节点统计
            if "NODE_NAME" in df.columns:
                print("\n按节点统计:")
                node_stats = df.groupby("NODE_NAME").size().reset_index(name="数量")
                print(node_stats.to_string(index=False))

            # 按天数统计
            if "DAYS_ELAPSED" in df.columns:
                print("\n按流程天数统计:")
                days_stats = df.groupby("DAYS_ELAPSED").size().reset_index(name="数量")
                print(days_stats.to_string(index=False))
        else:
            print("\n未查询到符合条件的告警记录")

        # 输出到Excel
        output_dir = os.path.dirname(os.path.abspath(__file__))
        timestamp = datetime.now().strftime("%Y%m%d_%H%M")
        output_file = os.path.join(
            output_dir, f"产品发行审批提醒_自测结果_{timestamp}.xlsx"
        )
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
