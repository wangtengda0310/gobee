/**
 * 测试 Fixtures
 * 提供可复用的测试设置和状态管理
 *
 * 重要：本项目的 E2E 测试通过 CDP 连接运行中的 WebView2 实例。
 * 不能通过 HTTP 访问 Vite dev server，因为 Wails 使用内存路由和 Bridge。
 *
 * 使用方式：
 * 1. 先启动应用: wails3 dev
 * 2. 再运行测试: npx playwright test
 */

import { test as base, Page, BrowserContext, chromium } from '@playwright/test';
import { HomePage } from '../pages/HomePage';
import { FunctionTestPage } from '../pages/FunctionTestPage';
import { ExcelTestPage } from '../pages/ExcelTestPage';
import { HeroWikiCheckPage } from '../pages/HeroWikiCheckPage';
import { SettingsPage } from '../pages/SettingsPage';
import { HeroVoiceResourceCheckPage } from '../pages/HeroVoiceResourceCheckPage';
import { RoadmapPage } from '../pages/RoadmapPage';
import { ProtoTestPage } from '../pages/ProtoTestPage';

// 获取 CDP 端口（与 main.go 的 getCDPPort() 一致）
function getCDPPort(): string {
  const port = process.env.CDP_PORT || '9223';
  if (!/^\d+$/.test(port)) {
    console.warn(`[fixture] 无效的 CDP_PORT: ${port}，使用默认 9223`);
    return '9223';
  }
  return port;
}

/**
 * 扩展的测试 Fixtures
 */
type AppFixtures = {
  // Page Objects
  homePage: HomePage;
  functionTestPage: FunctionTestPage;
  excelTestPage: ExcelTestPage;
  heroWikiCheckPage: HeroWikiCheckPage;
  settingsPage: SettingsPage;
  heroVoiceResourceCheckPage: HeroVoiceResourceCheckPage;
  roadmapPage: RoadmapPage;
  protoTestPage: ProtoTestPage;

  // CDP 连接的页面
  cdpPage: Page;
};

/**
 * 扩展 base test 添加自定义 fixtures
 */
export const test = base.extend<AppFixtures>(
  {
    // 覆盖默认的 page fixture —— 通过 CDP 连接 WebView2
    page: async ({}, use) => {
      const port = getCDPPort();
      const endpoint = `http://127.0.0.1:${port}`;

      // 检查 CDP 是否可用
      try {
        const response = await fetch(`${endpoint}/json/version`);
        if (!response.ok) {
          throw new Error(`CDP 端口 ${port} 无响应`);
        }
      } catch (e) {
        throw new Error(
          `无法连接到 WebView2 CDP 端口 ${port}。\n` +
          `请确保应用已启动: wails3 dev (或 CDP_PORT=${port} wails3 dev)\n` +
          `原始错误: ${e}`
        );
      }

      // 通过 CDP 连接已有浏览器实例
      const browser = await chromium.connectOverCDP(endpoint);
      const context = browser.contexts()[0];
      const page = context.pages()[0];

      await use(page);

      // 不断开连接，让应用继续运行
      // await browser.close();
    },

    // CDP 连接的页面（与 page 相同，保留兼容性）
    cdpPage: async ({ page }, use) => {
      await use(page);
    },

    // 首页 Page Object
    homePage: async ({ page }, use) => {
      const homePage = new HomePage(page);
      await use(homePage);
    },

    // 功能测试页 Page Object
    functionTestPage: async ({ page }, use) => {
      const functionTestPage = new FunctionTestPage(page);
      await use(functionTestPage);
    },

    // 配表测试页 Page Object
    excelTestPage: async ({ page }, use) => {
      const excelTestPage = new ExcelTestPage(page);
      await use(excelTestPage);
    },

    // Wiki 检查页 Page Object
    heroWikiCheckPage: async ({ page }, use) => {
      const heroWikiCheckPage = new HeroWikiCheckPage(page);
      await use(heroWikiCheckPage);
    },

    // 设置页 Page Object
    settingsPage: async ({ page }, use) => {
      const settingsPage = new SettingsPage(page);
      await use(settingsPage);
    },

    // 语音资源检查页 Page Object
    heroVoiceResourceCheckPage: async ({ page }, use) => {
      const heroVoiceResourceCheckPage = new HeroVoiceResourceCheckPage(page);
      await use(heroVoiceResourceCheckPage);
    },

    // 路线图页 Page Object
    roadmapPage: async ({ page }, use) => {
      const roadmapPage = new RoadmapPage(page);
      await use(roadmapPage);
    },

    // 协议重放页 Page Object
    protoTestPage: async ({ page }, use) => {
      const protoTestPage = new ProtoTestPage(page);
      await use(protoTestPage);
    },
  },
  { scope: 'test' }
);

/**
 * 每个 test 后清理残留的 modal/drawer 遮罩与 select 菜单，避免跑完后 webview 残留 overlay
 * 拦截手动操作点击（曾出现：E2E 跑完残留 mask 导致手动点 action select 无反应、点空白也不关，
 * 重启 dev 才恢复；E2E CLAUDE.md 第8/14条）
 */
test.afterEach(async ({page}) => {
    // 先按 Escape + 点 body 空白，关闭可能残留的 select 菜单（naive-ui select 点外部/Escape 关闭）
    await page.keyboard.press('Escape');
    await page.locator('body').click({position: {x: 1, y: 1}}).catch(() => {});
    await page.waitForTimeout(150);
    // 再清残留的 modal/drawer 遮罩
    for (let i = 0; i < 3; i++) {
        const masks = page.locator('.n-modal-mask:visible, .n-drawer-mask:visible');
        if (await masks.count() === 0) break;
        await page.keyboard.press('Escape');
        await page.waitForTimeout(200);
    }
    // 孤儿 select 菜单：naive-ui select unmounted 后 Follower 容器（.v-binder-follower-content）残留，
    // Escape/点 body 关不掉、display:none 会被 Follower 重新覆盖。只能连同 Follower 一起移除。
    // 仅移除"含可见菜单"的 Follower —— 正常 select 已被上面 Escape/点 body 关闭（菜单不可见），不会误伤
    await page.evaluate(() => {
        document.querySelectorAll('.v-binder-follower-content').forEach((f) => {
            const menu = f.querySelector('.n-base-select-menu');
            if (menu && (menu as HTMLElement).offsetParent !== null) {
                f.remove();
            }
        });
    }).catch(() => {});
});

/**
 * 重新导出 expect
 */
export { expect } from '@playwright/test';

/**
 * 测试描述分组
 */
export const describe = test.describe;
export const it = test;
