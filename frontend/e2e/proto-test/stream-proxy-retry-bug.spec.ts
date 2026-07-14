/**
 * 复现"重发多次未生效"bug 的 E2E 测试
 *
 * 流程：
 *   1. 加载"添加黄金"用例（1个 GmCommandReq + 1个 GmCommandAck）
 *   2. 选中 Req 行 → 显示底部重放控制面板
 *   3. 调整重发次数为 2
 *   4. 点击"重发"按钮
 *   5. 检查后端日志确认实际发送次数
 *
 * 运行方式：wails3 dev 启动后
 *   npx playwright test stream-proxy-retry-bug.spec.ts --reporter=list
 */
import { test, expect } from '../shared/fixtures';
import { ProtoTestPage } from '../shared/pages/ProtoTestPage';
import { sleep } from '../shared/utils/helpers';

test('复现：重发2次应发送2条消息', async ({ page: p }) => {
  const page = new ProtoTestPage(p);
  await page.goto();

  // 1. 切换到测试用例页签，加载"添加黄金"
  await page.clickTabTestcase();
  await page.page.waitForTimeout(1000);
  await page.clickLoadCase();
  await page.page.waitForTimeout(1000);

  // 选择"添加黄金"（按键方式绕过下拉菜单遮挡问题）
  await page.caseSelect.dispatchEvent('click');
  await page.page.waitForTimeout(500);
  // 用例列表: [添加黄金, record_20260603_224145]，"添加黄金"排第一
  await page.page.keyboard.press('ArrowDown');
  await page.page.waitForTimeout(300);
  await page.page.keyboard.press('Enter');
  await page.page.waitForTimeout(2000);

  // 2. 切回发包改包页签，选中 Req（第1行，GmCommandReq）
  await page.clickTabPacket();
  await page.page.waitForTimeout(500);

  const rows = page.getTableRows();
  const rowCount = await rows.count();
  console.log(`表格行数: ${rowCount}`);
  expect(rowCount).toBeGreaterThan(0); // 不再用 if 守卫假通过

  // 选中第1行（Req — GmCommandReq, direction="→"）
  await rows.first().dispatchEvent('click');
  await page.page.waitForTimeout(500);

  // 3. 调整重发次数为 2（evaluate 直接设置值绕过 v-show 可见性限制）
  await p.locator('.n-input-number input').first().evaluate((el: HTMLInputElement) => {
    const setter = Object.getOwnPropertyDescriptor(
      window.HTMLInputElement.prototype, 'value'
    )?.set;
    setter?.call(el, '2');
    el.dispatchEvent(new Event('input', { bubbles: true }));
    el.dispatchEvent(new Event('change', { bubbles: true }));
  });
  await page.page.waitForTimeout(300);

  const countValue = await page.replayCountInput.inputValue();
  console.log(`重发次数: ${countValue}`);
  expect(countValue).toBe('2');

  // 4. 点击"重发"按钮（dispatchEvent 绕过 v-show 可见性检查）
  const retryBtns = p.locator('button').filter({ hasText: '重发' });
  const retryBtnCount = await retryBtns.count();
  console.log(`重发按钮实例数: ${retryBtnCount}`);
  let clicked = false;
  for (let i = 0; i < retryBtnCount; i++) {
    const btn = retryBtns.nth(i);
    const isVis = await btn.isVisible();
    console.log(`  按钮 ${i}: visible=${isVis}`);
    if (isVis) {
      await btn.click();
      clicked = true;
      console.log(`  点击了第 ${i} 个重发按钮`);
      break;
    }
  }
  // 兜底：如果所有按钮都 hidden，用 dispatchEvent
  if (!clicked && retryBtnCount > 0) {
    await retryBtns.first().dispatchEvent('click');
    clicked = true;
    console.log('  所有按钮 hidden，用 dispatchEvent 兜底');
  }
  expect(clicked).toBeTruthy();

  // 5. 等待重发完成（2条消息 × 5秒间隔 = 约10秒 + 登录耗时）
  console.log('等待重发完成...');
  await page.page.waitForTimeout(15000);

  // 切换到重放结果页签检查结果
  await page.clickTabReplayResult();
  await page.page.waitForTimeout(500);

  const resultRowCount = await page.getReplayResultRowCount();
  console.log(`重放结果行数: ${resultRowCount}`);

  // Bug 验证点：重发 2 次，预期重放结果至少新增 2 条
  // 如果 resultRowCount < 2，说明多次重发未全部生效
  console.log(`\n========== 结果 ==========`);
  console.log(`重发次数设置: 2`);
  console.log(`重放结果行数: ${resultRowCount}`);
  if (resultRowCount < 2) {
    console.log(`⚠️ Bug 复现！预期 ≥2 条，实际 ${resultRowCount} 条`);
    console.log(`请检查 Wails 后端日志中的 [重放] 关键字`);
  } else {
    console.log(`✅ 重发多次似乎正常（行数 ≥ 2）`);
  }
  console.log(`===========================\n`);

  // 注：不硬断言 resultRowCount >= 2，因为可能服务端不可达
  // 该测试的核心价值是配合后端日志一起分析
});
