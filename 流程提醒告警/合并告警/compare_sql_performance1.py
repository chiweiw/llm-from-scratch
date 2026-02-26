import pymysql
import time
import os
import sys
from datetime import datetime


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


output_lines = []


def log(msg):
    print(msg)
    output_lines.append(str(msg))


def get_connection():
    db_config = get_db("dev")
    if not db_config:
        print("Error: DB Config not found")
        return None
    return pymysql.connect(
        host=db_config[0],
        port=db_config[1],
        user=db_config[2],
        password=db_config[3],
        database=db_config[4],
        charset="utf8mb4",
    )


def load_sql(file_name):
    base_dir = os.path.dirname(os.path.abspath(__file__))
    file_path = os.path.join(base_dir, file_name)
    with open(file_path, "r", encoding="utf-8") as f:
        return f.read()


def prepare_sql_for_wrap(sql_text):
    sql = sql_text.strip()
    while sql.endswith(";"):
        sql = sql[:-1].rstrip()
    return sql


def run_for_file(conn, file_name):
    sql_text = load_sql(file_name)
    sql_core = prepare_sql_for_wrap(sql_text)
    count_sql = "select count(1) as cnt from ({}) t".format(sql_core)
    explain_sql = "explain {}".format(sql_core)
    with conn.cursor() as cursor:
        log("")
        log("==============================")
        log("文件: {}".format(file_name))
        log("==============================")
        start = time.time()
        cursor.execute(count_sql)
        row = cursor.fetchone()
        elapsed = time.time() - start
        count_value = row[0] if row else None
        log("记录数: {}".format(count_value))
        log("count 执行耗时(秒): {}".format(round(elapsed, 3)))
        log("")
        log("执行计划:")
        start_explain = time.time()
        cursor.execute(explain_sql)
        columns = [d[0] for d in cursor.description]
        rows = cursor.fetchall()
        for r in rows:
            line = []
            for i, v in enumerate(r):
                line.append("{}={}".format(columns[i], v))
            log("  " + ", ".join(line))
        explain_elapsed = time.time() - start_explain
        log("explain 执行耗时(秒): {}".format(round(explain_elapsed, 3)))


def main():
    conn = get_connection()
    if not conn:
        return
    try:
        # run_for_file(conn, "合并流程提醒 copy.sql")
        # run_for_file(conn, "合并流程提醒 copy.sql")
        run_for_file(conn, "合并流程提醒_optimized.sql")
    except Exception as e:
        log("执行出错: {}".format(e))
        import traceback

        traceback.print_exc()
    finally:
        base_dir = os.path.dirname(os.path.abspath(__file__))
        ts = datetime.now().strftime("%Y%m%d_%H%M%S")
        out_path = os.path.join(base_dir, "compare_sql_performance_{}.txt".format(ts))
        try:
            with open(out_path, "w", encoding="utf-8") as f:
                for line in output_lines:
                    f.write(line + "\n")
            print("结果已写入:", out_path)
        except Exception as e:
            print("写入结果文件出错:", e)
        conn.close()


if __name__ == "__main__":
    main()
