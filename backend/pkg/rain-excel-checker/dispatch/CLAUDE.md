# dispatch 包说明

通知分发策略包，负责将 rain-excel-checker 的检查结果按分支策略分发到不同通知通道。

## 职责

- **策略决策**：根据 git 分支名决定通知方式（群消息 vs 私聊）
- **Handler 组装**：按策略注册 Console/Card/DM handler，构建 Dispatcher

## 目录结构

```
dispatch/
├── strategy.go       # NotifyMode + ResolveNotifyMode 纯函数
├── strategy_test.go  # 分支策略单元测试（7 个用例）
├── router.go         # NotifyRouter：封装 Dispatcher 构建逻辑
└── router_test.go    # Router 单元测试（9 个用例）
```

## 核心接口

### strategy.go

| 函数/类型 | 说明 |
|-----------|------|
| `NotifyMode` | 通知通道开关：`{Group bool, DM bool}` |
| `ResolveNotifyMode(branch, rules, defaultMode)` | 纯函数，按分支名返回通知模式 |
| `DefaultRules()` | 预定义规则：`{"v0.0.8-pre-release": {Group: true, DM: false}}` |
| `DefaultMode()` | 默认模式：`{Group: false, DM: true}` |

### router.go

| 函数/类型 | 说明 |
|-----------|------|
| `NotifyRouter` | 封装 Dispatcher 构建逻辑 |
| `NewNotifyRouter(mode, robotGUID, dmHandler)` | 构造函数 |
| `BuildDispatcher(authors)` | 按 mode 注册 handler，返回 Dispatcher |
| `Mode()` | 返回当前通知模式 |

## 分支策略

| 分支 | 群消息 | 私聊 |
|------|--------|------|
| `v0.0.8-pre-release` | 发送 | 不发送 |
| 其他分支 | 不发送 | 发送 |

## 使用方式

```go
// main.go 中的典型用法
branch, _ := gitutil.GetCurrentBranch(config.excelPath)
mode := dispatch.ResolveNotifyMode(branch, dispatch.DefaultRules(), dispatch.DefaultMode())

dmHandler := newDMHandler(config)              // -noDM 时仍创建 handler (dryRun)
monitorClient := feishu.NewOpenAPIClient(...)
dmHandler = handlers.WrapDMMonitor(dmHandler, monitorClient) // 装饰器包装
router := dispatch.NewNotifyRouter(mode, config.feishuRobot, dmHandler)

dispatcher := router.BuildDispatcher([]string{"作者名"})
dispatcher.Dispatch(event)
```

## 执行流程

```
                          main.go
                            │
                    ┌───────┴───────┐
                    │  argParse()   │  解析命令行参数
                    └───────┬───────┘
                            │
                    ┌───────┴──────────────────┐
                    │ config.mode == "full" ?   │
                    └───┬──────────────────┬───┘
                       YES                 NO
                        │                  │
              handleFullCheck()    获取 HEAD hash
              (本地磁盘全量检查)         │
                        │       ┌────────┴────────┐
                        │       │ IsMergeCommit ? │
                        │       └───┬─────────┬───┘
                        │          YES        NO
                        │           │          │
                        │   handleMergeCommit  handleNormalCommit
                        │   (遍历子commit)     (单个commit)
                        │           │          │
                        └─────┬─────┴──────────┘
                              │
                     构建 []CommitCheckResult
                              │
                  ┌───────────┴───────────┐
                  │  dispatchResults()    │
                  │  （main.go）          │
                  │  获取当前分支名        │
                  └───────────┬───────────┘
                              │
        ┌─────────────────────┴─────────────────────┐
        │  【dispatch 包介入点】                      │
        │                                           │
        │  dispatch.ResolveNotifyMode()             │
        │  branch + DefaultRules + DefaultMode      │
        │  → NotifyMode {Group, DM}                 │
        └─────────────────────┬─────────────────────┘
                              │
                  ┌───────────┴──────────────┐
                  │ mode.Group | mode.DM     │
                  │ (true/false 组合)        │
                  └───┬──────────────────┬───┘
                {G:true,D:false}    {G:false,D:true}
                      │                  │
              预发布分支(v0.0.8)      其他分支
                      │                  │
                  只发群消息           只发私聊
                      │                  │
                      └────────┬─────────┘
                               │
                    ┌──────────┴──────────┐
                    │ newDMHandler(config)       │  main.go：从环境变量读取凭证
                    │ -noDM → dryRun=true       │  有凭证时始终创建，dryRun 跳过发送
                    │ 无凭证 → nil              │
                    └──────────┬──────────┘
                               │
        ┌──────────────────────┴──────────────────────┐
        │  【dispatch 包介入点】                        │
        │                                              │
        │  dispatch.NewNotifyRouter(mode, robot, dm)   │
        │  → NotifyRouter                              │
        └──────────────────────┬──────────────────────┘
                               │
        ┌──────────────────────┴──────────────────────┐
        │  【dispatch 包介入点】                        │
        │                                              │
        │  router.BuildDispatcher(authors)             │
        │                                              │
        │  1. notification.NewDispatcher()             │
        │  2. Register(ConsoleHandler)   ← 始终        │
        │  3. if mode.Group              ← dispatch 决策│
        │     Register(FeishuCardHandler)               │
        │  4. if mode.DM && dm != nil    ← dispatch 决策│
        │     Register(FeishuDMHandler)                 │
        │  → *CheckResultDispatcher                     │
        └──────────────────────┬──────────────────────┘
                               │
                    ┌──────────┴──────────┐
                    │  事件构建（main.go）  │
                    │  - 聚合 ColResults   │
                    │  - 统计计算          │
                    │  - 构建 Event        │
                    └──────────┬──────────┘
                               │
                    ┌──────────┴──────────┐
                    │ dispatcher.Dispatch()│  main.go
                    │                     │
                    │  → ConsoleHandler    │
                    │  → FeishuCardHandler │ (如 mode.Group)
                    │  → FeishuDMHandler   │ (如 mode.DM)
                    └─────────────────────┘
```

## 测试

```bash
cd rain-excel-checker/dispatch
go test -v
```

- `strategy_test.go`：7 个用例，覆盖分支匹配/默认/自定义规则
- `router_test.go`：9 个用例，覆盖群消息/私聊/全开/全关/边界

## 扩展指南

### 新增分支策略

修改 `DefaultRules()` 或调用 `ResolveNotifyMode` 时传入自定义 rules map：

```go
customRules := map[string]dispatch.NotifyMode{
    "v0.0.9-pre-release": {Group: true, DM: true}, // 群消息+私聊
}
mode := dispatch.ResolveNotifyMode(branch, customRules, dispatch.DefaultMode())
```

### 修改默认模式

```go
customDefault := dispatch.NotifyMode{Group: true, DM: true}
mode := dispatch.ResolveNotifyMode(branch, dispatch.DefaultRules(), customDefault)
```

## 注意事项

- `ResolveNotifyMode` 是纯函数，无副作用，易于测试
- `BuildDispatcher` 中 ConsoleHandler 始终注册，不受 mode 影响
- DM handler 为 nil 时（未配置凭证），即使 mode.DM=true 也不注册
- `-noDM` 时 handler 仍创建（dryRun=true），Handle 正常执行（允许装饰器触发 debug 监控），但 sendDM 跳过实际发送
- 装饰器 `WrapDMMonitor` 包装私聊 handler，每次 Handle 后自动发送 debug 摘要到监控邮箱
- authors 为空切片时，群消息 handler 正常注册但不带 @ 提及
