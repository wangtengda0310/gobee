---
name: check-config-commit-message
description: |
  检查策划配表提交的 commit message 是否与修改内容匹配。
  分析 Excel 文件的实际变更，对比提交描述，识别描述不准确或随意填写的情况。
  当用户需要验证配表提交质量、检查 commit message 是否反映真实修改时使用。
tags: [excel, config, commit-message, quality-check, git]
---

# 配表提交 Commit Message 检查

## 概述

分析策划配表提交的 commit message 是否与 Excel 文件的实际修改内容匹配。识别以下问题：
- 提交描述与修改内容不符
- commit message 过于笼统或随意填写
- 修改了多个系统但描述只提到一个
- 数值调整未在描述中体现

## 触发条件

满足以下任一条件时触发：
- 提示词明确要求检查 commit message 质量
- 提示词提到「验证提交描述」「检查 message 是否准确」等关键词
- 分析配表提交影响时需要评估提交质量

## 工作流步骤

### 步骤 1: 获取提交信息

从环境变量或参数读取：
- `COMMIT_HASH`: 分析的目标 commit
- `CONFIG_REPO`: 配表仓库路径

执行 git 命令获取：
```bash
git log -1 --format='%s' COMMIT_HASH
git log -1 --format='%an' COMMIT_HASH
git diff --stat COMMIT_HASH~1 COMMIT_HASH
```

### 步骤 2: 分析 Excel 文件变更

对每个变更的 Excel 文件，使用 `/excel-parser` skill：
- 读取变更前后的关键字段
- 识别新增/删除行、修改的数值

### 步骤 3: 对比 Commit Message 与实际修改

评估匹配度：
- **完全匹配**: 描述准确反映了所有主要变更
- **部分匹配**: 描述提到了部分变更，但遗漏了重要内容
- **不匹配**: 描述与实际修改明显不符
- **过于笼统**: 描述如"更新配表"、"修改数据"等，没有具体信息

### 步骤 4: 生成检查报告

输出结构化报告，包含：
- Commit Message 匹配度评级
- 具体问题列表（如有）
- git 提交信息原文
- 改进建议（如有）

## 执行工具

- `git log`: 获取提交元信息
- `git diff --stat`: 获取变更统计
- `git diff`: 获取具体差异（必要时）
- `/excel-parser`: 解析 Excel 变更内容

## 输出格式

```markdown
# 配表提交 Commit Message 检查报告

## 匹配度

✅ 完全匹配 / ⚠️ 部分匹配 / ❌ 不匹配 / 📝 过于笼统

## ⚠️发现问题（如果没有则不显示此项）
问题1 . 描述提到"调整千里单骑数值"但未说明具体字段变化
问题2. 新增了 buff 配置但描述未提及

## 原始提交日志

> **Commit**: 张飞技能配置
> **作者**: 张三
> **日期**: 2026-05-21 10:00

## 改进建议（如果不需要则不显示此项）

建议的 commit message:
> 调整千里单骑无尽模式数值：
> - 动作表：降低精英怪触发概率 20%→15%
> - 计算器表：提高层数奖励系数 1.2→1.5
> - buff表：新增暴击buff效果

```

## 注意事项

- 中文文件名需确保 git `core.quotepath=false`
- 大量行变更时只展示关键字段的变化
- 新增/删除整表属于重大变更，必须在描述中体现

## 补充知识
*无尽诸葛|RougeEndlessZhuGeLiangNode* 表的`BuffDesc`字段1和2两个key为策划与客户端约定的特殊值，1代表绿色增益效果2代表红色减益效果

## 自迭代机制

当 skill 遇到无法准确判断的情况时，应通过 multica 创建 issue 驱动 skill 进化，确保规则持续完善。

### 触发场景

以下任一情况发生时，必须创建迭代 issue：

1. **发现新的 commit message 问题模式（当前规则未覆盖）**
   - 例如：commit message 使用了项目内部黑话/缩写，skill 无法识别其含义
   - 例如：commit message 包含特殊标记（如 `[紧急]`、`[热更]`）但 skill 未识别其重要性
   - 例如：commit message 引用了外部系统（如 Jira 工单号、飞书文档链接）但 skill 未验证其有效性

2. **误判：commit message 实际准确但被判定为不匹配**
   - 例如：commit message 使用了行业通用缩写（如 DPS、HP、CD），skill 误报为"描述不清"
   - 例如：commit message 描述了设计意图而非具体数值，skill 误报为"与修改内容不符"
   - 例如：commit message 引用了配置表之间的关联关系，skill 无法验证其正确性

3. **漏判：commit message 有问题但未被识别**
   - 例如：commit message 描述了 A 系统的修改，但实际修改了 B 系统，skill 未检测到
   - 例如：commit message 遗漏了破坏性变更（如删除字段、修改枚举值），skill 未标记为严重问题
   - 例如：commit message 数值描述与实际修改不一致（如说"增加"实际是"减少"），skill 未识别

4. **遇到非标准提交格式（如合并提交、revert 提交）**
   - 例如：合并提交（Merge commit）的 message 被误判为"过于笼统"
   - 例如：revert 提交的 message 格式（`Revert "xxx"`）未被正确识别
   - 例如：squash 合并后的 commit message 包含多个提交描述，skill 无法逐条验证
   - 例如：cherry-pick 提交的 message 包含 `(cherry picked from commit ...)` 标记，skill 未正确处理

### Issue 创建规范

使用 multica 创建 issue，命令格式：

```bash
multica issue create \
  --title "[check-config-commit-message] 问题简述" \
  --description-stdin --project rain-qa-func <<'EOF'
## 问题描述

[清晰描述遇到的问题，属于哪种触发场景]

## 发现位置

- **Commit Hash**: `abc1234`
- **涉及文件**: `config/hero.xlsx`, `config/skill.xlsx`
- **提交作者**: 张三
- **提交时间**: 2026-05-28

## 当前 Skill 判定逻辑

[引用 SKILL.md 中相关的判定规则，说明 skill 当前是如何判断的]

## 验证步骤

1. 进入配表仓库：`cd $CONFIG_REPO`
2. 查看提交信息：`git log -1 --format='%s' abc1234`
3. 查看变更统计：`git diff --stat abc1234~1 abc1234`
4. 解析 Excel 变更：`[具体的 excel-parser 命令或手动检查步骤]`
5. 预期结果：[skill 应该如何判定]
6. 实际结果：[skill 当前的判定结果]

## 建议修复方向

[简要说明如何修改 skill 规则以解决此问题]

## 必须迭代的产物

- [ ] 修改 SKILL.md 的 [具体章节/段落]
- [ ] 新增/更新 [具体规则/示例/注意事项]
EOF
```

### 必须迭代的产物

根据问题类型，明确产出物：

| 问题类型 | 迭代产物 | 修改位置 |
|---------|---------|---------|
| 新的问题模式 | 新增检查规则或识别逻辑 | "概述" 的问题列表 或 "步骤 3" 的评估标准 |
| 误判 | 更新评估标准，增加例外处理 | "步骤 3: 对比 Commit Message 与实际修改" |
| 漏判 | 增强检测逻辑，补充边界 case | "步骤 3" 的匹配度定义 或 "注意事项" |
| 非标准提交 | 新增提交类型处理分支或排除规则 | "工作流步骤" 新增步骤 或 "注意事项" |

### 示例

**场景**：合并提交被误判为"过于笼统"

**问题描述**：
- Commit message: `Merge branch 'feature/hero-rework' into dev`
- Skill 判定：过于笼统（描述如"更新配表"、"修改数据"等，没有具体信息）
- 实际：这是一个标准的 Git 合并提交，message 格式正确，不应被判定为问题

**创建的 issue**：

```bash
multica issue create \
  --title "[check-config-commit-message] 合并提交被误判为过于笼统" \
  --description-stdin --project rain-qa-func <<'EOF'
## 问题描述

合并提交（Merge commit）的 commit message 被 skill 误判为"过于笼统"。

属于触发场景：遇到非标准提交格式。

## 发现位置

- **Commit Hash**: `a1b2c3d`
- **涉及文件**: `config/hero.xlsx`, `config/skill.xlsx`, `config/buff.xlsx`
- **提交作者**: 李四
- **提交时间**: 2026-05-27

## 当前 Skill 判定逻辑

SKILL.md "步骤 3" 中定义：
> - **过于笼统**: 描述如"更新配表"、"修改数据"等，没有具体信息

当前 skill 将 `Merge branch 'feature/hero-rework' into dev` 匹配为"过于笼统"，因为 message 中未提及具体的 Excel 修改内容。

## 验证步骤

1. 进入配表仓库：`cd $CONFIG_REPO`
2. 查看提交信息：`git log -1 --format='%s' a1b2c3d`
   - 输出：`Merge branch 'feature/hero-rework' into dev`
3. 查看变更统计：`git diff --stat a1b2c3d~1 a1b2c3d`
   - 输出：涉及 3 个文件，大量变更
4. 检查提交类型：`git log -1 --format='%p' a1b2c3d`
   - 输出：两个 parent commit，确认是合并提交
5. 预期结果：skill 应识别为合并提交，跳过"过于笼统"检查，或提示"合并提交，建议补充合并说明"
6. 实际结果：skill 判定为"过于笼统"

## 建议修复方向

在"步骤 2: 分析 Excel 文件变更"之前，增加提交类型识别：
- 检测合并提交（两个 parent 或 message 以 `Merge ` 开头）
- 检测 revert 提交（message 以 `Revert ` 开头）
- 对合并提交降低检查严格度，或单独输出"合并提交"评级

## 必须迭代的产物

- [ ] 修改 SKILL.md "工作流步骤"，在"步骤 1"和"步骤 2"之间新增"步骤 1.5: 识别提交类型"
- [ ] 新增提交类型识别逻辑：合并提交、revert 提交、cherry-pick 提交
- [ ] 更新"步骤 3"评估标准，对合并提交增加特殊处理分支
- [ ] 在"注意事项"中补充：合并提交如包含非快进合并，建议补充合并原因说明
EOF
```

**迭代后的 SKILL.md 变更**：

```markdown
### 步骤 1.5: 识别提交类型（新增）

识别特殊提交类型，调整后续检查策略：
- **合并提交**: message 以 `Merge ` 开头 或 有两个 parent commit
  - 评级上限为"部分匹配"（合并提交本身允许不描述具体变更）
  - 如包含冲突解决，建议检查冲突解决是否被描述
- **Revert 提交**: message 以 `Revert ` 开头
  - 验证被 revert 的原始提交是否存在
  - 检查 revert 原因是否说明
- **Cherry-pick 提交**: message 包含 `(cherry picked from commit ...)`
  - 验证原始提交是否存在于其他分支
  - 检查 cherry-pick 原因是否说明
- **标准提交**: 以上都不是，执行完整检查
```