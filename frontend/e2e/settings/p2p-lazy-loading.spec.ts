/**
 * P2P 动态加载优化验收测试
 *
 * 验收目标：确认 IPFS 和 WebTorrent 面板使用 defineAsyncComponent 异步加载后，
 * Settings 页面功能正常，没有引入 bug。
 *
 * 连接方式：通过 CDP (Chrome DevTools Protocol) 连接 WebView2
 * CDP 端口：9223（main.go 中配置 --remote-debugging-port=9223）
 */

import { test, expect, describe } from '../shared/fixtures';
import { resolveRoute } from '../shared/pages/BasePage';
import { sleep } from '../shared/utils/helpers';

describe('P2P 动态加载优化验收', () => {
  /**
   * 每个测试前：加载首页并切换到设置页面
   */
  test.beforeEach(async ({ page }) => {
    // 加载根页面
    await page.goto(resolveRoute(page, '/'));
    await page.waitForLoadState('networkidle');

    // 点击设置导航按钮
    await page.locator('#layout-header button:has-text("设置")').click();
    await sleep(500);
  });

  /**
   * 测试1：Settings 页面正常加载
   * 验证所有配置卡片都正确显示，包括新加入的 P2P 相关卡片
   */
  describe('Settings 页面加载', () => {
    test('页面加载 - 显示所有配置卡片', async ({ page }) => {
      // 验证设置页面容器存在
      await expect(page.locator('#settings-page')).toBeVisible();

      // 验证飞书通知配置卡片存在
      await expect(page.locator('.n-card:has-text("飞书通知配置")')).toBeVisible();

      // 验证 MCP 配置卡片存在
      await expect(page.locator('.n-card:has-text("MCP 服务配置")')).toBeVisible();

      // 验证服务端日志卡片存在
      await expect(page.locator('.n-card:has-text("服务端日志")')).toBeVisible();

      // 验证 IPFS 分布式存储卡片存在（异步加载的面板入口）
      await expect(page.locator('.n-card:has-text("IPFS 分布式存储")')).toBeVisible();

      // 验证 WebTorrent P2P 传输卡片存在（异步加载的面板入口）
      await expect(page.locator('.n-card:has-text("WebTorrent P2P 传输")')).toBeVisible();

      // 验证开发路线图卡片存在
      await expect(page.locator('.n-card:has-text("开发路线图")')).toBeVisible();
    });

    test('IPFS 卡片显示正确的状态信息', async ({ page }) => {
      const ipfsCard = page.locator('.n-card:has-text("IPFS 分布式存储")');
      await expect(ipfsCard).toBeVisible();

      // 验证节点状态标签存在
      await expect(ipfsCard.locator('text=/节点状态/')).toBeVisible();

      // 验证连接数显示存在
      await expect(ipfsCard.locator('text=/连接/')).toBeVisible();

      // 验证上传记录显示存在
      await expect(ipfsCard.locator('text=/上传记录/')).toBeVisible();

      // 验证打开按钮存在且可点击
      const openButton = ipfsCard.locator('button:has-text("打开 IPFS 面板")');
      await expect(openButton).toBeVisible();
      await expect(openButton).toBeEnabled();
    });

    test('WebTorrent 卡片显示正确的状态信息', async ({ page }) => {
      const torrentCard = page.locator('.n-card:has-text("WebTorrent P2P 传输")');
      await expect(torrentCard).toBeVisible();

      // 验证传输方式标签存在
      await expect(torrentCard.locator('text=/传输方式/')).toBeVisible();

      // 验证 WebRTC 标签存在
      await expect(torrentCard.locator('.n-tag:has-text("WebRTC")')).toBeVisible();

      // 验证打开按钮存在且可点击
      const openButton = torrentCard.locator('button:has-text("打开 P2P 传输面板")');
      await expect(openButton).toBeVisible();
      await expect(openButton).toBeEnabled();
    });
  });

  /**
   * 测试2：IPFS 面板异步加载正常
   * 点击打开 IPFS 面板，验证抽屉正常打开，组件异步加载无报错
   */
  describe('IPFS 面板异步加载', () => {
    test('点击打开 IPFS 面板 - 抽屉正常显示', async ({ page }) => {
      const ipfsCard = page.locator('.n-card:has-text("IPFS 分布式存储")');
      const openButton = ipfsCard.locator('button:has-text("打开 IPFS 面板")');

      // 点击打开按钮
      await openButton.click();
      await sleep(800); // 等待异步组件加载和抽屉动画

      // 验证抽屉可见
      const drawer = page.locator('.n-drawer');
      await expect(drawer).toBeVisible();

      // 验证抽屉标题正确
      await expect(page.locator('.n-drawer-header__main:has-text("IPFS 分布式存储")')).toBeVisible();
    });

    test('IPFS 面板内容 - 节点控制区显示正常', async ({ page }) => {
      const ipfsCard = page.locator('.n-card:has-text("IPFS 分布式存储")');
      await ipfsCard.locator('button:has-text("打开 IPFS 面板")').click();
      await sleep(800);

      // 验证节点控制卡片存在
      await expect(page.locator('.n-card:has-text("节点控制")')).toBeVisible();

      // 验证状态标签存在
      await expect(page.locator('.n-drawer .n-tag:has-text("已停止")').first()).toBeVisible();

      // 验证启动/停止按钮存在
      await expect(page.locator('.n-drawer button:has-text("启动节点")')).toBeVisible();
    });

    test('IPFS 面板内容 - 文件上传区显示正常', async ({ page }) => {
      const ipfsCard = page.locator('.n-card:has-text("IPFS 分布式存储")');
      await ipfsCard.locator('button:has-text("打开 IPFS 面板")').click();
      await sleep(800);

      // 验证文件上传卡片存在
      await expect(page.locator('.n-card:has-text("文件上传")')).toBeVisible();

      // 验证选择文件按钮存在
      await expect(page.locator('.n-drawer button:has-text("选择文件")')).toBeVisible();

      // 验证上传到 IPFS 按钮存在（初始状态可能禁用）
      await expect(page.locator('.n-drawer button:has-text("上传到 IPFS")')).toBeVisible();
    });

    test('IPFS 面板内容 - CID 下载区显示正常', async ({ page }) => {
      const ipfsCard = page.locator('.n-card:has-text("IPFS 分布式存储")');
      await ipfsCard.locator('button:has-text("打开 IPFS 面板")').click();
      await sleep(800);

      // 验证 CID 下载卡片存在
      await expect(page.locator('.n-card:has-text("CID 下载")')).toBeVisible();

      // 验证下载按钮存在
      await expect(page.locator('.n-drawer button:has-text("下载")').first()).toBeVisible();
    });

    test('IPFS 面板关闭后重新打开 - 正常显示', async ({ page }) => {
      const ipfsCard = page.locator('.n-card:has-text("IPFS 分布式存储")');

      // 第一次打开
      await ipfsCard.locator('button:has-text("打开 IPFS 面板")').click();
      await sleep(800);
      await expect(page.locator('.n-drawer')).toBeVisible();

      // 关闭抽屉（点击遮罩或关闭按钮）
      await page.keyboard.press('Escape');
      await sleep(300);

      // 再次打开
      await ipfsCard.locator('button:has-text("打开 IPFS 面板")').click();
      await sleep(800);

      // 验证抽屉再次正常显示
      await expect(page.locator('.n-drawer')).toBeVisible();
      await expect(page.locator('.n-drawer-header__main:has-text("IPFS 分布式存储")')).toBeVisible();
    });
  });

  /**
   * 测试3：WebTorrent 面板异步加载正常
   * 点击打开 WebTorrent 面板，验证抽屉正常打开，组件异步加载无报错
   */
  describe('WebTorrent 面板异步加载', () => {
    test('点击打开 WebTorrent 面板 - 抽屉正常显示', async ({ page }) => {
      const torrentCard = page.locator('.n-card:has-text("WebTorrent P2P 传输")');
      const openButton = torrentCard.locator('button:has-text("打开 P2P 传输面板")');

      // 点击打开按钮
      await openButton.click();
      await sleep(800); // 等待异步组件加载和抽屉动画

      // 验证抽屉可见
      const drawer = page.locator('.n-drawer');
      await expect(drawer).toBeVisible();

      // 验证抽屉标题正确
      await expect(page.locator('.n-drawer-header__main:has-text("WebTorrent P2P 文件传输")')).toBeVisible();
    });

    test('WebTorrent 面板内容 - WebRTC 状态标签显示正常', async ({ page }) => {
      const torrentCard = page.locator('.n-card:has-text("WebTorrent P2P 传输")');
      await torrentCard.locator('button:has-text("打开 P2P 传输面板")').click();
      await sleep(800);

      // 验证 WebRTC 支持标签存在
      const webrtcTag = page.locator('.n-drawer .n-tag:has-text("WebRTC")');
      await expect(webrtcTag).toBeVisible();
    });

    test('WebTorrent 面板关闭后重新打开 - 正常显示', async ({ page }) => {
      const torrentCard = page.locator('.n-card:has-text("WebTorrent P2P 传输")');

      // 第一次打开
      await torrentCard.locator('button:has-text("打开 P2P 传输面板")').click();
      await sleep(800);
      await expect(page.locator('.n-drawer')).toBeVisible();

      // 关闭抽屉
      await page.keyboard.press('Escape');
      await sleep(300);

      // 再次打开
      await torrentCard.locator('button:has-text("打开 P2P 传输面板")').click();
      await sleep(800);

      // 验证抽屉再次正常显示
      await expect(page.locator('.n-drawer')).toBeVisible();
      await expect(page.locator('.n-drawer-header__main:has-text("WebTorrent P2P 文件传输")')).toBeVisible();
    });
  });

  /**
   * 测试4：页面整体功能回归
   * 验证异步加载没有破坏页面其他功能
   */
  describe('页面整体功能回归', () => {
    test('设置页面可以正常滚动查看所有卡片', async ({ page }) => {
      const scrollbar = page.locator('.n-scrollbar');
      await expect(scrollbar).toBeVisible();

      // 滚动到底部
      await page.evaluate(() => {
        const container = document.querySelector('.n-scrollbar-container');
        if (container) {
          container.scrollTop = container.scrollHeight;
        }
      });
      await sleep(200);

      // 验证底部的开发路线图卡片仍然可见
      await expect(page.locator('.n-card:has-text("开发路线图")')).toBeVisible();
    });

    test('其他配置卡片功能不受影响', async ({ page }) => {
      // 验证飞书开关可以交互
      const feishuCard = page.locator('.n-card:has-text("飞书通知配置")');
      const feishuSwitch = feishuCard.locator('.n-switch').first();
      await expect(feishuSwitch).toBeVisible();
      await expect(feishuSwitch).toBeEnabled();

      // 验证 MCP 开关可以交互
      const mcpCard = page.locator('.n-card:has-text("MCP 服务配置")');
      const mcpSwitch = mcpCard.locator('.n-switch');
      await expect(mcpSwitch).toBeVisible();
      await expect(mcpSwitch).toBeEnabled();

      // 验证服务端日志按钮可以交互
      const logCard = page.locator('.n-card:has-text("服务端日志")');
      const logButton = logCard.locator('button:has-text("查看服务端日志")');
      await expect(logButton).toBeVisible();
      await expect(logButton).toBeEnabled();
    });
  });
});
