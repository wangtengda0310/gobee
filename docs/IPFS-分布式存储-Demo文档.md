# IPFS 分布式存储 Demo 文档

> **边界说明**：本文档聚焦于 **Helia 浏览器端 SDK** 的知识整理。
> Kubo CLI 服务端操作请参考 IPFS Skill (`~/.claude/skills/ipfs/SKILL.md`)。

## 一、IPFS 与 Helia 概述

IPFS 是分布式文件系统协议，核心特性：内容寻址（CID）、去中心化、不可篡改。

Helia 是 IPFS 的 TypeScript 实现（js-IPFS 继任者）。两种模式：

| 模式 | 包名 | P2P | 上传 | 下载 |
|------|------|-----|------|------|
| 完整 P2P | `helia` | 有 | 支持 | 支持 |
| 纯 HTTP | `@helia/http` | 零 | 不支持 | 支持 |

## 二、核心 API 用法

详见 [Helia API 文档](https://ipfs.github.io/helia/)。

关键 API：
- `createHelia({ connectionManager: { maxConnections: 3 } })` — 创建 P2P 节点
- `createHeliaHTTP({ blockBrokers: [trustlessGateway()] })` — 创建纯 HTTP 节点
- `fs.addBytes(data)` — 上传，返回 CID
- `fs.cat(cid, { signal })` — 下载，返回 AsyncIterable
- `helia.pins.add(cid)` — 固定数据
- `CID.parse(string)` — 解析 CID
- `helia.stop()` — 释放资源

## 三、NPM 包

核心包：`helia`（P2P）、`@helia/http`（纯 HTTP）、`@helia/unixfs`、`vite-plugin-node-polyfills`（Vite 构建必需）。

## 四、连接管理和安全注意事项

### 4.1 连接数限制

`connectionManager.maxConnections` 是修剪机制（非硬限制），建议配合运行时前置检查实现双重保护。

### 4.2 内存管理

- PeerStore 不持久化（避免内存膨胀），使用 MemoryBlockstore
- 及时调用 `helia.stop()` 释放资源
- 通过 `helia.libp2p.getConnections()` 监控连接数

### 4.3 数据安全

- CID 公开可访问，不要上传敏感数据
- 数据一旦被复制不可删除

## 五、Vite 构建注意事项

helia 和 libp2p 依赖 Node.js 内置模块，需要使用 `vite-plugin-node-polyfills`（`vite.config.ts` 中配置）。

已知问题：`vite dev` 下正常，`vite build` 后 polyfill 可能失效。

## 六、本项目实现说明

### 架构

```
use-ipfs.ts (Composable)
  ├── startP2PNode()   → createHelia()        [上传模式]
  ├── stopNode()       → helia.stop()
  ├── uploadFile()     → fs.addBytes() + pins  [P2P 模式]
  └── downloadByCid()  → createHeliaHTTP()     [HTTP 网关模式]
       ↓
ipfs-panel.vue (Drawer 组件)
  ├── 节点控制区
  ├── 文件上传区
  ├── CID 下载区
  └── 上传历史区
```

### 安全措施

1. `maxConnections: 3` 配置级限制
2. 上传前连接数前置检查
3. 每 5 秒监控连接数更新 UI
4. 下载使用 HTTP 网关，零 P2P 连接
5. 10MB 上传大小限制
6. CID 公开性安全提示

### 参考

- [Helia API 文档](https://ipfs.github.io/helia/)
