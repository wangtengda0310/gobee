# gameconfig 开发指南

本文档面向 gameconfig 项目的**维护者和贡献者**。

**使用者文档**：参考 [README.md](README.md)

---

## 项目概览

gameconfig 是一个游戏配置管理工具，支持双模式配置加载（Excel/CSV）、条件字段、Schema 迁移、Mock 数据测试。

**Go Module**: `github.com/wangtengda0310/gobee/gameconfig`

---

## 代码结构

### 目录布局

```
gameconfig/
├── internal/config/          # 核心实现（不对外暴露）
│   ├── loader.go             # 统一加载器入口
│   ├── excel.go              # Excel 读取
│   ├── csv.go                # CSV 读取
│   ├── mapper.go             # 结构体反射映射
│   ├── condition.go          # 条件字段处理
│   ├── schema.go             # Schema 版本管理
│   ├── watcher.go            # 热重载
│   ├── exporter.go           # Excel 导出器
│   ├── comment.go            # 批注处理
│   ├── errors.go             # 错误处理
│   ├── test_helper.go        # 测试辅助函数
│   └── *_test.go             # 单元测试
├── pkg/config/               # 对外 API
│   └── config.go             # 公开接口
├── cmd/
│   ├── xlsx2csv/             # Excel 导出工具
│   └── install-skill/        # Skill 安装工具
├── tests/                    # 集成测试和回归测试
├── testdata/                 # 测试数据
├── .claude/skills/gameconfig/ # Skill 源文件（唯一修改位置）
├── scripts/                  # 构建和维护脚本
└── docs/                     # 项目文档
```

### 架构分层

```
┌─────────────────────────────────────────┐
│           pkg/config (对外 API)          │
├─────────────────────────────────────────┤
│    internal/config (核心实现)            │
│  ┌─────────┐ ┌─────────┐ ┌───────────┐  │
│  │ Loader  │ │ Mapper  │ │ Schema    │  │
│  │         │ │         │ │ Watcher   │  │
│  └─────────┘ └─────────┘ └───────────┘  │
├─────────────────────────────────────────┤
│     excelize (Excel) + fsnotify         │
└─────────────────────────────────────────┘
```

---

## 开发指南

### 添加新功能

1. 在 `internal/config/` 实现核心逻辑
2. 在 `pkg/config/` 暴露对外 API（如需要）
3. 编写单元测试 `*_test.go`
4. 在 `tests/` 添加集成测试（如适用）
5. 更新相关文档

### 代码规范

- **内部实现**：放在 `internal/config/`，不对外暴露
- **公开 API**：放在 `pkg/config/`，保持向后兼容
- **错误处理**：使用 `internal/config/errors.go` 中的错误类型
- **测试优先**：每个新功能都需要对应的测试

### 常见任务

#### 添加新的配置模式

1. 在 `internal/config/loader.go` 添加新的 Mode 常量
2. 实现对应的加载逻辑
3. 添加测试用例
4. 更新 README.md 中的模式说明

#### 添加新的 Struct Tag

1. 在 `internal/config/mapper.go` 扩展 tag 解析
2. 更新 `FieldMeta` 结构（如需要）
3. 添加测试用例
4. 更新 README.md 中的 tag 说明

---

## Skill 文件维护

### 唯一修改位置

**所有 skill 内容都在这里修改**：
```
.claude/skills/gameconfig/
├── SKILL.md              # 主 skill 文件
├── abilities/
│   └── AI指导.md          # AI 能力指导
└── README.md
```

### Skill 文件类型对比

| 文件 | 读者 | 内容类型 | 更新频率 |
|------|------|----------|----------|
| SKILL.md | 用户 + AI | 使用指南、API 参考、场景示例 | 中 |
| AI指导.md | AI | 触发场景、决策树、响应流程 | 中 |
| README.md | 开发者 | 目录说明、快速链接 | 低 |

### Skill 内容组织原则

#### SKILL.md 结构

```
1. YAML frontmatter（name + description）
2. 项目定位（Module + 核心价值）
3. 快速开始（最简用法代码）
4. 核心概念（表格：模式、Tag、格式）
5. 使用场景（5+ 个实际场景）
6. 项目结构（目录树）
7. 最佳实践（DO/DON'T）
8. 常见问题（Q&A）
9. 运行测试（命令示例）
10. 依赖（go require）
11. 相关文档（链接）
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
| ModeAuto | 默认，自动检测（优先 CSV） |
| ModeExcel | 开发环境直接读 Excel |
| ModeCSV | 生产环境读 CSV |
| ModeMemory | 测试环境使用 Mock 数据 |
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

- ✅ 生产环境使用 `ModeCSV`（Git diff 友好）
- ✅ 测试环境使用 `ModeMemory`（无需文件）

### DON'T 避免

- ❌ 不要在生产环境直接读取 Excel
- ❌ 不要忘记处理 `Load()` 返回的 error
```

#### 4. 使用代码引用

```markdown
- 条件字段实现: `gameconfig/internal/config/conditional_test.go:100`
- 映射逻辑: `gameconfig/internal/config/mapper.go:50`
- 测试示例: `gameconfig/tests/integration_test.go:200`
```

### 同步到安装工具

修改 skill 后，**必须同步**到嵌入文件：

```bash
cd cmd/install-skill
go generate
git add skills/
```

### 验证同步

```bash
# 检查文件是否同步
bash scripts/check-sync.sh

# 运行完整测试
go test ./tests/... -run TestInstallSkill -v
```

### Skill 同步工作流

```
修改 Skill 源文件
    ↓
cd cmd/install-skill
    ↓
go generate           # 同步到 skills 目录
    ↓
# 验证同步
ls skills/gameconfig/
    ↓
go test ./... -v       # 运行测试
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
   go generate
   go test ./... -v
   ```

3. **编译安装工具**
   ```bash
   go build -o ../../gameconfig-install-skill .
   # 验证可执行文件已生成
   ls ../../gameconfig-install-skill
   ```

4. **本地测试**
   ```bash
   # 测试安装到临时目录
   ../../gameconfig-install-skill -target /tmp/gameconfig-skill
   # 验证文件已正确安装
   ls /tmp/gameconfig-skill/
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
   go install github.com/wangtengda0310/gobee/gameconfig/cmd/install-skill@latest
   gameconfig-install-skill
   ```

### 常见问题

#### Q: 修改 Skill 后用户看不到更新？

**A**: 用户需要重新运行安装工具：
```bash
go install github.com/wangtengda0310/gobee/gameconfig/cmd/install-skill@latest
gameconfig-install-skill
```

#### Q: 如何验证 Skill 文件格式正确？

**A**: 检查 YAML frontmatter：
```bash
head -5 .claude/skills/gameconfig/SKILL.md
# 应该看到：
# ---
# name: gameconfig 配置管理工具
# description: ...
# ---
```

#### Q: `go generate` 失败怎么办？

**A**: 确保 skill 源文件存在：
```bash
ls -la .claude/skills/gameconfig/
# 应该看到 SKILL.md 和 abilities/AI指导.md
```

#### Q: 如何添加新的 AI 触发场景？

**A**: 在 `abilities/AI指导.md` 中添加：

1. 在 "AI 触发场景" 部分添加新场景
2. 更新 "快速决策树" 添加对应分支
3. 添加示例对话（如果复杂）
4. 运行 `go generate && go test` 验证

#### Q: 技能文件放在哪里？

**A**: 唯一修改位置是 `.claude/skills/gameconfig/`：
```
gameconfig/.claude/skills/gameconfig/
├── SKILL.md        # 主 skill 文件
├── README.md       # 目录说明
└── abilities/
    └── AI指导.md   # AI 能力指导
```

### Skill 内容原则

- **面向 AI**：使用简洁的指令和示例
- **代码引用**：使用 `file:line` 格式
- **渐进式披露**：先简单后复杂
- **场景驱动**：按用户使用场景组织

---

## 测试

### 测试结构

| 类型 | 位置 | 说明 |
|------|------|------|
| **单元测试** | `internal/config/*_test.go` | 测试单个功能 |
| **集成测试** | `tests/integration_test.go` | 测试完整流程 |
| **回归测试** | `tests/regression_test.go` | 防止 bug 再现 |
| **安装测试** | `tests/install_skill_test.go` | 测试 skill 安装 |

### 运行测试

```bash
# 单元测试
go test ./internal/config/... -v

# 集成测试
go test ./tests/... -v

# 回归测试
go test ./tests/... -run TestRegression

# 带竞态检测
go test ./... -race

# 覆盖率
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 添加回归测试

修复 bug 后，在 `tests/regression_test.go` 添加对应的回归测试：

```go
// TestRegression_IssueX_<描述> 测试<简要描述>
//
// 问题描述：<详细描述 bug 的表现>
//
// 复现步骤：
//   <代码示例>
//
// 解决方案：<如何修复的>
//
// 修复日期：YYYY-MM-DD
//
// 相关提交：<commit hash> <commit message>
//
// 测试覆盖：<测试验证了什么>
func TestRegression_IssueX_...(t *testing.T) {
    // 测试代码
}
```

详细说明：[tests/REGRESSION.md](tests/REGRESSION.md)

---

## 发布流程

### 发布前检查

```bash
# 1. 检查 skill 同步
cd cmd/install-skill && go generate

# 2. 运行完整测试
go test ./... -v

# 3. 运行回归测试
go test ./tests/... -run TestRegression

# 4. 检查 git 状态
git status
```

### 版本发布步骤

1. 更新 `CHANGELOG.md`
2. 更新版本号（如有）
3. 创建 git tag
4. 推送到远程

```bash
git tag -a v1.x.x -m "Release v1.x.x"
git push origin v1.x.x
```

---

## 常见问题

### 并发安全

- ✅ 多 goroutine 同时读取：支持
- ⚠️ SetMockData 并发：有锁保护，但最终值不确定

### 错误处理

所有错误定义在 `internal/config/errors.go`，使用 `errors.Join()` 组合多个错误。

### 性能考虑

- 使用 `sync.RWMutex` 保护缓存
- StructMapper 缓存反射结果
- 避免频繁的文件 I/O

---

## 相关文档

| 文档 | 说明 | 读者 |
|------|------|------|
| [README.md](README.md) | 使用说明和示例 | 使用者 |
| [.claude/skills/gameconfig/SKILL.md](.claude/skills/gameconfig/SKILL.md) | AI Skill 使用指南 | 用户 + AI |
| [.claude/skills/gameconfig/abilities/AI指导.md](.claude/skills/gameconfig/abilities/AI指导.md) | AI 能力指导 | AI |
| [DESIGN.md](DESIGN.md) | 设计文档 | 开发者 |
| [CHANGELOG.md](CHANGELOG.md) | 版本变更记录 | 所有人 |
| [AI_NATIVE.md](AI_NATIVE.md) | AI 原生规划 | 开发者 |
| [docs/DEVELOPER.md](docs/DEVELOPER.md) | 详细开发者指南 | 开发者 |
| [tests/REGRESSION.md](tests/REGRESSION.md) | 回归测试索引 | 开发者 |

---

## AI 原生规划

gameconfig 正在演进为 AI 原生开发工具，未来将支持：
- 🎯 AI 主动配置配置表
- 🔍 AI 审查和优化配置
- 🧪 AI 自动生成测试数据
- 📊 AI 生成文档和示例

详细规划：[AI_NATIVE.md](AI_NATIVE.md)
