import os
import sys
import json
from typing import List, Tuple, Optional

import pymysql


sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
try:
    from db_envs import get_db
except ImportError:

    def get_db(env):
        return None


def get_connection(env: str = "dev") -> Optional[pymysql.connections.Connection]:
    db_config = get_db(env)
    if not db_config:
        return None
    return pymysql.connect(
        host=db_config[0],
        port=db_config[1],
        user=db_config[2],
        password=db_config[3],
        database=db_config[4],
        charset="utf8mb4",
    )


def fetch_templates(tmpl_filter: str) -> List[Tuple[str, str]]:
    conn = get_connection("dev")
    if not conn:
        print("无法获取数据库连接，请检查 db_envs.get_db 配置")
        return []
    try:
        with conn.cursor() as cursor:
            sql = (
                "select TMPL_NAME, CFG_JSON "
                "from SYS_WF_TEMPLATE "
                "where TMPL_NAME like %s "
                "order by TMPL_NAME"
            )
            cursor.execute(sql, (tmpl_filter,))
            rows = cursor.fetchall()
            return [(r[0], r[1]) for r in rows]
    finally:
        conn.close()


def classify_scene(task_name: str) -> Tuple[str, str]:
    name = task_name or ""
    if any(k in name for k in ["发起", "复核"]):
        return "PRE_DEPT_APPR", "提交部门审批前"
    if any(k in name for k in ["待提交", "重新处理"]):
        return "WAIT_SUBMIT", "提交审批中（经办提交节点）"
    if any(k in name for k in ["审批", "领导", "会签"]):
        return "OTHERS", "审批中/其他"
    return "OTHERS", "其他情况"


def collect_task_nodes(obj) -> List[Tuple[str, str, str, str]]:
    result: List[Tuple[str, str, str, str]] = []

    def _walk(o):
        if isinstance(o, dict):
            if o.get("type") == "task":
                node_id = str(o.get("id", ""))
                text = o.get("text")
                params = o.get("parameters") or {}
                task_name = text or params.get("taskName") or ""
                scene_code, scene_name = classify_scene(task_name)
                result.append((node_id, task_name, scene_code, scene_name))
            for v in o.values():
                _walk(v)
        elif isinstance(o, list):
            for item in o:
                _walk(item)

    _walk(obj)
    return result


def print_nodes(tmpl_name: str, nodes: List[Tuple[str, str, str, str]]) -> None:
    print("=" * 80)
    print("模板:", tmpl_name)
    print("-" * 80)
    if not nodes:
        print("未解析到 type='task' 的节点")
        return
    print("{:<40}  {:<20}  {:<15}  {}".format("NODE_ID", "TASK_NAME", "SCENE_CODE", "SCENE_NAME"))
    for node_id, task_name, scene_code, scene_name in nodes:
        print(
            "{:<40}  {:<20}  {:<15}  {}".format(
                node_id, task_name or "", scene_code, scene_name
            )
        )


def main():
    if len(sys.argv) > 1:
        raw = sys.argv[1]
        if "%" in raw:
            tmpl_filter = raw
        else:
            tmpl_filter = f"%{raw}%"
    else:
        tmpl_filter = "%产品发行审批%"

    rows = fetch_templates(tmpl_filter)
    if not rows:
        print("未在 SYS_WF_TEMPLATE 中找到匹配模板，过滤条件:", tmpl_filter)
        return

    for tmpl_name, cfg in rows:
        if not cfg:
            print(f"模板 {tmpl_name} 的 CFG_JSON 为空")
            continue
        try:
            data = json.loads(cfg)
        except Exception as e:
            print(f"模板 {tmpl_name} 的 CFG_JSON 解析失败:", e)
            continue
        nodes = collect_task_nodes(data)
        print_nodes(tmpl_name, nodes)


if __name__ == "__main__":
    main()

