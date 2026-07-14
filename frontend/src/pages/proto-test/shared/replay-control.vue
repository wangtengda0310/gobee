<template>
  <div v-if="selectedEntry" style="margin-top: 12px; border: 1px solid var(--n-border-color); border-radius: 4px; padding: 12px;">
    <div style="font-size: 14px; font-weight: 500; margin-bottom: 8px;">重放控制</div>
    <div style="display: flex; justify-content: space-between; align-items: center; gap: 16px;">
      <!-- 左侧：重发控制 -->
      <div style="display: flex; gap: 8px; align-items: center; flex: 0 0 auto;">
        <template v-if="!interceptMode">
          <n-button type="primary" @click="handleRetry" :loading="running" :disabled="running" size="small">重发</n-button>
          <div style="display: flex; gap: 4px; align-items: center;">
            <n-input-number v-model:value="repeatCount" :min="1" :max="999" style="width: 80px;" size="small" />
            <span style="font-size: 13px; color: var(--n-text-color-3);">次</span>
          </div>
          <n-button type="warning" @click="handleIterativeSend" :loading="running" :disabled="running" size="small">迭代发送</n-button>
          <n-button @click="handleStop" :disabled="!running" size="small">停止</n-button>
        </template>
        <n-tag v-if="progress" :type="statusType" size="small">
          {{ statusLabel }}
        </n-tag>
      </div>

      <!-- 右侧：Ntf 显示（录制/重放场景）；测试用例仅 Req 无 Ack，不显示配对状态 -->
      <div v-if="selectedEntry.ntf" style="flex: 1; display: flex; gap: 8px; align-items: center; font-size: 13px; color: var(--n-text-color-2); min-width: 0;">
        <span style="white-space: nowrap; color: var(--n-text-color-3);">Ntf:</span>
        <span style="overflow: hidden; text-overflow: ellipsis; white-space: nowrap;">{{ selectedEntry.ntf.msg_name }}</span>
        <n-tag v-if="selectedEntry.ntf.payload && Object.keys(selectedEntry.ntf.payload).length > 0" size="tiny" type="info" style="flex-shrink: 0;">
          {{ formatPayload(selectedEntry.ntf.payload) }}
        </n-tag>
      </div>
      <div v-else-if="showPairStatus && selectedEntry.type === 'pair'" style="flex: 1; font-size: 13px; color: var(--n-text-color-3);">
        {{ selectedEntry.ack ? '已配对' : '等待 Ack...' }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { NButton, NTag, NInputNumber } from 'naive-ui'
import type { PairedEntry } from './composables/use-paired-messages'

const props = withDefaults(defineProps<{
  running: boolean
  progress: any
  selectedEntry: PairedEntry | null
  hasIterativeConfig?: boolean
  interceptMode?: boolean
  /** 是否显示 Req/Ack 配对状态（测试用例页签无 Ack，应关闭） */
  showPairStatus?: boolean
}>(), {
  showPairStatus: true,
})

const emit = defineEmits<{
  retry: [count: number]
  stop: []
  'iterative-send': []
}>()

const repeatCount = ref(1)

const statusType = computed(() => {
  if (!props.progress) return 'info'
  const s = props.progress.status
  if (s === 'completed') return 'success'
  if (s === 'error') return 'error'
  if (s === 'cancelled') return 'warning'
  return 'info'
})

const statusLabel = computed(() => {
  if (!props.progress) return ''
  const p = props.progress
  const map: Record<string, string> = {
    running: `运行中 (${p.sent}/${p.total})`,
    completed: `完成 (${p.total}/${p.total})`,
    error: `错误: ${p.error_message}`,
    cancelled: '已取消',
  }
  return map[p.status] || p.status
})

function formatPayload(payload: Record<string, any>): string {
  try {
    if (typeof payload === 'object' && payload !== null) {
      const keys = Object.keys(payload)
      if (keys.length > 0) {
        const firstKey = keys[0]
        const value = payload[firstKey]
        return `${firstKey}: ${JSON.stringify(value).slice(0, 30)}`
      }
    }
    return JSON.stringify(payload).slice(0, 30)
  } catch {
    return JSON.stringify(payload).slice(0, 30)
  }
}

function handleRetry() { emit('retry', repeatCount.value) }
function handleStop() { emit('stop') }
function handleIterativeSend() { emit('iterative-send') }
</script>
