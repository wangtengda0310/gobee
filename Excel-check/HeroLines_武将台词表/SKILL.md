---
name: HeroLines_武将台词表
description: |
  校验名将杀 HeroLines_武将台词表.xlsx。
  当用户提到 HeroLines_武将台词表.xlsx、HeroLines_武将台词表配置检查时使用。
  公共流程见 Excel-check/SKILL.md；本文件仅含本表规则。
---

# HeroLines_武将台词表 — 配置表检查

校验 `HeroLines_武将台词表.xlsx`（sheet：`武将台词|HeroLines`）。公共约定见 [Excel-check/SKILL.md](../SKILL.md)。

**表结构**：A（经典行表）
**主键列**：`Id`
**运行时输入**：用户提供 `HeroLines_武将台词表.xlsx` 路径。本表会读取同目录（或 `--skill` / `--skill-lines` / `--hero`）下的 `Skill.xlsx`、`SkillLines_技能台词表.xlsx`、`Hero.xlsx`。

**业务标识列**：`TabName`。

**读表补充**：类型行=1，字段名行=2，数据起始行（0-based）=`4`；连续空 3 行截断；`#` 分区行跳过；第 2、3 行皆空列丢弃。

## 脚本

```bash
python "HeroLines_武将台词表/scripts/check_HeroLines_wujiangtaicibiao.py" "<路径>"
python "HeroLines_武将台词表/scripts/check_HeroLines_wujiangtaicibiao.py" "<路径>" --skill "<Skill.xlsx路径>"
python "HeroLines_武将台词表/scripts/check_HeroLines_wujiangtaicibiao.py" "<路径>" --skill-lines "<SkillLines路径>"
python "HeroLines_武将台词表/scripts/check_HeroLines_wujiangtaicibiao.py" "<路径>" --hero "<Hero.xlsx路径>"
python "HeroLines_武将台词表/scripts/check_HeroLines_wujiangtaicibiao.py" "<路径>" --json
python "HeroLines_武将台词表/scripts/check_HeroLines_wujiangtaicibiao.py" "<路径>" --semantic-json
```

Issue：`Id=<Id> | TabName=… | <字段> | <说明>`；若归属武将 `IsOpen=0`，说明末尾追加 `；武将还未开放（武将名）`。

Agent 汇报：原样列出每条 Issue 行；禁止用分类汇总表代替。L1 语义若 `--semantic-json` 行含 `hero_not_open`，汇报时同样追加。

---

## 通用规则

### 结构化规则（脚本）

| 编号 | 适用? | 落到本表 |
|------|-------|----------|
| S1 | 是 | `Id` |
| S12 | 是 | 有值字段按类型行（int/bool/string/E*/数组等） |
| S2 | 是 | `TabName`、`Text`、`AudioId` 非空（见独有） |
| S4 | 是 | 见独有：`TabName`→`Type` 枚举 |
| S9 | 部分 | 类型行为数组等时 |
| S11 | 部分 | 成对时间字段（若存在） |
| L1 | 是 | `TabName` / `Text` |

主要字段：`Id`, `Type`, `TabName`, `Text`, `AudioId`, `Achievements`, `GroupId`

#### 字段细则

- **Id**：int，不重复（S1）
- **TabName**：必填非空（S2 / 独有）
- **Text**：必填非空（**年兽除外**，见独有）；全表非空 Text 不重复（S2 / 独有）
- **AudioId**：必填非空；全表不重复；武将行须符合 `Vo_Hero_…` 命名（**年兽**为 `Vo_Boss_NianShou_…`，见独有）
- 其余字段默认可空；有值时按类型行做格式校验（S12）
- 若存在 `Name` / `Title` / `SkillName`：非空（S2）

### 语义规则

- **L1 文案质量**：对非空 `TabName`、`Text` 检查错字、漏字、病句、标点不规范（剥富文本标签后审）。
- 检查时用 `--semantic-json` 取待审行，由 Agent（LLM）执行；脚本不维护错别字词表。

---

## 独有规则

### 结构化规则（脚本）

业务钥匙：`TabName`；关联键：`Type`；外联：`Skill.xlsx` 的 `SkillName` / `E#ESkillId`、`SkillLines_技能台词表.xlsx`、`Hero.xlsx`（`Skill` / `IsOpen`）。

#### 1. TabName 必填

`TabName` 不能为空。

#### 2. TabName → Type（固定页签）

| TabName | Type（必须） |
|---------|--------------|
| `登场` | `LinesType_Dengchang` |
| `击杀` | `LinesType_Kill` |
| `阵亡` | `LinesType_Dead` |
| `自选` | 空（必须为空） |
| `重伤` | 空（必须为空） |
| `退场` | 空（必须为空） |

#### 3. TabName 为技能名 → Type + Skill.xlsx

当 `TabName` **不是**上表固定值（含 `自选`/`重伤`/`退场`）时，视为技能名称：

| 条件 | 要求 |
|------|------|
| `TabName` 能在同目录（或 `--skill`）`Skill.xlsx` 的 `SkillName` 中找到 | `Type` 必须为 `LinesType_Skill` |
| `TabName` 在 `SkillName` 中找不到 | 报：在 Skill 表找不到此技能 |

`Skill.xlsx` 缺失则报错并跳过本条外联逐行校验（仍校验固定页签映射与 TabName 非空）。

#### 4. LinesType_Skill → SkillLines 台词引用

当本行 `Type`=`LinesType_Skill`（**年兽除外**）时：

1. 用 `TabName` 在 `Skill.xlsx` 找同名 `SkillName`，取其 **`E#ESkillId`（技能拼音 ID）**；同名多行则取全部（如羁绊/皮肤变体）
2. 在同目录（或 `--skill-lines`）`SkillLines_技能台词表.xlsx` 中，按 `SkillId`=`E#ESkillId` 定位行（含不同 `SkinId` 的多行）
3. 本行 `Id` **必须**出现在这些行的 `SkillFirstLine` / `SkillSecondLine` / `SkillThirdLine` / `SkillForthLine` 任一列（`int` 或逗号分隔 `int[]`）中

| 情况 | 报错 |
|------|------|
| 对应 `SkillId` 在 SkillLines 中均不存在 | SkillLines 中找不到 SkillId=… |
| 找到了但各 `Skill*Line` 均未包含本行 `Id` | Id 须出现在 SkillLines … Skill*Line 中 |

`SkillLines` 缺失则报错并跳过本条外联。`TabName` 已在第 3 条报「Skill 找不到」时，本条不再重复查拼音。

#### 5. Text 必填与唯一

- `Text` 不能为空（**年兽除外**，见下「年兽特殊」）
- 全表所有非空 `Text`（去首尾空白后）不得重复；重复时报涉及的 `Id` 列表

#### 6. AudioId 必填、唯一与命名格式

- `AudioId` 不能为空
- 全表所有非空 `AudioId`（去首尾空白后）不得重复；重复时报涉及的 `Id` 列表

**格式**（武将行；段之间用 `_`）：

`Vo_Hero_{武将拼音}[_{SkinN}]_{页签段}_{编号}_{台词首字母}`

| 段 | 规则 |
|----|------|
| 前缀 | 固定 `Vo_Hero` |
| 武将拼音 | 武将中文名逐字拼音，每字拼音**首字母大写**后拼接（如貂蝉→`DiaoChan`） |
| SkinN | **可选**；皮肤台词为 `Skin`+数字（如 `Skin1`），无皮肤则省略 |
| 页签段 | 见下表；技能类为 `JN1`/`JN2`/…（`JN`+正整数） |
| 编号 | 两位数字（如 `01`、`02`） |
| 台词首字母 | 取 `Text` 中前 **6** 个汉字（不足则有几个用几个；忽略标点/非汉字），按**词组**取拼音后取**首字母大写**拼接（多音字按词消歧，如「长枪」→`CQ` 而非 `ZQ`） |

**TabName → 页签段：**

| TabName | 页签段 |
|---------|--------|
| `登场` | `DC` |
| `击杀` | `JS` |
| `阵亡` | `ZW` |
| `自选` | `ZX` |
| `重伤` | `ZS` |
| `退场` | `TC` |
| 其它（技能名） | `JN`+序号（`JN1`、`JN2`…） |

例：`Vo_Hero_DiaoChan_DC_01_TYYQZC`；带皮肤：`Vo_Hero_ZhaoYun_Skin1_DC_01_PMDQGC`。

本表无独立武将名列时：脚本校验整段形态、页签段与 `TabName`/`Text` 的对应，以及 `AudioId` 非空/唯一；武将拼音是否与某张武将表一致不在本条强制外联（后续可增补）。

#### 7. 年兽特殊（无台词）

判定：`AudioId` 以 `Vo_Boss_NianShou_` 开头（分区 `#年兽` 下普通/狂暴等）。

| 项 | 处理 |
|----|------|
| `Text` | **可空**；不报「Text 不能为空」；空则不做 L1 Text |
| `AudioId` | 仍须非空、全表唯一；**不做** `Vo_Hero_…` 格式 / 页签段 / 台词首字母校验 |
| Skill / SkillLines 外联 | **不做**（Boss 技能 TabName 如 `烈焰噬心1` 不必出现在 `Skill.xlsx` / `SkillLines`） |
| TabName / Type | 固定页签（登场/击杀/退场/重伤等）仍按第 2 条校验 |
| Hero 开放态标注 | **不做**（无常规武将归属） |

#### 8. 检查结果标注武将开放态（Hero.IsOpen）

对**每一条**结构化 Issue（及 L1 语义汇报）：

1. 定位本行归属武将：
   - 优先：`TabName` 为技能名 → `Skill.xlsx` 得数值 `Id` → `Hero.xlsx` 的 `Skill` 列含该 Id 的武将
   - 否则：从 `AudioId` 的 `Vo_Hero_{E#EHeroId}_…` 匹配 `Hero.xlsx` 类型行 `E#EHeroId`
2. 若归属武将 **`IsOpen=0`（未开放）**：在 Issue 说明末尾追加  
   `；武将还未开放（武将名）`  
   （多名未开放武将用顿号连接）
3. `IsOpen=1` 或不标；找不到归属武将则不加本标注
4. `Hero.xlsx` 缺失则报错并跳过本标注（其它规则照常）

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
用户: 检查 HeroLines_武将台词表.xlsx
→ python "HeroLines_武将台词表/scripts/check_HeroLines_wujiangtaicibiao.py" "<路径>" [--skill ...] [--skill-lines ...] [--hero ...]
→ 需要时 --semantic-json，按 L1（及独有语义）由 Agent 审；含 hero_not_open 则追加标注
→ 合并结构化 + 语义报告
```

<!-- sync: L1；TabName→Type；SkillName；SkillLines；Hero.IsOpen=0 标注；年兽例外 -->
