<!--
    IPFS 分布式存储操作面板

    以抽屉方式展示 IPFS 文件上传、下载和历史记录功能。
    支持 Kubo 本机 Pin、Remote Pinning 和 WebRTC 网络调试。
-->
<script setup lang="ts">
import { onUnmounted, ref, computed } from 'vue'
import { useMessage } from 'naive-ui'
import { useIpfs, type IpfsDownloadResult } from '../composables/use-ipfs'

interface Props {
    visible?: boolean
}

withDefaults(defineProps<Props>(), {
    visible: false
})

interface Emits {
    (e: 'update:visible', value: boolean): void
}

const emit = defineEmits<Emits>()
const message = useMessage()

const {
    ipfsNodeRunning,
    ipfsUploading,
    ipfsDownloading,
    ipfsConnectionCount,
    ipfsUploadHistory,
    lastDownloadResult,
    kuboAvailable,
    kuboApiUrl,
    kuboChecking,
    remotePinConfig,
    remotePinEnabled,
    ipfsPeerId,
    ipfsPeers,
    ipfsProtocols,
    bootstrapConnected,
    lastDhtProvideStatus,
    targetPeerAddr,
    targetPeerConnected,
    MAX_CONNECTIONS,
    MAX_UPLOAD_SIZE,
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
} = useIpfs()

// ─── 上传区 ────────────────────────────────────────────

const selectedFile = ref<File | null>(null)
const lastUploadedCid = ref<string>('')
const fileInputRef = ref<HTMLInputElement>()

const handleFileSelect = (event: Event) => {
    const input = event.target as HTMLInputElement
    if (input.files && input.files.length > 0) {
        selectedFile.value = input.files[0]
    }
}

const triggerFileSelect = () => {
    fileInputRef.value?.click()
}

const handleUpload = async () => {
    if (!selectedFile.value) return
    try {
        const cid = await uploadFile(selectedFile.value)
        lastUploadedCid.value = cid
    } catch {
        // 错误已在 composable 中处理
    }
}

const handleCopyCid = async (cid: string) => {
    try {
        await copyToClipboard(cid)
        message.success('CID 已复制到剪贴板')
    } catch {
        message.error('复制失败')
    }
}

const handleOpenCid = (cid: string) => {
    window.open(getGatewayUrl(cid), '_blank')
}

// ─── 下载区 ────────────────────────────────────────────

const inputCid = ref('')
const downloadResult = ref<IpfsDownloadResult | null>(null)
const downloadError = ref<string>('')

const handleDownload = async () => {
    if (!inputCid.value.trim()) return
    downloadResult.value = null
    downloadError.value = ''
    try {
        const result = await downloadByCid(inputCid.value)
        downloadResult.value = result
    } catch (err) {
        downloadError.value = err instanceof Error ? err.message : String(err)
    }
}

const handleSaveFile = () => {
    if (!downloadResult.value) return
    saveDownloadToFile(downloadResult.value)
}

// ─── 节点控制 ──────────────────────────────────────────

const handleToggleNode = async () => {
    if (ipfsNodeRunning.value) {
        await stopNode()
    } else {
        try {
            await startP2PNode()
            message.success('P2P 节点已启动')
        } catch (err) {
            message.error(`启动失败: ${err instanceof Error ? err.message : String(err)}`)
        }
    }
}

// ─── Remote Pinning 配置 ───────────────────────────────

const showRemotePinConfig = ref(false)
const tempEndpointUrl = ref('')
const tempAccessToken = ref('')

const openRemotePinConfig = () => {
    tempEndpointUrl.value = remotePinConfig.value.endpointUrl
    tempAccessToken.value = remotePinConfig.value.accessToken
    showRemotePinConfig.value = true
}

const saveRemotePinConfig = () => {
    updateRemotePinConfig({
        endpointUrl: tempEndpointUrl.value,
        accessToken: tempAccessToken.value,
    })
    showRemotePinConfig.value = false
    message.success('Remote Pinning 配置已保存')
}

// ─── 调试面板 ──────────────────────────────────────────

const showDebugPanel = ref(false)
const showUploadHistory = ref(false)

const protocolsSummary = computed(() => {
    const entries = Object.entries(ipfsProtocols.value)
    if (entries.length === 0) return '无连接'
    return entries.map(([k, v]) => `${k}: ${v}`).join(', ')
})

onUnmounted(async () => {
    if (ipfsNodeRunning.value) {
        await stopNode()
    }
})
</script>

<template>
    <n-drawer
        :show="visible"
        @update:show="(val: boolean) => emit('update:visible', val)"
        :width="500"
        placement="right"
    >
        <n-drawer-content>
            <template #header>
                <span>IPFS 分布式存储</span>
            </template>

            <!-- 安全提示 -->
            <n-alert type="warning" style="margin-bottom: 12px;" closable>
                上传到 IPFS 的内容 CID 是公开可访问的，请勿上传敏感数据
            </n-alert>

            <!-- 节点控制区 -->
            <n-card title="节点控制" size="small" style="margin-bottom: 12px;">
                <div style="display: flex; align-items: center; gap: 10px; flex-wrap: wrap;">
                    <span>状态:</span>
                    <n-tag :type="ipfsNodeRunning ? 'success' : 'default'" size="small">
                        {{ ipfsNodeRunning ? '运行中' : '已停止' }}
                    </n-tag>
                    <span style="margin-left: 6px;">连接:</span>
                    <span style="color: #ccc;">{{ ipfsConnectionCount }} / {{ MAX_CONNECTIONS }}</span>
                    <span style="margin-left: 6px;">Kubo:</span>
                    <n-tag :type="kuboAvailable ? 'success' : 'default'" size="small">
                        {{ kuboAvailable ? '在线' : '离线' }}
                    </n-tag>
                    <n-button
                        size="small"
                        :type="ipfsNodeRunning ? 'error' : 'primary'"
                        @click="handleToggleNode"
                        style="margin-left: auto;"
                    >
                        {{ ipfsNodeRunning ? '停止节点' : '启动节点' }}
                    </n-button>
                </div>
                <div style="color: #999; font-size: 12px; margin-top: 8px;">
                    WebRTC + WebSocket P2P 模式，连接上限 {{ MAX_CONNECTIONS }}，上传限制 {{ formatFileSize(MAX_UPLOAD_SIZE) }}
                </div>
                <div style="margin-top: 6px; display: flex; align-items: center; gap: 6px;">
                    <span style="color: #999; font-size: 12px;">Kubo API:</span>
                    <n-input
                        v-model:value="kuboApiUrl"
                        size="tiny"
                        style="flex: 1; font-size: 11px; font-family: monospace;"
                        placeholder="http://127.0.0.1:5001"
                    />
                    <n-button size="tiny" quaternary @click="refreshKuboStatus" :loading="kuboChecking">
                        检测
                    </n-button>
                </div>
                <!-- 目标 Peer 连接状态 -->
                <div style="margin-top: 6px; display: flex; align-items: center; gap: 6px;">
                    <span style="color: #999; font-size: 12px;">目标节点:</span>
                    <n-tag :type="targetPeerConnected ? 'success' : 'default'" size="small">
                        {{ targetPeerConnected ? '已连接' : '未连接' }}
                    </n-tag>
                    <span style="color: #666; font-size: 10px; font-family: monospace; word-break: break-all;">
                        {{ targetPeerAddr.slice(0, 40) }}...
                    </span>
                </div>
            </n-card>

            <!-- 文件上传区 -->
            <n-card title="文件上传" size="small" style="margin-bottom: 12px;">
                <div style="display: flex; align-items: center; gap: 10px; margin-bottom: 10px;">
                    <input
                        ref="fileInputRef"
                        type="file"
                        style="display: none;"
                        @change="handleFileSelect"
                    />
                    <n-button @click="triggerFileSelect" size="small">
                        选择文件
                    </n-button>
                    <span style="color: #ccc; font-size: 13px;">
                        {{ selectedFile ? `${selectedFile.name} (${formatFileSize(selectedFile.size)})` : '未选择文件' }}
                    </span>
                </div>

                <!-- Pin 选项 -->
                <div style="display: flex; gap: 12px; margin-bottom: 10px; align-items: center;">
                    <span style="color: #999; font-size: 12px;">Pin 到:</span>
                    <n-tag :type="kuboAvailable ? 'success' : 'default'" size="small">
                        Kubo {{ kuboAvailable ? '(自动)' : '(不可用)' }}
                    </n-tag>
                    <n-checkbox v-model:checked="remotePinEnabled" size="small">
                        <span style="font-size: 12px;">Remote Pinning</span>
                    </n-checkbox>
                    <n-button size="tiny" quaternary @click="openRemotePinConfig">配置</n-button>
                </div>

                <n-button
                    type="primary"
                    size="small"
                    :loading="ipfsUploading"
                    :disabled="!selectedFile || !ipfsNodeRunning"
                    @click="handleUpload"
                >
                    上传到 IPFS
                </n-button>

                <!-- 上传结果 -->
                <div v-if="lastUploadedCid" style="margin-top: 10px;">
                    <n-divider style="margin: 8px 0;" />
                    <div style="font-size: 12px; color: #999; margin-bottom: 4px;">上传结果:</div>
                    <div style="display: flex; align-items: center; gap: 6px;">
                        <span style="color: #66ff5c; font-family: monospace; font-size: 12px; word-break: break-all;">
                            {{ lastUploadedCid }}
                        </span>
                    </div>
                    <div style="display: flex; gap: 6px; margin-top: 6px;">
                        <n-button size="small" @click="handleCopyCid(lastUploadedCid)">复制 CID</n-button>
                        <n-button size="small" @click="handleOpenCid(lastUploadedCid)">在浏览器打开</n-button>
                    </div>
                </div>
            </n-card>

            <!-- Remote Pinning 配置弹窗 -->
            <n-modal v-model:show="showRemotePinConfig" preset="dialog" title="Remote Pinning 配置" positive-text="保存" negative-text="取消" @positive-click="saveRemotePinConfig">
                <div style="display: flex; flex-direction: column; gap: 10px;">
                    <n-input v-model:value="tempEndpointUrl" placeholder="Endpoint URL (如 https://api.pinata.cloud/psa)" size="small" />
                    <n-input v-model:value="tempAccessToken" type="password" show-password-on="click" placeholder="Access Token / JWT" size="small" />
                    <div style="color: #999; font-size: 12px;">
                        兼容 IPFS Pinning Service API 规范的服务（Pinata、Web3.storage、自建等）
                    </div>
                </div>
            </n-modal>

            <!-- CID 下载区 -->
            <n-card title="CID 下载" size="small" style="margin-bottom: 12px;">
                <div style="display: flex; gap: 8px; margin-bottom: 10px;">
                    <n-input
                        v-model:value="inputCid"
                        placeholder="输入 CID (如 QmXyz... 或 bafy...)"
                        size="small"
                        style="flex: 1;"
                    />
                    <n-button
                        type="primary"
                        size="small"
                        :loading="ipfsDownloading"
                        :disabled="!inputCid.trim()"
                        @click="handleDownload"
                    >
                        下载
                    </n-button>
                </div>
                <div style="color: #999; font-size: 12px;">
                    本地节点优先读取，未命中则回退 HTTP 网关
                </div>

                <n-alert v-if="downloadError" type="error" style="margin-top: 8px;" closable @close="downloadError = ''">
                    {{ downloadError }}
                </n-alert>

                <div v-if="downloadResult" style="margin-top: 10px;">
                    <n-divider style="margin: 8px 0;" />
                    <div style="font-size: 12px; color: #999;">
                        大小: {{ formatFileSize(downloadResult.size) }}
                    </div>
                    <div v-if="downloadResult.textContent" style="margin-top: 8px;">
                        <div style="font-size: 12px; color: #999; margin-bottom: 4px;">内容预览:</div>
                        <n-scrollbar style="max-height: 200px;">
                            <pre style="background: #1a1a1a; padding: 8px; border-radius: 4px; font-size: 12px; color: #ccc; margin: 0; white-space: pre-wrap; word-break: break-all;">{{ downloadResult.textContent.slice(0, 5000) }}</pre>
                        </n-scrollbar>
                    </div>
                    <div v-else style="margin-top: 8px; color: #999; font-size: 12px;">
                        二进制文件，无法预览
                    </div>
                    <n-button size="small" style="margin-top: 8px;" @click="handleSaveFile">
                        保存到本地
                    </n-button>
                </div>
            </n-card>

            <!-- 上传历史 -->
            <n-card size="small" style="margin-bottom: 12px;">
                <template #header>
                    <div style="display: flex; align-items: center; gap: 8px;">
                        <span>上传历史</span>
                        <span style="color: #999; font-size: 12px;">({{ ipfsUploadHistory.length }})</span>
                    </div>
                </template>
                <template #header-extra>
                    <div style="display: flex; gap: 6px; align-items: center;">
                        <n-button v-if="ipfsUploadHistory.length > 0" size="tiny" type="error" quaternary @click="clearUploadHistory">
                            清空
                        </n-button>
                        <n-button v-if="ipfsUploadHistory.length > 1" size="tiny" quaternary @click="showUploadHistory = !showUploadHistory">
                            {{ showUploadHistory ? '收起' : '全部' }}
                        </n-button>
                    </div>
                </template>

                <div v-if="ipfsUploadHistory.length === 0" style="color: #999; font-size: 12px; text-align: center; padding: 10px;">
                    暂无上传记录
                </div>
                <div v-else>
                    <!-- 最新一条始终显示 -->
                    <div
                        style="display: flex; align-items: center; gap: 6px; padding: 6px 0; font-size: 12px;"
                    >
                        <n-tag :type="ipfsUploadHistory[0].status === 'success' ? 'success' : 'error'" size="small">
                            {{ ipfsUploadHistory[0].status === 'success' ? '成功' : '失败' }}
                        </n-tag>
                        <n-tag v-if="ipfsUploadHistory[0].kuboPinned" type="info" size="small">Kubo</n-tag>
                        <n-tag v-if="ipfsUploadHistory[0].remotePinned" type="warning" size="small">Remote</n-tag>
                        <span
                            v-if="ipfsUploadHistory[0].cid"
                            style="color: #66ff5c; font-family: monospace; cursor: pointer;"
                            @click="handleCopyCid(ipfsUploadHistory[0].cid)"
                            :title="ipfsUploadHistory[0].cid"
                        >
                            {{ ipfsUploadHistory[0].cid.slice(0, 12) }}...
                        </span>
                        <span style="color: #ccc;">{{ ipfsUploadHistory[0].filename }}</span>
                        <span style="color: #999;">{{ formatFileSize(ipfsUploadHistory[0].size) }}</span>
                        <span style="color: #666; margin-left: auto;">{{ ipfsUploadHistory[0].timestamp }}</span>
                    </div>

                    <!-- 折叠区域：其余历史记录 -->
                    <div v-if="showUploadHistory && ipfsUploadHistory.length > 1">
                        <n-divider style="margin: 4px 0;" />
                        <div
                            v-for="(record, index) in ipfsUploadHistory.slice(1)"
                            :key="index"
                            style="display: flex; align-items: center; gap: 6px; padding: 6px 0; border-bottom: 1px solid #333; font-size: 12px;"
                        >
                            <n-tag :type="record.status === 'success' ? 'success' : 'error'" size="small">
                                {{ record.status === 'success' ? '成功' : '失败' }}
                            </n-tag>
                            <n-tag v-if="record.kuboPinned" type="info" size="small">Kubo</n-tag>
                            <n-tag v-if="record.remotePinned" type="warning" size="small">Remote</n-tag>
                            <span
                                v-if="record.cid"
                                style="color: #66ff5c; font-family: monospace; cursor: pointer;"
                                @click="handleCopyCid(record.cid)"
                                :title="record.cid"
                            >
                                {{ record.cid.slice(0, 12) }}...
                            </span>
                            <span style="color: #ccc;">{{ record.filename }}</span>
                            <span style="color: #999;">{{ formatFileSize(record.size) }}</span>
                            <span style="color: #666; margin-left: auto;">{{ record.timestamp }}</span>
                        </div>
                    </div>
                </div>
            </n-card>

            <!-- 网络调试面板（实验性） -->
            <n-card size="small">
                <template #header>
                    <div style="display: flex; align-items: center; gap: 8px;">
                        <span>网络调试</span>
                        <n-tag type="warning" size="small">实验性</n-tag>
                    </div>
                </template>
                <template #header-extra>
                    <n-button size="tiny" quaternary @click="showDebugPanel = !showDebugPanel">
                        {{ showDebugPanel ? '收起' : '展开' }}
                    </n-button>
                </template>

                <div v-if="showDebugPanel">
                    <div v-if="!ipfsNodeRunning" style="color: #999; font-size: 12px; text-align: center; padding: 10px;">
                        请先启动节点
                    </div>
                    <div v-else>
                        <!-- PeerID -->
                        <div style="margin-bottom: 8px;">
                            <span style="color: #999; font-size: 12px;">PeerID:</span>
                            <span style="color: #ccc; font-family: monospace; font-size: 11px; word-break: break-all; margin-left: 4px;">
                                {{ ipfsPeerId || '获取中...' }}
                            </span>
                        </div>

                        <!-- Bootstrap 状态 -->
                        <div style="margin-bottom: 8px;">
                            <span style="color: #999; font-size: 12px;">Bootstrap:</span>
                            <n-tag :type="bootstrapConnected ? 'success' : 'default'" size="small" style="margin-left: 4px;">
                                {{ bootstrapConnected ? '已连接' : '未连接' }}
                            </n-tag>
                        </div>

                        <!-- 协议统计 -->
                        <div style="margin-bottom: 8px;">
                            <span style="color: #999; font-size: 12px;">协议:</span>
                            <span style="color: #ccc; font-size: 12px; margin-left: 4px;">{{ protocolsSummary }}</span>
                        </div>

                        <!-- DHT Provide 状态 -->
                        <div style="margin-bottom: 8px;">
                            <span style="color: #999; font-size: 12px;">DHT Provide:</span>
                            <span style="color: #ccc; font-size: 12px; margin-left: 4px;">
                                {{ lastDhtProvideStatus || '未触发' }}
                            </span>
                        </div>

                        <!-- Peer 列表 -->
                        <div>
                            <div style="color: #999; font-size: 12px; margin-bottom: 4px;">
                                已连接 Peers ({{ ipfsPeers.length }}):
                            </div>
                            <n-scrollbar style="max-height: 150px;">
                                <div v-if="ipfsPeers.length === 0" style="color: #666; font-size: 11px;">
                                    暂无连接
                                </div>
                                <div
                                    v-for="(peer, i) in ipfsPeers"
                                    :key="i"
                                    style="padding: 3px 0; border-bottom: 1px solid #222; font-size: 11px;"
                                >
                                    <span style="color: #999;">{{ peer.peerId.slice(0, 8) }}...</span>
                                    <span style="color: #666; margin-left: 6px;">{{ peer.addr.slice(0, 50) }}...</span>
                                    <n-tag v-for="p in peer.protocols" :key="p" size="small" style="margin-left: 4px;">{{ p }}</n-tag>
                                </div>
                            </n-scrollbar>
                        </div>
                    </div>
                </div>
            </n-card>

            <!-- 底部说明 -->
            <template #footer>
                <div style="font-size: 11px; color: #666;">
                    Kubo: {{ kuboAvailable ? '在线' : '离线' }} |
                    Remote Pin: {{ remotePinEnabled ? '启用' : '关闭' }} |
                    上传限制: {{ formatFileSize(MAX_UPLOAD_SIZE) }}
                </div>
            </template>
        </n-drawer-content>
    </n-drawer>
</template>
