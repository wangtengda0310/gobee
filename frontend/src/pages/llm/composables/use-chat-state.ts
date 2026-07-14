/**
 * 聊天状态管理 composable
 */
import { reactive, ref } from 'vue'
import type { ChatMessage, ChatConfig } from './chat.types'
import { defaultChatConfig } from './chat.types'

// 消息列表
export const messages = reactive<ChatMessage[]>([])

// 流式状态
export const isStreaming = ref(false)
export const streamingContent = ref('')

// 输入
export const inputText = ref('')

// 配置面板
export const showConfig = ref(false)
export const currentConfig = ref<ChatConfig>({ ...defaultChatConfig })

// 加载状态
export const isLoading = ref(false)

/**
 * 添加消息
 */
export function addMessage(message: ChatMessage) {
  messages.push(message)
}

/**
 * 清空消息
 */
export function clearMessages() {
  messages.length = 0
  streamingContent.value = ''
}

/**
 * 重置流式状态
 */
export function resetStreaming() {
  isStreaming.value = false
  streamingContent.value = ''
}
