/**
 * IPFS 分布式存储管理 Composable
 *
 * 基于 Helia (IPFS TypeScript 实现) 提供分布式文件上传和访问功能。
 *
 * 上传策略：Helia P2P 节点 → 可选 Kubo Pin → 可选 Remote Pinning → DHT Provide
 * 下载策略：本地节点优先 → HTTP 网关回退
 *
 * 网络传输：WebRTC + WebSocket（显式配置），通过 Bootstrap 节点连接 IPFS 网络
 */
import { reactive, ref } from 'vue'
import { createHelia, type Helia } from 'helia'
import { createHeliaHTTP } from '@helia/http'
import { unixfs, type UnixFS } from '@helia/unixfs'
import { trustlessGateway } from '@helia/block-brokers'
import { httpGatewayRouting } from '@helia/routers'
import { CID } from 'multiformats/cid'
import { multiaddr } from '@multiformats/multiaddr'
import { webRTC } from '@libp2p/webrtc'
import { webSockets } from '@libp2p/websockets'
import { circuitRelayTransport } from '@libp2p/circuit-relay-v2'
import { bootstrap } from '@libp2p/bootstrap'
import { IDBBlockstore } from 'blockstore-idb'
import { IDBDatastore } from 'datastore-idb'
import { createRemotePins, type RemotePins } from '@helia/remote-pinning'

// ─── 常量 ──────────────────────────────────────────────────

const MAX_CONNECTIONS = 10
const MAX_UPLOAD_SIZE = 10 * 1024 * 1024 // 10MB
/** Kubo HTTP API 地址（可在 UI 中配置，如本机 http://127.0.0.1:5001 或远程 http://47.100.180.176:5001） */
export const kuboApiUrl = ref('http://127.0.0.1:5001')

/** 指定连接的 Kubo 节点 P2P 地址（可通过 UI 配置） */
export const targetPeerAddr = ref('/ip4/47.100.180.176/tcp/4001/ws/p2p/12D3KooWSKz7vParWykwgWfZbhTEGutvJSkxumCFGx5QEh4U7cnN')
export const targetPeerConnected = ref(false)

/** Kubo 默认的 Bootstrap 节点（WebSocket 地址，浏览器可用） */
const BOOTSTRAP_NODES = [
    '/dns4/ams-1.bootstrap.libp2p.io/tcp/443/wss/p2p/QmSoLer265NRgSp2LA3dPaeykiS1J6DifTC88f5uVQKNAd',
    '/dns4/lon-1.bootstrap.libp2p.io/tcp/443/wss/p2p/QmSoLMeWqB7YGVLJN3pNLQpmmEk35v6wYtsMGLzSr5QBU3',
    '/dns4/sfo-3.bootstrap.libp2p.io/tcp/443/wss/p2p/QmSoLPppuBtQSGwKDZT2M73ULpjvfd3aZ6ha4oFGL1KrGM',
    '/dns4/sgp-1.bootstrap.libp2p.io/tcp/443/wss/p2p/QmSoLSafTMBsPKadTEgaXctDQVcqN88CNLHXMkTNwMKPnu',
    '/dns4/nyc-1.bootstrap.libp2p.io/tcp/443/wss/p2p/QmSoLueR4xBeUbY9WZ9xGUUxunbKWcrNFTAYadPMTz4G5Q',
]

// ─── 类型定义 ──────────────────────────────────────────────

export interface IpfsUploadRecord {
    cid: string
    filename: string
    size: number
    timestamp: string
    status: 'success' | 'failed'
    kuboPinned?: boolean
    remotePinned?: boolean
}

export interface IpfsDownloadResult {
    cid: string
    content: Uint8Array
    size: number
    textContent?: string
}

export interface RemotePinningConfig {
    endpointUrl: string
    accessToken: string
}

/** 已连接的 Peer 信息 */
export interface PeerInfo {
    peerId: string
    addr: string
    protocols: string[]
}

// ─── 模块级单例状态 ────────────────────────────────────────

export const ipfsNodeRunning = ref(false)
export const ipfsUploading = ref(false)
export const ipfsDownloading = ref(false)
export const ipfsConnectionCount = ref(0)
export const ipfsLastError = ref<string | null>(null)
const UPLOAD_HISTORY_KEY = 'ipfs-upload-history'
export const ipfsUploadHistory = reactive<IpfsUploadRecord[]>(
    JSON.parse(localStorage.getItem(UPLOAD_HISTORY_KEY) || '[]')
)

/** 持久化上传历史到 localStorage */
function saveUploadHistory(): void {
    localStorage.setItem(UPLOAD_HISTORY_KEY, JSON.stringify(ipfsUploadHistory))
}
export const lastUploadCid = ref<string | null>(null)
export const lastDownloadResult = ref<IpfsDownloadResult | null>(null)

// Kubo 状态
export const kuboAvailable = ref(false)
export const kuboChecking = ref(false)

// Remote Pinning 状态
export const remotePinConfig = ref<RemotePinningConfig>({
    endpointUrl: '',
    accessToken: '',
})
export const remotePinEnabled = ref(false)

// WebRTC 调试状态
export const ipfsPeerId = ref<string>('')
export const ipfsPeers = ref<PeerInfo[]>([])
export const ipfsProtocols = ref<Record<string, number>>({})
export const bootstrapConnected = ref(false)
export const lastDhtProvideStatus = ref<string>('')

// ─── 内部变量 ──────────────────────────────────────────────

let heliaInstance: Helia | null = null
let fsInstance: UnixFS | null = null
let remotePinsInstance: RemotePins | null = null
let connectionMonitor: ReturnType<typeof setInterval> | null = null

// ─── 辅助函数 ──────────────────────────────────────────────

export function formatFileSize(bytes: number): string {
    if (bytes < 1024) return bytes + ' B'
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

export function getGatewayUrl(cid: string): string {
    return `https://ipfs.io/ipfs/${cid}`
}

export async function copyToClipboard(text: string): Promise<void> {
    await navigator.clipboard.writeText(text)
}

function setError(msg: string): void {
    ipfsLastError.value = msg
    console.error('[IPFS]', msg)
}

function clearError(): void {
    ipfsLastError.value = null
}

function concatenateUint8Arrays(chunks: Uint8Array[]): Uint8Array {
    const totalLength = chunks.reduce((acc, chunk) => acc + chunk.length, 0)
    const result = new Uint8Array(totalLength)
    let offset = 0
    for (const chunk of chunks) {
        result.set(chunk, offset)
        offset += chunk.length
    }
    return result
}

function tryDecodeAsText(bytes: Uint8Array): string | undefined {
    try {
        return new TextDecoder('utf-8', { fatal: true }).decode(bytes)
    } catch {
        return undefined
    }
}

/** 从连接地址中提取协议名 */
function extractProtocol(addr: string): string {
    if (addr.includes('/webrtc/')) return 'WebRTC'
    if (addr.includes('/ws') || addr.includes('/websockets/')) return 'WebSocket'
    if (addr.includes('/p2p-circuit/')) return 'Circuit Relay'
    if (addr.includes('/tcp/')) return 'TCP'
    return 'Unknown'
}

// ─── Kubo HTTP API ─────────────────────────────────────────

/**
 * 检测本机 Kubo daemon 是否可用
 */
export async function checkKuboAvailable(): Promise<boolean> {
    kuboChecking.value = true
    try {
        const controller = new AbortController()
        const timeout = setTimeout(() => controller.abort(), 3000)
        const resp = await fetch(`${kuboApiUrl.value}/api/v0/id`, {
            method: 'POST',
            signal: controller.signal,
        })
        clearTimeout(timeout)
        kuboAvailable.value = resp.ok
        return resp.ok
    } catch {
        kuboAvailable.value = false
        return false
    } finally {
        kuboChecking.value = false
    }
}

/**
 * 上传文件到 Kubo 并 pin
 *
 * 通过 /api/v0/add 将文件内容直接上传到 Kubo（不只是 pin CID）。
 * 因为 Helia 浏览器节点的 block 数据不会自动同步到 Kubo，
 * 必须显式传输文件内容。
 */
async function pinToKubo(fileBytes: Uint8Array, filename: string): Promise<boolean> {
    try {
        const apiUrl = `${kuboApiUrl.value}/api/v0/add?pin=true&cid-version=1`
        console.log('[IPFS] 正在上传到 Kubo:', apiUrl, '文件:', filename, '大小:', fileBytes.length)

        const formData = new FormData()
        formData.append('file', new File([new Uint8Array(fileBytes)], filename), filename)

        const resp = await fetch(apiUrl, {
            method: 'POST',
            body: formData,
        })
        if (!resp.ok) {
            console.warn('[IPFS] Kubo 上传失败:', resp.status, resp.statusText)
            return false
        }
        const result = await resp.json()
        console.log('[IPFS] Kubo 上传+pin 成功:', result.Name, '→', result.Hash)
        return true
    } catch (err) {
        console.warn('[IPFS] Kubo pin 失败（daemon 未运行?）:', err)
        return false
    }
}

// ─── Remote Pinning ────────────────────────────────────────

/**
 * 通过 Remote Pinning Service API pin CID
 */
async function pinToRemoteService(cidStr: string): Promise<boolean> {
    const config = remotePinConfig.value
    if (!config.endpointUrl || !config.accessToken) {
        console.warn('[IPFS] Remote Pinning 未配置')
        return false
    }

    try {
        // 直接调用 IPFS Pinning Service API
        const resp = await fetch(`${config.endpointUrl}/pins`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${config.accessToken}`,
            },
            body: JSON.stringify({ cid: cidStr }),
        })

        if (!resp.ok) {
            const body = await resp.text()
            console.warn('[IPFS] Remote Pin 失败:', resp.status, body)
            return false
        }

        console.log('[IPFS] Remote Pin 成功:', cidStr)
        return true
    } catch (err) {
        console.warn('[IPFS] Remote Pin 失败:', err)
        return false
    }
}

// ─── DHT Provide ───────────────────────────────────────────

/**
 * 向 DHT 广播自己拥有某个 CID（实验性）
 *
 * 浏览器端 KadDHT 默认 clientMode，效果有限但可用于调研。
 */
async function provideToDht(cidStr: string): Promise<void> {
    if (!heliaInstance) return

    try {
        lastDhtProvideStatus.value = 'providing...'
        const cid = CID.parse(cidStr)
        await heliaInstance.routing.provide(cid)
        lastDhtProvideStatus.value = 'success'
        console.log('[IPFS] DHT provide 成功:', cidStr)
    } catch (err) {
        lastDhtProvideStatus.value = `failed: ${err instanceof Error ? err.message : String(err)}`
        console.warn('[IPFS] DHT provide 失败:', err)
    }
}

// ─── 连接信息更新 ──────────────────────────────────────────

function updateNetworkInfo(): void {
    if (!heliaInstance?.libp2p) return

    // 连接数
    const conns = heliaInstance.libp2p.getConnections()
    ipfsConnectionCount.value = conns.length

    // Peer 列表
    const peers: PeerInfo[] = []
    const protocolCounts: Record<string, number> = {}
    let bsConnected = false

    for (const conn of conns) {
        const addrStr = conn.remoteAddr.toString()
        const peerId = conn.remotePeer.toString()
        const proto = extractProtocol(addrStr)
        const protocols = [proto]

        peers.push({ peerId, addr: addrStr, protocols })

        protocolCounts[proto] = (protocolCounts[proto] || 0) + 1

        // 检查是否连接了 Bootstrap 节点
        for (const bs of BOOTSTRAP_NODES) {
            if (bs.includes(peerId)) {
                bsConnected = true
            }
        }
    }

    ipfsPeers.value = peers
    ipfsProtocols.value = protocolCounts
    bootstrapConnected.value = bsConnected

    if (conns.length > MAX_CONNECTIONS) {
        console.debug('[IPFS] 连接数超过限制:', conns.length, '/', MAX_CONNECTIONS)
    }
}

/**
 * 连接到指定的 Kubo peer
 *
 * 通过 libp2p dial 建立直连，连接后可通过 Bitswap 交换数据。
 * 异步执行，不阻塞节点启动。带超时和重试。
 */
async function connectToTargetPeer(): Promise<void> {
    if (!heliaInstance || !targetPeerAddr.value) return

    const maxRetries = 3
    for (let i = 0; i < maxRetries; i++) {
        try {
            const controller = new AbortController()
            const timeout = setTimeout(() => controller.abort(), 15_000)
            await heliaInstance.libp2p.dial(multiaddr(targetPeerAddr.value), {
                signal: controller.signal,
            })
            clearTimeout(timeout)
            targetPeerConnected.value = true
            console.log('[IPFS] 已连接到指定 Kubo peer:', targetPeerAddr.value)
            return
        } catch (err) {
            const msg = err instanceof Error ? err.message : String(err)
            console.warn(`[IPFS] 连接 Kubo peer 失败 (${i + 1}/${maxRetries}):`, msg)
            if (i < maxRetries - 1) {
                await new Promise(r => setTimeout(r, 3000))
            }
        }
    }
    targetPeerConnected.value = false
}

// ─── 核心操作 ──────────────────────────────────────────────

/**
 * 启动 P2P 节点
 *
 * 显式配置 WebRTC + WebSocket 传输，连接 Bootstrap 节点。
 */
export async function startP2PNode(): Promise<void> {
    if (heliaInstance) {
        setError('节点已在运行中')
        return
    }

    clearError()

    try {
        // 持久化存储：blockstore 和 datastore 均使用 IndexedDB，需显式打开
        const blockstore = new IDBBlockstore('ipfs-blocks')
        const datastore = new IDBDatastore('ipfs-data')
        await blockstore.open()
        await datastore.open()

        // 尝试完整 WebRTC + WebSocket 配置
        try {
            heliaInstance = await createHelia({
                blockstore,
                datastore,
                libp2p: {
                    transports: [
                        circuitRelayTransport(),
                        webSockets(),
                        webRTC(),
                    ],
                    addresses: {
                        listen: ['/p2p-circuit', '/webrtc'],
                    },
                    connectionManager: {
                        maxConnections: MAX_CONNECTIONS,
                    },
                    connectionGater: {
                        denyDialMultiaddr: async () => {
                            if (!heliaInstance) return false
                            const conns = heliaInstance.libp2p.getConnections()
                            return conns.length >= MAX_CONNECTIONS
                        },
                    },
                    peerDiscovery: [
                        bootstrap({
                            list: BOOTSTRAP_NODES,
                            timeout: 30_000,
                        }),
                    ],
                },
            })
            console.log('[IPFS] 使用 WebRTC + WebSocket 传输')
        } catch (transportErr) {
            // WebRTC 在 Wails WebView 中可能不可用，回退到默认配置
            console.warn('[IPFS] WebRTC 传输初始化失败，回退默认配置:', transportErr)
            heliaInstance = await createHelia({
                blockstore,
                datastore,
                libp2p: {
                    connectionManager: {
                        maxConnections: MAX_CONNECTIONS,
                    },
                    connectionGater: {
                        denyDialMultiaddr: async () => {
                            if (!heliaInstance) return false
                            const conns = heliaInstance.libp2p.getConnections()
                            return conns.length >= MAX_CONNECTIONS
                        },
                    },
                    peerDiscovery: [
                        bootstrap({
                            list: BOOTSTRAP_NODES,
                            timeout: 30_000,
                        }),
                    ],
                },
            })
            console.log('[IPFS] 使用默认传输配置')
        }

        fsInstance = unixfs(heliaInstance)
        ipfsNodeRunning.value = true
        ipfsPeerId.value = heliaInstance.libp2p.peerId.toString()

        // 初始化 Remote Pinning（如果配置了）
        if (remotePinConfig.value.endpointUrl && remotePinConfig.value.accessToken) {
            remotePinsInstance = createRemotePins(heliaInstance, {
                endpointUrl: remotePinConfig.value.endpointUrl,
                accessToken: remotePinConfig.value.accessToken,
            })
        }

        // 启动网络信息监控
        connectionMonitor = setInterval(updateNetworkInfo, 3000)
        updateNetworkInfo()

        // 异步检测 Kubo
        checkKuboAvailable()

        // 连接指定的 Kubo peer（不阻塞启动）
        connectToTargetPeer()

        console.log('[IPFS] P2P 节点已启动, PeerID:', ipfsPeerId.value)
    } catch (err) {
        heliaInstance = null
        fsInstance = null
        setError(`启动 P2P 节点失败: ${err instanceof Error ? err.message : String(err)}`)
        throw err
    }
}

/**
 * 停止 P2P 节点
 */
export async function stopNode(): Promise<void> {
    if (connectionMonitor) {
        clearInterval(connectionMonitor)
        connectionMonitor = null
    }

    if (heliaInstance) {
        try {
            await heliaInstance.stop()
        } catch (err) {
            console.warn('[IPFS] 停止节点时出错:', err)
        }
        heliaInstance = null
        fsInstance = null
        remotePinsInstance = null
    }

    ipfsNodeRunning.value = false
    ipfsConnectionCount.value = 0
    ipfsPeerId.value = ''
    ipfsPeers.value = []
    ipfsProtocols.value = {}
    bootstrapConnected.value = false
    lastDhtProvideStatus.value = ''
    targetPeerConnected.value = false
    console.log('[IPFS] P2P 节点已停止')
}

/**
 * 上传文件到 IPFS
 *
 * 流程：Helia addBytes → Helia pin → Kubo pin（可选）→ Remote Pin（可选）→ DHT provide
 */
export async function uploadFile(file: File): Promise<string> {
    if (!ipfsNodeRunning.value || !heliaInstance || !fsInstance) {
        throw new Error('P2P 节点未启动，请先启动节点')
    }
    if (ipfsUploading.value) {
        throw new Error('正在上传中，请等待当前上传完成')
    }
    if (file.size > MAX_UPLOAD_SIZE) {
        throw new Error(`文件大小超过限制（最大 ${formatFileSize(MAX_UPLOAD_SIZE)}）`)
    }

    clearError()
    ipfsUploading.value = true

    let kuboPinned = false
    let remotePinned = false

    try {
        const buffer = await file.arrayBuffer()
        const bytes = new Uint8Array(buffer)

        // 1. 上传到 Helia
        const cid = await fsInstance.addBytes(bytes)
        const cidStr = cid.toString()

        // 2. Helia 本地 pin
        await heliaInstance.pins.add(cid)

        // 3. Kubo pin（如果可用）
        if (kuboAvailable.value) {
            kuboPinned = await pinToKubo(bytes, file.name)
        }

        // 4. Remote Pin（如果启用且配置了）
        if (remotePinEnabled.value && remotePinConfig.value.endpointUrl) {
            remotePinned = await pinToRemoteService(cidStr)
        }

        // 5. DHT provide（实验性，不阻塞）
        provideToDht(cidStr).catch(() => {})

        lastUploadCid.value = cidStr

        ipfsUploadHistory.unshift({
            cid: cidStr,
            filename: file.name,
            size: file.size,
            timestamp: new Date().toLocaleTimeString(),
            status: 'success',
            kuboPinned,
            remotePinned,
        })
        saveUploadHistory()

        console.log('[IPFS] 文件上传成功:', file.name, '→', cidStr,
            kuboPinned ? '(Kubo pinned)' : '',
            remotePinned ? '(Remote pinned)' : '')
        return cidStr
    } catch (err) {
        const errMsg = err instanceof Error ? err.message : String(err)
        setError(`上传失败: ${errMsg}`)

        ipfsUploadHistory.unshift({
            cid: '',
            filename: file.name,
            size: file.size,
            timestamp: new Date().toLocaleTimeString(),
            status: 'failed',
        })
        saveUploadHistory()

        throw err
    } finally {
        ipfsUploading.value = false
    }
}

/**
 * 通过本地 P2P 节点读取 CID（无网络活动）
 */
async function downloadFromLocalNode(cid: CID): Promise<Uint8Array | null> {
    if (!ipfsNodeRunning.value || !heliaInstance || !fsInstance) return null

    try {
        const chunks: Uint8Array[] = []
        for await (const chunk of fsInstance.cat(cid, {
            signal: AbortSignal.timeout(5_000),
        })) {
            chunks.push(chunk)
        }
        return concatenateUint8Arrays(chunks)
    } catch {
        return null
    }
}

/**
 * 通过 HTTP 网关下载 CID（零 P2P 连接）
 */
async function downloadFromGateway(cid: CID): Promise<Uint8Array> {
    const httpHelia = await createHeliaHTTP({
        blockBrokers: [trustlessGateway()],
        routers: [
            httpGatewayRouting({
                gateways: ['https://ipfs.io', 'https://cloudflare-ipfs.com'],
            }),
        ],
    })

    try {
        const fs = unixfs(httpHelia)
        const chunks: Uint8Array[] = []
        for await (const chunk of fs.cat(cid, {
            signal: AbortSignal.timeout(30_000),
        })) {
            chunks.push(chunk)
        }
        return concatenateUint8Arrays(chunks)
    } finally {
        await httpHelia.stop().catch(() => {})
    }
}

/**
 * 通过 CID 下载文件（本地优先，HTTP 网关回退）
 */
export async function downloadByCid(cidStr: string): Promise<IpfsDownloadResult> {
    if (ipfsDownloading.value) {
        throw new Error('正在下载中，请等待当前下载完成')
    }

    let cid: CID
    try {
        cid = CID.parse(cidStr.trim())
    } catch {
        throw new Error(`无效的 CID 格式: ${cidStr}`)
    }

    clearError()
    ipfsDownloading.value = true

    try {
        let content = await downloadFromLocalNode(cid)
        const source = content ? '本地节点' : 'HTTP 网关'

        if (!content) {
            content = await downloadFromGateway(cid)
        }

        const textContent = tryDecodeAsText(content)

        const result: IpfsDownloadResult = {
            cid: cidStr.trim(),
            content,
            size: content.length,
            textContent,
        }

        lastDownloadResult.value = result
        console.log(`[IPFS] 文件下载成功(${source}):`, cidStr, '大小:', formatFileSize(content.length))
        return result
    } catch (err) {
        const errMsg = err instanceof Error ? err.message : String(err)
        setError(`下载失败: ${errMsg}`)
        throw err
    } finally {
        ipfsDownloading.value = false
    }
}

/** 清空上传历史 */
export function clearUploadHistory(): void {
    ipfsUploadHistory.splice(0, ipfsUploadHistory.length)
    saveUploadHistory()
}

/** 保存下载内容到本地文件 */
export function saveDownloadToFile(result: IpfsDownloadResult, filename?: string): void {
    const blob = new Blob([result.content as BlobPart])
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = filename || `ipfs_${result.cid.slice(0, 12)}`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
}

/** 手动触发 Kubo 可用性检测 */
export async function refreshKuboStatus(): Promise<void> {
    await checkKuboAvailable()
}

/** 更新 Remote Pinning 配置 */
export function updateRemotePinConfig(config: RemotePinningConfig): void {
    remotePinConfig.value = { ...config }
    // 重新初始化 Remote Pinning 实例
    if (heliaInstance && config.endpointUrl && config.accessToken) {
        remotePinsInstance = createRemotePins(heliaInstance, {
            endpointUrl: config.endpointUrl,
            accessToken: config.accessToken,
        })
    } else {
        remotePinsInstance = null
    }
}

// ─── Composable Hook ──────────────────────────────────────

export function useIpfs() {
    return {
        // 节点状态
        ipfsNodeRunning,
        ipfsUploading,
        ipfsDownloading,
        ipfsConnectionCount,
        ipfsLastError,
        ipfsUploadHistory,
        lastUploadCid,
        lastDownloadResult,

        // Kubo 状态
        kuboAvailable,
        kuboApiUrl,
        kuboChecking,

        // Remote Pinning 状态
        remotePinConfig,
        remotePinEnabled,

        // WebRTC 调试
        ipfsPeerId,
        ipfsPeers,
        ipfsProtocols,
        bootstrapConnected,
        lastDhtProvideStatus,
        targetPeerAddr,
        targetPeerConnected,

        // 常量
        MAX_CONNECTIONS,
        MAX_UPLOAD_SIZE,

        // 操作
        startP2PNode,
        stopNode,
        uploadFile,
        downloadByCid,
        getGatewayUrl,
        copyToClipboard,
        clearUploadHistory,
        saveDownloadToFile,
        refreshKuboStatus,
        updateRemotePinConfig,
        formatFileSize,
    }
}
