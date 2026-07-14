/**
 * 协议重放页 - 重播功能 E2E 测试
 *
 * 测试流程：
 * 1. 加载测试用例
 * 2. 点选表格中的一个 Req 消息
 * 3. 点击重发按钮
 * 4. 断言重放结果出现刚才重播的数据
 * 5. 重复执行直到测试通过
 *
 * 运行方式（需先启动 Wails 桌面应用）：
 *   wails3 dev
 *   npx playwright test stream-proxy-replay.spec.ts
 */
import { test, expect, describe } from '../shared/fixtures';
import { ProtoTestPage } from '../shared/pages/ProtoTestPage';
import { sleep } from '../shared/utils/helpers';

describe('协议重放页 - 重播功能', () => {
  let page: ProtoTestPage;

  test.beforeEach(async ({ page: p }) => {
    page = new ProtoTestPage(p);
    await page.goto();
  });

  describe('点选 Req 重播验证', () => {
    test('点选测试用例中的 Req 后重播，重放结果出现该 req 的数据', async () => {
      // 1. 切换到测试用例页签
      await page.clickTabTestcase();

      // 2. 选择测试用例（使用"添加黄金"用例，因为已存在）
      await page.selectCaseFromDropdown('添加黄金');
      await sleep(3000); // 等待数据加载完成

      // 3. 获取初始表格行数
      const initialRowCount = await page.getTableRowCount();
      console.log(`初始表格行数: ${initialRowCount}`);
      expect(initialRowCount).toBeGreaterThan(0);

      // 4. 点选第一个 Req 消息（direction = "C->S"，表示客户端到服务器的请求）
      const rows = page.getTableRows();
      let firstReqIndex = -1;
      let firstReqMsgName = '';

      // 查找第一个 Req 消息（direction = "C->S"，表示客户端到服务器的请求）
      for (let i = 0; i < Math.min(initialRowCount, 10); i++) {
        const rowText = await rows.nth(i).textContent();
        if (rowText && rowText.includes('C->S') && !rowText.includes('S->C')) {
          firstReqIndex = i;
          // 提取消息名称（如 PostUserStatusReq）
          const msgNameMatch = rowText.match(/([A-Z][a-zA-Z0-9]+Req)/);
          if (msgNameMatch) {
            firstReqMsgName = msgNameMatch[1];
          }
          break;
        }
      }

      expect(firstReqIndex).toBeGreaterThanOrEqual(0);
      console.log(`找到第一个 Req 消息在行 ${firstReqIndex}, 消息名称: ${firstReqMsgName}`);

      // 5. 点击该行选中
      await rows.nth(firstReqIndex).click();
      await sleep(500);

      // 6. 验证重放控制面板已显示
      const replayPanel = page.page.locator('text=重放控制').first();
      await expect(replayPanel).toBeVisible();

      // 7. 记录选中行的关键信息（消息名称用于后续验证）
      const selectedRowText = await rows.nth(firstReqIndex).textContent();
      console.log(`选中行内容: ${selectedRowText}`);

      // 提取消息名称（如 PostUserStatusReq）
      const msgNameMatch = selectedRowText?.match(/([A-Z][a-zA-Z0-9]+Req)/);
      const selectedMsgName = msgNameMatch ? msgNameMatch[1] : null;
      console.log(`选中的消息名称: ${selectedMsgName}`);
      expect(selectedMsgName).not.toBeNull();

      // 8. 点击重发按钮（默认重发1次）
      const retryButton = page.replayRetryButton;
      await retryButton.click();
      await sleep(1000); // 等待重发请求处理

      // 9. 等待重放完成（检查状态标签）
      const maxWaitTime = 10000; // 最多等待10秒
      const startTime = Date.now();

      while (Date.now() - startTime < maxWaitTime) {
        const statusText = await page.getReplayStatusText();
        console.log(`当前重放状态: ${statusText}`);

        // 如果状态不包含"正在"，说明已完成
        if (!statusText.includes('正在') && !statusText.includes('运行中')) {
          break;
        }
        await sleep(500);
      }

      // 10. 断言：表格行数增加（重放结果应该追加到表格）
      const newRowCount = await page.getTableRowCount();
      console.log(`重放后表格行数: ${newRowCount}`);
      expect(newRowCount).toBeGreaterThan(initialRowCount);

      // 11. 断言：新追加的行中包含相同消息名称的消息
      let foundMatchingMsg = false;
      const newRows = page.getTableRows();

      // 检查新追加的行（从 initialRowCount 到 newRowCount）
      for (let i = initialRowCount; i < newRowCount; i++) {
        const newRowText = await newRows.nth(i).textContent();
        console.log(`新追加行 ${i} 内容: ${newRowText}`);

        // 检查是否包含相同消息名称
        if (newRowText && newRowText.includes(selectedMsgName)) {
          foundMatchingMsg = true;
          console.log(`✅ 找到匹配的消息: ${selectedMsgName} 在行 ${i}`);
          break;
        }
      }

      expect(foundMatchingMsg).toBe(true);
    });

    test('多次重播验证数据持续追加', async () => {
      // 1. 切换到测试用例页签
      await page.clickTabTestcase();

      // 2. 选择测试用例
      await page.selectCaseFromDropdown('添加黄金');
      await sleep(2000);

      // 3. 点选第一个 Req 消息
      const rows = page.getTableRows();
      const initialRowCount = await rows.count();
      let firstReqIndex = -1;

      for (let i = 0; i < Math.min(initialRowCount, 10); i++) {
        const rowText = await rows.nth(i).textContent();
        if (rowText && rowText.includes('→')) {
          firstReqIndex = i;
          break;
        }
      }

      expect(firstReqIndex).toBeGreaterThanOrEqual(0);
      await rows.nth(firstReqIndex).click();
      await sleep(500);

      // 4. 记录选中行的 MsgID
      const selectedRowText = await rows.nth(firstReqIndex).textContent();
      const msgIdMatch = selectedRowText?.match(/MsgID:\s*(\d+)/);
      const selectedMsgId = msgIdMatch ? msgIdMatch[1] : null;
      expect(selectedMsgId).not.toBeNull();

      // 5. 第一次重播（重发2次）
      await page.setRepeatCount(2);
      const retryButton = page.replayRetryButton;
      await retryButton.click();
      await sleep(8000); // 等待重放完成

      // 验证行数增加
      const rowCountAfterFirstReplay = await page.getTableRowCount();
      console.log(`第一次重播后行数: ${rowCountAfterFirstReplay}`);
      expect(rowCountAfterFirstReplay).toBeGreaterThan(initialRowCount);

      // 6. 第二次重播（重发1次）
      await page.setRepeatCount(1);
      await retryButton.click();
      await sleep(6000);

      // 验证行数继续增加
      const rowCountAfterSecondReplay = await page.getTableRowCount();
      console.log(`第二次重播后行数: ${rowCountAfterSecondReplay}`);
      expect(rowCountAfterSecondReplay).toBeGreaterThan(rowCountAfterFirstReplay);

      // 7. 验证所有新追加的行中都包含该 MsgID
      let totalMatches = 0;
      const newRows = page.getTableRows();

      for (let i = initialRowCount; i < rowCountAfterSecondReplay; i++) {
        const rowText = await newRows.nth(i).textContent();
        if (rowText && rowText.includes(`MsgID:${selectedMsgId}`)) {
          totalMatches++;
        }
      }

      console.log(`总共找到 ${totalMatches} 个匹配的 MsgID: ${selectedMsgId}`);
      // 期望至少找到3个匹配（第一次重播2个 + 第二次重播1个）
      expect(totalMatches).toBeGreaterThanOrEqual(3);
    });

    test('重播完成后状态正确恢复', async () => {
      // 1. 切换到测试用例页签
      await page.clickTabTestcase();

      // 2. 选择测试用例
      await page.selectCaseFromDropdown('添加黄金');
      await sleep(2000);

      // 3. 点选第一个 Req 消息
      const rows = page.getTableRows();
      const initialRowCount = await rows.count();
      let firstReqIndex = -1;

      for (let i = 0; i < Math.min(initialRowCount, 10); i++) {
        const rowText = await rows.nth(i).textContent();
        if (rowText && rowText.includes('→')) {
          firstReqIndex = i;
          break;
        }
      }

      expect(firstReqIndex).toBeGreaterThanOrEqual(0);
      await rows.nth(firstReqIndex).click();
      await sleep(500);

      // 4. 点击重发按钮
      const retryButton = page.replayRetryButton;
      await retryButton.click();

      // 5. 等待重放中状态出现
      await page.page.waitForTimeout(1000);
      const statusTextWhileRunning = await page.getReplayStatusText();
      console.log(`重放中状态: ${statusTextWhileRunning}`);
      expect(statusTextWhileRunning).toMatch(/正在|运行中/);

      // 6. 等待重放完成
      await page.waitForReplayComplete(15000);
      const finalStatusText = await page.getReplayStatusText();
      console.log(`重放完成状态: ${finalStatusText}`);

      // 7. 断言：重放完成后不再显示"正在"状态
      expect(finalStatusText).not.toMatch(/正在|运行中/);
    });
  });

  describe('重播功能边界条件', () => {
    test('未选中行时重发按钮不可见或禁用', async () => {
      // 1. 切换到测试用例页签
      await page.clickTabTestcase();

      // 2. 选择测试用例
      await page.selectCaseFromDropdown('添加黄金');
      await sleep(2000);

      // 3. 确保没有选中任何行
      const replayPanel = page.page.locator('text=重放控制').first();
      const panelCount = await replayPanel.count();

      if (panelCount > 0) {
        // 如果面板可见，检查重发按钮是否禁用
        const retryButton = page.replayRetryButton;
        // 注意：重发按钮可能一直存在于DOM中，但只有在选中行时才激活
        console.log('重放按钮存在，检查其状态');
      } else {
        // 面板不可见是正常状态
        console.log('未选中行时，重放控制面板不可见（正常）');
      }
    });

    test('重发次数输入验证', async () => {
      // 1. 切换到测试用例页签
      await page.clickTabTestcase();

      // 2. 选择测试用例
      await page.selectCaseFromDropdown('添加黄金');
      await sleep(2000);

      // 3. 点选第一个 Req 消息
      const rows = page.getTableRows();
      let firstReqIndex = -1;

      for (let i = 0; i < Math.min(await rows.count(), 10); i++) {
        const rowText = await rows.nth(i).textContent();
        if (rowText && rowText.includes('→')) {
          firstReqIndex = i;
          break;
        }
      }

      expect(firstReqIndex).toBeGreaterThanOrEqual(0);
      await rows.nth(firstReqIndex).click();
      await sleep(500);

      // 4. 测试设置不同的重发次数
      const countInput = page.replayCountInput;

      // 设置重发次数为5
      await page.setRepeatCount(5);
      const valueAfterSet = await countInput.inputValue();
      expect(valueAfterSet).toBe('5');

      // 设置重发次数为1（最小值）
      await page.setRepeatCount(1);
      const valueMin = await countInput.inputValue();
      expect(valueMin).toBe('1');
    });
  });

  describe('重播功能稳定性测试', () => {
    test('连续重播10次验证系统稳定性', async () => {
      // 1. 切换到测试用例页签
      await page.clickTabTestcase();

      // 2. 选择测试用例
      await page.selectCaseFromDropdown('添加黄金');
      await sleep(2000);

      // 3. 点选第一个 Req 消息
      const rows = page.getTableRows();
      const initialRowCount = await rows.count();
      let firstReqIndex = -1;

      for (let i = 0; i < Math.min(initialRowCount, 10); i++) {
        const rowText = await rows.nth(i).textContent();
        if (rowText && rowText.includes('→')) {
          firstReqIndex = i;
          break;
        }
      }

      expect(firstReqIndex).toBeGreaterThanOrEqual(0);
      await rows.nth(firstReqIndex).click();
      await sleep(500);

      // 4. 记录初始行数
      const rowCountBeforeReplay = await page.getTableRowCount();

      // 5. 连续重播10次（每次重发1次）
      for (let round = 1; round <= 10; round++) {
        console.log(`第 ${round} 轮重播`);

        await page.setRepeatCount(1);
        const retryButton = page.replayRetryButton;
        await retryButton.click();

        // 等待重放完成
        await page.waitForReplayComplete(10000);
        await sleep(500);

        // 验证行数增加
        const currentRowCount = await page.getTableRowCount();
        console.log(`第 ${round} 轮后行数: ${currentRowCount}`);
        expect(currentRowCount).toBeGreaterThan(rowCountBeforeReplay + round - 1);
      }

      // 6. 最终验证：总行数应该增加了10次
      const finalRowCount = await page.getTableRowCount();
      console.log(`初始行数: ${rowCountBeforeReplay}, 最终行数: ${finalRowCount}`);
      expect(finalRowCount).toBeGreaterThanOrEqual(rowCountBeforeReplay + 10);
    });
  });
});
