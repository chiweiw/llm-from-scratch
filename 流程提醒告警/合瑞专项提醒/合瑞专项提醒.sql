with base_prd as (
    select
        b.PRD_O_CODE as PRD_CODE,
        b.PRD_O_NAME as PRD_NAME,
        b.CDATE,
        b.SALE_PRD_TYP_MKT
    from dw_prd_base_info b
    where b.D_FLAG = '0'
      and b.CDATE = date_format(current_date, '%Y-%m-%d')
      and (
          b.SALE_PRD_TYP_MKT = '7'
          or b.SALE_PRD_TYP_MKT like '%"7"%'
      )
      and b.PRD_O_NAME like '%合瑞%'
),
he_rui_targets as (
    select
        PRD_CODE,
        PRD_NAME,
        CDATE as ESTABLISH_DATE
    from base_prd
),
receivers as (
    select
        u.USER_O_CODE as account,
        u.USER_O_NAME
    from sys_rbac_user u
    where u.USER_O_NAME in ('杨波', '李旭晨')
)
select
    r.account,
    concat(
        '【产品运营管理系统】合瑞产品成立提醒',
        '\n产品名称：',
        coalesce(t.PRD_NAME, '未知'),
        '\n成立日期：',
        date_format(t.ESTABLISH_DATE, '%Y年%m月%d日'),
        '\n',
        '\n产品已于今日成立，请及时确认一次性支付销售手续费的金额明细并核对。'
    ) as CONTENT,
    'SYSTEM' as OPERATOR,
    'SYSTEM' as OPERATOR_ACCOUNT,
    '定时消息提醒' as `TRIGGER`,
    t.PRD_CODE,
    t.PRD_NAME,
    t.ESTABLISH_DATE as SETUP_DATE,
    r.USER_O_NAME as RECEIVER_NAME
from he_rui_targets t
cross join receivers r
