# Proto Test Page Layout

> File path: `src/pages/proto-test/index.vue`
> Route: `/ProtoTest`

## Overview

协议录制/重放页面，通过"发包改包"、"测试用例"和"重放结果"三个页签切换。发包改包和测试用例页签共享消息表格和 Payload 编辑器，顶部表单区域各自独立（通过 `protocol-content` 组件的 slot 实现）。目标服务配置区域在页签标题上方全局共享（发包改包和测试用例共用）。

**核心交互：两个独立的重放操作**
- **开始重放/执行用例**（顶部按钮）：将表格中所有 Req 发送到服务器，1 次
- **重发**（选中行后底部面板）：将选中的单条 Req 发送 N 次（用户指定次数）

**新增功能**：重放结果页签独立展示发包改包、测试用例、重发控制的重放结果，与录制/测试用例页签状态完全隔离。

## ASCII Layout Diagram

### 三个页签布局（结构相同，顶部按钮行不同）

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ *发包改包* | 测试用例 | 重放结果                                               │ ← Tab Bar
├─────────────────────────────────────────────────────────────────────────────┤
│ 目标服务: [TCP地址:18000    ] [HTTP地址:20144  ] [登录账号   ] [起始:1] [终止:1]  │ ← Target Service Row (全局共享)
├─────────────────────────────────────────────────────────────────────────────┤
│ 发包改包: [多选] [实时修改] [开始录制] [停止录制] [开始重放] 录制进度            │ ← Top Button Row
│ 测试用例: [多选] [加载用例] [执行用例] [新增模块] [删除模块] [下拉菜单...       │
│ 重放结果: [选择重放结果 ▼] [清除结果]                                          │
├─────────────────────────────────────────────────────────────────────────────┤
│ 版本: 1 | 录制时间: ... | 消息数: 42 | 来源: ...                              │ ← File Info
├─────────────────────────────────────────────────────────────────────────────┤
│ 已选 3 条 [取消多选] [保存到用例/删除/批量重放]  (多选模式下)                   │ ← MultiSelect Action Bar
├─────────────────────────────────────────────────────────────────────────────┤
│  发包改包（variant="packet"，请求列标题旁有 [过滤] 按钮，开启后只显示带 Req 的行）：    │
│  ┌────┬────┬──────────┬──────┬──────────────────┬─────────────┬──────┬──────────┬──────┐│
│  │    │ #  │ 时间      │ MsgID │ 请求(Req) [过滤] │ 响应(Ack)   │ SeqID │ 方向     │ 结果  ││ ← Message Table
│  ├────┼────┼──────────┼──────┼──────────────────┼─────────────┼──────┼──────────┼──────┤│
│  │ ⠿  │ 0  │ 10:00:00│ 1001  │ LoginReq         │ LoginAck    │ 0     │ C->S,S->C│ 成功  ││
│  │ ⠿  │ 3  │ 10:00:03│ 1006  │ BuyReq           │ (拦截中...) │ 2     │ C->S     │ 拦截  ││ ← 被拦截：橙色左边框
│  └────┴────┴──────────┴──────┴──────────────────┴─────────────┴──────┴──────────┴──────┘│
│                                                                             │
│  测试用例（variant="testcase"，无时间/SeqID/结果列，响应列改为描述列）：              │
│  ┌────┬────┬──────┬──────┬─────────────┬─────────────┬──────────┐            │
│  │    │ #  │ MsgID │ 请求(Req)   │ 描述         │ 方向     │            │
│  ├────┼────┼──────┼──────┼─────────────┼─────────────┼──────────┤            │
│  │ ⠿  │ 0  │ test │ 1001  │ GmCommandReq│ 添加粮草     │ C->S     │            │
│  └────┴────┴──────┴──────┴─────────────┴─────────────┴──────────┘            │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────────┐ │ ← 选中行后显示
│  │ 重放控制  [重发] [ 1 ] 次 [停止] [状态: 运行中 (5/5)] │ Ntf: RoomNtf     │ │    ReplayControl
│  ├─────────────────────────────────────────────────────────────────────────┤ │
│  │ HelloReq (MsgID=1001)                              [格式化] [应用]      │ │ ← ReqCardEditor
│  │ ┌────────────────────────────────────────────────────────────────────┐  │ │
│  │ │ name: [原始值▼] test1  hp: [范围▼] 100→200  heroId: [枚举▼] [h1,h2]│  │ │
│  │ └────────────────────────────────────────────────────────────────────┘  │ │
│  └─────────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Layout Dimensions

| Area | Size | Description |
|------|------|-------------|
| Tab Bar | Auto (fixed) | 发包改包 / 测试用例 页签切换 |
| Target Service Row | Auto (fixed) | 目标服务配置（TCP/HTTP/账号），页签标题上方全局共享 |
| Top (slot) | Auto (fixed) | 各页签自定义顶部区域（按钮行，多选按钮在左） |
| Record Progress | Auto (条件显示) | 录制状态标签，嵌入按钮行（在开始重放右侧） |
| File Info / Result Info | Auto (条件显示) | 发包改包/测试用例：版本/录制时间/消息数；重放结果：版本/录制时间/消息数/来源（合并单行） |
| Stop Button | Auto (条件显示) | 仅重放结果页签运行时显示 — [停止所有重放] 按钮行（信息栏和表格之间） |
| **MultiSelect Action Bar** | Auto (条件显示) | **多选模式时显示** — 已选计数 + 取消多选 + 操作按钮（文件信息和表格之间） |
| Message Table | Adaptive (flex: 1) | 消息列表表格，占据主要空间 |
| **Replay Control** | Auto (条件显示) | **仅当选中某行时显示** — 重发 + N次 + 停止 + Ntf 显示 |
| **拦截消息标记** | Auto (条件显示) | 被拦截但未放行行的橙色边框 + 微黄背景 |
| Payload Viewer | 300px (fixed) | 配对 Payload JSON 编辑器 |

## Component Tree Structure

```
pages/proto-test/index.vue                       # 页面编排层
├── Target Service Config (component)               # 目标服务配置（页签标题上方，全局共享）
│   ├── serverAddr input                           # TCP 地址输入框
│   ├── httpAddr input                             # HTTP 地址输入框
│   ├── openID input                               # 登录账号输入框
│   ├── rangeStart input                           # 起始序号输入框（账号范围迭代）
│   ├── rangeEnd input                             # 终止序号输入框（账号范围迭代）
│   ├── 设置按钮                                    # 打开重放设置抽屉
│   └── 设置抽屉 (n-drawer)                         # 发送间隔/等待Ack/最大并发
├── Tab Bar (div)                                 # 页签切换
├── ProtocolContent (component) v-if="activeTab"  # 协议内容共享组件
│   ├── #top slot                                 # 各页签自定义顶部
│   │   ├── 发包改包:
│   │   │   └── Top Button Row                    # [多选] [实时修改] [开始录制] [停止录制] [开始重放] 录制进度
│   │   │       └── 实时修改切换按钮 (filterMode)   # 开启/关闭拦截模式
│   │   └── 测试用例:
│   │       └── Case Selector + Buttons           # [多选] [加载用例] [执行用例] [新增模块] [删除模块] + 下拉菜单
│   ├── Record Progress (v-if, 嵌入按钮行内)       # 录制进度标签（在开始重放右侧）
│   ├── File Info Row (v-if)                      # 文件元信息（版本/录制时间/消息数）
│   ├── MessageTable                              # 消息列表（variant 区分页签）
│   │   ├── 发包改包 variant="packet"             # 拖柄|#|账号|时间|MsgID|请求[过滤]|响应|SeqID|方向|结果
│   │   │      请求列标题旁 [过滤] 按钮：只显示带 Req 数据的行
│   │   └── 测试用例 variant="testcase"           # 拖柄|#|MsgID|请求|描述(descript)|方向
│   │      行可拖拽排序，多选模式显示 checkbox，点击行高亮（蓝色背景）
│   │      拦截消息行：橙色左边框 + 微黄背景
│   ├── 拦截状态 (packet-tab.vue)
│   │   ├── filterMode (ref<boolean>)             # 拦截模式开关
│   │   └── interceptedSeqIDs (ref<Set<number>>)  # 被拦截但未放行的消息 SeqID 集合
│   ├── BatchActionBar (v-if=selectMode)          # 多选操作栏（文件信息和表格之间）
│   │   ├── 发包改包                              # "已选择 N 条消息" [取消多选] [保存到用例]
│   │   └── 测试用例                              # "已选择 N 条消息" [取消多选] [删除用例消息] [批量重放]
│   ├── ReplayControl (v-if=selectedPairedEntry)  # 重发控制+Ntf显示（选中行显示）
│   │   ├── 左侧控制区                            # 重发按钮 + 次数输入 + 停止按钮 + 状态标签
│   │   └── 右侧 Ntf 显示区                       # Ntf 消息名称 + 预览标签
│   └── PairedPayloadEditor (v-if=selectedEntry)  # Payload 编辑器
│       ├── ReqCardEditor                         # Req 卡片式编辑器
│       │   ├── 工具栏                            # 标题 + 格式化/应用按钮
│       │   ├── 卡片容器                          # Payload 字段卡片
│       │   │   └── FieldItem (v-for)             # 字段项（key/value 对）
│       │   │       ├── 组件选择下拉菜单             # 原始值/范围/枚举/组合/变量
│       │   │       ├── 原始值显示（只读）           # n-input (readonly)
│       │   │       ├── RangeInput (条件)          # 范围输入（hp/mp/attack等）
│       │   │       ├── EnumSelect (条件)          # 枚举值选择（heroId/itemId等）
│       │   │       ├── ComboSelect (条件)          # 组合选择（heroIds/itemIds等）
│       │   │       ├── VariableSelect (条件)       # 变量选择（从Ntf提取的动态变量）
│       │   │       ├── 嵌套对象递归                # FieldItem 递归渲染
│       │   │       └── 数组项编辑                  # 添加/删除/编辑数组元素
│       │   └── 空状态                            # 无数据/解析失败提示
│       └── AckCardEditor (TODO)                  # Ack 卡片式编辑器（待实现）
```

## Replay 操作对比

| 操作 | 触发 | 发送内容 | 次数 | 账号范围迭代 | 面板 |
|------|------|---------|------|------------|------|
| **开始重放** | 顶部按钮 | 表格全部 Req | 1 | **支持** (rangeStart~rangeEnd) | 不显示 |
| **执行用例** | 测试用例页签顶部按钮 | 从用例文件加载后全部 Req | 1 | **支持** (rangeStart~rangeEnd) | 不显示 |
| **重发** | 底部 ReplayControl 面板，选中行后 | 选中行的单条 Req | N（用户设） | 不支持 (始终单账号) | 显示 ReplayControl + ReqCardEditor |
| **迭代发送** | 底部 ReplayControl 面板 | 单条 Req 按字段配置迭代生成 | 1 | 不支持 (始终单账号) | 显示 ReplayControl + ReqCardEditor |

**拦截模式下的语义变化**：
- 拦截模式开启时，底部"重发"按钮语义变为"放行"——将编辑后的被拦截消息发送到服务端
- 放行成功后，该消息从 `interceptedSeqIDs` 集合移除，UI 恢复普通行样式

## Component File Mapping

| Component | File Path | Description |
|-----------|-----------|-------------|
| Main Page | `pages/proto-test/index.vue` | 编排层，页签切换 + slot 内容 |
| **Target Service Config** | `pages/proto-test/shared/target-service-config.vue` | **目标服务配置（TCP/HTTP/监听端口/注入unity服务器列表，页签标题上方全局共享）** |
| Message Table | `pages/proto-test/shared/message-table.vue` | 配对消息列表（可拖拽/多选） |
| Paired Payload Editor | `pages/proto-test/shared/paired-payload-editor.vue` | 配对 JSON 编辑器容器 |
| Replay Control | `pages/proto-test/shared/replay-control.vue` | 重发控制+Ntf显示（左侧重发+N次+停止，右侧Ntf预览） |
| Req Card Editor | `pages/proto-test/shared/req-card-editor.vue` | Req 卡片式编辑器（字段列表 + 格式化/应用） |
| Field Item | `pages/proto-test/shared/field-item.vue` | 单个字段编辑项（用户手动选择输入组件类型） |
| **Range Input** | `pages/proto-test/shared/range-input.vue` | **范围输入组件（起始值/步长/终值）** |
| **Enum Select** | `pages/proto-test/shared/enum-select.vue` | **枚举值选择组件（多选标签）** |
| **Combo Select** | `pages/proto-test/shared/combo-select.vue` | **组合选择组件（多选标签）** |
| **Variable Select** | `pages/proto-test/shared/variable-select.vue` | **变量选择组件（从Ntf提取的动态变量）** |
| ~~Replay Panel~~ | ~~`pages/proto-test/shared/replay-panel.vue`~~ | ~~已废弃~~（由 replay-control.vue 替代） |
| Paired Algorithm | `pages/proto-test/shared/composables/use-paired-messages.ts` | Req/Ack 配对算法 |
| Selected Entry Management | `pages/proto-test/shared/composables/use-selected-entry.ts` | 选中项管理 |

## 时序图

### 重放流程（双事件通道架构）

详见 [data-flow.md](data-flow.md) 的"录制流程"和"重放流程"章节。核心要点：

- 用户触发重放 → `emit('replay-start', source)` → index.vue 切换页签并初始化结果
- 后端 `SendMessages` 执行：HTTP登录 → TCP连接 → 逐条发送Req
- `EmitReplayMessage` 同时发射两个事件：
  - `record:progress` → packet-tab.vue 追加到录制表格
  - `replay:result` → index.vue 追加到重放结果表格

### 多账号迭代重放流程

用户在 target-service-config 设置 rangeStart/rangeEnd，点击"开始重放"或"执行用例"后，后端最外层循环遍历账号范围，每个账号独立执行：HTTP登录 → TCP连接 → 逐条发送Req → 关闭。rangeStart=rangeEnd=1 时为单账号模式。重发和迭代发送不受账号范围影响，始终单账号。

### 重发流程（底部面板）

用户选中表格某行后显示 ReplayControl + ReqCardEditor，修改字段值并"应用"后，可设置重发次数并点击"重发"。后端调用 `SendMessages` 循环发送指定次数，Ntf 推送到 ReplayControl 显示。

### 手动组件选择流程

用户选中某行展开字段卡片后，FieldItem 为每个字段渲染组件选择下拉菜单（原始值/范围/枚举/组合/变量），用户选择组件类型后渲染对应输入组件，修改字段值触发更新，应用按钮启用。
