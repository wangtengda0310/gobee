---
name: mail-check
description: |
  校验名将杀 Mail.xlsx 邮件配置表。
  当用户提到 Mail.xlsx、邮件配置检查、邮件表校验、信件配置审核时使用此 skill。
  结构化规则由脚本校验；Title/Sender 等语义匹配由 Agent 读表后 LLM 分析。
---

# mail-check — 邮件配置表检查

校验 `Mail.xlsx` 是否符合策划规范。表结构前 5 行为元数据，数据从第 6 行起（Id 列）。

**运行时输入**：用户提供 `Mail.xlsx` 路径，不读取项目内其他对照文件。

## 依赖

- Python 3
- `pandas`、`openpyxl`（读 xlsx）

```bash
pip install pandas openpyxl
```

## 目录结构

```
mail-check/
├── SKILL.md
└── scripts/
    └── check_mail.py
```

## 输入

- **Mail.xlsx 路径**（必填）

## 两阶段校验

| 阶段 | 执行者 | 内容 |
|------|--------|------|
| 1. 结构化 | 脚本 | Id/格式/枚举/重复/固定 Sender 等硬规则 |
| 2. 语义 | Agent（LLM） | Title 与 Sender 是否指向同一人物；核心意思一致即可 |

**禁止**在脚本中维护字号→武将名对照表；人物匹配一律由 LLM 判断。

---

## 阶段 1：运行脚本

```bash
python mail-check/scripts/check_mail.py "<Mail.xlsx 路径>"
python mail-check/scripts/check_mail.py "<路径>" --json
python mail-check/scripts/check_mail.py "<路径>" --semantic-json
```

| 命令 | 用途 |
|------|------|
| 默认 | 输出结构化问题 + 待语义分析行数 |
| `--json` | 完整报告（结构化问题 + semantic_rows） |
| `--semantic-json` | 仅导出待 LLM 分析的行 |

退出码：`0` = 结构化无问题，`1` = 结构化有问题，`2` = 文件不存在。

### 结构化规则（脚本）

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
| 武将来信 / `熔炼武将道具返还` | 非空即可；**是否与 Title 同一人物 → 阶段 2** |

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

## 阶段 2：LLM 语义分析（Agent 必须执行）

脚本跑完后，**必须**读取 `--semantic-json` 输出（或 `--json` 中的 `semantic_rows`），对每一行做语义判断。

### 分析对象

- `category=warrior_letter`：武将来信（Title 以「来信」结尾，不含山河争锋来信）
- `category=melt_return`：`熔炼武将道具返还`

每行字段：`id`、`title`、`sender`、`body_preview`（正文摘要，辅助判断）

### 判断标准

1. **Title 与 Sender 是否同一人物**（核心规则）
   - Title 常见为 `{字号/别名}来信`，Sender 为武将全名
   - 核心意思一致即通过，不要求字面相同
   - 例：`汉升来信` + `黄忠` ✓（汉升是黄忠的字）
   - 例：`刘季来信` + `刘邦` ✓
   - 例：`汉天来信` + `黄忠` ✗（汉天不是黄忠的字/别称，应为汉升一类）
   - 例：表内全部改成 `汉天来信` + `黄忠` 仍应判 ✗

2. **Body 与 Title/Sender**（可选辅助）
   - 若 Body 明显以第一人称叙述且与 Sender 矛盾，可报 issue
   - 仅 Body 风格问题、无法确定时不报

3. **熔炼邮件**
   - Sender 应为合理武将名，且与 Title 含义（熔炼返还）不矛盾

### 输出原则

- **明确不一致**才报 issue；无法确定则标注「需人工确认」，不算违规
- 不要因缺少外部配置表而跳过语义分析
- 批量分析时可分批（每批 50 行），合并结果

### 语义 issue 格式

与结构化一致：

```
Id=<Id> | Title=<Title> | <字段名> | <问题说明>
```

字段名常用 `Sender` 或 `Title`；说明中写清人物为何不匹配。

---

## 最终汇报

合并两阶段结果：

1. 检查文件路径
2. 结构化问题数 + 明细
3. 语义问题数 + 明细
4. 总问题数；若无任何问题，说明「结构化与语义校验均已通过」

## 工作流程

```
用户: 检查一下 Mail.xlsx
→ python mail-check/scripts/check_mail.py "<路径>"
→ python mail-check/scripts/check_mail.py "<路径>" --semantic-json
→ Agent 对 semantic_rows 逐批 LLM 分析 Title/Sender 人物是否一致
→ 合并输出完整报告
```
