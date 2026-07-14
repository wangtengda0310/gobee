# IPFS Panel - IPFS 分布式存储抽屉面板

> File path: `src/pages/settings/components/ipfs-panel.vue`
> Composable: `src/pages/settings/composables/use-ipfs.ts`

## Overview

以右侧抽屉（n-drawer, 500px）展示 IPFS 文件上传、下载和历史记录功能。
基于 Helia（浏览器端 IPFS 实现），支持 Kubo Pin、Remote Pinning 和 P2P 网络调试。

## ASCII Layout Diagram

```
┌─── IPFS 分布式存储 ──────────────────────────────────── 500px ──┐
│                                                                    │
│  ┌── ⚠ 安全提示 ──────────────────────────────── [n-alert] ───┐  │
│  │ 上传到 IPFS 的内容 CID 是公开可访问的，请勿上传敏感数据    [x]│  │
│  └────────────────────────────────────────────────────────────┘  │
│                                                                    │
│  ┌── 节点控制 ──────────────────────────────────── [n-card] ───┐  │
│  │ 状态: [运行中]  连接: 7/10  Kubo: [在线]       [停止节点] │  │
│  │ WebRTC + WebSocket P2P 模式，连接上限 10，上传限制 10MB   │  │
│  │ Kubo API: [http://127.0.0.1:5001___________] [检测]      │  │
│  │ 目标节点: [已连接] /ip4/47.100.180.176/tcp/4001/ws/...    │  │
│  └────────────────────────────────────────────────────────────┘  │
│                                                                    │
│  ┌── 文件上传 ──────────────────────────────────── [n-card] ───┐  │
│  │ [选择文件]  未选择文件 / filename.txt (1.5 KB)            │  │
│  │ Pin 到: [Kubo (自动)]  ☐ Remote Pinning  [配置]          │  │
│  │ [上传到 IPFS]                                             │  │
│  │ ──────────────────────────────────────────────────────── │  │
│  │ 上传结果:                                                 │  │
│  │ bafkreidqpozhofddpmcoibweowcs5g2ay...                    │  │
│  │ [复制 CID]  [在浏览器打开]                                │  │
│  └────────────────────────────────────────────────────────────┘  │
│                                                                    │
│  ┌── CID 下载 ──────────────────────────────────── [n-card] ───┐  │
│  │ [输入 CID (如 QmXyz... 或 bafy...)______] [下载]          │  │
│  │ 本地节点优先读取，未命中则回退 HTTP 网关                   │  │
│  │ ──────────────────────────────────────────────────────── │  │
│  │ 大小: 1.5 KB                                              │  │
│  │ 内容预览:                                                 │  │
│  │ ┌──────────────────────────────────────────────────────┐ │  │
│  │ │ 文本内容预览区域 (max 200px)                          │ │  │
│  │ └──────────────────────────────────────────────────────┘ │  │
│  │ [保存到本地]                                              │  │
│  └────────────────────────────────────────────────────────────┘  │
│                                                                    │
│  ┌── 上传历史 (3) ──────────────────── [清空] [全部] [n-card] ─┐  │
│  │ [成功] [Kubo] bafkreidqpo... 新建文件.txt 215B  07:09:59 │  │
│  │ ──────────────────────────────────────────────────────── │  │
│  │ (折叠区: 其余历史记录，点击"全部"/"收起"切换)             │  │
│  │ [成功] [Kubo] bafkreielz... file2.md 1.6KB  07:24:01    │  │
│  │ [成功] [Kubo] bafkreiaxr... file3.txt 927B  07:23:21    │  │
│  └────────────────────────────────────────────────────────────┘  │
│                                                                    │
│  ┌── 网络调试 [实验性] ──────────────────── [展开] [n-card] ──┐  │
│  │ (折叠区，点击展开/收起)                                   │  │
│  │ PeerID: 12D3KooWGcFm3E9SQWTeCc4zfuLQwuPZxzLXFRKN...    │  │
│  │ Bootstrap: [未连接]                                       │  │
│  │ 协议: WebSocket: 12                                       │  │
│  │ DHT Provide: success                                      │  │
│  │ 已连接 Peers (12):                                        │  │
│  │ ┌──────────────────────────────────────────────────────┐ │  │
│  │ │ 12D3KooW... /dns4/34-88-248-98... [WebSocket]       │ │  │
│  │ │ 12D3KooW... /dns4/80-241-211-3... [WebSocket]       │ │  │
│  │ │ ...                                                   │ │  │
│  │ └──────────────────────────────────────────────────────┘ │  │
│  └────────────────────────────────────────────────────────────┘  │
│                                                                    │
│  ─────────────────────────────────────────────────────────────── │
│  Kubo: 在线 | Remote Pin: 关闭 | 上传限制: 10MB                  │
└──────────────────────────────────────────────────────────────────┘
```

## Component Tree Structure

```
pages/settings/components/ipfs-panel.vue
├── n-drawer (placement="right", width=500)
│   └── n-drawer-content
│       ├── template #header → "IPFS 分布式存储"
│       ├── n-alert (安全提示, type=warning, closable)
│       │
│       ├── n-card (节点控制, size=small)
│       │   ├── .status-row → 状态/连接/Kubo 标签 + 启停按钮
│       │   ├── .hint-text → 模式/限制说明
│       │   ├── .kubo-api-row → Kubo API 地址输入 + 检测按钮
│       │   └── .target-peer-row → 目标节点连接状态
│       │
│       ├── n-card (文件上传, size=small)
│       │   ├── input[type=file] (hidden)
│       │   ├── .file-select-row → 选择文件按钮 + 文件名
│       │   ├── .pin-options-row → Kubo(自动) + Remote Pinning 复选框 + 配置按钮
│       │   ├── n-button (上传到 IPFS)
│       │   └── .upload-result → CID 显示 + 复制/打开按钮
│       │
│       ├── n-modal (Remote Pinning 配置弹窗)
│       │   ├── n-input (Endpoint URL)
│       │   ├── n-input (Access Token, type=password)
│       │   └── .hint → 兼容说明
│       │
│       ├── n-card (CID 下载, size=small)
│       │   ├── .input-row → CID 输入框 + 下载按钮
│       │   ├── .hint → 策略说明
│       │   ├── n-alert (错误提示, v-if=downloadError)
│       │   └── .download-result → 大小 + 内容预览 + 保存按钮
│       │
│       ├── n-card (上传历史, size=small)
│       │   ├── template #header → "上传历史" + 记录数
│       │   ├── template #header-extra → 清空 + 全部/收起按钮
│       │   ├── .latest-record → 最新一条记录（始终显示）
│       │   └── .collapsed-records → 其余记录（展开时显示）
│       │
│       ├── n-card (网络调试, size=small)
│       │   ├── template #header → "网络调试" + 实验性标签
│       │   ├── template #header-extra → 展开/收起按钮
│       │   └── .debug-content → PeerID/Bootstrap/协议/DHT/Peer列表
│       │
│       └── template #footer → 底部状态栏
```

## Component File Mapping

| Component | File Path | Description |
|-----------|-----------|-------------|
| IPFS Panel | pages/settings/components/ipfs-panel.vue | 抽屉面板 UI |
| IPFS Logic | pages/settings/composables/use-ipfs.ts | Helia 节点管理、上传/下载逻辑 |

## Key State (use-ipfs.ts)

### 节点状态

| State | Type | Description |
|-------|------|-------------|
| `ipfsNodeRunning` | Ref\<boolean\> | P2P 节点运行状态 |
| `ipfsConnectionCount` | Ref\<number\> | 当前连接数 |
| `ipfsPeerId` | Ref\<string\> | 本节点 PeerID |
| `ipfsPeers` | Ref\<PeerInfo[]\> | 已连接 Peer 列表 |
| `ipfsProtocols` | Ref\<Record\<string, number\>\> | 协议统计 |
| `bootstrapConnected` | Ref\<boolean\> | Bootstrap 连接状态 |
| `targetPeerAddr` | Ref\<string\> | 目标 Kubo 节点 P2P 地址 |
| `targetPeerConnected` | Ref\<boolean\> | 目标节点连接状态 |

### Kubo 状态

| State | Type | Description |
|-------|------|-------------|
| `kuboAvailable` | Ref\<boolean\> | Kubo daemon 在线状态 |
| `kuboApiUrl` | Ref\<string\> | Kubo HTTP API 地址（可在 UI 中配置） |
| `kuboChecking` | Ref\<boolean\> | Kubo 状态检测中 |

### Remote Pinning 状态

| State | Type | Description |
|-------|------|-------------|
| `remotePinConfig` | Ref\<RemotePinningConfig\> | Remote Pinning 配置（endpoint + token） |
| `remotePinEnabled` | Ref\<boolean\> | Remote Pinning 开关 |

### 上传/下载状态

| State | Type | Description |
|-------|------|-------------|
| `ipfsUploading` | Ref\<boolean\> | 上传中 |
| `ipfsDownloading` | Ref\<boolean\> | 下载中 |
| `ipfsUploadHistory` | Reactive\<IpfsUploadRecord[]\> | 上传历史（持久化到 localStorage） |
| `lastDhtProvideStatus` | Ref\<string\> | 最近一次 DHT provide 状态 |

## Constants

| Constant | Value | Description |
|----------|-------|-------------|
| `MAX_CONNECTIONS` | 10 | libp2p 最大连接数 |
| `MAX_UPLOAD_SIZE` | 10MB | 文件上传大小限制 |
| `UPLOAD_HISTORY_KEY` | 'ipfs-upload-history' | localStorage 持久化键 |

## Data Flow

### 文件上传流程

```
User selects file → handleFileSelect()
    │
    ▼
User clicks upload → handleUpload()
    │
    ▼
uploadFile(file)  [composable]
    │
    ├──► fsInstance.addBytes(bytes)          # Helia 上传到本地 blockstore
    │         │
    │         ▼
    │    heliaInstance.pins.add(cid)          # Helia pin
    │
    ├──► pinToKubo(bytes, filename)           # Kubo /api/v0/add?pin=true&cid-version=1
    │         │
    │         ├──► kuboAvailable=true → FormData POST → return true/false
    │         └──► kuboAvailable=false → skip
    │
    ├──► pinToRemoteService(cid)              # Remote Pinning (optional)
    │         │
    │         ├──► remotePinEnabled=true → POST /pins → return true/false
    │         └──► remotePinEnabled=false → skip
    │
    └──► provideToDht(cid)                   # DHT 广播 (experimental, non-blocking)
              │
              └──► helia.routing.provide(cid)
```

### 文件下载流程

```
User inputs CID → handleDownload()
    │
    ▼
downloadByCid(cidStr)  [composable]
    │
    ├──► downloadFromLocalNode(cid)          # 优先本地节点 (5s timeout)
    │         │
    │         ├──► fsInstance.cat(cid) → return content
    │         └──► fail/timeout → return null
    │
    └──► downloadFromGateway(cid)            # 回退 HTTP 网关 (30s timeout)
              │
              └──► createHeliaHTTP + trustlessGateway → fs.cat(cid)
```

### 节点启动流程

```
User clicks "启动节点" → handleToggleNode()
    │
    ▼
startP2PNode()  [composable]
    │
    ├──► new IDBBlockstore('ipfs-blocks') → blockstore.open()
    ├──► new IDBDatastore('ipfs-data')    → datastore.open()
    │
    ├──► createHelia({                      # 尝试完整 WebRTC 配置
    │         blockstore, datastore,
    │         transports: [circuitRelay, webSockets, webRTC],
    │         bootstrap: [5 public nodes],
    │         connectionGater: { denyDialMultiaddr }
    │     })
    │     └──► fail → createHelia(默认配置)  # 回退
    │
    ├──► unixfs(helia) → fsInstance
    ├──► connectToTargetPeer()               # 连接指定 Kubo 节点
    ├──► checkKuboAvailable()                # 检测 Kubo API
    └──► startConnectionMonitor()            # 定时更新网络状态
```

## Related Files

### Frontend
| File | Description |
|------|-------------|
| `pages/settings/components/ipfs-panel.vue` | IPFS 面板 UI 组件 |
| `pages/settings/composables/use-ipfs.ts` | IPFS 核心逻辑 |
| `pages/settings/index.vue` | 设置页面（包含 IPFS 入口按钮） |

### Dependencies
| Package | Version | Description |
|---------|---------|-------------|
| `helia` | ^6.1.3 | 浏览器端 IPFS 实现 |
| `@helia/unixfs` | ^7.2.1 | UnixFS 文件操作 |
| `@helia/http` | ^3.1.3 | HTTP 网关下载 |
| `@helia/remote-pinning` | ^3.0.0 | Remote Pinning Service API |
| `@helia/block-brokers` | ^5.2.3 | Trustless Gateway block broker |
| `@libp2p/circuit-relay-v2` | ^4.2.2 | Circuit Relay 传输 |
| `blockstore-idb` | ^4.0.3 | IndexedDB blockstore 后端 |
| `datastore-idb` | ^4.0.3 | IndexedDB datastore 后端 |
| `multiformats` | ^13.4.2 | CID 解析 |

### External Resources
| Resource | Description |
|----------|-------------|
| `http://127.0.0.1:5001` (configurable) | Kubo HTTP API（需 SSH 端口转发或本机 daemon） |
| `http://ipfs.itsnot.fun` | Kubo 网关（Nginx 代理） |
| `/ip4/47.100.180.176/tcp/4001/ws` | 目标 Kubo P2P 节点 |

## Notes

- 节点启动优先尝试 WebRTC+WebSocket 配置，失败回退到默认（纯 WebSocket）
- Kubo API 地址可在 UI 中配置（默认 `http://127.0.0.1:5001`）
- 上传历史持久化到 localStorage，重启不丢失
- 上传历史默认折叠，只显示最新一条；点击"全部"展开
- 网络调试面板为实验性功能，默认折叠
- 组件销毁时自动停止 P2P 节点（onUnmounted）
- 连接数超过 MAX_CONNECTIONS 时 connectionGater 拒绝新连接
