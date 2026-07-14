<!-- 发包改包页签组件 -->
<template>
  <div style="flex: 1; min-height: 0; display: flex; flex-direction: column;">
    <!-- 顶部按钮行 -->
    <div style="display: flex; gap: 8px; margin-bottom: 12px; align-items: center;">
      <n-button @click="toggleMultiSelect">{{ selectMode ? '退出多选' : '多选' }}</n-button>
      <n-button @click="toggleFilterMode">{{ filterMode ? '关闭实时修改' : '开启实时修改' }}</n-button>
      <n-button type="primary" @click="handleStartRecord" :loading="recordRunning" :disabled="recordRunning">开始录制</n-button>
      <n-button @click="handleStopRecord" :disabled="!recordRunning">停止录制</n-button>
      <n-button
        v-if="filterMode && recordRunning"
        type="warning"
        @click="handleReleaseAll"
        :disabled="interceptedSeqIDs.size === 0"
      >放行全部</n-button>
      <n-button type="primary" @click="handleStartReplay" :loading="false" :disabled="filterMode">开始重放</n-button>
      <!-- 录制进度 -->
      <div v-if="recordProgress" style="display: flex; gap: 4px; align-items: center; margin-left: 8px;">
        录制进度<n-tag :type="recordStatusType" size="small">{{ recordStatusLabel }}</n-tag>
      </div>
    </div>

    <!-- 文件信息 -->
    <div v-if="recordData" style="margin-bottom: 8px; display: flex; gap: 16px; font-size: 13px;">
      <span>版本: {{ recordData.version }}</span>
      <span>录制时间: {{ recordData.recorded_at }}</span>
      <span>消息数: {{ recordData.message_count }}</span>
    </div>

    <!-- 多选操作栏 -->
    <div v-if="selectMode" style="display: flex; gap: 8px; align-items: center; margin-bottom: 8px; padding: 6px 10px; background: rgba(255,255,255,0.05); border-radius: 4px;">
      <span style="font-size: 13px;">已选择 {{ selectedRowIds.length }} 条消息</span>
      <n-button size="tiny" @click="handleCancelSelect">取消多选</n-button>
      <n-button
        v-if="selectedRowIds.length > 0"
        size="tiny"
        type="primary"
        :disabled="!selectedRowsHasReq"
        @click="handleSaveToCase"
      >保存到用例</n-button>
    </div>

    <!-- 消息表格 -->
    <message-table
      ref="messageTableRef"
      variant="packet"
      :messages="messages"
      :selected-index="selectedIndex"
      :recorded-at="recordData?.recorded_at ?? ''"
      :select-mode="selectMode"
      :selected-row-ids="selectedRowIds"
      :enable-add-to-case="true"
      :intercepted-seq-ids="interceptedSeqIDs"
      @select="selectMessage"
      @reorder="handleReorder"
      @toggle-select-row="handleToggleSelectRow"
      @add-to-case="handleAddToCase"
    />
    <!-- 单条重发控制面板：选中行后显示 -->
    <replay-control
      v-if="selectedPairedEntry"
      :running="false"
      :progress="null"
      :selected-entry="selectedPairedEntry"
      :has-iterative-config="hasIterativeConfig"
      :intercept-mode="filterMode"
      @retry="handleRetryMessage"
      @stop="handleStopReplay"
      @iterative-send="handleIterativeSend"
    />

    <!-- 底部：配对 Payload 编辑器 -->
    <paired-payload-editor
      ref="pairedPayloadEditorRef"
      v-if="selectedPairedEntry"
      :key="editorRowKey"
      :entry="selectedPairedEntry"
      :req-payload-override="selectedReqPayloadOverride"
      :show-req-apply="!filterMode"
      @apply="handleApplyPayload"
      @config-change="updateIterativeConfig"
    />

    <!-- 保存用例对话框 -->
    <n-modal v-model:show="showCaseDialog" preset="card" title="保存到用例" style="width: 400px;" :mask-closable="false">
      <n-select
        v-model:value="selectedCaseName"
        :options="caseList"
        placeholder="选择用例或输入新用例名称"
        filterable
        tag
        clearable
      />
      <template #footer>
        <div style="display: flex; gap: 8px; justify-content: flex-end;">
          <n-button @click="showCaseDialog = false">取消</n-button>
          <n-button type="primary" @click="confirmSaveCase">追加</n-button>
        </div>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useMessage, NTag, NSelect } from 'naive-ui'
import MessageTable from '../shared/message-table.vue'
import PairedPayloadEditor from '../shared/paired-payload-editor.vue'
import ReplayControl from '../shared/replay-control.vue'
import { createWailsReplayControlService } from '../shared/replay-control.requirement'
import { createWailsRecordControlService, createWailsProtoTestConfigService } from '../shared/protocol-content.requirement'
import { createWailsTestCaseService } from '../shared/case-selector.requirement'
import { useSelectedEntry } from '../shared/composables/use-selected-entry'
import type { RecordFileData } from '@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/models'
import { RecordEntryView } from '@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/models'

const message = useMessage()

// ============ 事件定义 ============
const emit = defineEmits<{
  'replay-start': [source: 'packet' | 'retry']
}>()

// ============ 状态管理 ============
// ============ Props ============
const props = defineProps<{
  rangeStart?: number
  rangeEnd?: number
  targetService?: {
    serverAddr: string
    httpAddr: string
    openID: string
    rangeStart: number
    rangeEnd: number
  }
}>()

// ============ 状态管理 ============
const recordData = ref<RecordFileData | null>(null)
const selectedIndex = ref<number | null>(null)
const selectedRowIds = ref<number[]>([])

// 目标服务配置（默认值与父组件 targetService 保持一致）
const replayServerAddr = ref('10.254.114.204:18000')
const replayHttpAddr = ref('10.254.114.204:20144')
const replayOpenID = ref('test')

// 从父组件 targetService 同步配置
watch(() => props.targetService, (ts) => {
  if (ts) {
    replayServerAddr.value = ts.serverAddr
    replayHttpAddr.value = ts.httpAddr
    replayOpenID.value = ts.openID
  }
}, { immediate: true, deep: true })

// 录制状态
const recordRunning = ref(false)
const recordProgress = ref<any>(null)

// 重放状态（已移除：replayRunning/replayProgress/repeatCount 由父组件管理）

// 多选模式
const selectMode = ref(false)
// 实时编辑模式 开启此模式后 录制过程中客户端发往服务端的数据包会强制前端页面修改后才能转发给服务端
const filterMode = ref(false)
// 拦截队列：记录被拦截但未放行的消息 SeqID
const interceptedSeqIDs = ref<Set<number>>(new Set())
// 代理连接 ID（单连接假设，取最新 record:intercepted 的 conn_id）
const connId = ref<number>(0)
// 内存中的编辑 payload（key=消息 index，放行时按 seq_id 提交）
const editedPayloads = ref<Map<number, Record<string, any>>>(new Map())

// 合并拦截 toast
let interceptToastPending = 0
let interceptToastTimer: ReturnType<typeof setTimeout> | null = null

// 保存用例对话框
const showCaseDialog = ref(false)
const selectedCaseName = ref<string | null>(null)

const pendingCaseData = ref<RecordFileData | null>(null) // 暂存要追加的数据
// 用例列表（用于保存到用例）
const caseList = ref<any[]>([])

// 计算属性
const messages = computed(() => recordData.value?.messages ?? [])
const { pairedMessages, selectedPairedEntry } = useSelectedEntry(
  selectedIndex,
  messages,
  computed(() => recordData.value?.recorded_at)
)

// 当前多选行中是否至少包含一条 Req（用于控制"保存到用例"按钮可用性）
const selectedRowsHasReq = computed(() => {
  return pairedMessages.value
    .filter(p => selectedRowIds.value.includes(p.id))
    .some(p => p.req !== null)
})

const selectedReqPayloadOverride = computed(() => {
  const reqIndex = selectedPairedEntry.value?.req?.index
  if (reqIndex === undefined) return null
  return editedPayloads.value.get(reqIndex) ?? null
})

const editorRowKey = computed(() => {
  const idx = selectedPairedEntry.value?.req?.index ?? selectedIndex.value
  return idx ?? -1
})

const recordStatusType = computed(() => {
  if (!recordProgress.value) return 'info'
  const s = recordProgress.value.status
  if (s === 'completed') return 'success'
  if (s === 'error') return 'error'
  if (s === 'cancelled') return 'warning'
  return 'info'
})

const recordStatusLabel = computed(() => {
  if (!recordProgress.value) return ''
  const p = recordProgress.value
  const map: Record<string, string> = {
    listening: `监听中`,
    running: `录制中 (${p.message_count ?? 0})`,
    completed: `录制完成 (${p.message_count ?? 0})`,
    error: `录制错误: ${p.error_message}`,
    cancelled: '录制已取消',
  }
  return map[p.status] || p.status
})

// ============ Service 实例 ============
const recordControlService = createWailsRecordControlService()
const replayControlService = createWailsReplayControlService()
const testCaseService = createWailsTestCaseService()

/** 同 seq_id+msg_id+direction 的 C→S 消息只保留一条（拦截时 progress 与 intercepted 曾重复推送） */
function isDuplicateMessage(entry: RecordEntryView, messages: RecordEntryView[]): boolean {
  return messages.some(m =>
    m.seq_id === entry.seq_id &&
    m.msg_id === entry.msg_id &&
    m.direction === entry.direction
  )
}

function appendRecordedMessage(raw: Record<string, unknown>): RecordEntryView | null {
  if (!recordData.value) return null
  const newEntry = RecordEntryView.createFrom({
    ...raw,
    index: recordData.value.messages.length,
  })
  if (isDuplicateMessage(newEntry, recordData.value.messages)) {
    return null
  }
  recordData.value.messages = [...recordData.value.messages, newEntry]
  recordData.value.message_count = recordData.value.messages.length
  return newEntry
}

// ============ 事件监听 ============
// 注意：replay:progress 事件已移至父组件 index.vue 统一处理，以支持重放结果记录

import { Events } from '@wailsio/runtime'

let unsubscribeRecordProgress: (() => void) | null = null
let unsubscribeRecordIntercepted: (() => void) | null = null

onMounted(() => {
  // 监听录制进度事件
  unsubscribeRecordProgress = Events.On('record:progress', (raw: any) => {
    const data = raw.data ?? raw
    try {
      // 仅处理录制异常终止：正常停止录制后状态会回到 listening，running 由 handleStopRecord 设置
      if (data.status === 'error') {
        recordRunning.value = false
      }
      recordProgress.value = data

      if (data.latest_msg) {
        appendRecordedMessage(data.latest_msg)
      }
    } catch (e) {
      console.error('[record:progress] 处理事件失败:', e)
    }
  })

  // 监听拦截事件（拦截模式下，Req 消息被拦截后推送到前端）
  unsubscribeRecordIntercepted = Events.On('record:intercepted', (raw: any) => {
    const data = raw.data ?? raw
    try {
      if (data.conn_id !== undefined) {
        connId.value = Number(data.conn_id)
      }
      if (data.latest_msg) {
        const newEntry = appendRecordedMessage(data.latest_msg)
        if (newEntry?.seq_id !== undefined) {
          interceptedSeqIDs.value = new Set([...interceptedSeqIDs.value, newEntry.seq_id])
          scheduleInterceptToast(1)
        }
      }
    } catch (e) {
      console.error('[record:intercepted] 处理事件失败:', e)
    }
  })
})

onUnmounted(() => {
  if (unsubscribeRecordProgress) {
    unsubscribeRecordProgress()
    unsubscribeRecordProgress = null
  }
  if (unsubscribeRecordIntercepted) {
    unsubscribeRecordIntercepted()
    unsubscribeRecordIntercepted = null
  }
})

function scheduleInterceptToast(count: number) {
  interceptToastPending += count
  if (interceptToastTimer) clearTimeout(interceptToastTimer)
  interceptToastTimer = setTimeout(() => {
    if (interceptToastPending > 0) {
      message.info(`新增 ${interceptToastPending} 条待放行`)
    }
    interceptToastPending = 0
    interceptToastTimer = null
  }, 500)
}

// ============ 暴露给父组件的方法 ============
function setRecordData(data: RecordFileData | null) {
  recordData.value = data
}

function setTargetService(serverAddr: string, httpAddr: string, openID: string) {
  replayServerAddr.value = serverAddr
  replayHttpAddr.value = httpAddr
  replayOpenID.value = openID
}

// ============ 录制操作 ============
async function handleStartRecord() {
  // 检查监听是否已启动
  try {
    const status = await recordControlService.getRecordStatus()
    if (status.status !== 'listening' && status.status !== 'running') {
      message.warning('监听未启动，请在设置抽屉中配置并启动监听')
      return
    }
  } catch (e: any) {
    message.error('获取录制状态失败: ' + (e.message || e))
    return
  }

  // 录制数据驻留内存，不再自动生成落盘文件名
  // 持久化由用户点"保存为用例"时手动触发
  recordData.value = {
    version: 1,
    recorded_at: new Date().toISOString(),
    server_addr: replayServerAddr.value,
    message_count: 0,
    messages: [],
  }
  interceptedSeqIDs.value = new Set()
  editedPayloads.value = new Map()
  connId.value = 0

  await recordControlService.startRecord(filterMode.value)
  recordRunning.value = true
}

async function handleStopRecord() {
  if (filterMode.value && interceptedSeqIDs.value.size > 0) {
    await releasePendingWithEdits()
  }
  if (filterMode.value && recordRunning.value) {
    await recordControlService.setFilterMode(false)
  }
  await recordControlService.stopRecord()
  recordRunning.value = false
}

// ============ 重放操作 ============
async function handleStartReplay() {
  if (!messages.value.length) {
    message.warning('无协议数据可重放')
    return
  }
  // 筛选表格中的 Req（direction = "→"），全部发送一次
  const reqs = messages.value.filter(m => m.direction === '→')
  if (!reqs.length) {
    message.warning('表格中没有 Req 消息可重放')
    return
  }
  try {
    emit('replay-start', 'packet')
    await replayControlService.sendMessages(replayServerAddr.value, replayHttpAddr.value, replayOpenID.value, reqs, 1,
      props.rangeStart ?? 1, props.rangeEnd ?? 1)
    message.info(`正在重放 ${reqs.length} 条 Req...`)
  } catch (e: any) {
    message.error('重放失败: ' + (e.message || e))
  }
}

async function handleStopReplay() {
  try {
    await replayControlService.stopReplay()
    message.info('正在停止重放...')
  } catch (e: any) {
    message.error('停止失败: ' + (e.message || e))
  }
}

// ============ 多选操作 ============
function toggleMultiSelect() {
  selectMode.value = !selectMode.value
  if (!selectMode.value) {
    selectedRowIds.value = []
  }
}

function handleCancelSelect() {
  selectMode.value = false
  selectedRowIds.value = []
}

function handleToggleSelectRow(rowId: number) {
  const idx = selectedRowIds.value.indexOf(rowId)
  if (idx >= 0) {
    selectedRowIds.value = [...selectedRowIds.value.slice(0, idx), ...selectedRowIds.value.slice(idx + 1)]
  } else {
    selectedRowIds.value = [...selectedRowIds.value, rowId]
  }
}

// 加载用例列表并显示对话框
async function loadCaseListAndShowDialog(dataToAppend: RecordFileData) {
  try {
    const list = await testCaseService.loadCaseList()
    caseList.value = list.map((item) => ({ label: item.name, value: item.name }))
    pendingCaseData.value = dataToAppend
    selectedCaseName.value = null
    showCaseDialog.value = true
  } catch (e: any) {
    message.error('加载用例列表失败: ' + (e.message || e))
  }
}

// 接收 message-table 传递的配对行原始消息（修复索引错位：不再用 rowId 过滤 messages[idx]）
async function handleAddToCase(messagesToAdd: any[]) {
  if (!recordData.value) {
    message.warning('无录制数据')
    return
  }
  if (messagesToAdd.length === 0) {
    message.warning('未找到选中行的消息')
    return
  }

  // 直接使用 message-table 提取的原始消息（req + ack / ntf）
  const singleRecordData: RecordFileData = {
    version: recordData.value.version,
    recorded_at: recordData.value.recorded_at,
    server_addr: recordData.value.server_addr,
    message_count: messagesToAdd.length,
    messages: messagesToAdd,
  }

  await loadCaseListAndShowDialog(singleRecordData)
}


async function handleSaveToCase() {
  if (!recordData.value || selectedRowIds.value.length === 0) return
  // 从配对行中提取实际选中的原始消息（修复索引错位：不再用 selectedRowIds 过滤 messages[idx]）
  const selectedMessages = pairedMessages.value
    .filter(p => selectedRowIds.value.includes(p.id))
    .flatMap(p => {
      const msgs: any[] = []
      if (p.req) msgs.push(p.req)
      // 只保存 Req，不保存 Ack/Ntf
      return msgs
    })

  if (selectedMessages.length === 0) {
    // 匹配到了行但没有 Req，说明选中了 Ntf/Ack 单行
    message.warning('选中的行不包含 Req 消息，无法保存到用例（用例仅保存客户端请求）')
    return
  }

  const multiRecordData: RecordFileData = {
    version: recordData.value.version,
    recorded_at: recordData.value.recorded_at,
    server_addr: recordData.value.server_addr,
    message_count: selectedMessages.length,
    messages: selectedMessages,
  }
  await loadCaseListAndShowDialog(multiRecordData)
}

async function confirmSaveCase() {
  const caseName = selectedCaseName.value
  if (!caseName) {
    message.warning('请选择用例')
    return
  }
  if (!pendingCaseData.value) {
    message.warning('无数据可保存')
    return
  }

  try {
    // 追加数据到现有用例（不会覆盖 testcase-tab 中已维护的用例数据）
    await testCaseService.appendTestCase(caseName, pendingCaseData.value!)
    message.success(`已追加到用例: ${caseName}`)

    // 保存完整数据标志（在清空前检查）
    const isFullData = recordData.value && pendingCaseData.value?.message_count === recordData.value.message_count

    showCaseDialog.value = false
    selectedCaseName.value = null
    pendingCaseData.value = null

    if (isFullData) {
      toggleMultiSelect()
    }
  } catch (e: any) {
    message.error('保存失败: ' + (e.message || e))
  }
}

// ============ 切换实时编辑模式 ============
async function toggleFilterMode() {
  if (filterMode.value && recordRunning.value && interceptedSeqIDs.value.size > 0) {
    await releasePendingWithEdits()
  }
  if (filterMode.value && recordRunning.value) {
    await recordControlService.setFilterMode(false)
  }
  filterMode.value = !filterMode.value
  if (!filterMode.value) {
    interceptedSeqIDs.value = new Set()
    editedPayloads.value = new Map()
  }
}

function saveCurrentEditToMap() {
  const payload = pairedPayloadEditorRef.value?.getCurrentReqPayload()
  const req = selectedPairedEntry.value?.req
  if (payload && req?.index !== undefined && req.seq_id !== undefined && interceptedSeqIDs.value.has(req.seq_id)) {
    editedPayloads.value = new Map(editedPayloads.value).set(req.index, payload)
  }
}

function collectEditsJSON(): string {
  saveCurrentEditToMap()
  if (!recordData.value) return '{}'
  const edits: Record<string, Record<string, any>> = {}
  for (const seqId of interceptedSeqIDs.value) {
    const msg = recordData.value.messages.find(
      m => m.seq_id === seqId && m.direction === '→'
    )
    if (!msg || msg.index === undefined) continue
    const payload = editedPayloads.value.get(msg.index)
    if (payload) {
      edits[String(seqId)] = payload
    }
  }
  return JSON.stringify(edits)
}

async function releasePendingWithEdits() {
  const editsJSON = collectEditsJSON()
  if (connId.value > 0) {
    await recordControlService.releasePendingMessages(connId.value, editsJSON)
  } else {
    await recordControlService.releaseAllPending(editsJSON)
  }
  interceptedSeqIDs.value = new Set()
  editedPayloads.value = new Map()
}

async function handleReleaseAll() {
  if (interceptedSeqIDs.value.size === 0) return
  try {
    await releasePendingWithEdits()
    message.success('已放行全部待发送消息')
  } catch (e: any) {
    message.error('放行失败: ' + (e.message || e))
  }
}

// ============ 表格交互 ============
function selectMessage(index: number) {
  if (selectedIndex.value !== index) {
    saveCurrentEditToMap()
  }
  selectedIndex.value = index
}
function handleReorder(reordered: any) {
  if (!recordData.value) return
  recordData.value.messages = reordered
}

// ============ Payload 编辑（纯内存操作，不落盘） ============
// packet-tab 编辑的是录制中的临时数据，持久化发生在用户点"保存为用例"时
async function handleApplyPayload(index: number, payload: Record<string, any>) {
  try {
    const payloadDirty = pairedPayloadEditorRef.value?.hasReqPayloadChanges?.() ?? false
    const editorFieldStates = pairedPayloadEditorRef.value?.getFieldFourStates() ?? {}
    const hasConfigChanges = Object.values(editorFieldStates).some((s: any) => s?.input_type && s.input_type !== 'original')

    if (!payloadDirty && !hasConfigChanges) {
      message.info('没有需要保存的修改')
      return
    }

    // 直接修改内存中的录制数据
    if (recordData.value && recordData.value.messages[index]) {
      if (payloadDirty) {
        recordData.value.messages[index].payload = payload
      }
      if (hasConfigChanges) {
        recordData.value.messages[index].field_values = editorFieldStates
      }
    }
    message.success('修改已保存')
  } catch (e: any) {
    message.error('保存失败: ' + (e.message || e))
    throw e
  }
}

// ============ 重发控制 ============

// paired-payload-editor ref（用于收集迭代配置）
// message-table ref（用于自动滚动到拦截消息）
const messageTableRef = ref<any>(null)

// paired-payload-editor ref（用于收集迭代配置）
const pairedPayloadEditorRef = ref<InstanceType<typeof PairedPayloadEditor> | null>(null)

// 是否有迭代配置（事件驱动更新，非 computed）
const hasIterativeConfig = ref(false)

// 收集迭代配置：卡片编辑器活跃状态优先；编辑器未挂载（JSON 模式）时
// 回退到用例中持久化的 field_values。无有效迭代配置时返回空对象。
function collectIterativeStates(): Record<string, any> {
  const editorStates = pairedPayloadEditorRef.value?.getFieldFourStates() ?? {}
  if (Object.keys(editorStates).length > 0) {
    return Object.values(editorStates).some((s: any) => s?.input_type !== 'original') ? editorStates : {}
  }
  const persisted = selectedPairedEntry.value?.req?.field_values
  if (persisted && Object.values(persisted).some((s: any) => s?.input_type && s.input_type !== 'original')) {
    return persisted as Record<string, any>
  }
  return {}
}

// 由 paired-payload-editor 的 config-change 事件触发
function updateIterativeConfig() {
  hasIterativeConfig.value = Object.keys(collectIterativeStates()).length > 0
}

async function handleRetryMessage(count: number) {
  if (!selectedPairedEntry.value || !replayServerAddr.value) return
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
  // 从编辑器获取当前 payload（用户可能在 JSON/卡片编辑器中有未应用的修改）
  const editorPayload = pairedPayloadEditorRef.value?.getCurrentReqPayload()
  if (editorPayload) {
    targetMsg = { ...targetMsg, payload: editorPayload }
  }

  // 判断是否为拦截消息（需要放行）
  const isIntercepted = targetMsg.seq_id !== undefined && interceptedSeqIDs.value.has(targetMsg.seq_id)

  // 配置了字段迭代（枚举/范围/组合）时自动路由到迭代发送：
  // 用户预期是按配置展开为多条 Req，而非把迭代配置当作单条消息的字段值发送。
  // 拦截放行必须精确发送当前这一条，不参与迭代。
  if (!isIntercepted && Object.keys(collectIterativeStates()).length > 0) {
    await handleIterativeSend()
    return
  }

  try {
    emit('replay-start', 'retry')
    await replayControlService.sendMessages(replayServerAddr.value, replayHttpAddr.value, replayOpenID.value, [targetMsg], count)

    if (isIntercepted && targetMsg.seq_id !== undefined) {
      // 拦截消息放行成功，从拦截队列移除（使用不可变更新确保 Vue 响应式）
      const newSet = new Set(interceptedSeqIDs.value)
      newSet.delete(targetMsg.seq_id)
      interceptedSeqIDs.value = newSet
      message.success(`已放行 ${targetMsg.msg_name}`)
    } else {
      message.info(`正在重发 ${targetMsg.msg_name} (${count} 次)`)
    }
  } catch (e: any) {
    message.error('重发失败: ' + (e.message || e))
  }
}

// 迭代发送：收集字段4态配置（编辑器优先，回退持久化 field_values），调用后端生成并批量发送迭代消息
async function handleIterativeSend() {
  if (!selectedPairedEntry.value?.req || !replayServerAddr.value) return
  const req = selectedPairedEntry.value.req
  const fieldStates = collectIterativeStates()
  if (Object.keys(fieldStates).length === 0) {
    message.warning('没有字段迭代配置')
    return
  }
  // 从编辑器获取当前 payload（用户可能在 JSON/卡片编辑器中有未应用的修改）
  const editorPayload = pairedPayloadEditorRef.value?.getCurrentReqPayload()
  const payload = editorPayload ?? req.payload
  try {
    emit('replay-start', 'retry')
    await replayControlService.sendIterativeMessages(
      replayServerAddr.value, replayHttpAddr.value, replayOpenID.value,
      req.msg_id, req.msg_name, JSON.stringify(payload), req.direction,
      fieldStates
    )
    message.info(`正在迭代发送 ${req.msg_name}...`)
  } catch (e: any) {
    message.error('迭代发送失败: ' + (e.message || e))
  }
}

// ============ 监听录制文件变化，自动填充服务器地址 ============
watch(
  () => recordData.value,
  (val) => {
    if (val?.server_addr) {
      replayServerAddr.value = val.server_addr
      const colonIdx = val.server_addr.lastIndexOf(':')
      if (colonIdx > 0) {
        replayHttpAddr.value = val.server_addr.substring(0, colonIdx) + ':20144'
      }
    }
  }
)

// ============ 暴露状态给父组件 ============
defineExpose({
  setRecordData,
  setTargetService,
  recordData,
  messages,
  selectedPairedEntry,
  replayServerAddr,
  replayHttpAddr,
  replayOpenID,
  selectMessage,
})
</script>
