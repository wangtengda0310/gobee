/**
 * 协议重放页 - 重播功能 E2E 测试（简化版）
 *
 * 测试流程：
 * 1. 加载测试用例
 * 2. 点选表格中的一个 Req 消息
 * 3. 点击重发按钮
 * 4. 断言重放结果出现刚才重播的数据
 *
 * 运行方式（需先启动 Wails 桌面应用）：
 *   wails3 dev
 *   npx playwright test stream-proxy-replay-simple.spec.ts
 */
import { test, expect, describe } from '../shared/fixtures';
import { ProtoTestPage } from '../shared/pages/ProtoTestPage';
import { sleep } from '../shared/utils/helpers';

describe('协议重放页 - 重播功能简化测试', () => {
  let page: ProtoTestPage;

  test.beforeEach(async ({ page: p }) => {
    page = new ProtoTestPage(p);
    await page.goto();
  });

  describe('核心重播功能', () => {
    test('点选 Req 后重播，验证重放数据追加到表格', async () => {
      // 1. 切换到测试用例页签
      await page.clickTabTestcase();

      // 2. 选择测试用例
      await page.selectCaseFromDropdown('添加黄金');
      await sleep(3000);

      // 3. 获取初始表格行数
      const initialRowCount = await page.getTableRowCount();
      console.log(`初始表格行数: ${initialRowCount}`);
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

      // 7. 点击重发按钮（默认重发1次）
      const retryButton = page.replayRetryButton;
      await retryButton.click();
      console.log('已点击重发按钮');

      // 8. 等待重放完成
      await sleep(8000);

      // 9. 断言：表格行数增加
      const newRowCount = await page.getTableRowCount();
      console.log(`重放后表格行数: ${newRowCount}, 增加了 ${newRowCount - initialRowCount} 行`);
      expect(newRowCount).toBeGreaterThan(initialRowCount);

      // 10. 断言：新追加的行中包含相同消息名称
      let foundMatchingMsg = false;
      const newRows = page.getTableRows();

      for (let i = initialRowCount; i < newRowCount; i++) {
        const newRowText = await newRows.nth(i).textContent();
        if (newRowText && newRowText.includes(selectedMsgName)) {
          foundMatchingMsg = true;
          console.log(`✅ 在行 ${i} 找到匹配的消息: ${selectedMsgName}`);
          break;
        }
      }

      expect(foundMatchingMsg).toBe(true);
    });

    test('多次重播验证数据持续追加', async () => {
      // 1. 切换到测试用例页签并选择用例
      await page.clickTabTestcase();
      await page.selectCaseFromDropdown('添加黄金');
      await sleep(3000);

      // 2. 选中第一个 Req
      const rows = page.getTableRows();
      const initialRowCount = await rows.count();
      let firstReqIndex = -1;

      for (let i = 0; i < Math.min(initialRowCount, 10); i++) {
        const rowText = await rows.nth(i).textContent();
        if (rowText && rowText.includes('C->S') && !rowText.includes('S->C')) {
          firstReqIndex = i;
          break;
        }
      }

      await rows.nth(firstReqIndex).click();
      await sleep(500);

      // 3. 第一次重播（重发2次）
      await page.setRepeatCount(2);
      await page.replayRetryButton.click();
      await sleep(8000);

      const rowCountAfterFirst = await page.getTableRowCount();
      console.log(`第一次重播后行数: ${rowCountAfterFirst}`);

      // 4. 第二次重播（重发1次）
      await page.setRepeatCount(1);
      await page.replayRetryButton.click();
      await sleep(8000);

      const rowCountAfterSecond = await page.getTableRowCount();
      console.log(`第二次重播后行数: ${rowCountAfterSecond}`);

      // 5. 验证行数持续增加
      expect(rowCountAfterSecond).toBeGreaterThan(rowCountAfterFirst);
    });
  });

  describe('重播功能状态验证', () => {
    test('重播完成后状态正确恢复', async () => {
      // 1. 切换到测试用例页签并选择用例
      await page.clickTabTestcase();
      await page.selectCaseFromDropdown('添加黄金');
      await sleep(3000);

      // 2. 选中第一个 Req
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

      // 3. 点击重发按钮
      await page.replayRetryButton.click();

      // 4. 等待重放中状态出现
      await sleep(1000);
      const statusWhileRunning = await page.getReplayStatusText();
      console.log(`重放中状态: ${statusWhileRunning}`);

      // 5. 等待重放完成
      await page.waitForReplayComplete(15000);
      const finalStatus = await page.getReplayStatusText();
      console.log(`重放完成状态: ${finalStatus}`);

      // 6. 断言：重放完成后状态标签消失或显示完成
      if (finalStatus) {
        expect(finalStatus).not.toMatch(/正在|运行中/);
      }
    });
  });
});
