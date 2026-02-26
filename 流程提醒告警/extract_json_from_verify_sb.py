import os
import json
import sys
import pymysql


sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
try:
    from db_envs import get_db
except ImportError:
    def get_db(env):
        return None


def collect_nodes(obj):
    result = []
    if isinstance(obj, dict):
        node_type = obj.get("type")
        if node_type == "task" and "id" in obj:
            node_id = obj.get("id")
            node_name = obj.get("name") or obj.get("text")
            if not node_name:
                params = obj.get("parameters")
                if isinstance(params, dict):
                    node_name = params.get("taskName")
            if node_id and node_name:
                result.append((str(node_id), str(node_name)))
        for v in obj.values():
            result.extend(collect_nodes(v))
    elif isinstance(obj, list):
        for item in obj:
            result.extend(collect_nodes(item))
    return result


def get_cfg_json_from_db():
    db_config = get_db("dev")
    if not db_config:
        return None, None
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
            cursor.execute(
                "select TMPL_NAME, CFG_JSON from SYS_WF_TEMPLATE "
                "where TMPL_NAME like '%产品发行审批%' limit 1"
            )
            row = cursor.fetchone()
    finally:
        conn.close()
    if not row:
        return None, None
    name, cfg = row
    return name, cfg


def generate_issue_node_cte(nodes):
    if not nodes:
        return None
    lines = []
    lines.append("issue_node_cfg as (")
    for idx, (node_id, name) in enumerate(nodes):
        node_id_sql = node_id.replace("'", "''")
        name_sql = name.replace("'", "''")
        if idx == 0:
            prefix = "    select"
        else:
            prefix = "    union all select"
        lines.append(
            f"{prefix} '{node_id_sql}' as NODE_ID, '{name_sql}' as NODE_NAME"
        )
    lines.append(")")
    return "\n".join(lines)


def update_sql_file(sql_path, node_cte_sql):
    with open(sql_path, "r", encoding="utf-8") as f:
        text = f.read()
    if "issue_node_cfg as (" in text:
        start = text.index("issue_node_cfg as (")
        end = text.index(")\nselect distinct", start)
        before = text[:start]
        after = text[end + 2 :]
        new_text = before + node_cte_sql + "\n" + after
    else:
        pattern = ")\nselect distinct"
        if pattern not in text:
            print("未找到插入 issue_node_cfg 的位置")
            return False
        new_text = text.replace(
            pattern, "),\n" + node_cte_sql + "\nselect distinct", 1
        )
    with open(sql_path, "w", encoding="utf-8") as f:
        f.write(new_text)
    return True


def main():
    tmpl_name, cfg = get_cfg_json_from_db()
    if not cfg:
        print("未从数据库获取到 CFG_JSON")
        return
    try:
        data = json.loads(cfg)
    except Exception as e:
        print("CFG_JSON 解析失败:", e)
        return
    nodes = collect_nodes(data)
    if not nodes:
        print("CFG_JSON 中未发现包含 id 和 name 的节点")
        return
    base_dir = os.path.dirname(os.path.abspath(__file__))
    sql_path = os.path.join(base_dir, "流程提醒.sql")
    node_cte_sql = generate_issue_node_cte(nodes)
    if not node_cte_sql:
        print("节点列表为空，未生成 SQL 片段")
        return
    ok = update_sql_file(sql_path, node_cte_sql)
    if ok:
        print("已基于模板", tmpl_name, "更新流程提醒.sql 中的 issue_node_cfg 片段")
    else:
        print("流程提醒.sql 更新失败")


if __name__ == "__main__":
    main()


