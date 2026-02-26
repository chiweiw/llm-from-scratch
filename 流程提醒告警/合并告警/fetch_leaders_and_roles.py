"""
脚本功能：
1. 根据用户提供的部门领导名单，查询 sys_rbac_user 表获取对应的 USER_O_CODE。
   * 采用全量查询 + Python 过滤方式，规避 SQL 参数化执行的奇怪问题。
2. 查询 sys_rbac_role 表，尝试寻找"产品管理"相关的角色 SQL。
"""

import pymysql
import pandas as pd
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

# 部门-领导 映射数据
leader_data = [
    ("固定收益投资部", "杨牧", "贾志敏"),
    ("多资产投资部", "倪春", "贾志敏"),
    ("多策略投资部", "王莎", "贾志敏"),
    ("组合投资部", "张晓华", "贾志敏"),
    ("资产创设部", "刘潇", "贾志敏"),
    ("策略创新部", "杨杰", "贾志敏"),
    ("产品营销部", "张晓华2", "贾志敏"),
    ("运营管理部", "张怀珍", "毛伟"),
    ("客户体验部", "严律", "-"),
    ("投资研究部", "杨勤宇", "-"),
    ("风险管理部", "-", "-"),
    ("法律合规部", "李峰", "-"),
    ("金融科技部", "-", "-"),
    ("机构投资部", "杜建智", "贾志敏"),
    ("资金财务部", "-", "-"),
    ("ESG事业部", "王汉魁", "-"),
]

# 提取所有不重复的姓名 (排除 '-')
all_names = set()
for _, l1, l2 in leader_data:
    if l1 and l1 != "-":
        all_names.add(l1.strip())
    if l2 and l2 != "-":
        all_names.add(l2.strip())


def run_query():
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
            # 1. 全量查询用户表
            print("正在查询全量人员表...")
            sql_user = "SELECT USER_O_NAME, USER_O_CODE, ACCOUNT FROM sys_rbac_user"
            cursor.execute(sql_user)
            users = cursor.fetchall()
            df_all = pd.DataFrame(users, columns=["姓名", "人员代码", "登录名"])

            print(f"总用户数: {len(df_all)}")

            # Python端过滤
            # 处理 '张晓华2' -> '张晓华'
            df_matched = df_all[df_all["姓名"].isin(all_names)]

            # 额外处理此时包含 '张晓华' 但 all_names如果有 '张晓华2' 的情况
            if "张晓华2" in all_names:
                df_zh = df_all[df_all["姓名"].str.startswith("张晓华")]
                df_matched = pd.concat([df_matched, df_zh])

            df_matched = df_matched.drop_duplicates()

            print("-" * 30)
            print("匹配到的用户:")
            print(df_matched)
            print("-" * 30)

            # 2. 查询产品管理相关角色
            print("\n正在查询产品管理相关角色...")
            sql_role = "SELECT ROLE_O_NAME, ROLE_O_CODE FROM sys_rbac_role WHERE ROLE_O_NAME LIKE '%产品%' OR ROLE_O_NAME LIKE '%经理%'"
            cursor.execute(sql_role)
            roles = cursor.fetchall()
            df_roles = pd.DataFrame(roles, columns=["角色名称", "角色代码"])
            print(df_roles)

            # 3. 输出 SQL 语句建议
            print("\n\n=== 构建的领导集合 SQL 片段 (Map) ===")
            print("-- 领导名称 -> 代码映射")

            # 建立 name -> code 字典
            # 如果重名，这里只能取一个，或者需要更复杂的逻辑。暂时取第一个非空的。
            name_map = {}
            for index, row in df_matched.iterrows():
                if row["姓名"] not in name_map:
                    name_map[row["姓名"]] = row["人员代码"]

            # 映射修正
            if "张晓华" in name_map and "张晓华2" not in name_map:
                name_map["张晓华2"] = name_map["张晓华"]

            for dept, l1, l2 in leader_data:
                codes = []

                # 领导1
                l1_clean = l1.strip()
                if l1_clean in name_map:
                    codes.append(f"'{name_map[l1_clean]}'/*{l1_clean}*/")
                elif l1_clean != "-":
                    codes.append(f"NULL/*{l1_clean}未找到*/")

                # 领导2
                l2_clean = l2.strip()
                if l2_clean in name_map:
                    codes.append(f"'{name_map[l2_clean]}'/*{l2_clean}*/")
                elif l2_clean != "-":
                    codes.append(f"NULL/*{l2_clean}未找到*/")

                print(f"-- {dept}: {', '.join(codes)}")

    finally:
        conn.close()


if __name__ == "__main__":
    run_query()
