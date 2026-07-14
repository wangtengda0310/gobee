/**
 * Playwright 隔离测试模板 — 在浏览器环境中隔离测试第三方库
 *
 * 使用方法：
 * 1. 复制此文件到目标项目
 * 2. 修改 TARGET_LIB_IMPORT 和测试逻辑
 * 3. 确保 dev server 正在运行
 * 4. node playwright-isolate.mjs
 */
import { chromium } from 'playwright'

// ========== 配置区 ==========
const DEV_SERVER_URL = 'http://localhost:9246'
// Vite 预构建路径，或改为库的 CDN URL
const TARGET_LIB_IMPORT = '/node_modules/.vite/deps/target-lib.js'
// 测试超时（毫秒）
const TEST_TIMEOUT = 15000
// ========== 配置区结束 ==========

async function sleep(ms) {
    return new Promise(r => setTimeout(r, ms))
}

async function main() {
    console.log('[调试] 启动 Chromium...')
    const browser = await chromium.launch({ headless: true })
    const page = await browser.newPage()

    const consoleLogs = []
    page.on('console', msg => {
        consoleLogs.push(`[${msg.type()}] ${msg.text()}`)
        console.log(`[页面${msg.type()}] ${msg.text()}`)
    })
    page.on('pageerror', err => {
        consoleLogs.push(`[pageerror] ${err.message}`)
        console.log(`[页面错误] ${err.message}`)
    })

    console.log('[调试] 打开页面:', DEV_SERVER_URL)
    await page.goto(DEV_SERVER_URL, { waitUntil: 'domcontentloaded' })
    // 等待页面稳定
    await sleep(3000)

    console.log('[调试] 执行隔离测试...')
    const testResults = await page.evaluate(async (libImport) => {
        const results = []
        function log(level, ...args) {
            const text = args.map(a => typeof a === 'object' ? JSON.stringify(a) : String(a)).join(' ')
            results.push({ level, text })
        }

        try {
            // 导入库
            const mod = await import(libImport)
            const Lib = mod.default || mod
            log('info', '库导入成功')

            // ========== 测试逻辑（根据实际情况修改） ==========
            const instance = new Lib()
            log('info', '实例创建成功, constructor:', instance.constructor.name)

            // 检查关键属性是否被 Proxy 包装
            const arrayProps = Object.keys(instance).filter(k => Array.isArray(instance[k]))
            arrayProps.forEach(prop => {
                const constructorName = instance[prop].constructor.name
                log('info', `属性 ${prop} 类型: ${constructorName}`, constructorName === 'Proxy' ? '⚠️ 已被 Proxy 包装!' : '✅ 正常')
            })

            // TODO: 在此添加核心 API 测试
            // instance.someMethod(...)
            // log('success', '核心方法调用成功')

            log('success', '所有测试通过')
            // ========== 测试逻辑结束 ==========
        } catch (err) {
            log('error', '测试失败:', err.message)
            log('error', '堆栈:', err.stack)
        }

        return results
    }, TARGET_LIB_IMPORT)

    console.log('\n=== 测试结果 ===')
    testResults.forEach(r => console.log(`[${r.level}] ${r.text}`))

    const errors = testResults.filter(r => r.level === 'error')
    console.log('\n=== 汇总 ===')
    console.log('总日志数:', testResults.length)
    console.log('错误数:', errors.length)

    await browser.close()
}

main().catch(err => {
    console.error('[调试] 脚本执行失败:', err)
    process.exit(1)
})
