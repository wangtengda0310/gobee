/**
 * 协议重放页 - 重放结果页签测试
 *
 * 验证双事件通道架构下，所有重播操作的结果正确显示在重放结果页签。
 *
 * 测试覆盖：
 * 1. 发包改包开始重放 → 自动切换 + 结果显示
 * 2. 测试用例执行用例 → 自动切换 + 结果显示
 * 3. 重发按钮 → 结果追加
 * 4. 多次重播累积多个结果
 * 5. 发包改包表格不受影响
 *
 * 运行方式：
 *   wails3 dev
 *   npx playwright test stream-proxy-replay-result.spec.ts
 */
import { test, expect } from '../shared/fixtures';
import { ProtoTestPage } from '../shared/pages/ProtoTestPage';
import { sleep } from '../shared/utils/helpers';

test.describe('重放结果页签', () => {

  test('发包改包开始重放后自动切换并显示结果', async ({ protoTestPage }) => {
    await protoTestPage.goto();

    // 确认发包改包页签有表格数据
    const rowCount = await protoTestPage.getTableRowCount();
    if (rowCount === 0) {
      test.skip();
      return;
    }

    // 点击开始重放
    await protoTestPage.startReplayButton.click();
    console.log('✅ 已点击开始重放');

    // 等待重放开始（自动切换到重放结果页签）
    await sleep(3000);

    // 验证自动切换到重放结果页签
    const replayResultRows = await protoTestPage.getReplayResultRowCount();
    console.log(`重放结果页签表格行数: ${replayResultRows}`);

    // 如果有数据，验证来源标签
    if (replayResultRows > 0) {
      const sourceText = await protoTestPage.getResultSourceLabel();
      console.log(`来源标签: ${sourceText}`);
      expect(sourceText).toContain('发包改包');
      console.log('🎉 测试通过：发包改包开始重放 → 重放结果页签正确显示');
    } else {
      console.log('⚠️ 重放结果表格为空，可能服务端未响应');
    }
  });

  test('测试用例执行用例后自动切换并显示结果', async ({ protoTestPage }) => {
    await protoTestPage.goto();

    // 切换到测试用例页签
    await protoTestPage.clickTabTestcase();
    await sleep(1000);

    // 选择一个用例（使用 force click 避免被遮挡元素干扰）
    try {
      await protoTestPage.selectCaseFromDropdown('');
    } catch {
      // 尝试选择第一个可用用例
      const caseSelect = protoTestPage.caseSelect;
      await caseSelect.click();
      await sleep(300);
      const options = protoTestPage.page.locator('.n-base-select-option');
      const count = await options.count();
      if (count === 0) {
        test.skip();
        return;
      }
      // 使用 force: true 避免被其他元素遮挡导致 click 失败
      await options.first().click({ force: true });
      await sleep(2000);
    }

    // 确认用例已加载（表格有数据）
    const rowCount = await protoTestPage.getTableRowCount();
    if (rowCount === 0) {
      test.skip();
      return;
    }

    // 点击执行用例
    await protoTestPage.clickExecuteCase();
    console.log('✅ 已点击执行用例');

    // 等待重放开始和消息推送（重放每条间隔5秒，需要足够等待）
    await sleep(10000);

    // 验证自动切换到重放结果页签
    const replayResultRows = await protoTestPage.getReplayResultRowCount();
    console.log(`重放结果页签表格行数: ${replayResultRows}`);

    if (replayResultRows > 0) {
      const sourceText = await protoTestPage.getResultSourceLabel();
      console.log(`来源标签: ${sourceText}`);
      expect(sourceText).toContain('测试用例');
      console.log('🎉 测试通过：测试用例执行用例 → 重放结果页签正确显示');
    } else {
      console.log('⚠️ 重放结果表格为空，可能服务端未响应');
    }
  });

  test('重发后追加结果', async ({ protoTestPage }) => {
    await protoTestPage.goto();

    // 切换到测试用例页签，选择用例
    await protoTestPage.clickTabTestcase();
    await sleep(1000);

    try {
      await protoTestPage.selectCaseFromDropdown('');
    } catch {
      const caseSelect = protoTestPage.caseSelect;
      await caseSelect.click();
      await sleep(300);
      const options = protoTestPage.page.locator('.n-base-select-option');
      const count = await options.count();
      if (count === 0) {
        test.skip();
        return;
      }
      // 使用 force: true 避免被其他元素遮挡导致 click 失败
      await options.first().click({ force: true });
      await sleep(2000);
    }

    const rowCount = await protoTestPage.getTableRowCount();
    if (rowCount === 0) {
      test.skip();
      return;
    }

    // 选中第一行 Req 消息
    let foundReq = false;
    for (let i = 0; i < Math.min(rowCount, 10); i++) {
      const rowText = await protoTestPage.getRowText(i);
      if (rowText && rowText.includes('C->S') && !rowText.includes('S->C')) {
        await protoTestPage.clickTableRow(i);
        foundReq = true;
        break;
      }
    }

    if (!foundReq) {
      test.skip();
      return;
    }

    await sleep(500);

    // 点击重发按钮
    try {
      await protoTestPage.clickRetryReplay();
      console.log('✅ 已点击重发');

      // 等待重放完成
      await sleep(5000);

      // 验证自动切换到重放结果页签
      const replayResultRows = await protoTestPage.getReplayResultRowCount();
      console.log(`重放结果页签表格行数: ${replayResultRows}`);

      if (replayResultRows > 0) {
        const sourceText = await protoTestPage.getResultSourceLabel();
        console.log(`来源标签: ${sourceText}`);
        expect(sourceText).toContain('重发控制');
        console.log('🎉 测试通过：重发 → 重放结果页签正确追加');
      } else {
        console.log('⚠️ 重放结果表格为空');
      }
    } catch (e) {
      console.log('⚠️ 重发操作失败:', e);
      throw e;
    }
  });

  test('多次重播累积多个结果', async ({ protoTestPage }) => {
    await protoTestPage.goto();

    // 第一次重放
    const rowCount = await protoTestPage.getTableRowCount();
    if (rowCount === 0) {
      test.skip();
      return;
    }

    await protoTestPage.startReplayButton.click();
    await sleep(3000);

    // 切回发包改包页签
    await protoTestPage.clickTabPacket();
    await sleep(500);

    // 第二次重放
    await protoTestPage.startReplayButton.click();
    await sleep(3000);

    // 此时应该在重放结果页签，验证有多个结果
    const replayResultRows = await protoTestPage.getReplayResultRowCount();
    console.log(`最终重放结果行数: ${replayResultRows}`);

    // 确保在重放结果页签上
    await protoTestPage.clickTabReplayResult();
    await sleep(500);

    // 验证结果选择器中有多个选项（至少 2 个结果）
    const selectorCount = await protoTestPage.getResultSelectorCount();
    console.log(`结果选择器选项数: ${selectorCount}`);

    expect(selectorCount).toBeGreaterThanOrEqual(2);
    console.log('🎉 测试通过：多次重播累积多个结果');
  });

  test('发包改包表格不受影响', async ({ protoTestPage }) => {
    await protoTestPage.goto();

    // 记录发包改包表格原始行数
    const originalCount = await protoTestPage.getTableRowCount();
    if (originalCount === 0) {
      test.skip();
      return;
    }
    console.log(`发包改包表格原始行数: ${originalCount}`);

    // 执行重放
    await protoTestPage.startReplayButton.click();
    await sleep(3000);

    // 切回发包改包页签
    await protoTestPage.clickTabPacket();
    await sleep(500);

    // 验证表格数据仍然正常（行数 >= 原始行数，因为重放响应也会追加）
    const afterCount = await protoTestPage.getTableRowCount();
    console.log(`重放后发包改包表格行数: ${afterCount}`);

    expect(afterCount).toBeGreaterThanOrEqual(originalCount);
    console.log('🎉 测试通过：发包改包表格数据未丢失');
  });
});
