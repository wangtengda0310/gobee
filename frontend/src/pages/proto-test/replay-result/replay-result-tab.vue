<!-- 重放结果页签组件 -->
<template>
  <div style="flex: 1; min-height: 0; display: flex; flex-direction: column;">
    <!-- 结果选择器 -->
    <replay-result-selector
      :results="replayResults"
      :current-id="currentResultId"
      @select="handleSelectResult"
      @clear="handleClearResults"
      @delete="handleDeleteResult"
    />

    <!-- 结果文件信息栏（版本、录制时间、消息数、来源合并为一行） -->
    <div v-if="recordData" style="margin-bottom: 8px; display: flex; gap: 16px; font-size: 13px;">
      <span>版本: {{ recordData.version }}</span>
      <span>录制时间: {{ recordData.recorded_at }}</span>
      <span>消息数: {{ recordData.message_count }}</span>
      <span v-if="currentResult">来源: {{ sourceLabel }}</span>
    </div>

    <!-- 停止重放按钮（仅当有运行中的重放时显示） -->
    <div v-if="hasRunningResults" style="display: flex; gap: 8px; align-items: center; margin-bottom: 8px; padding: 6px 10px; background: rgba(255,100,100,0.08); border-radius: 4px;">
      <n-button type="error" size="small" @click="handleStopAllReplay">停止所有重放</n-button>
      <span style="font-size: 13px; color: var(--n-text-color-2);">
        正在运行 {{ runningCount }} 个重放任务
      </span>
    </div>

    <!-- 消息表格 -->
    <message-table
      variant="packet"
      :messages="messages"
      :selected-index="selectedIndex"
      :recorded-at="recordData?.recorded_at ?? ''"
      :select-mode="false"
      @select="selectMessage"
      @reorder="handleReorder"
    />

    <!-- 单条重发控制面板：选中行后显示 -->
    <replay-control
      v-if="selectedPairedEntry"
      :file-path="''"
      :running="false"
      :progress="null"
      :selected-entry="selectedPairedEntry"
      @retry="handleRetryMessage"
      @stop="() => {}"
    />

    <!-- 底部：配对 Payload 编辑器 -->
    <paired-payload-editor
      ref="pairedPayloadEditorRef"
      v-if="selectedPairedEntry"
      :entry="selectedPairedEntry"
      :show-req-apply="false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useMessage, NButton } from 'naive-ui'
import MessageTable from '../shared/message-table.vue'
import PairedPayloadEditor from '../shared/paired-payload-editor.vue'
import ReplayControl from '../shared/replay-control.vue'
import ReplayResultSelector from './replay-result-selector.vue'
import { buildPairedEntries, type PairedEntry } from '../shared/composables/use-paired-messages'
import { createWailsReplayControlService } from '../shared/replay-control.requirement'
import type { RecordFileData } from '@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/models'

const message = useMessage()

// ============ Props ============
const props = defineProps<{
  targetService: {
    serverAddr: string
    httpAddr: string
    openID: string
  }
}>()

// ============ Service 实例 ============
const replayControlService = createWailsReplayControlService()

// ============ 状态管理 ============
const replayResults = ref<ReplayResult[]>([])
const currentResultId = ref<string | null>(null)
const selectedIndex = ref<number | null>(null)

// paired-payload-editor ref（用于获取编辑器当前 payload）
const pairedPayloadEditorRef = ref<InstanceType<typeof PairedPayloadEditor> | null>(null)

// 计算属性
const currentResult = computed(() => {
  if (!currentResultId.value) return null
  return replayResults.value.find(r => r.id === currentResultId.value) ?? null
})

const hasRunningResults = computed(() =>
  replayResults.value.some(r => r.status === 'running')
)

const runningCount = computed(() =>
  replayResults.value.filter(r => r.status === 'running').length
)

const recordData = computed(() => currentResult.value?.recordData ?? null)
const messages = computed(() => recordData.value?.messages ?? [])
const pairedMessages = computed(() => buildPairedEntries(messages.value, recordData.value?.recorded_at))

const selectedPairedEntry = computed<PairedEntry | null>(() => {
  if (selectedIndex.value === null) return null
  for (const entry of pairedMessages.value) {
    if (entry.type === 'pair') {
      if (entry.req?.index === selectedIndex.value || entry.ack?.index === selectedIndex.value) {
        return entry
      }
    } else {
      if (entry.ntf?.index === selectedIndex.value || entry.ack?.index === selectedIndex.value) {
        return entry
      }
    }
  }
  return null
})

const sourceLabel = computed(() => {
  if (!currentResult.value) return ''
  const sourceMap: Record<string, string> = {
    packet: '发包改包 - 开始重放',
    testcase: '测试用例 - 执行用例',
    retry: '重发控制 - 重发',
  }
  return sourceMap[currentResult.value.source] || currentResult.value.source
})

// ============ 数据结构 ============
export interface ReplayResult {
  id: string
  source: 'packet' | 'testcase' | 'retry'
  timestamp: string
  recordData: RecordFileData
  status: 'running' | 'completed' | 'error' | 'cancelled'
  error?: string
}

// ============ 暴露给父组件的方法 ============
function addReplayResult(result: ReplayResult) {
  replayResults.value.push(result)
  currentResultId.value = result.id
  selectedIndex.value = null
}

function setCurrentResult(id: string) {
  currentResultId.value = id
}

// ============ 停止所有重放 ============
async function handleStopAllReplay() {
  try {
    await replayControlService.stopReplay()
    message.info('已停止所有重放')
    // 将所有运行中的结果标记为 cancelled
    replayResults.value.forEach(r => {
      if (r.status === 'running') {
        r.status = 'cancelled'
      }
    })
  } catch (e: any) {
    message.error('停止失败: ' + (e.message || e))
  }
}

// ============ 结果管理操作 ============
function handleSelectResult(id: string | null) {
  console.log('[DEBUG handleSelectResult] 切换到结果=', id, '之前的 currentResultId=', currentResultId.value)
  currentResultId.value = id
}

function handleClearResults() {
  console.log('[DEBUG handleClearResults] 清空所有结果, 之前 replayResults.length=', replayResults.value.length)
  replayResults.value = []
  currentResultId.value = null
  selectedIndex.value = null
}

function handleDeleteResult(id: string) {
  const index = replayResults.value.findIndex(r => r.id === id)
  if (index !== -1) {
    replayResults.value.splice(index, 1)
    // 如果删除的是当前结果，清空当前选中
    if (currentResultId.value === id) {
      currentResultId.value = null
      selectedIndex.value = null
    }
  }
}

// ============ 表格交互 ============
function selectMessage(index: number) {
  selectedIndex.value = index
}

function handleReorder(reordered: any) {
  if (!recordData.value) return
  recordData.value.messages = reordered
}

// ============ 重发控制 ============
async function handleRetryMessage(count: number) {
  if (!selectedPairedEntry.value) return
  const entry = selectedPairedEntry.value
  let targetMsg = entry.req
  if (!targetMsg && entry.type === 'single' && entry.direction === 'S->C') {
    message.warning('只能重发 Req 消息')
    return
  }
  if (!targetMsg) {
    message.warning('未找到 Req 消息')
    return
  }
  // 从编辑器获取当前 payload（用户可能在 JSON 编辑器中有未应用的修改）
  const editorPayload = pairedPayloadEditorRef.value?.getCurrentReqPayload()
  if (editorPayload) {
    targetMsg = { ...targetMsg, payload: editorPayload }
  }
  try {
    await replayControlService.sendMessages(
      props.targetService.serverAddr,
      props.targetService.httpAddr,
      props.targetService.openID,
      [targetMsg],
      count
    )
    message.info(`正在重发 ${targetMsg.msg_name} (${count} 次)`)
  } catch (e: any) {
    message.error('重发失败: ' + (e.message || e))
  }
}


// ============ 暴露状态给父组件 ============
defineExpose({
  addReplayResult,
  setCurrentResult,
  replayResults, // 暴露 ref 对象本身
  currentResultId,
  currentResult,
  recordData,
  messages,
  selectedPairedEntry,
  selectMessage,
  selectedIndex,
})
</script>
