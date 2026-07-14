# hero/ 目录文档

武将相关的跨表校验规则（package: `hero`）。

## 规则文件

| 文件 | 规则类型 | 说明 | 源表 | 依赖表 |
|------|---------|------|------|--------|
| hero_drop.go | HERO_DROP_CHECK | 武将抽卡掉落检查：已开放武将加入掉落库、战令/大将军武将掉落配置符合保护期 | Hero | DropItem, SeasonPassReward, SeasonPass, ArenaScoreReward, ArenaSeason |
| hero_drop_validdate.go | HERO_DROP_VALIDDATE_CHECK | 普通武将掉落时间检查：已开放普通武将 DropItem.ValidDate 必须 <= 下一个周四 5:00 | DropItem | Hero, SeasonPassReward, SeasonPass, ArenaScoreReward |
| hero_melt.go | HERO_MELT_CHECK | 武将熔炼检查：开放武将必须可熔炼、技能配置熔炼、战令/大将军熔炼时间 | Hero | SkillMelt, SeasonPassReward, SeasonPass, ArenaScoreReward, ArenaSeason |
| hero_skill_buff.go | HERO_SKILL_BUFF_CHECK | 武将技能和Buff引用检查：Hero.Skill 引用 Skill 表、Hero.Buff 引用 Buff 表 | Hero | Skill, Buff |
| item_synthesis.go | HERO_SYNTHESIS_CHECK | 武将合成检查：DropItem.ValidDate 不早于保护期截止时间 | Item | Hero, SeasonPassReward, SeasonPass, ArenaScoreReward, ArenaSeason, DropItem |

## 保护期规则

- 战令武将：SeasonPass.StartTime + protectMonths（默认4个月）
- 大将军武将：ArenaSeason.SeasonEndTime（赛季结束即保护期结束）

## 关键共享函数

| 函数 | 位置 | 说明 |
|------|------|------|
| getHeroDropValidDate | hero_drop.go | 从 DropItem 表获取 ValidDate（被 item_synthesis.go 复用） |
| getNextThursday5AM | hero_drop_validdate.go | 计算下一个周四 5:00 |

## 开发注意事项

- ArenaScoreReward 表名是单数无 s
- hero_drop.go 的 getHeroDropValidDate 被 item_synthesis.go 复用
- 保护期相关规则的数据流注释已标注在各文件顶部
- 规则只检查 HeroType=1 的普通武将，跳过特殊武将
