"""
合并流程提醒告警 - 测试执行脚本
基于 合并流程提醒.sql 执行测试，验证能否同时覆盖申报登记和发行审批
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
    # 尝试从当前目录的上两级查找 (兼容性处理)
    sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
    try:
        from db_envs import get_db
    except ImportError:

        def get_db(env):
            return None


# 读取SQL
ALERT_SQL_FILE = os.path.join(
    os.path.dirname(os.path.abspath(__file__)), "合并流程提醒.sql"
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
            print("正在执行合并流程提醒SQL...")
            cursor.execute(ALERT_SQL)
            columns = [d[0] for d in cursor.description]
            rows = cursor.fetchall()

        df = pd.DataFrame(rows, columns=columns)

        print("\n" + "=" * 60)
        print("合并流程提醒告警 - 测试结果")
        print("=" * 60)
        print(f"执行时间: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
        print(f"总记录数: {len(df)}")

        if not df.empty:
            # 统计不同告警类型
            if "CONTENT" in df.columns:
                # 简单分类逻辑：看内容前缀
                df["Type_Check"] = df["CONTENT"].apply(
                    lambda x: (
                        "申报登记"
                        if "申报登记" in str(x)
                        else ("发行审批" if "发行审批" in str(x) else "其他")
                    )
                )

                print("\n按告警类型统计:")
                type_stats = df.groupby("Type_Check").size().reset_index(name="数量")
                print(type_stats.to_string(index=False))

            print("\n前20条记录预览:")
            if "CONTENT" in df.columns:
                df_display = df.copy()
                df_display["CONTENT"] = (
                    df_display["CONTENT"].astype(str).str.slice(0, 60) + "..."
                )
                # 选择关键列展示
                cols_to_show = [
                    "account",
                    "CONTENT",
                    "PRD_CODE",
                    "ROLE_TYPE",
                    "DAYS_ELAPSED",
                ]
                valid_cols = [c for c in cols_to_show if c in df.columns]
                print(df_display[valid_cols].head(20).to_string())
            else:
                print(df.head(20).to_string())

        else:
            print("\n未查询到任何告警记录")

        # 输出到Excel
        output_dir = os.path.dirname(os.path.abspath(__file__))
        timestamp = datetime.now().strftime("%Y%m%d_%H%M")
        output_file = os.path.join(output_dir, f"合并告警测试结果_{timestamp}.xlsx")
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
