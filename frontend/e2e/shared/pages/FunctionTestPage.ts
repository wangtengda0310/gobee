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

  // 用例步骤面板
  readonly stepsPanel: Locator;

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

    // Tabs — 限定主 Tab 栏，避免与执行日志内嵌 Tab 混淆
    const mainTabs = page.locator('[data-testid="fight-main-tabs"]');
    this.tabsContainer = mainTabs;
    this.configTab = mainTabs.locator('.n-tabs-tab:has-text("用例配置")');
    this.stepsTab = mainTabs.locator('.n-tabs-tab:has-text("用例步骤")');
    this.logTab = mainTabs.locator('.n-tabs-tab:has-text("执行日志")');

    // Footer
    this.footerBar = page.locator('.n-layout-footer');
    this.caseCount = page.locator('.n-statistic').first();
    this.stepCount = page.locator('.n-statistic').nth(1);

    // Steps panel
    this.stepsPanel = page.locator('[data-testid="steps-panel"]');
  }

  /**
   * 导航到功能测试页
   */
  async goto(): Promise<void> {
    // 清前一个测试残留的 modal/drawer 遮罩（串行测试状态污染，E2E CLAUDE.md 第8/14条），
    // 否则后续点击导航按钮会被残留 mask 拦截
    for (let i = 0; i < 3; i++) {
      if (await this.page.locator('.n-modal-mask:visible, .n-drawer-mask:visible').count() === 0) break;
      await this.page.keyboard.press('Escape');
      await sleep(300);
    }
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
   *
   * 内存中已有用例时，前端会弹出"加载用例会覆盖当前内存中已加载的用例"的二次确认 dialog，
   * 这里检测到就点「确定」放行；未弹出（内存为空）则跳过。否则确认 dialog 的 mask 会拦截
   * 后续所有点击（曾导致 loadCaseWithSteps 点用例节点超时）。
   */
  async clickLoadCases(): Promise<void> {
    await this.loadButton.click();
    await sleep(500);
    const confirmDialog = this.page.locator('.n-modal').filter({hasText: '加载用例会覆盖'});
    if (await confirmDialog.isVisible({timeout: 1500}).catch(() => false)) {
      await confirmDialog.locator('button:has-text("确定")').click();
      // 等确认 dialog 及其 mask 完全关闭，避免后续点击被残留 mask 拦截
      await expect(confirmDialog).toBeHidden({timeout: 10000});
      await sleep(200);
    }
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

  /**
   * 断言主 Tab 导航在切换前后不发生水平偏移（回归：点击执行日志后标题整行右移）
   */
  async expectMainTabNavStableAfter(action: () => Promise<void>): Promise<void> {
    const nav = this.tabsContainer.locator('> .n-tabs-nav');
    const boxBefore = await nav.boundingBox();
    expect(boxBefore).not.toBeNull();

    await action();
    await sleep(300);

    const boxAfter = await nav.boundingBox();
    expect(boxAfter).not.toBeNull();
    expect(Math.abs(boxAfter!.x - boxBefore!.x)).toBeLessThanOrEqual(1);
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

  // ==================== 用例步骤面板 ====================

  /**
   * 加载并选中包含步骤的用例，切换到「用例步骤」Tab
   *
   * nth(0)=分类、nth(1)=第一个用例，依赖搜索结果单一分类匹配；多分类匹配场景需改用 data-node-type
   * （曾尝试给 n-tree node-props 透传 data-node-type，但 naive-ui node-props 实测不会把 data-* 写到
   * .n-tree-node 的 DOM attribute，仅绑定事件监听；要落地需改 render-label 包裹 data 属性，改动较大，暂不做）。
   */
  async loadCaseWithSteps(searchKeyword = '行动牌'): Promise<void> {
    await this.clickLoadCases();
    await sleep(500);
    await this.searchCases(searchKeyword);
    await sleep(300);

    const matched = this.testCaseTree
      .locator('.n-tree-node-content')
      .filter({ hasText: searchKeyword });

    // 分类节点默认折叠时用例不渲染（nth(1) 找不到）；若 nth(1) 不可见则展开第一个匹配（分类）的 switcher
    if (!(await matched.nth(1).isVisible().catch(() => false))) {
      const firstMatchedNode = this.testCaseTree
        .locator('.n-tree-node')
        .filter({ hasText: searchKeyword })
        .first();
      await firstMatchedNode.locator('.n-tree-node-switcher').click().catch(() => {});
      await sleep(400);
    }

    // 点第一个具体用例（nth(0) 是分类"N_xxx"，nth(1) 是用例），确保 caseSteps 被加载
    const caseNode = matched.nth(1);
    await expect(caseNode).toBeVisible({ timeout: 10000 });
    await caseNode.click();
    await sleep(400);

    await this.switchToStepsTab();
    await expect(this.stepsPanel).toBeVisible({ timeout: 10000 });
    await expect(this.getActionStepCards().first()).toBeVisible({ timeout: 10000 });
  }

  /** 顶层动作卡片列表 */
  getActionStepCards(): Locator {
    return this.stepsPanel.locator('[data-testid="action-step-card"]');
  }

  /** 动作卡片 header-extra 区域 */
  getActionStepHeaderExtra(stepIndex: number): Locator {
    return this.getActionStepCards()
      .nth(stepIndex)
      .locator('[data-testid="action-step-header-extra"]');
  }

  /** 指定动作下的断言卡片 */
  getAssertionCards(stepIndex: number): Locator {
    return this.getActionStepCards()
      .nth(stepIndex)
      .locator('[data-testid="asset-card"]');
  }

  /** 断言卡片 header-extra 区域 */
  getAssertionHeaderExtra(stepIndex: number, assetIndex: number): Locator {
    return this.getAssertionCards(stepIndex)
      .nth(assetIndex)
      .locator('[data-testid="asset-card-header-extra"]');
  }

  /** 断言卡片正文中的类型选择行 */
  getAssertionTypeRow(stepIndex: number, assetIndex: number): Locator {
    return this.getAssertionCards(stepIndex)
      .nth(assetIndex)
      .locator('[data-testid="asset-type-row"]');
  }

  /**
   * 验证卡片 header-extra 具备统一布局：拖动 + 应用智能描述 + 描述输入
   */
  async expectCardHeaderExtraLayout(headerExtra: Locator): Promise<void> {
    await expect(headerExtra.locator('button').filter({ hasText: '拖动' })).toBeVisible();
    await expect(headerExtra.locator('button').filter({ hasText: '应用智能描述' })).toBeVisible();
    await expect(headerExtra.locator('input').first()).toBeVisible();
  }

  /** 点击动作卡片底部的「新增断言」 */
  async clickAddAssertion(stepIndex = 0): Promise<void> {
    const card = this.getActionStepCards().nth(stepIndex);
    await card.locator('button').filter({ hasText: '新增断言' }).click();
    await sleep(300);
  }

  /** 点击动作卡片底部的「新增」步骤 */
  async clickAddStep(stepIndex = 0): Promise<void> {
    const card = this.getActionStepCards().nth(stepIndex);
    await card.locator('[data-testid="add-step-btn"]').click();
    await sleep(300);
  }

  /** 读取动作卡片标题文本（如「动作 1」） */
  async getActionStepTitle(stepIndex: number): Promise<string> {
    // 限定动作卡片自身的 header（断言卡片是动作卡片的子元素，也有 .n-card-header__main，
    // 用 > 直属子选择器跳过 .n-card__content 内的断言卡片标题，避免 strict mode 命中多个）
    const card = this.getActionStepCards().nth(stepIndex);
    const header = card.locator(':scope > .n-card-header .n-card-header__main');
    return (await header.textContent())?.trim() ?? '';
  }

  /** 读取断言卡片标题文本（如「断言 1」） */
  async getAssertionTitle(stepIndex: number, assetIndex: number): Promise<string> {
    const header = this.getAssertionCards(stepIndex).nth(assetIndex).locator('.n-card-header__main');
    return (await header.textContent())?.trim() ?? '';
  }

  // ==================== 下拉分组 / 悬浮提示回归 ====================

  /** 第 stepIndex 个动作的「使用卡牌」下拉（步骤 Tab；PlayCard/DisCard/OptRoomAction(确认)/UseHeroSkill 时出现） */
  getStepCardsSelect(stepIndex: number): Locator {
    return this.getActionStepCards().nth(stepIndex).locator('[data-testid="step-cards-select"]');
  }

  /** 「应用智能描述」按钮（动作卡片 header-extra 内，外包 n-tooltip，内容=aiDesc(step)） */
  getApplyAiDescButton(stepIndex: number): Locator {
    return this.getActionStepHeaderExtra(stepIndex).locator('[data-testid="apply-ai-desc-btn"]');
  }

  /** 第 heroIndex 个武将配置卡片（演武面板 / 配置 Tab） */
  getHeroConfigCard(heroIndex: number): Locator {
    return this.page.locator('[data-testid="hero-config-card"]').nth(heroIndex);
  }

  /** 武将的「删除技能」下拉（option hover 文字区显示技能描述 tooltip） */
  getDelSkillsSelect(heroIndex: number): Locator {
    return this.getHeroConfigCard(heroIndex).locator('[data-testid="del-skills-select"]');
  }

  /** 武将的「增加技能」下拉 */
  getAddSkillsSelect(heroIndex: number): Locator {
    return this.getHeroConfigCard(heroIndex).locator('[data-testid="add-skills-select"]');
  }

  /** 当前可见 select 菜单内的分组标题（卡牌下拉的"分割线"——naive-ui group header） */
  getVisibleSelectGroupHeaders(): Locator {
    return this.page.locator('.n-base-select-menu:visible .n-base-select-group-header');
  }

  /**
   * 当前可见 select 菜单内第 optionIndex 个选项的「文字区」。
   *
   * 技能描述 tooltip 的 NTooltip trigger 仅包裹选项文字 span，**必须 hover 文字区才会触发**，
   * hover 选项空白/checkmark 区不触发。回归测试用此 locator hover。
   */
  getVisibleOptionContent(optionIndex = 0): Locator {
    return this.page.locator('.n-base-select-option:visible').nth(optionIndex)
      .locator('.n-base-select-option__content');
  }

  /** 当前可见的 tooltip 浮层（naive-ui n-tooltip teleport 到 body，class 含 n-tooltip） */
  getVisibleTooltip(): Locator {
    return this.page.locator('.n-tooltip:visible').last();
  }

  /**
   * 点开指定 select 触发器并等待下拉菜单可见（naive-ui 菜单 teleport 到 body，统一 :visible 等待）。
   *
   * 替代散落在各 spec 里的 `trigger.click() + expect(.n-base-select-menu:visible).toBeVisible` 模式，
   * 统一行为与超时，便于后续维护。
   *
   * 用 .first()：避免前一个测试残留的不可见菜单被 :visible 误匹配（naive-ui 关闭后 menu DOM 残留），
   * 或本测试中多个 select 同时打开时 strict mode 崩溃（资产断言 test 里 pickAssertionType 后紧跟卡下拉打开）。
   */
  async openSelectAndWait(trigger: Locator): Promise<void> {
    await trigger.click();
    await expect(this.page.locator('.n-base-select-menu:visible').first())
      .toBeVisible({timeout: 5000});
  }

  // ==================== 资产断言下拉分组回归 ====================

  /** 指定断言卡片的「断言类型」下拉（asset-card.vue 内 .assetTypeSelect）。
   *  通过 data-testid="asset-type-row" 定位行，再取其中唯一的 n-select。 */
  getAssertionTypeSelect(stepIndex: number, assetIndex: number): Locator {
    return this.getAssertionTypeRow(stepIndex, assetIndex).locator('.n-select').first();
  }

  /** 资产断言的「卡」下拉（AssetNormalCardSection 中 Cards 字段）。
   *  定位策略：断言卡片内 .normalCardTypeAsset 容器最后一个子 div 的第一个 .n-select（=「卡」）。
   *  「不期望卡」是同 div 的第二个 .n-select，故用 first() 精确取到「卡」。
   *  不依赖 data-testid，避免运行中 wails3 dev 未 HMR 到 testid 改动时定位失败。 */
  getAssetCardsSelect(stepIndex: number, assetIndex: number): Locator {
    return this.getAssertionCards(stepIndex)
      .nth(assetIndex)
      .locator('.normalCardTypeAsset')
      .locator(':scope > :last-child .n-select')
      .first();
  }

  /**
   * 在指定断言卡片上选择断言类型（按可见下拉选项的文本匹配，如 "出牌(PlayCard)"）。
   *
   * naive-ui 的 filterable n-select 打开后选项 teleport 到 body 的 .n-base-select-menu，
   * 关闭后仍残留（不可见）；必须用 :visible 过滤残留，且只点当前打开菜单内的选项。
   *
   * 该 select 为 filterable，选项/虚拟列表偶发加载延迟，单次点击偶发"点了但未选中"
   * （菜单未关闭、值未更新），因此用「重试 + 校验 trigger 已变更」保证可靠性。
   *
   * 校验时机坑（曾导致确定性失败）：filterable select 点选项后**菜单保持打开**（不自动收起），
   * 此时 trigger 处于"搜索态"，textContent 可能不含完整 label → picked 永远 false。
   * 故每次点击后必须先关闭菜单（点空白），再在 trigger 稳态下校验文本。
   */
  async pickAssertionType(stepIndex: number, assetIndex: number, optionText: string): Promise<void> {
    const sel = this.getAssertionTypeSelect(stepIndex, assetIndex);
    await sel.scrollIntoViewIfNeeded();

    // 重试打开+点选，直到 trigger 显示目标文本（确认选中）
    let picked = false;
    for (let i = 0; i < 5 && !picked; i++) {
      // 若当前没有可见菜单，先打开
      if (await this.page.locator('.n-base-select-menu:visible').count() === 0) {
        await sel.click();
        await expect(this.page.locator('.n-base-select-menu:visible').first())
          .toBeVisible({timeout: 5000});
      }
      // 等目标选项可见（虚拟列表/异步加载兜底）
      const opt = this.page
        .locator('.n-base-select-menu:visible')
        .locator('.n-base-select-option:visible')
        .filter({hasText: optionText})
        .first();
      const visible = await opt.waitFor({state: 'visible', timeout: 3000})
        .then(() => true).catch(() => false);
      if (!visible) {
        // 选项可能在虚拟列表下方，滚动菜单重试
        await this.page.locator('.n-base-select-menu:visible').first()
          .evaluate((el) => el.scrollBy(0, 400)).catch(() => {});
        await sleep(250);
        continue;
      }
      await opt.click();
      await sleep(450);
      // 关键：先关闭菜单（点空白收起 filterable select），让 trigger 回到稳态再校验，
      // 否则菜单打开时 trigger 处于搜索态，textContent 不含完整 label → 误判未选中
      if (await this.page.locator('.n-base-select-menu:visible').count() > 0) {
        await this.page.locator('body').click({position: {x: 1, y: 1}}).catch(() => {});
        await sleep(300);
      }
      // 校验：trigger 内已显示目标文本 → 视为选中成功（菜单已关，trigger 为稳态选中值）
      picked = await sel.filter({hasText: optionText}).count() > 0;
    }
    expect(picked).toBe(true);

    // 兜底关闭可能残留的菜单（点击页面空白处收起 naive-ui select）
    if (await this.page.locator('.n-base-select-menu:visible').count() > 0) {
      await this.page.locator('body').click({position: {x: 1, y: 1}}).catch(() => {});
      await sleep(300);
    }
  }
}
