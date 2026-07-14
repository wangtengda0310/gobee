# settings/ - 全局配置管理页面

路由: `/Settings`

## 职责

提供飞书通知、MCP 服务、消息劫持、服务端日志、路线图和 WebTorrent 的配置界面。
所有配置通过 `composables/` 导出的响应式变量（如 `FeiShuNtf`、`MCPEnabled`）跨页面共享。

## 文件清单

| 文件 | 作用 |
|------|------|
| `index.vue` | 页面入口，组装各功能卡片 |
| `components/intercept-notification.vue` | 全局劫持通知条 |
| `components/intercept-detail.vue` | 消息劫持详情弹窗 |
| `components/roadmap-panel.vue` | 路线图管理面板 |
| `components/roadmap/RoadmapDetail.vue` | 功能详情（投票、评论） |
| `components/roadmap/SubmitModal.vue` | 提交新功能弹窗 |
| `components/server-log-panel.vue` | 服务端日志面板 |
| `components/ipfs-panel.vue` | IPFS 面板 |
| `components/webtorrent-panel.vue` | WebTorrent 抽屉面板 |
| `components/webtorrent/` | P2P 传输子组件 |
| `composables/use-settings.ts` | 飞书/MCP/统一策划配表目录配置状态，被其他页面引用 |
| `composables/use-intercept.ts` | 消息劫持逻辑 |

## 配置项

| 配置卡片 | 对应后端 Service | 持久化位置 |
|----------|-----------------|------------|
| 飞书通知配置 | `FuncCaseConfigService` (function-test) | `.rain-qa-func.json` / `func_case_config` |
| MCP 服务配置 | `MCPConfigService` | `.rain-qa-func.json` / `mcp` |
| 策划配表目录 | `ExcelConfigService` | `.rain-qa-func.json` / `excel_config` |
| `composables/use-server-logs.ts` | 服务端日志管理 |
| `composables/use-ipfs.ts` | IPFS 节点管理 |
| `composables/use-webtorrent.ts` | WebTorrent 客户端生命周期 |
| `composables/use-download.ts` | 下载逻辑 |
| `composables/use-seed.ts` | 做种逻辑 |
| `config/roadmap-api.ts` | 路线图 API 服务（Wails bindings 封装） |
| `config/roadmap-types.ts` | 路线图类型定义 |

## 关键数据流

- `use-settings.ts` 导出的 `FeiShuNtf`、`FeiShuGuid` 等是模块级 ref，被 `function-test` 等页面直接 import
- 路线图通过 `roadmap-api.ts` 调用 Wails 生成的 `RoadmapService` binding
- WebTorrent/IPFS 各自独立管理生命周期，不互相依赖

## 开发注意事项

- `use-settings.ts` 是跨页面共享模块，修改时注意对 `function-test/` 的影响
- 新增配置卡片只需在 `index.vue` 添加 `n-card` 和对应 composable
- 路线图类型在 `config/roadmap-types.ts` 中维护，与后端 models.js 保持兼容

## E2E 测试

| 测试文件 | Page Object | 覆盖范围 |
|----------|-------------|----------|
| `e2e/settings/settings.spec.ts` | [`SettingsPage`](../../e2e/shared/pages/SettingsPage.ts) | 页面加载（配置卡片显示）、飞书通知配置（开关/GUID 输入/消息劫持开关）、MCP 服务配置（开关/绑定地址/端口/运行状态）、策划配表目录配置（输入/保存）、服务端日志（日志计数/查看按钮）、页面布局（卡片滚动） |
| `e2e/roadmap.spec.ts` | [`RoadmapPage`](../../e2e/shared/pages/RoadmapPage.ts) | 抽屉打开（筛选栏/列表/提交按钮）、筛选排序（状态/搜索/空状态）、项目详情（打开/功能描述/投票/评论/关闭）、投票功能、评论功能（空评论不发送）、提交新建议（取消/空标题验证）、导航集成 |

