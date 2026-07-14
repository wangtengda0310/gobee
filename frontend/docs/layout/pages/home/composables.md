# Home - Composables (Logic Layer)

> Parent: [index.md](./index.md)

## Overview

Composables 包含 AI 聊天页面的响应式逻辑和状态管理。

## Composables Structure

```
composables/
├── use-chat-state.ts            # 聊天状态管理
└── chat.types.ts                # 类型定义
```

## State Management

### Global State

详见 [index.md#Key State](./index.md#key-state)

### Composable Details

#### use-chat-state.ts

聊天状态管理。导出 `messages`, `inputText`, `isStreaming`, `streamingContent`, `showConfig` 及 `addMessage`, `clearMessages`, `resetStreaming` 方法。

#### chat.types.ts

类型定义。`ChatMessage` 接口定义见源码 [chat.types.ts](../../composables/chat.types.ts)。

## Data Flow Diagrams

### 发送消息流程

```
User inputs text + Enter
    │
    ▼
handleSend() → ChatService.SendMessageStream()
    │
    ├──► Event: 'chatStreamChunk' → streamingContent += chunk
    ├──► Event: 'chatStreamDone' → addMessage() + resetStreaming()
    └──► Event: 'chatStreamError' → message.error() + resetStreaming()
```

## Related Files

### Components
- `components/chat-config-panel.vue` - AI 配置面板
- `components/chat-message-item.vue` - 单条消息组件

### Main Page
- [index.md](./index.md) - 主页面布局

## Notes

- 消息自动滚动到底部
- 支持流式生成（实时显示生成内容）
- 生成过程中显示"停止"按钮
- 无消息时显示欢迎界面
- 配置面板可设置 Provider 和 API Key
- 消息历史持久化到后端
