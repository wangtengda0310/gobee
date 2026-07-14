<!--
    seed-panel.vue - 做种面板组件

    功能：文件拖拽上传、文件选择做种、magnet URI 显示与复制、移除做种
-->
<script setup lang="ts">
import { ref } from 'vue'
import { NCard, NButton, NInput, NTag, NScrollbar, NIcon, useMessage } from 'naive-ui'
import { CloudUploadOutline, TrashOutline, CopyOutline } from '@vicons/ionicons5'
import type { SeedTorrentInfo } from '../../../composables/use-seed'
import { formatFileSize, formatSpeed } from '@shared/utils/format'

const props = defineProps<{
    seedingTorrents: readonly SeedTorrentInfo[]
    isSeeding: boolean
}>()

const emit = defineEmits<{
    seed: [files: FileList | File[]]
    remove: [infoHash: string]
}>()

const message = useMessage()
const isDragging = ref(false)

/** 拖拽进入 */
function onDragOver(e: DragEvent) {
    e.preventDefault()
    isDragging.value = true
}

/** 拖拽离开 — 用 relatedTarget 过滤子元素冒泡 */
function onDragLeave(e: DragEvent) {
    const dropZone = (e.currentTarget as HTMLElement)
    if (!dropZone || !dropZone.contains(e.relatedTarget as Node)) {
        isDragging.value = false
    }
}

/** 拖拽放下 - 触发做种 */
function onDrop(e: DragEvent) {
    e.preventDefault()
    isDragging.value = false
    if (props.isSeeding) {
        message.warning('正在做种中，请等待完成')
        return
    }
    if (e.dataTransfer?.files.length) {
        emit('seed', e.dataTransfer.files)
    }
}

/** 通过文件选择器选择文件 */
function onFileSelect(e: Event) {
    const input = e.target as HTMLInputElement
    if (props.isSeeding) {
        message.warning('正在做种中，请等待完成')
        input.value = ''
        return
    }
    if (input.files?.length) {
        emit('seed', input.files)
    }
    input.value = ''
}

/** 复制 magnet URI 到剪贴板 */
async function copyMagnet(magnetURI: string) {
    try {
        await navigator.clipboard.writeText(magnetURI)
        message.success('已复制 Magnet URI')
    } catch {
        message.error('复制失败')
    }
}
</script>

<template>
    <n-card title="做种 (上传)" size="small" style="height: 100%">
        <div class="seed-panel">
            <!-- 拖拽上传区域 -->
            <div
                class="drop-zone"
                :class="{ 'drop-zone-active': isDragging }"
                @dragover="onDragOver"
                @dragleave="onDragLeave"
                @drop="onDrop"
            >
                <n-icon size="32" :component="CloudUploadOutline" />
                <div class="drop-zone-text">
                    {{ isDragging ? '释放文件开始做种' : '拖拽文件到这里' }}
                </div>
                <label class="file-select-label">
                    或 点击选择文件
                    <input type="file" multiple @change="onFileSelect" style="display: none" />
                </label>
            </div>

            <!-- 做种中提示 -->
            <n-tag v-if="isSeeding" type="info" size="small" style="margin-top: 8px">
                正在做种中...
            </n-tag>

            <!-- 做种列表 -->
            <div class="seed-list" v-if="seedingTorrents.length > 0">
                <n-scrollbar style="max-height: 300px">
                    <div
                        v-for="torrent in seedingTorrents"
                        :key="torrent.infoHash"
                        class="seed-item"
                    >
                        <div class="seed-item-header">
                            <span class="seed-item-name">{{ torrent.name }}</span>
                            <n-button size="tiny" quaternary @click="emit('remove', torrent.infoHash)">
                                <template #icon><n-icon :component="TrashOutline" /></template>
                            </n-button>
                        </div>
                        <div class="seed-item-files">
                            <span v-for="f in torrent.files" :key="f.name" class="seed-file-tag">
                                {{ f.name }} ({{ formatFileSize(f.size) }})
                            </span>
                        </div>
                        <div class="seed-item-magnet">
                            <n-input
                                :value="torrent.magnetURI"
                                size="tiny"
                                readonly
                                style="flex: 1"
                            />
                            <n-button size="tiny" quaternary @click="copyMagnet(torrent.magnetURI)">
                                <template #icon><n-icon :component="CopyOutline" /></template>
                            </n-button>
                        </div>
                        <div class="seed-item-stats">
                            <span>上传: {{ formatSpeed(torrent.uploadSpeed) }}</span>
                            <span>已传: {{ formatFileSize(torrent.uploaded) }}</span>
                            <span>Peers: {{ torrent.numPeers }}</span>
                        </div>
                    </div>
                </n-scrollbar>
            </div>
        </div>
    </n-card>
</template>

<style scoped>
.seed-panel {
    display: flex;
    flex-direction: column;
    height: 100%;
}

.drop-zone {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 20px;
    border: 2px dashed #555;
    border-radius: 8px;
    cursor: pointer;
    transition: all 0.2s ease;
    min-height: 100px;
}

.drop-zone:hover,
.drop-zone-active {
    border-color: #63e2b7;
    background-color: rgba(99, 226, 183, 0.05);
}

.drop-zone-text {
    color: #999;
    font-size: 13px;
}

.file-select-label {
    color: #63e2b7;
    font-size: 12px;
    cursor: pointer;
    text-decoration: underline;
}

.seed-list {
    margin-top: 10px;
    flex: 1;
}

.seed-item {
    padding: 8px;
    margin-bottom: 8px;
    background: #2a2a2a;
    border-radius: 6px;
    border: 1px solid #3a3a3a;
}

.seed-item-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 4px;
}

.seed-item-name {
    font-weight: bold;
    font-size: 13px;
    color: #ddd;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: 200px;
}

.seed-item-files {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
    margin-bottom: 6px;
}

.seed-file-tag {
    font-size: 11px;
    color: #999;
    background: #333;
    padding: 1px 6px;
    border-radius: 3px;
}

.seed-item-magnet {
    display: flex;
    gap: 4px;
    align-items: center;
    margin-bottom: 4px;
}

.seed-item-stats {
    display: flex;
    gap: 12px;
    font-size: 11px;
    color: #888;
}
</style>
