<!--
    download-panel.vue - Magnet 下载面板组件

    功能：输入 magnet URI 下载、实时进度显示、保存文件到本地、取消下载
-->
<script setup lang="ts">
import { ref, computed } from 'vue'
import { NCard, NButton, NInput, NTag, NProgress, NScrollbar, NIcon } from 'naive-ui'
import { DownloadOutline, TrashOutline, SaveOutline } from '@vicons/ionicons5'
import type { DownloadTorrentInfo } from '../../../composables/use-download'
import { formatFileSize, formatSpeed } from '@shared/utils/format'

const props = defineProps<{
    downloadingTorrents: readonly DownloadTorrentInfo[]
    isAdding: boolean
    addError: string
}>()

const emit = defineEmits<{
    download: [magnetURI: string]
    cancel: [infoHash: string]
    save: [infoHash: string]
}>()

const magnetInput = ref('')

/** 是否可以开始下载 */
const canDownload = computed(() => {
    const trimmed = magnetInput.value.trim()
    return trimmed.length > 0 && trimmed.includes('btih:')
})

/** 开始下载 */
function startDownload() {
    let uri = magnetInput.value.trim()
    if (!uri) return
    // 如果用户粘贴的只是 infoHash，自动补全为 magnet URI
    if (/^[a-fA-F0-9]{40}$/.test(uri)) {
        uri = `magnet:?xt=urn:btih:${uri}`
    }
    // 如果不是以 magnet: 开头但包含 btih，尝试补全
    if (!uri.startsWith('magnet:') && uri.includes('btih:')) {
        uri = `magnet:?xt=urn:${uri}`
    }
    emit('download', uri)
    magnetInput.value = ''
}

/** 格式化剩余时间 */
function formatTime(ms: number): string {
    if (!ms || ms === Infinity) return '--'
    const seconds = Math.floor(ms / 1000)
    if (seconds < 60) return seconds + '秒'
    if (seconds < 3600) return Math.floor(seconds / 60) + '分' + (seconds % 60) + '秒'
    return Math.floor(seconds / 3600) + '时' + Math.floor((seconds % 3600) / 60) + '分'
}
</script>

<template>
    <n-card title="Magnet 下载" size="small" style="height: 100%">
        <div class="download-panel">
            <!-- Magnet URI 输入区域 -->
            <div class="magnet-input-area">
                <n-input
                    v-model:value="magnetInput"
                    type="textarea"
                    placeholder="粘贴 Magnet URI (magnet:?xt=...)"
                    :rows="2"
                    size="small"
                />
                <n-button
                    type="primary"
                    size="small"
                    :disabled="!canDownload || isAdding"
                    :loading="isAdding"
                    @click="startDownload"
                    style="margin-top: 6px"
                >
                    <template #icon><n-icon :component="DownloadOutline" /></template>
                    开始下载
                </n-button>
                <div v-if="addError" class="error-text">{{ addError }}</div>
            </div>

            <!-- 下载列表 -->
            <div class="download-list" v-if="downloadingTorrents.length > 0">
                <n-scrollbar style="max-height: 350px">
                    <div
                        v-for="torrent in downloadingTorrents"
                        :key="torrent.infoHash"
                        class="download-item"
                    >
                        <div class="download-item-header">
                            <span class="download-item-name">{{ torrent.name }}</span>
                            <n-tag :type="torrent.done ? 'success' : 'info'" size="small">
                                {{ torrent.done ? '完成' : '下载中' }}
                            </n-tag>
                        </div>

                        <!-- 进度条 -->
                        <n-progress
                            type="line"
                            :percentage="Number((torrent.progress * 100).toFixed(1))"
                            :indicator-placement="'inside'"
                            :height="16"
                            style="margin: 6px 0"
                        />

                        <!-- 文件列表 -->
                        <div class="download-item-files">
                            <span v-for="f in torrent.files" :key="f.name" class="download-file-tag">
                                {{ f.name }} ({{ formatFileSize(f.size) }})
                            </span>
                        </div>

                        <!-- 统计信息 -->
                        <div class="download-item-stats">
                            <span>速度: {{ formatSpeed(torrent.downloadSpeed) }}</span>
                            <span>已下: {{ formatFileSize(torrent.downloaded) }}</span>
                            <span>Peers: {{ torrent.numPeers }}</span>
                            <span>剩余: {{ formatTime(torrent.timeRemaining) }}</span>
                        </div>

                        <!-- 操作按钮 -->
                        <div class="download-item-actions">
                            <n-button
                                v-if="torrent.done"
                                size="tiny"
                                type="primary"
                                @click="emit('save', torrent.infoHash)"
                            >
                                <template #icon><n-icon :component="SaveOutline" /></template>
                                保存文件
                            </n-button>
                            <n-button size="tiny" quaternary @click="emit('cancel', torrent.infoHash)">
                                <template #icon><n-icon :component="TrashOutline" /></template>
                                取消
                            </n-button>
                        </div>
                    </div>
                </n-scrollbar>
            </div>

            <!-- 空状态 -->
            <div v-else class="empty-hint">
                输入 Magnet URI 开始下载文件
            </div>
        </div>
    </n-card>
</template>

<style scoped>
.download-panel {
    display: flex;
    flex-direction: column;
    height: 100%;
}

.magnet-input-area {
    margin-bottom: 10px;
}

.error-text {
    color: #e88080;
    font-size: 12px;
    margin-top: 4px;
}

.download-list {
    flex: 1;
}

.download-item {
    padding: 8px;
    margin-bottom: 8px;
    background: #2a2a2a;
    border-radius: 6px;
    border: 1px solid #3a3a3a;
}

.download-item-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 4px;
}

.download-item-name {
    font-weight: bold;
    font-size: 13px;
    color: #ddd;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: 250px;
}

.download-item-files {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
    margin-bottom: 4px;
}

.download-file-tag {
    font-size: 11px;
    color: #999;
    background: #333;
    padding: 1px 6px;
    border-radius: 3px;
}

.download-item-stats {
    display: flex;
    gap: 12px;
    font-size: 11px;
    color: #888;
    margin-bottom: 4px;
}

.download-item-actions {
    display: flex;
    gap: 6px;
}

.empty-hint {
    color: #666;
    font-size: 12px;
    text-align: center;
    margin-top: 40px;
}
</style>
