<!--
    WebTorrent P2P 文件传输操作面板

    以抽屉方式展示 WebTorrent 做种（上传）和 Magnet 下载功能。
    基于 WebRTC 实现浏览器端点对点文件传输。
-->
<script setup lang="ts">
import { watch, ref } from 'vue'
import { NTag } from 'naive-ui'
import SeedPanel from './webtorrent/components/seed-panel.vue'
import DownloadPanel from './webtorrent/components/download-panel.vue'
import TorrentStatsBar from './webtorrent/components/torrent-stats-bar.vue'
import { useWebTorrent } from '../composables/use-webtorrent'
import { useSeed } from '../composables/use-seed'
import { useDownload } from '../composables/use-download'

// Props
interface Props {
    visible?: boolean
}

const props = withDefaults(defineProps<Props>(), {
    visible: false
})

// Emits
interface Emits {
    (e: 'update:visible', value: boolean): void
}

const emit = defineEmits<Emits>()

// 初始化 WebTorrent 客户端
const {
    client,
    webrtcSupported,
    globalDownloadSpeed,
    globalUploadSpeed,
    globalRatio,
    initClient,
} = useWebTorrent()

// 做种逻辑
const { seedingTorrents, isSeeding, seedFiles, removeSeed } = useSeed(() => client.value)

// 下载逻辑
const { downloadingTorrents, isAdding, addError, addDownload, cancelDownload, saveFile } =
    useDownload(() => client.value)

// 懒初始化：只在抽屉首次打开时初始化 WebTorrent 客户端
const initialized = ref(false)
watch(() => props.visible, (show) => {
    if (show && !initialized.value) {
        initialized.value = true
        try {
            initClient()
        } catch (err) {
            console.error('[WebTorrent] 初始化失败:', err)
        }
    }
})
</script>

<template>
    <n-drawer
        :show="visible"
        @update:show="(val: boolean) => emit('update:visible', val)"
        :width="700"
        placement="right"
    >
        <n-drawer-content>
            <template #header>
                <div style="display: flex; align-items: center; gap: 10px;">
                    <span>WebTorrent P2P 文件传输</span>
                    <n-tag :type="webrtcSupported ? 'success' : 'error'" size="small">
                        {{ webrtcSupported ? 'WebRTC 支持' : 'WebRTC 不支持' }}
                    </n-tag>
                </div>
            </template>

            <!-- 做种面板 -->
            <div style="margin-bottom: 12px;">
                <SeedPanel
                    :seeding-torrents="seedingTorrents"
                    :is-seeding="isSeeding"
                    @seed="seedFiles"
                    @remove="removeSeed"
                />
            </div>

            <!-- 下载面板 -->
            <div style="margin-bottom: 12px;">
                <DownloadPanel
                    :downloading-torrents="downloadingTorrents"
                    :is-adding="isAdding"
                    :add-error="addError"
                    @download="addDownload"
                    @cancel="cancelDownload"
                    @save="saveFile"
                />
            </div>

            <!-- 底部统计栏 -->
            <template #footer>
                <TorrentStatsBar
                    :download-speed="globalDownloadSpeed"
                    :upload-speed="globalUploadSpeed"
                    :ratio="globalRatio"
                />
            </template>
        </n-drawer-content>
    </n-drawer>
</template>
