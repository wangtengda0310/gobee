/**
 * 跨表关系链检查 - 共享常量和工具函数
 *
 * regexOptions 和 findOptionByRegex 被 cross-reference-params 和 all-base-params 引用
 * 工具函数供 ChainReferenceParams.vue 使用
 */
import {ERuleParam} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"

// --- 共享常量 ---

/** 正则提取选项（-1 = 自定义，匹配时不对应任何预设即为自定义） */
export const regexOptions = [
    {label: '无特定格式, 整个单元格内容匹配', value: 0},
    {
        label: '匹配格式: {123;234} 捕获{}内;分隔后第一个 预计捕获结果: 123',
        value: 1,
        regex: '{(\\d+);\\d+}',
        groups: '1',
    },
    {
        label: '匹配格式: {123;234} 捕获{}内;分隔后第二个 预计捕获结果: 234',
        value: 2,
        regex: '{\\d+;(\\d+)}',
        groups: '1',
    },
    {
        label: '匹配格式: 1,A,鸡, 捕获组: 所有逗号分隔的元素 预计捕获结果: 1 A 鸡',
        value: 3,
        regex: '\\s*([^,\\s]+)\\s*',
        groups: '1',
    },
    {label: '自定义正则表达式', value: -1},
]

/** 比较类型选项（CHAIN_REFERENCE 的 chainCompare/chainMatchCompare 使用） */
export const compareTypeOptions = [
    {label: '验证存在', value: 'verify_exists'},
    {label: '缺失即报错', value: 'verify_must_exist'},
    {label: '时间匹配', value: 'time_overlap'},
    {label: '日期相同', value: 'date_equals'},
    {label: '日期早于或等于', value: 'date_before_or_equal'},
    {label: '日期晚于或等于', value: 'date_after_or_equal'},
]

/** 比较操作选项（CROSS_REFERENCE 的 compareOp 独立维度使用） */
export const compareOpOptions = [
    {label: '默认(精确匹配)', value: ''},
    {label: '完全匹配', value: 'exact_match'},
    {label: '字符串包含', value: 'contains'},
    {label: '日期相同', value: 'date_equals'},
    {label: '日期早于或等于', value: 'date_before_or_equal'},
    {label: '日期晚于或等于', value: 'date_after_or_equal'},
]

/** 值来源选项 */
export const valueSourceOptions = [
    {label: '自身值', value: 'self'},
    {label: '指定列', value: 'col'},
]

/** 预警窗口选项（CHAIN_REFERENCE 的 chainWarnBefore 使用） */
export const warnBeforeOptions = [
    {label: '不启用', value: ''},
    {label: '3天', value: '72h'},
    {label: '7天', value: '168h'},
    {label: '14天', value: '336h'},
    {label: '30天', value: '720h'},
    {label: '自定义', value: '__custom__'},
]

// --- 类型定义 ---

export interface ChainStep {
    sheet: string
    preCol: string
    findVal: string
    nextCol: string
    pattern: string
    groups: string
    filterCol: string
    filterVal: string
    filterIsArray?: string
    filterMode?: string    // 过滤模式: ''(单值) | 'multi'(多值) | 'withinDays'(距今<N天)
    filterDays?: string    // 距今天数
    isArray: string
}

export interface ChainSideConfig {
    steps: ChainStep[]
    compareCol: string
}

export interface ChainPairConfig {
    left: ChainSideConfig
    right: ChainSideConfig
}

// --- 工具函数 ---

export function createEmptyStep(): ChainStep {
    return {sheet: '', preCol: '', findVal: 'col', nextCol: '', pattern: '', groups: '', filterCol: '', filterVal: '', isArray: 'false'}
}

export function createEmptySideConfig(): ChainSideConfig {
    return {steps: [createEmptyStep()], compareCol: ''}
}

export function parseSteps(params: Record<string, string>, side: 'left' | 'right'): ChainStep[] {
    const raw = params['chainSteps']
    if (!raw) return [createEmptyStep()]
    try {
        const parsed = JSON.parse(raw) as ChainPairConfig
        const sideConfig = parsed[side]
        if (!sideConfig || !Array.isArray(sideConfig.steps) || sideConfig.steps.length === 0) return [createEmptyStep()]
        return sideConfig.steps
    } catch {
        return [createEmptyStep()]
    }
}

export function parseCompareCol(params: Record<string, string>, side: 'left' | 'right'): string {
    const raw = params['chainSteps']
    if (!raw) return ''
    try {
        const parsed = JSON.parse(raw) as ChainPairConfig
        return parsed[side]?.compareCol || ''
    } catch {
        return ''
    }
}

export function serializeSteps(params: Record<string, string>, side: 'left' | 'right', steps: ChainStep[], compareCol?: string): void {
    let config: ChainPairConfig
    try {
        const raw = params['chainSteps']
        config = raw ? JSON.parse(raw) as ChainPairConfig : {left: createEmptySideConfig(), right: createEmptySideConfig()}
    } catch {
        config = {left: createEmptySideConfig(), right: createEmptySideConfig()}
    }
    // 左链第一步：强制清空 sheet/preCol（仅取值模式，从当前列取值）
    const cleanSteps = steps.map((s, i) => {
        if (i === 0 && side === 'left') {
            return {...s, sheet: '', preCol: ''}
        }
        return s
    })
    config[side] = {steps: cleanSteps, compareCol: compareCol ?? ''}
    // JSON-in-string：ColRule.Params 统一为 map[string]string（Wails3 binding 限制），
    // 嵌套结构只能序列化为字符串存储，后端需 JSON.parse 二次反序列化
    params['chainSteps'] = JSON.stringify(config)
}

export const findOptionByRegex = (regex: string, groups: string): number => {
    return regexOptions.find(opt => opt.regex === regex && opt.groups === groups)?.value || 0
}

// --- 主组件导出 ---

import ChainReferenceParams from "./components/business/reference/ChainReferenceParams.vue"

export { ChainReferenceParams }
