import pymysql
import pandas as pd
import sys
import os

sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
try:
    from db_envs import get_db
except ImportError:
    def get_db(env):
        return None

ALERT_SQL_FILE = os.path.join(
    os.path.dirname(os.path.abspath(__file__)),
    "流程模板单个SQL",
    "产品申报提醒.sql",
)
with open(ALERT_SQL_FILE, "r", encoding="utf-8") as f:
    ALERT_SQL = f.read()

ALERT_SQL_MANAGER_ONLY = f"""
select *
from (
{ALERT_SQL}
) x
where x.ACCOUNT_ROLE <> ''
"""


def run_tests():
    db_config = get_db('dev')
    if not db_config:
        print("Error: DB Config not found")
        return

    conn = pymysql.connect(
        host=db_config[0],
        port=db_config[1],
        user=db_config[2],
        password=db_config[3],
        database=db_config[4],
        charset='utf8mb4'
    )

    try:
        with conn.cursor() as cursor:
            cursor.execute(ALERT_SQL)
            columns = [d[0] for d in cursor.description]
            rows = cursor.fetchall()

        df = pd.DataFrame(rows, columns=columns)
        if 'CONTENT' in df.columns:
            df['CONTENT'] = df['CONTENT'].astype(str).str.slice(0, 120)
        print("总记录数(全部待发送记录):", len(df))
        if not df.empty:
            print(df.head(20))

        output_dir = os.path.dirname(os.path.abspath(__file__))

        output_file_all = os.path.join(
            output_dir,
            '产品申报登记紧急预警_自测结果.xlsx'
        )
        df.to_excel(output_file_all, index=False)
        print("结果已输出到:", output_file_all)

        with conn.cursor() as cursor:
            cursor.execute(ALERT_SQL_MANAGER_ONLY)
            columns_mgr = [d[0] for d in cursor.description]
            rows_mgr = cursor.fetchall()

        df_mgr = pd.DataFrame(rows_mgr, columns=columns_mgr)
        if 'CONTENT' in df_mgr.columns:
            df_mgr['CONTENT'] = df_mgr['CONTENT'].astype(str).str.slice(0, 120)
        print("总记录数(投资/销售经理):", len(df_mgr))

        output_file_mgr = os.path.join(
            output_dir,
            '产品申报登记紧急预警_投资销售经理_自测结果.xlsx'
        )
        df_mgr.to_excel(output_file_mgr, index=False)
        print("投资/销售经理结果已输出到:", output_file_mgr)
    finally:
        conn.close()


if __name__ == "__main__":
    run_tests()
