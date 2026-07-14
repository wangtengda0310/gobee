# Home - AI 助手页面

> File path: `src/pages/home/index.vue`
> Route: `/Home`

## Overview

AI 聊天助手页面，提供与 AI 对话的功能。采用简单的上下布局：消息列表区域 + 输入区域。

## ASCII Layout Diagram

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                        消息列表区域（可滚动）                                 │
│                        最大宽度 900px，居中                                  │
│                                                                              │
│                    ┌──────────────────────────────┐                          │
│                    │                              │                          │
│                    │        🤖                    │                          │
│                    │    AI 助手                   │                          │
│                    │  有什么可以帮助你的？         │                          │
│                    │                              │                          │
│                    └──────────────────────────────┘                          │
│                                                                              │
│                    ┌──────────────────────────────┐                          │
│                    │ 用户: xxx                    │                          │
│                    │ ────────────────────────────  │                          │
│                    │                              │                          │
│                    └──────────────────────────────┘                          │
│                                                                              │
│                    ┌──────────────────────────────┐                          │
│                    │ 助手: xxx                    │                          │
│                    │ ────────────────────────────  │                          │
│                    │                              │                          │
│                    └──────────────────────────────┘                          │
│                                                                              │
│                                                                              │
├──────────────────────────────────────────────────────────────────────────────┤
│ Input Area (固定高度)                                                        │
│ ┌────┬──────────────────────────────────────────────┬──────┬──────┬──────┐  │
│ │ ⚙️ │ │ [输入消息...] (Enter发送, Shift+Enter换行) │ │ 发送 │ 清空 │      │  │
│ └────┴──────────────────────────────────────────────┴──────┴──────┴──────┘  │
└──────────────────────────────────────────────────────────────────────────────┘
```

## Component Tree Structure

```
pages/home/index.vue
├── ChatConfigPanel
├── .chat-container
│   └── n-scrollbar
│       └── .messages-wrapper
│           ├── .welcome-message
│           ├── ChatMessageItem (v-for) — 用户/助手消息
│           └── ChatMessageItem (流式) — isStreaming=true
└── .input-area
    ├── n-button (配置)
    ├── n-input (textarea)
    ├── n-button (发送/停止)
    └── n-button (清空)
```

## Layout Areas

| Area | Size | Description |
|------|------|-------------|
| Chat Container | `flex: 1` | 消息列表区域（可滚动） |
| Input Area | Auto | 输入区域（固定高度） |
| Messages Wrapper | max-width 900px | 消息居中显示 |

## Component File Mapping

| Component | File Path | Line | Description |
|-----------|-----------|------|-------------|
| Main Page | pages/home/index.vue | 189-276 | Main container |
| Config Panel | components/chat-config-panel.vue | 191 | 配置弹窗 |
| Messages Container | pages/home/index.vue | 194-220 | Scrollable message list |
| Welcome Message | pages/home/index.vue | 198-204 | Welcome screen |
| Message Items | components/chat-message-item.vue | 207, 214 | Individual messages |
| Input Area | pages/home/index.vue | 224-275 | Input controls |

## Data Flow

```
User Input
    │
    ▼
handleSend()
    │
    ├──► Add user message to messages
    ├──► Set isStreaming = true
    ├──► Call ChatService.SendMessageStream()
    │         │
    │         ▼
    │    Backend (Wails)
    │         │
    │         ▼
    │    Event: 'chatStreamChunk'
    │         │
    │         ▼
    └────► streamingContent += chunk
              │
              ▼
         Event: 'chatStreamDone'
              │
              ├──► Add assistant message to messages
              ├──► Reset streaming state
              └──► isStreaming = false
```

## Related Files

### Components (components/)
- `components/chat-config-panel.vue` - AI 配置面板（API Key、Provider）
- `components/chat-message-item.vue` - 单条消息组件

### Composables (composables/)
- `composables/use-chat-state.ts` - 聊天状态管理
- `composables/chat.types.ts` - 类型定义

## Event Flow

| 组件 | 事件 | 处理 |
|------|------|------|
| n-input | Enter | handleSend() |
| n-button (发送) | @click | handleSend() |
| n-button (停止) | @click | handleStop() |
| n-button (清空) | @click | handleClear() |
| n-button (配置) | @click | showConfig = true |
| Wails Event | chatStreamChunk | streamingContent += |
| Wails Event | chatStreamDone | addMessage() |
| Wails Event | chatStreamError | message.error() |

## Backend Services

`ChatService`: GetConfig, SaveConfig, SendMessageStream, StopStream, GetHistory, ClearHistory

支持模型: Claude 3.5 Sonnet/Haiku, Claude 3 Opus, GPT-4o, GPT-4o Mini, GPT-4 Turbo, 自定义 Provider

## Wails Events

| Event Name | Data | Description |
|------------|------|-------------|
| `chatStreamChunk` | string | 流式生成增量内容 |
| `chatStreamDone` | null | 流式生成完成 |
| `chatStreamError` | string | 流式生成错误 |

## Notes

- 消息列表最大宽度 900px，居中显示
- 支持 Enter 发送，Shift+Enter 换行
- 流式生成时显示"停止"按钮
- 首次使用时检查配置，无 API Key 时显示配置面板
- 欢迎消息在无消息时显示
- 消息自动滚动到底部
