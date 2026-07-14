// frontend/src/pages/webtorrent/composables/use-download.ts
import { ref, readonly, onUnmounted } from 'vue'
import type WebTorrent from 'webtorrent'

/** 下载中的 torrent 信息 */
export interface DownloadTorrentInfo {
    infoHash: string
    name: string
    progress: number
    downloadSpeed: number
    uploaded: number
    downloaded: number
    numPeers: number
    timeRemaining: number
    done: boolean
    files: readonly { name: string; size: number }[]
}

/**
 * 下载逻辑 composable
 *
 * 职责：
 * - 通过 magnet URI 添加下载任务
 * - 实时追踪下载进度和速度
 * - 下载完成后生成保存链接
 * - 取消下载任务
 */
export function useDownload(client: () => WebTorrent | null) {
    const downloadingTorrents = ref<DownloadTorrentInfo[]>([])
    const isAdding = ref(false)
    const addError = ref('')
    let refreshTimer: ReturnType<typeof setInterval> | null = null

    function startRefreshTimer() {
        if (refreshTimer) return
        refreshTimer = setInterval(refreshDownloadStats, 500)
    }

    function stopRefreshTimer() {
        if (refreshTimer) {
            clearInterval(refreshTimer)
            refreshTimer = null
        }
    }

    /** 刷新下载列表的进度统计 */
    function refreshDownloadStats() {
        const c = client()
        if (!c) return
        for (const info of downloadingTorrents.value) {
            const torrent = c.torrents.find(t => t.infoHash === info.infoHash)
            if (torrent) {
                info.progress = torrent.progress
                info.downloadSpeed = torrent.downloadSpeed
                info.uploaded = torrent.uploaded
                info.downloaded = torrent.downloaded
                info.numPeers = torrent.numPeers
                info.timeRemaining = torrent.timeRemaining
                info.done = torrent.done
            }
        }
    }

    /**
     * 添加下载任务
     * @param magnetURI magnet 链接
     */
    function addDownload(magnetURI: string) {
        const c = client()
        if (!c) return
        addError.value = ''
        isAdding.value = true

        try {
            // 从 magnet URI 中提取 infoHash 进行去重检查（支持 HEX 40位 和 Base32 32位）
            const infoHashMatch = magnetURI.match(/btih:([a-fA-F0-9]{40}|[A-Z2-7]{32})/i)
            if (infoHashMatch) {
                // Base32 转 hex（简化处理：WebTorrent 内部会做转换，这里只做字符串匹配）
                const infoHash = infoHashMatch[1].length === 32
                    ? infoHashMatch[1] // Base32 保留原始值做比较
                    : infoHashMatch[1].toLowerCase()
                // 检查是否已在下载列表中
                if (downloadingTorrents.value.some(t => t.infoHash === infoHash || t.infoHash === infoHashMatch[1])) {
                    addError.value = '该文件已在下载列表中'
                    isAdding.value = false
                    return
                }
                // 检查是否正在做种（同一个 client 内不能 add 已存在的 torrent）
                const existing = c.torrents.find(t => t.infoHash === infoHash || t.infoHash === infoHashMatch[1].toLowerCase())
                if (existing) {
                    // 文件已在做种中，直接加入下载列表并标记为完成（可直接保存）
                    const info: DownloadTorrentInfo = {
                        infoHash: existing.infoHash,
                        name: existing.name,
                        progress: 1,
                        downloadSpeed: 0,
                        uploaded: existing.uploaded,
                        downloaded: existing.downloaded,
                        numPeers: existing.numPeers,
                        timeRemaining: 0,
                        done: true,
                        files: existing.files.map(f => ({
                            name: f.name,
                            size: f.length,
                        })),
                    }
                    downloadingTorrents.value.push(info)
                    isAdding.value = false
                    startRefreshTimer()
                    return
                }
            }

            const torrent = c.add(magnetURI, (torrent) => {
                const info: DownloadTorrentInfo = {
                    infoHash: torrent.infoHash,
                    name: torrent.name,
                    progress: 0,
                    downloadSpeed: 0,
                    uploaded: 0,
                    downloaded: 0,
                    numPeers: 0,
                    timeRemaining: 0,
                    done: false,
                    files: torrent.files.map(f => ({
                        name: f.name,
                        size: f.length,
                    })),
                }
                downloadingTorrents.value.push(info)
                isAdding.value = false
                startRefreshTimer()

                torrent.on('done', () => {
                    const item = downloadingTorrents.value.find(t => t.infoHash === info.infoHash)
                    if (item) item.done = true
                })
            })

            // c.add() 返回 null 时（客户端已销毁等），重置状态
            if (!torrent) {
                addError.value = '添加下载失败：客户端不可用'
                isAdding.value = false
                return
            }

            // 监听单个 torrent 的错误（异步错误不会通过 try/catch 捕获）
            torrent.on('error', (err: Error) => {
                // 仅在 isAdding 为 true 时设置错误（避免覆盖后续状态）
                if (isAdding.value) {
                    addError.value = err.message || '添加下载失败'
                    isAdding.value = false
                }
            })
        } catch (err) {
            // c.add() 同步异常时重置状态
            addError.value = err instanceof Error ? err.message : '添加下载失败'
            isAdding.value = false
        }
    }

    /**
     * 取消下载任务
     * 仅从下载列表移除，不影响做种中的 torrent
     * @param infoHash 要取消的 torrent 的 infoHash
     */
    function cancelDownload(infoHash: string) {
        const c = client()
        if (!c) return
        // 从下载列表中移除
        const item = downloadingTorrents.value.find(t => t.infoHash === infoHash)
        downloadingTorrents.value = downloadingTorrents.value.filter(t => t.infoHash !== infoHash)
        // 仅移除非做种的 torrent（做种的 torrent 由 useSeed 管理）
        const torrent = c.torrents.find(t => t.infoHash === infoHash)
        if (torrent) {
            // 检查是否在做种列表中（做种列表来自父组件，这里通过判断是否为纯下载来决定是否 remove）
            // 如果 progress < 1 且不是 done，说明是真正的下载任务，可以安全移除
            if (item && !item.done) {
                c.remove(torrent.infoHash)
            }
            // 已完成的（来自做种的）不 destroy，只从下载列表移除
        }
        if (downloadingTorrents.value.length === 0) {
            stopRefreshTimer()
        }
    }

    /**
     * 保存已下载完成的文件到本地
     * 通过 Blob + URL.createObjectURL 触发浏览器下载
     * @param infoHash torrent 的 infoHash
     */
    async function saveFile(infoHash: string) {
        const c = client()
        if (!c) return
        const torrent = c.torrents.find(t => t.infoHash === infoHash)
        if (!torrent) return

        for (const file of torrent.files) {
            try {
                const blob = await file.blob()
                const url = URL.createObjectURL(blob)
                const a = document.createElement('a')
                a.href = url
                a.download = file.name
                document.body.appendChild(a)
                a.click()
                document.body.removeChild(a)
                setTimeout(() => URL.revokeObjectURL(url), 10000)
            } catch (err) {
                console.error('[下载] 保存文件失败:', file.name, err)
                // 继续处理下一个文件
            }
        }
    }

    // 页面卸载时清理定时器
    onUnmounted(() => {
        stopRefreshTimer()
    })

    return {
        downloadingTorrents: readonly(downloadingTorrents),
        isAdding: readonly(isAdding),
        addError: readonly(addError),
        addDownload,
        cancelDownload,
        saveFile,
    }
}
