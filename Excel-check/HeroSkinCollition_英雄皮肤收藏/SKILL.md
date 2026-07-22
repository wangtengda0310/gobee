---
name: HeroSkinCollition_英雄皮肤收藏
description: |
  校验名将杀 HeroSkinCollition_英雄皮肤收藏.xlsx。
  当用户提到 HeroSkinCollition_英雄皮肤收藏.xlsx、HeroSkinCollition_英雄皮肤收藏配置检查时使用。
  公共流程见 Excel-check/SKILL.md；本文件仅含本表规则。
---

# HeroSkinCollition_英雄皮肤收藏 — 配置表检查

校验 `HeroSkinCollition_英雄皮肤收藏.xlsx`（sheet：`英雄皮肤收藏|HeroSkinCollition`）。公共约定见 [Excel-check/SKILL.md](../SKILL.md)。

**表结构**：A（经典行表）
**主键列**：`Type`
**运行时输入**：用户提供 `HeroSkinCollition_英雄皮肤收藏.xlsx` 路径。默认不读外联表。

**业务标识列**：`Name`。

**读表补充**：类型行=1，字段名行=2，数据起始行（0-based）=`4`；连续空 3 行截断；`#` 分区行跳过；第 2、3 行皆空列丢弃。

## 脚本

```bash
python "HeroSkinCollition_英雄皮肤收藏/scripts/check_HeroSkinCollition_yingxiongpifushoucang.py" "<路径>"
python "HeroSkinCollition_英雄皮肤收藏/scripts/check_HeroSkinCollition_yingxiongpifushoucang.py" "<路径>" --json
python "HeroSkinCollition_英雄皮肤收藏/scripts/check_HeroSkinCollition_yingxiongpifushoucang.py" "<路径>" --semantic-json
```

Issue：`Type=<Type> | Name=… | <字段> | <说明>`

Agent 汇报：原样列出每条 Issue 行；禁止用分类汇总表代替。

---

## 通用规则

### 结构化规则（脚本）

| 编号 | 适用? | 落到本表 |
|------|-------|----------|
| S1 | 是 | `Type` |
| S12 | 是 | 有值字段按类型行（int/bool/string/E*/数组等） |
| S2 | 部分 | 见字段细则 |
| S9 | 部分 | 类型行为数组等时 |
| S11 | 部分 | 成对时间字段（若存在） |
| L1 | 是 | `Name` / `NameImg` / `NameBg` / `Desc` / `OpenDate` |

主要字段：`Type`, `Name`, `NameImg`, `NameBg`, `Desc`, `Weight`, `OpenDate`

#### 字段细则

- **Type**：非空且全表唯一（S1）
- 除主键外其余字段默认可空；有值时按类型行做格式校验（见脚本）
- 若存在 `Name` / `Title` / `SkillName`：非空（S2）

### 语义规则

- **L1 文案质量**：对非空 `Name`、`NameImg`、`NameBg`、`Desc`、`OpenDate` 检查错字、漏字、病句、标点不规范（剥富文本标签后审）。
- 检查时用 `--semantic-json` 取待审行，由 Agent（LLM）执行；脚本不维护错别字词表。

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
用户: 检查 HeroSkinCollition_英雄皮肤收藏.xlsx
→ python "HeroSkinCollition_英雄皮肤收藏/scripts/check_HeroSkinCollition_yingxiongpifushoucang.py" "<路径>"
→ 需要时 --semantic-json，按 L1（及独有语义）由 Agent 审
→ 合并结构化 + 语义报告
```

<!-- sync: 须含 L1：Name, NameImg, NameBg, Desc, OpenDate -->
