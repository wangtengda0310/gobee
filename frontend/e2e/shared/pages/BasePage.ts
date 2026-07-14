/**
 * Page Object 基类
 * 提供通用的页面操作方法
 */

import { Page, Locator, expect } from '@playwright/test';
import { sleep, waitForVisible, clickAndWait } from '../utils/helpers';

/**
 * 页面路由枚举
 */
export enum Route {
  HOME = '/Home',
  FUNCTION_TEST = '/Test',
  EXCEL_TEST = '/Excel',
  HERO_WIKI_CHECK = '/HeroWikiRes',
  HERO_VOICE_RESOURCE_CHECK = '/HeroRes',
  SETTINGS = '/Settings',
  ROADMAP = '/Roadmap',
}

/**
 * CDP 模式下需要完整 URL（如 http://wails.localhost:9245/Excel），
 * 相对路径会导致 "Cannot navigate to invalid URL" 错误。
 * 从当前页面 URL 提取 origin 拼接路由。
 */
export function resolveRoute(page: Page, route: Route | string): string {
  const currentUrl = page.url();
  if (currentUrl === 'chrome-error://chromewebdata/') {
    // 页面未加载完成时使用 wails 默认 origin
    return `http://wails.localhost:9245${route}`;
  }
  try {
    const url = new URL(currentUrl);
    return `${url.origin}${route}`;
  } catch {
    return `http://wails.localhost:9245${route}`;
  }
}

/**
 * Page Object 基类
 */
export abstract class BasePage {
  readonly page: Page;

  constructor(page: Page) {
    this.page = page;
  }

  /**
   * 导航到当前页面
   */
  abstract goto(): Promise<void>;

  /**
   * 等待页面加载完成
   */
  async waitForPageLoad(): Promise<void> {
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * 获取导航按钮
   */
  getNavButton(label: string): Locator {
    // 导航按钮可能是 .n-menu-item（Naive UI Menu）或 button（自定义按钮组）
    return this.page.locator(`.n-menu-item:has-text("${label}"), #layout-header button:has-text("${label}")`);
  }

  /**
   * 点击导航按钮
   */
  async clickNavButton(label: string): Promise<void> {
    const button = this.getNavButton(label);
    await button.click();
    await sleep(300);
  }

  /**
   * 获取按钮
   */
  getButton(label: string | RegExp): Locator {
    return this.page.locator(`button:has-text("${label}")`);
  }

  /**
   * 点击按钮
   */
  async clickButton(label: string | RegExp): Promise<void> {
    await this.getButton(label).click();
    await sleep(200);
  }

  /**
   * 获取输入框
   */
  getInput(placeholder?: string): Locator {
    if (placeholder) {
      return this.page.locator(`input[placeholder="${placeholder}"]`);
    }
    return this.page.locator('input');
  }

  /**
   * 填写输入框
   */
  async fillInput(placeholder: string, value: string): Promise<void> {
    const input = this.getInput(placeholder);
    await input.clear();
    await input.fill(value);
    await sleep(100);
  }

  /**
   * 获取文本内容
   */
  async getTextContent(locator: Locator): Promise<string> {
    return (await locator.textContent()) || '';
  }

  /**
   * 验证元素可见
   */
  async expectVisible(locator: Locator): Promise<void> {
    await expect(locator).toBeVisible();
  }

  /**
   * 验证元素隐藏
   */
  async expectHidden(locator: Locator): Promise<void> {
    await expect(locator).toBeHidden();
  }

  /**
   * 验证元素包含文本
   */
  async expectText(locator: Locator, text: string | RegExp): Promise<void> {
    await expect(locator).toContainText(text);
  }

  /**
   * 验证元素值为空
   */
  async expectEmpty(locator: Locator): Promise<void> {
    await expect(locator).toBeEmpty();
  }

  /**
   * 获取 Tab 标签
   */
  getTab(label: string): Locator {
    return this.page.locator(`.n-tabs-tab:has-text("${label}")`);
  }

  /**
   * 点击 Tab
   */
  async clickTab(label: string): Promise<void> {
    await this.getTab(label).click();
    await sleep(200);
  }

  /**
   * 获取卡片组件
   */
  getCard(title?: string): Locator {
    if (title) {
      return this.page.locator(`.n-card:has-text("${title}")`);
    }
    return this.page.locator('.n-card');
  }

  /**
   * 获取模态框
   */
  getModal(title?: string): Locator {
    if (title) {
      return this.page.locator(`.n-card:has-text("${title}")`);
    }
    return this.page.locator('.n-modal');
  }

  /**
   * 关闭模态框
   */
  async closeModal(): Promise<void> {
    const closeBtn = this.page.locator('.n-modal .n-base-close');
    if (await closeBtn.isVisible()) {
      await closeBtn.click();
      await sleep(200);
    }
  }

  /**
   * 获取抽屉组件
   */
  getDrawer(): Locator {
    return this.page.locator('.n-drawer');
  }

  /**
   * 关闭抽屉
   */
  async closeDrawer(): Promise<void> {
    const closeBtn = this.page.locator('.n-drawer .n-base-close');
    if (await closeBtn.isVisible()) {
      await closeBtn.click();
      await sleep(200);
    }
  }

  /**
   * 获取下拉菜单
   */
  getDropdown(): Locator {
    return this.page.locator('.n-dropdown-menu');
  }

  /**
   * 选择下拉菜单选项
   */
  async selectDropdownOption(label: string): Promise<void> {
    const option = this.page.locator(`.n-dropdown-option:has-text("${label}")`);
    await option.click();
    await sleep(200);
  }

  /**
   * 获取开关组件
   */
  getSwitch(): Locator {
    return this.page.locator('.n-switch');
  }

  /**
   * 打开开关
   */
  async turnOnSwitch(): Promise<void> {
    const switchEl = this.getSwitch();
    const checked = await switchEl.getAttribute('aria-checked');
    if (checked !== 'true') {
      await switchEl.click();
      await sleep(200);
    }
  }

  /**
   * 关闭开关
   */
  async turnOffSwitch(): Promise<void> {
    const switchEl = this.getSwitch();
    const checked = await switchEl.getAttribute('aria-checked');
    if (checked === 'true') {
      await switchEl.click();
      await sleep(200);
    }
  }

  /**
   * 获取树节点
   */
  getTreeNode(label: string): Locator {
    return this.page.locator(`.n-tree-node-content:has-text("${label}")`);
  }

  /**
   * 点击树节点
   */
  async clickTreeNode(label: string): Promise<void> {
    await this.getTreeNode(label).click();
    await sleep(200);
  }

  /**
   * 右键点击树节点
   */
  async rightClickTreeNode(label: string): Promise<void> {
    await this.getTreeNode(label).click({ button: 'right' });
    await sleep(200);
  }

  /**
   * 截图
   */
  async takeScreenshot(name: string): Promise<void> {
    await this.page.screenshot({ path: `test-results/screenshots/${name}.png` });
  }

  /**
   * 等待 Toast 消息
   */
  async waitForToast(message?: string): Promise<void> {
    const toast = message
      ? this.page.locator(`.n-message:has-text("${message}")`)
      : this.page.locator('.n-message');
    await toast.waitFor({ state: 'visible', timeout: 5000 });
  }

  /**
   * 获取加载状态
   */
  getLoadingSpinner(): Locator {
    return this.page.locator('.n-spin');
  }

  /**
   * 等待加载完成
   */
  async waitForLoadingComplete(): Promise<void> {
    const spinner = this.getLoadingSpinner();
    // 等待 spinner 出现（如果有的话）
    try {
      await spinner.waitFor({ state: 'visible', timeout: 1000 });
    } catch {
      // 没有 spinner，直接返回
      return;
    }
    // 等待 spinner 消失
    await spinner.waitFor({ state: 'hidden', timeout: 30000 });
  }

  /**
   * 执行 JavaScript
   */
  async evaluate<R>(fn: () => R): Promise<R> {
    return this.page.evaluate(fn);
  }

  /**
   * 滚动到元素
   */
  async scrollToElement(locator: Locator): Promise<void> {
    await locator.scrollIntoViewIfNeeded();
    await sleep(100);
  }

  /**
   * 滚动到页面底部
   */
  async scrollToBottom(): Promise<void> {
    await this.page.evaluate(() => window.scrollTo(0, document.body.scrollHeight));
    await sleep(100);
  }

  /**
   * 获取当前 URL
   */
  getCurrentUrl(): string {
    return this.page.url();
  }

  /**
   * 验证当前路由
   */
  async expectRoute(route: Route | string): Promise<void> {
    const url = this.getCurrentUrl();
    expect(url).toContain(route);
  }
}
