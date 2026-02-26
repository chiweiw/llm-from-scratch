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


def main():
    if len(sys.argv) < 2:
        print("用法: python dump_prechange_nodes.py <TMPL_NAME like 模式> [env]")
        print("示例: python dump_prechange_nodes.py '%产品发行前变更%' dev")
        return

    tmpl_like = sys.argv[1]
    env = sys.argv[2] if len(sys.argv) > 2 else "dev"

    db_config = get_db(env)
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
            cursor.execute(
                "select TMPL_NAME, CFG_JSON from SYS_WF_TEMPLATE "
                "where TMPL_NAME like %s",
                (tmpl_like,),
            )
            rows = cursor.fetchall()
    finally:
        conn.close()

    if not rows:
        print("未从数据库获取到 CFG_JSON")
        return

    for tmpl_name, cfg in rows:
        print("========================================")
        print("模板名称:", tmpl_name)

        try:
            data = json.loads(cfg)
        except Exception as e:
            print("CFG_JSON 解析失败:", e)
            continue

        nodes = collect_nodes(data)
        if not nodes:
            print("CFG_JSON 中未发现包含 id 和 name 的节点")
            continue

        print("节点列表(id, name)：")
        for node_id, name in nodes:
            print(node_id, "||", name)


if __name__ == "__main__":
    main()
