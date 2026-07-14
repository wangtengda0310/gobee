/**
 * 首页 Page Object
 * 对应 src/pages/home/index.vue
 */

import { Page, Locator, expect } from '@playwright/test';
import { BasePage, Route, resolveRoute } from './BasePage';
import { sleep } from '../utils/helpers';

/**
 * 首页 Page Object
 * AI 聊天助手页面
 */
export class HomePage extends BasePage {
  // 页面元素定位器
  readonly pageContainer: Locator;
  readonly chatContainer: Locator;
  readonly messagesWrapper: Locator;
  readonly welcomeMessage: Locator;
  readonly welcomeIcon: Locator;
  readonly welcomeTitle: Locator;
  readonly welcomeText: Locator;
  readonly inputArea: Locator;
  readonly messageInput: Locator;
  readonly sendButton: Locator;
  readonly stopButton: Locator;
  readonly clearButton: Locator;
  readonly configButton: Locator;
  readonly configDrawer: Locator;

  constructor(page: Page) {
    super(page);

    // 主容器
    this.pageContainer = page.locator('#ChatHome');

    // 聊天容器
    this.chatContainer = page.locator('.chat-container');
    this.messagesWrapper = page.locator('.messages-wrapper');

    // 欢迎消息
    this.welcomeMessage = page.locator('.welcome-message');
    this.welcomeIcon = page.locator('.welcome-icon');
    this.welcomeTitle = page.locator('.welcome-text h2');
    this.welcomeText = page.locator('.welcome-text p');

    // 输入区域
    this.inputArea = page.locator('.input-area');
    this.messageInput = page.locator('.input-area .n-input textarea');
    this.sendButton = page.locator('.input-area button[type="button"]').nth(1);
    this.stopButton = page.locator('.input-area button:has-text("停止")');
    this.clearButton = page.locator('.input-area button:has-text("清空")');
    this.configButton = page.locator('.input-area button').first();

    // 配置抽屉
    this.configDrawer = page.locator('.n-drawer');
  }

  /**
   * 导航到首页
   */
  async goto(): Promise<void> {
    await this.page.locator('#layout-header button:has-text("AI助手")').click();
    await sleep(800);
  }

  /**
   * 获取消息列表项
   */
  getMessageItems(): Locator {
    return this.page.locator('.chat-message-item');
  }

  /**
   * 获取最后一条消息
   */
  getLastMessage(): Locator {
    return this.getMessageItems().last();
  }

  /**
   * 输入消息
   */
  async inputMessage(text: string): Promise<void> {
    await this.messageInput.fill(text);
    await sleep(100);
  }

  /**
   * 发送消息（点击按钮）
   */
  async sendMessage(): Promise<void> {
    await this.sendButton.click();
    await sleep(300);
  }

  /**
   * 发送消息（Enter 键）
   */
  async sendMessageWithEnter(): Promise<void> {
    await this.messageInput.press('Enter');
    await sleep(300);
  }

  /**
   * 输入并发送消息
   */
  async inputAndSendMessage(text: string): Promise<void> {
    await this.inputMessage(text);
    await this.sendMessage();
  }

  /**
   * 停止生成
   */
  async stopGeneration(): Promise<void> {
    await this.stopButton.click();
    await sleep(200);
  }

  /**
   * 清空对话
   */
  async clearChat(): Promise<void> {
    await this.clearButton.click();
    await sleep(200);
  }

  /**
   * 打开配置面板
   */
  async openConfigPanel(): Promise<void> {
    await this.configButton.click();
    await sleep(300);
  }

  /**
   * 关闭配置面板
   */
  async closeConfigPanel(): Promise<void> {
    await this.closeDrawer();
  }

  /**
   * 验证欢迎消息可见
   */
  async expectWelcomeVisible(): Promise<void> {
    await this.expectVisible(this.welcomeMessage);
    await this.expectText(this.welcomeTitle, 'AI 助手');
  }

  /**
   * 验证消息数量
   */
  async expectMessageCount(count: number): Promise<void> {
    const items = this.getMessageItems();
    await expect(items).toHaveCount(count);
  }

  /**
   * 验证输入框为空
   */
  async expectInputEmpty(): Promise<void> {
    await this.expectEmpty(this.messageInput);
  }

  /**
   * 验证发送按钮状态
   */
  async expectSendButtonEnabled(enabled: boolean): Promise<void> {
    if (enabled) {
      await expect(this.sendButton).toBeEnabled();
    } else {
      await expect(this.sendButton).toBeDisabled();
    }
  }

  /**
   * 验证正在生成状态
   */
  async expectIsStreaming(): Promise<void> {
    await expect(this.stopButton).toBeVisible();
    await expect(this.sendButton).toBeHidden();
  }

  /**
   * 验证非生成状态
   */
  async expectNotStreaming(): Promise<void> {
    await expect(this.sendButton).toBeVisible();
    await expect(this.stopButton).toBeHidden();
  }

  /**
   * 等待 AI 响应完成
   */
  async waitForResponse(timeout = 60000): Promise<void> {
    // 等待生成开始
    try {
      await this.stopButton.waitFor({ state: 'visible', timeout: 5000 });
    } catch {
      // 可能响应很快，已经完成了
      return;
    }
    // 等待生成完成
    await this.stopButton.waitFor({ state: 'hidden', timeout });
  }

  /**
   * 验证消息包含文本
   */
  async expectMessageContains(index: number, text: string): Promise<void> {
    const message = this.getMessageItems().nth(index);
    await this.expectText(message, text);
  }

  /**
   * 获取消息内容
   */
  async getMessageContent(index: number): Promise<string> {
    const message = this.getMessageItems().nth(index);
    return this.getTextContent(message);
  }

  /**
   * Shift+Enter 换行
   */
  async pressShiftEnter(): Promise<void> {
    await this.messageInput.press('Shift+Enter');
    await sleep(100);
  }

  /**
   * 验证输入框多行内容
   */
  async expectMultilineInput(): Promise<void> {
    const value = await this.messageInput.inputValue();
    expect(value).toContain('\n');
  }

  /**
   * 获取配置面板中的提供商选择器
   */
  getProviderSelect(): Locator {
    return this.configDrawer.locator('.n-select');
  }

  /**
   * 获取配置面板中的 API Key 输入框
   */
  getApiKeyInput(): Locator {
    return this.configDrawer.locator('input[type="password"]');
  }

  /**
   * 选择 AI 提供商
   */
  async selectProvider(provider: string): Promise<void> {
    await this.getProviderSelect().click();
    await sleep(200);
    await this.selectDropdownOption(provider);
  }

  /**
   * 输入 API Key
   */
  async inputApiKey(apiKey: string): Promise<void> {
    await this.getApiKeyInput().fill(apiKey);
    await sleep(100);
  }

  /**
   * 获取模型选择器
   */
  getModelSelect(): Locator {
    return this.configDrawer.locator('.n-select').nth(1);
  }

  /**
   * 选择模型
   */
  async selectModel(model: string): Promise<void> {
    await this.getModelSelect().click();
    await sleep(200);
    await this.selectDropdownOption(model);
  }

  /**
   * 保存配置
   */
  async saveConfig(): Promise<void> {
    const saveButton = this.configDrawer.locator('button:has-text("保存")');
    await saveButton.click();
    await sleep(200);
  }

  /**
   * 验证页面已加载
   */
  async expectPageLoaded(): Promise<void> {
    await expect(this.pageContainer).toBeVisible();
    await expect(this.chatContainer).toBeVisible();
  }
}
