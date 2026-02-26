-- 合并流程提醒告警 SQL
-- 功能: 整合[产品申报登记]、[产品发行审批]、[产品发行前变更]、[产品暂停发行]、[产品发行登记]、[产品定价管理]、[增发份额]、[销售相关信息变更]、[开放计划调整]、[分红方案管理]、[产品到期日变更]的流程监控逻辑
-- 
-- 逻辑分支说明:
-- A. 分支逻辑 A: 产品申报登记 (Sb) 
--    基于临界时间 (Critical DT) 触发，关注 ods_prdtq_cpsbxxgl_info
-- B. 分支逻辑 B: 产品发行审批 (Issue)
--    基于募集起始日 T-1 日零点的临界时间触发，关注 ODS_PRDYZ_BASE_INFO，
--    同时结合 CFG_JSON 场景化节点映射进行接收人划分
-- C. 分支逻辑 C: 产品发行前变更 (PreChange)
--    基于募集开始日 T-2 日 13:00 的临界时间触发，关注 ODS_PRDYZ_BASE_INFO
-- D. 分支逻辑 D: 产品暂停发行 (Suspend)
--    基于募集起始日 T-2 日 13:00 的临界时间触发，关注 ODS_PRDYZ_ISSUE_SUSPEND / ODS_PRDYZ_BASE_INFO
-- E. 分支逻辑 E: 产品发行登记 (IssueReg)
--    基于募集开始日 T-1 日 13:00 的临界时间触发，关注 ODS_PRDYZ_ISSUE_REG / ODS_PRDYZ_BASE_INFO
-- F. 分支逻辑 F: 产品定价管理 (PriceAdjust)
--    基于生效日期（起）T-3 日 13:00 的临界时间触发，关注 ODS_PRDYZ_PRICE_ADJUST / ODS_PRDYZ_BASE_INFO
-- G. 分支逻辑 G: 增发份额 (AddIssue)
--    基于增发计划日期 T-3 日 13:00 的临界时间触发，关注 ODS_PRDYZ_ADD_ISSUE_SON / ODS_PRDYZ_ISSUE_INFO
-- H. 分支逻辑 H: 销售信息变更 (SalesChange)
--    基于生效日期 T-3 日 13:00 的临界时间触发，关注 ODS_PRDYZ_SALES_CHANGE
-- I. 分支逻辑 I: 开放计划调整 (OpenPlanAdjust)
--    基于开放日期 T-3 日 13:00 的临界时间触发，关注 ODS_PRDYZ_OPEN_PLAN_ADJUST
-- J. 分支逻辑 J: 分红方案管理 (DividendPlan)
--    基于除权除息日 T-3 日 13:00 的临界时间触发，关注 ODS_PRDYZ_INT_ASSIGN / SYS_APP_WF_INSTANCE
-- K. 分支逻辑 K: 产品到期日变更 (MaturityChange)
--    基于调整后到期日 T-3 日 13:00 的临界时间触发，关注 ODS_PRDJS_END

with 
-- ===============================================================
-- 基础数据准备 (Common)
-- ===============================================================
-- 1. 定义所有相关模板
tmpl as (
    select t.R_ID,
           t.TMPL_NAME
    from SYS_WF_TEMPLATE t
    where t.TMPL_NAME in (
        'HXLC_产品申报登记_维护',
        'HXLC_产品申报登记_封闭式',
        'HXLC_产品申报登记',
        'HXLC_产品发行审批',
        'HXLC_产品发行前变更',
        'HXLC_产品暂停发行管理',
        'HXLC_产品发行登记',
        'HXLC_产品定价管理-敏捷小组',
        'HXLC_产品定价管理-部门审批',
        'HXLC_产品定价管理-公司领导审批',
        'HXLC_增发份额',
        'HXLC_销售相关信息变更',
        'HXLC_开放计划调整',
        'HXLC_产品分红方案_定期分红',
        'HXLC_产品分红方案_不定期分红',
        'HXLC_产品到期日变更'
    )
),
-- 3. 发行审批 / 产品发行前变更 / 产品定价管理 节点配置 (按模板精确匹配)
--    发行审批: HXLC_产品发行审批
--    发行前变更: HXLC_产品发行前变更
--    产品定价管理: HXLC_产品定价管理-敏捷小组, HXLC_产品定价管理-部门审批, HXLC_产品定价管理-公司领导审批
issue_node_cfg as (
    select 'HXLC_产品发行审批' as TMPL_NAME, '提交需求' as NODE_NAME, 'PRE_DEPT_APPR' as SCENE_CODE, '流程未过提交部门审批节点' as SCENE_NAME
    union all select 'HXLC_产品发行审批', '复核需求', 'PRE_DEPT_APPR', '流程未过提交部门审批节点'
    union all select 'HXLC_产品发行审批', '提交部门审批', 'WAIT_SUBMIT', '流程在提交部门审批节点或流程在提交公司审批节点'
    union all select 'HXLC_产品发行审批', '提交公司审批', 'WAIT_SUBMIT', '流程在提交部门审批节点或流程在提交公司审批节点'
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

    -- 产品发行前变更 (HXLC_产品发行前变更)
    union all select 'HXLC_产品发行前变更', '提交需求', 'PRE_DEPT_APPR', '流程未过提交部门审批节点'
    union all select 'HXLC_产品发行前变更', '复核需求', 'PRE_DEPT_APPR', '流程未过提交部门审批节点'
    union all select 'HXLC_产品发行前变更', '提交部门审批', 'WAIT_SUBMIT', '流程在提交部门审批节点或流程在提交公司审批节点'
    union all select 'HXLC_产品发行前变更', '提交公司审批', 'WAIT_SUBMIT', '流程在提交部门审批节点或流程在提交公司审批节点'
    union all select 'HXLC_产品发行前变更', '产品营销部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品发行前变更', '固定收益投资部门领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品发行前变更', '多资产投资部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品发行前变更', '多策略投资部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品发行前变更', '组合投资部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品发行前变更', '资产创设部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品发行前变更', '机构投资部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品发行前变更', '运营管理部经办', 'OTHERS', '其他情况'
    union all select 'HXLC_产品发行前变更', '运营管理部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品发行前变更', '公司领导审批', 'OTHERS', '其他情况'
    union all select 'HXLC_产品发行前变更', '策略创新部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品发行前变更', '投资研究部领导', 'OTHERS', '其他情况'

    union all select 'HXLC_增发份额', '提交需求', 'PRE_DEPT_APPR', '流程未过提交部门审批节点'
    union all select 'HXLC_增发份额', '复核需求', 'PRE_DEPT_APPR', '流程未过提交部门审批节点'
    union all select 'HXLC_增发份额', '提交部门审批', 'WAIT_SUBMIT', '流程在提交部门审批节点或流程在提交公司审批节点'
    union all select 'HXLC_增发份额', '提交公司审批', 'WAIT_SUBMIT', '流程在提交部门审批节点或流程在提交公司审批节点'
    union all select 'HXLC_增发份额', '待提交部门领导审批', 'WAIT_SUBMIT', '流程在提交部门审批节点或流程在提交公司审批节点'
    union all select 'HXLC_增发份额', '待提交公司领导审批', 'WAIT_SUBMIT', '流程在提交部门审批节点或流程在提交公司审批节点'
    union all select 'HXLC_增发份额', '消费者权益保护岗', 'OTHERS', '其他情况'
    union all select 'HXLC_增发份额', '法律合规部', 'OTHERS', '其他情况'
    union all select 'HXLC_增发份额', '产品营销部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_增发份额', '固定收益投资部门领导', 'OTHERS', '其他情况'
    union all select 'HXLC_增发份额', '多资产投资部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_增发份额', '多策略投资部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_增发份额', '组合投资部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_增发份额', '资产创设部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_增发份额', '机构投资部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_增发份额', '公司领导审批', 'OTHERS', '其他情况'
    union all select 'HXLC_增发份额', '策略创新部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_增发份额', '投资研究部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_增发份额', '办结', 'OTHERS', '其他情况'

    union all select 'HXLC_销售相关信息变更', '提交需求', 'PRE_DEPT_APPR', '流程未过提交部门审批节点'
    union all select 'HXLC_销售相关信息变更', '复核需求', 'PRE_DEPT_APPR', '流程未过提交部门审批节点'
    union all select 'HXLC_销售相关信息变更', '投资经理复核', 'PRE_DEPT_APPR', '流程未过提交部门审批节点'
    union all select 'HXLC_销售相关信息变更', '销售经理复核', 'PRE_DEPT_APPR', '流程未过提交部门审批节点'
    union all select 'HXLC_销售相关信息变更', '提交部门审批', 'WAIT_SUBMIT', '流程在提交部门审批节点或流程在提交公司审批节点'
    union all select 'HXLC_销售相关信息变更', '部门领导审批', 'WAIT_SUBMIT', '流程在提交部门审批节点或流程在提交公司审批节点'
    union all select 'HXLC_销售相关信息变更', '运营部门经办', 'OTHERS', '其他情况'
    union all select 'HXLC_销售相关信息变更', '运营管理部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_销售相关信息变更', '办结', 'OTHERS', '其他情况'

    -- 开放计划调整 (HXLC_开放计划调整)
    union all select 'HXLC_开放计划调整', '提交需求', 'PRE_DEPT_APPR', '流程未过提交部门审批节点'
    union all select 'HXLC_开放计划调整', '复核需求', 'PRE_DEPT_APPR', '流程未过提交部门审批节点'
    union all select 'HXLC_开放计划调整', '草稿', 'PRE_DEPT_APPR', '流程未过提交部门审批节点'
    union all select 'HXLC_开放计划调整', '提交部门审批', 'WAIT_SUBMIT', '流程在提交部门审批节点或流程在提交公司审批节点'
    union all select 'HXLC_开放计划调整', '待提交部门领导审批', 'WAIT_SUBMIT', '流程在提交部门审批节点或流程在提交公司审批节点'
    union all select 'HXLC_开放计划调整', '待提交公司领导审批', 'WAIT_SUBMIT', '流程在提交部门审批节点或流程在提交公司审批节点'
    union all select 'HXLC_开放计划调整', '产品营销部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_开放计划调整', '固定收益投资部门领导', 'OTHERS', '其他情况'
    union all select 'HXLC_开放计划调整', '多资产投资部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_开放计划调整', '组合投资部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_开放计划调整', '机构投资部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_开放计划调整', '法律合规部', 'OTHERS', '其他情况'
    union all select 'HXLC_开放计划调整', '消费者权益保护岗', 'OTHERS', '其他情况'
    union all select 'HXLC_开放计划调整', '公司领导审批', 'OTHERS', '其他情况'
    union all select 'HXLC_开放计划调整', '办结', 'OTHERS', '其他情况'

    -- 产品到期日变更 (HXLC_产品到期日变更)
    union all select 'HXLC_产品到期日变更', '提交需求', 'PRE_DEPT_APPR', '提交部门审批前'
    union all select 'HXLC_产品到期日变更', '复核需求', 'PRE_DEPT_APPR', '提交部门审批前'
    union all select 'HXLC_产品到期日变更', '提交部门审批', 'WAIT_SUBMIT', '在部门审批/敏捷会审议/确认结果节点'
    union all select 'HXLC_产品到期日变更', '组织敏捷小组会审议', 'WAIT_SUBMIT', '在部门审批/敏捷会审议/确认结果节点'
    union all select 'HXLC_产品到期日变更', '确认审议结果', 'WAIT_SUBMIT', '在部门审批/敏捷会审议/确认结果节点'
    union all select 'HXLC_产品到期日变更', '固定收益投资部门领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品到期日变更', '多资产投资部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品到期日变更', '多策略投资部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品到期日变更', '组合投资部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品到期日变更', '资产创设部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品到期日变更', '机构投资部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品到期日变更', '消费者权益保护岗', 'OTHERS', '其他情况'
    union all select 'HXLC_产品到期日变更', '法律合规部', 'OTHERS', '其他情况'
    union all select 'HXLC_产品到期日变更', '策略创新部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品到期日变更', '投资研究部领导', 'OTHERS', '其他情况'

    -- 产品分红方案(HXLC_产品分红方案_不定期分红，HXLC_产品分红方案_定期分红)
    union all select 'HXLC_产品分红方案_不定期分红', '提交需求', 'PRE_DEPT_APPR', '流程未过提交部门审批节点'
    union all select 'HXLC_产品分红方案_不定期分红', '复核需求', 'PRE_DEPT_APPR', '流程未过提交部门审批节点'
    union all select 'HXLC_产品分红方案_不定期分红', '部门领导审批', 'OTHERS', '其他情况'
    union all select 'HXLC_产品分红方案_不定期分红', '提交公司审批', 'WAIT_SUBMIT', '流程在提交部门审批节点或流程在提交公司审批节点'
    union all select 'HXLC_产品分红方案_不定期分红', '公司领导审批', 'OTHERS', '其他情况'
    union all select 'HXLC_产品分红方案_不定期分红', '办结', 'OTHERS', '其他情况'
    union all select 'HXLC_产品分红方案_定期分红', '制定分红方案', 'PRE_DEPT_APPR', '流程未过提交部门审批节点'
    union all select 'HXLC_产品分红方案_定期分红', '复核分红方案', 'PRE_DEPT_APPR', '流程未过提交部门审批节点'
    union all select 'HXLC_产品分红方案_定期分红', '部门领导审批', 'OTHERS', '其他情况'
    union all select 'HXLC_产品分红方案_定期分红', '办结', 'OTHERS', '其他情况'


    -- 产品定价管理 (HXLC_产品定价管理-敏捷小组)
    union all select 'HXLC_产品定价管理-敏捷小组', '提交需求', 'PRE_DEPT_APPR', '流程未过提交部门审批节点'
    union all select 'HXLC_产品定价管理-敏捷小组', '复核需求', 'PRE_DEPT_APPR', '流程未过提交部门审批节点'
    union all select 'HXLC_产品定价管理-敏捷小组', '提交部门审批', 'WAIT_SUBMIT', '流程在提交部门审批节点或流程在提交公司审批节点'
    union all select 'HXLC_产品定价管理-敏捷小组', '消费者权益保护岗', 'OTHERS', '其他情况'
    union all select 'HXLC_产品定价管理-敏捷小组', '法律合规部', 'OTHERS', '其他情况'
    union all select 'HXLC_产品定价管理-敏捷小组', '固定收益投资部门领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品定价管理-敏捷小组', '多资产投资部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品定价管理-敏捷小组', '多策略投资部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品定价管理-敏捷小组', '组合投资部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品定价管理-敏捷小组', '资产创设部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品定价管理-敏捷小组', '机构投资部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品定价管理-敏捷小组', '组织敏捷小组会审议', 'OTHERS', '其他情况'
    union all select 'HXLC_产品定价管理-敏捷小组', '确认审议结果', 'OTHERS', '其他情况'
    union all select 'HXLC_产品定价管理-敏捷小组', '策略创新部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品定价管理-敏捷小组', 'ESG部门领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品定价管理-敏捷小组', '投资研究部领导', 'OTHERS', '其他情况'

    -- 产品定价管理 (HXLC_产品定价管理-部门审批)
    union all select 'HXLC_产品定价管理-部门审批', '提交需求', 'PRE_DEPT_APPR', '流程未过提交部门审批节点'
    union all select 'HXLC_产品定价管理-部门审批', '复核需求', 'PRE_DEPT_APPR', '流程未过提交部门审批节点'
    union all select 'HXLC_产品定价管理-部门审批', '消费者权益保护岗', 'OTHERS', '其他情况'
    union all select 'HXLC_产品定价管理-部门审批', '法律合规部', 'OTHERS', '其他情况'
    union all select 'HXLC_产品定价管理-部门审批', '固定收益投资部门领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品定价管理-部门审批', '多资产投资部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品定价管理-部门审批', '多策略投资部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品定价管理-部门审批', '组合投资部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品定价管理-部门审批', '资产创设部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品定价管理-部门审批', '机构投资部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品定价管理-部门审批', '办结生效', 'OTHERS', '其他情况'
    union all select 'HXLC_产品定价管理-部门审批', '策略创新部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品定价管理-部门审批', 'ESG部门领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品定价管理-部门审批', '投资研究部领导', 'OTHERS', '其他情况'

    -- 产品定价管理 (HXLC_产品定价管理-公司领导审批)
    union all select 'HXLC_产品定价管理-公司领导审批', '提交需求', 'PRE_DEPT_APPR', '流程未过提交部门审批节点'
    union all select 'HXLC_产品定价管理-公司领导审批', '复核需求', 'PRE_DEPT_APPR', '流程未过提交部门审批节点'
    union all select 'HXLC_产品定价管理-公司领导审批', '提交部门审批', 'WAIT_SUBMIT', '流程在提交部门审批节点或流程在提交公司审批节点'
    union all select 'HXLC_产品定价管理-公司领导审批', '提交公司审批', 'WAIT_SUBMIT', '流程在提交部门审批节点或流程在提交公司审批节点'
    union all select 'HXLC_产品定价管理-公司领导审批', '产品营销部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品定价管理-公司领导审批', '固定收益投资部门领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品定价管理-公司领导审批', '多资产投资部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品定价管理-公司领导审批', '多策略投资部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品定价管理-公司领导审批', '组合投资部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品定价管理-公司领导审批', '资产创设部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品定价管理-公司领导审批', '机构投资部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品定价管理-公司领导审批', '公司领导审批', 'OTHERS', '其他情况'
    union all select 'HXLC_产品定价管理-公司领导审批', '消费者权益保护岗', 'OTHERS', '其他情况'
    union all select 'HXLC_产品定价管理-公司领导审批', '法律合规部', 'OTHERS', '其他情况'
    union all select 'HXLC_产品定价管理-公司领导审批', '策略创新部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品定价管理-公司领导审批', 'ESG部门领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品定价管理-公司领导审批', '投资研究部领导', 'OTHERS', '其他情况'
    union all select 'HXLC_产品定价管理-公司领导审批', '办结生效', 'OTHERS', '其他情况'

    -- 产品申报登记 (HXLC_产品申报登记 / 封闭式 / 维护) - 通用场景
    union all select 'HXLC_产品申报登记', '提交申报', 'COMMON', '通用'
    union all select 'HXLC_产品申报登记', '上传流动性风险评估文件', 'COMMON', '通用'
    union all select 'HXLC_产品申报登记', '制作报备材料', 'COMMON', '通用'

    union all select 'HXLC_产品申报登记_封闭式', '提交申报', 'COMMON', '通用'
    union all select 'HXLC_产品申报登记_封闭式', '制作报备材料', 'COMMON', '通用'
    union all select 'HXLC_产品申报登记_封闭式', '上传流动性风险评估文件', 'COMMON', '通用'
    union all select 'HXLC_产品申报登记_封闭式', '复核报备材料', 'COMMON', '通用'

    union all select 'HXLC_产品申报登记_维护', '修改信息', 'COMMON', '通用'

    -- 产品暂停发行管理 (HXLC_产品暂停发行管理) - 通用场景
    union all select 'HXLC_产品暂停发行管理', '提交需求', 'COMMON', '通用'
    union all select 'HXLC_产品暂停发行管理', '复核需求', 'COMMON', '通用'
    union all select 'HXLC_产品暂停发行管理', '固定收益投资部门领导', 'COMMON', '通用'
    union all select 'HXLC_产品暂停发行管理', '多资产投资部领导', 'COMMON', '通用'
    union all select 'HXLC_产品暂停发行管理', '多策略投资部领导', 'COMMON', '通用'
    union all select 'HXLC_产品暂停发行管理', '组合投资部领导', 'COMMON', '通用'
    union all select 'HXLC_产品暂停发行管理', '策略创新部领导', 'COMMON', '通用'
    union all select 'HXLC_产品暂停发行管理', '资产创设部领导', 'COMMON', '通用'
    union all select 'HXLC_产品暂停发行管理', 'ESG部门领导', 'COMMON', '通用'
    union all select 'HXLC_产品暂停发行管理', '机构投资部领导', 'COMMON', '通用'
    union all select 'HXLC_产品暂停发行管理', '投资研究部领导', 'COMMON', '通用'

    -- 产品发行登记 (HXLC_产品发行登记) - 通用场景
    union all select 'HXLC_产品发行登记', '发行登记信息复核', 'COMMON', '通用'
),
unfinished_instance as (
    select ins.INSTANCE_ID,
           ins.TMPL_ID,
           t.TMPL_NAME,
           ins.CREATE_USER,
           ins.BEG_TIME as CREATE_TIME,
           ins.O_CODE,
           s.SCENE_CODE,
           s.NODE_NAME
    from sys_app_wf_instance ins
    join tmpl t on t.R_ID = ins.TMPL_ID
    left join (
        select
            ins2.INSTANCE_ID,
            max(ic.SCENE_CODE) as SCENE_CODE,
            max(ic.NODE_NAME) as NODE_NAME
        from sys_app_wf_activity act
        join sys_app_wf_act_permission per
          on per.INSTANCE_ID = act.INSTANCE_ID
         and per.ACTIVITY_ID = act.ACTIVITY_ID
        left join sys_app_wf_instance ins2
          on ins2.INSTANCE_ID = act.INSTANCE_ID
        left join tmpl t2
          on t2.R_ID = ins2.TMPL_ID
        left join issue_node_cfg ic
          on ic.NODE_NAME = act.TASK_NAME
         and ic.TMPL_NAME = t2.TMPL_NAME
        where act.ACTIVITY_STATE = '1'
          and (per.STATE = '1' or per.STATE = '2')
          and per.PARTAKE_USER is not null
        group by ins2.INSTANCE_ID
    ) s
      on s.INSTANCE_ID = ins.INSTANCE_ID
    where ins.INSTANCE_STATE = '0'
),

-- 4. 部门领导配置 (静态映射)
-- 基于 fetch_leaders_and_roles.py 获取的结果
dept_leader_cfg as (
    select '固定收益投资部' as DEPT_NAME, 'hxlcyangmu' as LEADER_1, 'N0002977' as LEADER_2
    union all select '多资产投资部', 'hxlcnichun', 'N0002977'
    union all select '多策略投资部', 'hxlcwangsha', 'N0002977'
    union all select '组合投资部', 'N0002979', 'N0002977'
    union all select '资产创设部', 'N0002980', 'N0002977'
    union all select '策略创新部', 'N0002981', 'N0002977'
    union all select '产品营销部', 'hxlczxh', 'N0002977' -- 张晓华2 -> hxlczxh
    union all select '运营管理部', 'N0002982', 'N0002986'
    union all select '客户体验部', 'hxlcyanlv', null
    union all select '投资研究部', 'hxlcyqy', null
    union all select '法律合规部', 'hxlclifeng', null
    union all select '机构投资部', 'hxlcdujianzhi', 'N0002977'
    union all select 'ESG事业部', 'hxlcwhk', null
),
-- 5. 产品管理角色人员 (HXLC00006)
-- 5. 产品管理角色人员 (HXLC00006)
prod_managers as (
    select ur.R_CODE1 as USER_O_CODE
    from sys_rbac_role_relate ur
    where ur.ROLE_CODE = 'HXLC00006' -- 产品管理角色
    -- and ur.R_TYPE = 'USER' -- 暂时注释，除非确认
),

-- 6. 领导黑名单集合 (用于发送时过滤)
excluded_leaders as (
    select LEADER_1 as account from dept_leader_cfg where LEADER_1 is not null
    union
    select LEADER_2 as account from dept_leader_cfg where LEADER_2 is not null
),

-- ===============================================================
-- 分支逻辑 A: 产品申报登记 (Sb)
-- ===============================================================
sb_raw as (
    select info.INSTANCE_ID,
           coalesce(info.HN_BSM, info.PRD_O_CODE) as PRD_CODE,
           info.REG_APPLICATION_DATE as RAISE_START_DATE,
           coalesce(d.DEPT_O_NAME, info.FQBM) as PRD_DEPT
    from ods_prdtq_cpsbxxgl_info info
    join unfinished_instance ui on ui.INSTANCE_ID = info.INSTANCE_ID
    left join sys_rbac_user u on u.USER_O_CODE = ui.CREATE_USER
    left join sys_rbac_depart d on u.ORG regexp d.DEPT_O_CODE
    where info.D_FLAG <> '7'
      and ui.TMPL_NAME like '%申报登记%' -- 确保只匹配申报登记的流程
),
sb_prd as (
    select s.INSTANCE_ID,
           s.PRD_CODE,
           b.PRD_O_NAME,
           b.PRD_COLL_MODE,
           s.RAISE_START_DATE,
           b.FIRST_INVEST_MANAGER,
           b.SALES_MANAGER,
           s.PRD_DEPT 
    from sb_raw s
    join dw_prd_base_info b
      on b.PRD_O_CODE = s.PRD_CODE
     and b.D_FLAG = '0'
),
sb_critical as (
    select
        p.INSTANCE_ID,
        case
            when p.PRD_COLL_MODE = '1' then (
                select concat(d.N_DAY, ' 00:00:00')
                from dw_dt_date d
                where d.DT_TYPE = 'A' and d.N_DAY <= p.RAISE_START_DATE
                order by d.N_DAY desc limit 10, 1
            )
            when p.PRD_COLL_MODE = '2' then (
                select concat(d.N_DAY, ' 00:00:00')
                from dw_dt_date d
                where d.DT_TYPE = 'A' and d.N_DAY <= p.RAISE_START_DATE
                order by d.N_DAY desc limit 2, 1
            )
            else null
        end as CRITICAL_DT
    from sb_prd p
),
-- 申报登记最终目标集
target_sb as (
    select
        p.INSTANCE_ID,
        p.PRD_CODE,
        p.PRD_O_NAME as PRD_NAME,
        '申报登记提醒' as ALERT_TYPE,
        p.FIRST_INVEST_MANAGER,
        p.SALES_MANAGER,
        null as START_TIME,
        null as DAYS_ELAPSED,
        c.CRITICAL_DT,
        p.PRD_COLL_MODE,
        concat(
            '当前时间已达到或超过临界时点 ',
            date_format(c.CRITICAL_DT, '%Y-%m-%d %H:%i:%s'),
            '，流程仍为未完结状态'
        ) as TRIGGER_REASON,
        p.PRD_DEPT,
        '' as PM_RECEIVER_TYPE,
        0 as SEND_INITIATOR,
        0 as SEND_INV_SALES,
        0 as SEND_SB_REGISTER,
        0 as SEND_ISSUE_REGISTER,
        0 as SEND_DISC_INFO
    from sb_prd p
    join sb_critical c on c.INSTANCE_ID = p.INSTANCE_ID
    where c.CRITICAL_DT is not null
      and current_timestamp >= c.CRITICAL_DT
),

-- ===============================================================
-- 分支逻辑 B: 产品发行审批 (Issue)
-- ===============================================================
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
      and ui.TMPL_NAME like '%发行审批%'
),
issue_critical as (
    select
        p.INSTANCE_ID,
        concat(date_sub(p.COLLECT_VDATE, interval 1 day), ' 00:00:00') as CRITICAL_DT
    from prd_yz p
),
target_issue as (
    select
        ui.INSTANCE_ID,
        pyz.PRD_CODE,
        pyz.PRD_NAME,
        '发行审批提醒' as ALERT_TYPE,
        null as FIRST_INVEST_MANAGER,
        null as SALES_MANAGER,
        ui.CREATE_TIME as START_TIME,
        datediff(current_date, date(ui.CREATE_TIME)) as DAYS_ELAPSED,
        c.CRITICAL_DT,
        null as PRD_COLL_MODE,
        concat(
            '当前时间已达到或超过临界时点 ',
            date_format(c.CRITICAL_DT, '%Y-%m-%d %H:%i:%s'),
            '，流程仍为未完结状态'
        ) as TRIGGER_REASON,
        pyz.PRD_DEPT,
        case
            when ui.SCENE_CODE = 'WAIT_SUBMIT' then 'ALL_PM'
            when ui.SCENE_CODE = 'OTHERS' then 'FLOW_PM'
            else ''
        end as PM_RECEIVER_TYPE,
        1 as SEND_INITIATOR,
        1 as SEND_INV_SALES,
        case
            when ui.SCENE_CODE in ('WAIT_SUBMIT', 'OTHERS') then 1
            else 0
        end as SEND_SB_REGISTER,
        0 as SEND_ISSUE_REGISTER,
        case
            when ui.SCENE_CODE in ('WAIT_SUBMIT', 'OTHERS') then 1
            else 0
        end as SEND_DISC_INFO
    from unfinished_instance ui
    join prd_yz pyz on pyz.INSTANCE_ID = ui.INSTANCE_ID
    join issue_critical c on c.INSTANCE_ID = ui.INSTANCE_ID
    where c.CRITICAL_DT is not null
      and current_timestamp >= c.CRITICAL_DT
),

-- ===============================================================
-- 分支逻辑 C: 发行前变更 (Prechange)
-- ===============================================================
prechange_prd as (
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
      and ui.TMPL_NAME like '%发行前变更%'
),
prechange_critical as (
    select
        p.INSTANCE_ID,
        concat(date_sub(p.COLLECT_VDATE, interval 2 day), ' 13:00:00') as CRITICAL_DT
    from prechange_prd p
),
target_prechange as (
    select
        ui.INSTANCE_ID,
        p.PRD_CODE,
        p.PRD_NAME,
        '发行前变更提醒' as ALERT_TYPE,
        null as FIRST_INVEST_MANAGER,
        null as SALES_MANAGER,
        ui.CREATE_TIME as START_TIME,
        datediff(current_date, date(ui.CREATE_TIME)) as DAYS_ELAPSED,
        pc.CRITICAL_DT,
        null as PRD_COLL_MODE,
        concat(
            '当前时间已达到或超过临界时点 ',
            date_format(pc.CRITICAL_DT, '%Y-%m-%d %H:%i:%s'),
            '，流程仍为未完结状态'
        ) as TRIGGER_REASON,
        p.PRD_DEPT,
        case
            when ui.SCENE_CODE = 'WAIT_SUBMIT' then 'ALL_PM'
            when ui.SCENE_CODE = 'OTHERS' then 'FLOW_PM'
            else ''
        end as PM_RECEIVER_TYPE,
        1 as SEND_INITIATOR,
        1 as SEND_INV_SALES,
        case
            when ui.SCENE_CODE in ('WAIT_SUBMIT', 'OTHERS') then 1
            else 0
        end as SEND_SB_REGISTER,
        0 as SEND_ISSUE_REGISTER,
        case
            when ui.SCENE_CODE in ('WAIT_SUBMIT', 'OTHERS') then 1
            else 0
        end as SEND_DISC_INFO
    from unfinished_instance ui
    join prechange_prd p on p.INSTANCE_ID = ui.INSTANCE_ID
    join prechange_critical pc on pc.INSTANCE_ID = ui.INSTANCE_ID
    where pc.CRITICAL_DT is not null
      and current_timestamp >= pc.CRITICAL_DT
),

-- ===============================================================
-- 分支逻辑 D: 暂停发行 (Suspend)
-- ===============================================================
suspend_prd as (
    select s.INSTANCE_ID,
           yz.PRD_O_CODE as PRD_CODE,
           yz.PRD_O_NAME as PRD_NAME,
           yz.COLLECT_VDATE,
           coalesce(d.DEPT_O_NAME, yz.FQBM) as PRD_DEPT
    from ODS_PRDYZ_ISSUE_SUSPEND s
    join unfinished_instance ui on ui.INSTANCE_ID = s.INSTANCE_ID
    join ODS_PRDYZ_BASE_INFO yz on yz.PRD_O_CODE = s.PRD_O_CODE
    left join sys_rbac_user u
      on u.USER_O_CODE = ui.CREATE_USER
    left join sys_rbac_depart d
      on u.ORG regexp d.DEPT_O_CODE
    where s.D_FLAG <> '7'
      and yz.D_FLAG <> '7'
      and ui.TMPL_NAME like '%暂停发行%'
),
suspend_critical as (
    select
        p.INSTANCE_ID,
        concat(date_sub(p.COLLECT_VDATE, interval 2 day), ' 13:00:00') as CRITICAL_DT
    from suspend_prd p
),
target_suspend as (
    select
        ui.INSTANCE_ID,
        p.PRD_CODE,
        p.PRD_NAME,
        '暂停发行提醒' as ALERT_TYPE,
        null as FIRST_INVEST_MANAGER,
        null as SALES_MANAGER,
        ui.CREATE_TIME as START_TIME,
        datediff(current_date, date(ui.CREATE_TIME)) as DAYS_ELAPSED,
        sc.CRITICAL_DT,
        null as PRD_COLL_MODE,
        concat(
            '当前时间已达到或超过临界时点 ',
            date_format(sc.CRITICAL_DT, '%Y-%m-%d %H:%i:%s'),
            '，流程仍为未完结状态'
        ) as TRIGGER_REASON,
        p.PRD_DEPT,
        'ALL_PM' as PM_RECEIVER_TYPE,
        1 as SEND_INITIATOR,
        1 as SEND_INV_SALES,
        1 as SEND_SB_REGISTER,
        0 as SEND_ISSUE_REGISTER,
        1 as SEND_DISC_INFO
    from unfinished_instance ui
    join suspend_prd p on p.INSTANCE_ID = ui.INSTANCE_ID
    join suspend_critical sc on sc.INSTANCE_ID = ui.INSTANCE_ID
    where sc.CRITICAL_DT is not null
      and current_timestamp >= sc.CRITICAL_DT
),

-- ===============================================================
-- 分支逻辑 E: 发行登记 (IssueReg)
-- ===============================================================
issue_reg_prd as (
    select r.INSTANCE_ID,
           yz.PRD_O_CODE as PRD_CODE,
           yz.PRD_O_NAME as PRD_NAME,
           yz.COLLECT_VDATE,
           coalesce(d.DEPT_O_NAME, yz.FQBM) as PRD_DEPT
    from ODS_PRDYZ_ISSUE_REG r
    join unfinished_instance ui on ui.INSTANCE_ID = r.INSTANCE_ID
    join ODS_PRDYZ_BASE_INFO yz on yz.PRD_O_CODE = r.PRD_O_CODE
    left join sys_rbac_user u
      on u.USER_O_CODE = ui.CREATE_USER
    left join sys_rbac_depart d
      on u.ORG regexp d.DEPT_O_CODE
    where r.D_FLAG <> '7'
      and yz.D_FLAG <> '7'
      and ui.TMPL_NAME like '%发行登记%'
),
issue_reg_critical as (
    select
        p.INSTANCE_ID,
        concat(date_sub(p.COLLECT_VDATE, interval 1 day), ' 13:00:00') as CRITICAL_DT
    from issue_reg_prd p
),
target_issue_reg as (
    select
        ui.INSTANCE_ID,
        p.PRD_CODE,
        p.PRD_NAME,
        '发行登记提醒' as ALERT_TYPE,
        null as FIRST_INVEST_MANAGER,
        null as SALES_MANAGER,
        ui.CREATE_TIME as START_TIME,
        datediff(current_date, date(ui.CREATE_TIME)) as DAYS_ELAPSED,
        ic.CRITICAL_DT,
        null as PRD_COLL_MODE,
        concat(
            '当前时间已达到或超过临界时点 ',
            date_format(ic.CRITICAL_DT, '%Y-%m-%d %H:%i:%s'),
            '，流程仍为未完结状态'
        ) as TRIGGER_REASON,
        p.PRD_DEPT,
        '' as PM_RECEIVER_TYPE,
        0 as SEND_INITIATOR,
        0 as SEND_INV_SALES,
        0 as SEND_SB_REGISTER,
        0 as SEND_ISSUE_REGISTER,
        0 as SEND_DISC_INFO
    from unfinished_instance ui
    join issue_reg_prd p on p.INSTANCE_ID = ui.INSTANCE_ID
    join issue_reg_critical ic on ic.INSTANCE_ID = ui.INSTANCE_ID
    where ic.CRITICAL_DT is not null
      and current_timestamp >= ic.CRITICAL_DT
),

-- ===============================================================
-- 分支逻辑 F: 产品定价管理 (PriceAdjust)
-- ===============================================================
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
price_critical as (
    select
        p.INSTANCE_ID,
        concat(date_sub(p.COLLECT_VDATE, interval 3 day), ' 13:00:00') as CRITICAL_DT
    from price_prd p
),
target_price as (
    select
        ui.INSTANCE_ID,
        p.PRD_CODE,
        p.PRD_NAME,
        '产品定价管理提醒' as ALERT_TYPE,
        null as FIRST_INVEST_MANAGER,
        null as SALES_MANAGER,
        ui.CREATE_TIME as START_TIME,
        datediff(current_date, date(ui.CREATE_TIME)) as DAYS_ELAPSED,
        pc.CRITICAL_DT,
        null as PRD_COLL_MODE,
        concat(
            '当前时间已达到或超过临界时点 ',
            date_format(pc.CRITICAL_DT, '%Y-%m-%d %H:%i:%s'),
            '，流程仍为未完结状态'
        ) as TRIGGER_REASON,
        p.PRD_DEPT,
        case
            when ui.TMPL_NAME = 'HXLC_产品定价管理-部门审批'
                 and ui.NODE_NAME like '%部门领导%' then 'ALL_PM'
            when ui.TMPL_NAME = 'HXLC_产品定价管理-公司领导审批'
                 and ui.SCENE_CODE = 'WAIT_SUBMIT' then 'ALL_PM'
            when ui.TMPL_NAME = 'HXLC_产品定价管理-公司领导审批'
                 and ui.SCENE_CODE = 'OTHERS' then 'FLOW_PM'
            when ui.TMPL_NAME = 'HXLC_产品定价管理-敏捷小组'
                 and ui.SCENE_CODE in ('WAIT_SUBMIT', 'OTHERS') then 'ALL_PM'
            else ''
        end as PM_RECEIVER_TYPE,
        1 as SEND_INITIATOR,
        1 as SEND_INV_SALES,
        0 as SEND_SB_REGISTER,
        0 as SEND_ISSUE_REGISTER,
        case
            when ui.TMPL_NAME = 'HXLC_产品定价管理-部门审批'
                 and ui.NODE_NAME like '%部门领导%' then 1
            when ui.TMPL_NAME in (
                     'HXLC_产品定价管理-公司领导审批',
                     'HXLC_产品定价管理-敏捷小组'
                 )
                 and ui.SCENE_CODE in ('WAIT_SUBMIT', 'OTHERS') then 1
            else 0
        end as SEND_DISC_INFO
    from unfinished_instance ui
    join price_prd p on p.INSTANCE_ID = ui.INSTANCE_ID
    join price_critical pc on pc.INSTANCE_ID = ui.INSTANCE_ID
    where pc.CRITICAL_DT is not null
      and current_timestamp >= pc.CRITICAL_DT
),

-- ===============================================================
-- 分支逻辑 G: 增发份额 (AddIssue)
-- ===============================================================
add_issue_prd as (
    select s.INSTANCE_ID,
           s.PRD_O_CODE as PRD_CODE,
           s.PRD_O_NAME as PRD_NAME,
           i.SPRD_PALN_CDATE as PLAN_CDATE,
           coalesce(d.DEPT_O_NAME, s.FQBM) as PRD_DEPT
    from ODS_PRDYZ_ADD_ISSUE_SON s
    join unfinished_instance ui on ui.INSTANCE_ID = s.INSTANCE_ID
    left join ODS_PRDYZ_ISSUE_INFO i
      on i.PRDYZ_O_CODE = s.PRDYZ_O_CODE
    left join sys_rbac_user u
      on u.USER_O_CODE = ui.CREATE_USER
    left join sys_rbac_depart d
      on u.ORG regexp d.DEPT_O_CODE
    where s.D_FLAG != 7
),
add_issue_critical as (
    select
        p.INSTANCE_ID,
        concat(date_sub(p.PLAN_CDATE, interval 3 day), ' 13:00:00') as CRITICAL_DT
    from add_issue_prd p
),
target_add_issue as (
    select
        ui.INSTANCE_ID,
        p.PRD_CODE,
        p.PRD_NAME,
        '增发份额提醒' as ALERT_TYPE,
        null as FIRST_INVEST_MANAGER,
        null as SALES_MANAGER,
        ui.CREATE_TIME as START_TIME,
        datediff(current_date, date(ui.CREATE_TIME)) as DAYS_ELAPSED,
        ac.CRITICAL_DT,
        null as PRD_COLL_MODE,
        concat(
            '当前时间已达到或超过临界时点 ',
            date_format(ac.CRITICAL_DT, '%Y-%m-%d %H:%i:%s'),
            '，流程仍为未完结状态'
        ) as TRIGGER_REASON,
        p.PRD_DEPT,
        case
            when ui.SCENE_CODE = 'WAIT_SUBMIT' then 'ALL_PM'
            when ui.SCENE_CODE = 'OTHERS' then 'FLOW_PM'
            else ''
        end as PM_RECEIVER_TYPE,
        1 as SEND_INITIATOR,
        1 as SEND_INV_SALES,
        0 as SEND_SB_REGISTER,
        0 as SEND_ISSUE_REGISTER,
        case
            when ui.SCENE_CODE in ('WAIT_SUBMIT', 'OTHERS') then 1
            else 0
        end as SEND_DISC_INFO
    from unfinished_instance ui
    join add_issue_prd p on p.INSTANCE_ID = ui.INSTANCE_ID
    join add_issue_critical ac on ac.INSTANCE_ID = ui.INSTANCE_ID
    where ac.CRITICAL_DT is not null
      and current_timestamp >= ac.CRITICAL_DT
),

-- ===============================================================
-- 分支逻辑 H: 销售信息变更 (SalesChange)
-- ===============================================================
sales_change_prd as (
    select a.INSTANCE_ID,
           a.PRD_O_CODE as PRD_CODE,
           a.PRD_O_NAME as PRD_NAME,
           a.EFF_DATE,
           d.DEPT_O_NAME as PRD_DEPT
    from ODS_PRDYZ_SALES_CHANGE a
    join unfinished_instance ui on ui.INSTANCE_ID = a.INSTANCE_ID
    left join sys_app_wf_instance ins on ins.INSTANCE_ID = a.INSTANCE_ID
    left join sys_rbac_user u on u.USER_O_CODE = ins.CREATE_USER
    left join sys_rbac_depart d on u.ORG regexp d.DEPT_O_CODE
),
sales_change_critical as (
    select
        p.INSTANCE_ID,
        concat(date_sub(p.EFF_DATE, interval 3 day), ' 13:00:00') as CRITICAL_DT
    from sales_change_prd p
),
target_sales_change as (
    select
        ui.INSTANCE_ID,
        p.PRD_CODE,
        p.PRD_NAME,
        '销售信息变更提醒' as ALERT_TYPE,
        null as FIRST_INVEST_MANAGER,
        null as SALES_MANAGER,
        ui.CREATE_TIME as START_TIME,
        datediff(current_date, date(ui.CREATE_TIME)) as DAYS_ELAPSED,
        sc.CRITICAL_DT,
        null as PRD_COLL_MODE,
        concat(
            '当前时间已达到或超过临界时点 ',
            date_format(sc.CRITICAL_DT, '%Y-%m-%d %H:%i:%s'),
            '，流程仍为未完结状态'
        ) as TRIGGER_REASON,
        p.PRD_DEPT,
        case
            when ui.SCENE_CODE = 'WAIT_SUBMIT' then 'ALL_PM'
            when ui.SCENE_CODE = 'OTHERS' then 'FLOW_PM'
            else ''
        end as PM_RECEIVER_TYPE,
        1 as SEND_INITIATOR,
        1 as SEND_INV_SALES,
        0 as SEND_SB_REGISTER,
        0 as SEND_ISSUE_REGISTER,
        case
            when ui.SCENE_CODE in ('WAIT_SUBMIT', 'OTHERS') then 1
            else 0
        end as SEND_DISC_INFO
    from unfinished_instance ui
    join sales_change_prd p on p.INSTANCE_ID = ui.INSTANCE_ID
    join sales_change_critical sc on sc.INSTANCE_ID = ui.INSTANCE_ID
    where sc.CRITICAL_DT is not null
      and current_timestamp >= sc.CRITICAL_DT
),
open_plan_prd as (
    select a.INSTANCE_ID,
           a.PRD_O_CODE as PRD_CODE,
           a.PRD_O_NAME as PRD_NAME,
           str_to_date(a.CDATE, '%Y-%m-%d') as OPEN_DATE,
           d.DEPT_O_NAME as PRD_DEPT
    from ODS_PRDYZ_OPEN_PLAN_ADJUST a
    join unfinished_instance ui on ui.INSTANCE_ID = a.INSTANCE_ID
    left join sys_rbac_depart d on a.ORG regexp d.DEPT_O_CODE
    where a.D_FLAG != 7
),
open_plan_critical as (
    select
        p.INSTANCE_ID,
        concat(date_sub(p.OPEN_DATE, interval 3 day), ' 13:00:00') as CRITICAL_DT
    from open_plan_prd p
),
target_open_plan as (
    select
        ui.INSTANCE_ID,
        p.PRD_CODE,
        p.PRD_NAME,
        '开放计划调整提醒' as ALERT_TYPE,
        null as FIRST_INVEST_MANAGER,
        null as SALES_MANAGER,
        ui.CREATE_TIME as START_TIME,
        datediff(current_date, date(ui.CREATE_TIME)) as DAYS_ELAPSED,
        oc.CRITICAL_DT,
        null as PRD_COLL_MODE,
        concat(
            '当前时间已达到或超过临界时点 ',
            date_format(oc.CRITICAL_DT, '%Y-%m-%d %H:%i:%s'),
            '，流程仍为未完结状态'
        ) as TRIGGER_REASON,
        p.PRD_DEPT,
        case
            when ui.SCENE_CODE = 'WAIT_SUBMIT' then 'ALL_PM'
            when ui.SCENE_CODE = 'OTHERS' then 'FLOW_PM'
            else ''
        end as PM_RECEIVER_TYPE,
        1 as SEND_INITIATOR,
        1 as SEND_INV_SALES,
        0 as SEND_SB_REGISTER,
        0 as SEND_ISSUE_REGISTER,
        case
            when ui.SCENE_CODE in ('WAIT_SUBMIT', 'OTHERS') then 1
            else 0
        end as SEND_DISC_INFO
    from unfinished_instance ui
    join open_plan_prd p on p.INSTANCE_ID = ui.INSTANCE_ID
    join open_plan_critical oc on oc.INSTANCE_ID = ui.INSTANCE_ID
    where oc.CRITICAL_DT is not null
      and current_timestamp >= oc.CRITICAL_DT
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
dividend_critical as (
    select
        p.INSTANCE_ID,
        concat(date_sub(p.EX_DATE, interval 3 day), ' 13:00:00') as CRITICAL_DT
    from dividend_prd p
),
target_dividend as (
    select
        ui.INSTANCE_ID,
        p.PRD_CODE,
        p.PRD_NAME,
        '分红方案管理提醒' as ALERT_TYPE,
        null as FIRST_INVEST_MANAGER,
        null as SALES_MANAGER,
        ui.CREATE_TIME as START_TIME,
        datediff(current_date, date(ui.CREATE_TIME)) as DAYS_ELAPSED,
        dc.CRITICAL_DT,
        null as PRD_COLL_MODE,
        concat(
            '当前时间已达到或超过临界时点 ',
            date_format(dc.CRITICAL_DT, '%Y-%m-%d %H:%i:%s'),
            '，流程仍为未完结状态'
        ) as TRIGGER_REASON,
        p.PRD_DEPT,
        case
            when ui.TMPL_NAME = 'HXLC_产品分红方案_不定期分红'
                 and ui.SCENE_CODE = 'WAIT_SUBMIT' then 'ALL_PM'
            when ui.TMPL_NAME = 'HXLC_产品分红方案_不定期分红' then 'FLOW_PM'
            else ''
        end as PM_RECEIVER_TYPE,
        1 as SEND_INITIATOR,
        1 as SEND_INV_SALES,
        0 as SEND_SB_REGISTER,
        0 as SEND_ISSUE_REGISTER,
        case
            when ui.TMPL_NAME = 'HXLC_产品分红方案_不定期分红' then 1
            else 0
        end as SEND_DISC_INFO
    from unfinished_instance ui
    join dividend_prd p on p.INSTANCE_ID = ui.INSTANCE_ID
    join dividend_critical dc on dc.INSTANCE_ID = ui.INSTANCE_ID
    where dc.CRITICAL_DT is not null
      and current_timestamp >= dc.CRITICAL_DT
),

-- ===============================================================
-- 分支逻辑 K: 产品到期日变更 (MaturityChange)
-- ===============================================================
maturity_prd as (
    select e.INSTANCE_ID,
           e.PRD_O_CODE as PRD_CODE,
           coalesce(b.PRD_O_NAME, e.PRD_O_SNAME, e.PRD_O_CODE) as PRD_NAME,
           str_to_date(e.A_MDATE, '%Y-%m-%d') as NEW_END_DATE,
           coalesce(d.DEPT_O_NAME, e.FQBM) as PRD_DEPT
    from ODS_PRDJS_END e
    join unfinished_instance ui on ui.INSTANCE_ID = e.INSTANCE_ID
    left join dw_prd_base_info b
      on b.PRD_O_CODE = e.PRD_O_CODE
     and b.D_FLAG = '0'
    left join sys_rbac_user u
      on u.USER_O_CODE = ui.CREATE_USER
    left join sys_rbac_depart d
      on u.ORG regexp d.DEPT_O_CODE
    where e.D_FLAG != 7
      and ui.TMPL_NAME = 'HXLC_产品到期日变更'
),
maturity_critical as (
    select
        p.INSTANCE_ID,
        concat(date_sub(p.NEW_END_DATE, interval 3 day), ' 13:00:00') as CRITICAL_DT
    from maturity_prd p
),
target_maturity_change as (
    select
        ui.INSTANCE_ID,
        p.PRD_CODE,
        p.PRD_NAME,
        '产品到期日变更提醒' as ALERT_TYPE,
        null as FIRST_INVEST_MANAGER,
        null as SALES_MANAGER,
        ui.CREATE_TIME as START_TIME,
        datediff(current_date, date(ui.CREATE_TIME)) as DAYS_ELAPSED,
        mc.CRITICAL_DT,
        null as PRD_COLL_MODE,
        concat(
            '当前时间已达到或超过临界时点 ',
            date_format(mc.CRITICAL_DT, '%Y-%m-%d %H:%i:%s'),
            '，流程仍为未完结状态'
        ) as TRIGGER_REASON,
        p.PRD_DEPT,
        case
            when ui.SCENE_CODE = 'WAIT_SUBMIT' then 'ALL_PM'
            when ui.SCENE_CODE = 'OTHERS' then 'FLOW_PM'
            else ''
        end as PM_RECEIVER_TYPE,
        1 as SEND_INITIATOR,
        1 as SEND_INV_SALES,
        0 as SEND_SB_REGISTER,
        0 as SEND_ISSUE_REGISTER,
        0 as SEND_DISC_INFO
    from unfinished_instance ui
    join maturity_prd p on p.INSTANCE_ID = ui.INSTANCE_ID
    join maturity_critical mc on mc.INSTANCE_ID = ui.INSTANCE_ID
    where mc.CRITICAL_DT is not null
      and current_timestamp >= mc.CRITICAL_DT
),

combined_targets as (
    select * from target_sb
    union all
    select * from target_issue
    union all
    select * from target_prechange
    union all
    select * from target_suspend
    union all
    select * from target_issue_reg
    union all
    select * from target_price
    union all
    select * from target_add_issue
    union all
    select * from target_sales_change
    union all
    select * from target_open_plan
    union all
    select * from target_dividend
    union all
    select * from target_maturity_change
),

-- ===============================================================
-- 关联当前待办人与生成最终结果
-- ===============================================================
current_todo as (
    select act.INSTANCE_ID,
           act.ACTIVITY_ID,
           act.TASK_NAME,
           per.PARTAKE_USER as account,
           u.USER_O_NAME as USER_NAME,
           ui.TMPL_NAME as TMPL_NAME,
           ic.NODE_NAME as NODE_NAME,
           ic.SCENE_CODE as SCENE_CODE,
           ic.SCENE_NAME as SCENE_NAME,
           case when pm.USER_O_CODE is not null then 1 else 0 end as IS_PROD_MANAGER
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
    left join prod_managers pm
      on pm.USER_O_CODE = per.PARTAKE_USER
    where act.ACTIVITY_STATE = '1'
      and (per.STATE = '1' or per.STATE = '2')
      and per.PARTAKE_USER is not null
      and ic.SCENE_CODE is not null
),

base_alerts as (
    select
        ct.account,
        case
            when ti.ALERT_TYPE = '申报登记提醒' then
                concat(
                    '【产品运营管理系统】紧急流程预警\n',
                    '流程类型：【产品申报登记】\n',
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
                    date_format(ti.CRITICAL_DT, '%Y-%m-%d'),
                    '\n',
                    '当前流程已处于业务临界时点，请及时关注进度！'
                )
            when ti.ALERT_TYPE = '发行审批提醒' then
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
                    coalesce(ct.TASK_NAME, '未知'),
                    '\n',
                    '待办人员：',
                    coalesce(ct.USER_NAME, ct.account),
                    '\n',
                    '业务生效日：',
                    date_format(ti.CRITICAL_DT, '%Y-%m-%d'),
                    '\n',
                    '当前流程已处于业务临界时点，请及时关注进度！'
                )
            when ti.ALERT_TYPE = '发行前变更提醒' then
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
                    coalesce(ct.TASK_NAME, '未知'),
                    '\n',
                    '待办人员：',
                    coalesce(ct.USER_NAME, ct.account),
                    '\n',
                    '业务生效日：',
                    date_format(ti.CRITICAL_DT, '%Y-%m-%d'),
                    '\n',
                    '当前流程已处于业务临界时点，请及时关注进度！'
                )
            when ti.ALERT_TYPE = '增发份额提醒' then
                concat(
                    '【产品运营管理系统】紧急流程预警\n',
                    '流程类型：【增发份额】\n',
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
                    date_format(ti.CRITICAL_DT, '%Y-%m-%d'),
                    '\n',
                    '当前流程已处于业务临界时点，请及时关注进度！'
                )
            when ti.ALERT_TYPE = '销售信息变更提醒' then
                concat(
                    '【产品运营管理系统】紧急流程预警\n',
                    '流程类型：【销售相关信息变更】\n',
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
                    date_format(ti.CRITICAL_DT, '%Y-%m-%d'),
                    '\n',
                    '当前流程已处于业务临界时点，请及时关注进度！'
                )
            when ti.ALERT_TYPE = '开放计划调整提醒' then
                concat(
                    '【产品运营管理系统】紧急流程预警\n',
                    '流程类型：【开放计划调整】\n',
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
                    date_format(ti.CRITICAL_DT, '%Y-%m-%d'),
                    '\n',
                    '当前流程已处于业务临界时点，请及时关注进度！'
                )
            when ti.ALERT_TYPE = '暂停发行提醒' then
                concat(
                    '【产品运营管理系统】紧急流程预警\n',
                    '流程类型：【产品暂停发行管理】\n',
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
                    date_format(ti.CRITICAL_DT, '%Y-%m-%d'),
                    '\n',
                    '当前流程已处于业务临界时点，请及时关注进度！'
                )
            when ti.ALERT_TYPE = '发行登记提醒' then
                concat(
                    '【产品运营管理系统】紧急流程预警\n',
                    '流程类型：【产品发行登记】\n',
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
                    date_format(ti.CRITICAL_DT, '%Y-%m-%d'),
                    '\n',
                    '当前流程已处于业务临界时点，请及时关注进度！'
                )
            when ti.ALERT_TYPE = '产品定价管理提醒' then
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
                    date_format(ti.CRITICAL_DT, '%Y-%m-%d'),
                    '\n',
                    '当前流程已处于业务临界时点，请及时关注进度！'
                )
            when ti.ALERT_TYPE = '产品到期日变更提醒' then
                concat(
                    '【产品运营管理系统】紧急流程预警\n',
                    '流程类型：【产品到期日变更】\n',
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
                    date_format(ti.CRITICAL_DT, '%Y-%m-%d'),
                    '\n',
                    '当前流程已处于业务临界时点，请及时关注进度！'
                )
            when ti.ALERT_TYPE = '分红方案管理提醒' then
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
                    coalesce(ct.TASK_NAME, '未知'),
                    '\n',
                    '待办人员：',
                    coalesce(ct.USER_NAME, ct.account),
                    '\n',
                    '业务生效日：',
                    date_format(ti.CRITICAL_DT, '%Y-%m-%d'),
                    '\n',
                    '当前流程已处于业务临界时点，请及时关注进度！'
                )
            else ''
        end as CONTENT,
        'SYSTEM' as OPERATOR,
        'SYSTEM' as OPERATOR_ACCOUNT,
        '定时消息提醒' as `TRIGGER`,
        ti.PRD_CODE,
        ti.PRD_NAME,
        ti.ALERT_TYPE,
        ti.PRD_DEPT,
        ti.INSTANCE_ID,
        ct.ACTIVITY_ID,
        ct.TASK_NAME,
        ct.NODE_NAME,
        ct.SCENE_NAME,
        ct.USER_NAME,
        ct.IS_PROD_MANAGER,
        dl.LEADER_1 as DEPT_LEADER_1,
        dl.LEADER_2 as DEPT_LEADER_2,
        ti.PM_RECEIVER_TYPE,
        ti.START_TIME,
        ti.DAYS_ELAPSED,
        ti.CRITICAL_DT,
        ti.TRIGGER_REASON,
        ti.SEND_INITIATOR,
        ti.SEND_INV_SALES,
        ti.SEND_SB_REGISTER,
        ti.SEND_ISSUE_REGISTER,
        ti.SEND_DISC_INFO
    from combined_targets ti
    join current_todo ct on ct.INSTANCE_ID = ti.INSTANCE_ID
    left join dept_leader_cfg dl on dl.DEPT_NAME = ti.PRD_DEPT
),

-- 产品在 DW 基础信息表中的三类角色映射（产品经理 / 投资助理 / 销售经理）
prd_roles as (
    select
        b.PRD_O_CODE as PRD_CODE,
        b.PM_MANAGER,
        b.SECOND_MANAGER,
        b.SALES_MANAGER
    from dw_prd_base_info b
    where b.D_FLAG = '0'
),

-- 角色【产品管理】下的全部用户集合
product_mgmt_users as (
    select
        u.USER_O_CODE as account
    from sys_rbac_role r
    join sys_rbac_role_relate ur
      on ur.ROLE_CODE = r.ROLE_O_CODE
    join sys_rbac_user u
      on u.R_ID = ur.R_CODE1
    where r.ROLE_O_NAME = '产品管理'
),

-- 角色【申报登记岗】下的全部用户集合
sb_register_users as (
    select
        u.USER_O_CODE as account
    from sys_rbac_role r
    join sys_rbac_role_relate ur
      on ur.ROLE_CODE = r.ROLE_O_CODE
    join sys_rbac_user u
      on u.R_ID = ur.R_CODE1
    where r.ROLE_O_NAME = '申报登记'
),

-- 角色【发行登记岗】下的全部用户集合
issue_register_users as (
    select
        u.USER_O_CODE as account
    from sys_rbac_role r
    join sys_rbac_role_relate ur
      on ur.ROLE_CODE = r.ROLE_O_CODE
    join sys_rbac_user u
      on u.R_ID = ur.R_CODE1
    where r.ROLE_O_NAME = '发行登记'
),

-- 角色【运管部信披岗】下的全部用户集合
disc_info_users as (
    select
        u.USER_O_CODE as account
    from sys_rbac_role r
    join sys_rbac_role_relate ur
      on ur.ROLE_CODE = r.ROLE_O_CODE
    join sys_rbac_user u
      on u.R_ID = ur.R_CODE1
    where r.ROLE_O_NAME = '运管部信披'
),

-- 当前节点待办人（ct.account）接收集合
recipient_todo as (
    select
        ba.account,
        ba.CONTENT,
        ba.OPERATOR,
        ba.OPERATOR_ACCOUNT,
        ba.`TRIGGER`,
        ba.PRD_CODE,
        ba.PRD_NAME,
        ba.ALERT_TYPE,
        ba.PRD_DEPT,
        ba.INSTANCE_ID,
        ba.ACTIVITY_ID,
        ba.TASK_NAME,
        ba.NODE_NAME,
        ba.SCENE_NAME,
        ba.USER_NAME,
        ba.IS_PROD_MANAGER,
        ba.DEPT_LEADER_1,
        ba.DEPT_LEADER_2,
        ba.PM_RECEIVER_TYPE,
        ba.START_TIME,
        ba.DAYS_ELAPSED,
        ba.CRITICAL_DT
    from base_alerts ba
),

-- 流程发起人接收集合（受 SEND_INITIATOR 控制）
recipient_initiator as (
    select
        ui.CREATE_USER as account,
        ba.CONTENT,
        ba.OPERATOR,
        ba.OPERATOR_ACCOUNT,
        ba.`TRIGGER`,
        ba.PRD_CODE,
        ba.PRD_NAME,
        ba.ALERT_TYPE,
        ba.PRD_DEPT,
        ba.INSTANCE_ID,
        ba.ACTIVITY_ID,
        ba.TASK_NAME,
        ba.NODE_NAME,
        ba.SCENE_NAME,
        ba.USER_NAME,
        ba.IS_PROD_MANAGER,
        ba.DEPT_LEADER_1,
        ba.DEPT_LEADER_2,
        ba.PM_RECEIVER_TYPE,
        ba.START_TIME,
        ba.DAYS_ELAPSED,
        ba.CRITICAL_DT
    from base_alerts ba
    join unfinished_instance ui
      on ui.INSTANCE_ID = ba.INSTANCE_ID
    where ba.SEND_INITIATOR = 1
),

-- 流程中产品经理 / 投资助理 / 销售经理接收集合（FLOW_PM 场景）
recipient_inv_sales_flow as (
    select
        pr.PM_MANAGER as account,
        ba.CONTENT,
        ba.OPERATOR,
        ba.OPERATOR_ACCOUNT,
        ba.`TRIGGER`,
        ba.PRD_CODE,
        ba.PRD_NAME,
        ba.ALERT_TYPE,
        ba.PRD_DEPT,
        ba.INSTANCE_ID,
        ba.ACTIVITY_ID,
        ba.TASK_NAME,
        ba.NODE_NAME,
        ba.SCENE_NAME,
        ba.USER_NAME,
        ba.IS_PROD_MANAGER,
        ba.DEPT_LEADER_1,
        ba.DEPT_LEADER_2,
        ba.PM_RECEIVER_TYPE,
        ba.START_TIME,
        ba.DAYS_ELAPSED,
        ba.CRITICAL_DT
    from base_alerts ba
    join prd_roles pr
      on pr.PRD_CODE = ba.PRD_CODE
    where ba.SEND_INV_SALES = 1
      and pr.PM_MANAGER is not null

    union all

    select
        pr.SECOND_MANAGER as account,
        ba.CONTENT,
        ba.OPERATOR,
        ba.OPERATOR_ACCOUNT,
        ba.`TRIGGER`,
        ba.PRD_CODE,
        ba.PRD_NAME,
        ba.ALERT_TYPE,
        ba.PRD_DEPT,
        ba.INSTANCE_ID,
        ba.ACTIVITY_ID,
        ba.TASK_NAME,
        ba.NODE_NAME,
        ba.SCENE_NAME,
        ba.USER_NAME,
        ba.IS_PROD_MANAGER,
        ba.DEPT_LEADER_1,
        ba.DEPT_LEADER_2,
        ba.PM_RECEIVER_TYPE,
        ba.START_TIME,
        ba.DAYS_ELAPSED,
        ba.CRITICAL_DT
    from base_alerts ba
    join prd_roles pr
      on pr.PRD_CODE = ba.PRD_CODE
    where ba.SEND_INV_SALES = 1
      and pr.SECOND_MANAGER is not null

    union all

    select
        pr.SALES_MANAGER as account,
        ba.CONTENT,
        ba.OPERATOR,
        ba.OPERATOR_ACCOUNT,
        ba.`TRIGGER`,
        ba.PRD_CODE,
        ba.PRD_NAME,
        ba.ALERT_TYPE,
        ba.PRD_DEPT,
        ba.INSTANCE_ID,
        ba.ACTIVITY_ID,
        ba.TASK_NAME,
        ba.NODE_NAME,
        ba.SCENE_NAME,
        ba.USER_NAME,
        ba.IS_PROD_MANAGER,
        ba.DEPT_LEADER_1,
        ba.DEPT_LEADER_2,
        ba.PM_RECEIVER_TYPE,
        ba.START_TIME,
        ba.DAYS_ELAPSED,
        ba.CRITICAL_DT
    from base_alerts ba
    join prd_roles pr
      on pr.PRD_CODE = ba.PRD_CODE
    where ba.SEND_INV_SALES = 1
      and pr.SALES_MANAGER is not null
),

recipient_inv_sales_allpm as (
    select
        u.account,
        ba.CONTENT,
        ba.OPERATOR,
        ba.OPERATOR_ACCOUNT,
        ba.`TRIGGER`,
        ba.PRD_CODE,
        ba.PRD_NAME,
        ba.ALERT_TYPE,
        ba.PRD_DEPT,
        ba.INSTANCE_ID,
        ba.ACTIVITY_ID,
        ba.TASK_NAME,
        ba.NODE_NAME,
        ba.SCENE_NAME,
        ba.USER_NAME,
        ba.IS_PROD_MANAGER,
        ba.DEPT_LEADER_1,
        ba.DEPT_LEADER_2,
        ba.PM_RECEIVER_TYPE,
        ba.START_TIME,
        ba.DAYS_ELAPSED,
        ba.CRITICAL_DT
    from base_alerts ba
    join product_mgmt_users u
      on 1 = 1
    where ba.SEND_INV_SALES = 1
      and ba.PM_RECEIVER_TYPE = 'ALL_PM'
),

recipient_flow_pm as (
    select
        ct.account,
        ba.CONTENT,
        ba.OPERATOR,
        ba.OPERATOR_ACCOUNT,
        ba.`TRIGGER`,
        ba.PRD_CODE,
        ba.PRD_NAME,
        ba.ALERT_TYPE,
        ba.PRD_DEPT,
        ba.INSTANCE_ID,
        ba.ACTIVITY_ID,
        ba.TASK_NAME,
        ba.NODE_NAME,
        ba.SCENE_NAME,
        ba.USER_NAME,
        ba.IS_PROD_MANAGER,
        ba.DEPT_LEADER_1,
        ba.DEPT_LEADER_2,
        ba.PM_RECEIVER_TYPE,
        ba.START_TIME,
        ba.DAYS_ELAPSED,
        ba.CRITICAL_DT
    from base_alerts ba
    join current_todo ct
      on ct.INSTANCE_ID = ba.INSTANCE_ID
     and ct.IS_PROD_MANAGER = 1
    where ba.PM_RECEIVER_TYPE = 'FLOW_PM'
),

recipient_sb_register as (
    select
        u.account,
        ba.CONTENT,
        ba.OPERATOR,
        ba.OPERATOR_ACCOUNT,
        ba.`TRIGGER`,
        ba.PRD_CODE,
        ba.PRD_NAME,
        ba.ALERT_TYPE,
        ba.PRD_DEPT,
        ba.INSTANCE_ID,
        ba.ACTIVITY_ID,
        ba.TASK_NAME,
        ba.NODE_NAME,
        ba.SCENE_NAME,
        ba.USER_NAME,
        ba.IS_PROD_MANAGER,
        ba.DEPT_LEADER_1,
        ba.DEPT_LEADER_2,
        ba.PM_RECEIVER_TYPE,
        ba.START_TIME,
        ba.DAYS_ELAPSED,
        ba.CRITICAL_DT
    from base_alerts ba
    join sb_register_users u
      on 1 = 1
    where ba.SEND_SB_REGISTER = 1
),

-- 发行登记岗位接收集合（受 SEND_ISSUE_REGISTER 控制）
recipient_issue_register as (
    select
        u.account,
        ba.CONTENT,
        ba.OPERATOR,
        ba.OPERATOR_ACCOUNT,
        ba.`TRIGGER`,
        ba.PRD_CODE,
        ba.PRD_NAME,
        ba.ALERT_TYPE,
        ba.PRD_DEPT,
        ba.INSTANCE_ID,
        ba.ACTIVITY_ID,
        ba.TASK_NAME,
        ba.NODE_NAME,
        ba.SCENE_NAME,
        ba.USER_NAME,
        ba.IS_PROD_MANAGER,
        ba.DEPT_LEADER_1,
        ba.DEPT_LEADER_2,
        ba.PM_RECEIVER_TYPE,
        ba.START_TIME,
        ba.DAYS_ELAPSED,
        ba.CRITICAL_DT
    from base_alerts ba
    join issue_register_users u
      on 1 = 1
    where ba.SEND_ISSUE_REGISTER = 1
),

-- 运管部信披岗位接收集合（受 SEND_DISC_INFO 控制）
recipient_disc_info as (
    select
        u.account,
        ba.CONTENT,
        ba.OPERATOR,
        ba.OPERATOR_ACCOUNT,
        ba.`TRIGGER`,
        ba.PRD_CODE,
        ba.PRD_NAME,
        ba.ALERT_TYPE,
        ba.PRD_DEPT,
        ba.INSTANCE_ID,
        ba.ACTIVITY_ID,
        ba.TASK_NAME,
        ba.NODE_NAME,
        ba.SCENE_NAME,
        ba.USER_NAME,
        ba.IS_PROD_MANAGER,
        ba.DEPT_LEADER_1,
        ba.DEPT_LEADER_2,
        ba.PM_RECEIVER_TYPE,
        ba.START_TIME,
        ba.DAYS_ELAPSED,
        ba.CRITICAL_DT
    from base_alerts ba
    join disc_info_users u
      on 1 = 1
    where ba.SEND_DISC_INFO = 1
),

-- 所有接收人集合汇总，后续统一去重与黑名单过滤
all_recipients as (
    select * from recipient_todo
    union all
    select * from recipient_initiator
    union all
    select * from recipient_inv_sales_flow
    union all
    select * from recipient_inv_sales_allpm
    union all
    select * from recipient_flow_pm
    union all
    select * from recipient_sb_register
    union all
    select * from recipient_issue_register
    union all
    select * from recipient_disc_info
)

select distinct
    ar.account,
    ar.CONTENT,
    ar.OPERATOR,
    ar.OPERATOR_ACCOUNT,
    ar.`TRIGGER`,
    ar.PRD_CODE,
    ar.PRD_NAME,
    ar.ALERT_TYPE,
    ar.PRD_DEPT,
    ar.INSTANCE_ID,
    ar.ACTIVITY_ID,
    ar.TASK_NAME,
    ar.NODE_NAME,
    ar.SCENE_NAME,
    ar.USER_NAME,
    ar.IS_PROD_MANAGER,
    ar.DEPT_LEADER_1,
    ar.DEPT_LEADER_2,
    ar.PM_RECEIVER_TYPE,
    ar.START_TIME,
    ar.DAYS_ELAPSED,
    ar.CRITICAL_DT
from all_recipients ar
where not exists (select 1 from excluded_leaders el where el.account = ar.account)
order by
    ar.PRD_NAME,
    ar.ALERT_TYPE
