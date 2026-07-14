# frontend/src 前端代码

> 代码按功能内聚组织，每个菜单项对应的代码自包含在 `pages/` 子目录下。

## 目录结构

```
src/
├── main.ts                 # 应用入口（createApp + mount）
├── App.vue                 # 根组件（布局 + 菜单 + 路由配置 + 路由视图）
├── shared/                 # 可被多个pages/下的页面跨页面共享的资源
│   ├── components/         # 通用组件（PathConfigInput、StatusBar、TreeSideNav 等）
│   ├── composables/        # 通用 composable（tree、format、rule-badge 等）
│   ├── config/
│   │   └── hero.ts         # 武将配置（跨 function-test 和 hero-voice-resource-check 共享）
│   ├── utils/              # 工具函数
│   ├── polyfills/          # WebTorrent 浏览器 shim
│   │   ├── bittorrent-dht.js  # DHT 模块浏览器替代
│   │   └── empty.js           # 空模块占位（Node.js 专用包）
│   └── polyfills.ts        # 按需加载入口（仅 WebTorrent/IPFS 启用时生效）
├── pages/
│   ├── settings/           # 设置页（/Settings）
│   │   ├── components/
│   │   │   ├── roadmap/    # 路线图组件（RoadmapDetail、SubmitModal）
│   │   │   ├── webtorrent/ # P2P 文件传输组件
│   │   │   └── ...         # intercept-notification、server-log-panel 等
│   │   ├── composables/    # use-server-logs、use-webtorrent、use-download 等
│   │   └── config/         # roadmap-api.ts、roadmap-types.ts
│   ├── llm/                # AI助手（/Home）
│   │   ├── components/     # chat-config-panel、chat-message-item
│   │   └── composables/    # use-chat-state、chat.types
│   ├── function-test/      # 战斗测试（/Test）
│   │   ├── components/     # steps-panel、init-yanwu-panel、asset-sections 等
│   │   ├── composables/    # Func、AssetProtoOptions、excel-rules-template 等
│   │   └── config/         # Card、Skill、Identity、ECountry、ErrorCode
│   ├── excel-test/         # 配表测试（/Excel）
│   ├── hero-voice-resource-check/  # 武将资源检查（/HeroRes）
│   ├── hero-wiki-check/    # 武将Wiki检查（/HeroWikiRes）
│   └── activity-wiki-check/ # 活动Wiki（/ActivityWiki）
└── vite-env.d.ts
```

## 前端变更强制同步规范

修改前端页面代码时，**必须**同步更新文档和 E2E 测试。详见 [E2E 测试同步规范](../docs/e2e-sync-rules.md)。

## 开发约定

### 页面模块内聚

每个 `pages/` 子目录是一个完整的功能单元，包含该功能所需的：

- `index.vue` — 页面入口
- `components/` — 页面专属组件
- `config/` — 页面专属配置数据（如有）

### 跨页面共享

- **共享组件** → `shared/components/`
- **共享 composable** → `shared/composables/`
- **共享配置数据** → `shared/config/`
- 不在 `pages/` 之间直接引用，通过 `shared/` 中转

### import 规范

- 页面内部：使用相对路径（`../config/`、`./components/`）
- 跨页面引用共享资源：使用 `@shared/` alias
- 绝对路径引用页面内资源：使用 `@/pages/...`

### 路由与菜单

- 路由定义在 `App.vue` 的普通 `<script>` 块中（`<script setup>` 不允许 export）
- 菜单配置在 `App.vue` 的 `<script setup>` 块中（`menuItems` 数组）
- 新增页面需同时更新路由和菜单两处

## Alias 配置（vite.config.ts）

| Alias         | 路径            | 用途                     |
| ------------- | --------------- | ------------------------ |
| `@`         | `src/`        | 根路径引用               |
| `@pages`    | `src/pages/`  | 页面引用                 |
| `@shared`   | `src/shared/` | 共享资源引用             |
| `@bindings` | `bindings/`   | Wails 生成的 Go bindings |

## E2E 测试

- 测试目录：`frontend/e2e/`（按模块分子目录）
- 共享模块：`frontend/e2e/shared/`（fixtures、pages、utils、docs）
- [E2E 测试同步规范](../docs/e2e-sync-rules.md) — 前端变更与 E2E 测试的同步规则
- 完整目录结构和模块对照见 [e2e/CLAUDE.md](../e2e/CLAUDE.md)

### E2E 测试索引

e2e-index.md

## 更多文档

前端布局文档 — 各页面 ASCII 布局可视化、组件层次、数据流
截图注释文档 — 页面截图 UI 元素标注
