<!--
    WebTorrent P2P 文件传输页面

    演示浏览器端基于 WebRTC 的点对点文件传输功能。
    左侧：做种（上传）面板
    右侧：Magnet 下载面板
    底部：全局传输统计
-->
<script setup lang="ts">
import { onMounted } from 'vue'
import { NTag, NScrollbar } from 'naive-ui'
import SeedPanel from './components/seed-panel.vue'
import DownloadPanel from './components/download-panel.vue'
import TorrentStatsBar from './components/torrent-stats-bar.vue'
import { useWebTorrent } from '../../composables/use-webtorrent'
import { useSeed } from '../../composables/use-seed'
import { useDownload } from '../../composables/use-download'

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

onMounted(() => {
    initClient()
})
</script>

<template>
    <div class="webtorrent-page">
        <!-- 标题栏 -->
        <div class="page-header">
            <span class="page-title">WebTorrent P2P 文件传输</span>
            <n-tag :type="webrtcSupported ? 'success' : 'error'" size="small">
                {{ webrtcSupported ? 'WebRTC 支持' : 'WebRTC 不支持' }}
            </n-tag>
        </div>

        <!-- 主内容区：左右分栏 -->
        <n-scrollbar style="flex: 1">
            <div class="panels-container">
                <div class="panel-left">
                    <SeedPanel
                        :seeding-torrents="seedingTorrents"
                        :is-seeding="isSeeding"
                        @seed="seedFiles"
                        @remove="removeSeed"
                    />
                </div>
                <div class="panel-right">
                    <DownloadPanel
                        :downloading-torrents="downloadingTorrents"
                        :is-adding="isAdding"
                        :add-error="addError"
                        @download="addDownload"
                        @cancel="cancelDownload"
                        @save="saveFile"
                    />
                </div>
            </div>
        </n-scrollbar>

        <!-- 底部统计栏 -->
        <TorrentStatsBar
            :download-speed="globalDownloadSpeed"
            :upload-speed="globalUploadSpeed"
            :ratio="globalRatio"
        />
    </div>
</template>

<style scoped>
.webtorrent-page {
    display: flex;
    flex-direction: column;
    height: 100%;
    width: 100%;
}

.page-header {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 8px 12px;
    background: #252525;
    border-bottom: 1px solid #333;
}

.page-title {
    font-size: 14px;
    font-weight: bold;
    color: #ddd;
}

.panels-container {
    display: flex;
    gap: 10px;
    padding: 10px;
    min-height: 400px;
}

.panel-left {
    flex: 1;
    min-width: 0;
}

.panel-right {
    flex: 1;
    min-width: 0;
}
</style>
