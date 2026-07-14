/**
 * 活动Wiki页面 Page Object
 * 对应 src/pages/activity-wiki-check/index.vue
 */

import {Page, Locator, expect} from '@playwright/test';
import {BasePage, resolveRoute} from './BasePage';
import {sleep} from '../utils/helpers';

export class ActivityWikiPage extends BasePage {
  readonly page: Page;

  constructor(page: Page) {
    super(page);
    this.page = page;
  }

  async goto(): Promise<void> {
    // 通过菜单点击导航（CDP 模式下 page.goto 不触发 Wails 路由）
    const btn = this.page.locator('#layout-header button:has-text("活动Wiki")');
    await btn.click();
    await sleep(800);
  }

  // 配置区域
  async setExcelDir(path: string): Promise<void> {
    const input = this.page.locator('input[placeholder*="Excel"]').first();
    if (await input.isVisible()) {
      await input.clear();
      await input.fill(path);
      await sleep(100);
    }
  }

  async setOldJsonPath(path: string): Promise<void> {
    const input = this.page.locator('input[placeholder*="JSON"]').first();
    if (await input.isVisible()) {
      await input.clear();
      await input.fill(path);
      await sleep(100);
    }
  }

  async clickExecuteCheck(): Promise<void> {
    await this.clickButton('执行检查');
  }

  async waitForCheckComplete(): Promise<void> {
    await this.waitForLoadingComplete();
    await sleep(1000);
  }

  // 活动卡片
  getActivityCards(): Locator {
    return this.page.locator('.activity-card');
  }

  getAccumulatedRechargeCards(): Locator {
    return this.page.locator('.activity-card').filter({hasText: 'ActTypeAccumulatedRecharge'});
  }

  async clickAccumulatedRechargeCard(index: number): Promise<void> {
    const cards = this.getAccumulatedRechargeCards();
    const card = cards.nth(index);
    await card.scrollIntoViewIfNeeded();
    await sleep(200);
  }

  // 页签
  getTabNames(): Promise<string[]> {
    return this.page.locator('.n-tabs-tab').allTextContents();
  }

  async clickTab(name: string): Promise<void> {
    await this.clickTab(name);
  }

  // 累充奖励页签
  async expectAccumulatedRechargeTabVisible(): Promise<void> {
    const tab = this.page.locator('.n-tabs-tab').filter({hasText: '累充奖励'});
    await expect(tab).toBeVisible();
  }

  getAccumulatedRechargeTable(): Locator {
    return this.page.locator('.n-table').filter({hasText: '累充金额'});
  }

  async expectAccumulatedRechargeTableHasData(): Promise<void> {
    const table = this.getAccumulatedRechargeTable();
    await expect(table).toBeVisible();
    const rows = table.locator('tbody tr');
    const count = await rows.count();
    expect(count).toBeGreaterThan(0);
  }

  // 页面布局验证
  async expectConfigAreaVisible(): Promise<void> {
    const configArea = this.page.locator('.n-card').first();
    await expect(configArea).toBeVisible();
  }

  async expectFilterAreaVisible(): Promise<void> {
    const filterArea = this.page.locator('.n-space');
    await expect(filterArea.first()).toBeVisible();
  }
}
