---
name: ItemFrame_边框道具表
description: |
  校验名将杀 ItemFrame_边框道具表.xlsx。
  当用户提到 ItemFrame_边框道具表.xlsx、ItemFrame_边框道具表配置检查时使用。
  公共流程见 Excel-check/SKILL.md；本文件仅含本表规则。
---

# ItemFrame_边框道具表 — 配置表检查

校验 `ItemFrame_边框道具表.xlsx`（sheet：`边框表|FrameItem`）。公共约定见 [Excel-check/SKILL.md](../SKILL.md)。

**表结构**：A（经典行表）
**主键列**：`Id`
**运行时输入**：用户提供 `ItemFrame_边框道具表.xlsx` 路径。本表会读取同目录（或 `--item`）下的 `Item.xlsx`，校验本表 `Id` / `Name` 与道具表外联。

**业务标识列**：`Name`。

**读表补充**：类型行=1，字段名行=2，数据起始行（0-based）=`4`；连续空 3 行截断；`#` 分区行跳过；第 2、3 行皆空列丢弃。

## 脚本

```bash
python "ItemFrame_边框道具表/scripts/check_ItemFrame_biankuangdaojubiao.py" "<路径>"
python "ItemFrame_边框道具表/scripts/check_ItemFrame_biankuangdaojubiao.py" "<路径>" --item "<Item.xlsx路径>"
python "ItemFrame_边框道具表/scripts/check_ItemFrame_biankuangdaojubiao.py" "<路径>" --json
python "ItemFrame_边框道具表/scripts/check_ItemFrame_biankuangdaojubiao.py" "<路径>" --semantic-json
```

Issue：`Id=<Id> | Name=… | <字段> | <说明>`

Agent 汇报：原样列出脚本输出的每条 Issue 行；禁止用分类汇总表代替。

---

## 通用规则

### 结构化规则（脚本）

| 编号 | 适用? | 落到本表 |
|------|-------|----------|
| S1 | 是 | `Id` |
| S12 | 是 | 有值字段按类型行（int/bool/string/E*/数组等） |
| S2 | 部分 | 见字段细则 |
| S9 | 部分 | 类型行为数组等时 |
| S11 | 部分 | 成对时间字段（若存在） |
| L1 | 是 | `Name` / `HeadFrameName` / `HeadFrameName1` |

主要字段：`Id`, `Name`, `ImagePath`, `ImagePath1`, `HeadFrameName`, `HeadFrameName1`, `FramePrefab`, `HeadPrefab`

#### 字段细则

- **Id**：int，不重复（S1）
- 除主键外其余字段默认可空；有值时按类型行做格式校验（见脚本）
- 若存在 `Name` / `Title` / `SkillName`：非空（S2）

### 语义规则

- **L1 文案质量**：对非空 `Name`、`HeadFrameName`、`HeadFrameName1` 检查错字、漏字、病句、标点不规范（剥富文本标签后审）。
- 检查时用 `--semantic-json` 取待审行，由 Agent（LLM）执行；脚本不维护错别字词表。

---

## 独有规则

### 结构化规则（脚本）

#### 1. Id / Name ↔ Item.xlsx

本表每一行：

1. `Id` 必须出现在同目录 `Item.xlsx`（sheet：`道具表|Item`）的 `Id` 列中。
2. 本行 `Name` 必须与 Item 中该 `Id` 同行的 `Name` **完全一致**（去首尾空白后比较）。

可用 `--item` 覆盖 Item 路径；道具表缺失则报错并跳过本条逐行外联校验。

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
用户: 检查 ItemFrame_边框道具表.xlsx
→ python "ItemFrame_边框道具表/scripts/check_ItemFrame_biankuangdaojubiao.py" "<路径>" [--item "<Item.xlsx>"]
→ 需要时 --semantic-json，按 L1（及独有语义）由 Agent 审
→ 合并结构化 + 语义报告
```

<!-- sync: 须含 L1：Name, HeadFrameName, HeadFrameName1 -->
