---
name: excel-check
description: |
  名将杀 Excel 配置表检查 skill 的公共框架与写作规范。
  当用户要新增某张表的检查 skill、抽取/统一表校验流程、或询问 Excel-check 怎么写时使用。
  具体某张表的字段规则与表独有业务规则写在对应子目录（如 Mail/、Recharge_充值表/），不写在本文件。
---

# Excel-check — 配置表检查框架

## 结构

```
Excel-check/
├── SKILL.md                 # 本文件：流程、通用规则类型、写新表方法论
└── <Table>/                 # 文件夹 = xlsx 主文件名（可含中文）
    ├── SKILL.md             # 通用规则 + 独有规则（各含结构化 / 语义）
    └── scripts/check_<ScriptId>.py
```

| 名称 | 规则 |
|------|------|
| `<Table>` | = 表文件名（无扩展名），可含中文。例：`Mail`、`ShopGoods_商品表` |
| `<ScriptId>` | **禁止中文**；英文保留，中文改无声调小写拼音。例：`check_Mail.py`、`check_ShopGoods_shangpinbiao.py`。纯英文时与 `<Table>` 相同 |

能脚本化的进脚本；语境/人物/语义判断进 LLM。**禁止**在脚本维护字号→武将名等易变对照表。

**两层规则，勿混淆（单表 skill 内各拆结构化 / 语义）：**

| 层级 | 单表 skill 中的位置 | 是什么 |
|------|---------------------|--------|
| 通用规则 | `## 通用规则` → 结构化 / 语义 | 把公共 S*/L* **落到本表字段**的细则 |
| 独有规则 | `## 独有规则` → 结构化 / 语义 | 本表玩法专属约定；**不要**写进本文件 S*/L* 细目。Agent 生成后，**使用者可继续在该表 skill 里直接增补** |

```
## 通用规则
### 结构化规则（脚本）
### 语义规则

## 独有规则
### 结构化规则（脚本）
### 语义规则
```

---

## 表结构规范（仅两种）

写新表 skill 前先判定本表属于 **A** 还是 **B**，读表与截断按对应规范落地；单表 skill 须标明 `表结构：A` 或 `表结构：B`。

### A — 经典行表（绝大多数）

横向字段、一行一条配置。元数据 + 数据区典型布局（0-based 行号）：

| 行 | 内容 |
|----|------|
| 0 | 中文列名 |
| 1 | 类型（`int` / `string` / `bool` / `E*` / `*[]` 等） |
| 2 | 英文字段名 |
| 3 | `server` / `client` / `server/client` 等端标记 |
| 4 或 5 起 | 业务数据（少数表元数据行数略有增减，`DATA_START_ROW` 可按 sheet 调整） |

**主键**：字段名行中用于唯一标识行的那一列，**名称不限**。常见：`Id` / `ID`、`ActivityId` / `SkinItemId` 等 `*Id`、或 `Dan` / `Level` / `Rank` / `key` / `ShopType` 等。不要因主键不叫 `Id` 就判为「结构不合规」。

**空列丢弃（仅 A）**：Excel 第 2 行（类型行）与第 3 行（字段名行）对该列均为空时，该列不参与校验（读表丢弃）。

### B — Global 系（竖向键值）

首行即列名，无「类型行 + 字段名行 + 端标记」那套元数据；从第 2 行起每行一条配置。典型列：

`name` | `value` | `type` | `sign` | `description`

例：`Global_全局配置表`、`Global_创角配置表`、`GuildCityName_城池名称表`、`BattleGlobal_战斗全局配置表` 等。

**主键**：一般为 `name`（配置键）；唯一性、截断均按该列。

读 B 表时**不要**套用 A 的「类型行@1 / 字段名行@2 / 数据@4」假设。

### 公共运行约定（A/B 共用）

- **连续空 3 行后截断**：数据区主键列连续为空达 3 行时，其后内容业务不用，脚本读表即截断、**不再检查**（单行/双行空行仍继续）
- xlsx 路径由用户提供；默认不读外联表（单表 skill 另有说明除外）
- 依赖：`pip install pandas openpyxl`

### 公共读表约定不可擅自例外（必须）

下列约定对**所有**单表 skill / 脚本默认生效，**高于**「样例看起来像还有数据」的直觉：

| 约定 | 适用 | Agent 禁止事项 |
|------|------|----------------|
| 连续空 3 行截断 | A/B | **禁止**因截断点之后仍能看到主键/名称，就改成「中段空行不截断 / 读到文件末尾」 |
| 第 2、3 行皆空列丢弃 | **仅 A** | **禁止**仅因数据区某列有值，就强行纳入未命名空类型列（单表 skill 另有白名单例外除外，须用户确认） |
| 元数据行数 / `DATA_START_ROW` | **仅 A** | 仅当该 sheet 元数据行数与默认明显不同时可调整起始行；**不得**借机改写截断语义；**不得**把 B 表误读成 A |
| 表结构 A/B | 全部 | 摸表后写错结构类型，或用 A 读法硬读 B（反之亦然） |

若摸表时发现「空 3 行后仍有多行」，正确流程：

1. **仍按截断实现**（其后不检查）  
2. 在汇报规则清单时**单独提示**用户：截断点后约有 N 行未检（可举 1～3 个主键例）  
3. **仅当用户明确要求**「空三行后仍要检查 / 本表不截断」时，才写入该表独有规则并改脚本  

新建或改单表 skill 时：不得把「放宽公共读表约定」写成默认可归纳出的独有规则。

---

## 两阶段校验

| 阶段 | 执行者 | 规则集 |
|------|--------|--------|
| 1. 结构化 | 脚本 | 单表「通用规则 / 结构化」+「独有规则 / 结构化」 |
| 2. 语义 | Agent（LLM） | 单表「通用规则 / 语义」+「独有规则 / 语义」（皆无则标明仅结构化） |

**字段 ↔ 规则多对多**：一字段可挂多条；一规则可落多列。打标时不要「一字段一条就停」。例：Mail 的 `Sender` 可同时有 S2+S5+S6+L1。

### 结构化规则（S*）— 通用类型

| 编号 | 规则 | 说明 / 来源例 |
|------|------|----------------|
| S1 | 主键类型与唯一 | 主键列（A：字段名行中的标识列，名不限；B：多为 `name`）类型正确、不重复 |
| S2 | 必填 / 非空 | 如 `Title` |
| S3 | 固定前缀 / 模板开头 | 如 Body 的 `<color=…>占位</color>` |
| S4 | 条件 → 枚举值 | 如 Title → `SenderType` |
| S5 | 条件 → 固定字符串 | 如特定 Title → `Sender`=`系统` |
| S6 | 跨行一致性 | 同键多行某列相同（如同 Title → 同 Sender） |
| S7 | 成对字段联动 | 如 `Receiver` 形态 → `ReceiverType` |
| S8 | Excel 单元格类型 | 须文本的列不能是数值型 |
| S9 | 正则 / 格式串 | 如 Item `{道具id;数量}` |
| S10 | 长度 / 位数约束 | 数字位数、字符串长度 |
| S11 | 时间区间成对与先后 | 两列同时空或同时有值；格式统一；开始≤结束 |

### 语义规则（L*）— 通用类型

| 编号 | 规则 | 说明 / 来源例 |
|------|------|----------------|
| L1 | 别名 / 字号 ↔ 全名同一人物 | 意思一致即可（如 `汉升来信` ↔ `黄忠`） |
| L2 | 正文与身份是否矛盾 | 第一人称与 Sender 等明显冲突才报 |
| L3 | 文案类别与字段含义不矛盾 | 如熔炼返还类 + 合理武将名 |

明确不一致才报 issue；吃不准则「需人工确认」；不因缺外链表跳过。

---

## 表独有业务规则 — 怎么想、怎么写

写新表时，**不要凭空列抽象问题清单**。正确做法是：

1. **先读已有各表 skill 里的独有规则**（见下方「现有样例从哪读」）  
2. **归纳这些规则是怎么被总结出来的**（特性 / 写法）  
3. **带着这些总结方式看新表样例**，写出本表独有条文  
4. 具体条文**只写进新表 skill**，不写进本文件 S*/L*

### 现有样例从哪读

写新表前**必须打开**「现有子 skill」中的各表 `SKILL.md`，重点读：

| 位置 | 当作什么 |
|------|----------|
| 各表 `## 独有规则`（下含结构化 / 语义） | 主样例 |
| 旧 skill 尚未按四段拆开时 | 把超出纯 S*/L* 落地的细则仍视为独有样例 |

每新增一张表的独有节，下一次写别的表时要一并读入——**以仓库里已有单表 skill 为活教材**，而不是只记本文件几条抽象追问。

### 从现有独有规则里应学到的「总结方式」（特性）

对照已读样例，抽出可迁移的**归纳手法**（学写法，不抄某表具体取值）：

| 特性 | 常见总结形态 | 现有痕迹（仅指风格） |
|------|--------------|----------------------|
| 用某一列当业务钥匙 | 「若 A 满足条件 / 匹配模式 → B 必须为…」 | Mail：`Title`→多列；Recharge：`Name`→`Price` |
| 条件写成可执行的匹配 | 精确字符串、前缀/后缀、正则、枚举集合 | Mail 固定 Title；Recharge `^(\d+)元`、`S{n}…` |
| 一条规则管多列 | 同一条件下同时约束多张字段 | Mail 一行 Title 同时订 SenderType/Sender/SendCondType |
| 跨行取值 | 「本行字段 = 另一行（用钥匙定位）的字段」 | Recharge 典藏 Base → 同季豪华 RelateId |
| 同分叉变体 | 同一类表面名，因另一列有值/空或枚举不同拆成多条 | Recharge 典藏升级档 vs 直购档 |
| 跨行一致性 | 同一业务钥匙的多行，某列必须相同 | Mail 同 Title→同 Sender |
| 硬规则进脚本，软规则进语义 | 定值/引用走脚本；需常识判断走 LLM | Mail Title↔武将名语义 |

看新表时：用上表「特性」当检查表，问「本表有没有同类结构」，再对照样例**落成条文**，而不是复述抽象问句。

### 落到新表时怎么干（实用步骤）

1. 列出已读独有规则的特性清单（可自建短表：钥匙列 / 匹配方式 / 约束列 / 是否跨行 / 是否分叉）  
2. 打开新表样例与字段名，**逐条特性对照**：本表有无对应钥匙列、有无「文案锁数值」、有无成套行、有无同钥匙多变体…  
3. 能落到具体条件与期望值的 → 写入新表 `## 独有规则` → `### 结构化规则（脚本）`  
4. 需要语境、无法定死对照表的 → `## 独有规则` → `### 语义规则`  
5. 样例 + 用户说明都推不出 → 独有结构化 / 独有语义均注明「经归纳无」

### 写到哪里

- **具体业务条文** → 只写单表 `## 独有规则` + 脚本  
- **S*/L* 落到本表字段的细则** → 单表 `## 通用规则`  
- **不要**把某表独有细目抄进本文件 S*/L*  
- 仅当出现可跨表复用的**新规则类型**时，才在本文件 S*/L* 新开抽象编号  

### 使用者后续补充（重要）

- Agent 生成单表 skill 时，独有规则往往只是**首轮归纳**（样例 + 已有表写法 + 当时对话），**不保证一次写全**  
- 使用者可在 **`Excel-check/<Table>/SKILL.md` 的 `## 独有规则`**（及 `## 通用规则`）下直接增补条文（结构化 / 语义均可）  
- 增补后若需生效：同步改该表 `scripts/check_*.py`（硬规则），或检查时走语义流程（软规则）  
- 下次写别的表时，应把这些**人手补充过的独有规则**也当作活样例读入  

**补规则前的冲突/重复检查（必须）：**

在具体表 skill 中按用户要求**补充新规则**时，Agent 须先对照该表 `SKILL.md` 已有条文（通用 + 独有，结构化 + 语义）及对应脚本已实现逻辑：

| 情况 | 处理 |
|------|------|
| 与现有规则**实质重复**（同条件同期望，或新条已被现有条文覆盖） | **先反馈**：指出与哪条现有规则重复，**不要落盘/改脚本**；询问是否仍要保留、合并表述，或取消本次补充 |
| 与现有规则**冲突**（同字段/同条件下期望值或判定相反、互斥） | **先反馈**：列出冲突双方（现有 vs 新提），**停止实现**；询问以哪一方为准，或如何改写成一致规则 |
| 无重复且无冲突 | 再按用户意图写入 skill，并视需要改脚本 |

不得在未确认前擅自用新规则覆盖旧规则，或静默忽略冲突只加一条并行规则。

**写进单表 skill（必须）：** 生成或回补任一具体表 `SKILL.md` 时，须按下方「单表 SKILL 骨架」写入同名小节 `## 补充规则时（必须）`（可精简表述，表意不得弱于上表）。不要只留在本公共文件里——否则 Agent 只打开单表 skill 时看不到。已有表缺该节时，写新表或维护旧表时一并补上。

若用户说「给某表补一条独有规则」而未改脚本：先改 skill 条文，再问是否一并改脚本；用户只改 skill 文件也合法，检查时以 skill 为准并由 Agent/脚本对齐。前提是已通过上表冲突/重复确认。

### 自检

- 是否已读过当前仓库内**全部**单表 skill 的 `## 独有规则`？  
- 是否从中提炼了「特性 / 总结方式」，而不是只套 S1–S11？  
- 新表是否按「通用 / 独有」×「结构化 / 语义」四段写齐（空段注明无）？  
- 单表 skill 是否已写入 `## 补充规则时（必须）`？  
- 是否**遵守**公共空 3 行截断 / 空列丢弃，未擅自写成独有「不截断」？若截断后仍有数据，是否已在汇报中提示并等待用户决定？

---

## 脚本 CLI

路径含中文时加引号：

```bash
python "<Table>/scripts/check_<ScriptId>.py" "<xlsx 路径>"
python "<Table>/scripts/check_<ScriptId>.py" "<路径>" --json
python "<Table>/scripts/check_<ScriptId>.py" "<路径>" --semantic-json
```

| 参数 / 退出码 | 含义 |
|---------------|------|
| 默认 | 结构化问题 + 待语义行数 |
| `--json` | 完整 JSON（含 `semantic_rows`） |
| `--semantic-json` | 仅待 LLM 行 |
| `0` / `1` / `2` | 结构化无问题 / 有问题 / 文件不存在等 |

喂 Agent 用 JSON 时退出码可为 `0`（以内容为准）。

Issue 统一格式：`<主键列>=<值> | <展示列>=<值> | <字段名> | <问题说明>`（展示列如 Title/Name，无则省略或用 `-`；主键列按本表实际字段名，常见 `Id` / `*Id` / `name` 等）。

**Agent 汇报硬性要求（避免擅自改格式）：**

- 结构化问题必须按上式**逐条**贴出（与脚本 `format_line` / 终端输出一致）
- **禁止**用分类汇总表、只给数量统计、或改写问题说明来代替 Issue 行
- 问题很多也仍须全文列出；用户明确要求「只摘要」时才可汇总
- 可在 Issue 列表**之前**保留脚本摘要行：检查文件 / 映射表（若有）/ 结构化问题数量 / 待语义行数

语义导出行至少：`id`、展示列、待判断字段；可选 `body_preview` / `category`。

---

## 检查一张表时（Agent）

1. 读 `Excel-check/<Table>/SKILL.md`（含表独有业务规则）
2. 跑脚本（需要时再 `--semantic-json`）
3. 有语义则按单表标准分批分析（建议 ≤50 行/批）
4. 汇报：路径 + **完整 Issue 行列表**（格式见上）+ 语义问题 + 总数；全过则写明已通过。不得用表格/分类统计替代 Issue 行

---

## 如何写一张新表的检查 skill

须按 **1 → 1.5 → 2 → 3** 做完，再落盘 SKILL / 脚本。

### 1. 遍历现有通用规则（S*/L*）

1. **摸清新表结构**：先判定 **表结构 A 或 B**；再确认字段、样例、主键列名；按对应规范落地读表（A：空 3 行截断 + 空列丢弃；B：空 3 行截断，勿套 A 元数据行）  
2. **拉清单**：读本文件全部 S*/L*；读已有单表 skill（[Mail](Mail/SKILL.md)、[Recharge_充值表](Recharge_充值表/SKILL.md) 等）作范例  
3. **双向打标**（规则↔字段；一字段可多条）  
4. **按本表改写**实现；同字段多规则并列报错  
5. 若空 3 行后仍有行：按上文「公共读表约定不可擅自例外」**截断 + 汇报提示**，不得先改成不截断

### 1.5. 归纳表独有业务规则（必做）

按上文「表独有业务规则 — 怎么想、怎么写」执行：

1. 读现有各表 skill 的 `## 独有规则`  
2. 总结其**特性与写法**  
3. 用同一套写法对照新表样例，写入本表 `## 独有规则`（结构化 / 语义）并实现  
4. 具体条文不写进本文件；通用 S*/L* 落地写在 `## 通用规则`  

若独有推不出 → 独有结构化 / 独有语义注明「经归纳无」。

### 2. 回补公共总结（仅新类型）

仅当出现**可复用且总结表没有的规则类型**时，才在本文件 S*/L* 新开编号（抽象定义 + 一两句说明）。  
表内具体业务条目**不属于**回补内容。

### 3. 向用户汇报（必须）

```
## 本表规则清单：<Table>
### 通用规则
- 结构化：…
- 语义：… / 无
### 独有规则
- 结构化：…
- 语义：… / 无
### 公共类型回补
- 无 / 新开了 Sx
### 读表约定
- 表结构：A（经典行表） / B（Global 系）
- 主键列：…
- 空 3 行截断：已遵守 / 用户已确认本表例外为…
- 截断后未检行：无 / 约 N 行（例：主键=…）
### 产物
- …
```

### 单表 SKILL 骨架

```markdown
---
name: <Table>
description: |
  校验名将杀 <Table>.xlsx。
  当用户提到 <Table>.xlsx、<中文名>配置检查时使用。
  公共流程见 Excel-check/SKILL.md；本文件仅含本表规则。
---

# <Table> — <中文名>配置表检查

校验 `<Table>.xlsx`。公共约定见 [Excel-check/SKILL.md](../SKILL.md)。
**表结构**：A（经典行表） / B（Global 系）
**主键列**：`<PkField>`
**运行时输入**：用户提供 xlsx 路径。

## 脚本

\`\`\`bash
python "<Table>/scripts/check_<ScriptId>.py" "<路径>"
python "<Table>/scripts/check_<ScriptId>.py" "<路径>" --semantic-json
\`\`\`

## 通用规则

### 结构化规则（脚本）

（S*/公共约定落到本表字段；可附适用编号对照）

### 语义规则

（通用 L* 落到本表；无则写「无」）

## 独有规则

### 结构化规则（脚本）

（本表业务硬规则；经归纳无则注明。生成后可由使用者在此继续补充）

### 语义规则

（本表业务软规则；经归纳无则注明。生成后可由使用者在此继续补充）

## 补充规则时（必须）

按用户要求为本表 **新增/修改规则**前：先对照本文件已有「通用规则」「独有规则」（及对应脚本实现）。

| 情况 | 处理 |
|------|------|
| 与现有规则实质重复 | 先反馈重复点，勿落盘；询问是否保留/合并/取消 |
| 与现有规则冲突 | 先列出冲突双方，停止实现；询问以哪方为准 |
| 无重复且无冲突 | 再写入本文件，并视需要改脚本 |

细则见 [Excel-check/SKILL.md](../SKILL.md)「使用者后续补充」。

## 工作流程
…
```

单表只写本表四段细则与脚本路径、以及上节「补充规则时」；勿重复公共依赖/退出码教科书。

---

## 现有子 skill

共 **204** 张（均为表结构 **A** 经典行表基线或已人工补全）。其中 4 张已人工补全独有/语义规则；其余为批量基线（通用 S1/S2/S9/S11，主键列按本表实际字段；独有规则经归纳暂无，可后续增补）。

尚未覆盖的 **B（Global 系）** 等：`Global_*`、`BattleGlobal_*`、`GuildCityName_*`，以及特殊布局 `PartnerText_伙伴聊天文本`。

| Skill | 表 | 说明 |
|-------|-----|------|
| [AIPriorityChange_人机出牌优先级变更表](AIPriorityChange_人机出牌优先级变更表/SKILL.md) | `AIPriorityChange_人机出牌优先级变更表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [AIRobotInfo_AI伪装表](AIRobotInfo_AI伪装表/SKILL.md) | `AIRobotInfo_AI伪装表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [AIcombatsituation_人机战斗情况表](AIcombatsituation_人机战斗情况表/SKILL.md) | `AIcombatsituation_人机战斗情况表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [AccumulatedRecharge_累计充值奖励表](AccumulatedRecharge_累计充值奖励表/SKILL.md) | `AccumulatedRecharge_累计充值奖励表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [Achieve](Achieve/SKILL.md) | `Achieve.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [AchieveHero](AchieveHero/SKILL.md) | `AchieveHero.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [ActivityInvite_活动邀请表](ActivityInvite_活动邀请表/SKILL.md) | `ActivityInvite_活动邀请表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [ActivityKey_活动密钥表](ActivityKey_活动密钥表/SKILL.md) | `ActivityKey_活动密钥表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [ActivityLogin_登录活动表](ActivityLogin_登录活动表/SKILL.md) | `ActivityLogin_登录活动表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [ActivityTask_活动任务表](ActivityTask_活动任务表/SKILL.md) | `ActivityTask_活动任务表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [Activity_活动表](Activity_活动表/SKILL.md) | `Activity_活动表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [Animation配置表](Animation配置表/SKILL.md) | `Animation配置表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [ArenaScore](ArenaScore/SKILL.md) | `ArenaScore.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [ArenaScoreRewards](ArenaScoreRewards/SKILL.md) | `ArenaScoreRewards.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [ArenaSeason](ArenaSeason/SKILL.md) | `ArenaSeason.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [ArenaSeasonLimitedHero_赛季限定武将表](ArenaSeasonLimitedHero_赛季限定武将表/SKILL.md) | `ArenaSeasonLimitedHero_赛季限定武将表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [AudioBnk_音频配置bnk](AudioBnk_音频配置bnk/SKILL.md) | `AudioBnk_音频配置bnk.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [Audio_音频配置表](Audio_音频配置表/SKILL.md) | `Audio_音频配置表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [BattleTipText_战斗提示文本表](BattleTipText_战斗提示文本表/SKILL.md) | `BattleTipText_战斗提示文本表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [Bgm](Bgm/SKILL.md) | `Bgm.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [BossReward年兽挑战伤害奖励表](BossReward年兽挑战伤害奖励表/SKILL.md) | `BossReward年兽挑战伤害奖励表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [BountyTaskRule_悬赏任务规则表](BountyTaskRule_悬赏任务规则表/SKILL.md) | `BountyTaskRule_悬赏任务规则表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [BubbleText_台词气泡表](BubbleText_台词气泡表/SKILL.md) | `BubbleText_台词气泡表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [Buff](Buff/SKILL.md) | `Buff.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [Card](Card/SKILL.md) | `Card.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [CardLineIcon_手牌横图](CardLineIcon_手牌横图/SKILL.md) | `CardLineIcon_手牌横图.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [CardSkin_手牌皮肤表](CardSkin_手牌皮肤表/SKILL.md) | `CardSkin_手牌皮肤表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [CardUI_手牌表现表](CardUI_手牌表现表/SKILL.md) | `CardUI_手牌表现表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [ChallengeSeat_新手挑战v2座位配置](ChallengeSeat_新手挑战v2座位配置/SKILL.md) | `ChallengeSeat_新手挑战v2座位配置.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [ChannelCustom_渠道定制](ChannelCustom_渠道定制/SKILL.md) | `ChannelCustom_渠道定制.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [Chat](Chat/SKILL.md) | `Chat.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [ChatCodeGift](ChatCodeGift/SKILL.md) | `ChatCodeGift.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [CityWar](CityWar/SKILL.md) | `CityWar.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [CityWarSeason_兵临城下赛季信息](CityWarSeason_兵临城下赛季信息/SKILL.md) | `CityWarSeason_兵临城下赛季信息.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [CollectionLevel_收藏品阶等级表](CollectionLevel_收藏品阶等级表/SKILL.md) | `CollectionLevel_收藏品阶等级表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [CollectionList_收藏品阶清单表](CollectionList_收藏品阶清单表/SKILL.md) | `CollectionList_收藏品阶清单表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [Country_势力表](Country_势力表/SKILL.md) | `Country_势力表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [DailyAsk_每日问对](DailyAsk_每日问对/SKILL.md) | `DailyAsk_每日问对.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [DeviceLevel_设备分级表](DeviceLevel_设备分级表/SKILL.md) | `DeviceLevel_设备分级表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [DragonBoat_龙舟](DragonBoat_龙舟/SKILL.md) | `DragonBoat_龙舟.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [DragonboatLZJD_龙舟竞渡](DragonboatLZJD_龙舟竞渡/SKILL.md) | `DragonboatLZJD_龙舟竞渡.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [Draw](Draw/SKILL.md) | `Draw.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [DrawHJZP_黄金转盘](DrawHJZP_黄金转盘/SKILL.md) | `DrawHJZP_黄金转盘.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [DrawPet_结缘亭](DrawPet_结缘亭/SKILL.md) | `DrawPet_结缘亭.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [Drop](Drop/SKILL.md) | `Drop.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [EmojiTab_表情页签表](EmojiTab_表情页签表/SKILL.md) | `EmojiTab_表情页签表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [Emoji_表情配置表](Emoji_表情配置表/SKILL.md) | `Emoji_表情配置表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [ErrorCode](ErrorCode/SKILL.md) | `ErrorCode.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [FirstRecharge_首充表](FirstRecharge_首充表/SKILL.md) | `FirstRecharge_首充表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [Fx配置表](Fx配置表/SKILL.md) | `Fx配置表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [GiftItem_送礼配置](GiftItem_送礼配置/SKILL.md) | `GiftItem_送礼配置.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [GiftLimitShop_限时礼包商品](GiftLimitShop_限时礼包商品/SKILL.md) | `GiftLimitShop_限时礼包商品.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [GiftLimitShop_限时礼包商店](GiftLimitShop_限时礼包商店/SKILL.md) | `GiftLimitShop_限时礼包商店.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [GiftShop_礼包商品](GiftShop_礼包商品/SKILL.md) | `GiftShop_礼包商品.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [GiftShop_礼包商店](GiftShop_礼包商店/SKILL.md) | `GiftShop_礼包商店.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [Guide](Guide/SKILL.md) | `Guide.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [Guide2v2_新手2v2竞技](Guide2v2_新手2v2竞技/SKILL.md) | `Guide2v2_新手2v2竞技.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [GuildCityCreate](GuildCityCreate/SKILL.md) | `GuildCityCreate.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [GuildCityLevel](GuildCityLevel/SKILL.md) | `GuildCityLevel.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [GuildCityWarReward](GuildCityWarReward/SKILL.md) | `GuildCityWarReward.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [GuildCityWeekRankAwards_山河周榜奖励表](GuildCityWeekRankAwards_山河周榜奖励表/SKILL.md) | `GuildCityWeekRankAwards_山河周榜奖励表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [GuildEvent_公会事件表](GuildEvent_公会事件表/SKILL.md) | `GuildEvent_公会事件表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [GuildRedPack_公会红包表](GuildRedPack_公会红包表/SKILL.md) | `GuildRedPack_公会红包表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [GuildWeekRankAwards_公会周榜奖励表](GuildWeekRankAwards_公会周榜奖励表/SKILL.md) | `GuildWeekRankAwards_公会周榜奖励表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [Hero](Hero/SKILL.md) | `Hero.xlsx` | 结构化 + Name→拼音/国号/熔炼名/扩展包/性别语义 |
| [HeroBond_武将羁绊](HeroBond_武将羁绊/SKILL.md) | `HeroBond_武将羁绊.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [HeroExpansionPack_武将扩展包配置表](HeroExpansionPack_武将扩展包配置表/SKILL.md) | `HeroExpansionPack_武将扩展包配置表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [HeroLines_武将台词表](HeroLines_武将台词表/SKILL.md) | `HeroLines_武将台词表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [HeroRaise_武将养成表](HeroRaise_武将养成表/SKILL.md) | `HeroRaise_武将养成表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [HeroRetreat_武将归隐表](HeroRetreat_武将归隐表/SKILL.md) | `HeroRetreat_武将归隐表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [HeroSkillSkinAnimation_英雄皮肤技能动画表](HeroSkillSkinAnimation_英雄皮肤技能动画表/SKILL.md) | `HeroSkillSkinAnimation_英雄皮肤技能动画表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [HeroSkillSkinFx_英雄皮肤技能特效](HeroSkillSkinFx_英雄皮肤技能特效/SKILL.md) | `HeroSkillSkinFx_英雄皮肤技能特效.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [HeroSkinCollition_英雄皮肤收藏](HeroSkinCollition_英雄皮肤收藏/SKILL.md) | `HeroSkinCollition_英雄皮肤收藏.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [HeroSkinItem_英雄皮肤](HeroSkinItem_英雄皮肤/SKILL.md) | `HeroSkinItem_英雄皮肤.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [HeroSkinSpine_英雄皮肤Spine](HeroSkinSpine_英雄皮肤Spine/SKILL.md) | `HeroSkinSpine_英雄皮肤Spine.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [HeroUI_武将表现表](HeroUI_武将表现表/SKILL.md) | `HeroUI_武将表现表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [HeroWar_群雄逐鹿](HeroWar_群雄逐鹿/SKILL.md) | `HeroWar_群雄逐鹿.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [HotfixFuncs](HotfixFuncs/SKILL.md) | `HotfixFuncs.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [HundredDraw_百连招募表](HundredDraw_百连招募表/SKILL.md) | `HundredDraw_百连招募表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [IdentityDesignCompetition_身份设计大赛](IdentityDesignCompetition_身份设计大赛/SKILL.md) | `IdentityDesignCompetition_身份设计大赛.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [IdentityEncodeRule_身份编码规则表](IdentityEncodeRule_身份编码规则表/SKILL.md) | `IdentityEncodeRule_身份编码规则表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [IdentityPaintingTogether_丹青共绘设计](IdentityPaintingTogether_丹青共绘设计/SKILL.md) | `IdentityPaintingTogether_丹青共绘设计.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [IdentityRule](IdentityRule/SKILL.md) | `IdentityRule.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [IdentityRuleDetailed](IdentityRuleDetailed/SKILL.md) | `IdentityRuleDetailed.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [InputFieldForbidden](InputFieldForbidden/SKILL.md) | `InputFieldForbidden.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [Item](Item/SKILL.md) | `Item.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [ItemDesktop_牌局桌面](ItemDesktop_牌局桌面/SKILL.md) | `ItemDesktop_牌局桌面.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [ItemFrame_边框道具表](ItemFrame_边框道具表/SKILL.md) | `ItemFrame_边框道具表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [ItemHeroSkin_武将皮肤展示表](ItemHeroSkin_武将皮肤展示表/SKILL.md) | `ItemHeroSkin_武将皮肤展示表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [ItemInteractionSuite_牌桌章法](ItemInteractionSuite_牌桌章法/SKILL.md) | `ItemInteractionSuite_牌桌章法.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [ItemTimeLimitResolve_道具限时分解表](ItemTimeLimitResolve_道具限时分解表/SKILL.md) | `ItemTimeLimitResolve_道具限时分解表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [KaKaShop_良心小铺](KaKaShop_良心小铺/SKILL.md) | `KaKaShop_良心小铺.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [KeyWord_关键字配置表](KeyWord_关键字配置表/SKILL.md) | `KeyWord_关键字配置表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [LimitTimeSkinDrawReward_限时皮肤抽卡奖励配置表](LimitTimeSkinDrawReward_限时皮肤抽卡奖励配置表/SKILL.md) | `LimitTimeSkinDrawReward_限时皮肤抽卡奖励配置表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [LimitTimeSkinRechargeReward_限时皮肤充值奖励配置表](LimitTimeSkinRechargeReward_限时皮肤充值奖励配置表/SKILL.md) | `LimitTimeSkinRechargeReward_限时皮肤充值奖励配置表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [LimitTimeSkinRewardPrew_限时皮肤奖励一览表](LimitTimeSkinRewardPrew_限时皮肤奖励一览表/SKILL.md) | `LimitTimeSkinRewardPrew_限时皮肤奖励一览表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [LoadingBg_加载背景表](LoadingBg_加载背景表/SKILL.md) | `LoadingBg_加载背景表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [Mail](Mail/SKILL.md) | `Mail.xlsx` | 结构化 + Title/Sender 人物语义 |
| [MainTheme_主界面主题表](MainTheme_主界面主题表/SKILL.md) | `MainTheme_主界面主题表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [MeltVoice_熔炼语音表](MeltVoice_熔炼语音表/SKILL.md) | `MeltVoice_熔炼语音表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [ModeTitle_模式标题](ModeTitle_模式标题/SKILL.md) | `ModeTitle_模式标题.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [Module_模块下载奖励表](Module_模块下载奖励表/SKILL.md) | `Module_模块下载奖励表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [MonthlyCard_月卡表](MonthlyCard_月卡表/SKILL.md) | `MonthlyCard_月卡表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [NewbieChallenge_新手挑战v2](NewbieChallenge_新手挑战v2/SKILL.md) | `NewbieChallenge_新手挑战v2.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [NewbieInvite_新人入队](NewbieInvite_新人入队/SKILL.md) | `NewbieInvite_新人入队.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [NianShouLaiXi_年兽来袭](NianShouLaiXi_年兽来袭/SKILL.md) | `NianShouLaiXi_年兽来袭.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [OpenSysControl_功能开关表](OpenSysControl_功能开关表/SKILL.md) | `OpenSysControl_功能开关表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [PackageSignature_包签名](PackageSignature_包签名/SKILL.md) | `PackageSignature_包签名.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [PartnerTrainSeat](PartnerTrainSeat/SKILL.md) | `PartnerTrainSeat.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [PeakArena_巅峰竞技](PeakArena_巅峰竞技/SKILL.md) | `PeakArena_巅峰竞技.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [PeakRankReward_巅峰竞技排名奖励](PeakRankReward_巅峰竞技排名奖励/SKILL.md) | `PeakRankReward_巅峰竞技排名奖励.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [PeakWinsReward_巅峰竞技胜场奖励](PeakWinsReward_巅峰竞技胜场奖励/SKILL.md) | `PeakWinsReward_巅峰竞技胜场奖励.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [PetAudio_灵宠音效表](PetAudio_灵宠音效表/SKILL.md) | `PetAudio_灵宠音效表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [PetDialogue_灵宠对话表](PetDialogue_灵宠对话表/SKILL.md) | `PetDialogue_灵宠对话表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [PetEasterEggIdle_灵宠彩蛋待机](PetEasterEggIdle_灵宠彩蛋待机/SKILL.md) | `PetEasterEggIdle_灵宠彩蛋待机.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [PetFeedbackText_伙伴反馈](PetFeedbackText_伙伴反馈/SKILL.md) | `PetFeedbackText_伙伴反馈.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [PetInstruction_伙伴指令表](PetInstruction_伙伴指令表/SKILL.md) | `PetInstruction_伙伴指令表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [PetRaise_灵宠养成表](PetRaise_灵宠养成表/SKILL.md) | `PetRaise_灵宠养成表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [PetTriggerWeight_灵宠动画触发权重](PetTriggerWeight_灵宠动画触发权重/SKILL.md) | `PetTriggerWeight_灵宠动画触发权重.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [Pet_灵宠表](Pet_灵宠表/SKILL.md) | `Pet_灵宠表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [PioneerPlanData_先锋计划返利数据表](PioneerPlanData_先锋计划返利数据表/SKILL.md) | `PioneerPlanData_先锋计划返利数据表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [PioneerPlan_先锋计划返利表](PioneerPlan_先锋计划返利表/SKILL.md) | `PioneerPlan_先锋计划返利表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [PveBaWangZhiLuanStage_八王之乱关卡表](PveBaWangZhiLuanStage_八王之乱关卡表/SKILL.md) | `PveBaWangZhiLuanStage_八王之乱关卡表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [PveBossStage_龚行天罚](PveBossStage_龚行天罚/SKILL.md) | `PveBossStage_龚行天罚.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [PveLevel_千里单骑难度表](PveLevel_千里单骑难度表/SKILL.md) | `PveLevel_千里单骑难度表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [PveNewbieChallenge_新手挑战](PveNewbieChallenge_新手挑战/SKILL.md) | `PveNewbieChallenge_新手挑战.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [PveRougeBattleNode_千里单骑战斗节点配置](PveRougeBattleNode_千里单骑战斗节点配置/SKILL.md) | `PveRougeBattleNode_千里单骑战斗节点配置.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [PveRougeEnemyBuff_千里单骑敌方武将增益配置](PveRougeEnemyBuff_千里单骑敌方武将增益配置/SKILL.md) | `PveRougeEnemyBuff_千里单骑敌方武将增益配置.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [PveRougeEventNode_千里单骑事件节点配置](PveRougeEventNode_千里单骑事件节点配置/SKILL.md) | `PveRougeEventNode_千里单骑事件节点配置.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [PveRougeLayerBonus_千里单骑彩蛋敌人配置](PveRougeLayerBonus_千里单骑彩蛋敌人配置/SKILL.md) | `PveRougeLayerBonus_千里单骑彩蛋敌人配置.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [PveRougeNodeBuffText_千里单骑节点加成文本配置](PveRougeNodeBuffText_千里单骑节点加成文本配置/SKILL.md) | `PveRougeNodeBuffText_千里单骑节点加成文本配置.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [PveRougeNodeBuff_千里单骑节点buff](PveRougeNodeBuff_千里单骑节点buff/SKILL.md) | `PveRougeNodeBuff_千里单骑节点buff.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [PveRougeNodeEnd千里单骑结局表](PveRougeNodeEnd千里单骑结局表/SKILL.md) | `PveRougeNodeEnd千里单骑结局表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [PveRougeRestNode_千里单骑休息节点配置](PveRougeRestNode_千里单骑休息节点配置/SKILL.md) | `PveRougeRestNode_千里单骑休息节点配置.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [PveRougeReward_千里单骑奖励表](PveRougeReward_千里单骑奖励表/SKILL.md) | `PveRougeReward_千里单骑奖励表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [PveSeatGroup_Pve座位配置](PveSeatGroup_Pve座位配置/SKILL.md) | `PveSeatGroup_Pve座位配置.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [PveTrainSeat](PveTrainSeat/SKILL.md) | `PveTrainSeat.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [QYCollection_群英荟萃表](QYCollection_群英荟萃表/SKILL.md) | `QYCollection_群英荟萃表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [QuickChat_快速聊天表](QuickChat_快速聊天表/SKILL.md) | `QuickChat_快速聊天表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [Rank_排行榜配置](Rank_排行榜配置/SKILL.md) | `Rank_排行榜配置.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [RatePseudo_伪概率](RatePseudo_伪概率/SKILL.md) | `RatePseudo_伪概率.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [RechargeChannel_充值渠道映射表](RechargeChannel_充值渠道映射表/SKILL.md) | `RechargeChannel_充值渠道映射表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [Recharge_充值表](Recharge_充值表/SKILL.md) | `Recharge_充值表.xlsx` | 仅结构化；含表独有业务规则 |
| [Recharge_映射表](Recharge_映射表/SKILL.md) | `Recharge_映射表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [RecommendBd推荐加点表](RecommendBd推荐加点表/SKILL.md) | `RecommendBd推荐加点表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [Recommend_推荐规则表](Recommend_推荐规则表/SKILL.md) | `Recommend_推荐规则表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [ReportRule_举报规则表](ReportRule_举报规则表/SKILL.md) | `ReportRule_举报规则表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [RobotAction_人机行动表](RobotAction_人机行动表/SKILL.md) | `RobotAction_人机行动表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [RoomMode](RoomMode/SKILL.md) | `RoomMode.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [RoomModeReward](RoomModeReward/SKILL.md) | `RoomModeReward.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [RougeEndlessAction_千里单骑无尽动作表](RougeEndlessAction_千里单骑无尽动作表/SKILL.md) | `RougeEndlessAction_千里单骑无尽动作表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [RougeEndlessBattleNode_千里单骑无尽战斗节点](RougeEndlessBattleNode_千里单骑无尽战斗节点/SKILL.md) | `RougeEndlessBattleNode_千里单骑无尽战斗节点.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [RougeEndlessBg_千里单骑背景表](RougeEndlessBg_千里单骑背景表/SKILL.md) | `RougeEndlessBg_千里单骑背景表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [RougeEndlessBonusLines_千里单骑无尽彩蛋台词表](RougeEndlessBonusLines_千里单骑无尽彩蛋台词表/SKILL.md) | `RougeEndlessBonusLines_千里单骑无尽彩蛋台词表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [RougeEndlessCalc_千里单骑无尽计算器表](RougeEndlessCalc_千里单骑无尽计算器表/SKILL.md) | `RougeEndlessCalc_千里单骑无尽计算器表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [RougeEndlessCollectionAction_千里单骑无尽图鉴人物动作表](RougeEndlessCollectionAction_千里单骑无尽图鉴人物动作表/SKILL.md) | `RougeEndlessCollectionAction_千里单骑无尽图鉴人物动作表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [RougeEndlessEventRestNode_千里单骑无尽事件及休息节点表](RougeEndlessEventRestNode_千里单骑无尽事件及休息节点表/SKILL.md) | `RougeEndlessEventRestNode_千里单骑无尽事件及休息节点表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [RougeEndlessHeroBuff_千里单骑无尽武将加成对应表](RougeEndlessHeroBuff_千里单骑无尽武将加成对应表/SKILL.md) | `RougeEndlessHeroBuff_千里单骑无尽武将加成对应表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [RougeEndlessKaKaNode_千里单骑无尽卡卡节点表](RougeEndlessKaKaNode_千里单骑无尽卡卡节点表/SKILL.md) | `RougeEndlessKaKaNode_千里单骑无尽卡卡节点表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [RougeEndlessNodeBuff_千里单骑无尽节点buff效果表](RougeEndlessNodeBuff_千里单骑无尽节点buff效果表/SKILL.md) | `RougeEndlessNodeBuff_千里单骑无尽节点buff效果表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [RougeEndlessRankReward_千里单骑无尽排行奖励表](RougeEndlessRankReward_千里单骑无尽排行奖励表/SKILL.md) | `RougeEndlessRankReward_千里单骑无尽排行奖励表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [RougeEndlessSanLeiNode_千里单骑无尽三类节点表](RougeEndlessSanLeiNode_千里单骑无尽三类节点表/SKILL.md) | `RougeEndlessSanLeiNode_千里单骑无尽三类节点表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [RougeEndlessSeason_千里单骑无尽赛季表](RougeEndlessSeason_千里单骑无尽赛季表/SKILL.md) | `RougeEndlessSeason_千里单骑无尽赛季表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [RougeEndlessValidator_千里单骑无尽条件验证表](RougeEndlessValidator_千里单骑无尽条件验证表/SKILL.md) | `RougeEndlessValidator_千里单骑无尽条件验证表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [RougeEndlessWuJiangPai_千里单骑无尽武将牌表](RougeEndlessWuJiangPai_千里单骑无尽武将牌表/SKILL.md) | `RougeEndlessWuJiangPai_千里单骑无尽武将牌表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [RougeEndlessXinWu_千里单骑无尽信物表](RougeEndlessXinWu_千里单骑无尽信物表/SKILL.md) | `RougeEndlessXinWu_千里单骑无尽信物表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [RougeEndlessZhuGeLiangNode_千里单骑无尽诸葛亮节点表](RougeEndlessZhuGeLiangNode_千里单骑无尽诸葛亮节点表/SKILL.md) | `RougeEndlessZhuGeLiangNode_千里单骑无尽诸葛亮节点表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [RushHeroRank_开阁纳贤](RushHeroRank_开阁纳贤/SKILL.md) | `RushHeroRank_开阁纳贤.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [SeasonPassBag_赛季战令礼包表](SeasonPassBag_赛季战令礼包表/SKILL.md) | `SeasonPassBag_赛季战令礼包表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [SeasonPassReward_赛季战令奖励表](SeasonPassReward_赛季战令奖励表/SKILL.md) | `SeasonPassReward_赛季战令奖励表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [SeasonPassTask_赛季战令任务](SeasonPassTask_赛季战令任务/SKILL.md) | `SeasonPassTask_赛季战令任务.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [SeasonPass_赛季战令表](SeasonPass_赛季战令表/SKILL.md) | `SeasonPass_赛季战令表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [SecPwdRule_二级密码规则](SecPwdRule_二级密码规则/SKILL.md) | `SecPwdRule_二级密码规则.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [ShiRiYuXi_十日羽檄配置表](ShiRiYuXi_十日羽檄配置表/SKILL.md) | `ShiRiYuXi_十日羽檄配置表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [ShopGoods_商品表](ShopGoods_商品表/SKILL.md) | `ShopGoods_商品表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [Shop_商店表](Shop_商店表/SKILL.md) | `Shop_商店表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [SignIn](SignIn/SKILL.md) | `SignIn.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [Skill](Skill/SKILL.md) | `Skill.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [SkillCardSort](SkillCardSort/SKILL.md) | `SkillCardSort.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [SkillEnchant_技能强化](SkillEnchant_技能强化/SKILL.md) | `SkillEnchant_技能强化.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [SkillLines_技能台词表](SkillLines_技能台词表/SKILL.md) | `SkillLines_技能台词表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [SkillMelt_技能熔炼表](SkillMelt_技能熔炼表/SKILL.md) | `SkillMelt_技能熔炼表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [SkillTag_技能标签表](SkillTag_技能标签表/SKILL.md) | `SkillTag_技能标签表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [SkillUI_技能表现表](SkillUI_技能表现表/SKILL.md) | `SkillUI_技能表现表.xlsx` | 结构化 + SkillName↔Id / 文案语义；外联 Skill.xlsx |
| [SpecialHero_定制武将表](SpecialHero_定制武将表/SKILL.md) | `SpecialHero_定制武将表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [SpecialStack_定制牌堆表](SpecialStack_定制牌堆表/SKILL.md) | `SpecialStack_定制牌堆表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [Survey_腾讯问卷表](Survey_腾讯问卷表/SKILL.md) | `Survey_腾讯问卷表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [SystemSet_系统设置表](SystemSet_系统设置表/SKILL.md) | `SystemSet_系统设置表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [Task](Task/SKILL.md) | `Task.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [TaskCompleteCond](TaskCompleteCond/SKILL.md) | `TaskCompleteCond.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [Team](Team/SKILL.md) | `Team.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [ThousandPics_千机锦图](ThousandPics_千机锦图/SKILL.md) | `ThousandPics_千机锦图.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [TitleByRank_排行称号表](TitleByRank_排行称号表/SKILL.md) | `TitleByRank_排行称号表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [Title_称号表](Title_称号表/SKILL.md) | `Title_称号表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [TongPao_同袍共战](TongPao_同袍共战/SKILL.md) | `TongPao_同袍共战.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [TotalBattleMerit_累计战功](TotalBattleMerit_累计战功/SKILL.md) | `TotalBattleMerit_累计战功.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [UI文本配置表](UI文本配置表/SKILL.md) | `UI文本配置表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [UserExpLevel](UserExpLevel/SKILL.md) | `UserExpLevel.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [UserPrestige_玩家声望表](UserPrestige_玩家声望表/SKILL.md) | `UserPrestige_玩家声望表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [Validator](Validator/SKILL.md) | `Validator.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [WechatTask_小程序奖励表](WechatTask_小程序奖励表/SKILL.md) | `WechatTask_小程序奖励表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [WuHouJiangWu_武侯讲武](WuHouJiangWu_武侯讲武/SKILL.md) | `WuHouJiangWu_武侯讲武.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [YiBiQianJin_一笔千金](YiBiQianJin_一笔千金/SKILL.md) | `YiBiQianJin_一笔千金.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
| [服务器配置表](服务器配置表/SKILL.md) | `服务器配置表.xlsx` | 基线结构化（表结构 A；S1/S2/S9/S11）；独有待补 |
