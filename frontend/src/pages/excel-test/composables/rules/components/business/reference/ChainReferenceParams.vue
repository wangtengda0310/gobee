<!--
  ChainReferenceParams - 跨表关系链检查规则参数组件

  三列布局：左（来源链）| 中（比较操作）| 右（目标链）
  使用 CSS Grid 行对齐步骤卡片

  ⚠️ 响应式约束（修改本文件时务必遵守）：
  - leftSteps / rightSteps 必须用 ref() 包裹，
    否则 computed(maxSteps) 无法追踪 length 变化，导致"添加步骤"按钮点击后无视觉反馈。
  - script 中访问这些变量必须通过 .value（如 leftSteps.value.push），
    模板中 Vue 会自动解包 ref，不需要 .value。
  - syncLeft/syncRight 将响应式数据序列化回 props.params（外部可变对象），
    传入 serializeSteps 时必须传 .value 以匹配 ChainStep[] 参数类型。
-->
<script setup lang="ts">
import {ref, computed} from "vue"
import {NSelect, NInput} from "naive-ui"
import {ERuleParam} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
import {
    compareTypeOptions, warnBeforeOptions,
    createEmptyStep, createEmptySideConfig, parseSteps, serializeSteps,
} from '../../../chain-reference-params'
import ChainStepCard from './ChainStepCard.vue'

const props = defineProps<{ params: { [p: string]: string } }>()

// 响应式比较类型（直接修改 params 不会触发重渲染，必须用 ref）
const matchCompareType = ref(props.params['chainMatchCompare'] || '')
const compareType = ref(props.params['chainCompare'] || 'match')

// 预警配置响应式状态
const warnBefore = ref(props.params['chainWarnBefore'] || '')
const warnSheet = ref(props.params['chainWarnSheet'] || '')
const warnCol = ref(props.params['chainWarnCol'] || '')
const warnCustom = ref('')

// 初始化默认参数（写入 props.params 供外部读取）
const params = props.params || {}
if (!(ERuleParam.ALLOW_EMPTY in params)) params[ERuleParam.ALLOW_EMPTY] = 'false'
if (!(ERuleParam.EXCEPTS in params)) params[ERuleParam.EXCEPTS] = ''
if (!(ERuleParam.ALLOW_COMMIT in params)) params[ERuleParam.ALLOW_COMMIT] = 'true'
if (!(ERuleParam.BREAK_LINE in params)) params[ERuleParam.BREAK_LINE] = '3'
if (!('chainCompare' in params)) params['chainCompare'] = 'match'
if (!('chainSteps' in params) || !params['chainSteps']) {
    params['chainSteps'] = JSON.stringify({left: createEmptySideConfig(), right: createEmptySideConfig()})
}

// ⚠️ 必须用 ref() 包裹，否则 computed 无法追踪变化，按钮操作无视觉反馈
const leftSteps = ref(parseSteps(params, 'left'))
const rightSteps = ref(parseSteps(params, 'right'))

// 将响应式数据序列化回 params（注意传 .value）
const syncLeft = () => serializeSteps(params, 'left', leftSteps.value)
const syncRight = () => serializeSteps(params, 'right', rightSteps.value)

const maxSteps = computed(() => Math.max(leftSteps.value.length, rightSteps.value.length))
// Row 1: 标题行, Row 2: 预警配置行, Row 3+: 步骤行
const gridTemplateRows = computed(() => ['auto', 'auto', ...Array(maxSteps.value).fill('1fr')].join(' '))

// 计算步骤的 grid-row：最后一行对齐到 maxSteps+2（最后一行），保证左右链最后一步水平对齐
// 因为插入了预警配置行（row 2），步骤行从 row 3 开始
function stepGridRow(i: number, length: number) {
  return (i === length - 1 && length < maxSteps.value) ? maxSteps.value + 2 : i + 3
}

function onMatchCompareTypeChange(v: string | null) {
    matchCompareType.value = v || ''
    params['chainMatchCompare'] = v || ''
    syncLeft()
    syncRight()
}
function onCompareTypeChange(v: string) {
    compareType.value = v
    params['chainCompare'] = v
    syncLeft()
    syncRight()
}

// 预警配置联动逻辑
const warnEnabled = computed(() => warnBefore.value !== '')
const warnIsCustom = computed(() => warnBefore.value === '__custom__')

function onWarnBeforeChange(v: string | null) {
    const val = v || ''
    warnBefore.value = val
    params['chainWarnBefore'] = val
    // 选择"不启用"时清空表名和列名
    if (val === '') {
        warnSheet.value = ''
        warnCol.value = ''
        params['chainWarnSheet'] = ''
        params['chainWarnCol'] = ''
    }
}
function onWarnSheetChange(v: string) {
    warnSheet.value = v
    params['chainWarnSheet'] = v
}
function onWarnColChange(v: string) {
    warnCol.value = v
    params['chainWarnCol'] = v
}
function onWarnCustomChange(v: string) {
    warnCustom.value = v
    params['chainWarnBefore'] = v
}

// 统一处理步骤更新事件
function onLeftStepUpdate(i: number, field: string, value: string) {
    if (field === '__insert__') { leftSteps.value.splice(i, 0, createEmptyStep()) }
    else if (field === '__append__') { leftSteps.value.splice(i + 1, 0, createEmptyStep()) }
    else if (field === '__remove__') { if (leftSteps.value.length > 1) leftSteps.value.splice(i, 1) }
    else { (leftSteps.value[i] as any)[field] = value }
    syncLeft()
}
function onRightStepUpdate(i: number, field: string, value: string) {
    if (field === '__insert__') { rightSteps.value.splice(i, 0, createEmptyStep()) }
    else if (field === '__append__') { rightSteps.value.splice(i + 1, 0, createEmptyStep()) }
    else if (field === '__remove__') { if (rightSteps.value.length > 1) rightSteps.value.splice(i, 1) }
    else { (rightSteps.value[i] as any)[field] = value }
    syncRight()
}
</script>

<template>
  <div :style="`display: grid; grid-template-columns: 1fr minmax(160px, auto) 1fr; grid-template-rows: ${gridTemplateRows}; gap: 8px; align-items: stretch`">
    <!-- 标题行 -->
    <div style="grid-column: 1; grid-row: 1; font-weight: bold; font-size: 14px; padding: 4px 0">来源链 (left)</div>
    <div style="grid-column: 2; grid-row: 1; font-weight: bold; font-size: 12px; text-align: center; padding: 4px 0; color: white">比较</div>
    <div style="grid-column: 3; grid-row: 1; font-weight: bold; font-size: 14px; padding: 4px 0">目标链 (right)</div>

    <!-- 预警配置行（跨三列，位于标题行下方） -->
    <div style="grid-column: 1 / -1; grid-row: 2; display: flex; gap: 12px; align-items: center; padding: 8px 12px; border: 1px dashed rgba(255,200,0,0.3); border-radius: 6px; background: rgba(255,200,0,0.03)">
      <span style="font-weight: bold; font-size: 13px; color: rgba(255,200,0,0.8)">预警</span>
      <n-select size="small" style="width: 120px" :value="warnBefore" :options="warnBeforeOptions" @update:value="onWarnBeforeChange" />
      <n-input v-if="warnIsCustom" size="small" placeholder="如 48h" style="width: 100px" :value="warnCustom" @update:value="onWarnCustomChange" />
      <span style="font-size: 12px; color: rgba(255,255,255,0.5)">时间来源:</span>
      <n-input size="small" placeholder="表名" style="width: 140px" :disabled="!warnEnabled" :value="warnSheet" @update:value="onWarnSheetChange" />
      <n-input size="small" placeholder="列名" style="width: 100px" :disabled="!warnEnabled" :value="warnCol" @update:value="onWarnColChange" />
    </div>

    <!-- 中间列: 比较卡片（跨所有步骤行，从 row 3 开始） -->
    <div :style="`grid-column: 2; grid-row: 3 / -1; display: flex; flex-direction: column; justify-content: space-between`">
      <div style="border: 2px solid rgba(64,128,255,0.3); border-radius: 6px; padding: 10px; background: rgba(64,128,255,0.03); text-align: center">
        <div style="font-weight: bold; font-size: 13px; margin-bottom: 4px">比较规则</div>
        <div style="font-size: 11px; color: rgba(255,255,255,0.5); margin-bottom: 4px">操作第一步数据</div>
        <n-select style="width: 100%; min-width: 140px" size="small" :value="compareType" :options="compareTypeOptions" @update:value="onCompareTypeChange" />
      </div>
      <div style="border: 2px solid rgba(64,128,255,0.4); border-radius: 6px; padding: 10px; background: rgba(64,128,255,0.05); text-align: center">
        <div style="font-weight: bold; font-size: 13px; margin-bottom: 4px">匹配规则</div>
        <div style="font-size: 11px; color: rgba(255,255,255,0.5); margin-bottom: 4px">操作最后一步数据</div>
        <n-select style="width: 100%; min-width: 140px" size="small" placeholder="不使用" clearable :value="matchCompareType || null" :options="compareTypeOptions" @update:value="onMatchCompareTypeChange" />
      </div>
    </div>

    <!-- 左右列: 步骤卡片 -->
    <template v-for="(_, i) in maxSteps" :key="i">
      <div v-if="i < leftSteps.length" :style="`grid-column: 1; grid-row: ${stepGridRow(i, leftSteps.length)}`">
        <ChainStepCard :step="leftSteps[i]" :index="i" :total="leftSteps.length" side="left" @update="(f, v) => onLeftStepUpdate(i, f, v)" />
      </div>
      <div v-if="i < rightSteps.length" :style="`grid-column: 3; grid-row: ${stepGridRow(i, rightSteps.length)}`">
        <ChainStepCard :step="rightSteps[i]" :index="i" :total="rightSteps.length" side="right" @update="(f, v) => onRightStepUpdate(i, f, v)" />
      </div>
    </template>
  </div>
</template>
