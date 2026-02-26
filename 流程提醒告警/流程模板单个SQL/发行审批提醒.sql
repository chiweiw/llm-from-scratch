-- 产品发行审批流程提醒告警SQL
-- 基于 CFG_JSON 中的节点配置生成
-- 更新记录: 
-- 1. 修正表名为 ODS_PRDYZ_BASE_INFO，字段名为 PRD_O_CODE/PRD_O_NAME
-- 2. 修正流程开始时间字段为 BEG_TIME
-- 3. [NEW] 增加业务场景映射 (SCENE_CODE, SCENE_NAME)
-- 
-- 【辅助分析】完整流程节点列表 (基于 1.json 解析)
-- 类型说明: start=开始, task=人工任务, finished=结束状态
-- --------------------------------------------------------------------------------------
-- Node ID                               | Text (Name)          | Type     | 说明
-- --------------------------------------------------------------------------------------
-- daef4c0e-40f6-4dd4-99be-db95dbc4d7b0  | (无)                 | start    | 流程开始
-- 5ca31762-6ca9-4e02-87fb-b85963dfe8b2  | 发起                 | task     | 经办人发起
-- 700f1822-582a-4e4e-b88d-0639b1d4d9bf  | 复核需求             | task     | 经办人/复核人处理
-- 8fd38299-524a-4784-9fba-b1abd549b03c  | 待提交部门领导审批   | task     | 【关键卡点】等待提交给部门领导
-- 12b2feeb-c924-4a31-b806-27d6d15e7331  | 相关部门审批         | task     | 外部部门并联/串联审批
-- df2b12a6-e4c5-4a6b-99aa-ea82029c7e6b  | 待提交公司领导审批   | task     | 【关键卡点】等待提交给公司领导
-- f99fb340-0972-45be-8495-517a4fcc2b61  | 公司领导审批         | task     | 最终领导审批
-- 91dc4ec0-de80-4a61-a4c2-d48b8879278e  | 办结                 | task     | 【注意】这是人工节点，非自动结束
-- eeff8a71-308e-4dfa-9d53-9851d301d863  | 已完成               | finished | 流程真正结束状态
-- --------------------------------------------------------------------------------------

with tmpl as (
    select t.R_ID,
           t.TMPL_NAME
    from SYS_WF_TEMPLATE t
    where t.TMPL_NAME in ('HXLC_产品发行审批', 'HX_产品发行审批')
),
issue_node_cfg as (
    -- HXLC_产品发行审批
    -- 场景1: 流程未过提交部门审批节点
    select 'HXLC_产品发行审批' as TMPL_NAME, '提交需求' as NODE_NAME, 'PRE_DEPT_APPR' as SCENE_CODE, '流程未过提交部门审批节点' as SCENE_NAME
    union all select 'HXLC_产品发行审批', '复核需求', 'PRE_DEPT_APPR', '流程未过提交部门审批节点'
    
    -- 场景2: 流程在提交部门审批节点或流程在提交公司审批节点
    union all select 'HXLC_产品发行审批', '提交部门审批', 'WAIT_SUBMIT', '流程在提交部门审批节点或流程在提交公司审批节点'
    union all select 'HXLC_产品发行审批', '提交公司审批', 'WAIT_SUBMIT', '流程在提交部门审批节点或流程在提交公司审批节点'
    
    -- 场景3: 其他情况
    union all select 'HXLC_产品发行审批', '产品营销部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品发行审批', '固定收益投资部门领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品发行审批', '多资产投资部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品发行审批', '多策略投资部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品发行审批', '组合投资部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品发行审批', '资产创设部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品发行审批', '机构投资部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品发行审批', '运营管理部经办', 'OTHERS', '其他情况'
    union all select 'HXLC_产品发行审批', '运营管理部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品发行审批', '公司领导审批', 'OTHERS', '其他情况'
    union all select 'HXLC_产品发行审批', '策略创新部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品发行审批', '投资研究部领导', 'OTHERS', '其他情况'

    -- HX_产品发行审批
    union all select 'HX_产品发行审批', '发起', 'PRE_DEPT_APPR', '流程未过提交部门审批节点'
    union all select 'HX_产品发行审批', '复核需求', 'PRE_DEPT_APPR', '流程未过提交部门审批节点'
    union all select 'HX_产品发行审批', '待提交部门领导审批', 'WAIT_SUBMIT', '流程在提交部门审批节点或流程在提交公司审批节点'
    union all select 'HX_产品发行审批', '待提交公司领导审批', 'WAIT_SUBMIT', '流程在提交部门审批节点或流程在提交公司审批节点'
    union all select 'HX_产品发行审批', '相关部门审批', 'OTHERS', '其他情况'
    union all select 'HX_产品发行审批', '公司领导审批', 'OTHERS', '其他情况'
    union all select 'HX_产品发行审批', '办结', 'OTHERS', '其他情况'
),
unfinished_instance as (
    select ins.INSTANCE_ID,
           ins.TMPL_ID,
           t.TMPL_NAME,
           ins.CREATE_USER,
           ins.BEG_TIME as CREATE_TIME
    from sys_app_wf_instance ins
    join tmpl t on t.R_ID = ins.TMPL_ID
    where ins.INSTANCE_STATE = '0'
),
current_todo as (
    select act.INSTANCE_ID,
           act.ACTIVITY_ID,
           act.TASK_NAME,
           per.PARTAKE_USER as account,
           u.USER_O_NAME as USER_NAME,
           ic.NODE_NAME,
           ic.SCENE_CODE,
           ic.SCENE_NAME
    from sys_app_wf_activity act
    join sys_app_wf_act_permission per
      on per.INSTANCE_ID = act.INSTANCE_ID
     and per.ACTIVITY_ID = act.ACTIVITY_ID
    left join sys_rbac_user u
      on u.USER_O_CODE = per.PARTAKE_USER
    left join unfinished_instance ui
      on ui.INSTANCE_ID = act.INSTANCE_ID
    left join issue_node_cfg ic
      on ic.NODE_NAME = act.TASK_NAME
     and ic.TMPL_NAME = ui.TMPL_NAME
    where act.ACTIVITY_STATE = '1'
      and (per.STATE = '1' or per.STATE = '2')
      and per.PARTAKE_USER is not null
),
prd_yz as (
    select yz.INSTANCE_ID,
           yz.PRD_O_CODE as PRD_CODE,
           yz.PRD_O_NAME as PRD_NAME,
           yz.COLLECT_VDATE,
           coalesce(d.DEPT_O_NAME, yz.FQBM) as PRD_DEPT
    from ODS_PRDYZ_BASE_INFO yz
    join unfinished_instance ui on ui.INSTANCE_ID = yz.INSTANCE_ID
    left join sys_rbac_user u
      on u.USER_O_CODE = ui.CREATE_USER
    left join sys_rbac_depart d
      on u.ORG regexp d.DEPT_O_CODE
    where yz.D_FLAG <> '7'
),
critical as (
    select
        p.INSTANCE_ID,
        concat(date_sub(p.COLLECT_VDATE, interval 1 day), ' 00:00:00') as CRITICAL_DT
    from prd_yz p
),
target_instance as (
    select
        ui.INSTANCE_ID,
        pyz.PRD_CODE,
        pyz.PRD_NAME,
        pyz.COLLECT_VDATE,
        ui.CREATE_TIME as START_TIME,
        datediff(current_date, date(ui.CREATE_TIME)) as DAYS_ELAPSED,
        c.CRITICAL_DT,
        pyz.PRD_DEPT
    from unfinished_instance ui
    join prd_yz pyz on pyz.INSTANCE_ID = ui.INSTANCE_ID
    join critical c on c.INSTANCE_ID = ui.INSTANCE_ID
    where c.CRITICAL_DT is not null
      and current_timestamp >= c.CRITICAL_DT
)
select distinct
    ct.account,
    concat(
        '【产品运营管理系统】紧急流程预警\n',
        '流程类型：【产品发行审批】\n',
        '产品名称：',
        coalesce(ti.PRD_NAME, '未知'),
        '（代码 ',
        coalesce(ti.PRD_CODE, 'N/A'),
        '）\n',
        '发起部门：',
        coalesce(ti.PRD_DEPT, '未知'),
        '\n',
        '当前环节：',
        coalesce(ct.NODE_NAME, ct.TASK_NAME),
        '\n',
        '待办人员：',
        coalesce(ct.USER_NAME, ct.account),
        '\n',
        '业务生效日：',
        date_format(ti.COLLECT_VDATE, '%Y-%m-%d'),
        '\n',
        '当前流程已处于业务临界时点，请及时关注进度！'
    ) as CONTENT,
    'SYSTEM' as OPERATOR,
    'SYSTEM' as OPERATOR_ACCOUNT,
    '定时消息提醒' as `TRIGGER`,
    ti.PRD_CODE,
    ti.PRD_NAME,
    ti.INSTANCE_ID,
    ct.ACTIVITY_ID,
    ct.TASK_NAME,
    ct.NODE_NAME,
    ct.SCENE_NAME,
    ct.USER_NAME,
    case
        when ct.SCENE_CODE in ('PRE_DEPT_APPR', 'WAIT_SUBMIT') then '经办人/产品经理'
        when ct.SCENE_CODE = 'OTHERS' then '审批领导/相关部门'
        else '待办人'
    end as RECEIVER_ROLE_DESC,
    1 as SEND_INITIATOR,
    1 as SEND_INV_SALES,
    case
        when ct.SCENE_CODE in ('WAIT_SUBMIT', 'OTHERS') then 1
        else 0
    end as SEND_SB_REGISTER,
    0 as SEND_ISSUE_REGISTER,
    case
        when ct.SCENE_CODE in ('WAIT_SUBMIT', 'OTHERS') then 1
        else 0
    end as SEND_DISC_INFO,
    case
        when ct.SCENE_CODE = 'WAIT_SUBMIT' then 'ALL_PM'
        when ct.SCENE_CODE = 'OTHERS' then 'FLOW_PM'
        else ''
    end as PM_RECEIVER_TYPE,
    ti.START_TIME,
    ti.DAYS_ELAPSED,
    ti.CRITICAL_DT,
    concat(
        '当前时间已达到或超过临界时点 ',
        date_format(ti.CRITICAL_DT, '%Y-%m-%d %H:%i:%s'),
        '，流程仍为未完结状态'
    ) as REASON
from target_instance ti
join current_todo ct on ct.INSTANCE_ID = ti.INSTANCE_ID
