/**
 * 功能测试页 Page Object
 * 对应 src/pages/function-test/index.vue
 */

import { Page, Locator, expect } from '@playwright/test';
import { BasePage, Route, resolveRoute } from './BasePage';
import { sleep } from '../utils/helpers';

/**
 * 功能测试页 Page Object
 * 测试用例编辑器页面
 */
export class FunctionTestPage extends BasePage {
  // 页面容器
  readonly pageContainer: Locator;

  // Header 区域
  readonly headerMenu: Locator;
  readonly loadButton: Locator;
  readonly saveButton: Locator;
  readonly executeButton: Locator;
  readonly stopButton: Locator;
  readonly settingsButton: Locator;

  // 左侧树面板
  readonly siderPanel: Locator;
  readonly searchInput: Locator;
  readonly filterSwitch: Locator;
  readonly showDescSwitch: Locator;
  readonly testCaseTree: Locator;

  // Tab 面板
  readonly tabsContainer: Locator;
  readonly configTab: Locator;
  readonly stepsTab: Locator;
  readonly logTab: Locator;

  // Footer 区域
  readonly footerBar: Locator;
  readonly caseCount: Locator;
  readonly stepCount: Locator;

  constructor(page: Page) {
    super(page);

    // 页面容器
    this.pageContainer = page.locator('#Test');

    // Header - 菜单按钮在 .n-menu-item 内
    this.headerMenu = page.locator('.n-layout-header .n-menu');
    this.loadButton = page.locator('.n-menu-item:has-text("加载用例")');
    this.saveButton = page.locator('.n-menu-item:has-text("保存用例")');
    this.executeButton = page.locator('.n-menu-item:has-text("执行用例")');
    this.stopButton = page.locator('.n-menu-item:has-text("停止用例")');
    this.settingsButton = page.locator('.n-menu-item:has-text("设置")');

    // Sider
    this.siderPanel = page.locator('.n-layout-sider');
    this.searchInput = page.locator('.n-layout-sider input[placeholder="搜索"]');
    this.filterSwitch = page.locator('.n-layout-sider .n-switch').first();
    this.showDescSwitch = page.locator('.n-layout-sider .n-switch').nth(1);
    this.testCaseTree = page.locator('.n-tree');

    // Tabs
    this.tabsContainer = page.locator('.n-tabs');
    this.configTab = page.locator('.n-tabs-tab:has-text("用例配置")');
    this.stepsTab = page.locator('.n-tabs-tab:has-text("用例步骤")');
    this.logTab = page.locator('.n-tabs-tab:has-text("执行日志")');

    // Footer
    this.footerBar = page.locator('.n-layout-footer');
    this.caseCount = page.locator('.n-statistic').first();
    this.stepCount = page.locator('.n-statistic').nth(1);
  }

  /**
   * 导航到功能测试页
   */
  async goto(): Promise<void> {
    await this.page.locator('#layout-header button:has-text("战斗测试")').click();
    await sleep(800);
  }

  /**
   * 验证页面已加载
   */
  async expectPageLoaded(): Promise<void> {
    await expect(this.pageContainer).toBeVisible();
    await expect(this.headerMenu).toBeVisible();
  }

  // ==================== Header 操作 ====================

  /**
   * 点击加载用例
   */
  async clickLoadCases(): Promise<void> {
    await this.loadButton.click();
    await sleep(500);
  }

  /**
   * 点击保存用例
   */
  async clickSaveCases(): Promise<void> {
    await this.saveButton.click();
    await sleep(300);
  }

  /**
   * 点击执行用例
   */
  async clickExecuteCases(): Promise<void> {
    await this.executeButton.click();
    await sleep(500);
  }

  /**
   * 点击停止用例
   */
  async clickStopCases(): Promise<void> {
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

  // ==================== 树操作 ====================

  /**
   * 搜索用例
   */
  async searchCases(keyword: string): Promise<void> {
    await this.searchInput.clear();
    await this.searchInput.fill(keyword);
    await sleep(300);
  }

  /**
   * 清空搜索
   */
  async clearSearch(): Promise<void> {
    await this.searchInput.clear();
    await sleep(200);
  }

  /**
   * 切换过滤开关
   */
  async toggleFilter(show: boolean): Promise<void> {
    const isChecked = await this.filterSwitch.getAttribute('aria-checked');
    const isCurrentlyOn = isChecked === 'true';
    if (isCurrentlyOn !== show) {
      await this.filterSwitch.click();
      await sleep(200);
    }
  }

  /**
   * 切换描述显示
   */
  async toggleShowDesc(show: boolean): Promise<void> {
    const isChecked = await this.showDescSwitch.getAttribute('aria-checked');
    const isCurrentlyOn = isChecked === 'true';
    if (isCurrentlyOn !== show) {
      await this.showDescSwitch.click();
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
   * 展开树分类
   */
  async expandCategory(label: string): Promise<void> {
    const node = this.page.locator(`.n-tree-node:has-text("${label}") .n-tree-node-switcher`);
    const expanded = await node.getAttribute('data-expanded');
    if (expanded !== 'true') {
      await node.click();
      await sleep(200);
    }
  }

  /**
   * 点击树节点（选择用例）
   */
  async clickTreeNode(label: string): Promise<void> {
    const node = this.getTreeNode(label);
    await node.click();
    await sleep(300);
  }

  /**
   * 右键点击树节点
   */
  async rightClickTreeNode(label: string): Promise<void> {
    const node = this.getTreeNode(label);
    await node.click({ button: 'right' });
    await sleep(200);
  }

  /**
   * 获取右键菜单
   */
  getContextMenu(): Locator {
    return this.page.locator('.n-dropdown-menu');
  }

  /**
   * 点击右键菜单选项
   */
  async clickContextMenuOption(option: string): Promise<void> {
    const menu = this.getContextMenu();
    await menu.locator(`.n-dropdown-option:has-text("${option}")`).click();
    await sleep(300);
  }

  // ==================== Tab 切换 ====================

  /**
   * 切换到用例配置 Tab
   */
  async switchToConfigTab(): Promise<void> {
    await this.configTab.click();
    await sleep(200);
  }

  /**
   * 切换到用例步骤 Tab
   */
  async switchToStepsTab(): Promise<void> {
    await this.stepsTab.click();
    await sleep(200);
  }

  /**
   * 切换到执行日志 Tab
   */
  async switchToLogTab(): Promise<void> {
    await this.logTab.click();
    await sleep(200);
  }

  // ==================== Footer 操作 ====================

  /**
   * 获取用例数量统计
   */
  async getCaseCountValue(): Promise<string> {
    return await this.caseCount.locator('.n-statistic-value').textContent() || '0';
  }

  /**
   * 获取动作数量统计
   */
  async getStepCountValue(): Promise<string> {
    return await this.stepCount.locator('.n-statistic-value').textContent() || '0';
  }

  // ==================== 验证方法 ====================

  /**
   * 验证用例已加载
   */
  async expectCasesLoaded(): Promise<void> {
    const count = await this.getCaseCountValue();
    expect(parseInt(count)).toBeGreaterThan(0);
  }

  /**
   * 验证树节点存在
   */
  async expectTreeNodeExists(label: string): Promise<void> {
    await expect(this.getTreeNode(label)).toBeVisible();
  }

  /**
   * 验证树节点不存在
   */
  async expectTreeNodeNotExists(label: string): Promise<void> {
    await expect(this.getTreeNode(label)).not.toBeVisible();
  }
}
