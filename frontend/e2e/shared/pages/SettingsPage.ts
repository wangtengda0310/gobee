/**
 * 设置页 Page Object
 * 对应 src/pages/settings/index.vue
 */

import { Page, Locator, expect } from '@playwright/test';
import { BasePage, Route, resolveRoute } from './BasePage';
import { sleep } from '../utils/helpers';

/**
 * 设置页 Page Object
 * 全局配置管理页面
 */
export class SettingsPage extends BasePage {
  // 飞书通知配置
  readonly feishuCard: Locator;
  readonly feishuNotifySwitch: Locator;
  readonly feishuGuidInput: Locator;
  readonly interceptSwitch: Locator;
  readonly viewInterceptButton: Locator;

  // MCP 服务配置
  readonly mcpCard: Locator;
  readonly mcpEnabledSwitch: Locator;
  readonly mcpRunningTag: Locator;
  readonly mcpHostInput: Locator;
  readonly mcpPortInput: Locator;

  // 日志管理
  readonly logCard: Locator;
  readonly viewLogButton: Locator;

  // 服务端日志
  readonly serverLogCard: Locator;
  readonly serverLogCount: Locator;
  readonly viewServerLogButton: Locator;

  // 弹窗和抽屉
  readonly interceptModal: Locator;
  readonly logPanel: Locator;
  readonly serverLogDrawer: Locator;

  constructor(page: Page) {
    super(page);

    // 飞书配置
    this.feishuCard = page.locator('.n-card:has-text("飞书通知配置")');
    this.feishuNotifySwitch = page.locator('.n-card:has-text("飞书通知") .n-switch').first();
    this.feishuGuidInput = page.locator('.n-card:has-text("机器人GUID") input');
    this.interceptSwitch = page.locator('.n-card:has-text("消息劫持") .n-switch');
    this.viewInterceptButton = page.locator('button:has-text("查看消息")');

    // MCP 配置
    this.mcpCard = page.locator('.n-card:has-text("MCP 服务")');
    this.mcpEnabledSwitch = page.locator('.n-card:has-text("MCP") .n-switch');
    this.mcpRunningTag = page.locator('.n-tag:has-text("运行中")');
    this.mcpHostInput = page.locator('.n-card:has-text("绑定地址") input');
    this.mcpPortInput = page.locator('.n-card:has-text("端口") input');

    // 日志管理
    this.logCard = page.locator('.n-card:has-text("日志管理")');
    this.viewLogButton = page.locator('button:has-text("查看日志")');

    // 服务端日志
    this.serverLogCard = page.locator('.n-card:has-text("服务端日志")');
    this.serverLogCount = page.locator('.n-card:has-text("已捕获日志")');
    this.viewServerLogButton = page.locator('button:has-text("查看服务端日志")');

    // 弹窗
    this.interceptModal = page.locator('.intercept-modal');
    this.logPanel = page.locator('.log-panel');
    this.serverLogDrawer = page.locator('.server-log-drawer');
  }

  async goto(): Promise<void> {
    await this.page.locator('#layout-header button:has-text("设置")').click();
    await sleep(800);
  }

  // 飞书通知
  async toggleFeishuNotify(): Promise<void> {
    await this.feishuNotifySwitch.click();
    await sleep(200);
  }

  async setFeishuGuid(guid: string): Promise<void> {
    await this.feishuGuidInput.clear();
    await this.feishuGuidInput.fill(guid);
    await this.feishuGuidInput.blur();
    await sleep(200);
  }

  async getFeishuGuidValue(): Promise<string> {
    return await this.feishuGuidInput.inputValue();
  }

  async toggleIntercept(): Promise<void> {
    await this.interceptSwitch.click();
    await sleep(200);
  }

  async clickViewInterceptMessages(): Promise<void> {
    await this.viewInterceptButton.click();
    await sleep(300);
  }

  async expectInterceptModalVisible(): Promise<void> {
    await expect(this.interceptModal).toBeVisible();
  }

  // MCP 服务
  async toggleMCP(): Promise<void> {
    await this.mcpEnabledSwitch.click();
    await sleep(500);
  }

  async setMCPHost(host: string): Promise<void> {
    await this.mcpHostInput.clear();
    await this.mcpHostInput.fill(host);
    await this.mcpHostInput.blur();
    await sleep(200);
  }

  async setMCPPort(port: string): Promise<void> {
    await this.mcpPortInput.clear();
    await this.mcpPortInput.fill(port);
    await this.mcpPortInput.blur();
    await sleep(200);
  }

  // 日志
  async clickViewLog(): Promise<void> {
    await this.viewLogButton.click();
    await sleep(300);
  }

  async expectLogPanelVisible(): Promise<void> {
    await expect(this.logPanel).toBeVisible();
  }

  async getServerLogCount(): Promise<number> {
    const text = await this.serverLogCount.textContent() || '0 条';
    const match = text.match(/(\d+)/);
    return match ? parseInt(match[1]) : 0;
  }

  async clickViewServerLog(): Promise<void> {
    await this.viewServerLogButton.click();
    await sleep(300);
  }

  async expectServerLogDrawerVisible(): Promise<void> {
    await expect(this.serverLogDrawer).toBeVisible();
  }

  // 验证
  async expectFeishuConfigVisible(): Promise<void> {
    await expect(this.feishuCard).toBeVisible();
  }

  async expectMCPConfigVisible(): Promise<void> {
    await expect(this.mcpCard).toBeVisible();
  }

  async expectFeishuNotifyToggled(): Promise<void> {
    const checked = await this.feishuNotifySwitch.getAttribute('aria-checked');
    expect(checked).toBe('true');
  }

  async expectInterceptToggled(): Promise<void> {
    const checked = await this.interceptSwitch.getAttribute('aria-checked');
    expect(checked).toBe('true');
  }
}
