<!-- 协议重放页面 - 发包改包 / 测试用例 / 重放结果 -->
<template>
  <div
    ref="pageRef"
    tabindex="-1"
    style="padding: 12px; height: 100%; box-sizing: border-box; display: flex; flex-direction: column; color: #fff; outline: none;"
    @keydown.esc="handleCloseEditor"
    @click="handlePageClick"
  >
    <!-- 目标服务配置（页签标题上方，全局共享） -->
    <target-service-config
      :server-addr="targetService.serverAddr"
      :http-addr="targetService.httpAddr"
      :openID="targetService.openID"
      :range-start="targetService.rangeStart"
      :range-end="targetService.rangeEnd"
      :tcp-listen-port="targetService.tcpListenPort"
      :http-listen-port="targetService.httpListenPort"
      @update:server-addr="val => targetService.serverAddr = val"
      @update:http-addr="val => targetService.httpAddr = val"
      @update:openID="val => targetService.openID = val"
      @update:range-start="val => targetService.rangeStart = val"
      @update:range-end="val => targetService.rangeEnd = val"
      @update:tcpListenPort="val => targetService.tcpListenPort = val"
      @update:httpListenPort="val => targetService.httpListenPort = val"
    />

    <!-- 页签切换 -->
    <div style="display: flex; gap: 2px; flex-shrink: 0; border-bottom: 1px solid #333; margin-bottom: 8px;">
      <div
        v-for="tab in tabOptions"
        :key="tab.key"
        :style="{
          padding: '6px 16px',
          cursor: 'pointer',
          fontSize: '14px',
          color: activeTab === tab.key ? '#fff' : '#888',
          borderBottom: activeTab === tab.key ? '2px solid var(--n-primary-color)' : '2px solid transparent',
          transition: 'all 0.2s',
          userSelect: 'none',
        }"
        @click="handleTabChange(tab.key)"
      >{{ tab.label }}</div>
    </div>

    <!-- 发包改包页签 -->
    <packet-tab
      v-show="activeTab === 'packet'"
      ref="packetTabRef"
      :record-data="sharedRecordData"
      :target-service="targetService"
      :range-start="targetService.rangeStart"
      :range-end="targetService.rangeEnd"
      @replay-start="onReplayStart"
    />

    <!-- 测试用例页签 -->
    <testcase-tab
      v-show="activeTab === 'testcase'"
      ref="testcaseTabRef"
      :record-data="sharedRecordData"
      :target-service="targetService"
      :case-list="caseList"
      :selected-case="selectedCase"
      :range-start="targetService.rangeStart"
      :range-end="targetService.rangeEnd"
      @update:selected-case="val => selectedCase = val"
      @replay-start="onReplayStart"
    />

    <!-- 重放结果页签 -->
    <replay-result-tab
      v-show="activeTab === 'replay-result'"
      ref="replayResultTabRef"
      :target-service="targetService"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onUnmounted } from 'vue'
import { useMessage } from 'naive-ui'
import PacketTab from './stream-proxy/packet-tab.vue'
import TestcaseTab from './cases/testcase-tab.vue'
import ReplayResultTab from './replay-result/replay-result-tab.vue'
import TargetServiceConfig from './shared/target-service-config.vue'
import { Events } from '@wailsio/runtime'
import { createWailsTestCaseService } from './shared/case-selector.requirement'
import type { RecordFileData } from '@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/models'
import { RecordEntryView } from '@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/models'

const message = useMessage()
const pageRef = ref<HTMLDivElement | null>(null)

// 组件引用
const packetTabRef = ref<InstanceType<typeof PacketTab> | null>(null)
const testcaseTabRef = ref<InstanceType<typeof TestcaseTab> | null>(null)
const replayResultTabRef = ref<InstanceType<typeof ReplayResultTab> | null>(null)

// 目标服务配置（发包改包和测试用例页签共享）
const targetService = ref({
  serverAddr: '10.254.114.204:18000',
  httpAddr: '10.254.114.204:20144',
  openID: 'test',
  rangeStart: 1,
  rangeEnd: 1,
  tcpListenPort: 18000,
  httpListenPort: 20144,
})

// 共享录制数据（发包改包和测试用例页签共享）
const sharedRecordData = ref<any>(null)

// 测试用例管理
const testCaseService = createWailsTestCaseService()
const caseList = ref<any[]>([])
const selectedCase = ref('')

// 页签管理
const activeTab = ref('packet')
const tabOptions = [
  { key: 'packet', label: '发包改包' },
  { key: 'testcase', label: '测试用例' },
  { key: 'replay-result', label: '重放结果' },
]

// 重放结果管理
let currentReplayResultId: string | null = null
let currentReplayMessages: any[] = []
let currentReplaySource: 'packet' | 'testcase' | 'retry' = 'retry'

// 重放开始处理（由子组件 emit 触发）
function onReplayStart(source: 'packet' | 'testcase' | 'retry') {
  console.log('[DEBUG onReplayStart] source=', source, '之前的 currentReplayResultId=', currentReplayResultId)
  currentReplaySource = source

  // 初始化新的重放结果
  currentReplayMessages = []
  const newResult = {
    id: `replay_${Date.now()}`,
    source,
    timestamp: new Date().toISOString(),
    recordData: {
      version: 1,
      recorded_at: new Date().toISOString(),
      server_addr: '',
      message_count: 0,
      messages: [],
    },
    status: 'running' as const,
  }
  currentReplayResultId = newResult.id

  // 添加到重放结果页签
  if (replayResultTabRef.value) {
    replayResultTabRef.value.addReplayResult(newResult)
  }

  // 切换到重放结果页签
  activeTab.value = 'replay-result'
}

// 页签切换处理
function handleTabChange(tabKey: string) {
  activeTab.value = tabKey

  // 切换到测试用例页签时，如果用例列表为空则自动加载
  if (tabKey === 'testcase' && caseList.value.length === 0 && testcaseTabRef.value) {
    loadTestCaseList()
  }
}

let unsubscribeReplayResult: (() => void) | null = null
let unsubscribeReplayProgress: (() => void) | null = null

// 监听重放结果消息事件（后端专用通道）
unsubscribeReplayResult = Events.On('replay:result', (raw: any) => {
  const data = raw.data ?? raw
  console.log('[DEBUG replay:result] 收到事件, currentReplayResultId=', currentReplayResultId, 'msg_name=', data.msg_name, 'direction=', data.direction)
  if (!currentReplayResultId) {
    console.warn('[DEBUG replay:result] currentReplayResultId 为空，跳过消息')
    return
  }

  // 后端直接发送 RecordEntryView 合约类型，使用 createFrom 统一反序列化
  const newEntry = RecordEntryView.createFrom({
    ...data,
    index: currentReplayMessages.length,
  })
  currentReplayMessages.push(newEntry)
  console.log('[DEBUG replay:result] 追加消息, currentReplayMessages.length=', currentReplayMessages.length)

  // 更新重放结果页签数据
  // 注意：defineExpose 暴露的 ref 在模板 ref 中会被自动 unwrap，
  // 所以 replayResultTabRef.value.replayResults 直接就是数组，不需要再 .value
  const resultsArray = replayResultTabRef.value?.replayResults as any[] | undefined
  if (resultsArray) {
    const currentResult = resultsArray.find((r: any) => r.id === currentReplayResultId)
    if (currentResult) {
      currentResult.recordData.messages = [...currentReplayMessages]
      currentResult.recordData.message_count = currentReplayMessages.length
      console.log('[DEBUG replay:result] 更新结果成功, result.id=', currentResult.id, 'messages.length=', currentResult.recordData.messages.length)
    } else {
      console.warn('[DEBUG replay:result] 未找到 currentResult, id=', currentReplayResultId, 'resultsArray.ids=', resultsArray.map((r: any) => r.id))
    }
  } else {
    console.warn('[DEBUG replay:result] resultsArray 为空')
  }
})

// 监听重放进度事件（仅状态管理）
unsubscribeReplayProgress = Events.On('replay:progress', (raw: any) => {
  const data = raw.data ?? raw
  console.log('[DEBUG replay:progress] status=', data.status, 'currentReplayResultId=', currentReplayResultId, 'sent=', data.sent, 'total=', data.total)

  // 更新重放结果状态
  if (currentReplayResultId && replayResultTabRef.value) {
    const resultsArray = replayResultTabRef.value?.replayResults as any[] | undefined
    if (resultsArray) {
      const currentResult = resultsArray.find((r: any) => r.id === currentReplayResultId)
      if (currentResult) {
        currentResult.status = data.status
        if (data.status === 'error') {
          currentResult.error = data.error_message
        }
        console.log('[DEBUG replay:progress] 更新结果状态, result.id=', currentResult.id, 'status=', data.status)
      }
    }
  }

  // 重放完成时清空引用
  if (data.status === 'completed' || data.status === 'error' || data.status === 'cancelled') {
    console.log('[DEBUG replay:progress] 重放结束, 清空 currentReplayResultId, 当前 replayResults 数量=', (replayResultTabRef.value?.replayResults as any[])?.length)
    if (replayResultTabRef.value) {
      const resultsArray = replayResultTabRef.value?.replayResults as any[] | undefined
      if (resultsArray) {
        for (const r of resultsArray) {
          console.log('[DEBUG replay:progress] 结果=', r.id, 'source=', r.source, 'status=', r.status, 'messages.length=', r.recordData?.messages?.length)
        }
      }
    }
    currentReplayResultId = null
  }
})

onUnmounted(() => {
  if (unsubscribeReplayResult) {
    unsubscribeReplayResult()
    unsubscribeReplayResult = null
  }
  if (unsubscribeReplayProgress) {
    unsubscribeReplayProgress()
    unsubscribeReplayProgress = null
  }
})

// 加载测试用例列表
async function loadTestCaseList() {
  try {
    const list = await testCaseService.loadCaseList()
    caseList.value = list
    message.success(`已加载 ${list.length} 个测试用例`)
  } catch (e: any) {
    message.error('加载用例列表失败: ' + (e.message || e))
  }
}

// 关闭编辑器
function handleCloseEditor() {
  // 清空各页签的选中状态
  if (packetTabRef.value) packetTabRef.value.selectMessage(-1)
  if (testcaseTabRef.value) testcaseTabRef.value.selectMessage(-1)
  if (replayResultTabRef.value) replayResultTabRef.value.selectMessage(-1)
}

// 页面点击处理
function handlePageClick(e: MouseEvent) {
  if (e.target !== pageRef.value) return
  handleCloseEditor()
}
</script>
