<!--
  variable-select.vue — 动态变量选择组件
  位置关系：field-item.vue 内嵌组件（与 range-input.vue 同级）
  调用链：field-item.vue → variable-select.vue → variable-select.requirement.ts → GetAvailableVariables()
-->
<template>
  <div class="variable-select">
    <label class="var-label">选择变量:</label>
    <n-select
      v-model:value="selectedVar"
      :options="varOptions"
      :loading="loading"
      placeholder="选择变量"
      size="small"
      style="flex: 1; min-width: 0;"
      data-testid="variable-select-dropdown"
      @update:value="handleVarChange"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { NSelect } from 'naive-ui'
import {
  createWailsVariableService,
  toVarOptions,
  type VarOption,
} from './variable-select.requirement'

const props = withDefaults(defineProps<{
  modelValue?: string
  /** 当前 Req 的 msg_name，用于按 AvailableReqs 过滤变量选项 */
  msgName?: string
}>(), {
  modelValue: '',
  msgName: '',
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
  'change': []
}>()

// 加载状态
const loading = ref(false)

// 选中值
const selectedVar = ref<string | null>(props.modelValue || null)

// 选项列表
const varOptions = ref<VarOption[]>([])

// 加载可用变量列表，按当前 Req 的 msg_name 过滤：
// available_reqs 为空=对所有 Req 可用；非空=仅对该列表中的 Req 可用
async function loadVariables() {
  loading.value = true
  try {
    const svc = createWailsVariableService()
    const items = await svc.getAvailableVariables()
    const filtered = items.filter(it => !it.available_reqs || it.available_reqs.length === 0 || it.available_reqs.includes(props.msgName))
    varOptions.value = toVarOptions(filtered)
  } catch (e) {
    console.warn('加载可用变量列表失败:', e)
    varOptions.value = []
  } finally {
    loading.value = false
  }
}

// 处理变量选择变更
function handleVarChange(value: string | null) {
  const v = value ?? ''
  emit('update:modelValue', v)
  emit('change')
}

// 组件挂载时加载变量列表
onMounted(() => {
  loadVariables()
})

// 切换到不同 Req 时（同 key field-item 复用、msgName 变化）重新加载过滤
watch(() => props.msgName, () => {
  loadVariables()
})
</script>

<style scoped>
.variable-select {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
}

.var-label {
  font-size: 13px;
  color: var(--n-text-color-2);
  white-space: nowrap;
}
</style>
