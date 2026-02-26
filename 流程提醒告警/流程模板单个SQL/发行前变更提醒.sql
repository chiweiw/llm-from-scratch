with tmpl as (
    select t.R_ID,
           t.TMPL_NAME
    from SYS_WF_TEMPLATE t
    where t.TMPL_NAME in ('HXLC_产品发行前变更')
),
prechange_node_cfg as (
    -- 场景1: 提交部门审批前
    select '提交需求' as NODE_NAME, 'PRE_DEPT_APPR' as SCENE_CODE, '流程未过提交部门审批节点' as SCENE_NAME
    union all select '复核需求', 'PRE_DEPT_APPR', '流程未过提交部门审批节点'
    -- 场景2: 提交审批中
    union all select '提交部门审批', 'WAIT_SUBMIT', '流程在提交部门审批节点或流程在提交公司审批节点'
    union all select '提交公司审批', 'WAIT_SUBMIT', '流程在提交部门审批节点或流程在提交公司审批节点'
    -- 场景3: 其他情况
    union all select '产品营销部领导', 'OTHERS', '其他情况'
    union all select '固定收益投资部门领导', 'OTHERS', '其他情况'
    union all select '多资产投资部领导', 'OTHERS', '其他情况'
    union all select '多策略投资部领导', 'OTHERS', '其他情况'
    union all select '组合投资部领导', 'OTHERS', '其他情况'
    union all select '资产创设部领导', 'OTHERS', '其他情况'
    union all select '机构投资部领导', 'OTHERS', '其他情况'
    union all select '运营管理部经办', 'OTHERS', '其他情况'
    union all select '运营管理部领导', 'OTHERS', '其他情况'
    union all select '公司领导审批', 'OTHERS', '其他情况'
    union all select '策略创新部领导', 'OTHERS', '其他情况'
    union all select '投资研究部领导', 'OTHERS', '其他情况'
),
unfinished_instance as (
    select ins.INSTANCE_ID,
           ins.TMPL_ID,
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
           coalesce(nc.NODE_NAME, act.TASK_NAME) as NODE_NAME,
           nc.SCENE_CODE,
           nc.SCENE_NAME
    from sys_app_wf_activity act
    join sys_app_wf_act_permission per
      on per.INSTANCE_ID = act.INSTANCE_ID
     and per.ACTIVITY_ID = act.ACTIVITY_ID
    left join sys_rbac_user u
      on u.USER_O_CODE = per.PARTAKE_USER
    left join prechange_node_cfg nc
      on nc.NODE_NAME = act.TASK_NAME
    where act.ACTIVITY_STATE = '1'
      and (per.STATE = '1' or per.STATE = '2')
      and per.PARTAKE_USER is not null
),
prd_prechange as (
    select yz.INSTANCE_ID,
           yz.PRD_O_CODE as PRD_CODE,
           yz.PRD_O_NAME as PRD_NAME,
           yz.COLLECT_VDATE,
           yz.FQBM as PRD_DEPT
    from ODS_PRDYZ_BASE_INFO yz
    join unfinished_instance ui on ui.INSTANCE_ID = yz.INSTANCE_ID
    where yz.D_FLAG <> '7'
),
critical as (
    select
        p.INSTANCE_ID,
        concat(date_sub(p.COLLECT_VDATE, interval 2 day), ' 13:00:00') as CRITICAL_DT
    from prd_prechange p
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
    join prd_prechange p on p.INSTANCE_ID = ui.INSTANCE_ID
    join critical c on c.INSTANCE_ID = ui.INSTANCE_ID
    where c.CRITICAL_DT is not null
      and current_timestamp >= c.CRITICAL_DT
)
select distinct
    ct.account,
    concat(
        '【产品运营管理系统】紧急流程预警\n',
        '流程类型：【产品发行前变更】\n',
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
    ct.SCENE_CODE,
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
