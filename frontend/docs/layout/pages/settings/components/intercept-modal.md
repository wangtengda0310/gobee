# InterceptModal - 劫持消息弹窗组件

> File path: `src/pages/settings/components/intercept-modal.vue`
> Parent: [../composables.md](../composables.md)

## Overview

显示被劫持的飞书消息列表，支持查看、复制、清空等操作。

采用**终端/控制台美学**设计，深色背景配合高对比度文字，自动适配浅色/深色主题。

## ASCII Layout Diagram

```
┌──────────────────────────────────────────────────────────────────────────────┐
│  被劫持的飞书消息                                           [清空] [关闭]    │
│  共 3 条消息                                                              │
│  ┌────────────────────────────────────────────────────────────────────────┐  │
│  │ ┌──────────────────────────────────────────────────────────────────────┐ │  │
│  │ │ [文本消息] excel-check     2026-03-31 14:23:15                      │ │  │
│  │ │ ─────────────────────────────────────────────────────────────────── │ │  │
│  │ │ **2026-03-31 14:23:13**                                            │ │  │
│  │ │                                                                  │ │  │
│  │ │ 📁 表格名称: HeroTest                                                │ │  │
│  │ │ 🔄 变更类型: 新增行                                                 │ │  │
│  │ │ 变更范围: 共 1 条记录被新增                                         │ │  │
│  │ │                                                                  │ │  │
│  │ │ [新增记录详情]                                                     │ │  │
│  │ │ 主键值: 123                                                        │ │  │
│  │ │ 行号: 第 5 行                                                       │ │  │
│  │ │                                                                  │ │  │
│  │ └──────────────────────────────────────────────────────────────────────┘ │  │
│  │                                                          [复制内容]      │  │
│  └────────────────────────────────────────────────────────────────────────┘  │
│                                                                              │
│  ┌────────────────────────────────────────────────────────────────────────┐  │
│  │ [卡片消息] excel-check     2026-03-31 14:22:10                      │  │
│  │ ─────────────────────────────────────────────────────────────────── │  │
│  │ **2026-03-31 14:22:08**                                            │ │  │
│  │                                                                  │ │  │
│  │ 📁 表格名称: CardTest                                                │ │  │
│  │ 🔄 变更类型: 修改行                                                 │ │  │
│  │ ...                                                              │  │
│  └────────────────────────────────────────────────────────────────────────┘  │
│                                                                              │
│                                   （可滚动，max-height: 65vh）                │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
```

## Component Structure

```
intercept-modal.vue                          # Modal component
├── Template
│   └── n-modal (preset="card") → intercept-modal.vue:53
│       ├── Template #header-extra
│       │   └── n-space → intercept-modal.vue:54
│       │       ├── n-tag → intercept-modal.vue:55
│       │       │   └── "共 {{ messages.length }} 条消息"
│       │       ├── n-button (清空) → intercept-modal.vue:56
│       │       └── n-button (关闭) → intercept-modal.vue:57
│       │
│       └── n-scrollbar (max-height: 65vh) → intercept-modal.vue:61
│           ├── n-empty (v-if="messages.length === 0")
│           │
│           └── .message-item (v-for="msg in messages") → intercept-modal.vue:64
│               ├── .message-header → intercept-modal.vue:65
│               │   ├── .message-meta → intercept-modal.vue:67
│               │   │   ├── n-tag (类型标签) → intercept-modal.vue:68
│               │   │   └── .message-guid (机器人GUID) → intercept-modal.vue:69
│               │   └── .message-time (时间戳) → intercept-modal.vue:70
│               │
│               ├── .message-content → intercept-modal.vue:73
│               │   └── pre.message-text → intercept-modal.vue:74
│               │       └── {{ msg.content }}
│               │
│               └── .message-actions → intercept-modal.vue:76
│                   └── n-button (复制) → intercept-modal.vue:77
│
└── Script
    └── onMounted → intercept-modal.vue:17
        ├── loadMessages()
        └── checkInterceptEnabled()
```

## Component File Mapping

| Component | File Path | Line | Description |
|-----------|-----------|------|-------------|
| Main Modal | components/intercept-modal.vue | 52-81 | n-modal + content |
| Message Item | components/intercept-modal.vue | 64-79 | Single message display |
| Message Header | components/intercept-modal.vue | 65-70 | Type, GUID, time |
| Message Content | components/intercept-modal.vue | 73-74 | Preformatted text |
| Message Actions | components/intercept-modal.vue | 76-77 | Copy button |

## State Management

| State | Source | Type | Description |
|-------|--------|------|-------------|
| `messages` | composables/use-intercept.ts | Ref\<InterceptedMessage[]\> | 被劫持的消息列表 |
| `showModal` | composables/use-intercept.ts | Ref\<boolean\> | 显示弹窗状态 |

## Props & Emits

**Props**: 无（使用 composable 共享状态）

**Emits**: 无

## Styling

### CSS Variables

```css
.intercept-modal {
    /* 深色主题（默认） */
    --msg-bg: #1e1e1e;           /* 消息卡片背景 */
    --msg-border: #333;            /* 边框颜色 */
    --msg-text: #d4d4d4;           /* 主文字颜色 */
    --msg-text-secondary: #858585; /* 次要文字颜色 */
    --msg-accent: #4ec9b0;         /* 强调色（悬停、高亮） */
    --msg-header-bg: #2d2d2d;      /* 内容区背景 */

    /* 浅色主题（@media prefers-color-scheme: light） */
    --msg-bg: #ffffff;
    --msg-border: #e0e0e0;
    --msg-text: #1a1a1a;
    --msg-text-secondary: #666;
    --msg-accent: #18a058;
    --msg-header-bg: #f5f5f7;
}
```

### Key Styles

| Class | Purpose | Value |
|-------|---------|-------|
| `.message-item` | 消息卡片容器 | border, padding, transition |
| `.message-header` | 消息头部 | flex, justify-between, border-bottom |
| `.message-text` | 消息内容（pre） | monospace, white-space: pre-wrap |
| `.message-actions` | 操作按钮区 | flex-end, gap |

## Animation & Micro-interactions

| Animation | Trigger | Effect |
|----------|---------|--------|
| `slideIn` | Component mount | 消息从上方滑入 (0.3s ease-out) |
| `hover` | `.message-item:hover` | 边框高亮、阴影、微向上位移 |
| `hover` | `.message-text:hover` | 边框颜色变化 |
| `hover` | `button:hover` | 按钮放大 (scale: 1.05) |

## Functions

| Function | Description |
|----------|-------------|
| `formatTime(timestamp)` | 格式化时间戳为本地字符串 |
| `copyContent(content)` | 复制内容到剪贴板 |
| `handleClearMessages()` | 清空所有消息（带确认） |
| `handleClose()` | 关闭弹窗 |

## Data Flow

### 加载消息流程

```
onMounted()
    │
    ├──► loadMessages()
    │         │
    │         ▼
    │    InterceptService.GetMessages()
    │         │
    │         ▼
    │    messages.value = result || []
    │
    └──► checkInterceptEnabled()
              │
              ▼
         interceptEnabled.value = enabled
```

### 复制内容流程

```
User clicks "复制" button
    │
    ▼
copyContent(content)
    │
    ├──► navigator.clipboard.writeText(content)
    │
    ▼
window.alert("已复制到剪贴板")
```

### 清空消息流程

```
User clicks "清空" button
    │
    ▼
window.confirm("确定要清空所有消息吗？")
    │
    ├──► 用户确认
    │         │
    │         ▼
    │    clearMessages()
    │         │
    │         ▼
    │    InterceptService.ClearMessages()
    │         │
    │         ▼
    │    messages.value = []
    │
    └──► 用户取消（无操作）
```

## Backend Services

### InterceptService

**Methods**:
- `GetMessages()` - 获取劫持消息列表
- `ClearMessages()` - 清空劫持消息

**Type Definition**:
```typescript
interface InterceptedMessage {
    id: string         // 唯一 ID
    robotGuid: string  // 机器人 GUID
    msgType: string    // "text" | "interactive"
    content: string    // 消息内容
    timestamp: string  // 发送时间（格式化字符串）
}
```

## Design Notes

### Typography

- **代码字体**: SF Mono, Monaco, Cascadia Code, Courier New
- **字体大小**: 13px (消息内容), 11-12px (元数据)
- **行高**: 1.6 (消息内容，确保可读性)

### Color Strategy

- **深色主题**: 模拟终端/控制台，深色背景 + 浅色文字
- **浅色主题**: 自动适配 `prefers-color-scheme: light`
- **对比度**: 所有文字与背景对比度符合 WCAG 标准

### Accessibility

- **键盘导航**: 所有按钮支持 Tab 键导航
- **颜色对比**: 适配浅色/深色主题
- **语义化 HTML**: 使用正确的标签和结构

## Related Files

### Composables
- `../composables/use-intercept.ts` - 状态管理和事件监听

### Main Page
- `../index.vue` - 设置页面主组件

## Notes

- 弹窗宽度: 750px
- 最大高度: 85vh (视口高度的 85%)
- 滚动区域: 65vh
- 消息进入动画: 0.3s ease-out
- **事件监听器在模块加载时设置，不是组件挂载时**
- 支持复制全部消息内容到剪贴板
- 清空操作需要用户确认
