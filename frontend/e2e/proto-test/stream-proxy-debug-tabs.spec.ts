/**
 * 协议重放页 - 调试页签选择
 */
import { test } from '../shared/fixtures';
import { sleep } from '../shared/utils/helpers';

test('调试：查看页签结构', async ({ page }) => {
  // 导航到 Proto测试页（不是 /Test）
  await page.goto('http://wails.localhost:9245/ProtoTest');
  await sleep(2000);

  // 获取所有可能的页签元素
  const tabs = page.locator('.n-tabs-tab');
  const tabCount = await tabs.count();

  console.log(`找到 ${tabCount} 个页签元素`);

  for (let i = 0; i < tabCount; i++) {
    const text = await tabs.nth(i).textContent();
    console.log(`页签 ${i}: "${text}"`);
  }

  // 尝试不同选择器
  const testcaseTabs = page.locator('[data-tab-key="testcase"]');
  const count = await testcaseTabs.count();
  console.log(`data-tab-key="testcase" 数量: ${count}`);

  await sleep(5000);
});
