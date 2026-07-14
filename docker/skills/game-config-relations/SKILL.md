---
name: game-config-relations
description: |
  游戏配表跨表关系知识库 - 提供已知的 Excel 配表之间的引用关系、外键关联和业务规则。

  **必须使用此技能的场景**：
  - 用户询问"XX表引用了哪些表"、"XX字段关联哪个表"
  - 用户需要了解配表之间的引用关系和数据流
  - 用户问"修改XX表会影响哪些其他表"
  - 用户需要验证跨表引用的完整性
  - 用户询问保护期规则、掉落规则、活动规则等业务逻辑
  - 用户问"武将和道具表有什么关系"、"战令奖励怎么关联武将"
  - 用户需要了解某个系统的完整配表架构（涉及多张表）
  - 用户询问字段索引（某个字段在哪些表中出现）
  - 用户问枚举定义在哪个文件

  **不要触发**：
  - 用户只是查看单个表的数据内容（→ excel-parser）
  - 用户只是筛选/分页浏览单个表（→ excel-parser）
  - 用户需要对比 Git 版本差异（→ excel-parser）
  - 用户要生成测试数据（→ excel-parser）

  **判定关键**：需要"了解关系"、"跨表关联"、"引用完整性"、"业务规则"时触发。

  **自迭代触发**：遇到知识盲区、文档错误、新关系发现时，调用 `/multica-cli` 创建 issue 驱动知识库进化。详见下方"自迭代机制"章节。
version: 1.0.0
tags: [game-config, relations, cross-table, foreign-key, references, schema]
---

# 游戏配表跨表关系知识库

> 基于 rain-excel-checker 跨表校验规则、配表字段反查索引和实际代码分析构建的配表关系知识库。
> 覆盖名将杀项目核心配表系统：武将、掉落、抽奖、活动、战令、竞技场等。

## 核心表关系图谱

### 1. 武将系统（Hero）

```
Hero (武将基础配置)
  ├── Skill (技能表) ← Hero.Skill 引用 Skill.Id
  ├── Buff (Buff表) ← Hero.Buff 引用 Buff 第2列标识符
  ├── SkillMelt (技能熔炼表) ← Hero.Skill 引用 SkillMelt.Id
  ├── DropItem (掉落道具表) ← 武将道具ID(1000000+HeroId)出现在DropItem.Item
  │   └── Item (道具表) ← DropItem.Item 引用 Item.Id
  ├── SeasonPassReward (战令奖励表) ← HighReward 包含武将道具ID
  │   └── SeasonPass (战令表) ← SeasonPassReward.SeasonPassId 引用 SeasonPass.Id
  ├── ArenaScoreReward (竞技场积分奖励表) ← Reward 包含武将道具ID
  │   └── ArenaSeason (竞技场赛季表) ← ArenaScoreReward.Season 引用 ArenaSeason.Id
  └── HeroAchieve (角色成就表) ← HeroAchieve.Hooker 关联成就系统
```

**关键规则**：
- 武将道具ID = 1000000 + 武将ID（Item 表前两位为10表示武将道具）
- 战令武将保护期：SeasonPass.StartTime + 4个月（默认）
- 大将军武将保护期：ArenaSeason.SeasonEndTime（赛季结束即保护期结束）
- 普通武将掉落时间：下一个周四凌晨 5:00

### 2. 掉落系统（Drop）

```
Drop.xlsx (同一文件包含3个Sheet)
  ├── DropRule (掉落规则表)
  │   ├── DropGroup (掉落分组表) ← DropRule.DropGroup 引用 DropGroup.Id
  │   ├── DropGroup (小保底) ← DropRule.EnsureSmallGroup 引用 DropGroup.Id
  │   ├── DropGroup (大保底) ← DropRule.EnsureBigGroup 引用 DropGroup.Id
  │   └── Item (道具表) ← DropRule.EnsureItemID 引用 Item.Id
  ├── DropGroup (掉落分组表)
  │   └── DropItem (掉落道具表) ← DropGroup 通过规则关联 DropItem.DropGroup
  └── DropItem (掉落道具表)
      └── Item (道具表) ← DropItem.Item 中的道具ID引用 Item.Id
```

**关键规则**：
- DropItem.Item 格式：`{道具ID;数量}{道具ID;数量}...`
- DropRule.DropGroup 支持逗号分隔的多个组ID
- EnsureSmallGroup 不应与 DropGroup 完全相同

### 3. 抽奖系统（Draw）

```
Draw.xlsx
  ├── Draw (名将册)
  │   └── DropRule (掉落规则表) ← Draw.OnceDropRule/TenDropRule 引用 DropRule.Id
  └── DrawSkin (皮肤抽奖池)
      ├── DropRule (掉落规则表) ← DrawSkin.OnceDropRule/TenDropRule 引用 DropRule.Id
      ├── Item (道具表) ← DrawSkin.byproduct 引用 Item.Id
      ├── Item (道具表) ← DrawSkin.OnceItemConfig 引用 Item.Id
      └── Activity (活动表) ← DrawSkin.ActivityId 引用 Activity.Id

DrawFix (定向招募)
  ├── Item (道具表) ← DrawFix.ItemIds 通过 Item.Type=Hero 映射到 Hero
  ├── Hero (武将表) ← 通过 Item 语义映射
  ├── SeasonPassReward (战令奖励) ← 保护期检查
  │   └── SeasonPass (战令表)
  └── ArenaScoreReward (竞技场奖励) ← 保护期检查
      └── ArenaSeason (竞技场赛季)

DrawPet (结缘亭)
  └── DropRule (掉落规则表) ← 引用 DropRule.Id
```

**关键规则**：
- 定向招募保护期：战令武将保护期内不能出现在 DrawFix 中
- DrawFix.EndTime < 保护期截止时间 → 错误
- 皮肤抽奖池的 ActivityId 必须对应丹青阁类型活动（ActTypeSkinRaffle）

### 4. 活动系统（Activity）

```
Activity (活动表)
  ├── DrawSkin (皮肤抽奖池) ← 丹青阁活动 CustomParma 引用 DrawSkin.Id
  ├── ActivityTask (活动任务表) ← ActivityTask.ActivityId 引用 Activity.Id
  │   └── Item (道具表) ← ActivityTask.Reward 引用 Item.Id
  ├── ActivityInvite (活动邀请表) ← ActivityInvite.ActivityId 引用 Activity.Id
  ├── ActivityKey (活动密钥表) ← ActivityKey.ActivityId 引用 Activity.Id
  ├── ActivityLogin (登录活动表) ← ActivityLogin.ActivityId 引用 Activity.Id
  ├── AccumulatedRechargeReward (累充奖励表) ← AccumulatedRechargeReward.ActId 引用 Activity.Id
  ├── DrawPet (结缘亭) ← DrawPet.ActivityId 引用 Activity.Id
  └── GiftShop (礼包商店) ← GiftShop.ActivityId 引用 Activity.Id
```

**关键规则**：
- 丹青阁活动识别：Name 包含"丹青阁"关键字，ActivityType = ActTypeSkinRaffle
- 丹青阁活动 CustomParma 必须非空且对应 DrawSkin.Id
- Activity 与 DrawSkin 双向引用一致性检查
- 关联活动与抽奖池时间范围必须有交集

### 5. 战令系统（SeasonPass）

```
SeasonPass (战令表)
  ├── SeasonPassReward (战令奖励表) ← SeasonPassReward.SeasonPassId 引用 SeasonPass.Id
  ├── SeasonPassTask (战令任务表) ← SeasonPassTask.SeasonPassId 引用 SeasonPass.Id
  ├── SeasonPassBag (战令礼包表) ← SeasonPassBag.SeasonPassId 引用 SeasonPass.Id
  └── Hero (武将表) ← SeasonPassReward.HighReward 包含武将道具ID → Hero
```

### 6. 竞技场系统（Arena）

```
ArenaSeason (竞技场赛季表)
  ├── ArenaScore (竞技场积分表) ← ArenaScore 关联赛季
  ├── ArenaScoreReward (竞技场积分奖励表) ← ArenaScoreReward.Season 引用 ArenaSeason.Id
  │   └── Hero (武将表) ← DanName含"大将军"的奖励包含武将道具ID
  └── ArenaSeasonLimitedHero (赛季限定武将表) ← ArenaSeasonLimitedHero.HeroIds 引用 Hero.Id
```

**注意**：ArenaScoreRewards.xlsx 的 Sheet 英文后缀是 `ArenaScoreReward`（单数无 s）

### 7. 商店系统（Shop）

```
Shop (商店表)
  ├── EShopType (商店类型枚举表)
  ├── ShopGoods (商品表) ← ShopGoods 引用 Shop
  │   └── Item (道具表) ← 商品关联道具
  └── GiftShop (礼包商店) ← GiftShop.ActivityId 引用 Activity.Id
```

### 8. 信件系统（Mail）

```
Mail (信件表)
  └── Item (道具表) ← 信件奖励可能包含道具引用
```

### 9. 成就系统（Achieve）

```
Achieve (成就表)
  ├── HeroAchieve (角色成就表) ← 关联武将成就
  └── Hero (武将表) ← Achieve.HeroItemId 关联 Hero.Id
```

### 10. 皮肤系统（Skin）

```
HeroSkinItem (英雄皮肤主表)  ← 核心枢纽表
  ├── Hero (武将表) ← HeroSkinItem.HeroId 引用 Hero.Id
  ├── HeroSkinSpine (皮肤Spine) ← HeroSkinSpine.SkinItemId 1:1关联 HeroSkinItem.SkinItemId
  ├── ItemHeroSkin (皮肤展示) ← ItemHeroSkin.SkinItemId 1:1关联 HeroSkinItem.SkinItemId
  ├── HeroSkillSkinFx (皮肤技能特效) ← 按 HeroId+SkinItemId+SkillId 三字段关联
  │   ├── Hero (武将表) ← HeroId 引用
  │   └── Skill (技能表) ← SkillId 引用
  ├── HeroSkinCollition (皮肤收藏册) ← HeroSkinItem.CollitionType 引用 HeroSkinCollition.Type
  ├── HeroSkinItem (自关联) ← AssociationSkinId 关联同表其他皮肤（套装关系）
  └── Item (道具表) ← SkinItemId 与 Item 表中 HeroSkin 类型道具存在前缀匹配关系
                       （Item ID ≈ SkinItemId × 10 + 后缀，少数特殊皮肤直接对应）

DrawSkin (皮肤抽奖池/丹青阁)
  ├── DropRule (掉落规则) ← OnceDropRule/TenDropRule 引用，常用规则ID 900004
  │   └── DropGroup → DropItem → Item (掉落道具引用链)
  ├── Item (道具表) ← BigAwardItemId 引用 HeroSkin 类型道具（保底大奖）
  ├── Item (道具表) ← byproduct 引用 Image/CardSkin 类型道具（副产物）
  ├── Item (道具表) ← OnceItemCost 引用丹青券(4000001)（抽卡消耗）
  └── Activity (活动表) ← ActivityId 引用 Activity.Id

CardSkin (手牌皮肤) ← 无出向外键（扁平资源表，全字段 client 导出）；CardSkin.Id 即 Item 表 Type=CardSkin 道具 ID，被 DrawSkin.byproduct 反向引用（见关键规则）
ItemFrame/FrameItem (边框道具) ← 独立资源表，无外键关联

限时皮肤活动子表（均通过 ActId 关联 Activity 表，Reward 字段引用 Item 表）：
  ├── LimitSkinTimesReward (限时皮肤抽卡次数奖励)
  ├── LimitSkinRecharge (限时皮肤充值奖励)
  └── LimitSkinAwardPreview (限时皮肤奖励一览)

皮肤与已有系统的交叉引用：
  SeasonPassBag (战令礼包) → Item(HeroSkin) ← 通过 ShowReward 引用皮肤道具
  SeasonPassReward (战令奖励) → Item(HeroSkin/Image) ← 通过 HighReward 引用
  ArenaSeason (竞技场赛季) → Item(HeroSkin) ← 通过 SpRewardId 引用
```

**关键规则**：
- **SkinItemId 编码规则**：HeroSkinItem.SkinItemId 是皮肤基础ID，Item 表中 HeroSkin 道具ID 在此基础上追加后缀（如 SkinItemId=1000100 对应 Item ID 10001001/10001002/10001003）
- **DrawSkin.BigAwardItemId**：直接引用 Item 表中完整的 HeroSkin 道具ID（含后缀），非 SkinItemId 本身
- **DrawSkin.byproduct**：逗号分隔的道具ID列表，引用 Image/CardSkin 类型道具
- **皮肤道具类型**：Item 表中皮肤相关 Type 包括 HeroSkin（武将皮肤）、CardSkin（手牌皮肤）、Frame（边框）、Image（形象）、Desktop（桌面）、SaiJiPiFuLihe（赛季皮肤礼盒）
- **皮肤收藏册**：HeroSkinCollition.Type 被 HeroSkinItem.CollitionType 引用，定义皮肤主题收藏（如五虎上将、工笔白描等）
- **皮肤套装**：HeroSkinItem.AssociationSkinId 自关联同一套装的不同皮肤

### 11. 兵临城下系统（CityWar）

```
CityWar (兵临城下城池配置表)
  ├── Skill (技能表) ← CityWar.FraySkill / CitySkill 引用 Skill.Id
  └── RoomMode (模式配置表) ← 通过 ModeType=RoomCityWar 关联兵临城下玩法

GlobalConfig (全局配置表)
  ├── Item (道具表) ← GlobalConfig.CityWarItem 引用 Item.Id（士兵道具 1000011）
  ├── Item (道具表) ← GlobalConfig.CityWarAwardItem 引用 Item.Id（青梅道具 1000014）
  └── Mail (信件表) ← GlobalConfig.CityWarSeasonGreenPlumMailId 引用 Mail.Id（9008）

RoomMode (模式配置表)
  ├── Team (队伍表) ← Team.ModeId 引用 RoomMode.Id
  └── RoomModeReward (模式奖励表) ← RoomMode.ModeRewardId 引用 RoomModeReward.Id

Team (队伍表)
  └── Item (道具表) ← Team.Cost 引用 Item.Id

RoomModeReward (模式奖励表)
  ├── Item (道具表) ← WinReward 引用 Item.Id
  └── Item (道具表) ← LoseReward 引用 Item.Id
```

**关键规则**：
- **CityWarArmyInterval**：`GlobalConfig` 中逗号分隔的 `int[]`，表示兵临城下军队出兵间隔。在 commit `b0227968` 中值从 `50,100,200,500`（4个值）扩展为 `50,100,200,500,1000`（5个值），说明新增了一个等级/阶段的出兵间隔配置；当前 main 分支仍为 4 个值。
- **RoomMode.ModeType**：兵临城下模式对应枚举值为 `RoomCityWar`（commit `b0227968` 中无前导空格；当前 main 分支对应单元格存在前导空格 `" RoomCityWar"`，属于数据格式问题）。
- **已知局限**：`CityWar.CardPileID` 的引用目标未在现有独立配表中明确找到；当前 `rain-excel-checker` 中未发现 CityWar 专属跨表校验规则。

### 12. 任务系统（Task）

```
Task (任务表)
  ├── TaskCompleteCond (任务完成条件表) ← Task.CompleteCond 引用 TaskCompleteCond.Id
  │     └── ETaskCompleteCond (任务完成条件枚举) ← TaskCompleteCond.CompleteCond 引用枚举值
  └── Item (道具表) ← Task.Reward 中的道具ID引用 Item.Id
```

**表结构要点**：
- `Task.xlsx`（Sheet「任务表|Task」）：124 行（含 4 行表头），120 条数据行（42 条有效 + 78 条空行），18 列
- `TaskCompleteCond.xlsx`（Sheet「任务完成条件表|TaskCompleteConditon」）：357 行（350 条有效数据），7 列
- `enum/ETaskCompleteCond_enum.xlsx`：51 个枚举项（值 0~50），其中 `CompleteCondNone=0` 为默认值，50 个有效完成条件类型

**关键规则**：
- `Task.CompleteCond` 引用格式为**嵌套数组** `{int[] Id}[]`，如 `{2001}` 或 `{1,2,3}{4,5}`，引用 `TaskCompleteCond.Id`
- `Task.Reward` 引用格式为 `{道具ID;数量}{道具ID;数量}...`（ItemCfg[] 类型），引用 `Item.Id`
- 实测 43 个 CompleteCond 引用 ID（值 1~6、1011~1027、2001~2020）在 TaskCompleteCond 表中**全部存在**
- `Task.Class` 任务类型分布：Daily（21）、ActLoginEightDays（9）、NewbieTask（8）、Task_Questionnaire（1）、Main/Branch/LimitTime（各1）
- `Task.AcceptCond` 当前全部为 0

**已知局限**：
- `ETaskClass_enum.xlsx`、`ETaskAcceptCond_enum.xlsx` 枚举文件在 config 仓库中缺失，Class 与 AcceptCond 字段的枚举值语义需结合代码确认（AcceptCond=0 推测为"无条件接受"）
- `ACTIVITY_TASK_REWARD_CHECK` 规则（`cross_table/activity/task_reward.go`）针对的是预留的 **ActivityTask** 表（当前 config 仓库不存在），且在 `engine/table_registry.go:104` 处于**注释未注册**状态（`// TODO: ActivityTaskRewardCheckRule 尚未实现`），当前**未启用**

### 13. 称号/Title 系统（含公会称号）

```
Title (称号表)
  ├── ETitleType (称号类型枚举) ← Title.TitleType 引用枚举值（非外键到配表）
  └── ETitleTimeType (称号时间类型枚举) ← Title.TimeType 引用枚举值

TitleByRank (排行称号表) → Title (称号表)
  └── TitleByRank.TitleID 引用 Title.Id（外键式引用；当前 TitleByRank 无有效数据，属预留结构）
```

**表结构要点**：
- `Title_称号表.xlsx`（Sheet「称号表|Title」）：4 行表头 + 10 条数据（Id 1–10），8 列
  - 字段：Id(int)、Name(string)、TitleType(ETitleType)、Group(int)、Weight(int)、TimeType(ETitleTimeType)、Bg(string)、Extra(int[])
  - 称号类型：`TitleType_GuildWeekTop=1`（公会周榜第一相关：会长/副会长/成员，Id 1–3）、`TitleType_GuildWeekTen=2`（公会周榜第二到第十，Id 4–10）
  - 持续时间：10 条数据 `TimeType` 均为 `TitleTimeType_Week=3`（持续当周）
  - `Extra` 为 `int[]`（如 `{6,10}`），语义需结合代码确认
  - 最新修改 commit `eeeb25d5`（`#cards-17623 【公会战】称号修改`）：Id=6 中文名由"守土"改为"屏翰"
- `TitleByRank_排行称号表.xlsx`（Sheet「排行称号|TitleByRank」）：4 行表头，字段 Id(int)、RankMin(int)、RankMax(int)、TitleID(int)，当前**无有效数据**（预留结构）
- 枚举文件（均位于 `excel/enum/`）：
  - `ETitleType_enum.xlsx`（Sheet「称号类型|ETitleType」）：`TitleType_GuildWeekTop=1`、`TitleType_GuildWeekTen=2`
  - `ETitleTimeType_enum.xlsx`（Sheet「称号时间类型|ETitleTimeType」）：4 值齐全——`TitleTimeType_Forever=1`、`TitleTimeType_Time=2`、`TitleTimeType_Week=3`、`TitleTimeType_Month=4`
  - `ETitleID_enum.xlsx`（Sheet 名疑为笔误「称号类型|ETitleID」）：10 个称号 ID `TitleType_GuildCityWeek1..10`（值 1–10），description 为称号中文名

**关键规则**：
- `Title.TitleType` / `Title.TimeType` 引用的是**枚举值**，不是外键到其他配表
- `TitleByRank.TitleID` 引用 `Title.Id`，是 Title 系统唯一的**外键式跨表引用**（当前 TitleByRank 无数据，引用未实际触发）
- 公会称号与公会山河周榜是**业务语义关联**（按周榜排名授予会长/副会长/成员/前 10/前 50），非配表外键；周榜奖励由 `GuildWeekRankAwards` / `GuildCityWeekRankAwards` 等表配置（见已知局限）
- `Title.Bg` 为 client 端 UI 底图资源名（如 `ui_s1_gonghui_frame_chenghao02`），不引用配表

**已知局限**：
- Title 系统配表（`Title_称号表`、`TitleByRank_排行称号表` 及三个 `ETitle*` 枚举）当前仅在特性分支（commit `eeeb25d5` 及后续如 `v0.0.8-dev-ob0618`）存在，**尚未进入 config 的 HEAD/main**；使用本知识时需注意配表版本
- `rain-excel-checker` 中**无 Title 专属跨表校验规则**；潜在规则建议：`TitleByRank.TitleID` 应引用 `Title.Id`（当前无数据，暂不触发）
- `ETitleID_enum.xlsx` 的 description 未随 Title 表 Id=6 改名同步（枚举仍为"守土"，Title 表已为"屏翰"），属数据级命名漂移
- 完整公会系统其他配表（城池 `GuildCityCreate/Level/Name/WarReward`、山河周榜奖励 `GuildCityWeekRankAwards/GuildWeekRankAwards`、事件 `GuildEvent`、红包 `GuildRedPack` 等）**仍未纳入本知识库**

## 字段索引速查

### 按字段名反查表

字段反查索引位于 `D:/work/config/docs/field-index/`，按字段名首字母分组：
- 1870 个配置表字段，1352 个枚举项
- 格式：字段名 | 中文名 | 类型 | Excel文件 | Sheet名 | 导出标识

常用字段位置：

| 字段名 | 所在表（Sheet） | 说明 |
|--------|----------------|------|
| Id | 几乎所有表 | 主键 |
| Name | 几乎所有表 | 名称 |
| HeroId | HeroSkinItem, HeroSkillSkinFx, PartnerTrainSeat, PveRougeBattleNode, RecommendBd, SeasonPass | 武将ID |
| SkinItemId | HeroSkinItem, HeroSkinSpine, ItemHeroSkin, HeroSkillSkinFx | 皮肤道具ID（基础ID，与Item表前缀匹配） |
| CollitionType | HeroSkinItem | 皮肤收藏册类型 |
| BigAwardItemId | DrawSkin | 丹青阁保底大奖道具ID（引用Item.HeroSkin类型） |
| byproduct | DrawSkin | 丹青阁副产物道具ID列表（引用Image/CardSkin类型） |
| ItemId | 多个奖励/掉落表 | 道具ID |
| ActivityId | ActivityInvite, ActivityKey, ActivityLogin, ActivityTask, Draw, DrawPet, DailyAsk, NianShouLaiXi | 活动ID |
| DropGroup | DropRule, DropItem | 掉落组ID |
| DropRule | Draw, DrawSkin, ThousandPics | 掉落规则ID |
| SeasonPassId | SeasonPassBag, SeasonPassReward, SeasonPassTask | 战令ID |
| Season | ArenaScoreReward, PeakArena, PeakRankReward, PeakWinsReward | 赛季 |
| CompleteCond | Task | 任务完成条件（嵌套数组 `{int[] Id}[]`，引用 TaskCompleteCond.Id） |
| Reward | Task, ActivityTask | 任务奖励（`{道具ID;数量}` 格式，引用 Item.Id） |
| SendStatus | Task | 推送状态 |
| CityWarArmyInterval | Global_全局配置表.xlsx / 全局配置表\|GlobalConfig | 兵临城下军队出兵间隔配置（commit b0227968 从4个值扩展为5个值） |
| CityWarItem | Global_全局配置表.xlsx / 全局配置表\|GlobalConfig | 兵临城下士兵道具ID（引用Item.Id） |
| CityWarAwardItem | Global_全局配置表.xlsx / 全局配置表\|GlobalConfig | 兵临城下青梅道具ID（引用Item.Id） |
| CityWarSeasonGreenPlumMailId | Global_全局配置表.xlsx / 全局配置表\|GlobalConfig | 兵临城下赛季青梅邮件ID（引用Mail.Id） |
| RoomCityWar | RoomMode.xlsx / 模式配置\|RoomMode | 兵临城下玩法模式枚举值（ModeType字段） |
| TitleType | Title_称号表.xlsx / 称号表\|Title | 称号类型（引用 ETitleType 枚举：GuildWeekTop=1 公会周榜第一/GuildWeekTen=2 周榜前10） |
| TimeType | Title_称号表.xlsx / 称号表\|Title | 持续时间类型（引用 ETitleTimeType 枚举；Title 表 10 条数据均为 Week=3） |
| Group | Title_称号表.xlsx / 称号表\|Title | 称号组（当前数据均为 1） |
| Weight | Title_称号表.xlsx / 称号表\|Title | 称号权重（1–10） |
| Bg | Title_称号表.xlsx / 称号表\|Title | 称号底图（client UI 资源名，不引用配表） |
| Extra | Title_称号表.xlsx / 称号表\|Title | 额外参数（int[]，如 {6,10}，语义需结合代码） |
| TitleID | TitleByRank_排行称号表.xlsx / 排行称号\|TitleByRank | 称号配置ID（引用 Title.Id；当前表无有效数据，预留结构） |

### 枚举索引

枚举定义位于 `D:/work/config/docs/field-index/enum/`，按枚举Key首字母分组。

常用枚举：

| 枚举 | 文件 | 说明 |
|------|------|------|
| EActivityType | EActivityType_enum.xlsx | 活动类型（含 ActTypeSkinRaffle 丹青阁） |
| EHeroType | EHeroType_enum.xlsx | 武将类型（1=普通武将） |
| ESkillId | ESkillId_enum.xlsx | 技能ID枚举 |
| ECountry | ECountry_enum.xlsx | 国家/势力 |
| ETaskCompleteCond | ETaskCompleteCond_enum.xlsx | 任务完成条件 |
| ESkinCollitionType | ESkinCollitionType_enum.xlsx | 皮肤收藏册类型（如五虎上将、工笔白描） |
| ESkinRailyType | ESkinRailyType_enum.xlsx | 皮肤品质类型 |
| ESkinType | ESkinType_enum.xlsx | 皮肤类型（如 SkinNormalSkin） |
| ETitleType | ETitleType_enum.xlsx | 称号类型（GuildWeekTop=1 公会周榜第一、GuildWeekTen=2 周榜前10） |
| ETitleTimeType | ETitleTimeType_enum.xlsx | 称号时间类型（Forever=1/Time=2/Week=3/Month=4，4值齐全；Title 表均用 Week=3） |
| ETitleID | ETitleID_enum.xlsx | 称号ID枚举（GuildCityWeek1..10，值1-10，description 为称号中文名；Sheet 名疑笔误为「称号类型」） |

## 跨表校验规则清单

### activity/ 目录（4条规则）

| 规则 | 源表 | 依赖表 | 说明 |
|------|------|--------|------|
| DANQINGGE_CUSTOM_PARAM_IS_ITEMID_CHECK | Activity | DrawSkin | 丹青阁CustomParam非空且引用DrawSkin |
| ACTIVITY_DRAWSKIN_CROSS_REFERENCE_CHECK | DrawSkin/Activity | Activity/DrawSkin | 双向引用一致性+活动类型校验 |
| ACTIVITY_DRAWSKIN_TIME_OVERLAP_CHECK | DrawSkin | Activity | 关联活动与抽奖池时间交集 |
| ACTIVITY_TASK_REWARD_CHECK | ActivityTask | Item | 任务奖励道具存在性+数量>0 |

### draw/ 目录（5条规则）

| 规则 | 源表 | 依赖表 | 说明 |
|------|------|--------|------|
| DRAW_DROP_RULE_REFERENCE_CHECK | Draw/DrawSkin | DropRule | OnceDropRule/TenDropRule引用 |
| DRAWSKIN_BYPRODUCT_CHECK | DrawSkin | Item | byproduct道具存在性 |
| DRAWSKIN_ONCE_ITEM_COST_CHECK | DrawSkin | Item | 单抽消耗道具存在性+数量>0 |
| DRAWFIX_PROTECTION_CHECK | DrawFix | Item/SeasonPassReward/SeasonPass/Hero | 战令保护期检查 |
| DRAWFIX_ARENA_PROTECTION_CHECK | DrawFix | Item/ArenaScoreReward/ArenaSeason/Hero | 大将军保护期检查 |

### drop/ 目录（4条规则）

| 规则 | 源表 | 依赖表 | 说明 |
|------|------|--------|------|
| DROP_ITEM_MUST_IN_ITEM_CHECK | DropItem | Item | 掉落道具存在性 |
| DROP_ITEM_VALIDITY_CHECK | DropItem | DropGroup/Item | 条件和互斥检查 |
| DROP_RULE_CONDITIONAL_CHECK | DropRule | DropGroup/Item | 保底机制条件引用 |
| DROP_RULE_GROUP_ID_CHECK | DropRule | DropGroup | 组ID引用存在性 |

### hero/ 目录（5条规则）

| 规则 | 源表 | 依赖表 | 说明 |
|------|------|--------|------|
| HERO_DROP_CHECK | Hero | DropItem/SeasonPassReward/SeasonPass/ArenaScoreReward/ArenaSeason | 武将掉落池检查 |
| HERO_DROP_VALIDDATE_CHECK | DropItem | Hero/SeasonPassReward/SeasonPass/ArenaScoreReward | 普通武将掉落时间检查 |
| HERO_MELT_CHECK | Hero | SkillMelt/SeasonPassReward/SeasonPass/ArenaScoreReward/ArenaSeason | 熔炼配置检查 |
| HERO_SKILL_BUFF_CHECK | Hero | Skill/Buff | 技能Buff引用 |
| HERO_SYNTHESIS_CHECK | Item | Hero/SeasonPassReward/SeasonPass/ArenaScoreReward/ArenaSeason/DropItem | 合成保护期检查 |

## 保护期规则汇总

| 武将类型 | 保护期起算点 | 保护期截止时间 | 规则文件 |
|----------|-------------|---------------|----------|
| 战令武将 | SeasonPass.StartTime | StartTime + 4个月 | hero_drop.go, drawfix_protection.go |
| 大将军武将 | ArenaSeason.SeasonStartTime | ArenaSeason.SeasonEndTime | hero_drop.go, drawfix_arena_protection.go |
| 普通武将 | OpenDate | 下一个周四 5:00 | hero_drop_validdate.go |

## 常见查询模式

### Q: 修改 Hero 表会影响哪些检查？
触发规则：HERO_DROP_CHECK, HERO_MELT_CHECK, HERO_SKILL_BUFF_CHECK
关联表：DropItem, SeasonPassReward, SeasonPass, ArenaScoreReward, ArenaSeason, SkillMelt, Skill, Buff

### Q: 修改 DropRule 表会影响哪些检查？
触发规则：DROP_RULE_CONDITIONAL_CHECK, DROP_RULE_GROUP_ID_CHECK
关联表：DropGroup, Item
被引用：Draw/DrawSkin 的 OnceDropRule/TenDropRule

### Q: 修改 Activity 表会影响哪些检查？
触发规则：DANQINGGE_CUSTOM_PARAM_IS_ITEMID_CHECK, ACTIVITY_DRAWSKIN_CROSS_REFERENCE_CHECK
关联表：DrawSkin, ActivityTask

### Q: 新增武将需要配置哪些表？
必需：Hero（基础信息）、Skill（技能）、SkillMelt（熔炼）
掉落：DropItem（加入掉落池，ValidDate配置）
可选：HeroAchieve（成就）、HeroSkinItem（皮肤）

### Q: 战令武将配置涉及哪些表？
SeasonPass（战令时间）、SeasonPassReward（奖励含武将道具）、SeasonPassTask（任务）、
DropItem（掉落时间）、Hero（武将信息）

### Q: 修改 Task 表会影响哪些检查？
当前 Task 表**无自动跨表校验规则**（`ACTIVITY_TASK_REWARD_CHECK` 针对预留的 ActivityTask 表且未注册启用），但修改时需关注以下隐式约束：
- `Task.CompleteCond` 引用的 ID 必须在 `TaskCompleteCond` 表中存在
- `Task.Reward` 中的道具 ID 必须在 `Item` 表中存在
关联表：TaskCompleteCond, Item

### Q: 新增武将皮肤需要配置哪些表？
必需：HeroSkinItem（皮肤主数据，SkinItemId+HeroId+Name等）
关联：HeroSkinSpine（Spine动画，按SkinItemId 1:1）、ItemHeroSkin（展示立绘，按SkinItemId 1:1）
可选：HeroSkillSkinFx（技能特效，按HeroId+SkinItemId+SkillId）、HeroSkinCollition（收藏册主题）
道具：Item 表需新增对应的 HeroSkin 类型道具（ID = SkinItemId × 10 + 后缀）

### Q: 皮肤抽奖池（丹青阁）涉及哪些表？
DrawSkin（抽奖池配置）→ DropRule（掉落规则）→ DropGroup → DropItem → Item（道具）
DrawSkin.BigAwardItemId → Item(HeroSkin 类型，保底大奖)
DrawSkin.byproduct → Item(Image/CardSkin 类型，副产物)
DrawSkin.ActivityId → Activity（丹青阁活动关联）

### Q: 皮肤道具ID如何跨表追踪？
HeroSkinItem.SkinItemId（基础ID）
  → Item 表 HeroSkin 道具（ID ≈ SkinItemId × 10 + 后缀）
  → DrawSkin.BigAwardItemId（引用完整的 Item ID）
  → DropItem.Item（掉落道具引用完整的 Item ID）
  → SeasonPassBag/SeasonPassReward（战令奖励引用完整的 Item ID）
  → ArenaSeason.SpRewardId（竞技场奖励引用完整的 Item ID）

## 知识库覆盖范围

**当前覆盖**：武将、掉落、抽奖、活动、战令、竞技场、商店、信件、成就、皮肤、兵临城下、任务、称号/Title（13个系统，18条跨表规则）

**未覆盖系统**（查询时请明确告知用户）：邮件、好友、公会（除公会称号外，含公会战/城池/事件/红包/周榜奖励等）、PVE、伙伴、坐骑

**已知局限**：
- 部分规则细节可能依赖代码实现（如保护期月数配置）
- 字段索引可能随配表版本更新而滞后

## 自迭代机制（知识库持续进化）

本知识库通过 Multica issue 驱动自成长。分析过程中遇到以下场景时，**必须**调用 `/multica-cli` 创建 issue，由独立 agent 验证并修复知识库。

### 触发提 issue 的场景

| 场景 | 判定标准 | Issue 类型 |
|------|---------|-----------|
| **知识盲区** | 用户查询未覆盖的系统（邮件、好友、公会其他模块、PVE、伙伴、坐骑等） | `knowledge-gap` |
| **文档错误** | 发现知识库描述与代码/配表实际行为不符 | `doc-error` |
| **新关系发现** | 发现未被记录的跨表引用或业务规则 | `new-relation` |
| **规则滞后** | 代码实现已变更但知识库未同步 | `outdated-rule` |
| **索引错误** | 字段索引位置、枚举值与实际配表不一致 | `index-error` |

### 禁止行为

- **禁止**在分析过程中猜测未覆盖系统的关系
- **禁止**忽略知识库与代码的不一致
- **禁止**用"可能"、"大概"等模糊表述绕过知识盲区

### Issue 创建规范

创建 issue 时必须包含以下验证细节，确保接收 agent 无需额外上下文即可执行验证：

```bash
multica issue create --title "[类型] 简要描述" --description-stdin --project rain-qa-func <<'EOF'
## 问题描述
一句话概括发现的问题

## 发现位置
- 分析上下文：用户在分析 XX 配表提交时提出
- 涉及文件：rain-excel-checker/xxx.go:行号（如已知）
- 涉及配表：XXX.xlsx/SheetName

## 当前知识库状态
- 知识库声称：XXX
- 实际观察到的：YYY

## 验证步骤（必须可独立执行）
1. 打开文件 `D:/work/config/XXX.xlsx`，检查 Sheet `YYY` 的 ZZZ 字段
2. 对比代码 `rain-excel-checker/xxx.go:行号` 的实现逻辑
3. 确认具体差异点

## 建议修复方向
- 如果是知识盲区：建议补充 XXX 系统的表关系图谱
- 如果是文档错误：建议修正 XXX 描述为 YYY
- 如果是新关系：建议添加 XXX → YYY 的引用关系

## 必须迭代的产物

**所有 issue 的最终产出必须是修改 `game-config-relations/SKILL.md` 文件**，确保知识库真正进化：

| Issue 类型 | SKILL.md 修改位置 | 修改内容 |
|-----------|------------------|---------|
| `knowledge-gap` | 对应系统的"核心表关系图谱"章节 | 新增该系统的关系图谱和关键规则 |
| `knowledge-gap` | "知识库覆盖范围"小节 | 更新覆盖系统列表，从未覆盖移到已覆盖 |
| `doc-error` | 错误描述的所在章节 | 修正错误描述，保留正确信息 |
| `new-relation` | 相关系统的"核心表关系图谱"章节 | 添加新的引用关系箭头 |
| `outdated-rule` | 规则所在章节的"关键规则"小节 | 更新规则描述，删除过时内容 |
| `index-error` | "字段索引速查"章节 | 修正字段位置或枚举值 |

**禁止行为**：
- 禁止仅验证问题而不修改 SKILL.md
- 禁止将修复内容写在其他文件或注释中
- 禁止关闭 issue 时 SKILL.md 未同步更新

## 相关代码引用
```go
// 如有相关代码片段，在此引用
```
EOF
```

### Issue 类型标签

创建 issue 时标题前缀使用以下标签：
- `[knowledge-gap]` — 知识盲区，需要调研补充
- `[doc-error]` — 文档描述错误，需要修正
- `[new-relation]` — 新发现的跨表关系
- `[outdated-rule]` — 规则滞后于代码实现
- `[index-error]` — 字段索引或枚举错误

### 示例

**场景**：用户查询"伙伴系统配表关系"，但知识库未覆盖

```bash
multica issue create --title "[knowledge-gap] 伙伴系统配表关系缺失" --description-stdin --project rain-qa-func <<'EOF'
## 问题描述
用户查询伙伴系统（Partner）的配表关系，当前知识库未覆盖该系统。

## 发现位置
- 分析上下文：用户询问 Partner.xlsx 与其他表的关系
- 涉及配表：D:/work/config/Partner.xlsx

## 当前知识库状态
- 知识库声称：未覆盖系统列表包含"伙伴"
- 实际观察到的：Partner.xlsx 存在且可能关联 Hero、Item 等表

## 验证步骤
1. 列出 `D:/work/config/` 下所有 Partner*.xlsx 文件
2. 打开 Partner.xlsx，记录所有 Sheet 名及字段
3. 检查 Partner 表是否有 HeroId、ItemId 等外键字段
4. 检查代码中是否有 partner 相关的跨表校验规则
5. 构建 Partner 系统的表关系图谱

## 建议修复方向
补充 Partner 系统到知识库的核心表关系图谱，包括：
- Partner 基础表与 Hero/Item 的引用关系
- 伙伴养成相关的跨表规则

## 必须迭代的产物
修改 `game-config-relations/SKILL.md`：
1. 在"核心表关系图谱"章节新增"10. 伙伴系统（Partner）"小节
2. 更新"知识库覆盖范围"：将"伙伴"从未覆盖系统列表移除，已覆盖系统数改为 10
3. 如有跨表规则，添加到"跨表校验规则清单"

## 相关代码引用
无（知识盲区）
EOF
```

---

## 注意事项

1. **ArenaScoreReward 表名是单数无 s** - FindSheetBySuffix 必须使用 "ArenaScoreReward"
2. **武将道具ID规则** - 1000000 + HeroId，Item 表前两位为10表示武将道具
3. **Item.Type=Hero** - 用于 DrawFix 的语义映射（ItemId → HeroId）
4. **TargetSheets 挂载原则** - 规则必须挂载在核心检查字段所在的表上
5. **CustomParma 拼写** - Activity 表实际字段名是 CustomParma（不是 CustomParam）
6. **时间格式** - 配表时间通常为字符串，需用 helpers.ParseDate 解析
7. **未覆盖系统提示** - 如果用户查询邮件、好友、公会（除公会称号外）、PVE、伙伴、坐骑等系统，**在告知未覆盖的同时必须创建 issue**
8. **SkinItemId 前缀匹配** - HeroSkinItem.SkinItemId 是皮肤基础ID，Item 表中 HeroSkin 道具ID = SkinItemId × 10 + 后缀（大部分情况），不是直接1:1对应
9. **任务系统 CompleteCond 嵌套数组格式** - `Task.CompleteCond` 类型为 `{int[] Id}[]`，是嵌套数组（如 `{1,2,3}{4,5}`），解析时需先按 `{...}` 分组再按 `,`/`;` 拆分，引用 `TaskCompleteCond.Id`
10. **任务系统枚举文件缺失** - `ETaskClass_enum.xlsx`、`ETaskAcceptCond_enum.xlsx` 在 config 仓库中缺失，`Task.Class` 与 `Task.AcceptCond` 的枚举值语义需结合代码确认
11. **ACTIVITY_TASK_REWARD_CHECK 未启用** - 该规则在 `engine/table_registry.go:104` 处于注释状态（`// TODO: ...尚未实现`），针对的是预留的 ActivityTask 表（当前不存在），**当前未生效**，不要误以为 Task/ActivityTask 已有奖励校验
12. **皮肤道具类型** - Item 表皮肤相关 Type：HeroSkin、CardSkin、Frame、Image、Desktop、SaiJiPiFuLihe（赛季皮肤礼盒）
13. **DrawSkin.BigAwardItemId** - 引用的是 Item 表中完整的 HeroSkin 道具ID（含后缀），非 SkinItemId 本身
14. **Title 表分支状态** - `Title_称号表`、`TitleByRank_排行称号表` 及 `ETitleType/ETitleTimeType/ETitleID` 枚举当前仅在特性分支（commit `eeeb25d5` 及后续），**尚未进入 config HEAD/main**，使用本知识时需注意配表版本
15. **TitleByRank.TitleID 引用** - `TitleByRank.TitleID` 引用 `Title.Id`（Title 系统唯一外键式跨表引用），当前 TitleByRank 表无有效数据为预留结构，引用未实际触发
