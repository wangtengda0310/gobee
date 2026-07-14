<script setup lang="ts">
/**
 * 聊天消息组件
 * 用于显示单条聊天消息，支持用户消息和 AI 消息的区分显示
 */
import { computed } from 'vue'
import type { ChatMessage } from '../composables/chat.types'

const props = defineProps<{
  message: ChatMessage
  isStreaming?: boolean
}>()

// 格式化时间
const formattedTime = computed(() => {
  const date = new Date(props.message.timestamp * 1000)
  return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
})

// 是否为用户消息
const isUser = computed(() => props.message.role === 'user')
</script>

<template>
  <div
    class="message-item"
    :class="[isUser ? 'user' : 'assistant']"
  >
    <div class="message-avatar">
      <span v-if="isUser">👤</span>
      <span v-else>🤖</span>
    </div>
    <div class="message-content">
      <div class="message-header">
        <span class="message-role">{{ isUser ? '你' : 'AI' }}</span>
        <span class="message-time">{{ formattedTime }}</span>
      </div>
      <div class="message-text">
        <!-- 显示内容（流式时显示累积内容） -->
        <span v-if="isStreaming && !isUser">
          {{ message.content }}<span class="cursor">▌</span>
        </span>
        <span v-else>{{ message.content }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.message-item {
  display: flex;
  gap: 12px;
  padding: 16px;
  margin: 8px 0;
  border-radius: 12px;
  max-width: 85%;
}

.message-item.user {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  margin-left: auto;
  flex-direction: row-reverse;
}

.message-item.assistant {
  background: #3a3a3a;
  margin-right: auto;
}

.message-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.15);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  flex-shrink: 0;
}

.message-content {
  flex: 1;
  min-width: 0;
}

.message-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}

.message-role {
  font-weight: 600;
  font-size: 13px;
  color: rgba(255, 255, 255, 0.9);
}

.message-time {
  font-size: 11px;
  color: rgba(255, 255, 255, 0.5);
}

.message-text {
  font-size: 14px;
  line-height: 1.6;
  color: rgba(255, 255, 255, 0.95);
  word-break: break-word;
  white-space: pre-wrap;
}

/* 流式输出光标动画 */
.cursor {
  animation: blink 1s infinite;
}

@keyframes blink {
  0%, 50% { opacity: 1; }
  51%, 100% { opacity: 0; }
}
</style>
