<!--
  FilterConditionRow - 过滤条件统一组件

  两种数据绑定模式：
  1. params 模式：传入 params 对象，自动读写 params.filterCol/filterVal/filterIsArray/filterMode/filterDays
  2. v-model 模式：传入各 prop + 监听 update:* 事件

  被 ChainStepCard、CrossReferenceParams 等需要行条件过滤的规则参数组件复用。
-->
<script setup lang="ts">
import {computed} from "vue"
import {NInput, NSelect, NInputNumber} from "naive-ui"

const props = withDefaults(defineProps<{
  /** params 对象（params 模式） */
  params?: Record<string, string>
  /** 当前过滤列名（v-model 模式） */
  filterCol?: string
  /** 当前过滤值（v-model 模式） */
  filterVal?: string
  /** 当前是否多值（v-model 模式） */
  filterIsArray?: string
  /** 当前过滤模式（v-model 模式） */
  filterMode?: string
  /** 当前距今天数（v-model 模式） */
  filterDays?: string
  /** 前缀标签文字 */
  label?: string
}>(), {
  label: '过滤:',
})

const emit = defineEmits<{
  'update:filterCol': [value: string]
  'update:filterVal': [value: string]
  'update:filterIsArray': [value: string]
  'update:filterMode': [value: string]
  'update:filterDays': [value: string]
}>()

// 判断是否使用 params 模式
const useParamsMode = computed(() => !!props.params)

// params 模式下的初始化
if (props.params) {
  if (!('filterCol' in props.params)) props.params['filterCol'] = ''
  if (!('filterVal' in props.params)) props.params['filterVal'] = ''
  if (!('filterIsArray' in props.params)) props.params['filterIsArray'] = ''
  if (!('filterMode' in props.params)) props.params['filterMode'] = ''
  if (!('filterDays' in props.params)) props.params['filterDays'] = ''
}

// 统一读取
const currentFilterCol = computed(() => {
  if (!useParamsMode.value) return props.filterCol ?? ''
  return props.params!['filterCol'] || ''
})

const currentFilterVal = computed(() => {
  if (!useParamsMode.value) return props.filterVal ?? ''
  return props.params!['filterVal'] || ''
})

const currentFilterMode = computed(() => {
  if (!useParamsMode.value) return props.filterMode ?? ''
  return props.params!['filterMode'] || ''
})

const currentFilterDays = computed(() => {
  if (!useParamsMode.value) return props.filterDays ?? ''
  return props.params!['filterDays'] || ''
})

// 模式选项
const filterModeOptions = [
  {label: '值', value: ''},
  {label: '多值', value: 'multi'},
  {label: '距今<N天', value: 'withinDays'},
]

// 统一写入
function setField(field: string, value: string) {
  if (useParamsMode.value) {
    props.params![field] = value
  } else {
    switch (field) {
      case 'filterCol': emit('update:filterCol', value); break
      case 'filterVal': emit('update:filterVal', value); break
      case 'filterIsArray': emit('update:filterIsArray', value); break
      case 'filterMode': emit('update:filterMode', value); break
      case 'filterDays': emit('update:filterDays', value); break
    }
  }
}

function onModeChange(val: string) {
  setField('filterMode', val)
  // 联动 filterIsArray：multi 模式自动设为 "true"
  if (val === 'multi') {
    setField('filterIsArray', 'true')
  } else {
    setField('filterIsArray', '')
  }
  // 切换到 withinDays 时清空 filterVal
  if (val === 'withinDays') {
    setField('filterVal', '')
  }
}

function onDaysChange(val: number | null) {
  setField('filterDays', val?.toString() ?? '')
}
</script>

<template>
  <div style="display: flex; gap: 8px; align-items: center">
    <div style="flex: 0 0 45px; font-size: 12px">{{ label }}</div>
    <div style="font-size: 12px">列</div>
    <n-input
      style="flex: 2 1 0"
      size="small"
      placeholder="字段名"
      :value="currentFilterCol"
      @update:value="setField('filterCol', $event)"
    />
    <n-select
      style="flex: 0 0 100px"
      size="small"
      :value="currentFilterMode"
      :options="filterModeOptions"
      @update:value="onModeChange"
    />
    <!-- 值/多值模式：显示值输入框 -->
    <n-input
      v-if="currentFilterMode !== 'withinDays'"
      style="flex: 2 1 0"
      size="small"
      :placeholder="currentFilterMode === 'multi' ? '逗号分隔多个值' : '匹配值'"
      :value="currentFilterVal"
      @update:value="setField('filterVal', $event)"
    />
    <!-- withinDays 模式：显示天数输入 -->
    <template v-if="currentFilterMode === 'withinDays'">
      <n-input-number
        style="flex: 1 1 0"
        size="small"
        placeholder="天数"
        :min="1"
        :value="currentFilterDays ? Number(currentFilterDays) : null"
        @update:value="onDaysChange"
      />
      <div style="font-size: 12px; color: #888">天</div>
    </template>
  </div>
</template>
