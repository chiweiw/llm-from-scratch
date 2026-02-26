with tmpl as (
    select t.R_ID,
           t.TMPL_NAME
    from SYS_WF_TEMPLATE t
    where t.TMPL_NAME in (
        'HXLC_产品申报登记_维护',
        'HXLC_产品申报登记_封闭式',
        'HXLC_产品申报登记'
    )
),
unfinished_instance as (
    select ins.INSTANCE_ID,
           ins.TMPL_ID,
           ins.CREATE_USER
    from sys_app_wf_instance ins
    join tmpl t on t.R_ID = ins.TMPL_ID
    where ins.INSTANCE_STATE = '0'
),
sb as (
    select info.INSTANCE_ID,
           coalesce(info.HN_BSM, info.PRD_O_CODE) as PRD_CODE,
           info.REG_APPLICATION_DATE as RAISE_START_DATE
    from ods_prdtq_cpsbxxgl_info info
    join unfinished_instance ui on ui.INSTANCE_ID = info.INSTANCE_ID
    where info.D_FLAG <> '7'
),
prd as (
    select s.INSTANCE_ID,
           s.PRD_CODE,
           b.PRD_O_NAME,
           b.PRD_COLL_MODE,
           s.RAISE_START_DATE,
           b.FIRST_INVEST_MANAGER,
           b.SALES_MANAGER
    from sb s
    join dw_prd_base_info b
      on b.PRD_O_CODE = s.PRD_CODE
     and b.D_FLAG = '0'
),
critical as (
    select
        p.INSTANCE_ID,
        case
            when p.PRD_COLL_MODE = '1' then (
                select concat(d.N_DAY, ' 00:00:00')
                from dw_dt_date d
                where d.DT_TYPE = 'A'
                  and d.N_DAY <= p.RAISE_START_DATE
                order by d.N_DAY desc
                limit 10, 1
            )
            when p.PRD_COLL_MODE = '2' then (
                select concat(d.N_DAY, ' 00:00:00')
                from dw_dt_date d
                where d.DT_TYPE = 'A'
                  and d.N_DAY <= p.RAISE_START_DATE
                order by d.N_DAY desc
                limit 2, 1
            )
            else null
        end as CRITICAL_DT
    from prd p
),
instance_dept as (
    select ui.INSTANCE_ID,
           d.DEPT_O_NAME as PRD_DEPT
    from unfinished_instance ui
    left join sys_rbac_user u
      on u.USER_O_CODE = ui.CREATE_USER
    left join sys_rbac_depart d
      on u.ORG regexp d.DEPT_O_CODE
),
target_instance as (
    select
        p.INSTANCE_ID,
        p.PRD_CODE,
        p.PRD_O_NAME,
        p.PRD_COLL_MODE,
        p.RAISE_START_DATE,
        c.CRITICAL_DT,
        p.FIRST_INVEST_MANAGER,
        p.SALES_MANAGER,
        id.PRD_DEPT
    from prd p
    join critical c on c.INSTANCE_ID = p.INSTANCE_ID
    left join instance_dept id on id.INSTANCE_ID = p.INSTANCE_ID
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
        '流程类型：【产品申报登记】\n',
        '产品名称：',
        ti.PRD_O_NAME,
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
        date_format(ti.RAISE_START_DATE, '%Y-%m-%d'),
        '\n',
        '当前流程已处于业务临界时点，请及时关注进度！'
    ) as CONTENT,
    'SYSTEM' as OPERATOR,
    'SYSTEM' as OPERATOR_ACCOUNT,
    '定时消息提醒' as `TRIGGER`,
    ti.PRD_CODE,
    ti.PRD_O_NAME,
    ti.PRD_COLL_MODE,
    case
        when ti.PRD_COLL_MODE = '1' then '公募'
        when ti.PRD_COLL_MODE = '2' then '私募'
        else ''
    end as PRD_COLL_MODE_NAME,
    ti.INSTANCE_ID,
    ct.ACTIVITY_ID,
    ct.TASK_NAME,
    ct.USER_NAME,
    '待办人' as ROLE_TYPE,
    case
        when ct.account = ti.FIRST_INVEST_MANAGER and ct.account = ti.SALES_MANAGER then '投资经理,销售经理'
        when ct.account = ti.FIRST_INVEST_MANAGER then '投资经理'
        when ct.account = ti.SALES_MANAGER then '销售经理'
        else ''
    end as ACCOUNT_ROLE,
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
