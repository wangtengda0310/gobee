# gameactor AI 开发指南

> 本文档指导 AI 助手进行 gameactor 模块的开发和维护。

## 项目定位

gameactor 是一个基于哈希路由的 Go 线程分发系统，专为游戏服务器设计。

**核心价值**：
- 简单易用（DispatchBy 闭包捕获）
- 顺序保证（相同哈希串行执行）
- 高性能（不同哈希并行执行）

## 快速参考

| 文档 | 路径 | 说明 |
|------|------|------|
| 用户文档 | @gameactor/README.md | API 文档、使用示例、Skill 安装 |
| AI Skill | @gameactor/.claude/skills/gameactor/ | AI Skill 源文件（唯一修改位置） |
| 架构设计 | @gameactor/DESIGN.md | 技术架构、设计决策 |
| 开发进度 | @gameactor/TODO.md | 待办事项、技术债务 |
| 记忆索引 | @MEMORY.md | 项目概览、快速参考 |

## 开发规范

### 1. 代码注释标准

所有代码必须包含针对代码审查者的详细注释：

#### 函数注释模板

```go
// DispatchBy 提交任务到指定哈希的 Actor 异步执行
//
// 参数:
//   - hash: 用于路由的哈希值，相同 hash 的任务将按顺序在同一 Actor 中执行
//   - handler: 无参任务处理函数，通过闭包捕获所需参数
//
// 返回:
//   - error: 分发器已关闭时返回 ErrDispatcherClosed
//
// 并发安全:
//   - 本函数是并发安全的，可从多个 goroutine 同时调用
//   - handler 将在目标 Actor 的 goroutine 中串行执行
//
// 注意:
//   - handler 执行过程中 panic 会被捕获
//   - 避免在 handler 中执行长时间阻塞操作
//
// 示例:
//   gameactor.DispatchBy(playerID, func() {
//       processPlayer(playerID, data)
//   })
func DispatchBy(hash uint64, handler func()) error
```

#### 关键设计决策注释

```go
// 使用 sync.Once 确保 Dispatcher 单例
// 原因: 多次初始化会导致 Actor 状态不一致
var initOnce sync.Once

// 使用独立的全局锁保护 shutdownHooks
// 原因: Shutdown() 可能由不同 goroutine 触发
var shutdownMutex sync.Mutex

// 使用取模运算进行哈希路由
// 原因: 简单高效，预留一致性哈希优化空间
// 算法: actorID = hash % numActors
func (d *Dispatcher) route(hash uint64) uint64 {
    return hash % d.numActors
}
```

#### 性能相关注释

```go
// 使用非阻塞 select 提交任务
// 原因: 避免提交者被阻塞，提高吞吐量
// 队列满时立即返回错误，由调用者决定重试策略
select {
case actor.queue <- task:
    return nil
default:
    return fmt.Errorf("actor %d queue is full", actorID)
}
```

### 2. TDD 开发流程

**必须遵循的顺序**：

```
1. 编写测试用例
2. 验证测试失败（红状态）
3. 实现功能代码
4. 验证测试通过（绿状态）
5. 优化重构
```

**禁止行为**：
- ❌ 跳过测试直接编写实现代码
- ❌ API 设计未确定就开始实现
- ❌ 测试用例不完整就开始实现

### 3. 测试隔离

**使用 TestDispatcher 进行隔离测试**：

```go
func TestPlayerLogic(t *testing.T) {
    // ✅ 推荐：使用 TestDispatcher
    td := gameactor.NewTestDispatcher(t, gameactor.DefaultConfig())
    defer td.Shutdown(5 * time.Second)

    td.DispatchBy(1001, func() {
        // 测试逻辑
    })
}

// ❌ 不推荐：使用全局 API（测试间会干扰）
func TestPlayerLogic(t *testing.T) {
    gameactor.Init(gameactor.DefaultConfig())
    defer gameactor.Shutdown(5 * time.Second)
    // ...
}
```

**原因**：全局 API 使用 sync.Once，测试间状态会干扰。

### 4. API 设计原则

#### 简单优先

```go
// ✅ 推荐：DispatchBy - 最简单
gameactor.DispatchBy(playerID, func() {
    attackMonster(playerID, damage)
})

// ❌ 不推荐：Dispatch - 需要创建 Task 对象
task := gameactor.NewTaskWithHash(playerID, func() error {
    return attackMonster(playerID, damage)
})
gameactor.Dispatch(task)
```

#### 闭包捕获参数

```go
// ✅ 推荐：闭包捕获
playerID := 1001
damage := 100
gameactor.DispatchBy(playerID, func() {
    attackMonster(playerID, damage)
})

// ❌ 不推荐：结构体传参
type AttackParams struct {
    PlayerID uint64
    Damage   int
}
params := AttackParams{PlayerID: 1001, Damage: 100}
gameactor.DispatchBy(params.PlayerID, func() {
    attackMonster(params.PlayerID, params.Damage)
})
```

#### 一致的命名

| 类型 | 异步 | 同步 | Context |
|------|------|------|---------|
| 便利函数 | `DispatchBy` | `DispatchBySync` | `DispatchByCtx` |
| Hashable | `Dispatch` | `DispatchSync` | `DispatchCtx` |
| 哈希函数 | `DispatchWithFunc` | `DispatchWithFuncSync` | `DispatchWithFuncCtx` |
| 直接哈希 | `DispatchWithHash` | `DispatchWithHashSync` | `DispatchWithHashCtx` |

### 5. 错误处理

#### Panic 恢复

```go
// Actor 中必须捕获 panic
func (a *actor) executeTask(task Task) {
    defer func() {
        if r := recover(); r != nil {
            // 调用 panicHandler
            if a.dispatcher.panicHandler != nil {
                a.dispatcher.panicHandler(r)
            }
        }
    }()
    task.handler()
}
```

#### 错误返回

```go
// 同步版本必须返回错误
func DispatchBySync(hash uint64, handler func() error) error {
    // ...
    return <-done  // 返回 handler 的错误
}

// 异步版本不返回执行错误（无法获取）
func DispatchBy(hash uint64, handler func()) error {
    // 只返回分发器错误
}
```

## 常见任务

### 添加新的 Dispatch 函数

1. 在 `api.go` 中添加函数签名
2. 在 `isolation_test.go` 中添加测试
3. 实现函数逻辑
4. 运行测试验证

**模板**：

```go
// DispatchXxxx 描述函数功能
//
// 参数:
//   - xxx: 参数说明
//
// 返回:
//   - error: 错误说明
//
// 示例:
//   gameactor.DispatchXxxx(...)
func DispatchXxxx(...) error {
    if globalDispatcher == nil || !globalDispatcher.IsRunning() {
        return ErrNotInitialized
    }
    // 实现逻辑
}
```

### 修复 Bug

1. 在 `isolation_test.go` 中添加失败测试
2. 修复代码
3. 验证测试通过
4. 检查是否有类似问题

### 性能优化

1. 运行基准测试：`go test -bench=. -benchmem`
2. 使用 pprof 分析：`go test -cpuprofile=cpu.prof`
3. 优化热点代码
4. 验证性能提升

## 项目结构

```
gameactor/
├── api.go              # 对外 API（优先阅读）
├── dispatcher.go       # 核心实现
├── testing.go          # 测试工具
├── isolation_test.go   # 隔离测试（参考）
├── api_test.go         # API 测试用例（TDD）
├── examples/basic/     # 使用示例
├── README.md           # 用户文档
├── DESIGN.md           # 架构设计
├── TODO.md             # 开发进度
└── .claude/skills/     # AI Skill 源文件（唯一修改位置）
    └── gameactor/
        ├── SKILL.md
        ├── README.md
        └── abilities/
            └── AI指导.md
```

## AI Skill 维护

### Skill 文件位置

**唯一修改位置**：`.claude/skills/gameactor/`

```
gameactor/.claude/skills/gameactor/
├── README.md       # Skill 目录说明
├── SKILL.md        # 主 skill 文件（用户和 AI 都读）
└── abilities/
    └── AI指导.md   # AI 能力指导（AI 读）
```

### Skill 文件结构

#### SKILL.md 格式

**文件开头必须是 YAML frontmatter**：

```yaml
---
name: gameactor 线程分发系统
description: 基于哈希路由的 Go 线程分发系统，专为游戏服务器设计
---
```

**内容章节**：
1. 项目定位
2. 快速开始
3. 核心概念
4. 使用场景（5+ 个）
5. 项目结构
6. 最佳实践
7. 常见问题
8. 运行测试
9. 依赖
10. 相关文档

#### AI指导.md 格式

**内容章节**：
1. AI 触发场景（5+ 个）
2. AI 约束和原则
3. 快速决策树
4. 示例对话（2+ 个）
5. 代码引用格式
6. 常见代码模式

### 修改 Skill 文件的流程

```
1. 修改 .claude/skills/gameactor/ 中的文件
2. cd cmd/install-skill
3. make generate    # 同步到 skills 目录
4. make build       # 编译安装工具
5. make install     # 安装到全局
6. 重启 Claude Code
```

### Makefile 命令

```bash
cd cmd/install-skill

make generate   # 同步 skill 文件到嵌入目录
make clean      # 清理生成的文件
make test       # 运行测试
make build      # 编译安装工具
make install    # 编译并安装 skill
make help       # 查看所有命令
```

### Skill 文件编写原则

1. **面向 AI**：使用简洁的指令和示例
2. **代码引用**：使用 `file:line` 格式
3. **渐进式披露**：先简单后复杂
4. **场景驱动**：按用户使用场景组织
5. **表格优先**：使用表格展示对比信息

### 修改 Skill 后的检查清单

- [ ] 更新了 SKILL.md 或 AI指导.md
- [ ] 运行 `make generate` 同步文件
- [ ] 运行 `make test` 验证编译
- [ ] 运行 `make build` 生成安装工具
- [ ] 更新 README.md 中的安装说明
- [ ] 更新 TODO.md 记录进度

### 安装工具说明

**位置**：`cmd/install-skill/`

**功能**：将 skill 文件嵌入到二进制文件中，用户可以直接运行安装。

**使用方式**：

```bash
# 用户安装
go install github.com/wangtengda0310/gobee/gameactor/cmd/install-skill@latest
gameactor-install-skill

# 开发者测试
cd cmd/install-skill
go run main.go -target /tmp/test
```

**嵌入原理**：

```go
//go:generate cp -r ../../.claude/skills/gameactor ./skills
//go:embed skills/*
var skillFiles embed.FS
```

### 同步机制

**为什么需要 make generate？**

1. `//go:embed` 只能嵌入当前目录或子目录的文件
2. Skill 源文件在 `.claude/skills/gameactor/`
3. 需要复制到 `cmd/install-skill/skills/` 才能嵌入

**自动化**：

```bash
# 每次修改 Skill 后执行
make generate    # cp -r ../../.claude/skills/gameactor ./skills
make build       # go build
make install     # 运行安装工具
```

---

- [ ] 阅读 DESIGN.md 了解架构
- [ ] 阅读 TODO.md 查看是否有相关计划
- [ ] 在 isolation_test.go 中添加测试用例
- [ ] 实现功能代码
- [ ] 运行测试验证
- [ ] 更新 README.md 文档
- [ ] 更新 TODO.md 进度

## 禁止的操作

- ❌ 修改测试通过的核心算法（如路由算法）
- ❌ 移除错误处理（panic 恢复）
- ❌ 破坏 API 一致性（命名规范）
- ❌ 添加全局状态（除 Dispatcher 外）
- ❌ 在 Actor 中执行长时间阻塞操作

## 推荐的操作

- ✅ 使用 TestDispatcher 进行测试
- ✅ 添加详细注释（针对代码审查者）
- ✅ 遵循 TDD 流程
- ✅ 保持 API 简单
- ✅ 优先使用闭包捕获参数

## 联系方式

- 问题反馈：提交 GitHub Issue
- 代码审查：参考 CLAUDE.md 注释标准
- 架构讨论：参考 DESIGN.md

---

## Skill 开发完整指南

### Skill 文件类型对比

| 文件 | 读者 | 内容类型 | 更新频率 |
|------|------|----------|----------|
| SKILL.md | 用户 + AI | 使用指南、API 参考 | 中 |
| AI指导.md | AI | 触发场景、决策树 | 中 |
| README.md | 开发者 | 目录说明、快速链接 | 低 |

### Skill 内容组织原则

#### SKILL.md 结构

```
1. YAML frontmatter（name + description）
2. 项目定位（Module + 核心价值）
3. 快速开始（最简用法代码）
4. 核心概念（表格：模式、Tag、API）
5. 使用场景（5+ 个实际场景）
6. 项目结构（目录树）
7. 最佳实践（DO/DON'T）
8. 常见问题（Q&A）
9. 运行测试（命令示例）
10. 相关文档（链接）
```

#### AI指导.md 结构

```
1. AI 触发场景（5+ 个场景）
2. AI 约束和原则（必须/推荐/避免）
3. 快速决策树（ASCII 流程图）
4. 示例对话（完整交互流程）
5. 代码引用格式（file:line）
6. 常见代码模式（模板）
```

### Skill 文件编写技巧

#### 1. 使用表格展示对比

```markdown
### 加载模式

| 模式 | 何时使用 |
|------|----------|
| ModeAuto | 默认，自动检测 |
| ModeExcel | 开发环境直接读 Excel |
| ModeCSV | 生产环境读 CSV |
```

#### 2. 使用代码块展示示例

```markdown
### 场景 1: 基础加载

```go
loader := config.NewLoader[Equipment](
    "config/装备表.xlsx",
    "武器",
    config.LoadOptions{Mode: config.ModeAuto},
)
items, err := loader.Load()
```
```

#### 3. 使用 DO/DON'T 列表

```markdown
### DO 推荐

- ✅ 生产环境使用 `ModeCSV`
- ✅ 测试环境使用 `ModeMemory`

### DON'T 避免

- ❌ 不要在生产环境直接读取 Excel
- ❌ 不要忘记处理 `Load()` 返回的 error
```

#### 4. 使用代码引用

```markdown
- 条件字段实现: `gameconfig/internal/config/conditional_test.go:100`
- 映射逻辑: `gameconfig/internal/config/mapper.go:50`
```

### Skill 同步工作流

```
修改 Skill 源文件
    ↓
cd cmd/install-skill
    ↓
make generate
    ↓
# 验证同步
ls skills/gameactor/
    ↓
make test
    ↓
# 修复问题（如有）
    ↓
git add .claude/skills/ cmd/install-skill/
    ↓
git commit -m "docs: update skill"
```

### Skill 发布流程

1. **修改 Skill 内容**
   - 更新 SKILL.md 或 AI指导.md
   - 确保格式正确（YAML frontmatter）

2. **同步并测试**
   ```bash
   cd cmd/install-skill
   make generate
   make test
   ```

3. **编译安装工具**
   ```bash
   make build
   # 验证可执行文件已生成
   ls ../../gameactor-install-skill.exe
   ```

4. **本地测试**
   ```bash
   # 测试安装到临时目录
   ../../gameactor-install-skill.exe -target /tmp/gameactor-skill
   # 验证文件已正确安装
   ls /tmp/gameactor-skill/
   ```

5. **提交代码**
   ```bash
   git add .claude/skills/ cmd/install-skill/
   git commit -m "docs: update AI skill content"
   git push
   ```

6. **发布新版本**
   ```bash
   # 用户安装新版本
   go install github.com/wangtengda0310/gobee/gameactor/cmd/install-skill@latest
   ```

### 常见问题

#### Q: 修改 Skill 后用户看不到更新？

**A**: 用户需要重新运行安装工具：
```bash
go install github.com/wangtengda0310/gobee/gameactor/cmd/install-skill@latest
gameactor-install-skill
```

#### Q: 如何验证 Skill 文件格式正确？

**A**: 运行安装工具的 `help` 命令：
```bash
cd cmd/install-skill
go run main.go -help
```

#### Q: 技能文件放在哪里？

**A**: 唯一修改位置是 `.claude/skills/gameactor/`：
```
gameactor/.claude/skills/gameactor/
├── SKILL.md        # 主 skill 文件
├── README.md       # 目录说明
└── abilities/
    └── AI指导.md   # AI 能力指导
```

#### Q: 如何添加新的 AI 触发场景？

**A**: 在 `abilities/AI指导.md` 中添加：

1. 在 "AI 触发场景" 部分添加新场景
2. 更新 "快速决策树" 添加对应分支
3. 添加示例对话（如果复杂）
4. 运行 `make generate && make test` 验证

#### Q: 代码引用格式是什么？

**A**: 使用 `file:line` 格式：
```
- 函数实现: `gameactor/api.go:50`
- 测试示例: `gameactor/isolation_test.go:100`
- 核心逻辑: `gameactor/dispatcher.go:200`
```

---

## Skill 文件示例模板

### SKILL.md 开头模板

```markdown
---
name: gameactor 线程分发系统
description: 基于哈希路由的 Go 线程分发系统，专为游戏服务器设计
---

# gameactor 线程分发系统

## 项目定位

...

## 快速开始

### 最简用法

...
```

### AI指导.md 开头模板

```markdown
# AI 使用 gameactor 指导

## AI 触发场景

### 1. 代码生成

用户请求示例：
- "生成玩家 Actor 代码"
...

AI 行为：
...
```
