# draw/ 目录文档

抽奖掉落相关的跨表校验规则（package: `draw`）。

## 规则文件

| 文件 | 规则类型 | 说明 | 源表 | 依赖表 |
|------|---------|------|------|--------|
| draw_drop_rule_reference.go | DRAW_DROP_RULE_REFERENCE_CHECK | 抽奖池掉落规则引用检查：Draw/DrawSkin 的 OnceDropRule/TenDropRule 引用 DropRule | Draw, DrawSkin | DropRule |
| drawskin_byproduct.go | DRAWSKIN_BYPRODUCT_CHECK | 皮肤抽奖副产品检查：DrawSkin byproduct 配置道具在 Item 表存在 | DrawSkin | Item |
| drawskin_once_item_cost.go | DRAWSKIN_ONCE_ITEM_COST_CHECK | 皮肤单抽消耗检查：OnceItemConfig 非空、引用 Item 表、数量 > 0 | DrawSkin | Item |
| drawfix_protection.go | DRAWFIX_PROTECTION_CHECK | 定向招募战令保护期检查：保护期内武将不出现在定向招募 | DrawFix | Item, SeasonPassReward, SeasonPass, Hero |
| drawfix_arena_protection.go | DRAWFIX_ARENA_PROTECTION_CHECK | 定向招募大将军保护期检查：赛季期间大将军武将不出现在定向招募 | DrawFix | Item, ArenaScoreReward, ArenaSeason, Hero |

## 保护期规则

- 战令武将：SeasonPass.StartTime + protectMonths（默认4个月）
- 大将军武将：ArenaSeason.SeasonEndTime（赛季结束即保护期结束，不使用 protectMonths）

## 开发注意事项

- ArenaScoreReward 表名是单数无 s
- 定向招募保护期数据流注释已标注在 drawfix_protection.go 和 drawfix_arena_protection.go 顶部
- 两个保护期规则共享 BuildHeroItemIdMap（Item 表语义映射）
