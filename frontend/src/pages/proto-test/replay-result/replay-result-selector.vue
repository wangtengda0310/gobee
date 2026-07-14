<template>
  <div style="display: flex; gap: 8px; align-items: center;">
    <span style="font-size: 13px; white-space: nowrap;">重放结果:</span>
    <n-select
      v-model:value="selectedId"
      :options="resultOptions"
      placeholder="选择重放结果"
      style="flex: 1; min-width: 200px;"
      size="small"
      clearable
      @update:value="handleChange"
    />
    <n-button size="small" @click="handleClear">清空</n-button>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { NSelect, NButton } from 'naive-ui'
import type { RecordFileData } from '@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/models'

export interface ReplayResult {
  id: string
  source: 'packet' | 'testcase' | 'retry'
  timestamp: string
  recordData: RecordFileData
  status: 'running' | 'completed' | 'error' | 'cancelled'
  error?: string
}

const props = defineProps<{
  results: ReplayResult[]
  currentId: string | null
}>()

const emit = defineEmits<{
  select: [id: string | null]
  clear: []
}>()

const selectedId = computed({
  get: () => props.currentId,
  set: (val) => emit('select', val ?? null),
})

const resultOptions = computed(() => {
  return props.results.map(r => ({
    label: `${r.source === 'packet' ? '发包改包' : r.source === 'testcase' ? '测试用例' : '重发控制'} - ${r.timestamp}`,
    value: r.id,
  }))
})

function handleChange(value: string | null) {
  emit('select', value ?? null)
}

function handleClear() {
  emit('clear')
}
</script>
