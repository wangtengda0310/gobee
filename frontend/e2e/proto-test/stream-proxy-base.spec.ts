/**
 * 协议重放页 E2E 测试 — 拆分自 stream-proxy.spec.ts
 *
 * 运行方式：
 *   wails3 dev
 *   cd frontend && npx playwright test stream-proxy/<filename>
 */
import { test, expect, describe } from '../shared/fixtures';
import { ProtoTestPage } from '../shared/pages/ProtoTestPage';
describe('协议重放页 — 基础', () => {
  let page: ProtoTestPage;

  test.beforeEach(async ({ page: p }) => {
    page = new ProtoTestPage(p);
    await page.goto();
  });

  describe('页面结构与页签', () => {
    test('发包改包页签可见', async () => {
      await expect(page.tabPacket).toBeVisible();
    });

    test('测试用例页签可见', async () => {
      await expect(page.tabTestcase).toBeVisible();
    });

    test('默认页签包含目标服务和录制按钮', async () => {
      await expect(page.replayServerInput).toBeVisible();
      await expect(page.startRecordButton).toBeVisible();
      await expect(page.stopRecordButton).toBeVisible();
    });

    test('切换到测试用例页签显示操作按钮', async () => {
      await page.clickTabTestcase();
      await expect(page.loadCaseButton).toBeVisible();
      await expect(page.executeCaseButton).toBeVisible();
      await expect(page.newCaseButton).toBeVisible();
      await expect(page.deleteCaseButton).toBeVisible();
    });
  });

  // ==================== 目标服务输入 ====================

  describe('目标服务输入', () => {
    test('输入 TCP 地址', async () => {
      await page.setReplayServerAddr('10.0.0.1:18000');
      expect(await page.replayServerInput.inputValue()).toBe('10.0.0.1:18000');
    });

    test('输入 HTTP 地址', async () => {
      await page.setReplayHttpAddr('10.0.0.1:20144');
      expect(await page.replayHttpInput.inputValue()).toBe('10.0.0.1:20144');
    });

    test('输入登录账号', async () => {
      await page.setReplayOpenID('test_admin');
      expect(await page.replayOpenIDInput.inputValue()).toBe('test_admin');
    });

    test('输入值在页签切换后保持', async () => {
      await page.setReplayServerAddr('10.254.114.204:18000');
      await page.setReplayHttpAddr('10.254.114.204:20144');
      await page.setReplayOpenID('test2');

      await page.clickTabTestcase();
      await page.clickTabPacket();

      expect(await page.replayServerInput.inputValue()).toBe('10.254.114.204:18000');
      expect(await page.replayHttpInput.inputValue()).toBe('10.254.114.204:20144');
      expect(await page.replayOpenIDInput.inputValue()).toBe('test2');
    });
  });

  // ==================== 录制按钮 ====================

  describe('录制按钮', () => {
    test('开始录制按钮可见且可用', async () => {
      await expect(page.startRecordButton).toBeVisible();
      await expect(page.startRecordButton).toBeEnabled();
    });

    test('停止录制按钮初始禁用', async () => {
      await expect(page.stopRecordButton).toBeVisible();
      await expect(page.stopRecordButton).toBeDisabled();
    });
  });

  // ==================== 多选模式 ====================

  // ==================== 并发度参数 ====================

  describe('最大并发参数', () => {
    test('最大并发输入框可见', async () => {
      await page.openSettings();
      await expect(page.maxConcurrencyInput).toBeVisible();
    });

    test('最大并发输入框默认值为 0', async () => {
      await page.openSettings();
      expect(await page.maxConcurrencyInput.inputValue()).toBe('0');
    });
  });
});
