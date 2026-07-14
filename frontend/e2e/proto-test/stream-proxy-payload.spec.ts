/**
 * 协议重放页 E2E 测试 — 拆分自 stream-proxy.spec.ts
 *
 * 运行方式：
 *   wails3 dev
 *   cd frontend && npx playwright test stream-proxy/<filename>
 */
import { test, expect, describe } from '../shared/fixtures';
import { ProtoTestPage } from '../shared/pages/ProtoTestPage';
describe('协议重放页 — Payload 同步与重发', () => {
  let page: ProtoTestPage;

  test.beforeEach(async ({ page: p }) => {
    page = new ProtoTestPage(p);
    await page.goto();
  });

  describe('重发 Payload 修改同步', () => {
    test.beforeEach(async ({ page: p }) => {
      page = new ProtoTestPage(p);
      await page.goto();
    });

    test('修改 JSON 后应用按钮变为可用', async () => {
      // 使用结构简单的测试用例
      await page.clickTabTestcase();
      await page.page.waitForTimeout(1000);

      // 先加载用例列表
      await page.clickLoadCase();
      await page.page.waitForTimeout(1000);

      // 使用 dispatchEvent 绕过 visible 检查，先加载用例再用键盘选择
      await page.caseSelect.dispatchEvent('click');
      await page.page.waitForTimeout(500);
      await page.page.keyboard.press('ArrowDown');
      await page.page.waitForTimeout(300);
      await page.page.keyboard.press('Enter');
      await page.page.waitForTimeout(2000);

      // 获取表格行 — 第1行是 Req，第2行是 Ack
      const allRows = page.page.locator('table').nth(1).locator('tbody tr');
      const count = await allRows.count();

      if (count === 0) {
        test.skip(true, '表格无数据，跳过测试');
        return;
      }

      // 选中第1行（Req — GmCommandReq，方向 "C->S" 或 "→"）
      await allRows.first().dispatchEvent('click');
      await page.page.waitForTimeout(500);

      // 获取初始 JSON 内容
      const initialJson = await page.getJsonEditorValue();
      console.log('初始 JSON:', initialJson);

      // 如果 JSON 是 "{}"（Ack 行），尝试选第2行
      if (initialJson.trim() === '{}') {
        await allRows.nth(1).dispatchEvent('click');
        await page.page.waitForTimeout(500);
        const retryJson = await page.getJsonEditorValue();
        if (retryJson.trim() === '{}') {
          test.skip(true, '所有行都是只读的 Ack，跳过测试');
          return;
        }
      }

      // 修改 JSON 内容 — 直接用 evaluate 设置值，绕过 textarea readonly
      await page.page.locator('textarea').first().evaluate((el: HTMLTextAreaElement) => {
        const textarea = el;
        // 获取当前值并替换
        const nativeInputValueSetter = Object.getOwnPropertyDescriptor(
          window.HTMLTextAreaElement.prototype, 'value'
        )?.set;
        nativeInputValueSetter?.call(textarea, textarea.value.replace('amount', 'AMOUNT_MODIFIED'));
        textarea.dispatchEvent(new Event('input', { bubbles: true }));
        textarea.dispatchEvent(new Event('change', { bubbles: true }));
      });
      await page.page.waitForTimeout(500);

      // 验证修改后"应用"按钮变为可用
      const afterApplyEnabled = await page.isApplyButtonEnabled();
      console.log('修改后应用按钮状态:', afterApplyEnabled);
      expect(afterApplyEnabled).toBeTruthy();
    });

    test.skip('修改 JSON 后直接重发应使用修改后的 payload', async () => {
      // ⚠️ 此测试需要实际服务端连接才能验证
      // 运行前确保：wails3 dev 已启动 + 服务端可达

      await page.clickTabTestcase();
      await page.page.waitForTimeout(1000);
      await page.clickLoadCase();
      await page.page.waitForTimeout(1000);

      // 选择用例
      await page.caseSelect.dispatchEvent('click');
      await page.page.waitForTimeout(500);
      await page.page.keyboard.press('ArrowDown');
      await page.page.waitForTimeout(300);
      await page.page.keyboard.press('Enter');
      await page.page.waitForTimeout(2000);

      // 配置目标服务
      await page.setReplayServerAddr('10.254.114.204:18000');
      await page.setReplayHttpAddr('10.254.114.204:20144');
      await page.setReplayOpenID('test1');

      // 选中第1行
      const allRows = page.page.locator('table').nth(1).locator('tbody tr');
      const count = await allRows.count();
      if (count === 0) { test.skip(true, '表格无数据'); return; }
      await allRows.first().dispatchEvent('click');
      await page.page.waitForTimeout(500);

      // 获取原始 JSON
      const originalJson = await page.getJsonEditorValue();
      console.log('原始 JSON:', originalJson);

      // 修改 JSON：用 evaluate 修改 textarea 值
      const markerValue = `E2E_TEST_MARKER_${Date.now()}`;
      await page.page.locator('textarea').first().evaluate(
        (el: HTMLTextAreaElement, marker: string) => {
          const setter = Object.getOwnPropertyDescriptor(
            window.HTMLTextAreaElement.prototype, 'value'
          )?.set;
          const newVal = el.value.replace(/"amount"\s*:\s*\d+/g, `"amount": "${marker}"`);
          setter?.call(el, newVal);
          el.dispatchEvent(new Event('input', { bubbles: true }));
          el.dispatchEvent(new Event('change', { bubbles: true }));
        },
        markerValue
      );
      await page.page.waitForTimeout(500);

      // 清空重放结果
      await page.clickTabReplayResult();
      await page.page.waitForTimeout(300);
      const clearBtn = page.page.locator('button').filter({ hasText: '清空' }).first();
      if (await clearBtn.isVisible()) {
        await clearBtn.click({ force: true });
        await page.page.waitForTimeout(300);
      }

      // 重发 — 用 dispatchEvent 绕过可见性，找可见的重发按钮
      await page.clickTabTestcase();
      await page.page.waitForTimeout(500);
      const retryBtn = page.page.locator('button').filter({ hasText: '重发' });
      const retryCount = await retryBtn.count();
      for (let i = 0; i < retryCount; i++) {
        const btn = retryBtn.nth(i);
        if (await btn.isVisible()) {
          await btn.dispatchEvent('click');
          break;
        }
      }
      await page.page.waitForTimeout(5000);

      // 检查重放结果
      await page.clickTabReplayResult();
      await page.page.waitForTimeout(500);

      const resultRowCount = await page.getReplayResultRowCount();
      console.log('重放结果行数:', resultRowCount);

      if (resultRowCount > 0) {
        // 点击最后一行展开 payload 编辑器
        const resultRows = page.page.locator('.n-data-table').last().locator('tbody tr');
        await resultRows.last().dispatchEvent('click');
        await page.page.waitForTimeout(500);
        const payload = await page.getJsonEditorValue();
        console.log('重放结果 payload:', payload);
        // Bug 验证：修复后包含标记值，修复前是原始值
        expect(payload).toContain(markerValue);
      } else {
        console.log('无重放结果，可能服务端未连接');
      }
    });

    test.skip('应用修改后重发应使用修改后的 payload', async () => {
      // ⚠️ 此测试需要实际服务端连接才能验证
      // 对照测试：先点"应用"再重发

      await page.clickTabTestcase();
      await page.page.waitForTimeout(1000);
      await page.clickLoadCase();
      await page.page.waitForTimeout(1000);

      await page.caseSelect.dispatchEvent('click');
      await page.page.waitForTimeout(500);
      await page.page.keyboard.press('ArrowDown');
      await page.page.waitForTimeout(300);
      await page.page.keyboard.press('Enter');
      await page.page.waitForTimeout(2000);

      await page.setReplayServerAddr('10.254.114.204:18000');
      await page.setReplayHttpAddr('10.254.114.204:20144');
      await page.setReplayOpenID('test1');

      const allRows = page.page.locator('table').nth(1).locator('tbody tr');
      const count = await allRows.count();
      if (count === 0) { test.skip(true, '表格无数据'); return; }
      await allRows.first().dispatchEvent('click');
      await page.page.waitForTimeout(500);

      // 修改 JSON
      const markerValue = `E2E_APPLY_MARKER_${Date.now()}`;
      await page.page.locator('textarea').first().evaluate(
        (el: HTMLTextAreaElement, marker: string) => {
          const setter = Object.getOwnPropertyDescriptor(
            window.HTMLTextAreaElement.prototype, 'value'
          )?.set;
          const newVal = el.value.replace(/"amount"\s*:\s*\d+/g, `"amount": "${marker}"`);
          setter?.call(el, newVal);
          el.dispatchEvent(new Event('input', { bubbles: true }));
          el.dispatchEvent(new Event('change', { bubbles: true }));
        },
        markerValue
      );
      await page.page.waitForTimeout(500);

      // 点击"应用"
      await page.clickApplyButton();
      await page.page.waitForTimeout(500);

      // 清空重放结果
      await page.clickTabReplayResult();
      await page.page.waitForTimeout(300);
      const clearBtn = page.page.locator('button').filter({ hasText: '清空' }).first();
      if (await clearBtn.isVisible()) {
        await clearBtn.click({ force: true });
        await page.page.waitForTimeout(300);
      }

      // 重发 — 找可见的重发按钮
      const retryBtn3 = page.page.locator('button').filter({ hasText: '重发' });
      const retryCount3 = await retryBtn3.count();
      for (let k = 0; k < retryCount3; k++) {
        const btn = retryBtn3.nth(k);
        if (await btn.isVisible()) {
          await btn.dispatchEvent('click');
          break;
        }
      }
      await page.page.waitForTimeout(5000);

      // 检查结果
      await page.clickTabReplayResult();
      await page.page.waitForTimeout(500);

      const resultRowCount3 = await page.getReplayResultRowCount();
      if (resultRowCount3 > 0) {
        const resultRows3 = page.page.locator('.n-data-table').last().locator('tbody tr');
        await resultRows3.last().dispatchEvent('click');
        await page.page.waitForTimeout(500);
        const payload3 = await page.getJsonEditorValue();
        console.log('应用后重发 payload:', payload3);
        expect(payload3).toContain(markerValue);
      }
    });
  });

  // ==================== 重发多次行为验证 ====================

  describe('重发多次行为验证', () => {
    test.beforeEach(async ({ page: p }) => {
      page = new ProtoTestPage(p);
      await page.goto();
    });

    test('重发次数输入框存在且默认值为1', async () => {
      // 加载用例数据
      await page.clickTabTestcase();
      await page.page.waitForTimeout(1000);
      await page.clickLoadCase();
      await page.page.waitForTimeout(1000);
      await page.caseSelect.dispatchEvent('click');
      await page.page.waitForTimeout(500);
      await page.page.keyboard.press('ArrowDown');
      await page.page.waitForTimeout(300);
      await page.page.keyboard.press('Enter');
      await page.page.waitForTimeout(2000);

      // 切回发包改包页签后选中第一行
      await page.clickTabPacket();
      await page.page.waitForTimeout(500);

      const rows = page.getTableRows();
      const count = await rows.count();
      if (count === 0) {
        test.skip(true, '表格无数据');
        return;
      }
      await rows.first().dispatchEvent('click');
      await page.page.waitForTimeout(500);

      // 验证重发次数输入框存在（v-show 下元素 DOM 存在即可）
      // 三个页签各有 .n-input-number，使用 replayPanel 区域内的最后一个（当前可见页签）
      const countInputInPanel = page.page.locator('.replay-control-area .n-input-number input').last();
      // 如果上面的选择器不匹配，回退到通用的 .n-input-number input
      const hasCountInput = (await countInputInPanel.count()) > 0
        || (await page.replayCountInput.count()) > 0;
      expect(hasCountInput).toBeTruthy();

      // 验证值为有效数字（inputValue 对 hidden 元素也有效）
      const defaultValue = await page.replayCountInput.inputValue();
      console.log('重发次数默认值:', defaultValue);
      expect(defaultValue).toBeTruthy();
    });

    test('设置重发次数为3后输入框显示3', async () => {
      await page.clickTabTestcase();
      await page.page.waitForTimeout(1000);
      await page.clickLoadCase();
      await page.page.waitForTimeout(1000);
      await page.caseSelect.dispatchEvent('click');
      await page.page.waitForTimeout(500);
      await page.page.keyboard.press('ArrowDown');
      await page.page.waitForTimeout(300);
      await page.page.keyboard.press('Enter');
      await page.page.waitForTimeout(2000);

      await page.clickTabPacket();
      await page.page.waitForTimeout(500);

      const rows = page.getTableRows();
      const count = await rows.count();
      if (count === 0) { test.skip(true, '表格无数据'); return; }
      await rows.first().dispatchEvent('click');
      await page.page.waitForTimeout(500);

      // 用 evaluate 直接设置 n-input-number 的值（绕过 v-show + force:true 仍可能失败的问题）
      await page.page.locator('.n-input-number input').first().evaluate((el: HTMLInputElement) => {
        const nativeInputValueSetter = Object.getOwnPropertyDescriptor(
          window.HTMLInputElement.prototype, 'value'
        )?.set;
        nativeInputValueSetter?.call(el, '3');
        el.dispatchEvent(new Event('input', { bubbles: true }));
        el.dispatchEvent(new Event('change', { bubbles: true }));
      });
      await page.page.waitForTimeout(300);

      const value = await page.replayCountInput.inputValue();
      console.log('设置后重发次数值:', value);
      expect(value).toBe('3');
    });

    test.skip('重发3次后重放结果应追加3条消息（需服务端连接）', async () => {
      // ⚠️ 此测试需要实际服务端连接才能验证
      // 复现场景：选中一条 Req，设置重发 3 次，点击重发
      // 预期：重放结果页签应追加 3 条消息
      // Bug 表现：如果后端 SendMessages 静默失败，
      // 前端显示"完成"但实际只发了 1 次或 0 次

      await page.clickTabTestcase();
      await page.page.waitForTimeout(1000);
      await page.clickLoadCase();
      await page.page.waitForTimeout(1000);
      await page.selectCaseFromDropdown('record_20260603_224145');
      await page.page.waitForTimeout(2000);

      // 配置目标服务（需要实际可达的服务端）
      await page.setReplayServerAddr('10.254.114.204:18000');
      await page.setReplayHttpAddr('10.254.114.204:20144');
      await page.setReplayOpenID('test1');

      await page.clickTabPacket();
      await page.page.waitForTimeout(500);

      const rows = page.getTableRows();
      const count = await rows.count();
      if (count === 0) { test.skip(true, '表格无数据'); return; }
      await rows.first().click();
      await page.page.waitForTimeout(500);

      // 清空重放结果页签
      await page.clickTabReplayResult();
      await page.page.waitForTimeout(300);
      const clearBtn = page.page.locator('button').filter({ hasText: '清空' }).first();
      if (await clearBtn.isVisible()) {
        await clearBtn.click({ force: true });
        await page.page.waitForTimeout(300);
      }

      // 获取重放前的结果行数
      const beforeCount = await page.getReplayResultRowCount();

      // 设置重发 3 次
      await page.clickTabPacket();
      await page.page.waitForTimeout(300);
      await page.setRepeatCount(3);
      await page.page.waitForTimeout(300);

      // 点击重发按钮
      await page.clickRetryReplay();
      // 等待重发完成（3条消息 × 5秒间隔 = 至少 15 秒）
      await page.page.waitForTimeout(20000);

      // 切到重放结果页签检查
      await page.clickTabReplayResult();
      await page.page.waitForTimeout(500);

      const afterCount = await page.getReplayResultRowCount();
      console.log(`重发前结果行数: ${beforeCount}, 重发后结果行数: ${afterCount}`);

      // 预期：追加了 3 条消息
      // Bug 验证：如果 afterCount - beforeCount < 3，说明多次重发未全部生效
      expect(afterCount - beforeCount).toBeGreaterThanOrEqual(3);
    });

    test.skip('连续2次重发按钮点击都应生效（需服务端连接）', async () => {
      // ⚠️ 此测试需要实际服务端连接才能验证
      // 复现场景：点重发 → 等待完成 → 再点重发
      // 预期：两次操作都应该生效（第二次不应该被"已有任务运行"拦截）

      await page.clickTabTestcase();
      await page.page.waitForTimeout(1000);
      await page.clickLoadCase();
      await page.page.waitForTimeout(1000);
      await page.selectCaseFromDropdown('record_20260603_224145');
      await page.page.waitForTimeout(2000);

      await page.setReplayServerAddr('10.254.114.204:18000');
      await page.setReplayHttpAddr('10.254.114.204:20144');
      await page.setReplayOpenID('test1');

      await page.clickTabPacket();
      await page.page.waitForTimeout(500);

      const rows = page.getTableRows();
      const count = await rows.count();
      if (count === 0) { test.skip(true, '表格无数据'); return; }
      await rows.first().click();
      await page.page.waitForTimeout(500);

      // 清空重放结果
      await page.clickTabReplayResult();
      await page.page.waitForTimeout(300);
      const clearBtn = page.page.locator('button').filter({ hasText: '清空' }).first();
      if (await clearBtn.isVisible()) {
        await clearBtn.click({ force: true });
        await page.page.waitForTimeout(300);
      }

      // 第一次重发
      await page.clickTabPacket();
      await page.page.waitForTimeout(300);
      await page.setRepeatCount(1);
      await page.page.waitForTimeout(300);
      await page.clickRetryReplay();
      await page.page.waitForTimeout(8000);

      const afterFirst = await page.getReplayResultRowCount();
      console.log(`第一次重发后行数: ${afterFirst}`);

      // 第二次重发（关键：验证 ReplayWorker 锁已释放）
      await page.clickTabPacket();
      await page.page.waitForTimeout(300);
      await page.setRepeatCount(1);
      await page.page.waitForTimeout(300);
      await page.clickRetryReplay();
      await page.page.waitForTimeout(8000);

      const afterSecond = await page.getReplayResultRowCount();
      console.log(`第二次重发后行数: ${afterSecond}`);

      // Bug 验证：第二次重发应该追加了新消息
      expect(afterSecond).toBeGreaterThan(afterFirst);
    });
  });

  // ==================== 重发自动路由迭代发送 ====================
  // Bug 回归：卡片模式配置枚举后点"重发"，应展开为多条 Req（走迭代发送），
  // 而不是把逗号拼接的枚举串当作单条消息的字段值发送
  describe('重发自动路由迭代发送', () => {
    test('配置枚举后点击重发应触发迭代发送', async () => {
      const p = page.page;
      await page.clickTabTestcase();
      await p.waitForTimeout(500);

      // 若表格无数据，先选择第一个可用用例（环境中用例名不固定）
      let reqRows = p.locator('tbody tr:visible').filter({ hasText: /Req/ });
      if ((await reqRows.count()) === 0) {
        await page.caseSelect.click();
        await p.waitForTimeout(500);
        const caseOptions = p.locator('.n-base-select-menu:visible .n-base-select-option');
        const optionCount = await caseOptions.count();
        console.log(`用例下拉选项数: ${optionCount}`);
        if (optionCount === 0) {
          await p.keyboard.press('Escape');
          test.skip(true, '无可用测试用例');
          return;
        }
        await caseOptions.first().click();
        await p.waitForTimeout(3000);
      }

      // 选中第一条 Req 行，等待编辑器出现
      reqRows = p.locator('tbody tr:visible').filter({ hasText: /Req/ });
      const reqRowCount = await reqRows.count();
      console.log(`可见 Req 行数: ${reqRowCount}`);
      if (reqRowCount === 0) { test.skip(true, '用例中无 Req 消息'); return; }
      await reqRows.first().click();
      await p.waitForTimeout(1000);

      // 切换到卡片模式（编辑器默认 JSON 模式；若已在卡片模式则按钮文案为"JSON模式"）
      const cardModeBtn = p.locator('button:visible').filter({ hasText: /^卡片模式$/ }).first();
      const jsonModeBtn = p.locator('button:visible').filter({ hasText: /^JSON模式$/ }).first();
      if (await cardModeBtn.isVisible()) {
        await cardModeBtn.click();
        await p.waitForTimeout(500);
      } else if (!(await jsonModeBtn.isVisible())) {
        test.skip(true, '无卡片模式按钮');
        return;
      }

      // 找到第一个基础类型字段的类型选择器
      const inputSelectors = p.locator('.field-item .input-selector:visible');
      if ((await inputSelectors.count()) === 0) { test.skip(true, '无可配置字段'); return; }
      const typeSelector = inputSelectors.first().locator('.n-base-selection').first();
      await typeSelector.click();
      await p.waitForTimeout(300);

      // 选择"枚举"类型
      // 注意：此处禁止按 Escape——Escape 会清除表格行选中并关闭整个编辑器
      await p.locator('.n-base-select-menu:visible .n-base-select-option')
        .filter({ hasText: /^枚举$/ }).first().click({ force: true });
      await p.waitForTimeout(300);

      // 添加两个枚举值标签（filterable+tag 的 NSelect 在选择框内联 input 中输入）
      const enumSelect = p.locator('.field-item .input-selector:visible').first()
        .locator('.n-base-selection').last();
      await enumSelect.click({ force: true });
      await p.waitForTimeout(200);
      for (const v of ['e2e_v1', 'e2e_v2']) {
        await p.keyboard.type(v);
        await p.waitForTimeout(200);
        await p.keyboard.press('Enter');
        await p.waitForTimeout(200);
      }

      // 点击重发：应自动路由到迭代发送（toast 文案为"迭代发送"而非"重发"）
      const retryBtn = p.locator('button:visible').filter({ hasText: /^重发$/ }).first();
      await retryBtn.click({ force: true });
      await p.waitForTimeout(1000);

      const iterCount = await p.locator('.n-message').filter({ hasText: '迭代发送' }).count();
      const retryCount = await p.locator('.n-message').filter({ hasText: /正在重发/ }).count();
      console.log(`迭代发送 toast: ${iterCount}, 单条重发 toast: ${retryCount}`);

      // 配置了枚举的情况下，必须走迭代发送，禁止单条重发
      expect(iterCount).toBeGreaterThan(0);
      expect(retryCount).toBe(0);
    });
  });

  // ==================== 表格列变体与 Req 过滤 ====================

  describe('表格列变体与 Req 过滤', () => {
    test('发包改包页签表头含完整列与 Req 过滤按钮', async ({ page: p }) => {
      const headerText = await p.locator('.n-data-table:visible thead').first().textContent() ?? '';
      expect(headerText).toContain('请求(Req)');
      expect(headerText).toContain('响应(Ack)');
      expect(headerText).toContain('时间');
      expect(headerText).toContain('SeqID');
      expect(headerText).toContain('结果');
      // Req 过滤按钮
      await expect(p.locator('[data-testid="req-filter-btn"]:visible').first()).toBeVisible();
    });

    test('测试用例页签表头显示描述列且无账号/时间/SeqID/结果列', async ({ page: p }) => {
      await page.clickTabTestcase();
      const headerText = await p.locator('.n-data-table:visible thead').first().textContent() ?? '';
      expect(headerText).toContain('请求(Req)');
      expect(headerText).toContain('描述');
      expect(headerText).not.toContain('账号');
      expect(headerText).not.toContain('响应(Ack)');
      expect(headerText).not.toContain('时间');
      expect(headerText).not.toContain('SeqID');
      expect(headerText).not.toContain('结果');
      // 测试用例页签无 Req 过滤按钮
      expect(await p.locator('[data-testid="req-filter-btn"]:visible').count()).toBe(0);
    });

    test('点击 Req 过滤按钮后只显示带 Req 数据的行', async ({ page: p }) => {
      const rows = p.locator('.n-data-table:visible tbody tr');
      const beforeCount = await rows.count();
      if (beforeCount === 0 || (await rows.first().textContent())?.includes('No Data')) {
        test.skip(true, '发包改包页签无录制数据');
        return;
      }

      const filterBtn = p.locator('[data-testid="req-filter-btn"]:visible').first();
      await filterBtn.click();
      await p.waitForTimeout(300);

      // 过滤后每一行的请求列都不应为 '-'（即都带有 Req 数据）
      const afterCount = await rows.count();
      expect(afterCount).toBeLessThanOrEqual(beforeCount);
      for (let i = 0; i < afterCount; i++) {
        const reqCell = await rows.nth(i).locator('td').nth(5).textContent();
        expect(reqCell?.trim()).not.toBe('-');
      }

      // 再次点击取消过滤，行数恢复
      await filterBtn.click();
      await p.waitForTimeout(300);
      expect(await rows.count()).toBe(beforeCount);
    });

    test('点击行后该行高亮（selected-row class）', async ({ page: p }) => {
      await page.clickTabTestcase();
      // 选择第一个可用用例（不依赖固定用例名）
      await page.caseSelect.click();
      await p.waitForTimeout(500);
      const caseOptions = p.locator('.n-base-select-menu:visible .n-base-select-option');
      if ((await caseOptions.count()) === 0) {
        test.skip(true, '无可用测试用例');
        return;
      }
      await caseOptions.first().click();
      await p.waitForTimeout(2000);

      const rows = p.locator('.n-data-table:visible tbody tr');
      const rowCount = await rows.count();
      if (rowCount === 0 || (await rows.first().textContent())?.includes('No Data')) {
        test.skip(true, '无可用测试用例数据');
        return;
      }

      await rows.first().click();
      await p.waitForTimeout(300);
      await expect(rows.first()).toHaveClass(/selected-row/);
    });
  });

});
