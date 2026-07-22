---
name: SkillUI_技能表现表
description: |
  校验名将杀 SkillUI_技能表现表.xlsx（技能表现配置表）。
  当用户提到 SkillUI_技能表现表.xlsx、技能表现表检查、SkillUI 配置校验时使用。
  公共流程见 Excel-check/SKILL.md；本文件仅含本表规则。
---

# SkillUI_技能表现表 — 技能表现配置表检查

校验 `SkillUI_技能表现表.xlsx`（sheet：`技能表现配置表|SkillUI`）。公共约定见 [Excel-check/SKILL.md](../SKILL.md)。

**表结构**：A（经典行表）
**主键列**：`Id`

**运行时输入**：用户提供 `SkillUI_技能表现表.xlsx` 路径。本表会读取同目录（或 `--skill`）下的 `Skill.xlsx`，校验本表 `Id`（字符串技能标识）与 `RelatedSkill`（数值技能 Id）外联。

**业务标识列**：`SkillName`（技能中文名）。`Id` 为字符串技能标识（如 `Sha`、`TaiPingYaoShu`）。

**读表补充**：

- 元数据 4 行（中文说明 / 类型 / 字段名 / client），**数据从第 5 行起**（0-based 行号 4）
- 第 2、3 行皆空的列丢弃
- **连续空 3 行后截断**：与公共约定一致；其后内容（如冲杀/闪避/蟠桃等）**不再检查**

## 脚本

```bash
python "SkillUI_技能表现表/scripts/check_SkillUI_jinengbiaoxianbiao.py" "<路径>"
python "SkillUI_技能表现表/scripts/check_SkillUI_jinengbiaoxianbiao.py" "<路径>" --skill "<Skill.xlsx路径>"
python "SkillUI_技能表现表/scripts/check_SkillUI_jinengbiaoxianbiao.py" "<路径>" --semantic-json
python "SkillUI_技能表现表/scripts/check_SkillUI_jinengbiaoxianbiao.py" "<路径>" --json
```

Issue 展示列用 `SkillName`：`Id=<Id> | SkillName=<SkillName> | <字段> | <说明>`

Agent 向用户汇报时：原样列出脚本输出的每条 Issue 行；禁止用分类汇总表代替（细则见 [Excel-check/SKILL.md](../SKILL.md)「Agent 汇报硬性要求」）。

---

## 通用规则

### 结构化规则（脚本）

适用公共类型（落到本表）：

| 编号 | 适用? | 落到本表 |
|------|-------|----------|
| S1 | 是 | `Id`（string，全表唯一） |
| S12 | 是 | 有值字段按类型行（int/bool/string/E*/数组等） |
| S2 | 是 | `SkillName`、`SkillText`、`ShortSkillText` |
| S3 | 是 | `Allusion` / `DesignThought` 有值时须以 `<color=#FFFFFF00>占位</color>` 开头（见独有细化） |
| S4 | 是 | `HasRelation` 有值时为 bool 语义 |
| S5 | 否 | — |
| S6 | 否 | — |
| S7 | 是 | `HasRelation` ↔ `RelatedSkill` |
| S8 | 否 | — |
| S9 | 是 | `RelatedSkill`/`Audio`/`KeyWords`/`SkillTag`/`IdentityLine`/`PetDesKey` 等格式 |
| S10 | 否 | — |
| S11 | 否 | — |

#### 字段细则

- **Id**：非空 string；形如 `^[A-Za-z][A-Za-z0-9_]*$`；全表不重复  
- **SkillName / SkillText / ShortSkillText**：必填非空  
- **HasRelation**：可空；有值须为 bool 语义（`true`/`false`/`0`/`1`）  
- **RelatedSkill**：可空；有值须 `^\d+(,\d+)*$`  
- **Audio / KeyWords / SkillTag**：可空；有值须 `^\d+(,\d+)*$`  
- **IdentityLine**：可空；有值须非空音频标识串（可逗号分隔多个 `EAudioId`）  
- **PlayCardAudio / SpecialAudio**：可空；有值须非空白 string  
- **PetDesKey**：可空；有值须匹配一个或多个 `{int;int;int;int;int}` 紧邻拼接  
- **SettlementDes / AuraButtonDes / BattleSkillStep**：可空；有值须非空白 string  

### 语义规则

- **L1 文案质量**：覆盖 `SkillName`、`ShortSkillText`、`SkillText`、`SettlementDes`、`Allusion`、`DesignThought` 等文本列（错字/漏字/病句/标点）；细则与玩法对照见下列各节（含独有语义）。

检查时 **必须** 读取 `--semantic-json`（或 `--json` 的 `semantic_rows`），对每行做下列判断。

#### SkillName ↔ Id（中文 ↔ 拼音标识）

`SkillName` 为中文名，`Id` 为其拼音或约定英文标识（如 `杀`→`Sha`，`万箭齐发`→`WanJianQiFa`）。  
脚本**不**维护拼音对照表；明显不是该 `SkillName` 的合理拼音/惯用写法 → 报 issue（或「需人工确认」）。允许词牌缩写、后缀变体（如 `HanXinCard`、`ForSiMaLiang`）。

**特例（历史误配已外发，无法改正；不报 Name↔Id 偏差）**：

| Id | SkillName |
|----|-----------|
| `ChengYe` | 举贤为将 |
| `LuanShiJianXiong` | 天不负我 |
| `GuoSeTianXiang` | 秋水伊人 |
| `YiJiDingLiaoDong` | 遗策平辽 |

#### ShortSkillText 短文案（必须）

对每行非空 `ShortSkillText` 做 LLM 语义（与 `SkillText` 对照；脚本不维护错别字/技能名对照表）。先用全表 `SkillName` 作词表。

| 检查点 | 说明 |
|--------|------|
| 错字 / 漏字 / 衍字 | 同 SkillText：明显不通才报 |
| 涉及技能名 | 点名、引号或可明确识别的技能/牌名，须与本表某行 `SkillName` 一致（或为本行自身）；对照词表同 SkillText |
| 与 SkillText 语义一致 | 短文案应是长文案的合理精简或等价改写，**不得**改核心效果/对象/时机/数值/关键限制。例：仅措辞、换行、省「你的」✓；效果对象矛盾 ✗ |
| 牌面对象用语不可混用 | **手牌、牌、任意牌**等是有区别的（以及同类：装备区的牌、判定区的牌、弃牌堆的牌等）。长短文案对照时不得互相替换或省略限定导致范围变化。例：长写「获得 1 张手牌」、短写成「获得 1 张牌」→ 报；长写「弃置 1 张牌」、短写成「弃置 1 张手牌」→ 报；长写「任意牌」、短改成「手牌」或反之 → 报。吃不准（上下文已明确同指）可不报 |
| 角色范围用语不可混用 | **其他、任意其他**等是有区别的（以及同类：其他角色 / 任意其他角色、一名其他角色 / 任意一名其他角色等）。长短文案对照时不得互相替换或省略「任意」等限定导致可选范围变化。例：长写「选择一名其他角色」、短写成「选择任意其他角色」→ 报；长写「任意其他角色」、短改成「其他角色」→ 报。吃不准可不报 |
| 允许省略 | 可省略长文案里的关键字释义（`<color>` 灰字说明）、嵌套牌的完整复述、纯装饰标签；省略后剩余句子仍须自洽、且不与长文案矛盾 |
| 禁止额外加严/改义 | 长文案没有的随机性、次数、持续时机等，短文案**不得擅自加上**（如长写「获得 1 张牌」、短写成「随机获得 1 张牌」→ 报）。长文案有的关键持续/失效条件，短文案若省略导致效果范围明显变宽，报 issue 或「需人工确认」。**牌/手牌/任意牌**及**其他/任意其他**等范围用语见上两条，不得借「精简」改掉 |
| 长度角色 | 去掉标签后，短文案一般不应明显长于长文案；若短更长且在堆砌新信息 → 报 |
| 色标用途 | 与 SkillText 同一套（见「SkillText / ShortSkillText 色标用途」） |

吃不准不报。样例里多数 diff 仅为去释义/`\n`，视为通过。

#### SkillText 文案质量与技能名指称（必须）

对每行非空 `SkillText` 做 LLM 语义检查（脚本不维护错别字/技能名对照表）：

| 检查点 | 说明 |
|--------|------|
| 错字 / 别字 | 明显用错汉字（如「共」写「攻」导致不通、技能常用词写错）才报 |
| 漏字 / 衍字 | 明显缺字或多余字导致句意残缺、不通才报 |
| 标点/格式导致不通 | 仅当严重影响理解时报；富文本标签语法本身不报；**颜色用途**见下「SkillText / ShortSkillText 色标用途」 |
| 涉及其他技能名 | 文案中点名、引号或可明确识别的技能/牌名，应与本表某行 `SkillName` **一致**（或为本行自身 `SkillName`）。例：写「当作“闪”打出」→ 表内应有 `SkillName=闪`；写成「当作“閒”」或与任何 `SkillName` 对不上 → 报 |
| 色标用途 | 见「SkillText / ShortSkillText 色标用途」；与 ShortSkillText 同一套 |

执行方式：先从 `--semantic-json` **全部行**收集 `SkillName` 集合作对照词表，再分批审 `SkillText`。吃不准（泛称「技能」「装备」未点名、典故比喻）不报。

#### SkillText / ShortSkillText 色标用途（必须）

对非空 `SkillText`、`ShortSkillText` 中的 `<color=…>`（可带 `<size=…>`）按用途审；脚本不维护词表，由 Agent（LLM）对照标签内文本判定。明显用错色时报；吃不准、历史存量灰色地带不报。

| 色标 | 用途 |
|------|------|
| `<color=#FFFFFF00>` | 用于占位符 |
| `<color=#B9C7D8><size=30>` | 用于解释武将技能文本中提到的专有名词或新卡牌 |
| `<color=#FFD900>` | 用于高亮武将技能描述中的常见时机点或常见限制，例如“出牌阶段限1次”、“每轮开始时”、“每名其他角色限1次”、“每个回合限1次”、“回合开始时”等等 |
| `<color=#FC8416>` | 用于高亮武将技能描述中的“登场”、“限定”、“阵亡” |
| `<color=#98FF98>` | 用于高亮武将技能描述中的游戏特定时机点，例如“受伤”、“杀伤”、“应战”、“出杀”、“击杀”等类似词语 |
| `<color=#00FFB4>` | 早期技能中用于高亮特殊动作（效果）或特殊区域，现已不再新增 |

补充约定：

- 上表色值大小写不敏感（`#ffd900` 等同 `#FFD900`）。
- `#B9C7D8` 释义段通常配 `<size=30>`；缺 size 但色与释义用途明显匹配时，吃不准可不报。
- `#00FFB4`：**新技能/明显新配文案不应再使用**；表内历史存量且用途符合「特殊动作/区域」→ 不报。
- 出现上表以外的 `<color=#…>`，且不像笔误/旧版残留 → 报 issue 或「需人工确认」。
- 与「标签语法本身不报」不冲突：本条查的是**色↔语义是否匹配**，不是标签能否解析。

#### SettlementDes 结算详情（必须）

`SettlementDes` **有值**时纳入语义（空值不报；本字段为可选补充说明）。脚本不维护错别字/技能名对照表。

| 检查点 | 说明 |
|--------|------|
| 错字 / 漏字 / 衍字 | 同 SkillText：明显不通才报 |
| 涉及技能名 | 点名、引号或可明确识别的技能/牌名，须与本表某行 `SkillName` 一致（或为本行自身 `SkillName`）；对照词表同 SkillText |
| 与 SkillText 对应 | `SettlementDes` 是对 `SkillText` 的**补充解释**（细则、边界、结算顺序等），不得与 `SkillText` 的核心效果/对象/时机/条件**明显矛盾**（如文案写「摸 2 张」结算写「摸 1 张」且无说明例外）。允许只展开、举例、列注意事项而不复述全文；吃不准不报 |

执行方式：读 `--semantic-json` 同行的 `skill_text` + `settlement_des` 对照审；技能名仍用全表 `SkillName` 词表。

#### Allusion 技能典故（必须）

`Allusion` **有值**时纳入语义（空值不报）。结构化已要求占位色前缀；语义审的是**去掉** `<color=#FFFFFF00>占位</color>` 及富文本标签后的正文。脚本不维护典故/错别字对照表。

| 检查点 | 说明 |
|--------|------|
| 错字 / 漏字 / 衍字 | 明显不通才报。引文保留繁体/古字（如「殺」、《说文》原文）**不**当错字 |
| 涉及技能名 | 若点名本表其他技能/牌，须与 `SkillName` 词表一致；典故人名/书名不按技能名硬套 |
| 与本技能相关 | 典故应能对上本行 `SkillName`（或技能主题）：字源、出处、史事、典故与牌面含义相关。完全张冠李戴（如「桃」的典故写成无关武器史）才报；略牵强、【原创】、吃不准不报 |
| 仅占位无正文 | 有字段值但去前缀/标签后正文为空 → 报（配了典故栏却无内容） |

#### DesignThought 设计思路（必须）

`DesignThought` **有值**时纳入语义（空值不报）。审正文（去占位前缀与标签）同 Allusion。

| 检查点 | 说明 |
|--------|------|
| 错字 / 漏字 / 衍字 | 明显不通才报 |
| 涉及技能名 | 点名技能/牌须与 `SkillName` 词表一致（或为本行自身）；五行/`<sprite>` 等设定符号本身不报 |
| 与本技能 / SkillText 相关 | 设计思路应解释本技能命名或机制立意（为何叫此名、与效果如何挂钩）。与 `SkillName`/`SkillText` 主题明显无关或描述成另一个技能 → 报。允许简短、允许「原创」、允许只谈美术/五行隐喻；吃不准不报 |
| 仅占位无正文 | 去前缀后为空 → 报 |

执行：`--semantic-json` 含 `allusion`、`design_thought`（及同行 `skill_name`/`skill_text`）；有值再审。

## 独有规则

### 结构化规则（脚本）

业务钥匙：`Id`（字符串技能标识）；关联键：`RelatedSkill`（数值技能 Id）。

#### 1. Id ↔ Skill.xlsx（E#ESkillId）

本表每个 `Id` 必须出现在同目录 `Skill.xlsx`（技能表）无字段名列 `E#ESkillId`（类型行 `E#ESkillId`）中。  
可用 `--skill` 覆盖路径；技能表缺失则报错并跳过本条逐行校验。

#### 2. HasRelation ↔ RelatedSkill

| 条件 | 要求 |
|------|------|
| `HasRelation` 为真（`1`/`true`） | `RelatedSkill` **必填**，且格式 `^\d+(,\d+)*$` |
| `RelatedSkill` 有值 | `HasRelation` 须为真；不可为假或空 |
| `HasRelation` 为假或空，且 `RelatedSkill` 空 | 通过 |

#### 3. RelatedSkill ↔ Skill.xlsx（数值 Id）

`RelatedSkill` 拆出的每个 int 必须出现在 `Skill.xlsx` 的 `Id` 列中。

#### 4. Allusion / DesignThought 占位前缀

字段**有值**时，须以 `<color=#FFFFFF00>占位</color>` 开头（与 Mail Body 同类占位色）。空值不报。

### 语义规则

与「通用规则 / 语义」相同：须跑 **SkillName↔Id**、**ShortSkillText**、**SkillText 文案质量与技能名指称**、**SkillText / ShortSkillText 色标用途**、**SettlementDes**、**Allusion**、**DesignThought**。

另可选对照：同 `Id` 在 `Skill.xlsx` 的 `SkillName` 与本表 `SkillName` 是否通指同一技能（仅明显矛盾才报；字数简称如 `破阵`↔`破阵卸甲`、通假字等吃不准不报）。

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
用户: 检查 SkillUI_技能表现表.xlsx
→ python "SkillUI_技能表现表/scripts/check_SkillUI_jinengbiaoxianbiao.py" "<路径>" [--skill "<Skill.xlsx>"]
→ python "SkillUI_技能表现表/scripts/check_SkillUI_jinengbiaoxianbiao.py" "<路径>" --semantic-json
→ 按通用语义规则分析 SkillName↔Id、ShortSkillText、SkillText（含色标用途）、SettlementDes、Allusion、DesignThought
→ 合并结构化 + 语义报告（Issue 行格式；禁止用汇总表代替）
```
