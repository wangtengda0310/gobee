# WebTorrent P2P 文件传输 Demo - 实现详解

## 概述

在 rain-qa-func 应用中新增 `/WebTorrent` 独立页面，演示基于 WebTorrent 的浏览器端 P2P 文件传输功能。包括文件做种（上传）、Magnet URI 下载、拖拽上传、传输状态监控。

## WebTorrent 核心知识

WebTorrent 是浏览器端的 BitTorrent 客户端，使用 **WebRTC Data Channels** 传输。浏览器端 "web peer" 只能连接支持 WebRTC 的客户端。

核心 API：
- `client.seed(files)` — 做种上传，回调获得 `torrent.magnetURI`
- `client.add(magnetURI)` — 下载，完成后通过 `file.blob()` 获取内容
- `client.destroy()` — 页面卸载时释放资源

常用 Tracker：`wss://tracker.btorrent.xyz`、`wss://tracker.openwebtorrent.com`

限制：Web Peer 隔离、自动做种、数据不持久化（内存存储）、需监听 `error` 事件。

## 页面布局

页面采用左右分栏：左侧做种（上传），右侧 Magnet 下载，底部全局统计栏。

## 文件结构

```
pages/settings/components/webtorrent/
├── index.vue                    # 主页面 - 左右分栏布局
├── seed-panel.vue               # 做种面板 - 文件选择/拖拽/magnet 显示
├── download-panel.vue           # 下载面板 - magnet 输入/进度/保存
└── torrent-stats-bar.vue        # 底部全局统计栏

pages/settings/composables/
├── use-webtorrent.ts            # 客户端生命周期管理（创建/销毁/全局统计）
├── use-seed.ts                  # 做种逻辑（文件选择/拖拽/magnet 生成/移除）
└── use-download.ts              # 下载逻辑（添加/进度追踪/保存文件/取消）
```

## 路由与导航

- 路由路径: `/WebTorrent`
- 在 `router/index.ts` 中注册
- 在 `App.vue` 导航栏添加按钮

## 核心流程

### 做种流程

1. 用户选择文件或拖拽文件到上传区域
2. 调用 `client.seed(files)` 开始做种
3. 回调中获得 `torrent.magnetURI`
4. 显示 magnet URI，支持一键复制
5. 用户可手动移除做种（调用 `client.remove(torrentId)`）

### 下载流程

1. 用户粘贴 magnet URI
2. 调用 `client.add(magnetURI)` 开始下载
3. 监听 `torrent.on('download')` 实时更新进度
4. `torrent.on('done')` 后通过 `file.blob()` 生成下载链接
5. 用户点击保存，通过 `<a>` 标签触发浏览器下载

### 生命周期

- 页面 `onMounted` 时创建 WebTorrent 客户端
- 页面 `onUnmounted` 时销毁客户端，释放所有资源

## 技术选型

| 项 | 选择 | 原因 |
|----|------|------|
| UI 库 | Naive UI | 与现有页面保持一致 |
| 包引入 | `npm install webtorrent` | 需配合 vite-plugin-node-polyfills |
| 状态管理 | Vue Composables | 遵循项目现有模式 |
| 文件下载 | Blob + URL.createObjectURL | 浏览器端标准做法 |

## 参考资源

- [WebTorrent 官网](https://webtorrent.io/)
- [WebTorrent GitHub](https://github.com/webtorrent/webtorrent)
- [WebTorrent API 文档](https://webtorrent.io/docs)
