# hero-wiki-check -- 武将 Wiki 检查页面

路由 `/HeroWikiRes`。对比 Wiki 公开数据与服务端配表数据，按维度展示差异。

## 文件清单

| 文件 | 职责 |
|------|------|
| `index.vue` | 页面入口，组装列表 + 面板布局 |
| `components/hero-panel.vue` | 武将详情面板入口，tab 切换调度各子 tab |
| `components/hero-panel-utils.ts` | 面板工具函数（格式化、映射） |
| `components/hero-diff-display.vue` | 通用差异展示（新增/删除/变更标记） |
| `components/hero-basic-tab.vue` | 基础信息 tab |
| `components/hero-skills-tab.vue` | 技能 tab |
| `components/hero-skins-tab.vue` | 皮肤 tab |
| `components/hero-ui-tab.vue` | UI tab |
| `components/hero-country-tab.vue` | 阵营 tab |
| `components/hero-gacha-tab.vue` | 抽卡 tab |
| `components/hero-recommend-tab.vue` | 推荐 tab |
| `components/hero-robot-tab.vue` | 机器人 tab |
| `components/hero-achievements-tab.vue` | 成就 tab |
| `components/buff-display.vue` | Buff 展示（被多个 tab 引用） |
| `components/drop-display.vue` | 掉落展示（被多个 tab 引用） |
| `composables/hero-wiki.types.ts` | 类型定义（Wiki 数据结构、过滤参数） |
| `composables/use-hero-wiki.ts` | 核心逻辑：数据获取、过滤、差异计算 |

## 关键依赖

- `@bindings/` -- Wails 生成的 Go 后端 bindings
- `@shared/config/hero.ts` -- 武将配置

## 开发注意

- hero-panel.vue 是核心组件，通过 tab 切换展示不同维度的 Wiki 数据
- 新增检查维度：创建 `hero-xxx-tab.vue`，在 hero-panel.vue 注册 tab
- 详细布局见 `frontend/docs/layout/pages/hero-wiki-check/index.md`
- 过滤条件扩展见 `docs/武将Wiki检查-过滤条件扩展指南.md`

## E2E 测试

| 测试文件 | Page Object | 覆盖范围 |
|----------|-------------|----------|
| `e2e/hero-wiki-check/hero-wiki-check.spec.ts` | [`HeroWikiCheckPage`](../../e2e/shared/pages/HeroWikiCheckPage.ts) | 路径配置（Excel 目录、历史 JSON）、执行检查、统计标签（总变化/新增/删除/修改）、筛选功能（搜索武将、势力、新武将、抽奖、已开放）、武将面板（列表/展开/差异详情/Buff/掉落）、锚点导航、页面布局 |

