# Stream Proxy 数据流

## 三个页签独立数据管理

**架构变更（2026-06-04）**：移除共享 composable 状态，每个页签组件独立管理数据。

```
发包改包页签 (packet-tab.vue)
├── 独立 recordData
├── 独立 messages
├── 独立 selectedIndex
├── 独立录制/重放/多选状态
├── filterMode: boolean              # 拦截模式开关
└── interceptedSeqIDs: Set<number>   # 被拦截但未放行的消息 SeqID 集合

测试用例页签 (testcase-tab.vue)
├── 独立 recordData
├── 独立 messages
├── 独立 selectedIndex
└── 独立用例管理状态

重放结果页签 (replay-result-tab.vue)
├── 独立 replayResults
├── 独立 currentResult
└── 独立 selectedIndex
```

## 事件路由（双事件通道架构）

后端 EmitReplayMessage 同时发射两个事件：

| 事件 | 数据类型 | 消费者 | 用途 |
|------|---------|--------|------|
| record:progress | `{latest_msg: {...}}` | packet-tab.vue | 追加到录制/重放响应表格 |
| record:intercepted | `{latest_msg: {...}, conn_id}` | packet-tab.vue | 推送被拦截的 Req 消息 |
| replay:result | `{msg_id, msg_name, direction, payload}` | index.vue | 追加到重放结果页签表格 |
| replay:progress | `{status, total, sent}` | index.vue | 仅状态管理（开始/完成/错误） |

来源判断：子组件在触发重播时 emit('replay-start', source) 通知 index.vue，而非在事件回调中判断 activeTab。

## 拦截模式流程

```
客户端发送 Req ──> RecordWorker
    │
    ▵
filterMode=true 且 MsgID >= 1000 ?
    │
    ├── 是 → 拦截，不转发到服务端
    │   │
    │   ▵
    │   Event.Emit('record:intercepted', {latest_msg, conn_id})
    │   │
    │   ▵
    │   packet-tab.vue 追加到表格 → 自动选中
    │   │
    │   ▵
    │   用户编辑 Payload → 点击"重发"（语义变为"放行"）
    │   │
    │   ▵
    │   SendMessages() 发送修改后的 Req 到服务端
    │   │
    │   ▵
    │   从 interceptedSeqIDs 移除，UI 恢复普通样式
    │
    └── 否 → 原样转发到服务端（框架消息或 Ack/Ntf）
```

## 拦截消息 UI 状态

| 状态 | 视觉表现 | 含义 |
|------|---------|------|
| 普通行 | 默认背景 | 正常录制/重放的消息 |
| 被拦截行 | 左侧橙色边框 + 微黄背景 | 被拦截等待编辑的 Req |
| 选中行 | 主题色高亮背景 | 当前正在查看/编辑的行 |
| 已放行 | 边框消失，恢复普通行样式 | 拦截消息已发送给服务端 |
