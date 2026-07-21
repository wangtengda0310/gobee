---
name: SkillLines_技能台词表
description: |
  校验名将杀 SkillLines_技能台词表.xlsx。
  当用户提到 SkillLines_技能台词表.xlsx、SkillLines_技能台词表配置检查时使用。
  公共流程见 Excel-check/SKILL.md；本文件仅含本表规则。
---

# SkillLines_技能台词表 — 配置表检查

校验 `SkillLines_技能台词表.xlsx`（sheet：`技能台词配置表|SkillLines`）。公共约定见 [Excel-check/SKILL.md](../SKILL.md)。

**表结构**：A（经典行表）
**主键列**：`Id`
**运行时输入**：用户提供 `SkillLines_技能台词表.xlsx` 路径。本表会读取同目录下：
- `HeroSkinItem_英雄皮肤.xlsx`（或 `--hero-skin-item`）：有值 `SkinId` 外联
- `HeroLines_武将台词表.xlsx`（或 `--hero-lines`）：`Skill*Line` 台词 Id 外联
- `Skill.xlsx`（或 `--skill`）：`TabName`→技能拼音 `E#ESkillId`，与本行 `SkillId` 对照

**业务标识列**：`SkillFirstLine`。

**读表补充**：类型行=1，字段名行=2，数据起始行（0-based）=`4`；连续空 3 行截断；`#` 分区行跳过；第 2、3 行皆空列丢弃。

## 脚本

```bash
python "SkillLines_技能台词表/scripts/check_SkillLines_jinengtaicibiao.py" "<路径>"
python "SkillLines_技能台词表/scripts/check_SkillLines_jinengtaicibiao.py" "<路径>" --hero-skin-item "<HeroSkinItem路径>" --hero-lines "<HeroLines路径>" --skill "<Skill.xlsx路径>"
python "SkillLines_技能台词表/scripts/check_SkillLines_jinengtaicibiao.py" "<路径>" --json
python "SkillLines_技能台词表/scripts/check_SkillLines_jinengtaicibiao.py" "<路径>" --semantic-json
```

Issue：`Id=<Id> | SkillFirstLine=… | <字段> | <说明>`

Agent 汇报：原样列出每条 Issue 行；禁止用分类汇总表代替。

---

## 通用规则

### 结构化规则（脚本）

| 编号 | 适用? | 落到本表 |
|------|-------|----------|
| S1 | 是 | `Id` |
| S12 | 是 | 有值字段按类型行：`SkinId`→int，`SkillId`→ESkillId，`Skill*Line`→int[]，`SpecialAudio`→EAudioId 等 |
| S2 | 部分 | `SkillId` 不可空；`Name`/`Title`/`SkillName` 若存在则不可空 |
| S9 | 部分 | 复杂格式串（本表主要由 S12 覆盖数组/枚举） |
| S11 | 部分 | 成对时间字段（若存在） |
| L1 | 否 | `Skill*Line` 类型为 `int[]`（音效 ID），非自然语言文案 |

主要字段：`Id`, `SkillId`, `SkinId`, `SkillFirstLine`, `SkillSecondLine`, `SkillThirdLine`, `SkillForthLine`, `SpecialAudio`

#### 字段细则

- **Id**：int，不重复（S1）
- **SkinId**：可空；有值须为 int（S12）；有值须落在 `HeroSkinItem_英雄皮肤.xlsx` 的 `SkinItemId`（独有外联）
- **SkillId**：不可空（S2）；须为 ESkillId（标识串或 int）（S12）
- **SkillFirstLine / SkillSecondLine / SkillThirdLine / SkillForthLine**：可空；有值须为 int 或逗号分隔 int 列表（S12）；四列台词 Id 全表不重复，且外联 HeroLines（见独有）
- **SpecialAudio**：可空；有值须为 EAudioId（S12）
- 若存在 `Name` / `Title` / `SkillName`：非空（S2）

### 语义规则

- **L1**：本表台词列均为音效 ID（`int[]`），**不适用**文案质量审。

---

## 独有规则

### 结构化规则（脚本）

1. **SkinId → HeroSkinItem.SkinItemId**  
   `SkinId` 空：不报。有值且已通过 int 形态后：必须等于 `HeroSkinItem_英雄皮肤.xlsx`（sheet：`英雄皮肤|HeroSkinItem`）某行的 `SkinItemId`。  
   缺失外联表：报一次「缺少 HeroSkinItem…无法校验 SkinId 外联」，并跳过逐行外联。  
   找不到：`SkinId=<值> 在 HeroSkinItem.SkinItemId 中不存在`。

2. **Skill*Line 台词 Id 全表唯一**  
   `SkillFirstLine` / `SkillSecondLine` / `SkillThirdLine` / `SkillForthLine` 中所有解析出的台词 Id（含逗号列表各项）在本表**四列合计**不得重复。  
   重复：`台词Id=<id> 在 Skill*Line 中重复出现 n 次`。

3. **Skill*Line → HeroLines.Id + TabName 拼音与 SkillId**  
   每个有值台词 Id：  
   - 必须存在于 `HeroLines_武将台词表.xlsx`（sheet：`武将台词|HeroLines`）的 `Id`；找不到：`台词Id=<id> 在 HeroLines.Id 中不存在`  
   - 取该行 `TabName`，在同目录（或 `--skill`）`Skill.xlsx` 的 `SkillName` 中查找，取其 **`E#ESkillId`（技能拼音 ID）**；须与本行 `SkillId` 一致（同名多行则 `SkillId` 落在任一即可）  
   - `TabName` 在 Skill 找不到：`TabName=<名> 在 Skill.xlsx 找不到，无法校验与 SkillId 一致`  
   - 拼音不一致：`台词Id=<id> TabName=<名> 对应拼音=<E#ESkillId> 与本行 SkillId=<…> 不一致`  
   缺 HeroLines / Skill：各报一次并跳过本条对应外联。本行 `SkillId` 为空时不做拼音对照（已由 S2 报）。

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
用户: 检查 SkillLines_技能台词表.xlsx
→ python "SkillLines_技能台词表/scripts/check_SkillLines_jinengtaicibiao.py" "<路径>" [--hero-skin-item ...] [--hero-lines ...] [--skill ...]
→ 需要时 --semantic-json，按 L1（及独有语义）由 Agent 审
→ 合并结构化 + 语义报告
```

<!-- sync: S12；SkinId→HeroSkinItem；Skill*Line 唯一+HeroLines+Skill拼音→SkillId -->
