# activity-wiki-check -- 活动 Wiki 检查页面

路由 `/ActivityWiki`。展示活动 Wiki 数据，支持抽皮肤等子功能。

## 文件清单

| 文件 | 职责 |
|------|------|
| `index.vue` | 页面入口，组装活动列表 + 详情面板 |
| `components/activity-panel.vue` | 活动详情面板（数据展示、子功能入口） |
| `components/DrawSkinCard.vue` | 抽皮肤卡片组件 |
| `components/DrawPetCard.vue` | 结缘亭卡片组件 |
| `composables/use-activity-wiki.ts` | 核心逻辑：活动数据获取与展示 |

## 角标颜色机制

BadgeLabel 组件通过 `provide/inject` 获取 `ruleCoverage` 数据：
- **绿色角标**：字段有规则覆盖且校验通过（`count > 0 && errorCount === 0`）
- **红色角标**：字段有规则覆盖且校验不通过（`errorCount > 0`）
- **无角标**：字段无规则覆盖（`count === 0`）
- **角标数字**：始终显示关联规则数（`count`），不显示错误数

### 数据流

1. 配表测试页面执行检查 → `check_manager.StoreCheckResults()` 缓存结果到后端内存
2. 活动 Wiki 执行检查 → `GetRuleCoverageWithErrors()` 从缓存读取错误计数并合并
3. 无缓存时角标全绿（默认状态）

### 传递链

```
index.vue
  → useActivityWikiCheck()
  → ruleCoverage ref
  → ActivityPanel (:rule-coverage prop)
  → provide('ruleCoverage')
  → BadgeLabel (inject)
```

## 关键依赖

- `@bindings/` -- Wails 生成的 Go 后端 bindings

## 活动分组配置（参考 rain-excel-checker/activity/group.go）

共 **12 个子系统**，涵盖活动相关的全部配表。分组与 Sheet 是多对多关系。

| # | 分组名称 | 包含 Sheet |
|---|----------|-----------|
| 1 | 活动核心框架 | Activity、ActivityKey、ActivityLogin、ActivityTask |
| 2 | 赛季战令系统 | SeasonPass、SeasonPassBag、SeasonPassReward、SeasonPassTask、ArenaSeason、ArenaSeasonLimitedHero |
| 3 | 千里单骑 | PveRougeLevel、PveRougeBattleNode、PveRougeEventNode、PveRougeNodeBuff、PveRougeNodeEnd、PveRougeReward |
| 4 | Boss 系统 | PveBossStage、PveBossReward、PveBossRankReward、NianShouLaiXi、BossReward |
| 5 | 竞技对战系统 | HeroWar、HeroWarSeason、CityWarSeason、CityWarSeasonRankReward、PeakArena、PeakRankReward、PeakWinsReward |
| 6 | 充值消费活动 | AccumulatedRecharge、FirstRecharge、MonthlyCard、PioneerPlan、HundredDraw |
| 7 | 公会系统 | GuildWeekRankAwards、GuildCityWarReward、GuildCityLevel、GuildCityCreate、GuildRedPack |
| 8 | 签到任务系统 | SignIn、Task、TaskCompleteCond、WechatTask |
| 9 | 限时活动礼包 | GiftLimitShop、GiftLimitShopGoods、LimitTimeSkinDraw、ItemTimeLimitResolve |
| 10 | 其他玩法活动 | ShiRiYuXi、DailyAsk、TongPao、RushHeroRank、PveNewbieChallenge |
| 11 | 丹青阁 | Draw、Drop、Item、Shop |
| 12 | 山河争锋 | CityWar、Country、AIRobotInfo |

> 注意：`rain-excel-checker/activity/` 目录为遗留代码，未被引用可删除。分组数据已归档在此处供后续开发参考。

## 开发注意

- 新增活动类型展示：在 activity-panel.vue 中添加对应条件分支
- 扩展活动数据结构：修改 use-activity-wiki.ts 和对应后端 binding

## E2E 测试

| 测试文件 | Page Object | 覆盖范围 |
|----------|-------------|----------|
| `e2e/activity-wiki/activity-wiki.spec.ts` | [`ActivityWikiPage`](../../e2e/shared/pages/ActivityWikiPage.ts) | 路径配置（Excel 目录、历史 JSON）、累充活动卡片（卡片列表/类型标签）、累充奖励页签（页签显示/关联说明/奖励表格/数据验证/页签切换/打开 Excel）、页面布局 |

