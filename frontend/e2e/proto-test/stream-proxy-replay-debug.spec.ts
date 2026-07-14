/**
 * 协议重放页 - 重播功能 E2E 测试（调试版）
 *
 * 运行方式：cd frontend && npx playwright test stream-proxy-replay-debug.spec.ts --reporter=list
 */
import { test, expect, describe } from '../shared/fixtures';
import { ProtoTestPage } from '../shared/pages/ProtoTestPage';
import { sleep } from '../shared/utils/helpers';

describe('协议重放页 - 重播功能调试', () => {
  let page: ProtoTestPage;

  test.beforeEach(async ({ page: p }) => {
    page = new ProtoTestPage(p);
    await page.goto();
  });

  test('调试：测试用例下拉框选择', async () => {
    // 1. 切换到测试用例页签
    await page.clickTabTestcase();

    // 2. 检查下拉框是否存在
    const caseSelect = page.caseSelect;
    await expect(caseSelect).toBeVisible();
    console.log('✅ 下拉框可见');

    // 3. 点击下拉框
    await caseSelect.click();
    await sleep(1000);
    console.log('✅ 点击了下拉框');

    // 4. 检查下拉菜单是否出现
    const menu = page.page.locator('.n-base-select-menu');
    const menuCount = await menu.count();
    console.log(`下拉菜单数量: ${menuCount}`);

    if (menuCount > 0) {
      console.log('✅ 下拉菜单已出现');

      // 5. 获取所有选项
      const options = page.page.locator('.n-base-select-option');
      const optionCount = await options.count();
      console.log(`选项数量: ${optionCount}`);

      // 6. 列出所有选项
      for (let i = 0; i < Math.min(optionCount, 5); i++) {
        const text = await options.nth(i).textContent();
        console.log(`选项 ${i}: ${text}`);
      }

      // 7. 检查是否有"添加黄金"选项
      const goldOption = options.filter({ hasText: '添加黄金' });
      const goldOptionCount = await goldOption.count();
      console.log(`"添加黄金"选项数量: ${goldOptionCount}`);

      if (goldOptionCount > 0) {
        console.log('✅ 找到"添加黄金"选项');

        // 8. 点击选择
        await goldOption.first().click();
        await sleep(500);
        console.log('✅ 点击了"添加黄金"选项');
      } else {
        console.log('❌ 未找到"添加黄金"选项');
      }
    } else {
      console.log('❌ 下拉菜单未出现');
    }

    // 等待一段时间观察界面
    await sleep(3000);
  });

  test('调试：表格数据和行选择', async () => {
    // 1. 切换到测试用例页签
    await page.clickTabTestcase();

    // 2. 尝试选择一个简单的测试用例名称（如果有）
    try {
      await page.selectCaseFromDropdown('添加黄金');
      await sleep(2000);
      console.log('✅ 选择了"添加黄金"用例');
    } catch (e) {
      console.log('❌ 选择"添加黄金"用例失败:', e);
      // 尝试选择第一个可用的用例
      await page.caseSelect.click();
      await sleep(500);
      const firstOption = page.page.locator('.n-base-select-option').first();
      const optionCount = await page.page.locator('.n-base-select-option').count();

      if (optionCount > 0) {
        await firstOption.click();
        await sleep(2000);
        console.log('✅ 选择了第一个可用用例');
      }
    }

    // 3. 检查表格数据
    const rows = page.getTableRows();
    const rowCount = await rows.count();
    console.log(`表格行数: ${rowCount}`);

    if (rowCount > 0) {
      console.log('✅ 表格有数据');

    // 4. 检查前几行的内容，分析数据格式
      for (let i = 0; i < Math.min(rowCount, 5); i++) {
        const rowText = await rows.nth(i).textContent();
        console.log(`行 ${i} 完整内容: "${rowText}"`);
        console.log(`行 ${i} 是否包含C->S: ${rowText?.includes('C->S')}`);
        console.log(`行 ${i} 是否包含→: ${rowText?.includes('→')}`);
      }

      // 5. 尝试点击第一个包含"→"的行
      for (let i = 0; i < Math.min(rowCount, 10); i++) {
        const rowText = await rows.nth(i).textContent();
        if (rowText && rowText.includes('→')) {
          console.log(`找到 Req 消息在行 ${i}`);
          await rows.nth(i).click();
          await sleep(500);
          console.log('✅ 点击了该行');

          // 检查重放控制面板是否出现
          const replayPanel = page.page.locator('text=重放控制').first();
          const panelCount = await replayPanel.count();
          console.log(`重放控制面板数量: ${panelCount}`);

          if (panelCount > 0) {
            console.log('✅ 重放控制面板已显示');
          }

          break;
        }
      }
    } else {
      console.log('❌ 表格没有数据');
    }

    // 等待一段时间观察界面
    await sleep(3000);
  });
});
