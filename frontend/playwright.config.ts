import { defineConfig, devices } from '@playwright/test';

/**
 * Playwright 配置文件
 * 用于 rain-qa-func 前端 E2E 测试
 *
 * 重要：本项目的 E2E 测试必须通过 CDP 连接运行中的 WebView2 实例，
 * 不能通过 HTTP 访问 Vite dev server。因为 Wails 桌面应用使用内存路由
 * (createMemoryHistory) 和 Wails Bridge，浏览器直接访问 Vite 端口时
 * 这些功能不可用。
 *
 * 运行方式：
 * 1. 先启动应用: wails3 dev (或 CDP_PORT=9224 wails3 dev)
 * 2. 再运行测试: npx playwright test
 *
 * CDP 端口通过环境变量 CDP_PORT 指定，默认 9223。
 */

export default defineConfig({
  // 测试目录
  testDir: './e2e',

  // 完全并行运行测试
  fullyParallel: true,

  // CI 上失败时禁止 test.only
  forbidOnly: !!process.env.CI,

  // CI 上重试失败用例
  retries: process.env.CI ? 2 : 0,

  // CI 上限制并行 workers
  workers: process.env.CI ? 1 : undefined,

  // Reporter 配置
  reporter: [
    ['html', { outputFolder: 'playwright-report' }],
    ['json', { outputFile: 'test-results/results.json' }],
    ['list']
  ],

  // 全局测试配置
  use: {
    // 不设置 baseURL —— E2E 测试通过 CDP 连接 WebView2，不是 HTTP

    // 收集失败用例的 trace
    trace: 'on-first-retry',

    // 截图配置
    screenshot: 'only-on-failure',

    // 视频录制
    video: 'retain-on-failure',

    // 超时配置
    actionTimeout: 10000,
    navigationTimeout: 30000,
  },

  // 配置项目（不同浏览器）
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
});
