---
name: Mail-check
description: |
  校验名将杀 Mail.xlsx 邮件配置表。
  当用户提到 Mail.xlsx、邮件配置检查、邮件表校验、信件配置审核时使用此 skill。
  公共流程见 Excel-check；本文件仅含邮件表字段与语义规则。
  结构化规则由脚本校验；Title/Sender 人物匹配由 Agent LLM 分析。
---

# Mail-check — 邮件配置表检查

校验 `Mail.xlsx`。公共流程、依赖、CLI、issue 格式、汇报格式见上级 [Excel-check/SKILL.md](../SKILL.md)。

**运行时输入**：用户提供 `Mail.xlsx` 路径，不读取项目内其他对照文件。

## 脚本

```bash
python Mail-check/scripts/check_mail.py "<Mail.xlsx 路径>"
python Mail-check/scripts/check_mail.py "<路径>" --json
python Mail-check/scripts/check_mail.py "<路径>" --semantic-json
```

人物匹配**禁止**写进脚本对照表，一律阶段 2 LLM。

---

## 结构化规则（脚本）

#### 1. Id — `int`，不得重复

#### 2. Title — `string`，非空

#### 3. Body — 以 `<color=#FFFFFF00>占位</color>` 开头

#### 4. SenderType

| Title 条件 | SenderType |
|-----------|------------|
| 武将来信（不含山河争锋来信） | `2` |
| `熔炼武将道具返还` | `2` |
| 其他 | `1` |

#### 5. Sender（固定值，脚本校验）

| Title 条件 | Sender |
|-----------|--------|
| `举报成功` / `被举报通知` | `系统` |
| `山河争锋守城将来信` | `守城将` |
| `山河争锋攻城将来信` | `攻城将` |
| 其他系统邮件 | `名将杀运营团队` |
| 武将来信 / `熔炼武将道具返还` | 非空即可；**是否与 Title 同一人物 → 语义阶段** |

同一 Title 的多行必须使用相同 Sender。

#### 6. ReceiverType / Receiver

| Receiver | ReceiverType |
|----------|--------------|
| 空 | `1` |
| 等级（≤3 位数字） | `2` |
| 道具 id（≤5 位数字） | `3` |

#### 7. SendCondType

| Title | SendCondType |
|-------|--------------|
| `举报成功` | `4` |
| `被举报通知` | `5` |
| 武将来信 | `3` |
| 其他系统邮件 | `0` |

#### 8. SendCond

- 类型：**string**（Excel 单元格须为文本，不能是数值型）

| SendCondType | SendCond |
|-------------|----------|
| `4` / `5` | 字符串 `1` 或 `2` |
| `3` | 7 位数字字符串，如 `"1010804"` |
| `0` | 空 |

数值 `1010804` 与字符串 `"1010804"` 不等价，前者违规。

#### 9. Item — `{道具id;数量}` 或 `{a;b},{c;d}`

---

## 语义规则（Agent 必须执行）

脚本跑完后读取 `--semantic-json`（或 `--json` 的 `semantic_rows`），对每一行判断。

### 分析对象

| category | 含义 |
|----------|------|
| `warrior_letter` | 武将来信（Title 以「来信」结尾，不含山河争锋来信） |
| `melt_return` | `熔炼武将道具返还` |

每行字段：`id`、`title`、`sender`、`body_preview`、`category`

### 判断标准

1. **Title 与 Sender 是否同一人物**（核心）
   - Title 常见 `{字号/别名}来信`，Sender 为武将全名
   - 核心意思一致即通过，不要求字面相同
   - 例：`汉升来信` + `黄忠` ✓；`刘季来信` + `刘邦` ✓
   - 例：`汉天来信` + `黄忠` ✗（汉天不是黄忠的字/别称）

2. **Body 与 Title/Sender**（可选）
   - Body 明显第一人称且与 Sender 矛盾 → 可报 issue
   - 仅风格问题、无法确定 → 不报

3. **熔炼邮件**
   - Sender 应为合理武将名，且与「熔炼返还」含义不矛盾

字段名常用 `Sender` 或 `Title`；说明中写清人物为何不匹配。

---

## 工作流程

```
用户: 检查一下 Mail.xlsx
→ 按 Excel-check 公共流程
→ python Mail-check/scripts/check_mail.py "<路径>"
→ python Mail-check/scripts/check_mail.py "<路径>" --semantic-json
→ 按本文件语义规则分析 Title/Sender
→ 合并输出完整报告
```
