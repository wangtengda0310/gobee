import { test, expect } from '@playwright/test';
import { BasePage } from '../shared/pages/BasePage';

test('简化调试布局', async ({ page }) => {
  const basePage = new BasePage(page);

  // 确保在协议重放页面
  await basePage.gotoProtoTest();

  // 等待页面加载
  await page.waitForTimeout(2000);

  // 截图查看当前布局
  await page.screenshot({ path: 'test-results/simple-layout-debug.png', fullPage: true });

  // 简单检查页签是否存在
  const tabs = await page.locator('div').filter({ hasText: /^发包改包$/ }).count();
  console.log('页签数量:', tabs);
  expect(tabs).toBeGreaterThan(0);
});
