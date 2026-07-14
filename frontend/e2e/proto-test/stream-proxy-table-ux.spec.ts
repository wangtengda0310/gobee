/**
 * 协议重放页 E2E 测试 — 拆分自 stream-proxy.spec.ts
 *
 * 运行方式：
 *   wails3 dev
 *   cd frontend && npx playwright test stream-proxy/<filename>
 */
import { test, expect, describe } from '../shared/fixtures';
import { ProtoTestPage } from '../shared/pages/ProtoTestPage';
describe('协议重放页 — 表格 UX', () => {
  let page: ProtoTestPage;

  test.beforeEach(async ({ page: p }) => {
    page = new ProtoTestPage(p);
    await page.goto();
  });

  // ==================== 测试用例描述编辑 ====================

  describe('测试用例描述编辑', () => {
    // 串行执行：共用「创号」用例文件，避免并行修改/选择冲突
    test.describe.configure({ mode: 'serial' });

    /** 加载「创号」用例并选中含描述的行（依赖 cases/proto_cases/创号.json） */
    async function loadChuanghaoCaseWithDescriptRow(page: ProtoTestPage, p: import('@playwright/test').Page) {
      await page.clickTabTestcase();
      await page.caseSelect.click();
      await p.waitForTimeout(500);
      const option = p.locator('.n-base-select-menu:visible .n-base-select-option').filter({ hasText: '创号' });
      if ((await option.count()) === 0) {
        return false;
      }
      await option.first().click();
      await p.waitForTimeout(2000);

      const descriptRow = p.locator('.n-data-table:visible tbody tr').filter({ hasText: '添加粮草' });
      if ((await descriptRow.count()) === 0) {
        return false;
      }
      await descriptRow.first().click();
      await p.waitForTimeout(300);
      return true;
    }

    test('选中行后显示描述输入框并加载当前描述', async ({ page: p }) => {
      const pageObj = new ProtoTestPage(p);
      await pageObj.goto();
      if (!(await loadChuanghaoCaseWithDescriptRow(pageObj, p))) {
        test.skip(true, '无「创号」用例或缺少带描述的测试数据');
        return;
      }

      const input = p.locator('[data-testid="case-descript-input"]:visible');
      await expect(input).toBeVisible();
      await expect(input).toHaveValue('添加粮草');
    });

    test('修改描述失焦不会自动保存，应用按钮变为可用', async ({ page: p }) => {
      const pageObj = new ProtoTestPage(p);
      await pageObj.goto();
      if (!(await loadChuanghaoCaseWithDescriptRow(pageObj, p))) {
        test.skip(true, '无「创号」用例或缺少带描述的测试数据');
        return;
      }

      const input = pageObj.caseDescriptInput;
      await input.fill(`e2e描述_${Date.now()}`);
      await input.blur();
      await p.waitForTimeout(500);

      // 失焦不应自动保存
      await expect(p.locator('.n-message').filter({ hasText: '描述已保存' })).toHaveCount(0);
      await expect(p.locator('.n-message').filter({ hasText: '已保存' })).toHaveCount(0);
      expect(await pageObj.isApplyButtonEnabled()).toBe(true);
    });

    test('仅修改描述点应用后表格同步且按钮再次禁用', async ({ page: p }) => {
      const pageObj = new ProtoTestPage(p);
      await pageObj.goto();
      if (!(await loadChuanghaoCaseWithDescriptRow(pageObj, p))) {
        test.skip(true, '无「创号」用例或缺少带描述的测试数据');
        return;
      }

      const newDescript = `e2e描述_${Date.now()}`;
      await pageObj.caseDescriptInput.fill(newDescript);
      await pageObj.clickApplyButton();
      await pageObj.expectStepSavedToast();
      await expect(
        p.locator('[data-testid="descript-cell"]:visible').filter({ hasText: newDescript })
      ).toBeVisible();
      expect(await pageObj.isApplyButtonEnabled()).toBe(false);

      // 还原描述
      await pageObj.caseDescriptInput.fill('添加粮草');
      await pageObj.clickApplyButton();
      await pageObj.expectStepSavedToast();
    });

    test('同时修改 payload 与描述点应用一次保存', async ({ page: p }) => {
      const pageObj = new ProtoTestPage(p);
      await pageObj.goto();
      if (!(await loadChuanghaoCaseWithDescriptRow(pageObj, p))) {
        test.skip(true, '无「创号」用例或缺少带描述的测试数据');
        return;
      }

      const newDescript = `e2e联合_${Date.now()}`;
      await pageObj.caseDescriptInput.fill(newDescript);

      const textarea = p.locator('textarea').first();
      const marker = `"E2E_MARKER_${Date.now()}"`;
      await textarea.evaluate((el: HTMLTextAreaElement, m: string) => {
        const setter = Object.getOwnPropertyDescriptor(window.HTMLTextAreaElement.prototype, 'value')?.set;
        const parsed = JSON.parse(el.value) as Record<string, unknown>;
        parsed.e2e_marker = m;
        setter?.call(el, JSON.stringify(parsed, null, 2));
        el.dispatchEvent(new Event('input', { bubbles: true }));
      }, marker);
      await p.waitForTimeout(300);
      expect(await pageObj.isApplyButtonEnabled()).toBe(true);

      await pageObj.clickApplyButton();
      await pageObj.expectStepSavedToast();
      await expect(
        p.locator('[data-testid="descript-cell"]:visible').filter({ hasText: newDescript })
      ).toBeVisible();
      const json = await pageObj.getJsonEditorValue();
      expect(json).toContain('e2e_marker');

      // 还原
      await pageObj.caseDescriptInput.fill('添加粮草');
      await textarea.evaluate((el: HTMLTextAreaElement) => {
        const setter = Object.getOwnPropertyDescriptor(window.HTMLTextAreaElement.prototype, 'value')?.set;
        const parsed = JSON.parse(el.value) as Record<string, unknown>;
        delete parsed.e2e_marker;
        setter?.call(el, JSON.stringify(parsed, null, 2));
        el.dispatchEvent(new Event('input', { bubbles: true }));
      });
      await p.waitForTimeout(300);
      await pageObj.clickApplyButton();
      await pageObj.expectStepSavedToast();
    });

    test('双击描述列聚焦底部描述输入框', async ({ page: p }) => {
      const pageObj = new ProtoTestPage(p);
      await pageObj.goto();
      if (!(await loadChuanghaoCaseWithDescriptRow(pageObj, p))) {
        test.skip(true, '无「创号」用例或缺少带描述的测试数据');
        return;
      }

      const cell = p.locator('[data-testid="descript-cell"]:visible').filter({ hasText: '添加粮草' }).first();
      await cell.dblclick();
      await p.waitForTimeout(200);

      const input = p.locator('[data-testid="case-descript-input"]:visible');
      await expect(input).toBeFocused();
    });
  });

  describe('步骤顺序与拖放', () => {
    test.describe.configure({ mode: 'serial' });

    const PACKET_TAB_SAMPLE = {
      version: 1,
      recorded_at: '2026-06-08T10:00:00.000Z',
      server_addr: '10.0.0.1:18000',
      message_count: 4,
      messages: [
        { index: 0, offset_ms: 0, msg_id: 1001, msg_name: 'HelloReq', seq_id: 1, direction: '→', payload_json: '{"greeting":"hello"}', field_values: null },
        { index: 1, offset_ms: 50, msg_id: 1002, msg_name: 'HelloAck', seq_id: 1, direction: '←', payload_json: '{"result":0}', field_values: null },
        { index: 2, offset_ms: 100, msg_id: 2001, msg_name: 'LoginReq', seq_id: 2, direction: '→', payload_json: '{"username":"test1"}', field_values: null },
        { index: 3, offset_ms: 150, msg_id: 2002, msg_name: 'LoginAck', seq_id: 2, direction: '←', payload_json: '{"result":0}', field_values: null },
      ],
    };

    async function injectPacketTabData(pageObj: ProtoTestPage, data: typeof PACKET_TAB_SAMPLE) {
      await pageObj.page.evaluate((recordData) => {
        const recordBtn = Array.from(document.querySelectorAll('button'))
          .find(b => b.textContent?.includes('开始录制'));
        if (!recordBtn) throw new Error('未找到发包改包页签');
        let el: HTMLElement | null = recordBtn as HTMLElement;
        while (el) {
          const comp = (el as any).__vueParentComponent;
          if (comp?.exposed?.setRecordData) {
            comp.exposed.setRecordData(recordData);
            return;
          }
          el = el.parentElement;
        }
        throw new Error('未找到 packet-tab 组件');
      }, data);
    }

    /** 加载「创号」用例（不选中特定行） */
    async function loadChuanghaoCase(pageObj: ProtoTestPage, p: import('@playwright/test').Page) {
      await pageObj.clickTabTestcase();
      await pageObj.caseSelect.click();
      await p.waitForTimeout(500);
      const option = p.locator('.n-base-select-menu:visible .n-base-select-option').filter({ hasText: '创号' });
      if ((await option.count()) === 0) return false;
      await option.first().click();
      await p.waitForTimeout(2000);
      return (await pageObj.getTableRowCount()) >= 3;
    }

    test('发包改包页签不显示拖动手柄', async ({ page: p }) => {
      const pageObj = new ProtoTestPage(p);
      await pageObj.goto();
      await pageObj.clickTabPacket();
      await injectPacketTabData(pageObj, PACKET_TAB_SAMPLE);
      await p.waitForTimeout(500);
      expect(await pageObj.getDragHandleCount()).toBe(0);
    });

    test('重放结果页签不显示拖动手柄', async ({ page: p }) => {
      const pageObj = new ProtoTestPage(p);
      await pageObj.goto();
      await pageObj.clickTabReplayResult();
      await p.waitForTimeout(500);
      expect(await pageObj.getDragHandleCount()).toBe(0);
    });

    test('测试用例页签拖拽后显示顺序变更栏', async ({ page: p }) => {
      const pageObj = new ProtoTestPage(p);
      await pageObj.goto();
      if (!(await loadChuanghaoCase(pageObj, p))) {
        test.skip(true, '无「创号」用例');
        return;
      }

      expect(await pageObj.getDragHandleCount()).toBeGreaterThan(0);
      await pageObj.dragTableRow(0, 2);
      await expect(pageObj.orderDirtyBar).toBeVisible();
      await expect(pageObj.orderDirtyBar).toContainText('步骤顺序已变更');
    });

    test('还原按钮恢复步骤顺序', async ({ page: p }) => {
      const pageObj = new ProtoTestPage(p);
      await pageObj.goto();
      if (!(await loadChuanghaoCase(pageObj, p))) {
        test.skip(true, '无「创号」用例');
        return;
      }

      const originalFirstRow = await pageObj.getRowText(0);
      await pageObj.dragTableRow(0, 2);
      await expect(pageObj.orderDirtyBar).toBeVisible();
      await pageObj.clickRevertOrder();
      await expect(pageObj.orderDirtyBar).toHaveCount(0);
      expect(await pageObj.getRowText(0)).toContain('CreateRoleReq');
    });

    test('保存顺序后重新加载仍保持新顺序', async ({ page: p }) => {
      const pageObj = new ProtoTestPage(p);
      await pageObj.goto();
      if (!(await loadChuanghaoCase(pageObj, p))) {
        test.skip(true, '无「创号」用例');
        return;
      }

      const originalFirst = await pageObj.getRowText(0);
      await pageObj.dragTableRow(0, 1);
      await expect(pageObj.orderDirtyBar).toBeVisible();
      await pageObj.clickSaveOrder();
      await pageObj.expectOrderSavedToast();

      // 重新选择用例触发从磁盘加载
      await pageObj.caseSelect.click();
      await p.waitForTimeout(300);
      await p.locator('.n-base-select-menu:visible .n-base-select-option').filter({ hasText: '创号' }).first().click();
      await p.waitForTimeout(2000);

      const reloadedFirst = await pageObj.getRowText(0);
      expect(reloadedFirst).not.toContain('CreateRoleReq');

      // 还原文件：将 CreateRoleReq 拖回第一行并保存
      const rows = pageObj.getTableRows();
      const count = await rows.count();
      for (let i = 0; i < count; i++) {
        const text = (await rows.nth(i).textContent()) || '';
        if (text.includes('CreateRoleReq')) {
          if (i !== 0) {
            await pageObj.dragTableRow(i, 0);
            await pageObj.clickSaveOrder();
            await pageObj.expectOrderSavedToast();
          }
          break;
        }
      }
    });

    test('顺序未保存时多选按钮禁用', async ({ page: p }) => {
      const pageObj = new ProtoTestPage(p);
      await pageObj.goto();
      if (!(await loadChuanghaoCase(pageObj, p))) {
        test.skip(true, '无「创号」用例');
        return;
      }

      await pageObj.dragTableRow(0, 2);
      await expect(pageObj.orderDirtyBar).toBeVisible();
      await expect(p.locator('[data-testid="testcase-multi-select-btn"]:visible')).toBeDisabled();
      await pageObj.clickRevertOrder();
      await expect(p.locator('[data-testid="testcase-multi-select-btn"]:visible')).toBeEnabled();
    });
  });
});
