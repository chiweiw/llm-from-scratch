-- 产品发行登记流程提醒告警SQL
-- 需求：募集开始日 T-1 日 13:00 触发
-- 模板：HXLC_产品发行登记

with tmpl as (
    select t.R_ID,
           t.TMPL_NAME
    from SYS_WF_TEMPLATE t
    where t.TMPL_NAME in ('HXLC_产品发行登记')
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
issue_reg_prd as (
    select r.INSTANCE_ID,
           yz.PRD_O_CODE as PRD_CODE,
           yz.PRD_O_NAME as PRD_NAME,
           yz.COLLECT_VDATE,
           yz.FQBM as PRD_DEPT
    from ODS_PRDYZ_ISSUE_REG r
    join unfinished_instance ui on ui.INSTANCE_ID = r.INSTANCE_ID
    join ODS_PRDYZ_BASE_INFO yz on yz.PRD_O_CODE = r.PRD_O_CODE
    where r.D_FLAG <> '7'
      and yz.D_FLAG <> '7'
),
critical as (
    select
        p.INSTANCE_ID,
        concat(date_sub(p.COLLECT_VDATE, interval 1 day), ' 13:00:00') as CRITICAL_DT
    from issue_reg_prd p
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
        p.PRD_DEPT
    from unfinished_instance ui
    join issue_reg_prd p on p.INSTANCE_ID = ui.INSTANCE_ID
    join critical c on c.INSTANCE_ID = ui.INSTANCE_ID
    where c.CRITICAL_DT is not null
      and current_timestamp >= c.CRITICAL_DT
),
current_todo as (
    select act.INSTANCE_ID,
           act.ACTIVITY_ID,
           act.TASK_NAME,
           per.PARTAKE_USER as account,
           u.USER_O_NAME as USER_NAME
    from sys_app_wf_activity act
    join sys_app_wf_act_permission per
      on per.INSTANCE_ID = act.INSTANCE_ID
     and per.ACTIVITY_ID = act.ACTIVITY_ID
    left join sys_rbac_user u
      on u.USER_O_CODE = per.PARTAKE_USER
    where act.ACTIVITY_STATE = '1'
      and (per.STATE = '1' or per.STATE = '2')
      and per.PARTAKE_USER is not null
)
select distinct
    ct.account,
    concat(
        '【产品运营管理系统】紧急流程预警\n',
        '流程类型：【产品发行登记】\n',
        '产品名称：',
        ti.PRD_NAME,
        '（代码 ',
        ti.PRD_CODE,
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
    '待办人' as ROLE_TYPE,
    0 as SEND_INITIATOR,
    0 as SEND_INV_SALES,
    0 as SEND_SB_REGISTER,
    0 as SEND_ISSUE_REGISTER,
    0 as SEND_DISC_INFO,
    '' as PM_RECEIVER_TYPE,
    ti.CRITICAL_DT,
    concat(
        '当前时间已达到或超过临界时点 ',
        date_format(ti.CRITICAL_DT, '%Y-%m-%d %H:%i:%s'),
        '，流程仍为未完结状态'
    ) as REASON
from target_instance ti
join current_todo ct on ct.INSTANCE_ID = ti.INSTANCE_ID
