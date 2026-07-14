/**
 * P2P 库 Node.js Polyfill 注入工具
 *
 * Helia (IPFS) 和 WebTorrent 底层依赖 Node.js API（Buffer、global、process），
 * 在浏览器环境中需要手动提供兼容实现。
 *
 * 采用手动注入而非 vite-plugin-node-polyfills 的原因：
 * - 避免首屏加载时注入大量 polyfill，影响 wails3 dev 启动速度
 * - 只在用户实际打开 IPFS/WebTorrent 面板时才注入，实现按需加载
 *
 * 使用单例 Promise 模式防止并发调用时的 race condition。
 */

// 注入完成的 Promise 锁，确保并发调用只执行一次
let injectPromise: Promise<void> | null = null

/**
 * 注入 P2P 库所需的 Node.js polyfill。
 *
 * 注入内容：
 * - global: libp2p 和 WebTorrent 内部代码访问 global 对象
 * - Buffer: CID 解析、数据块处理、BitTorrent 协议编解码
 * - process: libp2p 调试日志级别控制、WebTorrent 环境检测
 *
 * 调用时机：在第一个 P2P 动态导入前调用，通常在 startP2PNode() 或 initClient() 开头。
 */
export async function injectP2PPolyfills(): Promise<void> {
  if (!injectPromise) {
    injectPromise = (async () => {
      try {
        // global: libp2p 和 WebTorrent 内部访问 global 对象
        if (typeof (globalThis as any).global === 'undefined') {
          ;(globalThis as any).global = globalThis
        }

        // Buffer: Helia 解析 CID 和 WebTorrent 处理数据必需
        if (typeof Buffer === 'undefined') {
          const { Buffer } = await import('buffer')
          ;(window as any).Buffer = Buffer
        }

        // process: libp2p 调试日志和 WebTorrent 环境检测
        if (typeof process === 'undefined') {
          ;(window as any).process = {
            env: {},
            version: '',
            platform: 'browser',
            nextTick: (fn: () => void) => setTimeout(fn, 0),
          }
        }
      } catch (error) {
        injectPromise = null // 重置，允许下次重试
        throw error
      }
    })()
  }

  return injectPromise
}
