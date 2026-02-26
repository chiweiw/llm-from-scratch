-- 产品定价管理流程提醒告警SQL
-- 需求：生效日期（起）T-3 日 13:00 触发
-- 模板：HXLC_产品定价管理-敏捷小组、HXLC_产品定价管理-部门审批、HXLC_产品定价管理-公司领导审批

with tmpl as (
    select t.R_ID,
           t.TMPL_NAME
    from SYS_WF_TEMPLATE t
    where t.TMPL_NAME in (
        'HXLC_产品定价管理-敏捷小组',
        'HXLC_产品定价管理-部门审批',
        'HXLC_产品定价管理-公司领导审批'
    )
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
price_node_cfg as (
    -- HXLC_产品定价管理-敏捷小组
    -- 场景1: 提交部门审批前
    select 'HXLC_产品定价管理-敏捷小组' as TMPL_NAME,
           '提交需求' as NODE_NAME,
           'PRE_DEPT_APPR' as SCENE_CODE,
           '流程未过提交部门审批节点' as SCENE_NAME
    union all select 'HXLC_产品定价管理-敏捷小组',
                     '复核需求',
                     'PRE_DEPT_APPR',
                     '流程未过提交部门审批节点'
    -- 场景2: 提交审批中
    union all select 'HXLC_产品定价管理-敏捷小组',
                     '提交部门审批',
                     'WAIT_SUBMIT',
                     '流程在提交部门审批节点或流程在提交公司审批节点'
    -- 场景3: 其他情况
    union all select 'HXLC_产品定价管理-敏捷小组',
                     '消费者权益保护岗',
                     'OTHERS',
                     '其他情况'
    union all select 'HXLC_产品定价管理-敏捷小组',
                     '法律合规部',
                     'OTHERS',
                     '其他情况'
    union all select 'HXLC_产品定价管理-敏捷小组',
                     '固定收益投资部门领导',
                     'OTHERS',
                     '其他情况'
    union all select 'HXLC_产品定价管理-敏捷小组',
                     '多资产投资部领导',
                     'OTHERS',
                     '其他情况'
    union all select 'HXLC_产品定价管理-敏捷小组',
                     '多策略投资部领导',
                     'OTHERS',
                     '其他情况'
    union all select 'HXLC_产品定价管理-敏捷小组',
                     '组合投资部领导',
                     'OTHERS',
                     '其他情况'
    union all select 'HXLC_产品定价管理-敏捷小组',
                     '资产创设部领导',
                     'OTHERS',
                     '其他情况'
    union all select 'HXLC_产品定价管理-敏捷小组',
                     '机构投资部领导',
                     'OTHERS',
                     '其他情况'
    union all select 'HXLC_产品定价管理-敏捷小组',
                     '组织敏捷小组会审议',
                     'OTHERS',
                     '其他情况'
    union all select 'HXLC_产品定价管理-敏捷小组',
                     '确认审议结果',
                     'OTHERS',
                     '其他情况'
    union all select 'HXLC_产品定价管理-敏捷小组',
                     '策略创新部领导',
                     'OTHERS',
                     '其他情况'
    union all select 'HXLC_产品定价管理-敏捷小组',
                     'ESG部门领导',
                     'OTHERS',
                     '其他情况'
    union all select 'HXLC_产品定价管理-敏捷小组',
                     '投资研究部领导',
                     'OTHERS',
                     '其他情况'

    -- HXLC_产品定价管理-部门审批
    union all select 'HXLC_产品定价管理-部门审批',
                     '提交需求',
                     'PRE_DEPT_APPR',
                     '流程未过提交部门审批节点'
    union all select 'HXLC_产品定价管理-部门审批',
                     '复核需求',
                     'PRE_DEPT_APPR',
                     '流程未过提交部门审批节点'
    union all select 'HXLC_产品定价管理-部门审批',
                     '消费者权益保护岗',
                     'OTHERS',
                     '其他情况'
    union all select 'HXLC_产品定价管理-部门审批',
                     '法律合规部',
                     'OTHERS',
                     '其他情况'
    union all select 'HXLC_产品定价管理-部门审批',
                     '固定收益投资部门领导',
                     'OTHERS',
                     '其他情况'
    union all select 'HXLC_产品定价管理-部门审批',
                     '多资产投资部领导',
                     'OTHERS',
                     '其他情况'
    union all select 'HXLC_产品定价管理-部门审批',
                     '多策略投资部领导',
                     'OTHERS',
                     '其他情况'
    union all select 'HXLC_产品定价管理-部门审批',
                     '组合投资部领导',
                     'OTHERS',
                     '其他情况'
    union all select 'HXLC_产品定价管理-部门审批',
                     '资产创设部领导',
                     'OTHERS',
                     '其他情况'
    union all select 'HXLC_产品定价管理-部门审批',
                     '机构投资部领导',
                     'OTHERS',
                     '其他情况'
    union all select 'HXLC_产品定价管理-部门审批',
                     '办结生效',
                     'OTHERS',
                     '其他情况'
    union all select 'HXLC_产品定价管理-部门审批',
                     '策略创新部领导',
                     'OTHERS',
                     '其他情况'
    union all select 'HXLC_产品定价管理-部门审批',
                     'ESG部门领导',
                     'OTHERS',
                     '其他情况'
    union all select 'HXLC_产品定价管理-部门审批',
                     '投资研究部领导',
                     'OTHERS',
                     '其他情况'

    -- HXLC_产品定价管理-公司领导审批
    union all select 'HXLC_产品定价管理-公司领导审批',
                     '提交需求',
                     'PRE_DEPT_APPR',
                     '流程未过提交部门审批节点'
    union all select 'HXLC_产品定价管理-公司领导审批',
                     '复核需求',
                     'PRE_DEPT_APPR',
                     '流程未过提交部门审批节点'
    union all select 'HXLC_产品定价管理-公司领导审批',
                     '提交部门审批',
                     'WAIT_SUBMIT',
                     '流程在提交部门审批节点或流程在提交公司审批节点'
    union all select 'HXLC_产品定价管理-公司领导审批',
                     '提交公司审批',
                     'WAIT_SUBMIT',
                     '流程在提交部门审批节点或流程在提交公司审批节点'
    union all select 'HXLC_产品定价管理-公司领导审批',
                     '产品营销部领导',
                     'OTHERS',
                     '其他情况'
    union all select 'HXLC_产品定价管理-公司领导审批',
                     '固定收益投资部门领导',
                     'OTHERS',
                     '其他情况'
    union all select 'HXLC_产品定价管理-公司领导审批',
                     '多资产投资部领导',
                     'OTHERS',
                     '其他情况'
    union all select 'HXLC_产品定价管理-公司领导审批',
                     '多策略投资部领导',
                     'OTHERS',
                     '其他情况'
    union all select 'HXLC_产品定价管理-公司领导审批',
                     '组合投资部领导',
                     'OTHERS',
                     '其他情况'
    union all select 'HXLC_产品定价管理-公司领导审批',
                     '资产创设部领导',
                     'OTHERS',
                     '其他情况'
    union all select 'HXLC_产品定价管理-公司领导审批',
                     '机构投资部领导',
                     'OTHERS',
                     '其他情况'
    union all select 'HXLC_产品定价管理-公司领导审批',
                     '公司领导审批',
                     'OTHERS',
                     '其他情况'
    union all select 'HXLC_产品定价管理-公司领导审批',
                     '消费者权益保护岗',
                     'OTHERS',
                     '其他情况'
    union all select 'HXLC_产品定价管理-公司领导审批',
                     '法律合规部',
                     'OTHERS',
                     '其他情况'
    union all select 'HXLC_产品定价管理-公司领导审批',
                     '策略创新部领导',
                     'OTHERS',
                     '其他情况'
    union all select 'HXLC_产品定价管理-公司领导审批',
                     'ESG部门领导',
                     'OTHERS',
                     '其他情况'
    union all select 'HXLC_产品定价管理-公司领导审批',
                     '投资研究部领导',
                     'OTHERS',
                     '其他情况'
    union all select 'HXLC_产品定价管理-公司领导审批',
                     '办结生效',
                     'OTHERS',
                     '其他情况'
),
price_prd as (
    select pa.INSTANCE_ID,
           pa.PRD_O_CODE as PRD_CODE,
           coalesce(pa.PRD_O_NAME, yz.PRD_O_NAME) as PRD_NAME,
           coalesce(yz.COLLECT_VDATE, current_date) as COLLECT_VDATE,
           coalesce(d.DEPT_O_NAME, pa.FQBM, yz.FQBM) as PRD_DEPT
    from ODS_PRDYZ_PRICE_ADJUST pa
    join unfinished_instance ui on ui.INSTANCE_ID = pa.INSTANCE_ID
    left join ODS_PRDYZ_BASE_INFO yz
      on yz.PRD_O_CODE = pa.PRD_O_CODE
     and yz.D_FLAG <> '7'
    left join sys_rbac_user u
      on u.USER_O_CODE = ui.CREATE_USER
    left join sys_rbac_depart d
      on u.ORG regexp d.DEPT_O_CODE
    where pa.D_FLAG != 7
),
critical as (
    select
        p.INSTANCE_ID,
        concat(date_sub(p.COLLECT_VDATE, interval 3 day), ' 13:00:00') as CRITICAL_DT
    from price_prd p
),
target_instance as (
    select
        ui.INSTANCE_ID,
        p.PRD_CODE,
        p.PRD_NAME,
        ui.CREATE_TIME as START_TIME,
        datediff(current_date, date(ui.CREATE_TIME)) as DAYS_ELAPSED,
        p.COLLECT_VDATE,
        c.CRITICAL_DT,
        p.PRD_DEPT,
        ui.TMPL_NAME
    from unfinished_instance ui
    join price_prd p on p.INSTANCE_ID = ui.INSTANCE_ID
    join critical c on c.INSTANCE_ID = ui.INSTANCE_ID
    where c.CRITICAL_DT is not null
      and current_timestamp >= c.CRITICAL_DT
),
current_todo as (
    select act.INSTANCE_ID,
           act.ACTIVITY_ID,
           act.TASK_NAME,
           per.PARTAKE_USER as account,
           u.USER_O_NAME as USER_NAME,
           pc.SCENE_CODE,
           pc.SCENE_NAME
    from sys_app_wf_activity act
    join sys_app_wf_act_permission per
      on per.INSTANCE_ID = act.INSTANCE_ID
     and per.ACTIVITY_ID = act.ACTIVITY_ID
    left join sys_rbac_user u
      on u.USER_O_CODE = per.PARTAKE_USER
    left join unfinished_instance ui
      on ui.INSTANCE_ID = act.INSTANCE_ID
    left join price_node_cfg pc
      on pc.NODE_NAME = act.TASK_NAME
     and pc.TMPL_NAME = ui.TMPL_NAME
    where act.ACTIVITY_STATE = '1'
      and (per.STATE = '1' or per.STATE = '2')
      and per.PARTAKE_USER is not null
)
select distinct
    ct.account,
    concat(
        '【产品运营管理系统】紧急流程预警\n',
        '流程类型：【产品定价管理】\n',
        '产品名称：',
        coalesce(ti.PRD_NAME, '未知'),
        '（代码 ',
        coalesce(ti.PRD_CODE, 'N/A'),
        '）\n',
        '发起部门：',
        coalesce(ti.PRD_DEPT, '未知'),
        '\n',
        '当前环节：',
        coalesce(ct.TASK_NAME, '未知'),
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
    ct.USER_NAME,
    ct.SCENE_CODE,
    ct.SCENE_NAME,
    1 as SEND_INITIATOR,
    1 as SEND_INV_SALES,
    0 as SEND_SB_REGISTER,
    0 as SEND_ISSUE_REGISTER,
    case
        when ti.TMPL_NAME = 'HXLC_产品定价管理-部门审批'
             and ct.TASK_NAME like '%部门领导%' then 1
        when ti.TMPL_NAME in ('HXLC_产品定价管理-公司领导审批', 'HXLC_产品定价管理-敏捷小组')
             and ct.SCENE_CODE in ('WAIT_SUBMIT', 'OTHERS') then 1
        else 0
    end as SEND_DISC_INFO,
    case
        when ti.TMPL_NAME = 'HXLC_产品定价管理-部门审批'
             and ct.TASK_NAME like '%部门领导%' then 'ALL_PM'
        when ti.TMPL_NAME = 'HXLC_产品定价管理-公司领导审批'
             and ct.SCENE_CODE = 'WAIT_SUBMIT' then 'ALL_PM'
        when ti.TMPL_NAME = 'HXLC_产品定价管理-公司领导审批'
             and ct.SCENE_CODE = 'OTHERS' then 'FLOW_PM'
        when ti.TMPL_NAME = 'HXLC_产品定价管理-敏捷小组'
             and ct.SCENE_CODE in ('WAIT_SUBMIT', 'OTHERS') then 'ALL_PM'
        else ''
    end as PM_RECEIVER_TYPE,
    ti.CRITICAL_DT,
    concat(
        '当前时间已达到或超过临界时点 ',
        date_format(ti.CRITICAL_DT, '%Y-%m-%d %H:%i:%s'),
        '，流程仍为未完结状态'
    ) as REASON
from target_instance ti
join current_todo ct on ct.INSTANCE_ID = ti.INSTANCE_ID
