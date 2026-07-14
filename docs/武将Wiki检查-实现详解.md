# 武将 Wiki 检查页面 - 实现详解

本文档详细说明武将 Wiki 检查页面的实现架构和关键技术点。

## 文件结构

```
frontend/src/
├── pages/
│   └── HeroWikiCheck.vue          # 主页面
├── components/HeroWiki/
│   ├── HeroPanel.vue              # 武将面板组件
│   ├── HeroDiffDisplay.vue        # 差异显示组件
│   ├── BuffDisplay.vue            # Buff 展示组件
│   └── DropDisplay.vue            # 掉落展示组件
└── scripts/HeroPanel/
    └── TypeDef.ts                 # 类型定义
```

## 数据流

```
┌─────────────────────────────────────────────────────────────────┐
│                      HeroWikiResCheckService                    │
│              (Go 后端服务，通过 Wails 绑定)                       │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                     DataContainer (diffExcels)                  │
│  ├── HeroDiff[]              # 武将差异列表                      │
│  ├── SkillDiff[]             # 技能差异列表                      │
│  ├── SkillLineDiff[]         # 技能台词差异                      │
│  ├── HeroLineDiff[]          # 武将台词差异                      │
│  ├── CountryDiff[]           # 国家差异列表                      │
│  ├── HeroSkinItemDiff[]      # 皮肤差异列表                      │
│  └── HeroWikiDiffResult      # Wiki 检查结果                     │
│       ├── HeroesDiff         # Map<eHeroId, HeroWikiDiff>       │
│       ├── Summary            # 统计摘要                          │
│       │   ├── AddedHeroes    # 新增武将 EHeroId 列表             │
│       │   ├── RemovedHeroes  # 删除武将 EHeroId 列表             │
│       │   └── ModifiedHeroes # 修改武将 EHeroId 列表             │
│       └── RemovedHeroesData  # 删除武将的完整数据                │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    filteredHeroes (计算属性)                     │
│              根据筛选条件过滤后的武将列表                          │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      HeroPanel 组件                              │
│              显示单个武将的详细信息                               │
└─────────────────────────────────────────────────────────────────┘
```

## 核心组件

### HeroWikiCheck.vue - 主页面

#### 页面布局

详见 [frontend/docs/layout/pages/hero-wiki-check/index.md](frontend/docs/layout/pages/hero-wiki-check/index.md)。

#### 关键状态变量

- `diffExcels` - 数据源（`DataContainer`）
- `searchName` - 名称搜索
- `filterCountry` - 势力筛选（多选）
- `filterIsNewHero` / `filterIsGacha` / `filterIsOpen` - 布尔三态筛选
- `filterDiffType` - diff 类型筛选（`'added' | 'modified' | 'removed' | null`）

筛选逻辑：`[filteredHeroes](frontend/src/pages/HeroWikiCheck.vue)` 计算属性，依次判断名称、势力、布尔值、diff 类型。

#### 删除武将处理

删除武将不在 `HeroDiff` 列表中，需从 `RemovedHeroesData` 构建虚拟 `HeroDiff` 对象。关键点：使用 `removed-` 前缀和 `EHeroId` 作为面板 ID，避免与正常武将数字 ID 冲突。

### HeroPanel.vue - 武将面板组件

#### Props 定义

- `seq` - 序号
- `heroInfo` - 武将差异数据
- `diffExcels` - 完整数据容器
- `diffIndexMap` - 索引映射（用于关联查询）
- `heroWikiData` - 完整 Wiki 数据（删除武将专用）
- `isRemoved` - 是否为删除武将

面板 ID 规则：正常武将 `HeroId:{Id}`，删除武将 `HeroId:removed-{EHeroId}`。

删除武将样式：红色删除线标题 + "已删除" 标签 + `removed-hero` CSS 类。

### DiffIndexMap - 索引映射

用于快速查询关联数据：`[transMap()](frontend/src/pages/HeroWikiCheck.vue)` 构建 skillsDiffMap、skillLinesDiffMap、heroLinesDiffMap、countryDiffMap、heroSkinItemDiffMap。

## 样式规范

颜色和国家/类型映射定义见 `[hero-panel-utils.ts](frontend/src/pages/hero-wiki-check/components/hero-panel-utils.ts)`。

### 导航名字标黄逻辑

当武将存在于 `HeroesDiff` 中时（有变更），导航中该武将的名字显示为黄色。见 `[HeroWikiCheck.vue](frontend/src/pages/HeroWikiCheck.vue)` 导航样式绑定。

## 相关文档

- [过滤条件扩展指南](武将Wiki检查-过滤条件扩展指南.md)
- 武将标黄的逻辑 — 见上文「标黄逻辑」章节
