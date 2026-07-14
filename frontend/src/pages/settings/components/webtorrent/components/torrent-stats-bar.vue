<!--
    torrent-stats-bar.vue - 底部全局统计栏

    显示所有 torrent 的汇总速度和分享率
-->
<script setup lang="ts">
import { ArrowDownOutline, ArrowUpOutline } from '@vicons/ionicons5'
import { NIcon } from 'naive-ui'
import { formatSpeed } from '@shared/utils/format'

defineProps<{
    downloadSpeed: number
    uploadSpeed: number
    ratio: number
}>()
</script>

<template>
    <div class="stats-bar">
        <div class="stat-item">
            <n-icon size="14" :component="ArrowDownOutline" color="#63e2b7" />
            <span class="stat-label">下载:</span>
            <span class="stat-value">{{ formatSpeed(downloadSpeed) }}</span>
        </div>
        <div class="stat-item">
            <n-icon size="14" :component="ArrowUpOutline" color="#f2c97d" />
            <span class="stat-label">上传:</span>
            <span class="stat-value">{{ formatSpeed(uploadSpeed) }}</span>
        </div>
        <div class="stat-item">
            <span class="stat-label">分享率:</span>
            <span class="stat-value">{{ ratio.toFixed(2) }}</span>
        </div>
    </div>
</template>

<style scoped>
.stats-bar {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 4px 12px;
    background: #1e1e1e;
    border-top: 1px solid #333;
    font-size: 12px;
}

.stat-item {
    display: flex;
    align-items: center;
    gap: 4px;
}

.stat-label {
    color: #888;
}

.stat-value {
    color: #ccc;
    font-family: monospace;
}
</style>
