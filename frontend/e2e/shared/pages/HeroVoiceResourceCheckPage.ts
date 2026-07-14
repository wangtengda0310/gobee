/**
 * 语音资源检查页 Page Object
 * 对应 src/pages/hero-voice-resource-check/index.vue
 */

import { Page, Locator, expect } from '@playwright/test';
import { BasePage, resolveRoute } from './BasePage';
import { sleep } from '../utils/helpers';

/**
 * 语音资源检查页 Page Object
 */
export class HeroVoiceResourceCheckPage extends BasePage {
  readonly configCard: Locator;
  readonly excelDirInput: Locator;
  readonly cardDirInput: Locator;
  readonly startCheckButton: Locator;
  readonly errorList: Locator;
  readonly heroGroup: Locator;

  constructor(page: Page) {
    super(page);
    this.configCard = page.locator('.path-config-row');
    this.excelDirInput = page.locator('input[placeholder*="配表"]');
    this.cardDirInput = page.locator('input[placeholder*="Card"]');
    this.startCheckButton = page.locator('button:has-text("开始检索")');
    this.errorList = page.locator('.err-list');
    this.heroGroup = page.locator('.errInfo');
  }

  async goto(): Promise<void> {
    await this.page.locator('#layout-header button:has-text("武将资源检查")').click();
    await sleep(800);
  }

  async setExcelDir(path: string): Promise<void> {
    await this.excelDirInput.clear();
    await this.excelDirInput.fill(path);
    await this.excelDirInput.blur();
    await sleep(200);
  }

  async setCardDir(path: string): Promise<void> {
    await this.cardDirInput.clear();
    await this.cardDirInput.fill(path);
    await this.cardDirInput.blur();
    await sleep(200);
  }

  async clickStartCheck(): Promise<void> {
    await this.startCheckButton.click();
    await sleep(500);
  }

  async waitForCheckComplete(timeout = 60000): Promise<void> {
    await this.waitForLoadingComplete();
  }

  getErrorItems(): Locator {
    return this.errorList.locator('.errInfo');
  }

  async getErrorCount(): Promise<number> {
    return await this.getErrorItems().count();
  }

  getErrorItem(index: number): Locator {
    return this.getErrorItems().nth(index);
  }

  async expectConfigAreaVisible(): Promise<void> {
    await expect(this.configCard).toBeVisible();
  }

  async expectNoErrors(): Promise<void> {
    await expect(this.errorList).not.toBeVisible();
  }

  async expectCheckButtonEnabled(): Promise<void> {
    await expect(this.startCheckButton).toBeEnabled();
  }
}
