# docker/skills/ 开发指导文档

> 游戏 QA 工具的 AI skill 目录，所有 skill 均运行在 Docker 流水线环境中，支持通过 multica-cli 自迭代进化。

---

## 1. 目录说明

### 1.1 用途

本目录存放 AI agent 在 Docker 流水线中调用的 skill 定义文件。每个 skill 是一个自包含的知识/能力单元，通过 YAML frontmatter 声明元信息，以 Markdown 格式编写核心文档。

Skill 不负责容器生命周期、环境准备、通用通知基础设施。这些由 Docker 公共脚本提供：

- `/app/entrypoint.sh`：容器入口，负责启动 Claude 并调用指定 skill
- `/app/notify.sh`：飞书通知公共脚本，提供 `dm` / `group` / `monitor` 子命令

Skill 中需要发送通知时，直接调用 `/app/notify.sh`（参见 `analyze-prompt/SKILL.md` 示例）。

### 1.2 Skill 命名规范

| 项目                 | 规范              | 示例                                     |
|--------------------|-----------------|----------------------------------------|
| 目录名                | kebab-case，语义清晰 | `excel-parser`、`game-config-relations` |
| SKILL.md 中的 `name` | 与目录名一致          | `name: excel-parser`                   |
| 文件名                | 固定为 `SKILL.md`  | `excel-parser/SKILL.md`                |
| 版本号                | SemVer，重大变更升主版本 | `version: 1.0.0`                       |

### 1.3 文件组织方式

```
docker/skills/
├── CLAUDE.md                          # 本文件：目录级开发指导
├── analyze-prompt/
│   └── SKILL.md                       # 配表提交影响分析（流水线默认 skill）
├── excel-parser/
│   ├── SKILL.md                       # 核心 skill 文档
│   └── references/                    # 渐进式披露：拆分的参考文档
│       ├── nodejs-usage.md
│       ├── git-diff.md
│       └── examples.md
├── game-config-relations/
│   └── SKILL.md
├── multica-cli/
│   └── SKILL.md
└── check-config-commit-message/
    └── SKILL.md
```

**组织原则**：
- `SKILL.md` 聚焦核心功能（触发条件、快速开始、主要功能导航）
- 详细参考文档、示例、子主题放入 `references/` 目录
- 避免单个 SKILL.md 超过 500 行，超过时应拆分

---

## 2. 新 Skill 开发规范

### 2.1 标准 YAML Frontmatter

每个 SKILL.md 文件必须以 YAML frontmatter 开头：

```yaml
---
name: skill-name
description: |
  一句话描述 skill 的核心能力。

  **必须使用此技能的场景**：
  - 场景1的具体描述
  - 场景2的具体描述

  **不要触发**：
  - 不应触发的场景1
  - 不应触发的场景2

  **判定关键**：一句话总结何时应该使用此 skill

  **自迭代触发**：遇到 XX 场景时，调用 `/multica-cli` 创建 issue 驱动 skill 进化。详见下方"自迭代机制"章节。
version: 1.0.0
tags: [tag1, tag2, tag3]
---
```

**字段说明**：

| 字段            | 必填 | 说明                 |
|---------------|----|--------------------|
| `name`        | 是  | skill 标识符，与目录名一致   |
| `description` | 是  | 多行描述，必须包含触发场景和排除场景 |
| `version`     | 是  | SemVer 版本号         |
| `tags`        | 是  | 关键词标签数组，便于分类检索     |

### 2.2 触发条件说明

在 `description` 中明确声明：

1. **必须使用此技能的场景** — 列出所有应该触发此 skill 的条件
2. **不要触发** — 列出易混淆但不应触发此 skill 的场景
3. **判定关键** — 一句话帮助 AI 快速判断是否应使用此 skill

### 2.3 核心功能文档

SKILL.md 正文应包含以下章节：

```markdown
# Skill 标题

> 一句话概述

## 快速开始

### 前置条件

说明运行环境要求、依赖安装等。

## 功能导航

### [功能A](#功能a) — 简要说明
### [功能B](#功能b) — 简要说明

---

## 功能A

### 用法

工具/命令说明。

### 参数

| 参数     | 必填 | 说明 |
|--------|----|----|
| param1 | 是  | 说明 |

### 示例

```json
{ "示例": "数据" }
```

---

## 功能B
...

## 注意事项

- 注意点1
- 注意点2
```

### 2.4 自迭代机制（必须包含）

每个 skill 必须包含自迭代机制章节，说明何时创建 issue 驱动 skill 进化。

详见下方第 3 节的标准模板。

### 2.5 注意事项

- 使用中文编写
- 代码块使用正确的语法高亮（`bash`、`json`、`markdown`、`go` 等）
- 表格优先于段落描述
- 引用其他 skill 使用 `/skill-name` 格式

---

## 3. 自迭代机制标准模板

以下模板可直接复制粘贴到新 skill 的 SKILL.md 中，根据 skill 类型调整触发场景。

### 3.1 触发场景表格

```markdown
## 自迭代机制（Skill 持续进化）

本 skill 通过 Multica issue 驱动自成长。分析/执行过程中遇到以下场景时，**必须**调用 `/multica-cli` 创建 issue，由独立 agent 验证并修复 skill 文档。

### 触发提 issue 的场景

| 场景 | 判定标准 | Issue 类型 |
|------|---------|-----------|
| **知识盲区** | 用户查询 skill 未覆盖的内容或功能 | `knowledge-gap` |
| **文档错误** | 发现 skill 描述与实际行为/代码不符 | `doc-error` |
| **新功能发现** | 发现 skill 应支持但当前未记录的功能 | `new-feature` |
| **规则滞后** | 底层实现已变更但 skill 文档未同步 | `outdated-rule` |
| **示例错误** | 示例代码/命令执行结果与文档描述不一致 | `example-error` |
```

### 3.2 Issue 创建规范

```markdown
### Issue 创建规范

创建 issue 时必须包含以下验证细节，确保接收 agent 无需额外上下文即可执行验证：

```bash
multica issue create --title "[类型] 简要描述" --description-stdin --project rain-qa-func <<'EOF'
## 问题描述
一句话概括发现的问题

## 发现位置
- 分析上下文：用户在执行 XX 操作时提出
- 涉及文件：skill 文件路径（如已知）
- 涉及代码：代码文件路径:行号（如已知）

## 当前 Skill 状态
- Skill 声称：XXX
- 实际观察到的：YYY

## 验证步骤（必须可独立执行）
1. 执行 XXX 操作
2. 观察 YYY 结果
3. 对比文档描述与实际行为

## 建议修复方向
- 如果是知识盲区：建议补充 XXX 功能说明
- 如果是文档错误：建议修正 XXX 描述为 YYY
- 如果是新功能：建议添加 XXX 功能章节

## 必须迭代的产物

**所有 issue 的最终产出必须是修改本 SKILL.md 文件**，确保 skill 真正进化：

| Issue 类型 | SKILL.md 修改位置 | 修改内容 |
|-----------|------------------|---------|
| `knowledge-gap` | 对应功能章节 | 新增未覆盖的功能说明 |
| `doc-error` | 错误描述的所在章节 | 修正错误描述 |
| `new-feature` | 功能导航 + 新增章节 | 添加新功能文档 |
| `outdated-rule` | 规则所在章节 | 更新规则描述 |
| `example-error` | 示例所在章节 | 修正示例代码 |

**禁止行为**：
- 禁止仅验证问题而不修改 SKILL.md
- 禁止将修复内容写在其他文件或注释中
- 禁止关闭 issue 时 SKILL.md 未同步更新
EOF
```
```

### 3.3 Issue 类型标签

```markdown
### Issue 类型标签

创建 issue 时标题前缀使用以下标签：
- `[knowledge-gap]` — 知识盲区，需要调研补充
- `[doc-error]` — 文档描述错误，需要修正
- `[new-feature]` — 新发现的功能需求
- `[outdated-rule]` — 规则滞后于实现
- `[example-error]` — 示例代码错误
```

### 3.4 示例

```markdown
### 示例

**场景**：用户发现 skill 文档中的示例命令执行失败

```bash
multica issue create --title "[example-error] XXX 示例命令参数错误" --description-stdin --project rain-qa-func <<'EOF'
## 问题描述
SKILL.md 中"功能A"章节的示例命令缺少必需参数 `--required-param`，直接复制执行会报错。

## 发现位置
- 分析上下文：用户按文档执行示例命令时失败
- 涉及文件：docker/skills/my-skill/SKILL.md
- 涉及章节：## 功能A → ### 示例

## 当前 Skill 状态
- Skill 声称：示例命令可直接执行
- 实际观察到的：执行后报错 "missing required flag --required-param"

## 验证步骤
1. 复制 SKILL.md 第 N 行的示例命令
2. 在 Docker 环境中执行
3. 观察错误输出

## 建议修复方向
在示例命令中添加 `--required-param` 参数

## 必须迭代的产物
修改 `docker/skills/my-skill/SKILL.md`：
1. 修正"功能A"章节的示例代码
2. 验证修正后的命令可正常执行
EOF
```
```

---

## 4. 与 multica-cli 的集成规范

### 4.1 环境假设

Docker 流水线环境已预置以下配置：

- `/home/analyzer/.multica/config.json` — 包含有效的 PAT，**无需登录**
- `multica` CLI 已安装并在 PATH 中
- `multica issue create` 是纯 API 调用，**无需启动 daemon**

### 4.2 标准调用方式

```bash
# 基础创建（固定 project 为 rain-qa-func）
multica issue create --title "标题" --description "描述" --project rain-qa-func

# 长描述（推荐，避免命令行长度限制）
cat <<'EOF' | multica issue create --title "标题" --description-stdin --project rain-qa-func
描述内容...
EOF
```

### 4.3 Issue 标题前缀规范

所有由 skill 自迭代机制创建的 issue 必须使用以下前缀：

| 前缀                | 含义    | 示例                                     |
|-------------------|-------|----------------------------------------|
| `[knowledge-gap]` | 知识盲区  | `[knowledge-gap] 伙伴系统配表关系缺失`           |
| `[doc-error]`     | 文档错误  | `[doc-error] HeroDropCheck 规则描述错误`     |
| `[new-relation]`  | 新关系发现 | `[new-relation] Activity 与 Mail 的引用关系` |
| `[outdated-rule]` | 规则滞后  | `[outdated-rule] 保护期月数配置已变更`           |
| `[index-error]`   | 索引错误  | `[index-error] SeasonPassId 字段位置错误`    |
| `[new-feature]`   | 新功能   | `[new-feature] 支持 Git 分支对比`            |
| `[example-error]` | 示例错误  | `[example-error] diff 命令示例参数缺失`        |

---

## 5. 最佳实践

### 5.1 渐进式披露原则

参考 `excel-parser` 的 `references/` 目录拆分：

- **SKILL.md** 只保留核心功能导航和快速开始（目标：5 分钟内上手）
- **references/ 子文档** 存放详细指南、完整示例、边缘场景
- 子文档通过相对路径链接：`[详细指南](references/guide.md)`

**拆分信号**：
- SKILL.md 超过 500 行
- 某个功能有 3 个以上子主题
- 示例代码超过 50 行
- 有多个独立的使用场景（如 MCP 方案 vs Node.js 方案）

### 5.2 保持 Skill 聚焦单一职责

- 一个 skill 只做一件事，做好一件事
- 功能有重叠时，通过触发条件明确区分（见 `excel-parser` vs `game-config-relations`）
- 避免"万能 skill"，宁拆勿合

### 5.3 文档与代码同步更新

- 修改底层实现后，必须同步更新 skill 文档
- 新增功能时，先更新 SKILL.md，再提交代码
- 示例命令必须实际可执行，禁止虚构参数

---

## 6. 现有 Skill 索引

| Skill 目录                       | 名称                          | 用途                          | 版本    | 自迭代                         |
|--------------------------------|-----------------------------|-----------------------------|-------|-----------------------------|
| `analyze-prompt/`              | analyze-prompt              | 配表提交影响分析（流水线 `CI_INVOKE_SKILL` 默认 skill，任务上下文由 entrypoint.sh 通过 `CI_*` 环境变量注入） | 1.0.0 | 否（流程编排型）                    |
| `excel-parser/`                | excel-config-guide          | Excel 配表解析、查询、生成、Git 版本对比   | 5.0.0 | 否（工具型）⚠️ 从用户 skill 复制，需注意同步 |
| `game-config-relations/`       | game-config-relations       | 游戏配表跨表关系知识库                 | 1.0.0 | 是                           |
| `multica-cli/`                 | multica-cli                 | Docker 流水线中创建 Multica issue | 1.1.0 | 否（基础工具）                     |
| `check-config-commit-message/` | check-config-commit-message | 检查配表提交 commit message 质量    | -     | 否                           |

**说明**：
- **工具型 skill**（如 multica-cli、excel-parser）：提供基础能力，通常不需要自迭代
- **知识型 skill**（如 game-config-relations）：包含业务知识，必须通过自迭代机制持续进化
- 新增知识型 skill 时，默认启用自迭代机制

⚠️ **外部 Skill 同步提醒**：
- `excel-parser` 是从用户全局 skill 目录复制而来，非本仓库原生维护
- 如用户更新全局 skill，需手动同步到本目录
- 同步时保留自迭代机制章节（如已添加）

---

## 附录：新 Skill 创建检查清单

创建新 skill 时，逐项确认：

- [ ] 目录名符合 kebab-case 规范
- [ ] SKILL.md 包含标准 YAML frontmatter
- [ ] description 中包含触发条件和排除条件
- [ ] 如果是知识型 skill，包含"自迭代机制"章节
- [ ] 自迭代章节包含触发场景表格、Issue 创建规范、类型标签、示例
- [ ] 示例代码/命令已验证可执行
- [ ] 文档超过 500 行时拆分到 references/ 目录
- [ ] 更新本 CLAUDE.md 的"现有 Skill 索引"表格
