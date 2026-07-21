---
name: Hero
description: |
  校验名将杀 Hero.xlsx（武将表）。
  当用户提到 Hero.xlsx、武将表检查、武将配置校验时使用。
  公共流程见 Excel-check/SKILL.md；本文件仅含本表规则。
---

# Hero — 武将表检查

校验 `Hero.xlsx`（sheet：`武将表|Hero`）。公共约定见 [Excel-check/SKILL.md](../SKILL.md)。

**表结构**：A（经典行表）
**主键列**：`Id`

**运行时输入**：用户提供 `Hero.xlsx` 路径。本表会读取同目录（或 `--skill`）下的 `Skill.xlsx`，校验武将 `Skill` 字段中的技能 Id 是否存在于技能表。

**业务标识列**：`Name`（武将名）。独有规则主要围绕 `HeroType` 分叉。

**读表补充**：第 3 行无字段名的列默认不参与校验；例外：`E#EHeroId`、`E#EHeroType` 纳入读表（与 `Name` / `HeroType` 对齐）。

## 脚本

```bash
python "Hero/scripts/check_Hero.py" "<路径>"
python "Hero/scripts/check_Hero.py" "<路径>" --skill "<Skill.xlsx路径>"
python "Hero/scripts/check_Hero.py" "<路径>" --semantic-json
python "Hero/scripts/check_Hero.py" "<路径>" --json
```

Issue 展示列用 `Name`：`Id=<Id> | Name=<Name> | <字段> | <说明>`

Agent 向用户汇报时：原样列出脚本输出的每条 Issue 行；禁止用分类汇总表代替（细则见 [Excel-check/SKILL.md](../SKILL.md)「Agent 汇报硬性要求」）。

---

## 通用规则

### 结构化规则（脚本）

适用公共类型（落到本表）：

| 编号 | 适用? | 落到本表 |
|------|-------|----------|
| S1 | 是 | `Id` |
| S12 | 是 | 有值字段按类型行（int/bool/string/E*/数组等） |
| S2 | 是 | `Name`；`HeroType`（与 `E#EHeroType` 同空时可空） |
| S3 | 否 | — |
| S4 | 是 | `HeroType`, `Gender`, `Country`, `BelongExpansionPack`, `IsOpen` 等 |
| S5 | 否 | — |
| S6 | 否 | —（`Name` **允许**重复；唯一性在 `E#EHeroId`） |
| S7 | 否 | —（`OpenDate` 可与 `IsOpen` 不同时有值；有值时走格式） |
| S8 | 否 | — |
| S9 | 是 | `Skill`, `ExcludeIdentity`, `NotUseModeType`, `Buff`, `AssociationHeroId`, `MeltName` |
| S10 | 否 | — |
| S11 | 部分 | `OpenDate` 有值时须为 `YYYY-MM-DD HH:MM:SS` |

读表约定：空/`#` Id 跳过；连续空 3 行截断；第 2、3 行皆空列丢弃；无字段名的列默认不检查，**例外**：`E#EHeroId`、`E#EHeroType` 纳入读表。

#### 字段细则

- **Id**：int，全表不重复  
- **Name**：非空 string（中文展示名，**允许重复**）  
- **E#EHeroId**：非空；形如拼音/标识符 `^[A-Za-z][A-Za-z0-9_]*$`；**全表不重复**  
- **HeroType** 与 **E#EHeroType**：须一致——二者同空，或同有值；`E#EHeroType` 为空时 `HeroType` 允许为空；有值时 `HeroType` ∈ `{1,2,3,4,5,6}`  
- **IsOpen / IsAlwaysZhuGong / CanMelt / IsNewHero / IsGacha**：可空（视 HeroType）；有值须为 bool 语义（`true`/`false`/`0`/`1`）  
- **OpenDate**：可空；有值须 `YYYY-MM-DD HH:MM:SS`  
- **Skill / ExcludeIdentity / NotUseModeType / AssociationHeroId**：可空；有值须 `^\d+(,\d+)*$`  
- **Buff**：可空；有值须非空白 string（或同逗号分隔枚举名，不作外联）  
- **MeltName**：可空；有值须逗号分隔的非空片段（`string[]`）

### 语义规则

- **L1 文案质量**：对非空 `Name`、`MeltName` 等展示名检查错字、漏字、病句、标点（专有名词吃不准标需人工确认）。
- 下列玩法语义为**独有**（非公共 L*）：

检查时 **必须** 读取 `--semantic-json`（或 `--json` 的 `semantic_rows`），对每行做下列判断。

#### Name ↔ E#EHeroId（中文 ↔ 拼音）

`Name` 为中文名，`E#EHeroId` 为其拼音（或约定英文标识，如 `CaoCao`、`ZhenFu`）。  
脚本**不**维护拼音对照表；明显不是该 `Name` 的合理拼音/惯用写法 → 报 issue（或「需人工确认」）。允许字号、通假、后缀变体（如 `甄宓`→`ZhenFu`，`张角`→`ZhangJiao_Guide`）。占位/熔炼槽等非人名（`局内1号位`→`MeltHeroPos1`）按约定英文标识通过，不强行拆成字拼音。

**特例（拼音偏差仍通过）**：

| Id | Name | E#EHeroId | 说明 |
|----|------|-----------|------|
| `10008` | 许褚 | `XuZhu` | 历史误配已外发，无法改正；不报 XuChu 偏差 |
| `10701` | 陈胜 | `ChengSheng` | 历史误配已外发，无法改正；不报 ChenSheng 偏差 |

#### Name → Country（由武将名推断国号）

当行上 `Country` 非空、且 `Name` 为可识别历史/演义人物（非局内位、熔炼槽等占位名）时：根据人物所属势力/时代，推断应属国号，与配置的 `Country` 比对。

| Country | 含义（推断用） |
|---------|----------------|
| `CaoWei` | 曹魏 |
| `Shu` | 蜀汉 |
| `SunWu` | 东吴 |
| `DongHan` | 东汉末群雄等（董卓、袁绍、刘协等） |
| `Huang` | 黄巾 |
| `XiHan` | 西汉/汉初 |
| `XiChu` | 西楚 |
| `Qin` | 秦 |
| `XiJin` | 西晋 |
| `Zhao` / `Wei` / `Han` / `Qi` / `Chu` / `Yan` | 战国对应国 |
| `ZhangChu` | 张楚（陈胜等） |
| `XiZhou` | 西周等特例 |
| `NianShou` / `Fei` 等 | 活动/特殊势力，不按常理人名硬推时跳过或「需人工确认」 |

明显矛盾才报（如 `曹操` 配成 `Shu`）；吃不准 →「需人工确认」或不报。跨势力人物以游戏配置阵营为准、存在合理歧义时不报。脚本**不**维护人名→国号对照表。

#### Name → MeltName（由武将名推断熔炼碎片名）

当 `MeltName` 非空（通常 `HeroType=1` 且可熔炼）时：结合 `Name` 与该人物惯用字号/别称，判断熔炼串是否合理。

样例形态（逗号分隔，常见 6 段）：`字号或别称,姓相关字,名相关字,…`  
如 `曹操`→`孟德,曹,孟,曹,孟,德`，`夏侯惇`→`元让,夏,惇,夏,侯,惇`。

| 检查点 | 说明 |
|--------|------|
| 与姓名相关 | `Name` 中的姓/名用字应在 `MeltName` 片段中有体现（复姓拆开亦可） |
| 字号/别称 | 首段多为字或别称，须与该人物相符（允许通假/别称变体） |
| 空值 | `MeltName` 为空时本条不报（未配熔炼名交给结构化/业务补全） |

**特例（不报 Name 用字缺失）**：下列行按别称/通假配熔炼名，不因 `Name` 中「姬」未出现在 `MeltName` 而报 issue：

| Id | Name | MeltName（约定通过） |
|----|------|----------------------|
| `10402` | 虞姬 | `虞美人,虞,美,虞,美,人` |
| `10901` | 如姬 | `如妃,如,妃,如,妃,妃` |

明显无关（如姓名用字全未出现、字号张冠李戴）才报；吃不准不报。脚本**不**维护人名→字号对照表。

#### Name → BelongExpansionPack（由武将名推断扩展包）

当 `BelongExpansionPack` 非空、且 `Name` 为可识别人物时：按人物所属历史时代/题材，推断应属扩展包，与配置比对。

| BelongExpansionPack | 题材（推断用） |
|---------------------|----------------|
| `HeroExpansionPack_SanGuoYunQi` | 三国 |
| `HeroExpansionPack_ChuHanZhiZheng` | 楚汉 |
| `HeroExpansionPack_WuDiShengShi` | 西汉武帝线等 |
| `HeroExpansionPack_QinSaoLiuHe` | 秦 |
| `HeroExpansionPack_HeZongLianHeng` | 战国合纵连横 |
| `HeroExpansionPack_JinLuoXingTi` | 西晋 |
| `HeroExpansionPack_JiangXinTianGong` | 机关/工匠向特例（如偃师、马钧） |

明显矛盾才报（如 `项羽` 配成三国包）；跨时代或活动向吃不准 →「需人工确认」或不报。脚本**不**维护人名→扩展包对照表。

#### Name → Gender（由武将名推断性别）

当 `Gender` 非空、且 `Name` 为可识别对象时：按人物/对象属性推断，与配置比对。

| Gender | 含义 |
|--------|------|
| `1` | 男 |
| `2` | 女 |
| `3` | 太监 |
| `0` | **无性别**（动物等无法按人类男女定义的对象，如年兽） |

明显矛盾才报（如 `甄宓` 配成 `1`，`曹操` 配成 `2`；太监误配 `1`/`2`/`0`，无性别对象误配 `1`/`2`/`3` 等）。脚本**不**维护人名→性别对照表。

---

## 独有规则

### 结构化规则（脚本）

业务钥匙：`HeroType`（与 `E#EHeroType` 对齐后方可分叉）。

#### 0. HeroType ↔ E#EHeroType 空/非空一致

| E#EHeroType | HeroType |
|-------------|----------|
| 空 | 必须为空（不报缺省；本行跳过按类型分叉） |
| 有值 | 必须有值且为合法枚举 int |

一侧有值另一侧空 → 报不一致。

#### 1. 按 HeroType 分叉必填 / 应空

| HeroType | 含义（据样例） | 额外要求 |
|----------|----------------|----------|
| `1` | 正式可玩武将 | 见 §2（含开放态加强） |
| `3` | 局内位 / 熔炼槽占位 | `IsOpen`/`Gender`/`Point`/`HpLimit`/`HandLimit`/`EquipLimit`/`Country`/`Skill`/`CanMelt`/`MeltName`/`BelongExpansionPack`/`OpenDate` **应为空** |
| `2` / `4` / `5` / `6` | 特殊/副本等 | `IsOpen`/`Gender`/`Point`/`HpLimit`/`HandLimit`/`EquipLimit`/`Country` **必填**；`CanMelt`/`MeltName`/`BelongExpansionPack` **应为空** |

#### 2. 正式武将（HeroType=1）

**基线（不论是否开放）**

- `IsOpen` / `Gender` / `Point` / `HpLimit` / `HandLimit` / `EquipLimit` / `Country` / `IsAlwaysZhuGong` / `CanMelt` / `BelongExpansionPack` **必填**
- `Gender` ∈ `{1,2}`（正式武将按男女；太监见语义 `3`、无性别见语义 `0`，一般不落在开放正式将）
- `Point` = `HpLimit`；`EquipLimit` = `3`
- `Country`、`BelongExpansionPack` ∈ 枚举白名单（脚本常量）

**开放态加强（`IsOpen` 为真 / `1`）**：相关业务信息**不得为空**，并核格式：

| 字段 | 要求 |
|------|------|
| `Skill` | 必填；`^\d+(,\d+)*$` |
| `MeltName` | 必填；逗号分隔且**每段非空**（常见 6 段字号/姓名拆字） |
| `NotUseModeType` | 必填；`^\d+(,\d+)*$` |
| `Name` / `E#EHeroId` | 已由通用规则保证非空与拼音格式 |

下列可空（样例多为空或选填）：`OpenDate`、`Buff`、`AssociationHeroId`、`ExcludeIdentity`、`IsGacha`。

未开放（`IsOpen=0`）时：`Skill` / `MeltName` 允许空（配置中未完成项）。

#### 3. Name（中文）与 E#EHeroId（拼音）

- `Name`：**允许**重复（跨行同名合法）  
- `E#EHeroId`：必填、格式 `^[A-Za-z][A-Za-z0-9_]*$`、**全表唯一**  
- 二者语义对应关系见上方「通用规则 / 语义」

#### 4. AssociationHeroId 指向本表 Id

`AssociationHeroId` 有值时，拆出的每个 int 必须是本表某行的 `Id`。

#### 5. Skill ↔ Skill.xlsx

武将行 `Skill` 有值时，拆出的每个技能 Id 必须出现在 `Skill.xlsx`（技能表）的 `Id` 列中。

默认技能表路径：与 `Hero.xlsx` 同目录的 `Skill.xlsx`；可用 `--skill` 覆盖。技能表缺失则报错并跳过本条逐行校验。

### 语义规则

与「通用规则 / 语义」相同：检查流程须跑 **Name↔E#EHeroId**、**Name→Country**、**Name→MeltName**、**Name→BelongExpansionPack**、**Name→Gender**。

---

## 补充规则时（必须）

按用户要求为本表 **新增/修改规则**前：先对照本文件已有「通用规则」「独有规则」（及对应脚本实现）。

| 情况 | 处理 |
|------|------|
| 与现有规则实质重复 | 先反馈重复点，勿落盘；询问是否保留/合并/取消 |
| 与现有规则冲突 | 先列出冲突双方，停止实现；询问以哪方为准 |
| 无重复且无冲突 | 再写入本文件，并视需要改脚本 |

细则见 [Excel-check/SKILL.md](../SKILL.md)「使用者后续补充」。

---

## 工作流程

```
用户: 检查 Hero.xlsx
→ python "Hero/scripts/check_Hero.py" "<路径>" [--skill "<Skill.xlsx>"]
→ python "Hero/scripts/check_Hero.py" "<路径>" --semantic-json
→ 按通用语义规则分析 Name ↔ E#EHeroId、Name → Country / MeltName / BelongExpansionPack / Gender
→ 合并结构化 + 语义报告（Issue 行格式；禁止用汇总表代替）
```
