# shared -- 跨页面共享资源

所有页面共用的组件、composable、配置和工具函数。页面间禁止直接引用，通过本目录中转。

## 目录结构

### components/ -- 通用组件

| 组件 | 职责 |
|------|------|
| `column-rule-list/` | 列规则列表（配表检查页使用） |
| `header-content-footer/` | 三段式布局（头部 + 内容 + 底部） |
| `page-menu-bar/` | 页面菜单栏 |
| `path-config-input/` | 路径配置输入框（Excel 目录 + 可配置第二路径） |
| `status-bar/` | 底部状态栏（版本、连接状态） |
| `table-rule-panel/` | 表级规则面板 |
| `tooltip-checkbox/` | 带 tooltip 的复选框 |
| `tree-side-nav/` | 树形侧边导航 |

### composables/ -- 通用 composable

| 文件 | 职责 |
|------|------|
| `use-tree-*.ts` | 树组件相关（状态、搜索、右键菜单、下拉、工具） |
| `use-format-utils.ts` | 格式化工具 |
| `use-open-excel.ts` | 打开 Excel 文件 |
| `use-rule-badge.ts` | 规则标签样式 |

### config/ -- 共享配置

| 文件 | 职责 |
|------|------|
| `hero.ts` | 武将配置（function-test 和 hero-voice-resource-check 共享） |

### utils/ -- 工具函数

| 文件 | 职责 |
|------|------|
| `format.ts` | 通用格式化函数 |

### polyfills/ -- 浏览器 shim

| 文件 | 职责 |
|------|------|
| `bittorrent-dht.js` | WebTorrent DHT 模块浏览器 shim |
| `empty.js` | 空模块占位（Node.js 专用包的浏览器替代） |

`polyfills.ts` 负责按需加载上述 shim，仅在 WebTorrent/IPFS 功能启用时生效。

## 开发注意

- 新增共享组件必须放入对应子目录，index.vue 作为入口
- 通过 `@shared/` alias 引用，不使用相对路径
