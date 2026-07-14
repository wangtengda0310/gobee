# pkg/ 目录规范

本目录包含所有后端功能模块。每个子目录是一个独立的业务包，通过 `llm.go` 为 AI Agent 提供工具。

## 模块文档组织规范

每个 pkg 子目录的 `CLAUDE.md` 必须包含以下章节，顺序固定：

```markdown
# <模块名>

## 1. requirement.ts 规范（前端依赖索引）

| Service | 后端文件 | 对应前端组件 | requirement.ts |
|---------|---------|------------|---------------|

## 2. wails.go 索引（后端接口索引）

| Service | 文件 | 职责 | 对应前端组件 |
|---------|------|------|------------|

## 3. 设计决策 / 时序图 / 已知问题
```

## Service 拆分规范

### 拆分原则

**按前端组件边界拆分 Service**，而非按功能类型。

一个前端功能组件（有独立数据/交互逻辑的）对应一个 Wails Service。

### 禁止

- 单 Service 文件超过 500 行
- 一个 Service 被多个不相关的前端组件共享
- Service 方法直接暴露内部数据结构
- wails.go 中包含业务逻辑

### Service 文件命名

```
service_<组件名>.go

示例：
service_test_case.go      # 对应 case-selector.vue
service_record_file.go    # 对应 message-table.vue + payload-editor.vue
service_replay_control.go # 对应 replay-panel.vue
```

### Service 注册

```go
// main.go
app.NewService(prototest.NewTestCaseService())
app.NewService(prototest.NewRecordFileService())
```

## llm.go 规范

每个 pkg 子目录必须包含 `llm.go`，为 AI Agent 注册工具。

```go
package xxx

func InitLLMTools(registry *tool.Registry, svc *XxxService) {
    // ...
}
```

详见 [pkg/CLAUDE.md](CLAUDE.md) llm.go 开发指导。

## 目录结构示例

```
pkg/<module-name>/
├── wails.go                    # Service 工厂 + 共享类型
├── service_<组件1>.go          # 组件1 对应 Service
├── service_<组件2>.go          # 组件2 对应 Service
├── llm.go                      # AI Agent 工具注册
├── <业务逻辑>.go               # 内部业务逻辑
└── CLAUDE.md                   # 模块文档（按本规范组织）
```

## 现有模块索引

| 模块 | 路径 | Service 拆分状态 |
|------|------|-----------------|
| proto-test | proto-test/ | 待拆分（当前单 Service） |
| excel-test | excel-test/ | 待拆分 |
| function-test | function-test/ | 待拆分 |
| hero-wiki-check | hero-wiki-check/ | 待拆分 |
| activity-wiki-check | activity-wiki-check/ | 待拆分 |
| rain-resources-checker | rain-resources-checker/ | 待拆分 |
| settings | settings/ | 待拆分 |
| common | common/ | 无 Wails Service |
