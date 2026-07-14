<!-- 测试用例页签组件 -->
<template>
  <div style="flex: 1; min-height: 0; display: flex; flex-direction: column;">
    <!-- 按钮行 + 用例选择器 -->
    <div style="display: flex; gap: 8px; margin-bottom: 12px; align-items: center; flex-wrap: wrap;">
      <n-button data-testid="testcase-multi-select-btn" @click="toggleMultiSelect" :disabled="orderDirty">{{ selectMode ? '退出多选' : '多选' }}</n-button>
      <n-button @click="handleLoadCase" :loading="loading">加载用例</n-button>
      <n-button type="primary" :disabled="!selectedCase" :loading="false" @click="handleExecuteCase">执行用例</n-button>
      <n-button @click="handleNewCase">新增模块</n-button>
      <n-button @click="handleDeleteCase" :disabled="!selectedCase">删除模块</n-button>

      <!-- 用例选择下拉框 -->
      <n-select
        v-model:value="selectedCase"
        :options="caseOptions"
        placeholder="选择用例"
        style="width: 300px;"
        size="small"
        @update:value="handleSelectCase"
      />
    </div>

    <!-- 文件信息 -->
    <div v-if="recordData" style="margin-bottom: 8px; display: flex; gap: 16px; font-size: 13px;">
      <span>版本: {{ recordData.version }}</span>
      <span>录制时间: {{ recordData.recorded_at }}</span>
      <span>消息数: {{ recordData.message_count }}</span>
    </div>

    <!-- 步骤顺序变更提示栏 -->
    <div
      v-if="orderDirty"
      data-testid="order-dirty-bar"
      style="display: flex; gap: 8px; align-items: center; margin-bottom: 8px; padding: 6px 10px; background: rgba(255,255,255,0.05); border-radius: 4px;"
    >
      <span style="font-size: 13px;">步骤顺序已变更（{{ stepCount }} 步）</span>
      <n-button size="tiny" data-testid="revert-order-btn" @click="handleRevertOrder">还原</n-button>
      <n-button size="tiny" type="primary" data-testid="save-order-btn" @click="handleSaveOrder">保存</n-button>
    </div>

    <!-- 多选操作栏 -->
    <div v-if="selectMode" style="display: flex; gap: 8px; align-items: center; margin-bottom: 8px; padding: 6px 10px; background: rgba(255,255,255,0.05); border-radius: 4px;">
      <span style="font-size: 13px;">已选择 {{ selectedRowIds.length }} 条消息</span>
      <n-button size="tiny" @click="handleCancelSelect">取消多选</n-button>
      <n-button size="tiny" type="error" @click="handleDeleteMessages">删除用例消息</n-button>
      <n-button size="tiny" type="primary" @click="handleBatchReplay">批量重放</n-button>
    </div>

    <!-- 消息表格 -->
    <message-table
      variant="testcase"
      :enable-reorder="true"
      :messages="messages"
      :selected-index="selectedIndex"
      :recorded-at="recordData?.recorded_at ?? ''"
      :select-mode="selectMode"
      :selected-row-ids="selectedRowIds"
      @select="selectMessage"
      @reorder="handleReorder"
      @update:selected-row-ids="val => selectedRowIds = val"
      @toggle-select-row="handleToggleSelectRow"
      @delete-row="handleDeleteRow"
      @focus-descript="focusDescriptInput"
    />

    <!-- 单条重发控制面板：选中行后显示 -->
    <replay-control
      v-if="selectedPairedEntry"
      :running="false"
      :progress="null"
      :selected-entry="selectedPairedEntry"
      :has-iterative-config="hasIterativeConfig"
      :show-pair-status="false"
      @retry="handleRetryMessage"
      @stop="handleStopReplay"
      @iterative-send="handleIterativeSend"
    />

    <!-- 底部：配对 Payload 编辑器（含描述输入与工具栏） -->
    <paired-payload-editor
      ref="pairedPayloadEditorRef"
      v-if="selectedPairedEntry"
      variant="testcase"
      v-model:descript="descriptDraft"
      :entry="selectedPairedEntry"
      :show-req-apply="true"
      :extra-apply-enabled="descriptDirty"
      @apply="handleApplyPayload"
      @config-change="updateIterativeConfig"
    />

    <!-- 新增用例对话框 -->
    <n-modal v-model:show="showNewCaseDialog" preset="card" title="新增测试模块" style="width: 400px;" :mask-closable="false">
      <n-input v-model:value="newCaseName" placeholder="输入测试模块名称" />
      <template #footer>
        <div style="display: flex; gap: 8px; justify-content: flex-end;">
          <n-button @click="showNewCaseDialog = false">取消</n-button>
          <n-button type="primary" @click="handleConfirmNewCase">创建</n-button>
        </div>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, nextTick } from 'vue'
import { useMessage, NSelect, NInput } from 'naive-ui'
import MessageTable from '../shared/message-table.vue'
import PairedPayloadEditor from '../shared/paired-payload-editor.vue'
import ReplayControl from '../shared/replay-control.vue'
import { createWailsTestCaseService } from '../shared/case-selector.requirement'
import { createWailsReplayControlService } from '../shared/replay-control.requirement'
import { createWailsRecordFileWriteService } from '../shared/paired-payload-editor.requirement'
import { useSelectedEntry } from '../shared/composables/use-selected-entry'
import type { RecordFileData } from '@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/models'
import { RecordEntryView } from '@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/models'
import { RecordFileData as RecordFileDataClass } from '@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/models'

const message = useMessage()

// ============ 事件定义 ============
const emit = defineEmits<{
  'replay-start': [source: 'testcase' | 'retry']
}>()

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
const filePath = ref('')
const recordData = ref<RecordFileData | null>(null)
const selectedIndex = ref<number | null>(null)

// 目标服务配置（默认值与父组件 targetService 保持一致）
const replayServerAddr = ref('10.254.114.204:18000')
const replayHttpAddr = ref('10.254.114.204:20144')
const replayOpenID = ref('test')

// 从父组件 targetService 同步配置
watch(() => props.targetService, (ts) => {
  console.log('[DEBUG testcase-tab] watch targetService 触发:', JSON.stringify(ts))
  if (ts) {
    replayServerAddr.value = ts.serverAddr
    replayHttpAddr.value = ts.httpAddr
    replayOpenID.value = ts.openID
    console.log('[DEBUG testcase-tab] 同步完成: openID=', replayOpenID.value, 'serverAddr=', replayServerAddr.value)
  }
}, { immediate: true, deep: true })

// 用例管理
const caseList = ref<any[]>([])
const selectedCase = ref('')
const loading = ref(false)
const showNewCaseDialog = ref(false)
const newCaseName = ref('')

// 多选模式
const selectMode = ref(false)
const selectedRowIds = ref<number[]>([])

// 步骤顺序变更追踪
const savedOrderFingerprint = ref('')

function orderFingerprint(msgs: RecordEntryView[]): string {
  return msgs.map(m => `${m.msg_id}:${m.msg_name}:${m.seq_id ?? ''}`).join('|')
}

function syncOrderBaseline() {
  savedOrderFingerprint.value = orderFingerprint(messages.value)
}

const orderDirty = computed(() => {
  if (!recordData.value || messages.value.length === 0) return false
  return orderFingerprint(messages.value) !== savedOrderFingerprint.value
})

const stepCount = computed(() => messages.value.length)

// 计算属性
const messages = computed(() => recordData.value?.messages ?? [])
const { pairedMessages, selectedPairedEntry } = useSelectedEntry(
  selectedIndex,
  messages,
  computed(() => recordData.value?.recorded_at)
)

const caseOptions = computed(() =>
  caseList.value.map((c: any) => ({ label: `${c.name} (${c.message_count}条)`, value: c.name }))
)

// ============ Service 实例 ============
const testCaseService = createWailsTestCaseService()
const replayControlService = createWailsReplayControlService()
const recordFileWriteService = createWailsRecordFileWriteService()

// ============ 事件监听 ============
// 注意：replay:progress 和 record:progress 事件已移至父组件 index.vue 统一处理，以支持重放结果记录

// ============ 暴露给父组件的方法 ============
function setRecordData(data: RecordFileData | null) {
  recordData.value = data
  if (data?.messages) {
    savedOrderFingerprint.value = orderFingerprint(data.messages as RecordEntryView[])
  } else {
    savedOrderFingerprint.value = ''
  }
}

function setTargetService(serverAddr: string, httpAddr: string, openID: string) {
  replayServerAddr.value = serverAddr
  replayHttpAddr.value = httpAddr
  replayOpenID.value = openID
}

function setCaseList(list: any[]) {
  caseList.value = list
}

function setSelectedCase(name: string) {
  selectedCase.value = name
}

// ============ 批量发送辅助 ============

/** 检查消息是否有有效的迭代/变量字段配置（input_type 非 original） */
function hasIterativeFieldValues(msg: RecordEntryView): boolean {
  const fv = msg.field_values
  if (!fv) return false
  return Object.values(fv).some((s: any) => s?.input_type && s.input_type !== 'original')
}

/** 统一发送消息（后端内部处理迭代/变量展开），一次调用避免 "已有重放任务正在运行" */
async function sendMessagesWithIterationSupport(
  reqs: RecordEntryView[],
  rangeStart: number,
  rangeEnd: number
): Promise<void> {
  // 筛选出 Req 消息，传递完整的 RecordEntryView（含 field_values）给后端
  const hasIter = reqs.some(m => hasIterativeFieldValues(m))
  if (hasIter) {
    await replayControlService.sendMessagesWithFieldValues(
      replayServerAddr.value, replayHttpAddr.value, replayOpenID.value,
      reqs, 1, rangeStart, rangeEnd
    )
  } else {
    await replayControlService.sendMessages(
      replayServerAddr.value, replayHttpAddr.value, replayOpenID.value,
      reqs, 1, rangeStart, rangeEnd
    )
  }
}

// ============ 用例管理操作 ============
async function handleLoadCase() {
  const list = await testCaseService.loadCaseList()
  caseList.value = list
  message.success(`已刷新用例列表，共 ${caseList.value.length} 个`)
}

// ============ 多选操作 ============
function toggleMultiSelect() {
  if (orderDirty.value && !selectMode.value) {
    message.warning('请先保存或还原步骤顺序')
    return
  }
  selectMode.value = !selectMode.value
  if (!selectMode.value) {
    selectedRowIds.value = []
  }
}

function handleCancelSelect() {
  selectMode.value = false
  selectedRowIds.value = []
}

async function handleBatchReplay() {
  if (selectedRowIds.value.length === 0) {
    message.warning('请先选择要重放的消息')
    return
  }
  // 筛选出选中的 Req 消息
  const selectedReqs = messages.value.filter((m, idx) => selectedRowIds.value.includes(idx) && m.direction === '→')
  if (!selectedReqs.length) {
    message.warning('所选消息中没有 Req 消息')
    return
  }
  const hasIter = selectedReqs.some(m => hasIterativeFieldValues(m))
  message.info(`正在重放 ${selectedReqs.length} 条消息${hasIter ? '（含迭代/变量配置）' : ''}...`)
  try {
    emit('replay-start', 'testcase')
    await sendMessagesWithIterationSupport(selectedReqs, 1, 1)
  } catch (e: any) {
    message.error('批量重放失败: ' + (e.message || e))
  }
}

function handleToggleSelectRow(rowId: number) {
  const idx = selectedRowIds.value.indexOf(rowId)
  if (idx >= 0) {
    selectedRowIds.value.splice(idx, 1)
  } else {
    selectedRowIds.value.push(rowId)
  }
}

async function handleDeleteRow(rowId: number) {
  if (!selectedCase.value) {
    message.warning('请先选择一个用例')
    return
  }
  try {
    // 从表格中移除指定行的消息
    const remainingMessages = messages.value.filter((m, idx) => idx !== rowId)
    if (!recordData.value) return

    // 更新 recordData
    recordData.value.messages = remainingMessages
    recordData.value.message_count = remainingMessages.length

    // 保存到文件
    await testCaseService.saveTestCase(selectedCase.value, recordData.value)

    syncOrderBaseline()
    message.success('已删除消息并保存到用例')
  } catch (e: any) {
    message.error('删除失败: ' + (e.message || e))
  }
}

async function handleDeleteMessages() {
  if (!selectedCase.value) {
    message.warning('请先选择一个用例')
    return
  }
  if (selectedRowIds.value.length === 0) {
    message.warning('请先选择要删除的消息')
    return
  }
  try {
    // 从表格中移除选中的消息
    const remainingMessages = messages.value.filter((m, idx) => !selectedRowIds.value.includes(idx))
    if (!recordData.value) return

    // 更新 recordData
    recordData.value.messages = remainingMessages
    recordData.value.message_count = remainingMessages.length

    // 保存到文件
    await testCaseService.saveTestCase(selectedCase.value, recordData.value)

    syncOrderBaseline()
    message.success(`已删除 ${selectedRowIds.value.length} 条消息并保存到用例`)
    toggleMultiSelect()
  } catch (e: any) {
    message.error('删除失败: ' + (e.message || e))
  }
}

async function handleDeleteCase() {
  if (!selectedCase.value) {
    message.warning('请先选择要删除的用例')
    return
  }
  try {
    await testCaseService.deleteTestCase(selectedCase.value)
    message.success('已删除用例')

    // 从下拉框中移除已删除的用例
    caseList.value = caseList.value.filter(caseItem => caseItem.name !== selectedCase.value)

    // 清空当前选中状态
    selectedCase.value = ''
  } catch (e: any) {
    message.error('删除失败: ' + (e.message || e))
  }
}

async function handleNewCase() {
  newCaseName.value = ''
  showNewCaseDialog.value = true
}

async function handleConfirmNewCase() {
  const name = newCaseName.value.trim()
  if (!name) {
    message.warning('请输入测试模块名称')
    return
  }
  try {
    const emptyData = new RecordFileDataClass({
      version: 1,
      recorded_at: new Date().toISOString(),
      server_addr: replayServerAddr.value,
      message_count: 0,
      messages: [],
    })
    await testCaseService.saveTestCase(name, emptyData)
    message.success(`已创建测试模块: ${name}`)
    showNewCaseDialog.value = false

    // 自动刷新下拉框并选中新创建的用例
    const list = await testCaseService.loadCaseList()
    caseList.value = list
    selectedCase.value = name
  } catch (e: any) {
    message.error('创建失败: ' + (e.message || e))
  }
}

async function handleSelectCase(value: string) {
  selectedCase.value = value
}

async function handleExecuteCase() {
  if (!selectedCase.value) {
    message.warning('请先选择一个测试模块')
    return
  }
  // 数据已通过 watch 监听器自动加载，这里只执行重放
  // 筛选所有 Req 发送
  const reqs = messages.value.filter(m => m.direction === '→')
  if (!reqs.length) {
    message.warning('用例中没有 Req 消息')
    return
  }
  const hasIter = reqs.some(m => hasIterativeFieldValues(m))
  message.info(`正在登录服务器执行重放${hasIter ? '（含迭代/变量配置）' : ''}...`)
  try {
    // 调试日志：打印实际使用的参数值
    console.log('[DEBUG testcase-tab handleExecuteCase]', JSON.stringify({
      replayOpenID: replayOpenID.value,
      replayServerAddr: replayServerAddr.value,
      replayHttpAddr: replayHttpAddr.value,
      rangeStart: props.rangeStart,
      rangeEnd: props.rangeEnd,
      targetService: props.targetService,
    }))
    emit('replay-start', 'testcase')
    await sendMessagesWithIterationSupport(
      reqs, props.rangeStart ?? 1, props.rangeEnd ?? 1
    )
  } catch (e: any) {
    message.error('执行失败: ' + (e.message || e))
  }
}

// ============ 表格交互 ============
function selectMessage(index: number) {
  selectedIndex.value = index
}

function handleReorder(reordered: any) {
  if (!recordData.value) return
  recordData.value.messages = reordered
  recordData.value.message_count = reordered.length
}

async function handleSaveOrder() {
  if (!selectedCase.value || !recordData.value || !orderDirty.value) return
  try {
    await testCaseService.saveTestCase(selectedCase.value, recordData.value)
    syncOrderBaseline()
    message.success('顺序已保存')
  } catch (e: any) {
    message.error('保存顺序失败: ' + (e.message || e))
  }
}

async function handleRevertOrder() {
  if (!selectedCase.value || !orderDirty.value) return
  try {
    const data = await testCaseService.loadTestCase(selectedCase.value)
    setRecordData(data)
    selectedIndex.value = null
    message.info('已还原步骤顺序')
  } catch (e: any) {
    message.error('还原失败: ' + (e.message || e))
  }
}

// ============ 描述编辑 ============
const descriptDraft = ref('')
const savedDescript = ref('')

watch(selectedPairedEntry, (entry) => {
  const d = entry?.req?.descript ?? ''
  descriptDraft.value = d
  savedDescript.value = d
}, { immediate: true })

const descriptDirty = computed(() => descriptDraft.value !== savedDescript.value)

function focusDescriptInput() {
  nextTick(() => {
    pairedPayloadEditorRef.value?.focusDescriptInput()
  })
}

// ============ Payload / 描述 / 字段配置 统一保存 ============
async function handleApplyPayload(index: number, payload: Record<string, any>) {
  if (!filePath.value) return

  const payloadDirty = pairedPayloadEditorRef.value?.hasReqPayloadChanges?.() ?? false
  const descriptChanged = descriptDirty.value
  // 收集卡片编辑器的字段4态配置（variable/range/enum/combo）
  const editorFieldStates = pairedPayloadEditorRef.value?.getFieldFourStates() ?? {}
  const hasConfigChanges = Object.values(editorFieldStates).some((s: any) => s?.input_type && s.input_type !== 'original')

  if (!payloadDirty && !descriptChanged && !hasConfigChanges) return

  try {
    let data: RecordFileData | null = recordData.value
    if (payloadDirty) {
      data = await recordFileWriteService.updateMessagePayload(filePath.value, index, payload)
    }
    if (descriptChanged) {
      data = await recordFileWriteService.updateMessageDescript(filePath.value, index, descriptDraft.value)
    }
    // 持久化字段配置（variable/range/enum/combo 的 input_type 及相关参数）
    if (hasConfigChanges) {
      data = await recordFileWriteService.updateMessageFieldValues(filePath.value, index, editorFieldStates)
    }
    if (data) {
      setRecordData(data)
      savedDescript.value = descriptDraft.value
    }
    message.success('已保存')
  } catch (e: any) {
    message.error('保存失败: ' + (e.message || e))
    throw e
  }
}

// ============ 重发控制 ============

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

  // 配置了字段迭代（枚举/范围/组合）时自动路由到迭代发送：
  // 用户预期是按配置展开为多条 Req，而非把迭代配置当作单条消息的字段值发送
  if (Object.keys(collectIterativeStates()).length > 0) {
    await handleIterativeSend()
    return
  }

  try {
    emit('replay-start', 'retry')
    await replayControlService.sendMessages(replayServerAddr.value, replayHttpAddr.value, replayOpenID.value, [targetMsg], count)
    message.info(`正在重发 ${targetMsg.msg_name} (${count} 次)`)
  } catch (e: any) {
    message.error('重发失败: ' + (e.message || e))
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

// 迭代发送：收集字段4态配置（编辑器优先，回退持久化 field_values）
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

// ============ 监听用例选择变化，自动加载数据 ============
watch(selectedCase, async (newCaseName) => {
  if (newCaseName) {
    try {
      const data = await testCaseService.loadTestCase(newCaseName)
      setRecordData(data)
      filePath.value = `cases/proto_cases/${newCaseName}.json`
    } catch (e: any) {
      message.error('加载用例失败: ' + (e.message || e))
    }
  }
})

// ============ 组件挂载时自动加载用例列表 ============
onMounted(async () => {
  try {
    const list = await testCaseService.loadCaseList()
    caseList.value = list
  } catch (e: any) {
    console.error('加载用例列表失败:', e)
  }
})

// ============ 暴露状态给父组件 ============
defineExpose({
  setRecordData,
  setTargetService,
  setCaseList,
  setSelectedCase,
  recordData,
  messages,
  selectedPairedEntry,
  replayServerAddr,
  replayHttpAddr,
  replayOpenID,
  selectMessage,
  selectedIndex,
})
</script>
