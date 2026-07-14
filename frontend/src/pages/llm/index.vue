<script setup lang="ts">
/**
 * AI 聊天主页面
 * Home 页面入口，提供 AI 对话功能
 */
import { ref, onMounted, onUnmounted, nextTick, computed } from 'vue'
import { NInput, NButton, NScrollbar, NSpin, useMessage } from 'naive-ui'
import { Events } from '@wailsio/runtime'
import ChatMessageItem from './components/chat-message-item.vue'
import ChatConfigPanel from './components/chat-config-panel.vue'
import {
  messages,
  isStreaming,
  streamingContent,
  inputText,
  showConfig,
  addMessage,
  clearMessages,
  resetStreaming
} from './composables/use-chat-state'
import { ChatService } from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/settings/home"
import type { ChatMessage } from './composables/chat.types'

const message = useMessage()

// 消息容器引用
const messagesContainer = ref<HTMLElement | null>(null)
const scrollbarRef = ref<any>(null)

// 流式消息（用于显示正在生成的消息）
const streamingMessage = computed<ChatMessage | null>(() => {
  if (!isStreaming.value || !streamingContent.value) return null
  return {
    role: 'assistant',
    content: streamingContent.value,
    timestamp: Math.floor(Date.now() / 1000)
  }
})

// 事件取消函数
let unsubscribeChunk: (() => void) | null = null
let unsubscribeDone: (() => void) | null = null
let unsubscribeError: (() => void) | null = null

// 发送消息
async function handleSend() {
  const content = inputText.value.trim()
  if (!content || isStreaming.value) return

  // 添加用户消息
  const userMsg: ChatMessage = {
    role: 'user',
    content: content,
    timestamp: Math.floor(Date.now() / 1000)
  }
  addMessage(userMsg)
  inputText.value = ''
  isStreaming.value = true
  streamingContent.value = ''

  // 滚动到底部
  await nextTick()
  scrollToBottom()

  try {
    await ChatService.SendMessageStream(content)
  } catch (e: any) {
    message.error('发送失败: ' + e.message)
    isStreaming.value = false
  }
}

// 停止生成
function handleStop() {
  ChatService.StopStream()
  resetStreaming()
}

// 清空历史
async function handleClear() {
  try {
    await ChatService.ClearHistory()
    clearMessages()
    message.success('已清空对话')
  } catch (e: any) {
    message.error('清空失败: ' + e.message)
  }
}

// 滚动到底部
function scrollToBottom() {
  if (scrollbarRef.value) {
    scrollbarRef.value.scrollTo({ top: 999999, behavior: 'smooth' })
  }
}

// 键盘事件处理
function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    handleSend()
  }
}

// 加载历史记录
async function loadHistory() {
  try {
    const session = await ChatService.GetHistory()
    if (session && session.messages) {
      messages.length = 0
      session.messages.forEach(msg => messages.push({
        role: msg.role as 'user' | 'assistant',
        content: msg.content,
        timestamp: msg.timestamp
      }))
    }
  } catch (e) {
    console.log('无历史记录')
  }
}

// 检查配置
async function checkConfig() {
  try {
    const config = await ChatService.GetConfig()
    if (config) {
      const hasKey = config.provider === 'anthropic'
        ? config.anthropicConfig?.apiKey
        : config.openaiConfig?.apiKey
      if (!hasKey) {
        showConfig.value = true
        message.info('请先配置 API Key')
      }
    }
  } catch (e) {
    showConfig.value = true
  }
}

onMounted(async () => {
  // 加载历史
  await loadHistory()

  // 检查配置
  await checkConfig()

  // 先清除可能残留的同名事件监听器，避免重复注册导致内容双倍
  Events.Off('chatStreamChunk', 'chatStreamDone', 'chatStreamError')
  // 订阅流式事件
  // 后端为每个 content chunk 附带递增序列号 seq，防止 WebView2 bridge 事件乱序导致中文错乱。
  // pendingChunks 暂存乱序到达的 chunk，flushPendingChunks 按 seq 顺序拼接。
  const pendingChunks = new Map<number, string>()
  let expectedSeq = 1

  // 按序列号顺序将待处理 chunk 追加到 streamingContent
  const flushPendingChunks = () => {
    while (pendingChunks.has(expectedSeq)) {
      streamingContent.value += pendingChunks.get(expectedSeq)!
      pendingChunks.delete(expectedSeq)
      expectedSeq++
    }
    nextTick(() => scrollToBottom())
  }

  unsubscribeChunk = Events.On('chatStreamChunk', (data: any) => {
    // Wails Event.Emit 传递 map[string]any 到 JS 后，data 的结构为 { data: [{ content, seq }] }
    let content = ''
    let seq = 0

    if (typeof data === 'string') {
      // 旧格式兼容（裸字符串，无序列号）
      content = data
      seq = expectedSeq
    } else if (data?.data) {
      const payload = Array.isArray(data.data) ? data.data[0] : data.data
      if (typeof payload === 'object' && payload !== null) {
        content = payload.content || ''
        seq = payload.seq || 0
      } else {
        content = String(payload)
        seq = expectedSeq
      }
    }

    if (seq > 0 && content !== '') {
      pendingChunks.set(seq, content)
      flushPendingChunks()
    }
  })

  unsubscribeDone = Events.On('chatStreamDone', () => {
    // 添加助手消息
    if (streamingContent.value) {
      const assistantMsg: ChatMessage = {
        role: 'assistant',
        content: streamingContent.value,
        timestamp: Math.floor(Date.now() / 1000)
      }
      addMessage(assistantMsg)
    }
    resetStreaming()
  })

  unsubscribeError = Events.On('chatStreamError', (data: any) => {
    const errorMsg = data.data?.[0] || '未知错误'
    message.error('请求失败: ' + errorMsg)
    resetStreaming()
  })
})

onUnmounted(() => {
  // 取消事件订阅
  unsubscribeChunk?.()
  unsubscribeDone?.()
  unsubscribeError?.()
})
</script>

<template>
  <div id="ChatHome">
    <!-- 配置面板 -->
    <ChatConfigPanel v-model:show="showConfig" />

    <!-- 消息列表区域 -->
    <div class="chat-container">
      <n-scrollbar ref="scrollbarRef" class="messages-scrollbar">
        <div class="messages-wrapper">
          <!-- 欢迎消息 -->
          <div v-if="messages.length === 0 && !isStreaming" class="welcome-message">
            <div class="welcome-icon">🤖</div>
            <div class="welcome-text">
              <h2>AI 助手</h2>
              <p>有什么可以帮助你的？</p>
            </div>
          </div>

          <!-- 消息列表 -->
          <ChatMessageItem
            v-for="(msg, index) in messages"
            :key="index"
            :message="msg"
          />

          <!-- 流式消息 -->
          <ChatMessageItem
            v-if="streamingMessage"
            :message="streamingMessage"
            :is-streaming="true"
          />
        </div>
      </n-scrollbar>
    </div>

    <!-- 输入区域 -->
    <div class="input-area">
      <div class="input-wrapper">
        <!-- 配置按钮 -->
        <n-button
          quaternary
          circle
          @click="showConfig = true"
          title="配置"
        >
          <template #icon>
            <span>⚙️</span>
          </template>
        </n-button>

        <!-- 输入框 -->
        <n-input
          v-model:value="inputText"
          type="textarea"
          :autosize="{ minRows: 1, maxRows: 5 }"
          placeholder="输入消息... (Enter发送, Shift+Enter换行)"
          @keydown="handleKeydown"
          :disabled="isStreaming"
        />

        <!-- 发送/停止按钮 -->
        <n-button
          v-if="!isStreaming"
          type="primary"
          @click="handleSend"
          :disabled="!inputText.trim()"
        >
          发送
        </n-button>
        <n-button
          v-else
          type="error"
          @click="handleStop"
        >
          停止
        </n-button>

        <!-- 清空按钮 -->
        <n-button
          quaternary
          @click="handleClear"
          :disabled="messages.length === 0"
          title="清空对话"
        >
          清空
        </n-button>
      </div>
    </div>
  </div>
</template>

<style scoped>
#ChatHome {
  position: relative;
  width: 100%;
  height: 100%;
  background-color: #1a1a1a;
  display: flex;
  flex-direction: column;
  color: white;
  box-sizing: border-box;
}

.chat-container {
  flex: 1;
  min-height: 0;
  position: relative;
}

.messages-scrollbar {
  height: 100%;
}

.messages-wrapper {
  max-width: 900px;
  margin: 0 auto;
  padding: 20px;
}

.welcome-message {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 300px;
  text-align: center;
}

.welcome-icon {
  font-size: 64px;
  margin-bottom: 20px;
}

.welcome-text h2 {
  margin: 0 0 10px 0;
  font-size: 24px;
  color: #fff;
}

.welcome-text p {
  margin: 0;
  font-size: 14px;
  color: #888;
}

.input-area {
  border-top: 1px solid #333;
  background: #222;
  padding: 12px 16px;
}

.input-wrapper {
  max-width: 900px;
  margin: 0 auto;
  display: flex;
  gap: 8px;
  align-items: flex-end;
}

.input-wrapper :deep(.n-input) {
  flex: 1;
}

.input-wrapper :deep(.n-input .n-input__textarea-el) {
  font-size: 14px;
}
</style>
