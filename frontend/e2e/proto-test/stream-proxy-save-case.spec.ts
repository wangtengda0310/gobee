/**
 * 配对消息索引错位回归测试
 *
 * Bug 描述：发包改包页签中，用户右键选择一条 Req 消息"增加到用例"，
 * 实际保存的是错误的 Ack 消息（配对行 ID 与原始消息索引错位）。
 *
 * 根因：message-table.vue 的 addToCase 事件传递 PairedEntry.id（配对行序号），
 * 但 packet-tab.vue 的 handleAddToCase 用它去过滤原始消息数组（messages[idx]），
 * 导致索引错位选择到错误的消息。
 *
 * 数据映射关系（4条原始消息 → 2个配对行）：
 *   原始索引 0: HelloReq  (→) ──┐
 *   原始索引 1: HelloAck  (←) ──┤── Pair id=0: HelloReq | HelloAck
 *   原始索引 2: LoginReq  (→) ──┐
 *   原始索引 3: LoginAck  (←) ──┘── Pair id=1: LoginReq | LoginAck
 *
 * Bug 场景：右键 Pair id=1 → rowId=1 → messages[1] = HelloAck（错位！应该是 LoginReq+LoginAck）
 * Bug 场景：多选 [0,1] → messages[0,1] = [HelloReq, HelloAck]（错位！应该是全部4条）
 *
 * 运行方式（需先启动 Wails 桌面应用）：
 *   wails3 dev
 *   cd frontend && npx playwright test stream-proxy-save-case.spec.ts
 */
import { test, expect, describe } from '../shared/fixtures';
import { ProtoTestPage } from '../shared/pages/ProtoTestPage';
import { sleep } from '../shared/utils/helpers';

/** 4条原始消息的测试数据 */
const TEST_MESSAGES = [
  { index: 0, offset_ms: 0, msg_id: 1001, msg_name: 'HelloReq', seq_id: 1, direction: '→', payload_json: '{"greeting":"hello"}', field_values: null },
  { index: 1, offset_ms: 50, msg_id: 1002, msg_name: 'HelloAck', seq_id: 1, direction: '←', payload_json: '{"result":0}', field_values: null },
  { index: 2, offset_ms: 100, msg_id: 2001, msg_name: 'LoginReq', seq_id: 2, direction: '→', payload_json: '{"username":"test1"}', field_values: null },
  { index: 3, offset_ms: 150, msg_id: 2002, msg_name: 'LoginAck', seq_id: 2, direction: '←', payload_json: '{"result":0,"token":"abc"}', field_values: null },
];

const TEST_RECORD_DATA_LOGIN = {
  version: 1,
  recorded_at: '2026-06-08T11:00:00.000Z',
  server_addr: '10.0.0.2:18000',
  message_count: 2,
  messages: [
    { index: 0, offset_ms: 0, msg_id: 2001, msg_name: 'LoginReq', seq_id: 2, direction: '→', payload_json: '{"username":"test1"}', field_values: null },
    { index: 1, offset_ms: 50, msg_id: 2002, msg_name: 'LoginAck', seq_id: 2, direction: '←', payload_json: '{"result":0,"token":"abc"}', field_values: null },
  ],
};

/**
 * 通过 CDP 调用后端 SaveTestCase 预创建用例（用于追加场景的前置条件）
 */
async function createCaseByBackend(pageObj: ProtoTestPage, name: string, data: any): Promise<void> {
  await pageObj.page.evaluate(async ({ caseName, caseData }) => {
    try {
      const { SaveTestCase } = await import(
        /* @vite-ignore */
        '/bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/testcaseservice.js'
      );
      await SaveTestCase(caseName, caseData);
    } catch {
      // 备选：部分旧构建可能使用 streamproxy 路径
      const { SaveTestCase } = await import(
        /* @vite-ignore */
        '/bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/testcaseservice.js'
      );
      await SaveTestCase(caseName, caseData);
    }
  }, { caseName: name, caseData: data });
}

/**
 * 通过 CDP 读取用例文件中的消息名称列表
 */
async function getCaseMessageNames(pageObj: ProtoTestPage, name: string): Promise<string[] | null> {
  return await pageObj.page.evaluate(async (caseName) => {
    try {
      const { LoadTestCase } = await import(
        /* @vite-ignore */
        '/bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/testcaseservice.js'
      );
      const data = await LoadTestCase(caseName);
      if (!data || !data.messages) return [];
      return data.messages.map((m: any) => m.msg_name);
    } catch {
      try {
        const { LoadTestCase } = await import(
          /* @vite-ignore */
          '/bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/testcaseservice.js'
        );
        const data = await LoadTestCase(caseName);
        if (!data || !data.messages) return [];
        return data.messages.map((m: any) => m.msg_name);
      } catch {
        return null;
      }
    }
  }, name);
}

/**
 * 在保存到用例对话框中输入用例名称并点击"追加"
 */
async function fillCaseNameAndAppend(pageObj: ProtoTestPage, name: string): Promise<void> {
  const selectInput = pageObj.page.locator('.n-modal .n-base-selection-input').first();
  if (await selectInput.count() > 0) {
    await selectInput.click();
    await sleep(200);
    await pageObj.page.keyboard.type(name);
    await sleep(200);
    await pageObj.page.keyboard.press('Enter');
    await sleep(300);
  }
  const appendBtn = pageObj.page.locator('.n-modal button').filter({ hasText: '追加' });
  await appendBtn.click();
  await sleep(1500);
}

describe('追加到已存在用例', () => {
  test('多选 Req 追加不应覆盖原数据', async () => {
    const caseName = `e2e_append_${Date.now()}`;
    createdCaseNames.push(caseName);

    // 前置：通过后端创建一个只含 HelloReq 的用例
    await createCaseByBackend(page, caseName, {
      version: 1,
      recorded_at: '2026-06-08T09:00:00.000Z',
      server_addr: '10.0.0.0:18000',
      message_count: 1,
      messages: [
        { index: 0, offset_ms: 0, msg_id: 1001, msg_name: 'HelloReq', seq_id: 1, direction: '→', payload_json: '{"greeting":"hello"}', field_values: null },
      ],
    });

    // 注入第二批数据：LoginReq/LoginAck
    await injectPacketTabData(page, TEST_RECORD_DATA_LOGIN);
    await sleep(500);

    // 进入多选模式并勾选 LoginReq
    await page.clickMultiSelect();
    await sleep(500);
    const checkboxes = page.page.locator('.n-data-table:visible .n-checkbox');
    expect(await checkboxes.count()).toBeGreaterThanOrEqual(1);
    await checkboxes.nth(0).click();
    await sleep(200);

    // 点击"保存到用例"
    const saveBtn = page.page.locator('button').filter({ hasText: '保存到用例' });
    await saveBtn.click();
    await sleep(800);

    // 追加到同一个用例
    await fillCaseNameAndAppend(page, caseName);

    // 验证：用例中应同时包含 HelloReq（原有）和 LoginReq（追加）
    const savedMessages = await getCaseMessageNames(page, caseName);
    expect(savedMessages).toContain('HelloReq');
    expect(savedMessages).toContain('LoginReq');
    expect(savedMessages).not.toContain('LoginAck');
    expect(savedMessages).toHaveLength(2);
  });

  test('右键"增加到用例"追加到已存在用例不应覆盖原数据', async () => {
    const caseName = `e2e_append_rightclick_${Date.now()}`;
    createdCaseNames.push(caseName);

    // 前置：通过后端创建一个只含 HelloReq 的用例
    await createCaseByBackend(page, caseName, {
      version: 1,
      recorded_at: '2026-06-08T09:00:00.000Z',
      server_addr: '10.0.0.0:18000',
      message_count: 1,
      messages: [
        { index: 0, offset_ms: 0, msg_id: 1001, msg_name: 'HelloReq', seq_id: 1, direction: '→', payload_json: '{"greeting":"hello"}', field_values: null },
      ],
    });

    // 注入第二批数据：LoginReq/LoginAck（1 个配对行）
    await injectPacketTabData(page, TEST_RECORD_DATA_LOGIN);
    await sleep(500);

    // 右键点击第一行（LoginReq|LoginAck）
    const rows = page.page.locator('.n-data-table:visible tbody tr');
    expect(await rows.count()).toBe(1);
    await rows.nth(0).click({ button: 'right' });
    await sleep(300);

    // 点击"增加到用例"
    const addToCaseOption = page.page.locator('.n-dropdown-menu .n-dropdown-option')
      .filter({ hasText: /^增加到用例/ });
    await addToCaseOption.click();
    await sleep(800);

    // 追加到同一个用例
    await fillCaseNameAndAppend(page, caseName);

    // 验证：用例中应同时包含 HelloReq（原有）和 LoginReq（追加），且不包含 Ack
    const savedMessages = await getCaseMessageNames(page, caseName);
    expect(savedMessages).toContain('HelloReq');
    expect(savedMessages).toContain('LoginReq');
    expect(savedMessages).not.toContain('LoginAck');
    expect(savedMessages).toHaveLength(2);
  });

  test('右键无 Req 行时不显示"增加到用例"菜单', async () => {
    // 注入纯 Ntf 数据
    const ntfOnlyData = {
      version: 1,
      recorded_at: '2026-06-08T12:00:00.000Z',
      server_addr: '10.0.0.3:18000',
      message_count: 1,
      messages: [
        { index: 0, offset_ms: 0, msg_id: 3001, msg_name: 'SomeNtf', seq_id: 5, direction: '←', payload_json: '{}', field_values: null },
      ],
    };
    await injectPacketTabData(page, ntfOnlyData);
    await sleep(500);

    // 右键点击 Ntf 行
    const rows = page.page.locator('.n-data-table:visible tbody tr');
    expect(await rows.count()).toBe(1);
    await rows.nth(0).click({ button: 'right' });
    await sleep(300);

    // 验证：自定义右键菜单不应出现（无"增加到用例"选项）
    const addToCaseOption = page.page.locator('.n-dropdown-menu .n-dropdown-option')
      .filter({ hasText: /^增加到用例/ });
    await expect(addToCaseOption).toHaveCount(0);

    // 关闭可能存在的菜单（按 Escape）
    await page.page.keyboard.press('Escape');
    await sleep(300);
  });

  test('多选模式下非 Req 行（Ntf）的 checkbox 应禁用', async () => {
    // 注入纯 Ntf 数据
    const ntfOnlyData = {
      version: 1,
      recorded_at: '2026-06-08T12:00:00.000Z',
      server_addr: '10.0.0.3:18000',
      message_count: 1,
      messages: [
        { index: 0, offset_ms: 0, msg_id: 3001, msg_name: 'SomeNtf', seq_id: 5, direction: '←', payload_json: '{}', field_values: null },
      ],
    };
    await injectPacketTabData(page, ntfOnlyData);
    await sleep(500);

    // 进入多选模式
    await page.clickMultiSelect();
    await sleep(500);

    // 验证：Ntf 行的 checkbox 应被禁用
    const checkboxes = page.page.locator('.n-data-table:visible .n-checkbox');
    expect(await checkboxes.count()).toBe(1);
    const checkbox = checkboxes.nth(0).locator('input');
    await expect(checkbox).toBeDisabled();

    // 退出多选模式
    await page.page.keyboard.press('Escape');
    await sleep(300);
  });
});


/**
 * 通过 CDP 找到 packet-tab 组件实例并调用 setRecordData
 *
 * Vue 3 在组件根元素的 DOM 节点上存储 __vueParentComponent 引用，
 * 通过向上遍历 DOM 树找到 expose 了 setRecordData 的组件。
 */
async function injectPacketTabData(pageObj: ProtoTestPage, data: any): Promise<void> {
  await pageObj.page.evaluate((recordData) => {
    // 策略：找到"开始录制"按钮所在的组件容器，它就是 packet-tab 的根元素
    // packet-tab 的特征：包含"开始录制"按钮
    const recordBtn = Array.from(document.querySelectorAll('button'))
      .find(b => b.textContent?.includes('开始录制'));
    if (!recordBtn) throw new Error('未找到"开始录制"按钮（不在发包改包页签？）');

    // 向上遍历 DOM 找到 packet-tab 组件实例
    let el: HTMLElement | null = recordBtn as HTMLElement;
    while (el) {
      const comp = (el as any).__vueParentComponent;
      if (comp?.exposed?.setRecordData) {
        comp.exposed.setRecordData(recordData);
        return true;
      }
      el = el.parentElement;
    }
    throw new Error('未找到 packet-tab 组件实例（setRecordData 未暴露）');
  }, data);
}

/**
 * 读取 packet-tab 组件内部状态
 */
async function getPacketTabState(pageObj: ProtoTestPage, keys: string[]): Promise<any> {
  return await pageObj.page.evaluate((stateKeys) => {
    const recordBtn = Array.from(document.querySelectorAll('button'))
      .find(b => b.textContent?.includes('开始录制'));
    if (!recordBtn) return null;

    let el: HTMLElement | null = recordBtn as HTMLElement;
    while (el) {
      const comp = (el as any).__vueParentComponent;
      if (comp?.exposed?.setRecordData) {
        const result: Record<string, any> = {};
        for (const key of stateKeys) {
          // Vue 3 defineExpose 的值通过 .exposed 访问
          // ref 需要 .value 来获取实际值
          const val = comp.exposed[key];
          if (val && typeof val === 'object' && '__v_isRef' in val) {
            result[key] = val.value;
          } else {
            result[key] = val;
          }
        }
        return result;
      }
      el = el.parentElement;
    }
    return null;
  }, keys);
}

describe('配对消息索引错位回归测试', () => {
  let page: ProtoTestPage;
  let createdCaseNames: string[] = [];

  test.beforeEach(async ({ page: p }) => {
    page = new ProtoTestPage(p);
    await page.goto();
    createdCaseNames = [];
  });

  test.afterEach(async () => {
    // 清理测试创建的用例文件
    for (const name of createdCaseNames) {
      try {
        await page.page.evaluate(async (caseName) => {
          // Wails v3 绑定通过 @wailsio/runtime 的 $Call.ByID 调用
          // 直接 import 绑定模块使用
          try {
            const { DeleteTestCase } = await import(
              /* @vite-ignore */
              '/bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/testcaseservice.ts'
            );
            await DeleteTestCase(caseName);
          } catch (e) {
            // 备选：尝试 .js 后缀
            try {
              const { DeleteTestCase } = await import(
                /* @vite-ignore */
                '/bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/testcaseservice.js'
              );
              await DeleteTestCase(caseName);
            } catch {}
          }
        }, name);
      } catch {
        // 忽略清理失败
      }
    }

    // 确保退出多选模式（v-show 导致按钮存在但不可见，用 Escape 处理）
    try {
      const cancelBtn = page.page.locator('button').filter({ hasText: '取消多选' }).first();
      if (await cancelBtn.isVisible()) {
        await cancelBtn.click();
        await sleep(300);
      }
    } catch {
      // 忽略
    }

    // 确保模态框关闭
    const modal = page.page.locator('.n-modal');
    if (await modal.count() > 0) {
      await page.page.keyboard.press('Escape');
      await sleep(300);
    }
  });

  // ==================== 数据注入验证 ====================

  test('注入测试数据后表格应显示 2 行配对消息', async () => {
    await injectPacketTabData(page, TEST_RECORD_DATA);
    await sleep(500);

    // 4 条原始消息 → 2 个 Req/Ack 配对行
    // 注意：发包改包页签使用 v-show，DOM 中可能有多个 .n-data-table
    // 只读取可见的表格（发包改包页签的）
    const rowCount = await page.page.evaluate(() => {
      const tables = document.querySelectorAll('.n-data-table');
      for (const table of tables) {
        // 找到可见的表格
        if ((table as HTMLElement).offsetParent !== null) {
          return table.querySelectorAll('tbody tr').length;
        }
      }
      return 0;
    });
    expect(rowCount).toBe(2);

    // 验证消息数据正确注入
    const state = await getPacketTabState(page, ['messages']);
    expect(state).not.toBeNull();
    expect(state.messages).toHaveLength(4);
    expect(state.messages[0].msg_name).toBe('HelloReq');
    expect(state.messages[2].msg_name).toBe('LoginReq');
  });

  // ==================== 右键菜单"增加到用例"测试 ====================

  describe('右键菜单"增加到用例"索引正确性', () => {
    test('右键第二行（LoginReq|LoginAck）应保存 Login 相关消息而非 HelloAck', async () => {
      await injectPacketTabData(page, TEST_RECORD_DATA);
      await sleep(500);

      // 右键第二行（Pair id=1，对应 LoginReq|LoginAck）
      const rows = page.page.locator('.n-data-table:visible tbody tr');
      const rowCount = await rows.count();
      expect(rowCount).toBe(2);

      await rows.nth(1).click({ button: 'right' });
      await sleep(300);

      // 点击右键菜单"增加到用例"
      const addToCaseOption = page.page.locator('.n-dropdown-menu .n-dropdown-option')
        .filter({ hasText: '增加到用例' });
      await addToCaseOption.click();
      await sleep(800);

      // 保存对话框应已弹出，输入用例名并保存
      const caseName = `e2e_idx_rightclick_${Date.now()}`;
      createdCaseNames.push(caseName);

      // n-select 支持 tag 模式创建新值：输入名称后按回车
      const selectInput = page.page.locator('.n-modal .n-base-selection-input').first();
      if (await selectInput.count() > 0) {
        await selectInput.click();
        await sleep(200);
        await page.page.keyboard.type(caseName);
        await sleep(200);
        await page.page.keyboard.press('Enter');
        await sleep(300);
      }

      // 点击"追加"按钮
      const appendBtn = page.page.locator('.n-modal button').filter({ hasText: '追加' });
      await appendBtn.click();
      await sleep(1500);

      // 验证：直接通过后端读取保存的用例文件内容，检查消息名称
      // 这比通过 UI 加载更可靠（避免页签状态干扰）
      const savedMessages = await page.page.evaluate(async (caseName) => {
        try {
          const { LoadTestCase } = await import(
            /* @vite-ignore */
            '/bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/testcaseservice.ts'
          );
          const data = await LoadTestCase(caseName);
          if (!data || !data.messages) return [];
          return data.messages.map((m: any) => m.msg_name);
        } catch {
          // 备选路径
          try {
            const { LoadTestCase } = await import(
              /* @vite-ignore */
              '/bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/testcaseservice.js'
            );
            const data = await LoadTestCase(caseName);
            if (!data || !data.messages) return [];
            return data.messages.map((m: any) => m.msg_name);
          } catch {
            return null;
          }
        }
      }, caseName);

      // 核心断言：保存的消息应包含 LoginReq 和 LoginAck
      expect(savedMessages).toContain('LoginReq');
      expect(savedMessages).toContain('LoginAck');

      // 不应包含 Hello（用户选的是第二行 Login 配对行）
      expect(savedMessages).not.toContain('HelloReq');
      expect(savedMessages).not.toContain('HelloAck');
    });
  });

  // ==================== 多选"保存到用例"测试 ====================

  describe('多选模式"保存到用例"索引正确性', () => {
    test('多选两行保存应包含全部 4 条消息（Hello+Login）', async () => {
      await injectPacketTabData(page, TEST_RECORD_DATA);
      await sleep(500);

      // 进入多选模式
      await page.clickMultiSelect();
      await sleep(500);

      // 勾选两行 checkbox（只操作可见表格中的）
      const checkboxes = page.page.locator('.n-data-table:visible .n-checkbox');
      const cbCount = await checkboxes.count();
      expect(cbCount).toBeGreaterThanOrEqual(2);

      await checkboxes.nth(0).click();
      await sleep(200);
      await checkboxes.nth(1).click();
      await sleep(200);

      // 点击"保存到用例"
      const saveBtn = page.page.locator('button').filter({ hasText: '保存到用例' });
      await saveBtn.click();
      await sleep(800);

      // 输入用例名并保存
      const caseName = `e2e_idx_multiselect_${Date.now()}`;
      createdCaseNames.push(caseName);

      const selectInput = page.page.locator('.n-modal .n-base-selection-input').first();
      if (await selectInput.count() > 0) {
        await selectInput.click();
        await sleep(200);
        await page.page.keyboard.type(caseName);
        await sleep(200);
        await page.page.keyboard.press('Enter');
        await sleep(300);
      }

      const appendBtn = page.page.locator('.n-modal button').filter({ hasText: '追加' });
      await appendBtn.click();
      await sleep(1500);

      // 验证：直接通过后端读取保存的用例文件内容
      const savedMessages = await page.page.evaluate(async (caseName) => {
        try {
          const { LoadTestCase } = await import(
            /* @vite-ignore */
            '/bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/testcaseservice.ts'
          );
          const data = await LoadTestCase(caseName);
          if (!data || !data.messages) return [];
          return data.messages.map((m: any) => m.msg_name);
        } catch {
          try {
            const { LoadTestCase } = await import(
              /* @vite-ignore */
              '/bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/testcaseservice.js'
            );
            const data = await LoadTestCase(caseName);
            if (!data || !data.messages) return [];
            return data.messages.map((m: any) => m.msg_name);
          } catch {
            return null;
          }
        }
      }, caseName);

      // 核心断言：全部 4 条消息都应存在
      expect(savedMessages).toContain('HelloReq');
      expect(savedMessages).toContain('HelloAck');
      expect(savedMessages).toContain('LoginReq');
      expect(savedMessages).toContain('LoginAck');
    });
  });
});
