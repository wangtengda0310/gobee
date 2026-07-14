/**
 * 配表测试页 Page Object
 * 对应 src/pages/excel-test/index.vue
 */

import { Page, Locator, expect } from '@playwright/test';
import { BasePage, Route, resolveRoute } from './BasePage';
import { sleep } from '../utils/helpers';

/**
 * 配表测试页 Page Object
 * Excel 配置检查页面
 */
export class ExcelTestPage extends BasePage {
  // Header 区域
  readonly headerMenu: Locator;
  readonly loadButton: Locator;
  readonly saveButton: Locator;
  readonly executeButton: Locator;
  readonly stopButton: Locator;
  readonly settingsButton: Locator;
  readonly excelDirInput: Locator;
  readonly caseDirInput: Locator;

  // 左侧树面板
  readonly siderPanel: Locator;
  readonly searchInput: Locator;
  readonly excelTree: Locator;

  // Tab 面板
  readonly tabsContainer: Locator;
  readonly managerTab: Locator;
  readonly configTab: Locator;
  readonly logTab: Locator;

  // Footer 区域
  readonly footerBar: Locator;
  readonly excelCount: Locator;
  readonly sheetCount: Locator;
  readonly successCount: Locator;
  readonly errorCount: Locator;
  readonly errorCellCount: Locator;

  constructor(page: Page) {
    super(page);

    // Header — UI 文案在 excel-test 页面体已变更（"加载配表"等），用 #Excel 限定避免与 layout-header 冲突
    this.headerMenu = page.locator('#layout-header');
    this.loadButton = page.locator('#Excel button:has-text("加载配表")');
    this.saveButton = page.locator('#Excel button:has-text("保存用例")');
    this.executeButton = page.locator('#Excel button:has-text("执行检查")');
    this.stopButton = page.locator('#Excel button:has-text("停止检查")');
    this.settingsButton = page.locator('#Excel button:has-text("设置")');
    this.excelDirInput = page.locator('#Excel input[placeholder*="配表"]');
    this.caseDirInput = page.locator('#Excel input[placeholder*="用例"]');

    // Sider
    this.siderPanel = page.locator('.n-layout-sider');
    this.searchInput = page.locator('.n-layout-sider input[placeholder*="搜索"]');
    this.excelTree = page.locator('.n-tree');

    // Tabs
    this.tabsContainer = page.locator('.n-tabs');
    this.managerTab = page.locator('.n-tabs-tab:has-text("负责人")');
    this.configTab = page.locator('.n-tabs-tab:has-text("用例配置")');
    this.logTab = page.locator('.n-tabs-tab:has-text("执行日志")');

    // Footer
    this.footerBar = page.locator('.n-layout-footer');
    this.excelCount = page.locator('.n-statistic').first();
    this.sheetCount = page.locator('.n-statistic').nth(1);
    this.successCount = page.locator('.n-statistic').nth(2);
    this.errorCount = page.locator('.n-statistic').nth(3);
    this.errorCellCount = page.locator('.n-statistic').nth(4);
  }

  /**
   * 导航到配表测试页
   * 通过点击导航按钮，因为 CDP 连接的 WebView2 不支持 page.goto() 相对 URL
   */
  async goto(): Promise<void> {
    const isExcelTest = await this.page.locator('.excel-test-page').isVisible().catch(() => false);
    if (isExcelTest) {
      return;
    }
    await this.clickNavButton('配表测试');
    await this.waitForPageLoad();
  }

  // ==================== Header 操作 ====================

  /**
   * 点击加载配置
   */
  async clickLoadConfig(): Promise<void> {
    await this.loadButton.click();
    await sleep(500);
  }

  /**
   * 点击保存配置
   */
  async clickSaveConfig(): Promise<void> {
    await this.saveButton.click();
    await sleep(300);
  }

  /**
   * 点击执行检查
   */
  async clickExecuteCheck(): Promise<void> {
    await this.executeButton.click();
    await sleep(500);
  }

  /**
   * 点击停止检查
   */
  async clickStopCheck(): Promise<void> {
    await this.stopButton.click();
    await sleep(300);
  }

  /**
   * 点击设置按钮
   */
  async clickSettings(): Promise<void> {
    await this.settingsButton.click();
    await sleep(300);
  }

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
   * 设置用例目录
   */
  async setCaseDir(path: string): Promise<void> {
    await this.caseDirInput.clear();
    await this.caseDirInput.fill(path);
    await this.caseDirInput.blur();
    await sleep(200);
  }

  // ==================== 树操作 ====================

  /**
   * 搜索 Excel
   */
  async searchExcel(keyword: string): Promise<void> {
    await this.searchInput.clear();
    await this.searchInput.fill(keyword);
    await sleep(300);
  }

  /**
   * 获取 Excel 节点
   */
  getExcelNode(fileName: string): Locator {
    return this.excelTree.locator(`.n-tree-node:has-text("${fileName}")`);
  }

  /**
   * 获取 Sheet 节点
   */
  getSheetNode(sheetName: string): Locator {
    return this.excelTree.locator(`.n-tree-node:has-text("${sheetName}")`);
  }

  /**
   * 点击 Sheet 节点
   */
  async clickSheetNode(sheetName: string): Promise<void> {
    await this.getSheetNode(sheetName).click();
    await sleep(500);
  }

  /**
   * 展开 Excel 节点
   */
  async expandExcelNode(fileName: string): Promise<void> {
    const node = this.getExcelNode(fileName);
    const switcher = node.locator('.n-tree-node-switcher');
    const expanded = await switcher.getAttribute('data-expanded');
    if (expanded !== 'true') {
      await switcher.click();
      await sleep(200);
    }
  }

  /**
   * 勾选 Sheet
   */
  async checkSheet(sheetName: string): Promise<void> {
    const checkbox = this.getSheetNode(sheetName).locator('.n-checkbox');
    await checkbox.click();
    await sleep(200);
  }

  // ==================== Tab 切换 ====================

  /**
   * 切换到负责人 Tab
   */
  async switchToManagerTab(): Promise<void> {
    await this.managerTab.click();
    await sleep(200);
  }

  /**
   * 切换到用例配置 Tab
   */
  async switchToConfigTab(): Promise<void> {
    await this.configTab.click();
    await sleep(200);
  }

  /**
   * 切换到执行日志 Tab
   */
  async switchToLogTab(): Promise<void> {
    await this.logTab.click();
    await sleep(200);
  }

  // ==================== 负责人管理 ====================

  /**
   * 获取负责人输入框
   */
  getOwnerInput(sheetName: string): Locator {
    return this.page.locator(`.excel-check-manager tr:has-text("${sheetName}") input`);
  }

  /**
   * 设置负责人
   */
  async setOwner(sheetName: string, owner: string): Promise<void> {
    const input = this.getOwnerInput(sheetName);
    await input.clear();
    await input.fill(owner);
    await input.blur();
    await sleep(200);
  }

  // ==================== 用例配置 ====================

  /**
   * 获取规则卡片列表
   */
  getRuleCards(): Locator {
    return this.page.locator('.excel-check-panel .rule-card');
  }

  /**
   * 添加规则
   */
  async addRule(columnName: string): Promise<void> {
    const addBtn = this.page.locator(`.rule-group:has-text("${columnName}") button:has-text("添加规则")`);
    await addBtn.click();
    await sleep(200);
  }

  /**
   * 选择规则类型
   */
  async selectRuleType(ruleIndex: number, ruleType: string): Promise<void> {
    const rules = this.getRuleCards();
    const select = rules.nth(ruleIndex).locator('.n-select');
    await select.click();
    await sleep(200);
    await this.selectDropdownOption(ruleType);
  }

  /**
   * 切换规则开关
   */
  async toggleRule(ruleIndex: number): Promise<void> {
    const rules = this.getRuleCards();
    const switchEl = rules.nth(ruleIndex).locator('.n-switch');
    await switchEl.click();
    await sleep(200);
  }

  /**
   * 删除规则
   */
  async removeRule(ruleIndex: number): Promise<void> {
    const rules = this.getRuleCards();
    const deleteBtn = rules.nth(ruleIndex).locator('button:has-text("删除")');
    await deleteBtn.click();
    await sleep(200);
  }

  // ==================== 执行日志 ====================

  /**
   * 获取日志列表
   */
  getLogList(): Locator {
    return this.page.locator('.excel-check-log .log-item');
  }

  /**
   * 筛选日志级别
   */
  async filterLogLevel(level: string): Promise<void> {
    await this.page.locator(`.excel-check-log button:has-text("${level}")`).click();
    await sleep(200);
  }

  /**
   * 清空日志
   */
  async clearLogs(): Promise<void> {
    await this.page.locator('.excel-check-log button:has-text("清空")').click();
    await sleep(200);
  }

  // ==================== Footer 统计 ====================

  /**
   * 获取 Excel 文件数
   */
  async getExcelCountValue(): Promise<string> {
    return await this.excelCount.locator('.n-statistic-value').textContent() || '0';
  }

  /**
   * 获取 Sheet 总数
   */
  async getSheetCountValue(): Promise<string> {
    return await this.sheetCount.locator('.n-statistic-value').textContent() || '0';
  }

  /**
   * 获取成功数
   */
  async getSuccessCountValue(): Promise<string> {
    return await this.successCount.locator('.n-statistic-value').textContent() || '0';
  }

  /**
   * 获取错误数
   */
  async getErrorCountValue(): Promise<string> {
    return await this.errorCount.locator('.n-statistic-value').textContent() || '0';
  }

  /**
   * 获取错误单元格数
   */
  async getErrorCellCountValue(): Promise<string> {
    return await this.errorCellCount.locator('.n-statistic-value').textContent() || '0';
  }

  // ==================== 验证方法 ====================

  /**
   * 验证配置已加载
   */
  async expectConfigLoaded(): Promise<void> {
    const count = await this.getExcelCountValue();
    expect(parseInt(count)).toBeGreaterThan(0);
  }

  /**
   * 验证正在检查
   */
  async expectIsChecking(): Promise<void> {
    await expect(this.stopButton).toBeVisible();
    await expect(this.executeButton).toBeHidden();
  }

  /**
   * 验证检查完成
   */
  async expectCheckComplete(): Promise<void> {
    await expect(this.executeButton).toBeVisible();
    await expect(this.stopButton).toBeHidden();
  }
}
