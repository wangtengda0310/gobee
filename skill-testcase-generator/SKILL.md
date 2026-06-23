---
name: skill-testcase-generator
description: |
  为名将杀游戏的**技能或牌**自动生成功能测试用例。
  触发条件（同时满足两点即触发）：
  1. 消息中包含**对象名称**——技能名（如：武圣义绝、天不负我）或牌名（如：削弱冲杀、过河拆桥）
  2. 消息中包含意图关键词：测试 / 用例 / 测 / 生成
  即使用户表达较口语化也应触发，例如：「生成 XXX 的测试用例」「测一下 XXX」「XXX 测试」「给 XXX 出一份测例」「XXX 用例」。
  流程：定位对象（先技能 sheet 找、未命中再牌 sheet 找，定 `OBJECT_TYPE=skill|card`）→ 反向查表解析文案引用 → 按对象路径拆解（技能→5 子模块见 references/skill-decomposition.md；牌→一级模块清单见 references/card-decomposition.md）→ 生成用例 → 完整性校验 → 输出 Excel + Markdown 到飞书 Drive。
---

# 名将杀技能/牌测试用例生成器

输入一个技能名或牌名，自动收集背景知识并生成完整的功能测试用例。仅适用于名将杀项目。

## 默认配置

- **项目**: 名将杀
- **源文档 URL**: `https://ztgame.feishu.cn/sheets/Z9kFs9JWdhqxQ5tt0I9csmytnVg`
- **源文档 Token**: `Z9kFs9JWdhqxQ5tt0I9csmytnVg`
- **Python 环境**: `/usr/bin/python3`（已含 openpyxl）
- **本地输出路径**: `/Users/zt-3803045/.openclaw/workspace/名将杀技能测试用例/`
- **Drive 上传位置**: 云空间根目录
- **Drive 权限**: 所有者保持 bot 不变；上传后给固定人员 + 发起用户授「可管理」、群聊则整群授「可读」、两份文件开启组织内链接分享（详见 4.3–4.5）

## 触发方式

消息中同时满足两点即启动全流程：
1. 包含**对象名称**（技能名 或 牌名）
2. 包含意图关键词：`测试` / `用例` / `测` / `生成`

示例：「生成武圣义绝的技能测试用例」「测一下天不负我」「削弱冲杀测试用例」「过河拆桥 测试」。

提取对象名后做**对象判定**（决定后续走技能路径还是牌路径）：

1. 先去「技能」sheet（`iwM7X5`）「技能名称」列做精确/模糊匹配 → 命中即 `OBJECT_TYPE=skill`。
2. 未命中再去「牌」sheet（`a109ea`）「牌名称」列匹配 → 命中即 `OBJECT_TYPE=card`。
3. 两 sheet 均命中（罕见同名）→ 列两组候选问用户。
4. 两 sheet 均零候选 → 按 1.1 零候选分支处理。

`OBJECT_TYPE` 决定后续 1.1 / 1.3 / 1.4 / 3.1 / 3.2 / 3.3 / 3.5 各阶段走哪条分支。

## 阶段零：初始化工作目录

在所有阶段之前创建任务专属子目录，避免同名对象覆盖历史结果：

```bash
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
SKILL_DIR="/Users/zt-3803045/.openclaw/workspace/名将杀技能测试用例/<对象名>_${TIMESTAMP}"
mkdir -p "${SKILL_DIR}"
```

后续所有产物（overview.json、testcases.json、Excel、Markdown）全部写入 `${SKILL_DIR}`。并行生成多个对象时各子代理各自取 `TIMESTAMP`，天然隔离。

> ⚠️ 若 1.1 名称解析为零候选（两 sheet 都找不到），**不创建 `${SKILL_DIR}`**，避免空目录污染。

## 四阶段工作流

### 阶段一：数据采集（5 子步）

#### 1.0 快照校验/刷新（前置门控 — 必须先完成再进入 1.1）

进入定位与检索之前的**硬门控**。`+info` 拿结构与合并范围（检索全程复用），再一次小 `+read`（如读 A1:A1）拿**文档级 revision**，与快照头部记录比对：

- **一致** → 所有快照有效，直接进入 1.1。
- **不一致** → **先刷新内容、后写 revision**：刷新 [sheet-structure.md](references/sheet-structure.md)、[skill-test-point-modules.md](references/skill-test-point-modules.md)、[card-test-point-modules.md](references/card-test-point-modules.md)、[test-point-knowledge.md](references/test-point-knowledge.md) 中实际变化的内容，再把四处快照头部 `revision` 改为当前值。**严禁只改 revision 不核内容**（否则下一轮会误判快照新鲜、沿用旧数据）。刷新是本步的强制产出，无需向用户确认；**revision 未写回前不得进入 1.1**。
- **取不到** → 重试一次；仍失败按 fail-closed 分支如实告知用户，不静默沿用旧快照。

刷新哪些快照、完整校验链见 [snapshot-validation.md](references/snapshot-validation.md)。

> ⚠️ 为何独立成前置门控：快照刷新是**条件触发的写盘维护**，与主线交付（定位→拆解→生成）无直接关系。若与「定位技能」同处一步，注意力会被主线带走——刷新被口头宣告却跳过（已两次复现）。单列为先于定位的硬门控，正是为防这种漏刷。

> ⚠️ 索引转换：`+info` 返回的 `start_row_index`/`end_row_index` 是 **0-based**，而 `+read`/`+find` 的 range 用 **Excel 行号（1-based）**。拿到 merges 后立即对所有索引 **+1**，全程使用 1-based。

#### 1.1 定位对象

按"对象判定"已定 `OBJECT_TYPE`，分支处理：

**技能路径（OBJECT_TYPE=skill）**

在「技能」sheet（`iwM7X5`）「技能名称」列搜索技能名：

```bash
lark-cli sheets +find --find "<对象名>" --url "..." --sheet-id "iwM7X5"
```

命中后读该 cell 的 **C 列合并范围 = 该技能全部条目行**（单条目则范围为自身一行），逐行取 C/D/E/F 列得到条目列表 `[{发动方式, 文案}, …]`；武将/势力取包住这些行的 A/B 列合并范围值（嵌套合并解析规则见 `sheet-structure.md`）。

| 匹配情况 | 处理 |
|---------|------|
| 精确命中唯一技能 | 取全条目，继续主流程 |
| 模糊命中多个 | 列出候选（技能名 + 武将 + 文案摘要）让用户选择，等回复再继续 |
| 模糊命中单个（与原名不一致） | 返回候选 + 文案摘要给用户确认后继续 |
| 零候选 | **不创建 `${SKILL_DIR}`**，返回未找到并列可能原因（PM 未保存 / 写法不同 / 落在其他 sheet），等用户补充再重试 |

> 武将归属仅靠合并单元格判定，不靠 null 推断。真未合并的单行标「未标注」。

**牌路径（OBJECT_TYPE=card）**

在「牌」sheet（`a109ea`）「牌名称」列 `+find`：

```bash
lark-cli sheets +find --find "<牌名>" --url "..." --sheet-id "a109ea"
```

命中后读该行 A/B/C/D/E/F（牌名/牌类型大类/牌类型小类/花色/点数/效果文案）。牌 sheet 无合并、每张牌单行即全部信息。匹配候选与零候选处理表同技能路径。

**牌路径一级模块清单来源**：阶段三按 `card-decomposition.md` §3 派生；其中源表权威清单从 [card-test-point-modules.md](references/card-test-point-modules.md)（对应 sheet `hYAw0h`）取，已在 1.0 快照门控刷到最新。PM 增量加模块时，下轮 revision 变化即自动接入。后续拆解走 `card-decomposition.md`。

#### 1.2 解析并查询文案引用

对每个条目文案（牌路径只有 1 条文案 = F 列效果文案），采用**反向查表**：将文案中的名词/短语依次去各 sheet 搜索，搜到才视为引用。

> **牌路径分支**：本节反查规则通用——文案中的牌/术语/技能/机制引用都走反向查表。**牌路径跳过**：sub_skills 检测（牌不授予技能）、关联技能集合检测（无）、D 列技能分类反查（牌 D 列是花色不是分类）。其余子项通用。

- `<sprite=N>` → 花色引用
- 去「牌」sheet（`a109ea`）「牌名称」列搜 → 牌引用（**必读「效果文案」列**，那是牌的真实定义；「牌类型大类」只是分类标签）
- 去「基本术语」sheet（`cFEl74`，已是术语**唯一来源**）「术语」列搜 → 术语/机制 / 状态效果 / 伤害类型引用
- **动作/机制类术语**（打出、当作X、转化、摸、弃置、获得、限X次…）靠名字搜不到——「基本术语」的「牌与牌堆操作」+「卡牌使用与效果」+「技能机制与能力变化」三组术语名是参数化形式（如「当xx牌」「转化牌」「每个出牌阶段限次」）。改为**现读这三组当动作词汇表、按定义语义匹配**文案动作（匹配规则与易混区分详见 [skill-decomposition.md](references/skill-decomposition.md)）
- 去「技能」sheet「技能名称」/「归属武将」列搜 → 技能 / 武将引用
- **授予技能（→ sub_skills）**：上一条命中的技能名 X，若在文案中被授予式动词支配（获得 / 得到 / 令…获得 / 授予 / 拥有 / 视为拥有 + X），则 X 是被测技能**授予的另一独立技能**（如天时→归心），记入 overview `sub_skills`（取 X 自身行的文案；文案里内联的定义只是副本）。被授予技能**不计入** 1.4 关联技能集合。entries / sub_skills / 关联技能 三者区分见 `skill-decomposition.md`。
- **技能分类（D 列）也反查**：每个 entry / sub_skill 的 D 列值（技能分类，如 `武将技` / `独立型buff` / `绑定型Buff` 等）去「基本术语」B 列搜 → 类型标签术语。命中后录其定义到 overview「核心术语」section，阶段三相关模块预期按定义照搬（如 `独立型buff` 影响"技能封禁/被重置/被删除/被转移"四模块的预期——源技能被封禁/重置不影响已施加的独立型 buff）。

每个命中的术语**记录其所属大类**（来自基本术语 A 列），供 1.3 的 5 模块归类复用。同时将被引用概念与「所有机制」sheet（`n2eNub`）A 列标题比对，引用到某机制主题则读取该标题下方区域（B-T 列）的全部预设测试点一并收集，供阶段三末段逐条生成用例（见 3.4）。技能未引用到任何机制主题时，overview §4 仍输出 section 带"未引用"占位行（避免沉默跳过假象）。

引用概念逐条收集：名称、原文定义（完整照搬不简化）、来源坐标、所属大类。

> **检索未收录**：被引用概念若**重要（影响预期）、无法凭常识推断、且各 sheet 均查不到**，其 `原文定义` 记 `⚠️ 源文档未收录` + 一句可能含义，不编造。此标注随该概念落入对应 overview section（核心术语 / 测试点背景知识 / 概念范围枚举 / 拆解映射表来源列）及用例预期，使缺口在概览层即可见。能推断或无关紧要的不标。

#### 1.3 拆解（按对象类型分支）

**技能路径（OBJECT_TYPE=skill）→ 拆解技能定义 5 模块（逐条目）　详见 `skill-decomposition.md`**

- 每个条目文案细拆为原子项（如「出 1 张红色手牌」→ `[出][1张][红色][手牌]`），各分配 ID（拆-01…）。
- 把每个原子项归入 5 个二级模块（发动阶段 / 发动条件 / 发动方式 / 发动效果 / 重置时机），可双归类。
- 牌相关原子项：查**牌类型层级快照**（10 个叶子类型，见 `sheet-structure.md`）做大类→小类展开，每个叶子类型单独成一条用例的依据；**绝不展到 A 列具体牌名**。
- 发动阶段：永远现读「技能发动阶段」大类 16 节点逐节点成用例。文案有时机词+归属者时，命中节点位置按文案归属者集合展开（情况数=展开数）；不命中节点不展开。规则详见 `skill-decomposition.md` §3 发动阶段。
- 主动条目与被动条目分开拆解、分开生成用例。
- **节点拆分**：把技能流程切成有序节点，供阶段三封禁/被重置/被删除/被转移 4 个一级模块的「发动期间每节点插入」使用。

**牌路径（OBJECT_TYPE=card）→ 拆解效果文案　详见 `card-decomposition.md`**

- 按文案**动词序列**拆原子项（目标限定 / 动作 / 数量 / 代价 / 效果各成一项），各分配 ID（拆-01…）。
- 把每个原子项归入牌的一级模块清单（见 `card-decomposition.md` §3，部分源表词部分⚠️临时）。
- 牌相关原子项的枚举触发规则与技能侧不同（见 `card-decomposition.md` §5）：「弃置X张牌」无类型限定才触发 10 叶子枚举；「攻击范围内」不枚举（落入"与装备交互"）。
- **不**遍历「技能发动阶段」16 节点（牌不套用此铁律）。
- **不**做 sub_skills、**不**做节点拆分（牌路径无"被封禁/被重置/被删除/被转移"4 模块、不需要节点序列）。

#### 1.4 关联技能 + 组合分析　详见 [associated-skill-interaction.md](references/associated-skill-interaction.md)

**仅技能路径**。扫描以下交互原因，命中的技能**去重进一个集合**：同武将多技能联动 / 同技能多条目联动（被动+主动配合，如赵云）、状态效果叠加/互斥、同触发时机/同触发条件。集合大小 N 作为完整性校验基准。交互类型标签取值见 associated-skill-interaction.md。

> **牌路径分支**：跳过本步。牌与技能的交互测试落在 card-decomposition.md「与技能交互」一级模块下处理，不构成独立的关联技能辅助模块。

### 阶段二：落盘 overview.json

将阶段一结果序列化为 `${SKILL_DIR}/overview.json`，字段与各 section 见 [overview-format.md](references/overview-format.md)：顶层 `entries`（多条目）+ sections（技能定义拆解映射表 / 概念范围枚举 / 核心术语 / 机制测试点 / 关联技能(辅助) / 测试点背景知识）。

> **牌路径分支**：
> - `entries` 长度恒为 1（牌只有 F 列一条效果文案），`hero`/`faction`/`skill_category` 字段省略或填 `null`，新增 `card_type_major`（牌类型大类）/ `card_type_minor`（牌类型小类）/ `suit`（花色）/ `rank`（点数）字段。
> - sections 中「关联技能(辅助)」**不输出**；其余 5 个 section 通用。
> - 「技能定义拆解映射表」section heading 改写为「效果拆解映射表」、列名含义同（拆解项ID/所属条目/原子项/归入二级模块/关联文档/来源）；「所属条目」列恒填「条目1」。

### 阶段三：生成 testcases.json（8 列）

用例 8 列结构（`编号/一级模块/二级模块/拆解项/标题/前置条件/步骤/预期结果`）与撰写规范见 [testcase-format.md](references/testcase-format.md)。生成后保存为 `${SKILL_DIR}/testcases.json`。

**预期结果原则**：撰写规范见 `testcase-format.md`「预期结果撰写原则」（要点：引用外部定义照搬原文 + 标来源、自身/信息足够直接写、结算不确定写「需游戏内验证」）。

#### 3.1 技能定义（5 子模块，逐条目）— 仅技能路径

> **牌路径分支**：本节跳过。牌直接走 3.2-牌路径，按 card-decomposition.md §3 的一级模块清单逐模块展开。

> ⚠️ **用例列结构与源表相反，务必照排、勿"纠正"回源表形态**：源表里 技能定义 是一级、5 子模块是二级；**用例表里反过来**——5 子模块提到「一级模块」列、原子项落「二级模块」列，让每条用例一眼看出针对哪个点。
> - 「一级模块」列 = 发动阶段 / 发动条件 / 发动方式 / 发动效果 / 重置时机 之一（**绝不写「技能定义」**）
> - 「二级模块」列 = 该用例针对的原子项（发动阶段枚举行填对应时间节点）
> - 「拆解项」列 = 原子项 ID（`拆-NN`）
>
> | 编号 | 一级模块 | 二级模块 | 拆解项 |
> |---|---|---|---|
> | TC-07 | 发动效果 | 摸1张牌 | 拆-03 |
> | TC-12 | 发动阶段 | 出牌阶段 | 拆-09 |
>
> 列映射详见 `testcase-format.md`。

对每个原子项按其所属子模块展开正/反/边界用例 + 维度展开（规则见 `skill-decomposition.md`）。

发动阶段：永远遍历「技能发动阶段」16 节点。预期判定与归属者展开规则见 `skill-decomposition.md` §3。

无限次时仅跳过「重置时机」子模块，其余 4 个不跳。

#### 3.2 其余一级模块

**技能路径**：对 `skill-test-point-modules.md` 中除「技能定义」外的每个一级模块各生成用例（当前为：技能封禁 / 被重置 / 被删除 / 被转移 / 刻写复制 / 多个同时在场 / 武将重新初始化 / 断线重连——以快照当前清单为准）。封禁/被重置/被删除/被转移的「发动期间」二级模块按 1.3 拆出的节点序列**每节点各 1 条**，全覆盖不取代表。

> ⚠️ 模块清单从 `skill-test-point-modules.md` 派生，**不硬编码**；若快照刷新后出现陌生新模块，生成已知部分并提示用户人工确认其规则。

**牌路径**：按 `card-decomposition.md` §3 五档结构生成用例：

1. **牌定义档（核心）**：4 子模块作"一级模块"列：**身份 / 打出条件 / 效果 / 响应**。文案 拆-NN 原子项强校验落在本档（身份用例「拆解项」列填 `—`，其余落对应 ID）。响应仅适用时出（杀类/锦囊类），否则整块跳过；其余子模块永远适用。**绝不写"牌定义"本身**（同技能侧 5 子模块直接当一级模块的规则）。
2. **通用必跑（4 个）**：发动者和目标者状态 / 与技能交互 / 断线重连情况 / 边缘场景。"发动者和目标者状态"按 §3.2 适用判定（纯自用/被动响应/装备自己 → 整块跳过、**不出占位行**）。
3. **源表现读**：从 `card-test-point-modules.md` 取 hYAw0h 当前模块清单（当前唯一：「牌的状态」），每个二级模块 ≥1 条用例。PM 增量自动接入；命中同名时让位牌定义档/通用必跑里的临时清单。
4. **自由补充（末尾、克制使用）**：PM 未规定的少数特殊场景，模型按需自拟（参考 §3.4 类目表）。
5. **机制测试点**（追加段，沿用 3.4）。

各档展开规则见 card-decomposition.md §4。

**绝不出现**「技能定义/发动阶段/发动条件/发动方式/重置时机」这类技能专属一级模块名（出现即误套技能路径，回 1.3 修）。**也不得出现"牌定义"作为一级模块名**（应为身份/打出条件/效果/响应之一，违反即结构未重排）。

#### 3.3 关联技能交互（辅助模块）

**仅技能路径**。对 1.4 去重集合中每个关联技能生成用例，交互类型作二级模块/标题标签。规则见 `associated-skill-interaction.md`。

> **牌路径分支**：跳过本节。

#### 3.4 机制测试点（追加段，放在测试用例表最后）

对 1.2 反查到的每个机制主题，**把该主题下的所有预设测试点逐条生成用例**（一条预设 = 一条用例，无去重）：
- **一级模块**：填机制主题名（源表 A 列原文，如「受伤的机制测试点」）
- **二级模块**：填该条预设测试点的内容/简述
- **拆解项**：填「—」（无 拆-NN ID）
- **标题 / 前置条件 / 步骤 / 预期结果**：按预设测试点的语义撰写

技能未引用任何机制主题时，本段不生成用例（overview §4 仍出占位行作"已检索"标记）。

#### 3.5 完整性校验（唯一闸门，保存前强制执行）

> **先校验快照新鲜度**：断言快照头部 `revision` == 当前 live revision；不等 → 说明 1.0 前置门控被跳过，回 1.0 补刷后再继续。（仅数字级核对，防「漏刷」、不防「只改号不核内容」——后者需内容级核对。）

保存 `testcases.json` 前循环校验，差集非空 → **按名称**补用例（不按数量）→ 重校验，连续两次差集全空才进入阶段四。

**技能路径校验清单**：

| 校验对象 | 断言 |
|---------|------|
| 技能条目 | 每个条目都有 ≥1 条技能定义用例 |
| 原子拆解项 | 每个原子项 ID ∈ 某条用例的「拆解项」列 |
| 概念范围枚举（含牌子类型、时间节点） | 每个展开子类型**单独** ≥1 条用例；**所有条目此表都须含「技能发动阶段」16 节点**（无条件枚举） |
| 一级模块 | 技能定义的 5 子模块各 ≥1 用例（输出中作为一级模块）+ `skill-test-point-modules.md` 其余每个一级模块 ≥1 用例（零跳过） |
| 技能定义列结构 | 技能定义类用例「一级模块」列 ∈ {发动阶段/发动条件/发动方式/发动效果/重置时机}，**不得出现「技能定义」**（出现即结构未重排，回 3.1 修） |
| 发动阶段 | 全部条目都遍历「技能发动阶段」16 节点；文案有时机词+归属者时，命中节点位置按归属者情况数展开。预期判定按 `skill-decomposition.md` §3 表，违反回 §3 修 |
| 关联技能 | 集合中每个技能名 ∈ 某条辅助模块用例标题 |
| 机制测试点 | 1.2 反查到的每个机制主题下每条预设测试点都生成 ≥1 条用例（未引用则免）；用例「一级模块」列 = 机制主题原文 |

技能路径最终排序：发动阶段 → 发动条件 → 发动方式 → 发动效果 → 重置时机（技能定义 5 子模块）→ 其余一级模块 → 关联技能辅助模块 → 机制测试点；各组内按模块聚合并全局重新编号。

**牌路径校验清单**（按 card-decomposition.md §3 五档结构）：

| 校验对象 | 断言 |
|---------|------|
| 牌定义档 - 身份 | 牌名/牌类型/花色/点数/效果文案 每个非空字段在「身份」一级模块下 ≥1 条用例 |
| 牌定义档 - 打出条件 | 永远 ≥1 条用例（除非被测牌是被动响应类如闪/无懈可击，此时整块跳过） |
| 牌定义档 - 效果 | 永远 ≥1 条用例（极特殊无效果牌可跳过） |
| 牌定义档 - 响应 | 适用判定：可被响应（杀类/锦囊类）→ ≥1 条；不适用整块缺席 |
| 原子拆解项 | 每个 `拆-NN` ID ∈ 牌定义档中某条用例的「拆解项」列；双归类的 拆-NN 须在两个子模块都出现 |
| 概念范围枚举 | 触发的枚举（按 card-decomposition.md §5 表判定）每子类型 ≥1 条用例 |
| 通用必跑（4 个） | 与技能交互 / 断线重连情况 / 边缘场景 各 ≥1 条；发动者和目标者状态按 §3.2 适用判定：适用即 ≥1 条、不适用整块缺席（**整模块无用例**） |
| 源表现读（hYAw0h） | 当前每个一级模块及其每个二级模块 ≥1 条用例（按 card-test-point-modules.md 当前快照） |
| 自由补充位置 | 全部自由补充用例**排在牌定义档/通用必跑/源表现读之后**（测试用例表的末段） |
| 模块名禁列 | **不得出现**「技能定义/发动阶段/发动条件/发动方式/重置时机」等技能专属一级模块名；**也不得出现"牌定义"作为一级模块名**（应为身份/打出条件/效果/响应之一） |
| 16 节点禁遍历 | 牌路径**不应出现**「技能发动阶段」16 节点遍历产生的 16 行枚举用例（出现即把技能侧铁律误套到牌上，回 1.3 修） |
| 机制测试点 | 1.2 反查到的每个机制主题下每条预设测试点都生成 ≥1 条用例（未引用则免）；用例「一级模块」列 = 机制主题原文 |

牌路径最终排序：**牌定义档（身份 → 打出条件 → 效果 → 响应）→ 通用必跑（按 3.2 顺序）→ 源表现读（按 hYAw0h A 列顺序）→ 自由补充（末尾）→ 机制测试点**；各组内按模块聚合并全局重新编号。

### 阶段四：输出交付

#### 4.1 生成 Excel + Markdown

```bash
PYTHON=/usr/bin/python3
$PYTHON /Users/zt-3803045/.openclaw/skills/skill-testcase-generator/scripts/generate_outputs.py \
  "${SKILL_DIR}/overview.json" "${SKILL_DIR}/testcases.json" "${SKILL_DIR}" "<对象名>"
```

生成 Excel（Sheet 1 测试概览 + Sheet 2 测试用例，含自动筛选与冻结首行）与同内容 Markdown。**禁止手写 Excel/Markdown 生成逻辑**，统一通过脚本保证格式一致。

#### 4.2 上传到飞书 Drive

```bash
lark-cli drive +create-folder --name "<对象名>_${TIMESTAMP}" --as bot   # 返回 folder_token
cd "${SKILL_DIR}" && \
lark-cli drive +upload --file ./<对象名>_测试用例.xlsx --folder-token "<folder_token>" --as bot && \
lark-cli drive +upload --file ./<对象名>_测试用例.md   --folder-token "<folder_token>" --as bot
```

从 `+create-folder` 返回值取 `folder_token`；从两次 `+upload` 返回值各取 `file_token`（xlsx 与 md 各一）。后续授权全部以 `--as bot` 执行——bot 始终为所有者，不做所有者转移。

#### 4.3 固定人员 + 发起用户授「可管理」（folder 上）

对下列 open_id 逐个在文件夹上授 `full_access`：

| 姓名 | open_id |
|---|---|
| 李冬生 | `ou_7b40308405b24776fcd2d9459698895f` |
| 张姝祺 | `ou_4ea890058aab2ccd713d9ddc3dd26bc6` |
| 龚亮 | `ou_42a1ac1d8cc73e00a5980c4ed2686cc2` |
| 当前发起用户 | 由 lark-cli 按本次 `--as user` 绑定的发起人自动解析其 open_id；显式查询用 `lark-cli contact +whoami --as user` |

```bash
lark-cli drive permission.members create \
  --params '{"token":"<folder_token>","type":"folder"}' \
  --data '{"member_type":"openid","member_id":"<ou_xxx>","perm":"full_access","type":"user"}' \
  --yes --as bot
```

> 发起用户 open_id 与固定三人重复时去重，只授一次；目标已是协作者时接口可能返回「已存在」，忽略即可，不中断。

#### 4.4 群聊可读（条件触发，folder 上）

仅当本次请求来自群聊时，把整群授 `view`。chat_id 取自 OpenClaw 注入的触发消息上下文（非环境变量）；feishu cli 无「当前会话」概念、无法反查发起群，故取不到 chat_id 时直接跳过本步。

```bash
# 1. 先探测 chat 类型，确认是群再授权（防止把 p2p 私聊或无效 id 当群授权报错）
lark-cli im chats get --params '{"chat_id":"<chat_id>"}' --as bot   # 读返回的 chat_mode

# 2. 仅当 chat_mode ∈ {group, topic} 时授群读；否则跳过
lark-cli drive permission.members create \
  --params '{"token":"<folder_token>","type":"folder"}' \
  --data '{"member_type":"openchat","member_id":"<chat_id>","perm":"view","type":"chat"}' \
  --yes --as bot
```

> 取不到 chat_id / `chats get` 失败（bot 不在群、id 无效）/ `chat_mode == p2p` → 一律跳过群读，不报错、不中断主流程。

#### 4.5 两个文件开启链接分享（组织内可读）

对 xlsx 与 md 各执行一次（`permission.public patch` 不支持 `folder` 类型，只能逐文件设置）：

```bash
lark-cli api PATCH "/open-apis/drive/v1/permissions/<file_token>/public" \
  --params '{"type":"file"}' \
  --data '{"link_share_entity":"tenant_readable"}' \
  --as bot
```

> ⚠️ **`lark-cli api` 会丢弃 URL 中的 query string**（如 `?type=file`），所有查询参数必须经 `--params` 传递，不能直接拼在 URL 里。
>
> `tenant_readable` = 组织内任何人凭链接可读。本设置仅限组织内、不涉对外分享；若仍返回 91009/91011 等「对外分享被租户策略管控」错误码，照实告知用户，不重试绕过。

#### 4.6 返回文件夹链接

用 `lark-cli im` 把「总结 + Drive 文件夹链接」作为一条消息**显式发送**给用户（一个链接即可访问两份文件，两份文件均已开启组织内链接分享）。

> ⚠️ 必须显式发送，**不要**把总结留到回合结尾的收尾文本里——否则它会排在 4.8 的反馈卡片之后，导致用户先看到卡片、后看到结果。

#### 4.7 写入产出统计表（Bitable）

把本次产出创建一条记录到飞书 Bitable「测试用例生成数据统计」
（[app_token=AyOlbNrnzavuCzscvYPchXwtnUU](https://ztgame.feishu.cn/base/AyOlbNrnzavuCzscvYPchXwtnUU)，表 `tblnlTnPgX12x9gP`）。
本表对技能/牌**不去重、全表累加**，每跑一次创建一条记录（运行ID `<对象名>_${TIMESTAMP}` 天然唯一，无需幂等）。

**写入统计记录**

```bash
$PYTHON ~/.openclaw/skills/skill-testcase-generator/scripts/append_stats.py \
  "${SKILL_DIR}/testcases.json" "<对象名>" "$OBJECT_TYPE" "<xlsx_url>" "${TIMESTAMP}"
```

脚本读 `testcases.json` 的 `rows` 算用例数、由 `${TIMESTAMP}` 派生 ISO 生成时间与运行ID，拼 fields 对象（项目=名将杀；类型/来源按 `OBJECT_TYPE` 取 技能·技能用例 / 牌·牌用例；覆盖率五列留空；飞书链接=4.2 上传的 **xlsx 文件链接**）后调 `lark-cli api POST .../records` 创建 Bitable 记录。
成功时 stdout 末尾输出 `RECORD_ID=<record_id>`，供后续反馈功能更新同一条记录。

> `<xlsx_url>` 用 4.2 上传 xlsx 拿到的 `file_token` 拼成：`https://ztgame.feishu.cn/file/<xlsx_file_token>`（点开直达用例表本身，而非文件夹）。
>
> 写表是**非关键的统计收尾**——发生在文件已上传、链接已返回之后。脚本内部对任何失败只打印警告并正常退出（exit 0），不中断主交付，也无需向用户报错。

#### 4.8 触发反馈收集（交给 user-feedback skill）

**前置**：确认 4.6 的总结消息已发出，再发反馈卡片——保证用户先看到结果、再看到反馈卡。

本 skill 的职责到 4.7 拿到 `RECORD_ID` 为止。接着**调用 user-feedback skill** 收集反馈——发卡、回写的全部逻辑都在 user-feedback（见其 SKILL.md），本 skill 只在生成完成这一刻触发它并传入：

- `RECORD_ID`：4.7 捕获的记录 ID
- `<对象名>`：阶段零提取的技能名/牌名（仅用于卡片文案）
- 发起用户（传入 user-feedback 时的接收人 ID）：
    私聊 → 先把 SenderId（open_id）转 union_id，再传给 user-feedback
    群聊 → 直传 chat_id
  转换命令：
    ```bash
    UNION_ID=$(npx lark-cli api GET "/open-apis/contact/v3/users/${SENDER_OPEN_ID}" \
      --params '{"user_id_type":"open_id"}' \
      | $PYTHON -c "import json,sys; print(json.load(sys.stdin)['data']['user']['union_id'])")
    ```

## ⚠️ 注意事项

1. **合并单元格必须先解析** — 不先解析会导致技能归属与条目范围错误。技能 sheet 的嵌套合并解析规则见 `sheet-structure.md`。
2. **照搬原文** — 技能文案、术语定义、牌效果一律用源文档原文，禁止简化改写；推断结论标「待验证」，不编造。预期结果撰写规范见 `testcase-format.md`。
3. **JSON 中文引号用「」代替 `"`** — 写入 overview.json / testcases.json 时，中文文本内的引号必须用「」，否则 JSON 解析失败。
4. **飞书操作降级链** — `lark-cli` 失败后才用 `feishu` 插件；feishu 插件再失败则告知用户需 OAuth 授权。
5. **快照刷新不是改号** — revision 变化后，四份快照必须从 live 数据全量重建（行号、模块数量、合并范围、术语坐标），再改 revision 号。
