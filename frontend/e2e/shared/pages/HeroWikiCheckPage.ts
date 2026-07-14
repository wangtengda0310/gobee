/**
 * Wiki 检查页 Page Object
 * 对应 src/pages/hero-wiki-check/index.vue
 */

import { Page, Locator, expect } from '@playwright/test';
import { BasePage, Route, resolveRoute } from './BasePage';
import { sleep } from '../utils/helpers';

/**
 * Wiki 检查页 Page Object
 * 武将差异检查页面
 */
export class HeroWikiCheckPage extends BasePage {
  // 配置区域
  readonly configCard: Locator;
  readonly excelDirInput: Locator;
  readonly oldJsonInput: Locator;
  readonly executeButton: Locator;
  readonly saveButton: Locator;

  // 统计区域
  readonly summaryCard: Locator;
  readonly totalChangeTag: Locator;
  readonly addedTag: Locator;
  readonly deletedTag: Locator;
  readonly modifiedTag: Locator;

  // 筛选区域
  readonly filterCard: Locator;
  readonly searchInput: Locator;
  readonly countrySelect: Locator;
  readonly newHeroCheckbox: Locator;
  readonly gachaCheckbox: Locator;
  readonly isOpenCheckbox: Locator;

  // 武将列表
  readonly heroList: Locator;
  readonly heroPanel: Locator;

  // 锚点导航
  readonly anchorNav: Locator;

  constructor(page: Page) {
    super(page);

    // 配置区域
    this.configCard = page.locator('.n-card').first();
    this.excelDirInput = page.locator('input[placeholder*="Excel"]');
    this.oldJsonInput = page.locator('input[placeholder*="JSON"]');
    this.executeButton = page.locator('button:has-text("执行检查")');
    this.saveButton = page.locator('button:has-text("保存结果")');

    // 统计区域
    this.summaryCard = page.locator('.n-card').nth(1);
    this.totalChangeTag = page.locator('.n-tag:has-text("总变化")');
    this.addedTag = page.locator('.n-tag:has-text("新增")');
    this.deletedTag = page.locator('.n-tag:has-text("删除")');
    this.modifiedTag = page.locator('.n-tag:has-text("修改")');

    // 筛选区域
    this.filterCard = page.locator('.n-card').nth(2);
    this.searchInput = page.locator('input[placeholder*="搜索武将"]');
    this.countrySelect = page.locator('.n-select').first();
    this.newHeroCheckbox = page.locator('.n-checkbox').first();
    this.gachaCheckbox = page.locator('.n-checkbox').nth(1);
    this.isOpenCheckbox = page.locator('.n-checkbox').nth(2);

    // 武将列表
    this.heroList = page.locator('.hero-list');
    this.heroPanel = page.locator('.hero-panel');

    // 锚点导航
    this.anchorNav = page.locator('.n-anchor');
  }

  /**
   * 导航到 Wiki 检查页
   */
  async goto(): Promise<void> {
    await this.page.locator('#layout-header button:has-text("武将Wiki检查")').click();
    await sleep(800);
  }

  // ==================== 配置操作 ====================

  /**
   * 设置 Excel 目录
   */
  async setExcelDir(path: string): Promise<void> {
    await this.excelDirInput.clear();
    await this.excelDirInput.fill(path);
    await this.excelDirInput.blur();
    await sleep(200);
  }

  /**
   * 设置历史 JSON 路径
   */
  async setOldJsonPath(path: string): Promise<void> {
    await this.oldJsonInput.clear();
    await this.oldJsonInput.fill(path);
    await this.oldJsonInput.blur();
    await sleep(200);
  }

  /**
   * 点击执行检查
   */
  async clickExecuteCheck(): Promise<void> {
    await this.executeButton.click();
    await sleep(500);
  }

  /**
   * 点击保存结果
   */
  async clickSaveResult(): Promise<void> {
    await this.saveButton.click();
    await sleep(300);
  }

  // ==================== 统计操作 ====================

  /**
   * 获取总变化数
   */
  async getTotalChangeCount(): Promise<number> {
    const text = await this.totalChangeTag.textContent() || '0';
    const match = text.match(/\d+/);
    return match ? parseInt(match[0]) : 0;
  }

  /**
   * 获取新增数
   */
  async getAddedCount(): Promise<number> {
    const text = await this.addedTag.textContent() || '0';
    const match = text.match(/\d+/);
    return match ? parseInt(match[0]) : 0;
  }

  /**
   * 获取删除数
   */
  async getDeletedCount(): Promise<number> {
    const text = await this.deletedTag.textContent() || '0';
    const match = text.match(/\d+/);
    return match ? parseInt(match[0]) : 0;
  }

  /**
   * 获取修改数
   */
  async getModifiedCount(): Promise<number> {
    const text = await this.modifiedTag.textContent() || '0';
    const match = text.match(/\d+/);
    return match ? parseInt(match[0]) : 0;
  }

  /**
   * 点击统计标签筛选
   */
  async clickSummaryTag(type: 'total' | 'added' | 'deleted' | 'modified'): Promise<void> {
    const tags = {
      total: this.totalChangeTag,
      added: this.addedTag,
      deleted: this.deletedTag,
      modified: this.modifiedTag,
    };
    await tags[type].click();
    await sleep(200);
  }

  // ==================== 筛选操作 ====================

  /**
   * 搜索武将
   */
  async searchHero(name: string): Promise<void> {
    await this.searchInput.clear();
    await this.searchInput.fill(name);
    await sleep(300);
  }

  /**
   * 选择势力
   */
  async selectCountry(country: string): Promise<void> {
    await this.countrySelect.click();
    await sleep(200);
    await this.selectDropdownOption(country);
  }

  /**
   * 清除势力选择
   */
  async clearCountryFilter(): Promise<void> {
    const clearBtn = this.countrySelect.locator('.n-base-clear');
    if (await clearBtn.isVisible()) {
      await clearBtn.click();
      await sleep(200);
    }
  }

  /**
   * 勾选新武将筛选
   */
  async checkNewHeroFilter(): Promise<void> {
    if (!(await this.newHeroCheckbox.isChecked())) {
      await this.newHeroCheckbox.click();
      await sleep(200);
    }
  }

  /**
   * 取消新武将筛选
   */
  async uncheckNewHeroFilter(): Promise<void> {
    if (await this.newHeroCheckbox.isChecked()) {
      await this.newHeroCheckbox.click();
      await sleep(200);
    }
  }

  /**
   * 勾选抽奖筛选
   */
  async checkGachaFilter(): Promise<void> {
    if (!(await this.gachaCheckbox.isChecked())) {
      await this.gachaCheckbox.click();
      await sleep(200);
    }
  }

  /**
   * 取消抽奖筛选
   */
  async uncheckGachaFilter(): Promise<void> {
    if (await this.gachaCheckbox.isChecked()) {
      await this.gachaCheckbox.click();
      await sleep(200);
    }
  }

  /**
   * 勾选已开放筛选
   */
  async checkIsOpenFilter(): Promise<void> {
    if (!(await this.isOpenCheckbox.isChecked())) {
      await this.isOpenCheckbox.click();
      await sleep(200);
    }
  }

  /**
   * 取消已开放筛选
   */
  async uncheckIsOpenFilter(): Promise<void> {
    if (await this.isOpenCheckbox.isChecked()) {
      await this.isOpenCheckbox.click();
      await sleep(200);
    }
  }

  // ==================== 武将列表操作 ====================

  /**
   * 获取武将面板列表
   */
  getHeroPanels(): Locator {
    return this.heroPanel;
  }

  /**
   * 获取武将数量
   */
  async getHeroCount(): Promise<number> {
    return await this.heroPanel.count();
  }

  /**
   * 获取指定武将面板
   */
  getHeroPanelByName(name: string): Locator {
    return this.page.locator(`.hero-panel:has-text("${name}")`);
  }

  /**
   * 展开武将面板
   */
  async expandHeroPanel(name: string): Promise<void> {
    const panel = this.getHeroPanelByName(name);
    await panel.click();
    await sleep(200);
  }

  /**
   * 获取武将差异详情
   */
  getHeroDiffDisplay(panelName: string): Locator {
    return this.getHeroPanelByName(panelName).locator('.hero-diff-display');
  }

  /**
   * 获取技能加成显示
   */
  getBuffDisplay(panelName: string): Locator {
    return this.getHeroPanelByName(panelName).locator('.buff-display');
  }

  /**
   * 获取掉落显示
   */
  getDropDisplay(panelName: string): Locator {
    return this.getHeroPanelByName(panelName).locator('.drop-display');
  }

  // ==================== 锚点导航 ====================

  /**
   * 点击锚点链接
   */
  async clickAnchorLink(heroName: string): Promise<void> {
    await this.anchorNav.locator(`.n-anchor-link:has-text("${heroName}")`).click();
    await sleep(300);
  }

  /**
   * 获取锚点链接列表
   */
  getAnchorLinks(): Locator {
    return this.anchorNav.locator('.n-anchor-link');
  }

  // ==================== 验证方法 ====================

  /**
   * 验证有差异结果
   */
  async expectHasDiffResult(): Promise<void> {
    await expect(this.summaryCard).toBeVisible();
  }

  /**
   * 验证无差异结果
   */
  async expectNoDiffResult(): Promise<void> {
    await expect(this.summaryCard).toBeHidden();
  }

  /**
   * 验证武将存在
   */
  async expectHeroExists(name: string): Promise<void> {
    await expect(this.getHeroPanelByName(name)).toBeVisible();
  }

  /**
   * 验证武将不存在
   */
  async expectHeroNotExists(name: string): Promise<void> {
    await expect(this.getHeroPanelByName(name)).not.toBeVisible();
  }

  /**
   * 验证锚点导航可见
   */
  async expectAnchorNavVisible(): Promise<void> {
    await expect(this.anchorNav).toBeVisible();
  }
}
