/**
 * 设置页测试
 * 对应 src/pages/settings/index.vue
 */

import { test, expect, describe } from '../shared/fixtures';
import { resolveRoute } from '../shared/pages/BasePage';
import { sleep } from '../shared/utils/helpers';

describe('设置页测试', () => {
  test.beforeEach(async ({ page }) => {
    // 先加载根页面
    await page.goto(resolveRoute(page, '/'));
    await page.waitForLoadState('networkidle');

    // 点击设置导航按钮
    await page.locator('#layout-header button:has-text("设置")').click();
    await sleep(500);
  });

  describe('页面加载', () => {
    test('页面加载 - 显示配置卡片', async ({ page }) => {
      // 验证设置页面容器存在
      await expect(page.locator('#settings-page')).toBeVisible();

      // 验证飞书配置卡片存在
      await expect(page.locator('.n-card:has-text("飞书通知配置")')).toBeVisible();

      // 验证 MCP 配置卡片存在
      await expect(page.locator('.n-card:has-text("MCP 服务配置")')).toBeVisible();

      // 验证服务端日志卡片存在
      await expect(page.locator('.n-card:has-text("服务端日志")')).toBeVisible();
    });
  });

  describe('飞书通知配置', () => {
    test('飞书通知开关 - 验证开关存在', async ({ page }) => {
      const card = page.locator('.n-card:has-text("飞书通知配置")');
      const switchEl = card.locator('.n-switch').first();

      await expect(switchEl).toBeVisible();
      await expect(switchEl).toBeEnabled();
    });

    test('机器人GUID输入 - 验证输入框存在', async ({ page }) => {
      const card = page.locator('.n-card:has-text("飞书通知配置")');
      const input = card.locator('input[placeholder*="GUID"]');

      await expect(input).toBeVisible();
      await expect(input).toBeEnabled();
    });

    test('消息劫持开关 - 验证开关存在', async ({ page }) => {
      const card = page.locator('.n-card:has-text("飞书通知配置")');
      const switches = card.locator('.n-switch');
      const secondSwitch = switches.nth(1);

      await expect(secondSwitch).toBeVisible();
      await expect(secondSwitch).toBeEnabled();
    });
  });

  describe('MCP 服务配置', () => {
    test('MCP 配置卡片 - 显示MCP配置', async ({ page }) => {
      const card = page.locator('.n-card:has-text("MCP 服务配置")');
      await expect(card).toBeVisible();
    });

    test('MCP 开关 - 验证开关存在', async ({ page }) => {
      const card = page.locator('.n-card:has-text("MCP 服务配置")');
      const switchEl = card.locator('.n-switch');

      await expect(switchEl).toBeVisible();
      await expect(switchEl).toBeEnabled();
    });

    test('绑定地址 - 验证输入框存在', async ({ page }) => {
      const card = page.locator('.n-card:has-text("MCP 服务配置")');
      const input = card.locator('input[placeholder="127.0.0.1"]');

      await expect(input).toBeVisible();
      await expect(input).toBeEnabled();
    });

    test('端口配置 - 验证输入框存在', async ({ page }) => {
      const card = page.locator('.n-card:has-text("MCP 服务配置")');
      const portInput = card.locator('.n-input-number');

      await expect(portInput).toBeVisible();
      await expect(portInput).toBeEnabled();
    });

    test('运行状态标签 - 验证状态显示', async ({ page }) => {
      const card = page.locator('.n-card:has-text("MCP 服务配置")');
      const statusTag = card.locator('.n-tag');

      await expect(statusTag).toBeVisible();
      // 标签应该显示运行中或已停止
      const text = await statusTag.textContent();
      expect(['运行中', '已停止']).toContain(text);
    });
  });

  describe('服务端日志', () => {
    test('日志计数 - 显示已捕获日志数量', async ({ page }) => {
      const card = page.locator('.n-card:has-text("服务端日志")');
      const logCount = card.locator('text=/\\d+ 条/');

      await expect(logCount).toBeVisible();
    });

    test('查看服务端日志 - 验证按钮存在', async ({ page }) => {
      const card = page.locator('.n-card:has-text("服务端日志")');
      const button = card.locator('button:has-text("查看服务端日志")');

      await expect(button).toBeVisible();
      await expect(button).toBeEnabled();
    });
  });

  describe('页面布局验证', () => {
    test('页面布局 - 验证所有卡片显示', async ({ page }) => {
      // 验证设置页面容器
      await expect(page.locator('#settings-page')).toBeVisible();

      // 验证所有卡片都存在
      const cards = page.locator('.n-card');
      const cardCount = await cards.count();
      expect(cardCount).toBeGreaterThanOrEqual(3);
    });

    test('卡片滚动 - 验证可以滚动查看所有内容', async ({ page }) => {
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
    });
  });
});
