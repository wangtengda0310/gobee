/**
 * 协议重放页 - 重播功能 E2E 测试（正确版本）
 *
 * 测试流程：
 * 1. 加载测试用例
 * 2. 点选表格中的一个 Req 消息
 * 3. 点击重发按钮
 * 4. 切换到重放结果页签
 * 5. 在重放结果页签的表格中验证是否出现重播的数据
 *
 * 运行方式（需先启动 Wails 桌面应用）：
 *   wails3 dev
 *   npx playwright test stream-proxy-replay-correct.spec.ts
 */
import { test, expect, describe } from '../shared/fixtures';
import { ProtoTestPage } from '../shared/pages/ProtoTestPage';
import { sleep } from '../shared/utils/helpers';

describe('协议重放页 - 重播功能正确测试', () => {
  let page: ProtoTestPage;

  test.beforeEach(async ({ page: p }) => {
    page = new ProtoTestPage(p);
    await page.goto();
  });

  describe('重播功能完整流程测试', () => {
    test('点选 Req 后重播，在重放结果页签中验证数据', async () => {
      // 1. 切换到测试用例页签
      await page.clickTabTestcase();

      // 2. 选择测试用例
      await page.selectCaseFromDropdown('添加黄金');
      await sleep(3000);

      // 3. 获取初始表格行数
      const initialRowCount = await page.getTableRowCount();
      console.log(`测试用例表格行数: ${initialRowCount}`);
      expect(initialRowCount).toBeGreaterThan(0);

      // 4. 找到第一个 Req 消息并选中
      const rows = page.getTableRows();
      let firstReqIndex = -1;
      let selectedMsgName = '';

      for (let i = 0; i < Math.min(initialRowCount, 10); i++) {
        const rowText = await rows.nth(i).textContent();
        if (rowText && rowText.includes('C->S') && !rowText.includes('S->C')) {
          firstReqIndex = i;
          const msgNameMatch = rowText.match(/([A-Z][a-zA-Z0-9]+Req)/);
          if (msgNameMatch) {
            selectedMsgName = msgNameMatch[1];
          }
          break;
        }
      }

      expect(firstReqIndex).toBeGreaterThanOrEqual(0);
      console.log(`选中行 ${firstReqIndex}, 消息: ${selectedMsgName}`);

      // 5. 点击该行选中
      await rows.nth(firstReqIndex).click();
      await sleep(500);

      // 6. 验证重放控制面板已显示
      const replayPanel = page.page.locator('text=重放控制').first();
      await expect(replayPanel).toBeVisible();

      // 7. 记录测试用例页签的当前行数（用于后续对比）
      const testcaseRowCountBefore = await page.getTableRowCount();
      console.log(`测试用例页签行数（重播前）: ${testcaseRowCountBefore}`);

      // 8. 点击重发按钮（默认重发1次）
      const retryButton = page.replayRetryButton;
      await retryButton.click();
      console.log('已点击重发按钮');

      // 9. 等待重放完成
      await sleep(8000);

      // 10. 切换到"重放结果"页签
      // 先找到重放结果页签
      const tabReplayResult = page.page.locator('div').filter({ hasText: /^重放结果$/ }).first();
      await expect(tabReplayResult).toBeVisible();
      await tabReplayResult.click();
      await sleep(1000);

      console.log('已切换到重放结果页签');

      // 11. 验证重放结果页签的表格有数据
      const resultRows = page.getTableRows();
      const resultRowCount = await resultRows.count();
      console.log(`重放结果页签表格行数: ${resultRowCount}`);
      expect(resultRowCount).toBeGreaterThan(0);

      // 12. 在重放结果表格中查找相同消息名称的数据
      let foundMatchingMsg = false;

      for (let i = 0; i < Math.min(resultRowCount, 20); i++) {
        const rowText = await resultRows.nth(i).textContent();
        if (rowText && rowText.includes(selectedMsgName)) {
          foundMatchingMsg = true;
          console.log(`✅ 在重放结果页签行 ${i} 找到匹配的消息: ${selectedMsgName}`);
          console.log(`该行内容: ${rowText}`);
          break;
        }
      }

      expect(foundMatchingMsg).toBe(true);
    });

    test('多次重播在重放结果页签中验证数据累积', async () => {
      // 1. 切换到测试用例页签
      await page.clickTabTestcase();

      // 2. 选择测试用例
      await page.selectCaseFromDropdown('添加黄金');
      await sleep(3000);

      // 3. 选中第一个 Req
      const rows = page.getTableRows();
      let firstReqIndex = -1;
      let selectedMsgName = '';

      for (let i = 0; i < Math.min(await rows.count(), 10); i++) {
        const rowText = await rows.nth(i).textContent();
        if (rowText && rowText.includes('C->S') && !rowText.includes('S->C')) {
          firstReqIndex = i;
          const msgNameMatch = rowText.match(/([A-Z][a-zA-Z0-9]+Req)/);
          if (msgNameMatch) {
            selectedMsgName = msgNameMatch[1];
          }
          break;
        }
      }

      await rows.nth(firstReqIndex).click();
      await sleep(500);

      // 4. 第一次重播（重发2次）
      await page.setRepeatCount(2);
      await page.replayRetryButton.click();
      await sleep(10000);

      // 5. 切换到重放结果页签验证第一次重播结果
      const tabReplayResult = page.page.locator('div').filter({ hasText: /^重放结果$/ }).first();
      await tabReplayResult.click();
      await sleep(1000);

      const resultRows1 = page.getTableRows();
      const resultRowCount1 = await resultRows1.count();
      console.log(`第一次重播后，重放结果页签行数: ${resultRowCount1}`);
      expect(resultRowCount1).toBeGreaterThan(0);

      // 验证包含选中的消息
      let foundFirst1 = false;
      for (let i = 0; i < Math.min(resultRowCount1, 20); i++) {
        const rowText = await resultRows1.nth(i).textContent();
        if (rowText && rowText.includes(selectedMsgName)) {
          foundFirst1 = true;
          break;
        }
      }
      expect(foundFirst1).toBe(true);

      // 6. 切回测试用例页签，进行第二次重播
      const tabTestcase = page.page.locator('div').filter({ hasText: /^测试用例$/ }).first();
      await tabTestcase.click();
      await sleep(1000);

      await page.setRepeatCount(1);
      await page.replayRetryButton.click();
      await sleep(10000);

      // 7. 再次切换到重放结果页签验证累积数据
      await tabReplayResult.click();
      await sleep(1000);

      const resultRows2 = page.getTableRows();
      const resultRowCount2 = await resultRows2.count();
      console.log(`第二次重播后，重放结果页签行数: ${resultRowCount2}`);

      // 验证行数增加
      expect(resultRowCount2).toBeGreaterThan(resultRowCount1);

      // 8. 验证包含选中的消息
      let foundFirst2 = false;
      for (let i = 0; i < Math.min(resultRowCount2, 20); i++) {
        const rowText = await resultRows2.nth(i).textContent();
        if (rowText && rowText.includes(selectedMsgName)) {
          foundFirst2 = true;
          break;
        }
      }
      expect(foundFirst2).toBe(true);
    });
  });

  describe('重放结果页签功能验证', () => {
    test('重放结果页签显示重放结果选择器', async () => {
      // 1. 切换到重放结果页签
      const tabReplayResult = page.page.locator('div').filter({ hasText: /^重放结果$/ }).first();
      await tabReplayResult.click();
      await sleep(1000);

      // 2. 验证重放结果选择器可见
      const resultSelector = page.page.locator('text=重放结果:').first();
      await expect(resultSelector).toBeVisible();

      // 3. 验证清空按钮存在
      const clearButton = page.page.locator('button:has-text("清空")').first();
      await expect(clearButton).toBeVisible();
    });

    test('执行重播后重放结果页签自动显示新结果', async () => {
      // 1. 切换到测试用例页签
      await page.clickTabTestcase();

      // 2. 选择测试用例
      await page.selectCaseFromDropdown('添加黄金');
      await sleep(3000);

      // 3. 选中第一个 Req 并重播
      const rows = page.getTableRows();
      let firstReqIndex = -1;

      for (let i = 0; i < Math.min(await rows.count(), 10); i++) {
        const rowText = await rows.nth(i).textContent();
        if (rowText && rowText.includes('C->S') && !rowText.includes('S->C')) {
          firstReqIndex = i;
          break;
        }
      }

      await rows.nth(firstReqIndex).click();
      await sleep(500);

      // 4. 点击重发按钮
      await page.replayRetryButton.click();
      await sleep(8000);

      // 5. 切换到重放结果页签
      const tabReplayResult = page.page.locator('div').filter({ hasText: /^重放结果$/ }).first();
      await tabReplayResult.click();
      await sleep(1000);

      // 6. 验证重放结果选择器有选项
      const resultSelector = page.page.locator('.n-select').filter({ hasText: '选择重放结果' }).first();
      await expect(resultSelector).toBeVisible();

      // 点击下拉框查看是否有新的重放结果
      await resultSelector.click();
      await sleep(500);

      const options = page.page.locator('.n-base-select-option');
      const optionCount = await options.count();
      console.log(`重放结果选项数量: ${optionCount}`);

      // 验证至少有一个重放结果
      expect(optionCount).toBeGreaterThan(0);

      // 关闭下拉框
      await page.page.keyboard.press('Escape');
    });
  });
});
