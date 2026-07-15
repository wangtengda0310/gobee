---
name: Mail
description: |
  校验名将杀 Mail.xlsx 邮件配置表。
  当用户提到 Mail.xlsx、邮件配置检查、邮件表校验、信件配置审核时使用此 skill。
  公共流程见 Excel-check/SKILL.md；本文件仅含本表规则。
  结构化由脚本校验；Title/Sender 人物匹配为独有语义规则。
---

# Mail — 邮件配置表检查

校验 `Mail.xlsx`。公共约定见 [Excel-check/SKILL.md](../SKILL.md)。

**运行时输入**：用户提供 `Mail.xlsx` 路径，不读取项目内其他对照文件。

**业务标识列**：`Title`（邮件种类/人物来信标题）。

## 脚本

```bash
python Mail/scripts/check_Mail.py "<Mail.xlsx 路径>"
python Mail/scripts/check_Mail.py "<路径>" --json
python Mail/scripts/check_Mail.py "<路径>" --semantic-json
```

人物匹配**禁止**写进脚本对照表，一律独有语义阶段。

Issue 展示列用 `Title`：`Id=<Id> | Title=<Title> | <字段> | <说明>`

---

## 通用规则

### 结构化规则（脚本）

适用公共类型（落到本表）：

| 编号 | 适用? | 落到本表 |
|------|-------|----------|
| S1 | 是 | `Id` |
| S2 | 是 | `Title`；武将来信类 `Sender` 非空 |
| S3 | 是 | `Body` 固定前缀 |
| S4 | 是 | 见独有：`Title`→`SenderType` / `SendCondType` 等（条件枚举形态） |
| S5 | 是 | 见独有：`Title`→固定 `Sender` |
| S6 | 是 | 同 `Title` → 同 `Sender` |
| S7 | 是 | `Receiver`↔`ReceiverType`；`SendCondType`↔`SendCond` |
| S8 | 是 | `SendCond` 须为文本型单元格 |
| S9 | 是 | `Item` |
| S10 | 部分 | `SendCond` 长度/位数（随类型） |
| L1 | 是 | 见独有语义（Title↔Sender 人物） |
| L2–L3 | 部分/是 | 见独有语义 |

读表约定：连续空 3 行截断；第 2、3 行皆空列丢弃。

#### 字段细则（通用落地）

- **Id**：int，不重复  
- **Title**：string，非空  
- **Body**：以 `<color=#FFFFFF00>占位</color>` 开头  
- **Receiver / ReceiverType**：空→1；等级（≤3 位数字）→2；道具 id（≤5 位）→3  
- **SendCond**：单元格须为 string（不能是数值型）；具体取值见独有  
- **Item**：`{道具id;数量}` 或 `{a;b},{c;d}`  
- **同 Title 多行**：`Sender` 必须相同  

### 语义规则

无单独的「通用语义」条文；人物匹配等见下方独有语义（对应 L1 等落到本表）。

---

## 独有规则

### 结构化规则（脚本）

业务钥匙：`Title`。

#### SenderType

| Title 条件 | SenderType |
|-----------|------------|
| 武将来信（不含山河争锋来信） | `2` |
| `熔炼武将道具返还` | `2` |
| 其他 | `1` |

#### Sender（固定值）

| Title 条件 | Sender |
|-----------|--------|
| `举报成功` / `被举报通知` | `系统` |
| `山河争锋守城将来信` | `守城将` |
| `山河争锋攻城将来信` | `攻城将` |
| 其他系统邮件 | `名将杀运营团队` |
| 武将来信 / `熔炼武将道具返还` | 非空即可；是否与 Title 同一人物 → 独有语义 |

#### SendCondType

| Title | SendCondType |
|-------|--------------|
| `举报成功` | `4` |
| `被举报通知` | `5` |
| 武将来信 | `3` |
| 其他系统邮件 | `0` |

#### SendCond（随 SendCondType）

| SendCondType | SendCond |
|-------------|----------|
| `4` / `5` | 字符串 `1` 或 `2` |
| `3` | 7 位数字字符串，如 `"1010804"` |
| `0` | 空 |

数值 `1010804` 与字符串 `"1010804"` 不等价，前者违规。

### 语义规则（Agent 必须执行）

脚本跑完后读取 `--semantic-json`（或 `--json` 的 `semantic_rows`），对每一行判断。

#### 分析对象

| category | 含义 |
|----------|------|
| `warrior_letter` | 武将来信（Title 以「来信」结尾，不含山河争锋来信） |
| `melt_return` | `熔炼武将道具返还` |

每行字段：`id`、`title`、`sender`、`body_preview`、`category`

#### 判断标准

1. **Title 与 Sender 是否同一人物**（核心）  
   - Title 常见 `{字号/别名}来信`，Sender 为武将全名  
   - 核心意思一致即通过  
   - 例：`汉升来信`+`黄忠` ✓；`汉天来信`+`黄忠` ✗  

2. **Body 与 Title/Sender**（可选）  
   - 正文明显第一人称且与 Sender 矛盾 → 可报  
   - 无法确定 → 不报  

3. **熔炼邮件**  
   - Sender 应为合理武将名，且与「熔炼返还」不矛盾  

---

## 工作流程

```
用户: 检查 Mail.xlsx
→ python Mail/scripts/check_Mail.py "<路径>"
→ python Mail/scripts/check_Mail.py "<路径>" --semantic-json
→ 按独有语义规则分析 Title/Sender
→ 合并通用/独有结构化 + 独有语义报告
```
