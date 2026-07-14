# llm/ - AI 聊天助手页面

路由: `/Home`

## 职责

提供 AI 聊天界面，支持 Anthropic/OpenAI 模型的流式对话。通过 Wails Events 实现后端到前端的流式响应传输。

## 文件清单

| 文件 | 作用 |
|------|------|
| `index.vue` | 聊天主页面（消息列表 + 输入框） |
| `components/chat-config-panel.vue` | API 配置抽屉（模型选择、API Key、Base URL） |
| `components/chat-message-item.vue` | 消息气泡组件（区分用户/AI） |
| `composables/chat.types.ts` | 类型定义（ChatMessage、ChatConfig、AnthropicConfig） |
| `composables/use-chat-state.ts` | 状态管理（消息列表、流式状态、输入、配置） |

## 关键数据流

1. 用户输入 -> 前端调用后端服务 -> 后端发起 AI API 请求
2. 后端通过 Wails Events（`chatStreamChunk`/`chatStreamDone`/`chatStreamError`）推送流式响应
3. 前端 `use-chat-state.ts` 维护 `streamingContent` ref，实时更新 AI 回复内容
4. 事件带序列号，防止乱序

## 开发注意事项

- `use-chat-state.ts` 使用模块级响应式变量，状态在页面切换时保持
- 支持多模型切换（Anthropic/OpenAI），配置结构定义在 `chat.types.ts`
- 流式响应的事件名和序列号机制在前后端之间耦合，修改需同步 Go 后端

## E2E 测试

| 测试文件 | Page Object | 覆盖范围 |
|----------|-------------|----------|
| `e2e/home/home.spec.ts` | [`HomePage`](../../e2e/shared/pages/HomePage.ts) | 页面加载（聊天容器/输入框/按钮）、欢迎消息、消息输入（可输入/多行/长文本/特殊字符/焦点状态/空消息不发送）、输入区域按钮、配置面板布局（抽屉/表单元素）、流式生成（跳过 — 需 mock 后端） |

