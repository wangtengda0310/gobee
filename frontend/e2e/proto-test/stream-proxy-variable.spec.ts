/**
 * 协议重放页 E2E 测试 — 动态变量提取功能
 *
 * 运行方式：
 *   wails3 dev
 *   cd frontend && npx playwright test stream-proxy/stream-proxy-variable.spec.ts
 *
 * 测试范围：field-item 中的 inputTypeOptions 始终包含变量选项
 *           variable-select 组件的变量下拉选择
 *           选择变量后 FieldFourState 包含 variable_name
 *           带字段配置的 Req 默认卡片模式
 *
 * 数据说明：复用「观战」用例，在测试用例页签内验证（选用例只加载到测试用例
 * 表格，发包改包页签表格数据源不同，不能跨页签操作）。
 */
import { test, expect, describe } from '../shared/fixtures';
import { ProtoTestPage } from '../shared/pages/ProtoTestPage';

describe('协议重放页 — 动态变量提取', () => {
  let page: ProtoTestPage;

  test.beforeEach(async ({ page: p }) => {
    page = new ProtoTestPage(p);
    await page.goto();
  });

  // 防御性清理：关闭可能从前次测试残留的下拉菜单，避免 selectCaseFromDropdown
  // 内部 waitFor('.n-base-select-menu') 因多元素触发 strict mode 失败
  async function dismissStaleDropdowns() {
    if (await page.page.locator('.n-base-select-menu:visible').count() > 0) {
      await page.page.keyboard.press('Escape');
      await page.page.waitForTimeout(300);
    }
  }

  // 加载「观战」用例并在测试用例页签选中指定 Req 行，返回该页面对象便于后续操作
  async function loadWatchCaseAndSelectReq(reqName: string) {
    await page.clickTabTestcase();
    await page.page.waitForTimeout(1000);
    await dismissStaleDropdowns();
    await page.selectCaseFromDropdown('观战');
    await page.page.waitForTimeout(2000);

    const row = page.page.locator('tbody tr:visible').filter({ hasText: reqName }).first();
    await expect(row).toBeVisible();
    await row.click();
    await page.page.waitForTimeout(800);
  }

  // NewGetRoomListReq 用于测试「original → variable」切换流程。选中后确保进入卡片模式。
  // 不假设默认模式：paired-payload-editor 的 reqEditMode 是组件状态，goto 只切内存路由
  // 不重建组件，跨测试可能保留；根据当前显示的切换按钮决定是否需要点击。
  async function loadNewGetRoomListReqInCardMode() {
    await loadWatchCaseAndSelectReq('NewGetRoomListReq');
    const cardModeBtn = page.page.locator('button:visible').filter({ hasText: /^卡片模式$/ }).first();
    // 卡片模式按钮可见 = 当前 JSON 模式，需点击切换
    if (await cardModeBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
      await cardModeBtn.click();
      await page.page.waitForTimeout(500);
    }
    // 确认卡片编辑器已就绪
    await expect(page.page.locator('text=Payload 字段').first()).toBeVisible();
  }

  // ==================== G1: 变量选项始终可用 ====================
  describe('变量选项始终可用 (G1)', () => {
    test('变量选项默认出现在 inputTypeOptions 中', async () => {
      await loadNewGetRoomListReqInCardMode();

      // 查找第一个字段项的类型下拉菜单
      const firstField = page.page.locator('.field-item').first();
      const dropdown = firstField.locator('.n-select');
      await expect(dropdown.first()).toBeVisible();
      await dropdown.first().click();
      await page.page.waitForTimeout(300);

      // 验证选项数量为 5（含"变量"）
      const options = page.page.locator('.n-base-select-option:visible');
      const optionCount = await options.count();
      expect(optionCount).toBe(5);

      // 验证包含所有预期选项
      const optionTexts = await options.allTextContents();
      expect(optionTexts).toContain('原始值');
      expect(optionTexts).toContain('范围');
      expect(optionTexts).toContain('枚举');
      expect(optionTexts).toContain('组合');
      expect(optionTexts).toContain('变量');

      // 关闭下拉菜单（Escape 会同时清除行选中，但本测试已结束，下次 beforeEach 会重新导航）
      await page.page.keyboard.press('Escape');
      await page.page.waitForTimeout(200);
    });
  });

  // ==================== G2: 选择 variable 后下拉显示可用变量 ====================
  describe('选择 variable 后下拉显示可用变量 (G2)', () => {
    test('选择 variable 后 variable-select 下拉显示可用变量', async () => {
      await loadNewGetRoomListReqInCardMode();

      const firstField = page.page.locator('.field-item').first();
      const dropdown = firstField.locator('.n-select');
      await expect(dropdown.first()).toBeVisible();

      // 切换到"变量"模式
      await dropdown.first().click();
      await page.page.waitForTimeout(300);
      const varOption = page.page.locator('.n-base-select-option:visible').filter({ hasText: '变量' });
      await varOption.click();
      await page.page.waitForTimeout(500);

      // 验证变量选择下拉出现
      const varSelect = firstField.locator('[data-testid="variable-select-dropdown"]');
      await expect(varSelect.first()).toBeVisible();
    });

    test('variable-select 选项格式为 {DisplayName} ({ShortName})', async () => {
      await loadNewGetRoomListReqInCardMode();

      const firstField = page.page.locator('.field-item').first();
      const dropdown = firstField.locator('.n-select');

      // 切换到"变量"模式
      await dropdown.first().click();
      await page.page.waitForTimeout(300);
      const varOption = page.page.locator('.n-base-select-option:visible').filter({ hasText: '变量' });
      await varOption.click();
      await page.page.waitForTimeout(500);

      // 打开变量下拉
      const varSelect = firstField.locator('[data-testid="variable-select-dropdown"]');
      await varSelect.click();
      await page.page.waitForTimeout(500);

      // 验证下拉中有选项
      const varOptions = page.page.locator('.n-base-select-option:visible');
      const varOptionCount = await varOptions.count();
      expect(varOptionCount).toBeGreaterThan(0);

      // 验证格式包含已知 DisplayName（城池ID/房间列表/当前账号）
      const texts = await varOptions.allTextContents();
      const hasDisplayName = texts.some(t => t.includes('城池') || t.includes('房间') || t.includes('账号'));
      expect(hasDisplayName).toBeTruthy();

      // 关闭下拉（Escape 会清除行选中，但本测试已结束）
      await page.page.keyboard.press('Escape');
      await page.page.waitForTimeout(200);
    });
  });

  // ==================== G3: 选择变量后 FieldFourState 包含 variable_name ====================
  describe('选择变量后 FieldFourState 包含 variable_name (G3)', () => {
    test('选择变量后应能收集到 variable_name', async () => {
      await loadNewGetRoomListReqInCardMode();

      const firstField = page.page.locator('.field-item').first();
      const dropdown = firstField.locator('.n-select');

      // 切换到"变量"模式
      await dropdown.first().click();
      await page.page.waitForTimeout(300);
      const varOption = page.page.locator('.n-base-select-option:visible').filter({ hasText: '变量' });
      await varOption.click();
      await page.page.waitForTimeout(500);

      // 打开变量下拉并选择第一个
      const varSelect = firstField.locator('[data-testid="variable-select-dropdown"]');
      await varSelect.click();
      await page.page.waitForTimeout(500);

      const varOptions = page.page.locator('.n-base-select-option:visible');
      const varOptionCount = await varOptions.count();
      expect(varOptionCount).toBeGreaterThan(0);

      // 选择第一个变量
      await varOptions.first().click();
      await page.page.waitForTimeout(300);

      // 验证选中了某个值（variable-select 会更新，不再是占位"选择变量"）
      const selectedText = await varSelect.textContent();
      expect(selectedText).toBeTruthy();
      expect(selectedText).not.toBe('选择变量');
    });
  });

  // ==================== G4: 带字段配置的 Req 默认卡片模式 ====================
  describe('带字段配置的 Req 默认卡片模式 (G4)', () => {
    // 复用「观战」用例：RoomLookOnReq 含 variable 字段配置，NewGetRoomListReq 无字段配置。
    // 在测试用例页签内验证（选用例只加载到测试用例表格，发包改包页签表格数据源不同）。
    async function loadWatchCaseRows() {
      await page.clickTabTestcase();
      await page.page.waitForTimeout(1000);
      await dismissStaleDropdowns();
      await page.selectCaseFromDropdown('观战');
      await page.page.waitForTimeout(2000);
    }

    test('选中含 variable 字段配置的 Req 应默认进入卡片模式', async () => {
      await loadWatchCaseRows();

      const targetRow = page.page.locator('tbody tr:visible').filter({ hasText: 'RoomLookOnReq' }).first();
      await expect(targetRow).toBeVisible();
      await targetRow.click();
      await page.page.waitForTimeout(800);

      // 卡片模式标志：按钮文案为「JSON模式」+「Payload 字段」标题可见
      await expect(page.page.locator('button:visible').filter({ hasText: /^JSON模式$/ }).first()).toBeVisible();
      await expect(page.page.locator('text=Payload 字段').first()).toBeVisible();
    });

    test('选中无字段配置的 Req 应默认进入 JSON 模式', async () => {
      await loadWatchCaseRows();

      const targetRow = page.page.locator('tbody tr:visible').filter({ hasText: 'NewGetRoomListReq' }).first();
      await expect(targetRow).toBeVisible();
      await targetRow.click();
      await page.page.waitForTimeout(800);

      // JSON 模式标志：按钮文案为「卡片模式」
      await expect(page.page.locator('button:visible').filter({ hasText: /^卡片模式$/ }).first()).toBeVisible();
    });
  });

  // ==================== G5: 变量按 Req 过滤（AvailableReqs） ====================
  describe('变量按 Req 过滤 (G5)', () => {
    test('RoomLookOnReq 的变量下拉仅显示 roomCreator/roomID/openid', async () => {
      await loadWatchCaseAndSelectReq('RoomLookOnReq');

      // RoomLookOnReq 字段已是 variable，variable-select 已显示，直接打开下拉
      const varSelect = page.page.locator('[data-testid="variable-select-dropdown"]:visible').first();
      await varSelect.click();
      await page.page.waitForTimeout(500);

      const joined = (await page.page.locator('.n-base-select-option:visible').allTextContents()).join('|');
      expect(joined).toContain('房间列表');  // roomCreator / roomID
      expect(joined).toContain('当前账号');   // openid（全可用）
      expect(joined).not.toContain('城池');   // cityId 仅对 TeamSelectGuildCityReq 可见

      await page.page.keyboard.press('Escape');
    });

    test('NewGetRoomListReq 的变量下拉仅显示 openid（全可用变量）', async () => {
      await loadNewGetRoomListReqInCardMode();

      // 切换第一字段为 variable
      const firstField = page.page.locator('.field-item').first();
      await firstField.locator('.n-select').first().click();
      await page.page.waitForTimeout(300);
      await page.page.locator('.n-base-select-option:visible').filter({ hasText: '变量' }).click();
      await page.page.waitForTimeout(500);

      // 打开 variable-select 下拉
      const varSelect = firstField.locator('[data-testid="variable-select-dropdown"]');
      await varSelect.click();
      await page.page.waitForTimeout(500);

      const joined = (await page.page.locator('.n-base-select-option:visible').allTextContents()).join('|');
      expect(joined).toContain('当前账号');    // openid 全可用
      expect(joined).not.toContain('房间列表'); // roomCreator/roomID 仅对 RoomLookOnReq
      expect(joined).not.toContain('城池');     // cityId 仅对 TeamSelectGuildCityReq

      await page.page.keyboard.press('Escape');
    });
  });
});
