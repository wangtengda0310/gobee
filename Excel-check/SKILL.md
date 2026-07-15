---
name: excel-check
description: |
  名将杀 Excel 配置表检查 skill 的公共框架与写作规范。
  当用户要新增某张表的检查 skill、抽取/统一表校验流程、或询问 Excel-check 怎么写时使用。
  具体某张表（如 Mail.xlsx）的字段规则不写在这里，写在对应子目录 skill（如 Mail-check）中。
---

# Excel-check — 配置表检查框架

名将杀配置表检查采用「公共框架 + 单表 skill」结构：

| 层级 | 路径 | 职责 |
|------|------|------|
| 公共 | `Excel-check/SKILL.md`（本文件） | 流程、脚本约定、汇报格式、如何写新表 skill |
| 单表 | `Excel-check/<Table>-check/SKILL.md` | **仅**该表的字段规则、枚举、语义判断标准 |
| 脚本 | `Excel-check/<Table>-check/scripts/check_*.py` | 该表的结构化硬规则实现 |

**原则**：能脚本化的进脚本；需要语境/人物/语义判断的交给 Agent（LLM），禁止在脚本里维护易变对照表。

---

## 目录约定

```
Excel-check/
├── SKILL.md                 # 本文件：公共框架
└── <Table>-check/
    ├── SKILL.md             # 仅表内检查规则
    └── scripts/
        └── check_<Table>.py
```

命名：表文件 `Foo.xlsx` → skill 名 `Foo-check`（与表名大小写一致，后缀 `-check`）。

---

## 表结构约定（名将杀配置表）

除非单表 skill 另行说明：

- 前 5 行为元数据（类型/字段名等）
- **数据从第 6 行起**（脚本中常：`HEADER_ROW` / `DATA_START_ROW` 按实际 sheet 调整）
- 主键列一般为 `Id`

---

## 运行时输入

- **xlsx 路径由用户提供**（必填）
- 默认**不**读取项目内其他对照文件，除非该表 skill 明确要求外联表

---

## 依赖（所有表检查共用）

- Python 3
- `pandas`、`openpyxl`

```bash
pip install pandas openpyxl
```

---

## 两阶段校验（默认模式）

| 阶段 | 执行者 | 内容 |
|------|--------|------|
| 1. 结构化 | 脚本 | 类型、必填、枚举、格式、重复、固定映射等硬规则 |
| 2. 语义 | Agent（LLM） | 语境一致、人物/名称匹配、文案是否矛盾等软规则 |

无语义需求的表可省略阶段 2，在单表 skill 中标明「仅结构化」。

### 职责边界

| 放脚本 | 放 LLM（单表 skill 写清标准） |
|--------|------------------------------|
| 类型、非空、枚举、正则格式 | 字号/别名 ↔ 全名是否同一人物 |
| Id 唯一、跨行一致性（同 Title 同 Sender） | 正文与身份是否矛盾 |
| 固定 Title → 固定字段值 | 策划文案「意思对不对」 |

**禁止**在脚本中维护字号→武将名等易变对照表。

---

## 脚本 CLI 约定

统一参数与退出码，便于 Agent 编排：

```bash
python <Table>-check/scripts/check_<Table>.py "<xlsx 路径>"
python <Table>-check/scripts/check_<Table>.py "<路径>" --json
python <Table>-check/scripts/check_<Table>.py "<路径>" --semantic-json
```

| 命令 | 用途 |
|------|------|
| 默认 | 输出结构化问题 +（若有）待语义分析行数 |
| `--json` | 完整报告 JSON（结构化问题 + semantic_rows） |
| `--semantic-json` | 仅导出待 LLM 分析的行 |

| 退出码 | 含义 |
|--------|------|
| `0` | 结构化无问题 |
| `1` | 结构化有问题 |
| `2` | 文件不存在或其他致命错误 |

`--json` / `--semantic-json` 用于向 Agent 喂数据时，退出码可为 `0`（以 JSON 内容为准）。

### Issue 行格式（结构化与语义统一）

```
Id=<Id> | Title=<Title> | <字段名> | <问题说明>
```

- 无 Title 的表可用其他主键展示列（在单表 skill 中约定），或写 `-`
- `Id` 缺失时用 `-`

### 语义导出行建议字段

至少包含：行标识（`id`）、展示用文案列、待判断字段、可选 `body_preview` / `category`。具体以单表 skill 为准。

---

## Agent 执行流程（检查一张表时）

1. **读该表 skill**：`Excel-check/<Table>-check/SKILL.md`（规则以它为准）
2. **跑结构化脚本**（默认 + 需要时 `--semantic-json`）
3. **若有语义阶段**：按单表 skill 标准分批分析（建议每批 ≤50 行），合并结果
4. **最终汇报**（合并两阶段）：
   1. 检查文件路径
   2. 结构化问题数 + 明细
   3. 语义问题数 + 明细（若有）
   4. 总问题数；全通过时写明「结构化与语义校验均已通过」（仅结构化则写「结构化校验已通过」）

### 语义阶段通用原则

- **明确不一致**才报 issue；吃不准则「需人工确认」，不算违规
- 不因缺少外部对照表而跳过语义分析
- issue 格式与结构化一致

---

## 如何写一张新表的检查 skill

复制并填写以下骨架到 `Excel-check/<Table>-check/SKILL.md`（`<Table>` 与 xlsx 文件名大小写一致）：

```markdown
---
name: <Table>-check
description: |
  校验名将杀 <Table>.xlsx。
  当用户提到 <Table>.xlsx、<中文名>配置检查、<中文名>表校验时使用。
  公共流程见 Excel-check/SKILL.md；本文件仅含本表规则。
---

# <Table>-check — <中文名>配置表检查

校验 `<Table>.xlsx`。公共流程、依赖、CLI、汇报格式见上级 [Excel-check/SKILL.md](../SKILL.md)。

**运行时输入**：用户提供 xlsx 路径。

## 脚本

\`\`\`bash
python <Table>-check/scripts/check_<Table>.py "<路径>"
python <Table>-check/scripts/check_<Table>.py "<路径>" --semantic-json
\`\`\`

## 结构化规则（脚本）

#### 1. 字段A — 类型/约束
...

## 语义规则（若有）

### 分析对象 / category
### 判断标准
### 本表特例

## 工作流程

\`\`\`
用户: 检查 <Table>.xlsx
→ 按 Excel-check 公共流程跑脚本 +（可选）LLM
→ 按本文件规则产出报告
\`\`\`
```

### 单表 skill 应写 / 不应写

| 应写 | 不应写（已在公共 skill） |
|------|-------------------------|
| 字段级规则、枚举表、固定 Title 映射 | `pip install`、通用退出码说明长文 |
| 本表语义 category 与判断范例 | 重复的「两阶段是什么」教科书 |
| 脚本相对路径与本表特有参数 | 与本表无关的其他表规则 |
| 相对公共流程的差异（如「仅结构化」） | |

脚本实现放在同目录 `scripts/`，规则条文与脚本行为保持一致。

---

## 现有子 skill

| Skill | 表 | 说明 |
|-------|-----|------|
| [Mail-check](Mail-check/SKILL.md) | `Mail.xlsx` | 邮件配置；结构化 + Title/Sender 人物语义 |
