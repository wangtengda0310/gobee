---
name: HeroSkinItem_英雄皮肤
description: |
  校验名将杀 HeroSkinItem_英雄皮肤.xlsx。
  当用户提到 HeroSkinItem_英雄皮肤.xlsx、HeroSkinItem_英雄皮肤配置检查时使用。
  公共流程见 Excel-check/SKILL.md；本文件仅含本表规则。
---

# HeroSkinItem_英雄皮肤 — 配置表检查

校验 `HeroSkinItem_英雄皮肤.xlsx`（sheet：`英雄皮肤|HeroSkinItem`）。公共约定见 [Excel-check/SKILL.md](../SKILL.md)。

**表结构**：A（经典行表）
**主键列**：`SkinItemId`
**运行时输入**：用户提供 `HeroSkinItem_英雄皮肤.xlsx` 路径。本表会读取同目录（或 `--hero`）下的 `Hero.xlsx`，校验 `HeroId` / `SkinPinYin` 外联，以及上线武将四套皮肤齐全。

**业务标识列**：`Name`。

**读表补充**：类型行=1，字段名行=2，数据起始行（0-based）=`4`；连续空 3 行截断；`#` 分区行跳过；第 2、3 行皆空列丢弃。

## 脚本

```bash
python "HeroSkinItem_英雄皮肤/scripts/check_HeroSkinItem_yingxiongpifu.py" "<路径>"
python "HeroSkinItem_英雄皮肤/scripts/check_HeroSkinItem_yingxiongpifu.py" "<路径>" --hero "<Hero.xlsx路径>"
python "HeroSkinItem_英雄皮肤/scripts/check_HeroSkinItem_yingxiongpifu.py" "<路径>" --json
python "HeroSkinItem_英雄皮肤/scripts/check_HeroSkinItem_yingxiongpifu.py" "<路径>" --semantic-json
```

Issue：`SkinItemId=<SkinItemId> | Name=… | <字段> | <说明>`

Agent 汇报：原样列出每条 Issue 行；禁止用分类汇总表代替。

---

## 通用规则

### 结构化规则（脚本）

| 编号 | 适用? | 落到本表 |
|------|-------|----------|
| S1 | 是 | `SkinItemId` |
| S12 | 是 | 有值字段按类型行（int/bool/string/E*/数组等） |
| S2 | 部分 | `HeroId`/`SkinPinYin` 不可空；`Name`/`Title`/`SkillName` 若存在则不可空 |
| S9 | 部分 | 类型行为数组等时 |
| S11 | 部分 | 成对时间字段（若存在） |
| L1 | 是 | `Name` / `GetWay` / `Lines` / `DebutLines` / `KillLines` / `DeadLines` / `LinesDubbed` / `OriginalArtDesigner` / `BodyOffset` |

主要字段：`SkinItemId`, `HeroId`, `RailyType`, `SkinType`, `Name`, `GetWay`, `SkinPinYin`, `SeatSpecialImg`, `HeroUIExtraIcons`, `Lines`, `DebutLines`, `KillLines`, `DeadLines`, `HeroAudio`, `LinesDubbed`, `OriginalArtDesigner`, `CollitionType`, `CanSetBackground`, `IsOpen`, `OpenDate` 等共 29 列

#### 字段细则

- **SkinItemId**：int，不重复（S1）
- **HeroId**：不可空；须落在 `Hero.xlsx` 的 `Id`（独有外联）
- **SkinPinYin**：不可空；须与同行 `HeroId` 对应武将的 `E#EHeroId` 对齐（独有）
- 除上述外其余字段默认可空；有值时按类型行做格式校验（见脚本）
- 若存在 `Name` / `Title` / `SkillName`：非空（S2）

### 语义规则

- **L1 文案质量**：对非空 `Name`、`GetWay`、`Lines`、`DebutLines`、`KillLines`、`DeadLines`、`LinesDubbed`、`OriginalArtDesigner`、`BodyOffset` 检查错字、漏字、病句、标点不规范（剥富文本标签后审）。
- 检查时用 `--semantic-json` 取待审行，由 Agent（LLM）执行；脚本不维护错别字词表。

---

## 独有规则

### 结构化规则（脚本）

1. **HeroId → Hero.Id**  
   `HeroId` 不能为空；必须等于同目录（或 `--hero`）`Hero.xlsx`（sheet：`武将表|Hero`）某行的 `Id`。  
   缺失 Hero 表：报一次并跳过本条外联。  
   找不到：`HeroId=<值> 在 Hero.Id 中不存在`。

2. **SkinPinYin ↔ Hero.E#EHeroId**  
   在 `HeroId` 已命中武将行后：取该行类型列 `E#EHeroId`，与本行 `SkinPinYin` 比对（**大小写不敏感**）。下列任一即算对得上：  
   - `SkinPinYin` == `E#EHeroId`  
   - `SkinPinYin` 以 `E#EHeroId + '_'` 开头（如 `caocao_xiangao`、`caocao_gangman`）  
   - `SkinPinYin` 以 `E#EHeroId` 为前缀的粘连后缀（如 `ZhuRong`→`zhurongfuren`）  
   - 若 `E#EHeroId` 含下划线（如 `ZhangJiao_Guide`）：取其**第一段**作为 base，再按上三条与 `SkinPinYin` 比对  
   `SkinPinYin` 空：报不能为空。对不上：`SkinPinYin=<…> 与 Hero.E#EHeroId=<…>（HeroId=<…>）对不上`。  
   `HeroId` 未命中时本条不做拼音比对。

3. **正式上线武将须具备 4 种皮肤**  
   仅对 `Hero.xlsx` 中 **`HeroType=1` 且 `IsOpen=1`** 的武将检查（排除引导将、山贼、年兽、副本位等非正式类型）。  
   须在本表按 `HeroId` 找到以下 4 种 `SkinType` 各至少一行：  
   - `SkinNormalSkin`（原画）  
   - `SkinLineSkin`（工笔白描）  
   - `SkinNormalDynamicsSkin`（动态原画）  
   - `SkinHKComicsSkin`（兼工带写）  
   缺任一：`上线正式武将 HeroId=<Id>（Name=<名>）缺少皮肤类型: …`（展示列用武将名；`SkinItemId=-`）。  
   Hero 表缺失则本条随外联一并跳过。

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
用户: 检查 HeroSkinItem_英雄皮肤.xlsx
→ python "HeroSkinItem_英雄皮肤/scripts/check_HeroSkinItem_yingxiongpifu.py" "<路径>" [--hero ...]
→ 需要时 --semantic-json，按 L1（及独有语义）由 Agent 审
→ 合并结构化 + 语义报告
```

<!-- sync: HeroId→Hero.Id；SkinPinYin↔E#EHeroId；HeroType=1且IsOpen=1 四套 SkinType；L1 -->
