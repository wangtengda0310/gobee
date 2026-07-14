import { test, expect } from '@playwright/test';
import { ProtoTestPage } from '../shared/pages/ProtoTestPage';

test('调试布局顺序', async ({ page }) => {
  const protoTestPage = new ProtoTestPage(page);

  // 截图查看当前布局
  await page.screenshot({ path: 'test-results/layout-debug.png', fullPage: true });

  // 截图查看当前布局
  await page.screenshot({ path: 'test-results/layout-debug.png', fullPage: true });

  // 检查DOM顺序
  const tabs = await page.locator('div').filter({ hasText: /^发包改包$/ }).all();
  const serverInputs = await page.locator('input[placeholder*="TCP"]').all();
  const recordButtons = await page.locator('button').filter({ hasText: '开始录制' }).all();

  console.log('页签数量:', tabs.length);
  console.log('TCP输入框数量:', serverInputs.length);
  console.log('开始录制按钮数量:', recordButtons.length);

  // 获取DOM顺序
  const tabElement = tabs[0];
  const serverInput = serverInputs[0];
  const recordButton = recordButtons[0];

  // 检查元素在页面中的位置（从上到下）
  const tabBoundingBox = await tabElement.boundingBox();
  const serverBoundingBox = await serverInput.boundingBox();
  const recordBoundingBox = await recordButton.boundingBox();

  console.log('页签Y坐标:', tabBoundingBox?.y);
  console.log('服务器输入框Y坐标:', serverBoundingBox?.y);
  console.log('录制按钮Y坐标:', recordBoundingBox?.y);

  // 验证顺序：页签 → 服务器输入框 → 录制按钮
  if (tabBoundingBox && serverBoundingBox && recordBoundingBox) {
    expect(serverBoundingBox.y).toBeGreaterThan(tabBoundingBox.y);
    expect(recordBoundingBox.y).toBeGreaterThan(serverBoundingBox.y);
    console.log('✅ 布局顺序正确：页签 → 目标服务 → 按钮');
  } else {
    console.log('❌ 无法获取元素位置');
  }
});
