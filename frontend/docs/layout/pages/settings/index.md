# Settings - 设置页面

> File path: `src/pages/settings/index.vue`
> Route: `/Settings`

## Overview

全局配置管理页面，提供飞书通知、消息劫持和 MCP 服务的配置界面。采用简单的卡片式布局。

## ASCII Layout Diagram

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                                                                              │
│  ┌────────────────────────────────────────────────────────────────────────┐  │
│  │ 飞书通知配置                                            [n-card]       │  │
│  │ ┌──────────────────────────────────────────────────────────────────┐  │  │
│  │ │ 飞书通知:           [关闭] 开启                                   │  │  │
│  │ │ 机器人GUID:         [________________]                            │  │  │
│  │ │ ────────────────────────────────────────────────────────────── │  │  │
│  │ │ 消息劫持:           [关闭] 开启        [查看消息]               │  │  │
│  │ │                                                                  │  │  │
│  │ │ 开启劫持后消息不会发送到飞书，改为弹窗显示（用于测试）            │  │  │
│  │ └──────────────────────────────────────────────────────────────────┘  │  │
│  └────────────────────────────────────────────────────────────────────────┘  │
│                                                                              │
│  ┌────────────────────────────────────────────────────────────────────────┐  │
│  │ MCP 服务配置                                            [n-card]       │  │
│  │ ┌──────────────────────────────────────────────────────────────────┐  │  │
│  │ │ 启用 MCP 服务:      [关闭] 开启              [运行中]             │  │  │
│  │ │ 绑定地址:          [127.0.0.1]                                 │  │  │
│  │ │ 端口:              [_____]                                      │  │  │
│  │ │                                                                  │  │  │
│  │ │ MCP 服务用于外部 AI 工具调用，修改后自动重启                      │  │  │
│  │ └──────────────────────────────────────────────────────────────────┘  │  │
│  └────────────────────────────────────────────────────────────────────────┘  │
│                                                                              │
│  ┌────────────────────────────────────────────────────────────────────────┐  │
│  │ 策划配表目录                                            [n-card]       │  │
│  │ ┌──────────────────────────────────────────────────────────────────┐  │  │
│  │ │ Excel目录:          [../../config]                          │  │  │
│  │ │                                                                  │  │  │
│  │ │ 统一配置各模块读取的策划配表目录                                  │  │  │
│  │ └──────────────────────────────────────────────────────────────────┘  │  │
│  └────────────────────────────────────────────────────────────────────────┘  │
│                                                                              │
│  ┌────────────────────────────────────────────────────────────────────────┐  │
│  │ 日志管理                                                [n-card]       │  │
│  │ ┌──────────────────────────────────────────────────────────────────┐  │  │
│  │ │                                                          [查看日志]│  │
│  │ │                                                                  │  │  │
│  │ │ 查看后端运行日志，支持实时推送                                    │  │  │
│  │ └──────────────────────────────────────────────────────────────────┘  │  │
│  └────────────────────────────────────────────────────────────────────────┘  │
│                                                                              │
│  ┌────────────────────────────────────────────────────────────────────────┐  │
│  │ 服务端日志                                              [n-card]       │  │
│  │ ┌──────────────────────────────────────────────────────────────────┐  │  │
│  │ │ 已捕获日志:    592 条                                            │  │  │
│  │ │                                                  [查看服务端日志]│  │  │
│  │ │                                                                  │  │  │
│  │ │ 自动捕获 Go 后端输出，支持手动标记重要事件                        │  │  │
│  │ └──────────────────────────────────────────────────────────────────┘  │  │
│  └────────────────────────────────────────────────────────────────────────┘  │
│                                                                              │
│                               （可滚动）                                      │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
```

## Component Tree Structure

```
pages/settings/index.vue
├── n-scrollbar
│   └── .settings-container
│       ├── n-card (飞书通知配置)
│       ├── n-card (MCP 服务配置)
│       ├── n-card (策划配表目录)
│       ├── n-card (日志管理)
│       └── n-card (服务端日志)
├── LogPanel
├── ServerLogPanel
└── InterceptModal
```

## Layout Areas

| Area | Size | Description |
|------|------|-------------|
| Settings Container | 100% | 配置卡片容器 |
| Card | Auto | 配置卡片 |
| Setting Row | Auto | 单行配置 |
| Setting Label | min-width 100px | 配置项标签 |

## Component File Mapping

| Component | File Path | Line | Description |
|-----------|-----------|------|-------------|
| Main Page | pages/settings/index.vue | 35-101 | n-scrollbar + container |
| Feishu Card | pages/settings/index.vue | 39-63 | 飞书通知和劫持配置 |
| MCP Card | pages/settings/index.vue | 66-84 | MCP 服务配置 |
| Log Panel | pages/settings/index.vue:97 | components/log-panel.vue | 日志查看面板 |
| Server Log Panel | pages/settings/index.vue:114 | components/server-log-panel.vue | 服务端日志抽屉面板 |
| Intercept Modal | pages/settings/index.vue:100 | components/intercept-modal.vue | 劫持消息弹窗 |
| IPFS Panel | pages/settings/index.vue | components/ipfs-panel.vue | IPFS 分布式存储抽屉面板 |

## Key State

| 状态 | 文件 | 说明 |
|------|------|------|
| `FeiShuNtf` | use-settings.ts | 飞书通知开关 |
| `FeiShuGuid` | use-settings.ts | 飞书机器人 GUID |
| `MCPEnabled` | use-settings.ts | MCP 服务开关 |
| `MCPHost` | use-settings.ts | MCP 绑定地址 |
| `MCPPort` | use-settings.ts | MCP 端口 |
| `MCPRunning` | use-settings.ts | MCP 运行状态 |
| `interceptEnabled` | use-intercept.ts | 消息劫持开关 |
| `messages` | use-intercept.ts | 被劫持的消息列表 |
| `serverLogs` | use-server-logs.ts | 已捕获的服务端日志 |

## Data Flow

```
Switch/Input Change → v-model → @blur → saveXxxConfig() → SettingsService
```

## Event Flow

| 组件 | 事件 | 处理 |
|------|------|------|
| n-switch (飞书) | @update:value | saveFeishuConfig() |
| n-input (GUID) | @blur | saveFeishuConfig() |
| n-switch (劫持) | @update:value | toggleIntercept() |
| n-button (MCP) | @update:value | toggleMCPEnabled() |
| n-input (Host/Port) | @blur | saveMCPConfig() |

### Wails Events

`feishu:intercepted` → 更新消息列表 + 显示弹窗

## Backend Services

| Service | 文件 | 说明 |
|---------|------|------|
| Feishu Config | `FuncCaseConfigService` | `.rain-qa-func.json` (section: `function_test`) |
| Intercept | `InterceptService` | 消息劫持开关/列表 |
| MCP Config | `MCPConfigService` | `.rain-qa-func.json` (section: `mcp`)，修改后自动重启 |
| Excel Config | `ExcelConfigService` | `.rain-qa-func.json` (section: `excel_config`)，统一策划配表目录 |

## Notes

- 页面加载时自动加载配置（onMounted）
- 飞书通知用于功能测试完成后发送消息到飞书群
- **消息劫持功能用于测试阶段，避免频繁发送飞书消息**
- 劫持事件监听器在模块加载时设置，确保任何页面都能收到消息
- MCP 服务用于外部 AI 工具调用（如 Claude Code）
- 设置实时保存到后端
