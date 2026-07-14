<template>
  <div style="display: flex; gap: 8px; align-items: center; flex-wrap: wrap;">
    <!-- 只选模式：测试用例页签 -->
    <n-select
      v-if="mode === 'select'"
      :value="modelValue"
      :options="caseOptions"
      placeholder="选择用例"
      style="width: 300px;"
      size="small"
      @update:value="handleUpdate"
    />
    <!-- 可输入模式：保存到用例弹窗 -->
    <n-auto-complete
      v-else
      :value="modelValue"
      :options="caseOptions"
      placeholder="选择或输入用例名称"
      style="width: 300px;"
      size="small"
      clearable
      @update:value="handleUpdate"
    />
    <slot />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{
  modelValue: string
  caseList: any[]
  mode?: 'select' | 'editable'
}>(), {
  mode: 'select',
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()

const caseOptions = computed(() =>
  props.caseList.map((c: any) => ({ label: `${c.name} (${c.message_count}条)`, value: c.name }))
)

function handleUpdate(value: string) {
  emit('update:modelValue', value)
}
</script>
