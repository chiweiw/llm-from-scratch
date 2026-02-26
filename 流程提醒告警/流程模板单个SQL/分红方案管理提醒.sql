with tmpl as (
    select t.R_ID,
           t.TMPL_NAME
    from SYS_WF_TEMPLATE t
    where t.TMPL_NAME in (
        'HXLC_产品分红方案_定期分红',
        'HXLC_产品分红方案_不定期分红'
    )
),
dividend_node_cfg as (
    select 'HXLC_产品分红方案_不定期分红' as TMPL_NAME, '提交需求' as NODE_NAME, 'PRE_DEPT_APPR' as SCENE_CODE, '流程未过提交部门审批节点' as SCENE_NAME
    union all select 'HXLC_产品分红方案_不定期分红', '复核需求', 'PRE_DEPT_APPR', '流程未过提交部门审批节点'
    union all select 'HXLC_产品分红方案_不定期分红', '部门领导审批', 'OTHERS', '其他情况'
    union all select 'HXLC_产品分红方案_不定期分红', '提交公司审批', 'WAIT_SUBMIT', '流程在提交部门审批节点或流程在提交公司审批节点'
    union all select 'HXLC_产品分红方案_不定期分红', '公司领导审批', 'OTHERS', '其他情况'
    union all select 'HXLC_产品分红方案_不定期分红', '办结', 'OTHERS', '其他情况'
    union all select 'HXLC_产品分红方案_定期分红', '制定分红方案', 'PRE_DEPT_APPR', '流程未过提交部门审批节点'
    union all select 'HXLC_产品分红方案_定期分红', '复核分红方案', 'PRE_DEPT_APPR', '流程未过提交部门审批节点'
    union all select 'HXLC_产品分红方案_定期分红', '部门领导审批', 'OTHERS', '其他情况'
    union all select 'HXLC_产品分红方案_定期分红', '办结', 'OTHERS', '其他情况'
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
           dnc.NODE_NAME,
           dnc.SCENE_CODE,
           dnc.SCENE_NAME
    from sys_app_wf_activity act
    join sys_app_wf_act_permission per
      on per.INSTANCE_ID = act.INSTANCE_ID
     and per.ACTIVITY_ID = act.ACTIVITY_ID
    left join sys_rbac_user u
      on u.USER_O_CODE = per.PARTAKE_USER
    join unfinished_instance ui
      on ui.INSTANCE_ID = act.INSTANCE_ID
    left join dividend_node_cfg dnc
      on dnc.NODE_NAME = act.TASK_NAME
     and dnc.TMPL_NAME = ui.TMPL_NAME
    where act.ACTIVITY_STATE = '1'
      and (per.STATE = '1' or per.STATE = '2')
      and per.PARTAKE_USER is not null
),
dividend_prd as (
    select ia.INSTANCE_ID,
           ia.PRD_O_CODE as PRD_CODE,
           coalesce(yz.PRD_O_NAME, ia.PRD_O_CODE) as PRD_NAME,
           date(ia.DIVIDEND_DEL_DATE) as EX_DATE,
           d.DEPT_O_NAME as PRD_DEPT
    from ODS_PRDYZ_INT_ASSIGN ia
    join unfinished_instance ui on ui.INSTANCE_ID = ia.INSTANCE_ID
    left join ODS_PRDYZ_BASE_INFO yz
      on yz.PRD_O_CODE = ia.PRD_O_CODE
     and yz.D_FLAG <> '7'
    left join sys_app_wf_instance ins
      on ins.INSTANCE_ID = ia.INSTANCE_ID
    left join sys_rbac_user u
      on u.USER_O_CODE = ins.CREATE_USER
    left join sys_rbac_depart d
      on u.ORG regexp d.DEPT_O_CODE
    where ia.D_FLAG <> '7'
),
critical as (
    select
        p.INSTANCE_ID,
        concat(date_sub(p.EX_DATE, interval 3 day), ' 13:00:00') as CRITICAL_DT
    from dividend_prd p
),
target_instance as (
    select
        ui.INSTANCE_ID,
        p.PRD_CODE,
        p.PRD_NAME,
        ui.CREATE_TIME as START_TIME,
        datediff(current_date, date(ui.CREATE_TIME)) as DAYS_ELAPSED,
        p.EX_DATE,
        c.CRITICAL_DT,
        p.PRD_DEPT,
        ui.TMPL_NAME
    from unfinished_instance ui
    join dividend_prd p on p.INSTANCE_ID = ui.INSTANCE_ID
    join critical c on c.INSTANCE_ID = ui.INSTANCE_ID
    where c.CRITICAL_DT is not null
      and current_timestamp >= c.CRITICAL_DT
)
select distinct
    ct.account,
    concat(
        '【产品运营管理系统】紧急流程预警\n',
        '流程类型：【分红方案管理】\n',
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
        date_format(ti.EX_DATE, '%Y-%m-%d'),
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
    ct.SCENE_CODE,
    ct.SCENE_NAME,
    ct.USER_NAME,
    case
        when ct.account is not null then '流程中产品经理'
        else '待办人'
    end as RECEIVER_ROLE_DESC,
    1 as SEND_INITIATOR,
    1 as SEND_INV_SALES,
    0 as SEND_SB_REGISTER,
    0 as SEND_ISSUE_REGISTER,
    case
        when ti.TMPL_NAME = 'HXLC_产品分红方案_不定期分红' then 1
        else 0
    end as SEND_DISC_INFO,
    case
        when ti.TMPL_NAME = 'HXLC_产品分红方案_定期分红' then 'FLOW_PM'
        when ti.TMPL_NAME = 'HXLC_产品分红方案_不定期分红'
             and ct.TASK_NAME like '%公司领导审批%' then 'ALL_PM'
        when ti.TMPL_NAME = 'HXLC_产品分红方案_不定期分红' then 'FLOW_PM'
        else ''
    end as PM_RECEIVER_TYPE,
    ti.START_TIME,
    ti.DAYS_ELAPSED,
    ti.CRITICAL_DT,
    concat(
        '当前时间已达到或超过除权除息日前 T-3 日 13:00:00 的临界时点，流程仍为未完结状态'
    ) as REASON
from target_instance ti
join current_todo ct on ct.INSTANCE_ID = ti.INSTANCE_ID
