/**
 * Excel 规则模板配置
 *
 * 定义规则选项和规则参数组件映射
 */
import {type Component, type VNode, h} from "vue";
import {EColRule} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule";

/** 规则参数组件映射值类型 */
type RuleComponentEntry = Component | ((props: { params: { [p: string]: string } }) => VNode)

// 测试规则
import CustomParams from "./rules/components/test/CustomParams.vue"
// 基础规则
import AllBaseParams from "./rules/components/AllBaseParams.vue"
// 数据类型
import CellTypeCheckParams from "./rules/components/data-type/CellTypeCheckParams.vue"
import DateParams from "./rules/components/data-type/DateParams.vue"
import SpecialFormatParams from "./rules/components/data-type/SpecialFormatParams.vue"
// 业务关系规则 / 日期
import DateDurationParams from "./rules/components/business/date/DateDurationParams.vue"
import DateRangeParams from "./rules/components/business/date/DateRangeParams.vue"
// 业务关系规则 / 数值
import NumericRangeParams from "./rules/components/business/numeric/NumericRangeParams.vue"
import EnumParams from "./rules/components/business/numeric/EnumParams.vue"
// 业务关系规则 / 拼音&中文
import PinYinChsParams from "./rules/components/business/pinyin/PinYinChsParams.vue"
// 业务关系规则 / 关联表
import CrossReferenceParams from "./rules/components/business/reference/CrossReferenceParams.vue"
import ChainReferenceParams from "./rules/components/business/reference/ChainReferenceParams.vue"
import RegexParams from "./rules/components/business/reference/RegexParams.vue"
// 业务关系规则 / 数值计算
import WeightSumParams from "./rules/components/business/calculation/WeightSumParams.vue"
import DateConsistencyParams from "./rules/components/business/calculation/DateConsistencyParams.vue"
// 特殊
import ServerOrClientParams from "./rules/components/ServerOrClientParams.vue"
// 共享
import SimpleParams from "./rules/components/SimpleParams.vue"
// 资源检查
import ResourceParams from "./rules/components/ResourceParams.vue"

/** 为需要额外 props 的组件创建包装函数（注入 defaults） */
function withDefaults(component: Component, defaults: Record<string, string>) {
    return (props: { params: { [p: string]: string } }) => h(component, { params: props.params, defaults })
}

/** 规则选项配置 - 用于级联选择器 */
export const ruleOptions = [
    {
        label: '测试规则', value: '测试规则', children: [
            // EColRule.TEST — 测试
            {label: '测试', value: EColRule.TEST},
        ]
    },
    {
        // EColRule.ALL_BASE — 全基础规则
        label: '基础规则', value: EColRule.ALL_BASE
    },
    {
        // EColRule.RESOURCE — 资源路径校验
        label: '资源路径校验', value: EColRule.RESOURCE
    },
    {
        label: '数据类型', value: '数据类型', children: [
            // EColRule.NUMERIC — 检测数值类型
            {label: '检测数值类型', value: EColRule.NUMERIC,},
            // EColRule.DATE — 检测日期类型
            {label: '检测日期类型', value: EColRule.DATE,},
            // EColRule.BOOLEAN — 检测布尔类型
            {label: '检测布尔类型', value: EColRule.BOOLEAN,},
            // EColRule.STRING — 检测字符串类型（单元格应为字符串）
            {label: '检测字符串类型', value: EColRule.STRING,},
            // EColRule.RICH_TEXT — 富文本格式类型
            {label: '富文本格式类型', value: EColRule.RICH_TEXT,},
            // EColRule.SPECIAL_FORMAT — 特殊格式{id;count}
            {label: '特殊格式{id;count}', value: EColRule.SPECIAL_FORMAT,},
        ]
    },
    {
        label: '业务关系规则', value: '业务关系规则', children: [
            {
                label: '日期', value: '日期', children: [
                    // EColRule.DATE_DURATION — 检测日期跨度
                    {label: '检测日期跨度', value: EColRule.DATE_DURATION,},
                    // EColRule.DATE_RANGE — 检测日期范围
                    {label: '检测日期范围', value: EColRule.DATE_RANGE,},
                ]
            },
            {
                label: '数值', value: '数值', children: [
                    // EColRule.NUMERIC_RANGE — 范围内数
                    {label: '范围内数', value: EColRule.NUMERIC_RANGE,},
                    // EColRule.ENUM — 固定枚举
                    {label: '固定枚举', value: EColRule.ENUM,},
                ]
            },
            {
                label: '拼音&中文', value: '拼音&中文', children: [
                    // EColRule.PIN_YIN_CHS — 拼音列包含中文列转拼音的结果
                    {label: '拼音列包含中文列转拼音的结果', value: EColRule.PIN_YIN_CHS,},
                ]
            },
            {
                label: '关联表', value: '关联表', children: [
                    // EColRule.CROSS_REFERENCE — 跨表检查
                    {label: '跨表检查', value: EColRule.CROSS_REFERENCE,},
                    // EColRule.CHAIN_REFERENCE — 关系链检查
                    {label: '关系链检查', value: EColRule.CHAIN_REFERENCE,},
                    // EColRule.REGEX — 自定义正则
                    {label: '自定义正则', value: EColRule.REGEX,},
                ]
            },
            {
                label: '数值计算', value: '数值计算', children: [
                    // EColRule.WEIGHT_SUM — 单元格总和
                    {label: '单元格总和', value: EColRule.WEIGHT_SUM,},
                    // EColRule.DATE_CONSISTENCY — 配置时间与描述时间一致
                    {label: '配置时间与描述时间一致', value: EColRule.DATE_CONSISTENCY,},
                ]
            }
        ]
    },
]

/**
 * 规则参数组件映射 - 根据规则类型渲染对应的参数配置组件
 *
 * 值为 Component 或 (props) => VNode 的包装函数
 */
export const ruleComponents = new Map<EColRule, RuleComponentEntry>([
    // EColRule.TEST — 测试规则
    [EColRule.TEST, CustomParams],
    // EColRule.ALL_BASE — 全基础规则
    [EColRule.ALL_BASE, AllBaseParams],
    // EColRule.INCREASE_ID — 自增ID检查
    [EColRule.INCREASE_ID, withDefaults(SimpleParams, { allowEmpty: 'false' })],
    // EColRule.NOT_EMPTY — 非空检查
    [EColRule.NOT_EMPTY, withDefaults(SimpleParams, { allowEmpty: 'false' })],
    // EColRule.UNIQUE — 唯一性检查
    [EColRule.UNIQUE, withDefaults(SimpleParams, { allowEmpty: 'false' })],
    // EColRule.CHS_ONLY — 纯中文检查
    [EColRule.CHS_ONLY, withDefaults(SimpleParams, { allowEmpty: 'false' })],
    // EColRule.SERVER_OR_CLIENT — 服务端/客户端标识检查
    [EColRule.SERVER_OR_CLIENT, ServerOrClientParams],
    // EColRule.NUMERIC — 数值类型检查
    [EColRule.NUMERIC, withDefaults(SimpleParams, { allowEmpty: 'true' })],
    // EColRule.DATE — 日期类型检查
    [EColRule.DATE, DateParams],
    // EColRule.BOOLEAN — 布尔类型检查
    [EColRule.BOOLEAN, withDefaults(SimpleParams, { allowEmpty: 'false' })],
    // EColRule.STRING — 单元格应为字符串（参数组件：CellTypeCheckParams.vue）
    [EColRule.STRING, CellTypeCheckParams],
    // EColRule.DATE_DURATION — 日期跨度检查
    [EColRule.DATE_DURATION, DateDurationParams],
    // EColRule.DATE_RANGE — 日期范围检查
    [EColRule.DATE_RANGE, DateRangeParams],
    // EColRule.NUMERIC_RANGE — 数值范围检查
    [EColRule.NUMERIC_RANGE, NumericRangeParams],
    // EColRule.ENUM — 枚举值检查
    [EColRule.ENUM, EnumParams],
    // EColRule.CROSS_REFERENCE — 跨表引用检查
    [EColRule.CROSS_REFERENCE, CrossReferenceParams],
    // EColRule.CHAIN_REFERENCE — 关系链检查
    [EColRule.CHAIN_REFERENCE, ChainReferenceParams],
    // EColRule.SPECIAL_FORMAT — 特殊格式检查
    [EColRule.SPECIAL_FORMAT, SpecialFormatParams],
    // EColRule.REGEX — 自定义正则检查
    [EColRule.REGEX, RegexParams],
    // EColRule.DATE_CONSISTENCY — 配置时间与描述时间一致性检查
    [EColRule.DATE_CONSISTENCY, DateConsistencyParams],
    // EColRule.WEIGHT_SUM — 权重和检查
    [EColRule.WEIGHT_SUM, WeightSumParams],
    // EColRule.RESOURCE — 资源路径检查
    [EColRule.RESOURCE, ResourceParams],
    // EColRule.PIN_YIN_CHS — 拼音中文检查
    [EColRule.PIN_YIN_CHS, PinYinChsParams],
    // EColRule.RICH_TEXT — 富文本格式检查
    [EColRule.RICH_TEXT, withDefaults(SimpleParams, { allowEmpty: 'true' })],
])
