<!--
  ComboSelect - 组合选择组件
  支持多选组合值，输入输出格式为数组
-->
<script setup lang="ts">
import { NSelect } from 'naive-ui'

const props = withDefaults(defineProps<{
  modelValue?: any[]
}>(), {
  modelValue: () => [],
})

const emit = defineEmits<{
  'update:modelValue': [value: any[]]
}>()

function handleUpdate(values: string[]) {
  const parsed = values.map(v => {
    const num = Number(v)
    return isNaN(num) ? v : num
  })
  emit('update:modelValue', parsed)
}
</script>

<template>
  <div style="display: flex; gap: 8px; align-items: center; width: 100%;">
    <span style="font-size: 13px; color: var(--n-text-color-2); white-space: nowrap; min-width: 60px;">组合:</span>
    <n-select
      style="flex: 1; min-width: 200px;"
      placeholder="请选择或输入组合值"
      size="small"
      filterable
      multiple
      tag
      :value="modelValue.map(String)"
      @update:value="handleUpdate"
    />
  </div>
</template>
