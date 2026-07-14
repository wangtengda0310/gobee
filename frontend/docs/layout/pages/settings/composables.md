# Settings - Composables (Logic Layer)

> Parent: [index.md](./index.md)

## Overview

Composables 包含设置页面的响应式逻辑和状态管理。

## Composables Structure

```
composables/
├── use-settings.ts              # 设置状态管理
├── use-intercept.ts             # 消息劫持状态管理
└── use-server-logs.ts         # 服务端日志状态管理
```

## State Management

### Settings State (use-settings.ts)

| State | Type | Description |
|-------|------|-------------|
| `FeiShuNtf` | Ref\<boolean\> | 飞书通知开关 |
| `FeiShuGuid` | Ref\<string\> | 飞书机器人 GUID |
| `MCPEnabled` | Ref\<boolean\> | MCP 服务开关 |
| `MCPHost` | Ref\<string\> | MCP 绑定地址 |
| `MCPPort` | Ref\<number\> | MCP 端口 (1-65535) |
| `MCPRunning` | Ref\<boolean\> | MCP 运行状态 |
| `MCPLoading` | Ref\<boolean\> | MCP 加载状态 |

### Intercept State (use-intercept.ts)

| State | Type | Description |
|-------|------|-------------|
| `messages` | Ref\<InterceptedMessage[]\> | 被劫持的消息列表 |
| `showModal` | Ref\<boolean\> | 显示劫持消息弹窗 |
| `interceptEnabled` | Ref\<boolean\> | 劫持开关状态 |

### ServerLog State (use-server-logs.ts)

| State | Type | Description |
|-------|------|-------------|
| `serverLogs` | Reactive\<ServerLogEntry[]\> | 已捕获的服务端日志（最大 2000 条） |
| `levelFilters` | Reactive\<object\> | 级别过滤开关 (DEBUG/INFO/WARN/ERROR) |
| `searchQuery` | Ref\<string\> | 搜索关键词 |
| `filteredLogs` | Computed\<ServerLogEntry[]\> | 过滤后的日志列表 |
| `stats` | Computed\<object\> | 各级别日志计数 |

**Event Listener Pattern**: 模块级别注册（非 onMounted），与 use-intercept.ts 一致

```typescript
// Wails v3 Events.On 回调接收 WailsEvent 对象，含 .data 属性
Events.On('serverLog', (ev: any) => {
    const payload = ev?.data ?? ev
    // 兼容数组/对象两种格式
})
```

**Wails Events**:
```
Backend                     Frontend
─────────────────────────────────────────────
Emit("serverLog", map)     → Events.On("serverLog")
                            │
                            ▼
                      insertLog(entry) → serverLogs reactive array
```

### Composable Details

#### use-settings.ts

**Purpose**: 设置状态管理

**Exports**:
```typescript
export function useSettings() {
  return {
    // 飞书配置
    FeiShuNtf: Ref<boolean>
    FeiShuGuid: Ref<string>

    // MCP 配置
    MCPEnabled: Ref<boolean>
    MCPPort: Ref<number>
    MCPHost: Ref<string>
    MCPRunning: Ref<boolean>
    MCPLoading: Ref<boolean>

    // 函数
    loadAllSettings: () => Promise<void>
    saveFeishuConfig: () => Promise<void>
    saveMCPConfig: () => Promise<void>
    toggleMCPEnabled: () => Promise<void>
  }
}
```

**Functions**:

| Function | Description |
|----------|-------------|
| `loadAllSettings()` | 加载所有配置（飞书、MCP） |
| `saveFeishuConfig()` | 保存飞书配置 |
| `saveMCPConfig()` | 保存 MCP 配置 |
| `toggleMCPEnabled()` | 切换 MCP 开关（启动/停止服务） |

#### use-intercept.ts

**Purpose**: 消息劫持状态管理

**Key Feature**: 事件监听器在模块加载时立即设置，确保任何页面都能收到劫持消息

**Exports**:
```typescript
// 状态
messages: Ref<InterceptedMessage[]>
showModal: Ref<boolean>
interceptEnabled: Ref<boolean>

// 函数
loadMessages: () => Promise<void>
toggleIntercept: (enabled: boolean) => Promise<void>
clearMessages: () => Promise<void>
checkInterceptEnabled: () => Promise<void>

// 类型定义
interface InterceptedMessage {
  id: string
  robotGuid: string
  msgType: string    // "text" | "interactive"
  content: string
  timestamp: string
}
```

**Functions**:

| Function | Description |
|----------|-------------|
| `loadMessages()` | 加载劫持消息列表 |
| `toggleIntercept(enabled)` | 切换劫持开关 |
| `clearMessages()` | 清空劫持消息 |
| `checkInterceptEnabled()` | 检查劫持状态 |

**Wails Events**:
```typescript
Events.On("feishu:intercepted", (ev: any) => {
  // 在模块加载时设置，确保任何页面都能收到
  if (ev && ev.data) {
    messages.value.unshift(ev.data)
    showModal.value = true
  }
})
```

## Data Flow Diagrams

### 加载配置流程

```
onMounted()
    │
    ▼
loadAllSettings()
    │
    ├──► SettingsService.GetFeiShuConfig()
    │         │
    │         ▼
    │    FeiShuNtf.value = config.FeiShuNtf
    │    FeiShuGuid.value = config.FeiShuGuid
    │
    └──► SettingsService.GetMCPConfig()
              │
              ▼
         MCPEnabled.value = config.Enabled
         MCPPort.value = config.Port
         MCPHost.value = config.Host
         MCPRunning.value = config.Running
```

### 消息劫持流程

```
模块加载时 (use-intercept.ts)
    │
    ▼
设置 Wails 事件监听器
    │
    ├──► Events.On("feishu:intercepted", handler)
    │
    ▼
等待事件...
    │
    ▼
后端发送 "feishu:intercepted" 事件
    │
    ▼
事件处理器触发
    │
    ├──► messages.value.unshift(msg)
    └──► showModal.value = true
```

### 切换劫持开关流程

```
User toggles intercept switch
    │
    ▼
@update:value
    │
    ▼
toggleIntercept(enabled)
    │
    ├──► InterceptService.SetEnabled(enabled)
    │
    ▼
interceptEnabled.value = enabled
```

### 保存飞书配置流程

```
User edits 飞书配置
    │
    ▼
@update:value/@blur
    │
    ▼
saveFeishuConfig()
    │
    ├──► SettingsService.SetFeiShuConfig({
    │         FeiShuNtf: FeiShuNtf.value,
    │         FeiShuGuid: FeiShuGuid.value
    │       })
    │
    ▼
Config saved to backend
```

### 保存 MCP 配置流程

```
User edits MCP 配置
    │
    ▼
@update:value/@blur
    │
    ▼
saveMCPConfig()
    │
    ├──► SettingsService.SaveMCPConfig({
    │         Enabled: MCPEnabled.value,
    │         Port: MCPPort.value,
    │         Host: MCPHost.value
    │       })
    │
    ▼
Config saved to backend
    │
    └──► Service auto-restart
```

### 切换 MCP 开关流程

```
User toggles MCP switch
    │
    ▼
@update:value
    │
    ▼
toggleMCPEnabled()
    │
    ├──► MCPLoading.value = true
    │
    ├──► SettingsService.SetMCPEnabled(MCPEnabled.value)
    │
    ├──► Save config
    │
    ▼
Wait for service to start/stop
    │
    ▼
Check MCP service status
    │
    ├──► MCPRunning.value = status.Running
    │
    └──► MCPLoading.value = false
```

## Backend Services

### SettingsService

**Methods**:
- `GetFeiShuConfig()` - 获取飞书配置
- `SetFeiShuConfig(config)` - 设置飞书配置
- `GetMCPConfig()` - 获取 MCP 配置
- `SaveMCPConfig(config)` - 保存 MCP 配置
- `SetMCPEnabled(enabled)` - 设置 MCP 开关
- `GetMCPStatus()` - 获取 MCP 运行状态

### InterceptService

**Methods**:
- `SetEnabled(enabled)` - 设置劫持开关
- `IsEnabled()` - 检查劫持开关状态
- `GetMessages()` - 获取劫持消息列表
- `ClearMessages()` - 清空劫持消息

### Config Types

```typescript
interface FeiShuConfig {
  FeiShuNtf: boolean
  FeiShuGuid: string
}

interface MCPConfig {
  Enabled: boolean
  Port: number       // 1-65535
  Host: string       // e.g., "127.0.0.1"
}

interface MCPStatus {
  Running: boolean
}

interface InterceptedMessage {
  id: string
  robotGuid: string
  msgType: string    // "text" | "interactive"
  content: string
  timestamp: string
}
```

## Event Flow

### Component Events

```
Component          Event                  Handler                  Effect
────────────────────────────────────────────────────────────────────────────
n-switch (飞书)    @update:value          → saveFeishuConfig()     保存飞书配置
n-input (GUID)     @blur                  → saveFeishuConfig()     保存 GUID
n-switch (劫持)    @update:value          → toggleIntercept()      切换劫持开关
n-button (查看)    @click                 → showModal = true        显示消息列表
n-switch (MCP)     @update:value          → toggleMCPEnabled()     启动/停止 MCP
n-input (Host)     @blur                  → saveMCPConfig()        保存绑定地址
n-input-number     @blur                  → saveMCPConfig()        保存端口
```

### Wails Events

```
Backend                     Frontend
─────────────────────────────────────────────
发送 "feishu:intercepted"  → Events.On("feishu:intercepted")
                            │
                            ▼
                      更新消息列表 + 显示弹窗
```

## Related Files

### Main Page
- [index.md](./index.md) - 主页面布局

### Components
- `components/intercept-modal.vue` - 劫持消息弹窗组件
- `components/log-panel.vue` - 日志查看面板组件

## Notes

### Important Design Decisions

1. **事件监听器设置时机**: 在 `use-intercept.ts` 模块加载时立即设置，而不是在组件的 `onMounted` 中。这确保：
   - 用户在任何页面执行 Excel 检查都能收到劫持消息
   - 不需要先打开设置页面才能接收消息

2. **消息劫持优先级**: 劫持功能优先于飞书通知
   - 劫持开启时，消息不会发送到飞书
   - 劫持关闭时，正常发送飞书消息

3. **状态持久化**: 劫持开关状态保存在后端，页面刷新后保持

### General Notes

- 页面加载时自动加载配置（onMounted）
- 飞书通知用于功能测试完成后发送消息到飞书群
- **消息劫持功能用于测试阶段，避免频繁发送飞书消息**
- MCP 服务用于外部 AI 工具调用（如 Claude Code）
- MCP 配置修改后会自动重启服务
- 设置实时保存到后端
- MCP 启动/停止有 loading 状态
- 运行状态实时显示（"运行中"/"已停止"）
