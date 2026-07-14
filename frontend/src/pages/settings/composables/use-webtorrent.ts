// frontend/src/pages/webtorrent/composables/use-webtorrent.ts
import { ref, onUnmounted, readonly, markRaw } from 'vue'
import { injectP2PPolyfills } from '@/shared/polyfills'
import type WebTorrentType from 'webtorrent'

// 模块级变量存储定时器 ID（避免污染 WebTorrent 实例）
let statsTimer: ReturnType<typeof setInterval> | null = null

/**
 * WebTorrent 客户端管理 composable
 *
 * 职责：
 * - 创建和销毁 WebTorrent 客户端实例
 * - 提供全局传输统计（下载速度、上传速度、分享率）
 * - 检测 WebRTC 支持状态
 *
 * 注意：使用 markRaw 防止 Vue 递归包装 WebTorrent 实例，
 * 避免 Proxy 干扰 webtorrent 内部的数组操作（如 torrents.push/indexOf）
 * 性能优化：使用动态导入（import()）加载 webtorrent 模块，
 * 避免首屏 bundle 包含 P2P 代码，提升 wails3 dev 启动速度。
 */
export function useWebTorrent() {
    const client = ref<WebTorrentType.Instance | null>(null)
    const globalDownloadSpeed = ref(0)
    const globalUploadSpeed = ref(0)
    const globalRatio = ref(0)
    const webrtcSupported = ref(false)

    /**
     * 初始化 WebTorrent 客户端
     * 创建实例并启动全局统计定时刷新
     */
    async function initClient() {
        if (client.value) return

        // 注入 P2P 所需的 Node.js polyfill（Buffer/global/process）
        await injectP2PPolyfills()

        // 动态导入 webtorrent 模块，避免首屏加载
        const WebTorrent = (await import('webtorrent')).default

        webrtcSupported.value = WebTorrent.WEBRTC_SUPPORT
        // markRaw 防止 Vue reactive 系统递归包装 WebTorrent 实例，
        // 避免 client.torrents 变成 Proxy(Array) 导致 webtorrent 内部逻辑异常
        client.value = markRaw(new WebTorrent())
        client.value.on('error', (err: Error) => {
            console.error('[WebTorrent] 客户端错误:', err)
        })
        // 每秒刷新全局统计
        statsTimer = setInterval(() => {
            if (client.value) {
                globalDownloadSpeed.value = client.value.downloadSpeed
                globalUploadSpeed.value = client.value.uploadSpeed
                globalRatio.value = client.value.ratio
            }
        }, 1000)
    }

    /**
     * 销毁客户端，释放所有资源
     * 页面卸载时自动调用
     */
    function destroyClient() {
        if (statsTimer) {
            clearInterval(statsTimer)
            statsTimer = null
        }
        if (!client.value) return
        client.value.destroy((err: Error | undefined) => {
            if (err) console.error('[WebTorrent] 销毁错误:', err)
        })
        client.value = null
    }

    onUnmounted(() => {
        destroyClient()
    })

    return {
        client,
        webrtcSupported: readonly(webrtcSupported),
        globalDownloadSpeed: readonly(globalDownloadSpeed),
        globalUploadSpeed: readonly(globalUploadSpeed),
        globalRatio: readonly(globalRatio),
        initClient,
        destroyClient,
    }
}
