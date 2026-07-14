# coded_rules/cross_table 目录文档

跨表级别校验规则实现。按业务域分为 4 个子目录（独立 Go package）。

## 目录结构

| 目录 | package | 业务域 | 规则列表 |
|------|---------|--------|----------|
| [activity/](./activity/) | `activity` | 丹青阁、活动抽皮肤、活动任务奖励 | DANQINGGE_CUSTOM_PARAM, ACTIVITY_DRAWSKIN_CROSS_REFERENCE, ACTIVITY_DRAWSKIN_TIME_OVERLAP, ACTIVITY_TASK_REWARD |
| [draw/](./draw/) | `draw` | 抽卡掉落规则、皮肤抽奖、定向招募保护 | DRAW_DROP_RULE_REFERENCE, DRAWSKIN_BYPRODUCT, DRAWSKIN_ONCE_ITEM_COST, DRAWFIX_PROTECTION, DRAWFIX_ARENA_PROTECTION |
| [drop/](./drop/) | `drop` | 掉落道具、掉落规则条件/组ID | DROP_ITEM_MUST_IN_ITEM, DROP_ITEM_VALIDITY, DROP_RULE_CONDITIONAL, DROP_RULE_GROUP_ID |
| [hero/](./hero/) | `hero` | 武将掉落/熔炼/技能/合成 | HERO_DROP, HERO_DROP_VALIDDATE, HERO_MELT, HERO_SKILL_BUFF, HERO_SYNTHESIS |

## 规则列表

| 规则类型 | 说明 | 源表 | 依赖表 | 所属目录 |
|----------|------|------|--------|---------|
| `HERO_DROP_CHECK` | 武将掉落池检查 | Hero | Drop, SeasonPassReward, ArenaScoreReward | hero/ |
| `HERO_MELT_CHECK` | 武将熔炼检查 | Hero | — | hero/ |
| `HERO_SYNTHESIS_CHECK` | 武将合成检查 | Item | Hero, SeasonPassReward, SeasonPass, ArenaScoreReward, ArenaSeason | hero/ |
| `HERO_SKILL_BUFF_CHECK` | 武将技能Buff引用检查 | Hero | Skill, Buff | hero/ |
| `HERO_DROP_VALIDDATE_CHECK` | 普通武将掉落时间检查 | DropItem | Hero, SeasonPassReward, SeasonPass, ArenaScoreReward | hero/ |
| `ACTIVITY_TASK_REWARD_CHECK` | 活动任务奖励检查 | ActivityTask | Item | activity/ |
| `DANQINGGE_CUSTOM_PARAM_IS_ITEMID_CHECK` | 丹青阁自定义参数检查 | Activity | DrawSkin | activity/ |
| `ACTIVITY_DRAWSKIN_CROSS_REFERENCE_CHECK` | 活动与DrawSkin交叉引用检查 | Activity | DrawSkin | activity/ |
| `ACTIVITY_DRAWSKIN_TIME_OVERLAP_CHECK` | 活动与DrawSkin时间重叠检查 | Activity | DrawSkin | activity/ |
| `DROP_ITEM_MUST_IN_ITEM_CHECK` | 掉落道具检查 | DropItem | Item | drop/ |
| `DROP_ITEM_VALIDITY_CHECK` | 掉落道具条件和互斥检查 | DropItem | Item | drop/ |
| `DROP_RULE_CONDITIONAL_CHECK` | 掉落规则条件检查 | DropRule | DropGroup | drop/ |
| `DROP_RULE_GROUP_ID_CHECK` | 掉落规则组ID检查 | DropRule | DropGroup | drop/ |
| `DRAW_DROP_RULE_REFERENCE_CHECK` | 抽奖池掉落规则引用检查 | Draw | DropRule | draw/ |
| `DRAWSKIN_BYPRODUCT_CHECK` | 皮肤抽奖副产品检查 | DrawSkin | Item | draw/ |
| `DRAWSKIN_ONCE_ITEM_COST_CHECK` | 皮肤单抽消耗检查 | DrawSkin | Item | draw/ |
| `DRAWFIX_PROTECTION_CHECK` | 定向招募战令保护期检查 | DrawFix | Item, SeasonPassReward, SeasonPass, Hero | draw/ |
| `DRAWFIX_ARENA_PROTECTION_CHECK` | 定向招募大将军保护期检查 | DrawFix | Item, ArenaScoreReward, ArenaSeason, Hero | draw/ |

## 依赖表

| 表名 | 说明 | 被哪些规则使用 |
|------|------|----------------|
| `Hero` | 武将基础配置 | HERO_DROP_CHECK, HERO_MELT_CHECK, DRAWFIX_PROTECTION_CHECK, DRAWFIX_ARENA_PROTECTION_CHECK |
| `Item` | 道具配置 | ACTIVITY_TASK_REWARD_CHECK, DROP_ITEM_MUST_IN_ITEM_CHECK, HERO_SYNTHESIS_CHECK, DRAWSKIN_BYPRODUCT_CHECK, DRAWFIX_PROTECTION_CHECK, DRAWFIX_ARENA_PROTECTION_CHECK |
| `DrawSkin` | 皮肤抽奖池配置 | DANQINGGE_CUSTOM_PARAM_IS_ITEMID_CHECK, ACTIVITY_DRAWSKIN_CROSS_REFERENCE_CHECK, ACTIVITY_DRAWSKIN_TIME_OVERLAP_CHECK |
| `Drop` | 掉落配置 | HERO_DROP_CHECK |
| `SeasonPassReward` | 赛季战令奖励 | HERO_DROP_CHECK, DRAWFIX_PROTECTION_CHECK |
| `ArenaScoreReward` | 竞技场积分奖励 | HERO_DROP_CHECK, DRAWFIX_ARENA_PROTECTION_CHECK |
| `SeasonPass` | 赛季战令时间 | HERO_DROP_VALIDDATE_CHECK |
| `ArenaSeason` | 竞技场赛季 | DRAWFIX_ARENA_PROTECTION_CHECK |

## 注册方式

所有跨表规则在 `engine/table_registry.go` 的 `init()` 中统一注册，使用 `crossactivity`/`crossdraw`/`crossdrop`/`crosshero` 四个 import alias。

## 新增规则

详见 [跨表规则开发手册](../../docs/跨表规则开发手册.md)，包含：步骤清单、实现模板、测试规范、自检清单。

**新增规则时需选择正确的子目录**：
- 涉及活动表(Activity)为主 → `activity/`
- 涉及抽卡(Draw/DrawSkin/DrawFix)为主 → `draw/`
- 涉及掉落(Drop/DropRule/DropItem/DropGroup)为主 → `drop/`
- 涉及武将(Hero)为主 → `hero/`

## 开发注意事项

- 跨表规则通过 `sheetMap` 参数访问其他表
- 检查前必须验证所依赖的表是否存在
- 使用 `helpers.FindSheetBySuffix` 后缀匹配查找 Sheet
- 时间相关检查使用 `helpers` 包中的时间解析函数
- 错误信息须使用中文，包含足够定位信息(活动名、ID、具体值)
- **ArenaScoreRewards.xlsx 的 sheet 英文后缀是 `ArenaScoreReward`(单数无 s)，FindSheetBySuffix 必须使用单数**
- **TargetSheets 挂载原则**：规则的 `TargetSheets` 必须是核心检查字段所在的表(见上级 CLAUDE.md)，不要因为规则名称包含某个表名就挂载到该表
