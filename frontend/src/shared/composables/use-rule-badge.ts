/**
 * 规则覆盖角标渲染工具
 *
 * 提供规则覆盖数据的查询和角标渲染函数，
 * 供 activity-panel.vue 和 DrawSkinCard.vue 共用
 */
import {h} from "vue"
import {NBadge, NPopover, NText} from "naive-ui"
import type {RuleCoverageData, RuleNameInfo} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/activity-wiki-check/models.js"

// EColRule 英文枚举 → 中文名映射（规则 JSON 中 DisplayName 为空时使用）
const colRuleNameMap: Record<string, string> = {
  TEST: '测试',
  INCREASE_ID: '自增ID',
  UNIQUE: '唯一不重复',
  CHS_ONLY: '仅中文',
  NOT_EMPTY: '不为空',
  SERVER_OR_CLIENT: '前后端校验',
  ALL_BASE: '基础规则（全）',
  NUMERIC: '数值类型',
  DATE: '日期格式',
  BOOLEAN: '布尔类型',
  STRING: '单元格应为字符串',
  DATE_RANGE: '日期范围',
  DATE_DURATION: '日期跨度',
  NUMERIC_RANGE: '数值范围',
  ENUM: '固定枚举',
  FOREIGN_KEY: '关联表',
  CROSS_REFERENCE: '跨表引用',
  SPLIT_REGERENCE: '拆分引用',
  SPECIAL_FORMAT: '特殊格式',
  REGEX: '自定义正则',
  WEIGHT_SUM: '权重求和',
  DATE_CONSISTENCY: '日期一致性',
  RESOURCE: '资源检查',
  PIN_YIN_CHS: '拼音汉字匹配',
  RICH_TEXT: '富文本检查',
}

/** 获取规则显示名称：优先 displayName，否则查映射表，最后用英文枚举 */
export const getRuleDisplayName = (type: string, displayName: string) => {
  if (displayName) return displayName
  return colRuleNameMap[type] || type
}

/** 获取 Sheet 的表级规则信息（含错误计数） */
export const getTableRuleInfo = (ruleCoverage: RuleCoverageData | null, sheetName: string) => {
  const stats = ruleCoverage?.sheets?.[sheetName]
  return {count: stats?.tableRuleCount ?? 0, names: stats?.tableRuleNames ?? [], errorCount: stats?.tableErrorCount ?? 0}
}

/** 获取字段级规则信息（含错误计数） */
export const getFieldRuleInfo = (ruleCoverage: RuleCoverageData | null, sheetName: string, fieldName: string) => {
  const stats = ruleCoverage?.sheets?.[sheetName]
  const fieldStat = stats?.fieldRuleStats?.[fieldName]
  return {count: fieldStat?.totalCount ?? 0, names: fieldStat?.ruleNames ?? [], errorCount: fieldStat?.errorCount ?? 0}
}

/** 构建规则列表 VNode（表级 + 字段级共用） */
const buildRuleListVNode = (headerLabel: string, count: number, names: RuleNameInfo[]) => {
  return h('div', {style: 'max-width: 280px;'}, [
    h(NText, {depth: 2, style: 'font-size: 12px; margin-bottom: 4px; display: block;'}, () => `${headerLabel} (${count}条)`),
    ...names.map(r =>
      h('div', {style: 'font-size: 12px; line-height: 1.6; display: flex; align-items: center; gap: 4px;'}, [
        h('span', {style: 'color: #18a058; font-size: 10px;'}, '●'),
        h('span', null, getRuleDisplayName(r.type, r.displayName)),
      ])
    )
  ])
}

/** 渲染带角标的 Tab 标题（hover 弹出规则列表，校验失败变红） */
export const renderTabWithBadge = (ruleCoverage: RuleCoverageData | null, sheetName: string, label: string) => {
  const {count, names, errorCount} = getTableRuleInfo(ruleCoverage, sheetName)

  const tabContent = h('div', {style: 'display: flex; flex-direction: column; align-items: center; line-height: 1.2;'}, [
    h('span', {style: 'font-size: 10px; color: #888;'}, sheetName),
    h('span', {style: 'font-size: 12px;'}, label)
  ])

  if (count === 0) return tabContent

  const ruleList = buildRuleListVNode('表级规则', count, names)
  const badgeColor = errorCount > 0 ? '#d03050' : '#18a058'

  return h(NPopover, {trigger: 'hover', placement: 'top'}, {
    trigger: () => h(NBadge, {value: count, color: badgeColor, size: 'small'}, {default: () => tabContent}),
    default: () => ruleList
  })
}

/** 渲染带角标的字段 label（hover 弹出规则列表，校验失败变红） */
export const renderFieldWithBadge = (ruleCoverage: RuleCoverageData | null, sheetName: string, fieldName: string, label: string) => {
  const {count, names, errorCount} = getFieldRuleInfo(ruleCoverage, sheetName, fieldName)
  const labelContent = h('span', null, label)

  if (count === 0) return labelContent

  const ruleList = buildRuleListVNode('字段规则', count, names)
  const badgeColor = errorCount > 0 ? '#d03050' : '#18a058'

  return h(NPopover, {trigger: 'hover', placement: 'top'}, {
    trigger: () => h(NBadge, {value: count, color: badgeColor, size: 'small'}, {default: () => labelContent}),
    default: () => ruleList
  })
}
