/**
 * 测试工具函数
 * 提供 Playwright 测试中常用的辅助方法
 */

import { Page, Locator, expect } from '@playwright/test';

/**
 * 等待指定毫秒数
 */
export function sleep(ms: number): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, ms));
}

/**
 * 等待元素可见
 */
export async function waitForVisible(locator: Locator, timeout = 10000): Promise<void> {
  await locator.waitFor({ state: 'visible', timeout });
}

/**
 * 等待元素隐藏
 */
export async function waitForHidden(locator: Locator, timeout = 10000): Promise<void> {
  await locator.waitFor({ state: 'hidden', timeout });
}

/**
 * 等待元素出现在 DOM 中
 */
export async function waitForAttached(locator: Locator, timeout = 10000): Promise<void> {
  await locator.waitFor({ state: 'attached', timeout });
}

/**
 * 点击元素并等待响应
 */
export async function clickAndWait(locator: Locator, waitForNavigation = false): Promise<void> {
  if (waitForNavigation) {
    await locator.click();
  } else {
    await locator.click();
    // 等待可能的 UI 更新
    await sleep(100);
  }
}

/**
 * 填写输入框并触发失焦
 */
export async function fillAndBlur(locator: Locator, value: string): Promise<void> {
  await locator.clear();
  await locator.fill(value);
  await locator.blur();
}

/**
 * 获取 Naive UI 开关组件的状态
 */
export async function getSwitchState(locator: Locator): Promise<boolean> {
  const checked = await locator.getAttribute('aria-checked');
  return checked === 'true';
}

/**
 * 设置 Naive UI 开关组件的状态
 */
export async function setSwitchState(locator: Locator, targetState: boolean): Promise<void> {
  const currentState = await getSwitchState(locator);
  if (currentState !== targetState) {
    await locator.click();
    await sleep(100);
  }
}

/**
 * 等待 Naive UI 加载状态消失
 */
export async function waitForLoadingComplete(page: Page, timeout = 30000): Promise<void> {
  const loadingSpinner = page.locator('.n-spin-container');
  try {
    await loadingSpinner.waitFor({ state: 'hidden', timeout });
  } catch {
    // 可能没有 loading 组件，忽略
  }
}

/**
 * 获取 Naive UI 树节点
 */
export function getTreeNode(page: Page, label: string): Locator {
  return page.locator(`.n-tree-node-content:has-text("${label}")`);
}

/**
 * 点击树节点
 */
export async function clickTreeNode(page: Page, label: string): Promise<void> {
  const node = getTreeNode(page, label);
  await node.click();
  await sleep(200);
}

/**
 * 展开树节点
 */
export async function expandTreeNode(page: Page, label: string): Promise<void> {
  const node = page.locator(`.n-tree-node:has-text("${label}") .n-tree-node-switcher`);
  const expanded = await node.getAttribute('data-expanded');
  if (expanded !== 'true') {
    await node.click();
    await sleep(200);
  }
}

/**
 * 右键点击树节点
 */
export async function rightClickTreeNode(page: Page, label: string): Promise<void> {
  const node = getTreeNode(page, label);
  await node.click({ button: 'right' });
  await sleep(200);
}

/**
 * 获取 Naive UI 下拉菜单选项
 */
export function getDropdownOption(page: Page, label: string): Locator {
  return page.locator(`.n-dropdown-option:has-text("${label}")`);
}

/**
 * 选择下拉菜单选项
 */
export async function selectDropdownOption(page: Page, label: string): Promise<void> {
  const option = getDropdownOption(page, label);
  await option.click();
  await sleep(100);
}

/**
 * 获取 Tab 标签
 */
export function getTabLabel(page: Page, label: string): Locator {
  return page.locator(`.n-tabs-tab:has-text("${label}")`);
}

/**
 * 点击 Tab 标签
 */
export async function clickTab(page: Page, label: string): Promise<void> {
  const tab = getTabLabel(page, label);
  await tab.click();
  await sleep(200);
}

/**
 * 获取 Naive UI 按钮组件
 */
export function getButton(page: Page, label: string | RegExp): Locator {
  return page.locator(`button:has-text("${label}")`);
}

/**
 * 点击按钮
 */
export async function clickButton(page: Page, label: string | RegExp): Promise<void> {
  const button = getButton(page, label);
  await button.click();
  await sleep(200);
}

/**
 * 获取输入框
 */
export function getInput(page: Page, placeholder?: string): Locator {
  if (placeholder) {
    return page.locator(`input[placeholder="${placeholder}"]`).first();
  }
  return page.locator('input').first();
}

/**
 * 获取文本域
 */
export function getTextarea(page: Page, placeholder?: string): Locator {
  if (placeholder) {
    return page.locator(`textarea[placeholder="${placeholder}"]`).first();
  }
  return page.locator('textarea').first();
}

/**
 * 检查元素是否包含文本
 */
export async function hasText(locator: Locator, text: string): Promise<boolean> {
  const content = await locator.textContent();
  return content?.includes(text) ?? false;
}

/**
 * 等待文本出现
 */
export async function waitForText(page: Page, text: string, timeout = 10000): Promise<void> {
  await page.waitForSelector(`text=${text}`, { timeout });
}

/**
 * 等待请求完成
 */
export async function waitForResponse(page: Page, urlPattern: string | RegExp): Promise<void> {
  await page.waitForResponse(response => {
    if (typeof urlPattern === 'string') {
      return response.url().includes(urlPattern);
    }
    return urlPattern.test(response.url());
  });
}

/**
 * 拖拽元素
 */
export async function dragElement(page: Page, source: Locator, target: Locator): Promise<void> {
  await source.hover();
  await page.mouse.down();
  await target.hover();
  await page.mouse.up();
  await sleep(200);
}

/**
 * 截图并比较
 */
export async function takeScreenshotAndCompare(
  page: Page,
  name: string,
  options?: { fullPage?: boolean }
): Promise<void> {
  await expect(page).toHaveScreenshot(`${name}.png`, options);
}

/**
 * 模拟键盘快捷键
 */
export async function pressShortcut(page: Page, keys: string): Promise<void> {
  await page.keyboard.press(keys);
  await sleep(100);
}

/**
 * 滚动到元素
 */
export async function scrollToElement(locator: Locator): Promise<void> {
  await locator.scrollIntoViewIfNeeded();
  await sleep(100);
}

/**
 * 滚动到页面底部
 */
export async function scrollToBottom(page: Page): Promise<void> {
  await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight));
  await sleep(100);
}

/**
 * 获取元素数量
 */
export async function getElementCount(locator: Locator): Promise<number> {
  return await locator.count();
}

/**
 * 等待元素数量变化
 */
export async function waitForCountChange(
  locator: Locator,
  expectedCount: number,
  timeout = 10000
): Promise<void> {
  const startTime = Date.now();
  while (Date.now() - startTime < timeout) {
    const count = await locator.count();
    if (count === expectedCount) return;
    await sleep(100);
  }
  throw new Error(`Element count did not reach ${expectedCount} within ${timeout}ms`);
}
