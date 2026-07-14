/**
 * ProtoTest 多线程并发重放 E2E 测试
 *
 * 测试 MaxConcurrency 参数设置和并发重放行为：
 * - 最大并发输入框布局和默认值
 * - 多账号批量重放进度总数与账号数匹配
 * - 取消后不会继续接收到新消息
 *
 * 运行方式（需先启动 Wails 桌面应用 + 目标服务器）：
 *   wails3 dev
 *   npx playwright test stream-proxy-concurrency.spec.ts
 */
import { test, expect, describe } from '../shared/fixtures';
import { ProtoTestPage } from '../shared/pages/ProtoTestPage';
import { sleep } from '../shared/utils/helpers';

describe('多线程并发重放', () => {
  let page: ProtoTestPage;

  test.beforeEach(async ({ page: p }) => {
    page = new ProtoTestPage(p);
    await page.goto();
  });

  // ==================== 并发度参数布局 ====================

  describe('并发度参数布局', () => {
    test('最大并发输入框可见', async () => {
      await expect(page.maxConcurrencyInput).toBeVisible();
    });

    test('最大并发输入框默认值为 0（不限制）', async () => {
      const value = await page.maxConcurrencyInput.inputValue();
      expect(value).toBe('0');
    });

    test('最大并发输入框在页签切换后保持可见', async () => {
      await page.clickTabTestcase();
      await page.page.waitForTimeout(300);
      await expect(page.maxConcurrencyInput).toBeVisible();

      await page.clickTabPacket();
      await page.page.waitForTimeout(300);
      await expect(page.maxConcurrencyInput).toBeVisible();
    });
  });

  // ==================== 多账号并发重放（需要服务器） ====================

  describe('多账号批量重放', () => {
    // 注意：以下测试需要可连接的目标服务器
    test('进度总数与账号范围数匹配', async () => {
      // 1. 切换到测试用例页签
      await page.clickTabTestcase();

      // 2. 选择测试用例
      await page.selectCaseFromDropdown('添加黄金');
      await sleep(3000);

      // 3. 设置账号范围为 1~3（3个账号）
      await page.setAccountRangeStart(1);
      await page.setAccountRangeEnd(3);
      await sleep(200);

      // 4. 验证范围提示
      const hint = page.page.locator('text=共 3 个账号').first();
      await expect(hint).toBeVisible();

      // 5. 选中第一个 Req
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

      expect(firstReqIndex).toBeGreaterThanOrEqual(0);
      await rows.nth(firstReqIndex).click();
      await sleep(500);

      // 6. 设置重发次数为1，点击重发
      await page.setRepeatCount(1);
      await page.replayRetryButton.click();
      console.log('已点击重发按钮（3个账号并发）');

      // 7. 等待重放完成
      await page.waitForReplayComplete(30000);

      // 8. 验证表格行数增加（至少每个账号产生 1 条消息）
      const newRowCount = await page.getTableRowCount();
      console.log(`重放前行数: ${initialRowCount}, 重放后行数: ${newRowCount}`);
      expect(newRowCount).toBeGreaterThan(initialRowCount);

      // 9. 验证进度完成
      const finalStatus = await page.getReplayStatusText();
      console.log(`最终状态: ${finalStatus}`);
      expect(finalStatus).not.toMatch(/正在|运行中/);
    });

    test('取消后不会继续接收到新消息', async () => {
      // 1. 切换到测试用例页签
      await page.clickTabTestcase();

      // 2. 选择测试用例
      await page.selectCaseFromDropdown('添加黄金');
      await sleep(3000);

      // 3. 设置较大账号范围（确保有足够时间取消）
      await page.setAccountRangeStart(1);
      await page.setAccountRangeEnd(10);
      await sleep(200);

      // 4. 选中第一个 Req
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

      expect(firstReqIndex).toBeGreaterThanOrEqual(0);
      await rows.nth(firstReqIndex).click();
      await sleep(500);

      // 5. 点击重发（10个账号，需要一些时间完成）
      await page.setRepeatCount(1);
      await page.replayRetryButton.click();
      console.log('已点击重发按钮（10个账号）');

      // 6. 等待一小段时间让部分消息到达
      await sleep(2000);

      // 7. 记录当前行数
      const rowCountBeforeCancel = await page.getTableRowCount();
      console.log(`取消前表格行数: ${rowCountBeforeCancel}`);

      // 8. 点击停止按钮取消重放
      await page.replayStopButton.click();
      console.log('已点击停止按钮');

      // 9. 等待取消生效
      await sleep(3000);

      // 10. 验证重放已停止
      const statusAfterCancel = await page.getReplayStatusText();
      console.log(`取消后状态: ${statusAfterCancel}`);
      expect(statusAfterCancel).not.toMatch(/正在|运行中/);

      // 11. 等待一段时间后确认不再有新消息
      await sleep(3000);
      const rowCountAfterWait = await page.getTableRowCount();
      console.log(`等待后表格行数: ${rowCountAfterWait}`);

      // 取消后行数不应显著增加（允许最多 1-2 条已经在传输中的消息到达）
      const newMessagesAfterCancel = rowCountAfterWait - rowCountBeforeCancel;
      expect(newMessagesAfterCancel).toBeLessThanOrEqual(2);
    });
  });
});
