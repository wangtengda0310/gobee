<!--
  RegexFormatRow - 正则格式选择 + 是否为数组 共用组件

  两种数据绑定模式：
  1. params 模式（默认）：传入 params 对象，自动读写 params[ERuleParam.PATTERN]/[GROUPS]/['isArray']
  2. v-model 模式：传入 pattern/groups/isArray props + 监听 update:* 事件

  被 CrossReferenceParams、ChainStepCard 等需要正则提取的规则参数组件复用。
-->
<script setup lang="ts">
import {computed} from "vue"
import {NCheckbox, NSelect} from "naive-ui"
import {ERuleParam} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
import {regexOptions, findOptionByRegex} from '../chain-reference-params'

const props = withDefaults(defineProps<{
  /** params 对象（params 模式） */
  params?: Record<string, string>
  /** 当前正则值（v-model 模式） */
  pattern?: string
  /** 当前捕获组（v-model 模式） */
  groups?: string
  /** 当前是否为数组（v-model 模式） */
  isArray?: string
  /** 是否显示"是否为数组"复选框，默认 true */
  showIsArray?: boolean
}>(), {
  showIsArray: true,
})

const emit = defineEmits<{
  'update:pattern': [value: string]
  'update:groups': [value: string]
  'update:isArray': [value: string]
}>()

// 判断是否使用 params 模式
const useParamsMode = computed(() => !!props.params)

// params 模式下的初始化
if (props.params) {
  if (!(ERuleParam.PATTERN in props.params)) props.params[ERuleParam.PATTERN] = ''
  if (!(ERuleParam.GROUPS in props.params)) props.params[ERuleParam.GROUPS] = ''
  if (!('isArray' in props.params)) props.params['isArray'] = 'false'
}

// 统一读取：优先 props 值，否则从 params 读取
const currentPattern = computed(() => {
  if (!useParamsMode.value) return props.pattern ?? ''
  return props.params![ERuleParam.PATTERN] || ''
})

const currentGroups = computed(() => {
  if (!useParamsMode.value) return props.groups ?? ''
  return props.params![ERuleParam.GROUPS] || ''
})

const currentIsArray = computed(() => {
  if (!useParamsMode.value) return props.isArray ?? 'false'
  return props.params!['isArray'] || 'false'
})

const selectedRegexOption = computed(() => {
  return findOptionByRegex(currentPattern.value, currentGroups.value)
})

function onRegexSelect(_v: number, option: any) {
  const p = option.regex || ''
  const g = option.groups || ''
  if (useParamsMode.value) {
    props.params![ERuleParam.PATTERN] = p
    props.params![ERuleParam.GROUPS] = g
  } else {
    emit('update:pattern', p)
    emit('update:groups', g)
  }
}

function onIsArrayChange(val: boolean) {
  const v = val.toString()
  if (useParamsMode.value) {
    props.params!['isArray'] = v
  } else {
    emit('update:isArray', v)
  }
}
</script>

<template>
  <div style="display: flex; gap: 10px; align-items: center">
    <div>正则格式:</div>
    <n-select
      style="flex: 1 1 0"
      placeholder="选择正则提取格式"
      :value="selectedRegexOption"
      :options="regexOptions"
      @update:value="onRegexSelect"
    />
    <n-checkbox
      v-if="showIsArray"
      style="flex: 0 0 100px"
      :checked="currentIsArray === 'true'"
      @update:checked="onIsArrayChange"
    >
      是否为数组
    </n-checkbox>
  </div>
</template>
