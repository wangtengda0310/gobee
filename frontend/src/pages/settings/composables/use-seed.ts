// frontend/src/pages/webtorrent/composables/use-seed.ts
import { ref, readonly, onUnmounted } from 'vue'
import type WebTorrent from 'webtorrent'

/** 做种中的 torrent 信息 */
export interface SeedTorrentInfo {
    infoHash: string
    name: string
    magnetURI: string
    files: readonly { name: string; size: number }[]
    uploadSpeed: number
    uploaded: number
    numPeers: number
}

/**
 * 做种逻辑 composable
 *
 * 职责：
 * - 接收用户选择的文件并做种
 * - 维护做种列表和实时统计
 * - 生成可复制的 magnet URI
 * - 移除做种任务
 */
export function useSeed(client: () => WebTorrent | null) {
    const seedingTorrents = ref<SeedTorrentInfo[]>([])
    const isSeeding = ref(false)
    let refreshTimer: ReturnType<typeof setInterval> | null = null

    function startRefreshTimer() {
        if (refreshTimer) return
        refreshTimer = setInterval(refreshSeedingStats, 1000)
    }

    function stopRefreshTimer() {
        if (refreshTimer) {
            clearInterval(refreshTimer)
            refreshTimer = null
        }
    }

    /** 刷新做种列表中的速度/peers 等统计 */
    function refreshSeedingStats() {
        const c = client()
        if (!c) return
        for (const info of seedingTorrents.value) {
            const torrent = c.torrents.find(t => t.infoHash === info.infoHash)
            if (torrent) {
                info.uploadSpeed = torrent.uploadSpeed
                info.uploaded = torrent.uploaded
                info.numPeers = torrent.numPeers
            }
        }
    }

    /**
     * 开始做种
     * @param files 用户选择的 FileList 或 File 数组
     */
    function seedFiles(files: FileList | File[]) {
        const c = client()
        if (!c) return
        if (isSeeding.value) return
        isSeeding.value = true

        // 计算文件的唯一标识用于去重（文件名+大小）
        const fileKey = (f: File) => `${f.name}:${f.size}`
        const inputFiles = Array.from(files)
        const inputKeys = new Set(inputFiles.map(fileKey))

        // 检查是否已有完全相同的文件在做种
        for (const t of c.torrents) {
            const existingFiles = t.files.map(f => `${f.name}:${f.length}`)
            if (existingFiles.every(k => inputKeys.has(k)) && existingFiles.length === inputFiles.length) {
                console.warn('[做种] 文件已在做种列表中，跳过:', t.name)
                isSeeding.value = false
                return
            }
        }

        // 超时保护：30 秒后自动重置状态，防止 seed 回调永远不触发
        const timeout = setTimeout(() => {
            if (isSeeding.value) {
                console.warn('[做种] 超时，重置 isSeeding 状态')
                isSeeding.value = false
            }
        }, 30000)

        try {
            const torrent = c.seed(inputFiles, (torrent) => {
                clearTimeout(timeout)
                // 如果 torrent 已被销毁（重复文件情况），忽略
                if (torrent.destroyed) {
                    isSeeding.value = false
                    return
                }
                // 如果列表中已存在相同 infoHash，不重复添加
                if (seedingTorrents.value.some(t => t.infoHash === torrent.infoHash)) {
                    isSeeding.value = false
                    return
                }

                const info: SeedTorrentInfo = {
                    infoHash: torrent.infoHash,
                    name: torrent.name,
                    magnetURI: torrent.magnetURI,
                    files: torrent.files.map(f => ({
                        name: f.name,
                        size: f.length,
                    })),
                    uploadSpeed: 0,
                    uploaded: 0,
                    numPeers: 0,
                }
                seedingTorrents.value.push(info)
                isSeeding.value = false
                startRefreshTimer()
            })

            // 监听 torrent 级别的 error 事件，重置 isSeeding
            if (torrent) {
                torrent.on('error', (err: Error) => {
                    if (isSeeding.value) {
                        clearTimeout(timeout)
                        console.error('[做种] torrent 错误:', err.message)
                        isSeeding.value = false
                    }
                })
            }
        } catch (err) {
            clearTimeout(timeout)
            console.error('[做种] seed 调用失败:', err)
            isSeeding.value = false
        }
    }

    /**
     * 移除做种任务
     * @param infoHash 要移除的 torrent 的 infoHash
     */
    function removeSeed(infoHash: string) {
        const c = client()
        if (!c) return
        const torrent = c.torrents.find(t => t.infoHash === infoHash)
        if (torrent) {
            c.remove(torrent.infoHash)
        }
        seedingTorrents.value = seedingTorrents.value.filter(t => t.infoHash !== infoHash)
        if (seedingTorrents.value.length === 0) {
            stopRefreshTimer()
        }
    }

    // 页面卸载时清理定时器
    onUnmounted(() => {
        stopRefreshTimer()
    })

    return {
        seedingTorrents: readonly(seedingTorrents),
        isSeeding: readonly(isSeeding),
        seedFiles,
        removeSeed,
    }
}
