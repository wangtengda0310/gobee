---
name: GuildCityCreate
description: |
  校验名将杀 GuildCityCreate.xlsx。
  当用户提到 GuildCityCreate.xlsx、GuildCityCreate配置检查时使用。
  公共流程见 Excel-check/SKILL.md；本文件仅含本表规则。
---

# GuildCityCreate — 配置表检查

校验 `GuildCityCreate.xlsx`（sheet：`创建公会城池消耗表|GuildCityCreate`）。公共约定见 [Excel-check/SKILL.md](../SKILL.md)。

**表结构**：A（经典行表）
**主键列**：`Count`
**运行时输入**：用户提供 `GuildCityCreate.xlsx` 路径。默认不读外联表。

**业务标识列**：`Count`。

**读表补充**：类型行=1，字段名行=2，数据起始行（0-based）=`4`；连续空 3 行截断；`#` 分区行跳过；第 2、3 行皆空列丢弃。

## 脚本

```bash
python "GuildCityCreate/scripts/check_GuildCityCreate.py" "<路径>"
python "GuildCityCreate/scripts/check_GuildCityCreate.py" "<路径>" --json
python "GuildCityCreate/scripts/check_GuildCityCreate.py" "<路径>" --semantic-json
```

Issue：`Count=<Count> | Count=… | <字段> | <说明>`

Agent 汇报：原样列出每条 Issue 行；禁止用分类汇总表代替。

---

## 通用规则

### 结构化规则（脚本）

| 编号 | 适用? | 落到本表 |
|------|-------|----------|
| S1 | 是 | `Count` |
| S12 | 是 | 有值字段按类型行（int/bool/string/E*/数组等） |
| S2 | 部分 | 见字段细则 |
| S9 | 部分 | 类型行为数组等时 |
| S11 | 部分 | 成对时间字段（若存在） |
| L1 | 否 | 无文本描述字段 |

主要字段：`Count`, `RedJade`

#### 字段细则

- **Count**：int，不重复（S1）
- 除主键外其余字段默认可空；有值时按类型行做格式校验（见脚本）
- 若存在 `Name` / `Title` / `SkillName`：非空（S2）

### 语义规则

- 无文本描述字段，**L1 不适用**。

---

## 独有规则

### 结构化规则（脚本）

经归纳无（首轮对照 Mail/Recharge 等写法后，本表样例未落成可执行独有硬规则；后续可增补）。

### 语义规则

经归纳无（除公共 L1 外无额外玩法语义；后续可增补）。

---

## 补充规则时（必须）

按用户要求新增/修改规则前：对照本文件已有通用+独有及脚本。

| 情况 | 处理 |
|------|------|
| 实质重复 | 先反馈，勿落盘；问保留/合并/取消 |
| 冲突 | 先列双方，停止实现；问以谁为准 |
| 无重复无冲突 | 再写入，并视需要改脚本 |

细则见 [Excel-check/SKILL.md](../SKILL.md)「使用者后续补充」。

---

## 工作流程

```
用户: 检查 GuildCityCreate.xlsx
→ python "GuildCityCreate/scripts/check_GuildCityCreate.py" "<路径>"
→ 需要时 --semantic-json，按 L1（及独有语义）由 Agent 审
→ 合并结构化 + 语义报告
```

<!-- sync: 无文本描述列，L1 不适用 -->
