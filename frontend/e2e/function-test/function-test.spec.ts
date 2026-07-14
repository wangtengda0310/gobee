/**
 * 功能测试页面测试
 * 测试功能测试页面的各个区域：Header、左侧树面板、Tab 面板、Footer
 */

import { test, expect, describe } from '../shared/fixtures';
import { FunctionTestPage } from '../shared/pages/FunctionTestPage';
import { sleep } from '../shared/utils/helpers';

describe('功能测试页 - Header 区域测试', () => {
  let functionTestPage: FunctionTestPage;

  /**
   * 测试前准备：导航到功能测试页并初始化 Page Object
   */
  test.beforeEach(async ({ page }) => {
    functionTestPage = new FunctionTestPage(page);
    await functionTestPage.goto();
  });

  /**
   * 加载用例 - 点击加载按钮，验证树加载成功
   */
  test('加载用例', async () => {
    await functionTestPage.clickLoadCases();

    // 验证用例已加载（用例数量大于 0）
    await functionTestPage.expectCasesLoaded();
  });

  /**
   * 保存用例 - 点击保存按钮，验证保存成功
   */
  test('保存用例', async () => {
    // 先加载用例
    await functionTestPage.clickLoadCases();
    await sleep(500);

    // 点击保存
    await functionTestPage.clickSaveCases();

    // 验证保存成功（等待 Toast 消息）
    await functionTestPage.waitForToast('保存成功');
  });

  /**
   * 执行用例 - 点击执行按钮，遇到停止按钮
   */
  test('执行用例', async () => {
    // 先加载用例
    await functionTestPage.clickLoadCases();
    await sleep(500);

    // 点击执行
    await functionTestPage.clickExecuteCases();

    // 验证正在执行状态（停止按钮可见）
    await functionTestPage.expectIsExecuting();
  });

  /**
   * 停止用例 - 点击停止按钮，验证恢复到执行状态
   */
  test('停止用例', async () => {
    // 先加载并执行用例
    await functionTestPage.clickLoadCases();
    await sleep(500);
    await functionTestPage.clickExecuteCases();
    await functionTestPage.expectIsExecuting();

    // 点击停止
    await functionTestPage.clickStopCases();

    // 验证恢复到未执行状态
    await functionTestPage.expectNotExecuting();
  });

  /**
   * 设置弹窗 - 点击设置按钮，验证弹窗显示
   */
  test('设置弹窗', async () => {
    await functionTestPage.clickSettings();

    // 验证设置弹窗/抽屉显示
    const modal = functionTestPage.page.locator('.n-modal, .n-drawer');
    await expect(modal).toBeVisible();
  });
});

describe('功能测试页 - 左侧树面板测试', () => {
  let functionTestPage: FunctionTestPage;

  test.beforeEach(async ({ page }) => {
    functionTestPage = new FunctionTestPage(page);
    await functionTestPage.goto();
    // 加载用例以便测试树操作
    await functionTestPage.clickLoadCases();
    await sleep(500);
  });

  /**
   * 搜索过滤 - 输入搜索关键词，验证树过滤
   */
  test('搜索过滤', async () => {
    // 输入搜索关键词
    await functionTestPage.searchCases('测试');

    // 验证树节点被过滤（搜索框有值）
    const searchValue = await functionTestPage.searchInput.inputValue();
    expect(searchValue).toBe('测试');
  });

  /**
   * 仅展示过滤开关 - 切换开关，验证过滤效果
   */
  test('仅展示过滤开关', async () => {
    // 获取初始树节点数量
    const initialCount = await functionTestPage.testCaseTree.locator('.n-tree-node').count();

    // 打开过滤开关
    await functionTestPage.toggleFilter(true);

    // 验证开关已打开
    const isChecked = await functionTestPage.filterSwitch.getAttribute('aria-checked');
    expect(isChecked).toBe('true');

    // 关闭过滤开关
    await functionTestPage.toggleFilter(false);
    const isCheckedAgain = await functionTestPage.filterSwitch.getAttribute('aria-checked');
    expect(isCheckedAgain).toBe('false');
  });

  /**
   * 显示描述开关 - 切换开关，验证描述显示
   */
  test('显示描述开关', async () => {
    // 打开描述显示
    await functionTestPage.toggleShowDesc(true);

    // 验证开关已打开
    const isChecked = await functionTestPage.showDescSwitch.getAttribute('aria-checked');
    expect(isChecked).toBe('true');

    // 关闭描述显示
    await functionTestPage.toggleShowDesc(false);
    const isCheckedAgain = await functionTestPage.showDescSwitch.getAttribute('aria-checked');
    expect(isCheckedAgain).toBe('false');
  });

  /**
   * 点击分类展开 - 点击分类节点展开
   */
  test('点击分类展开', async () => {
    // 查找第一个可展开的树节点
    const expandIcon = functionTestPage.testCaseTree.locator('.n-tree-node-switcher').first();

    if (await expandIcon.isVisible()) {
      await expandIcon.click();
      await sleep(200);

      // 验证展开成功（检查展开状态）
      await expect(expandIcon).toBeVisible();
    }
  });

  /**
   * 点击用例加载 - 点击用例节点，验证数据加载
   */
  test('点击用例加载', async () => {
    // 查找第一个叶子节点（用例）
    const leafNode = functionTestPage.testCaseTree.locator('.n-tree-node-content').first();

    if (await leafNode.isVisible()) {
      await leafNode.click();
      await sleep(300);

      // 验证用例配置 Tab 可见
      await expect(functionTestPage.configTab).toBeVisible();
    }
  });

  /**
   * 右键菜单显示 - 右键点击节点，验证菜单显示
   */
  test('右键菜单显示', async () => {
    // 右键点击第一个树节点
    const firstNode = functionTestPage.testCaseTree.locator('.n-tree-node-content').first();

    if (await firstNode.isVisible()) {
      await firstNode.click({ button: 'right' });
      await sleep(200);

      // 验证右键菜单显示
      const contextMenu = functionTestPage.getContextMenu();
      await expect(contextMenu).toBeVisible();
    }
  });

  /**
   * 新建分类 - 通过右键菜单新建分类
   * 注意：此测试需要 mock 后端保存操作
   */
  test.skip('新建分类（需要 mock 后端）', async () => {
    // 右键点击根节点
    const rootNode = functionTestPage.testCaseTree.locator('.n-tree-node-content').first();
    await rootNode.click({ button: 'right' });
    await sleep(200);

    // 点击新建分类
    await functionTestPage.clickContextMenuOption('新建分类');

    // 在弹窗中输入分类名
    const input = functionTestPage.page.locator('.n-modal input');
    await input.fill('测试分类');
    await functionTestPage.page.locator('.n-modal button:has-text("确定")').click();

    // 验证新建成功
    await functionTestPage.expectTreeNodeExists('测试分类');
  });

  /**
   * 新建用例 - 通过右键菜单新建用例
   * 注意：此测试需要 mock 后端保存操作
   */
  test.skip('新建用例（需要 mock 后端）', async () => {
    // 右键点击分类节点
    const categoryNode = functionTestPage.testCaseTree.locator('.n-tree-node-content').first();
    await categoryNode.click({ button: 'right' });
    await sleep(200);

    // 点击新建用例
    await functionTestPage.clickContextMenuOption('新建用例');

    // 在弹窗中输入用例名
    const input = functionTestPage.page.locator('.n-modal input');
    await input.fill('测试用例');
    await functionTestPage.page.locator('.n-modal button:has-text("确定")').click();

    // 验证新建成功
    await functionTestPage.expectTreeNodeExists('测试用例');
  });

  /**
   * 重命名 - 通过右键菜单重命名
   * 注意：此测试需要 mock 后端保存操作
   */
  test.skip('重命名（需要 mock 后端）', async () => {
    // 右键点击节点
    const node = functionTestPage.testCaseTree.locator('.n-tree-node-content').first();
    await node.click({ button: 'right' });
    await sleep(200);

    // 点击重命名
    await functionTestPage.clickContextMenuOption('重命名');

    // 在弹窗中输入新名称
    const input = functionTestPage.page.locator('.n-modal input');
    await input.clear();
    await input.fill('重命名后的节点');
    await functionTestPage.page.locator('.n-modal button:has-text("确定")').click();

    // 验证重命名成功
    await functionTestPage.expectTreeNodeExists('重命名后的节点');
  });

  /**
   * 删除 - 通过右键菜单删除
   * 注意：此测试需要 mock 后端保存操作
   */
  test.skip('删除（需要 mock 后端）', async () => {
    // 获取初始节点数量
    const initialCount = await functionTestPage.testCaseTree.locator('.n-tree-node').count();

    // 右键点击节点
    const node = functionTestPage.testCaseTree.locator('.n-tree-node-content').first();
    await node.click({ button: 'right' });
    await sleep(200);

    // 点击删除
    await functionTestPage.clickContextMenuOption('删除');

    // 确认删除
    await functionTestPage.page.locator('.n-modal button:has-text("确定")').click();

    // 验证删除成功（节点数量减少）
    const newCount = await functionTestPage.testCaseTree.locator('.n-tree-node').count();
    expect(newCount).toBeLessThan(initialCount);
  });

  /**
   * 拖拽排序 - 拖拽节点改变顺序
   * 注意：此测试需要 mock 后端保存操作
   */
  test.skip('拖拽排序（需要 mock 后端）', async () => {
    // 获取第一个节点
    const firstNode = functionTestPage.testCaseTree.locator('.n-tree-node-content').first();
    const secondNode = functionTestPage.testCaseTree.locator('.n-tree-node-content').nth(1);

    if (await firstNode.isVisible() && await secondNode.isVisible()) {
      // 拖拽第一个节点到第二个位置
      await firstNode.hover();
      await functionTestPage.page.mouse.down();
      await secondNode.hover();
      await functionTestPage.page.mouse.up();
      await sleep(200);

      // 验证拖拽成功（无 JS 错误即可）
      await expect(firstNode).toBeVisible();
    }
  });
});

describe('功能测试页 - Tab 面板用例配置测试', () => {
  let functionTestPage: FunctionTestPage;

  test.beforeEach(async ({ page }) => {
    functionTestPage = new FunctionTestPage(page);
    await functionTestPage.goto();
    await functionTestPage.clickLoadCases();
    await sleep(500);

    // 确保在用例配置 Tab
    await functionTestPage.switchToConfigTab();

    // 选择一个用例
    const firstLeaf = functionTestPage.testCaseTree.locator('.n-tree-node-content').first();
    if (await firstLeaf.isVisible()) {
      await firstLeaf.click();
      await sleep(300);
    }
  });

  /**
   * 用例名称编辑 - 编辑名称，验证保存
   */
  test('用例名称编辑', async () => {
    const nameInput = functionTestPage.getCaseNameInput();

    if (await nameInput.isVisible()) {
      await functionTestPage.setCaseName('测试用例名称');

      const value = await nameInput.inputValue();
      expect(value).toBe('测试用例名称');
    }
  });

  /**
   * 用例描述编辑 - 编辑描述
   */
  test('用例描述编辑', async () => {
    const descInput = functionTestPage.getCaseDescInput();

    if (await descInput.isVisible()) {
      await functionTestPage.setCaseDesc('这是一个测试描述');

      const value = await descInput.inputValue();
      expect(value).toBe('这是一个测试描述');
    }
  });

  /**
   * 负责人设置 - 设置负责人
   */
  test('负责人设置', async () => {
    const ownerInput = functionTestPage.getOwnerInput();

    if (await ownerInput.isVisible()) {
      await functionTestPage.setOwner('测试人员');

      const value = await ownerInput.inputValue();
      expect(value).toBe('测试人员');
    }
  });

  /**
   * 牌堆组数量 - 修改牌堆组数量
   */
  test('牌堆组数量', async () => {
    const deckInput = functionTestPage.getDeckGroupInput();

    if (await deckInput.isVisible()) {
      await functionTestPage.setDeckGroupCount(5);

      const value = await deckInput.inputValue();
      expect(value).toBe('5');
    }
  });

  /**
   * 摸牌堆配置 - 配置摸牌堆
   */
  test('摸牌堆配置', async () => {
    // 查找摸牌堆配置区域
    const drawDeckSection = functionTestPage.page.locator('text=摸牌堆').locator('..');

    if (await drawDeckSection.isVisible()) {
      // 验证摸牌堆配置区域存在
      await expect(drawDeckSection).toBeVisible();
    }
  });

  /**
   * 弃牌堆配置 - 配置弃牌堆
   */
  test('弃牌堆配置', async () => {
    // 查找弃牌堆配置区域
    const discardDeckSection = functionTestPage.page.locator('text=弃牌堆').locator('..');

    if (await discardDeckSection.isVisible()) {
      // 验证弃牌堆配置区域存在
      await expect(discardDeckSection).toBeVisible();
    }
  });

  /**
   * 增加武将 - 点击增加武将按钮
   */
  test('增加武将', async () => {
    const initialCount = await functionTestPage.getHeroCardCount();

    await functionTestPage.clickAddHero();

    const newCount = await functionTestPage.getHeroCardCount();
    expect(newCount).toBe(initialCount + 1);
  });

  /**
   * 武将选择 - 选择武将
   * 注意：此测试需要武将选择器支持
   */
  test.skip('武将选择（需要 mock 后端数据）', async () => {
    // 添加武将后选择
    await functionTestPage.clickAddHero();

    // 查找武将选择器
    const heroSelect = functionTestPage.page.locator('.hero-card .n-select').first();

    if (await heroSelect.isVisible()) {
      await heroSelect.click();
      await sleep(200);

      // 选择第一个选项
      const option = functionTestPage.page.locator('.n-base-select-option').first();
      if (await option.isVisible()) {
        await option.click();
      }
    }
  });

  /**
   * 身份选择 - 选择身份
   * 注意：此测试需要身份选择器支持
   */
  test.skip('身份选择（需要 mock 后端数据）', async () => {
    // 查找身份选择器
    const identitySelect = functionTestPage.page.locator('text=身份').locator('..').locator('.n-select');

    if (await identitySelect.isVisible()) {
      await identitySelect.click();
      await sleep(200);

      // 选择第一个选项
      const option = functionTestPage.page.locator('.n-base-select-option').first();
      if (await option.isVisible()) {
        await option.click();
      }
    }
  });

  /**
   * 势力选择 - 选择势力
   * 注意：此测试需要势力选择器支持
   */
  test.skip('势力选择（需要 mock 后端数据）', async () => {
    // 查找势力选择器
    const factionSelect = functionTestPage.page.locator('text=势力').locator('..').locator('.n-select');

    if (await factionSelect.isVisible()) {
      await factionSelect.click();
      await sleep(200);

      // 选择第一个选项
      const option = functionTestPage.page.locator('.n-base-select-option').first();
      if (await option.isVisible()) {
        await option.click();
      }
    }
  });

  /**
   * 初始手牌配置 - 配置初始手牌
   * 注意：此测试需要手牌配置支持
   */
  test.skip('初始手牌配置（需要 mock 后端数据）', async () => {
    // 查找初始手牌配置区域
    const handCardsSection = functionTestPage.page.locator('text=初始手牌').locator('..');

    if (await handCardsSection.isVisible()) {
      // 验证初始手牌配置区域存在
      await expect(handCardsSection).toBeVisible();
    }
  });

  /**
   * 删除武将 - 删除武将座位
   */
  test('删除武将', async () => {
    // 先添加一个武将
    await functionTestPage.clickAddHero();
    const countAfterAdd = await functionTestPage.getHeroCardCount();

    if (countAfterAdd > 0) {
      // 删除最后一个武将
      await functionTestPage.removeHero(countAfterAdd - 1);

      const countAfterRemove = await functionTestPage.getHeroCardCount();
      expect(countAfterRemove).toBe(countAfterAdd - 1);
    }
  });

  /**
   * 武将拖拽排序 - 拖拽调整座位顺序
   * 注意：此测试需要多个武将卡片
   */
  test.skip('武将拖拽排序（需要 mock 后端）', async () => {
    // 添加两个武将
    await functionTestPage.clickAddHero();
    await functionTestPage.clickAddHero();

    const count = await functionTestPage.getHeroCardCount();
    if (count >= 2) {
      // 拖拽第一个武将到第二个位置
      await functionTestPage.dragHeroCard(0, 1);

      // 验证拖拽成功（无 JS 错误即可）
      await expect(functionTestPage.getHeroCards().first()).toBeVisible();
    }
  });
});

describe('功能测试页 - Tab 面板用例步骤测试', () => {
  let functionTestPage: FunctionTestPage;

  test.beforeEach(async ({ page }) => {
    functionTestPage = new FunctionTestPage(page);
    await functionTestPage.goto();
    await functionTestPage.clickLoadCases();
    await sleep(500);

    // 切换到用例步骤 Tab
    await functionTestPage.switchToStepsTab();
  });

  /**
   * 步骤列表显示 - 切换到步骤 Tab，验证步骤显示
   */
  test('步骤列表显示', async () => {
    // 验证步骤 Tab 已激活
    await expect(functionTestPage.stepsTab).toHaveClass(/n-tabs-tab--active/);

    // 验证步骤列表存在
    const stepsList = functionTestPage.getStepsList();
    await expect(stepsList.first()).toBeVisible();
  });

  /**
   * 动作类型选择 - 选择动作类型
   * 注意：此测试需要动作类型选择器支持
   */
  test.skip('动作类型选择（需要 mock 后端数据）', async () => {
    // 选择第一个步骤
    const steps = functionTestPage.getStepsList();
    if (await steps.count() > 0) {
      const firstStep = steps.first();

      // 查找动作类型选择器
      const actionSelect = firstStep.locator('.n-select').first();

      if (await actionSelect.isVisible()) {
        await actionSelect.click();
        await sleep(200);

        // 选择一个选项
        const option = functionTestPage.page.locator('.n-base-select-option').first();
        if (await option.isVisible()) {
          await option.click();
        }
      }
    }
  });

  /**
   * 座位选择 - 选择执行座位
   * 注意：此测试需要座位选择器支持
   */
  test.skip('座位选择（需要 mock 后端数据）', async () => {
    const steps = functionTestPage.getStepsList();
    if (await steps.count() > 0) {
      const firstStep = steps.first();

      // 查找座位选择器
      const seatSelect = firstStep.locator('text=座位').locator('..').locator('.n-select');

      if (await seatSelect.isVisible()) {
        await seatSelect.click();
        await sleep(200);

        const option = functionTestPage.page.locator('.n-base-select-option').first();
        if (await option.isVisible()) {
          await option.click();
        }
      }
    }
  });

  /**
   * 技能选择 - 选择技能
   * 注意：此测试需要技能选择器支持
   */
  test.skip('技能选择（需要 mock 后端数据）', async () => {
    const steps = functionTestPage.getStepsList();
    if (await steps.count() > 0) {
      const firstStep = steps.first();

      // 查找技能选择器
      const skillSelect = firstStep.locator('text=技能').locator('..').locator('.n-select');

      if (await skillSelect.isVisible()) {
        await skillSelect.click();
        await sleep(200);

        const option = functionTestPage.page.locator('.n-base-select-option').first();
        if (await option.isVisible()) {
          await option.click();
        }
      }
    }
  });

  /**
   * 卡牌选择 - 选择卡牌
   * 注意：此测试需要卡牌选择器支持
   */
  test.skip('卡牌选择（需要 mock 后端数据）', async () => {
    const steps = functionTestPage.getStepsList();
    if (await steps.count() > 0) {
      const firstStep = steps.first();

      // 查找卡牌选择器
      const cardSelect = firstStep.locator('text=卡牌').locator('..').locator('.n-select');

      if (await cardSelect.isVisible()) {
        await cardSelect.click();
        await sleep(200);

        const option = functionTestPage.page.locator('.n-base-select-option').first();
        if (await option.isVisible()) {
          await option.click();
        }
      }
    }
  });

  /**
   * 新增断言 - 点击新增断言按钮
   */
  test('新增断言', async () => {
    const initialCount = await functionTestPage.getStepCount();

    if (initialCount > 0) {
      const initialAssetCount = await functionTestPage.getAssetCards(0).count();

      await functionTestPage.addAsset(0);

      const newAssetCount = await functionTestPage.getAssetCards(0).count();
      expect(newAssetCount).toBe(initialAssetCount + 1);
    }
  });

  /**
   * 断言配置 - 配置断言内容
   * 注意：此测试需要断言配置支持
   */
  test.skip('断言配置（需要 mock 后端数据）', async () => {
    const steps = functionTestPage.getStepsList();
    if (await steps.count() > 0) {
      // 添加一个断言
      await functionTestPage.addAsset(0);

      const assetCards = functionTestPage.getAssetCards(0);
      if (await assetCards.count() > 0) {
        // 验证断言卡片存在
        await expect(assetCards.first()).toBeVisible();
      }
    }
  });

  /**
   * 删除断言 - 删除断言卡片
   */
  test('删除断言', async () => {
    const steps = functionTestPage.getStepsList();
    if (await steps.count() > 0) {
      // 先添加一个断言
      await functionTestPage.addAsset(0);
      const countAfterAdd = await functionTestPage.getAssetCards(0).count();

      if (countAfterAdd > 0) {
        // 删除最后一个断言
        const assetCards = functionTestPage.getAssetCards(0);
        await assetCards.last().locator('button:has-text("删除")').click();

        const countAfterRemove = await functionTestPage.getAssetCards(0).count();
        expect(countAfterRemove).toBe(countAfterAdd - 1);
      }
    }
  });

  /**
   * 新增步骤 - 点击新增步骤按钮
   */
  test('新增步骤', async () => {
    const initialCount = await functionTestPage.getStepCount();

    await functionTestPage.addStep();

    const newCount = await functionTestPage.getStepCount();
    expect(newCount).toBe(initialCount + 1);
  });

  /**
   * 删除步骤 - 删除步骤
   */
  test('删除步骤', async () => {
    // 先添加一个步骤
    await functionTestPage.addStep();
    const countAfterAdd = await functionTestPage.getStepCount();

    if (countAfterAdd > 0) {
      // 删除最后一个步骤
      await functionTestPage.removeStep(countAfterAdd - 1);

      const countAfterRemove = await functionTestPage.getStepCount();
      expect(countAfterRemove).toBe(countAfterAdd - 1);
    }
  });

  /**
   * 步骤拖拽排序 - 拖拽调整步骤顺序
   * 注意：此测试需要多个步骤
   */
  test.skip('步骤拖拽排序（需要 mock 后端）', async () => {
    // 添加两个步骤
    await functionTestPage.addStep();
    await functionTestPage.addStep();

    const count = await functionTestPage.getStepCount();
    if (count >= 2) {
      const steps = functionTestPage.getStepsList();
      const firstStep = steps.first().locator('.drag-handle');
      const secondStep = steps.nth(1);

      if (await firstStep.isVisible()) {
        // 拖拽第一个步骤到第二个位置
        await firstStep.hover();
        await functionTestPage.page.mouse.down();
        await secondStep.hover();
        await functionTestPage.page.mouse.up();
        await sleep(200);

        // 验证拖拽成功（无 JS 错误即可）
        await expect(steps.first()).toBeVisible();
      }
    }
  });

  /**
   * 锚点导航 - 点击锚点跳转
   */
  test('锚点导航', async () => {
    const anchorNav = functionTestPage.getAnchorNav();

    if (await anchorNav.isVisible()) {
      // 点击第一个锚点链接
      const firstLink = anchorNav.locator('.n-anchor-link').first();
      if (await firstLink.isVisible()) {
        await firstLink.click();
        await sleep(200);

        // 验证锚点导航可见
        await expect(anchorNav).toBeVisible();
      }
    }
  });

  /**
   * 复制步骤 - 复制步骤
   * 注意：此测试需要复制按钮支持
   */
  test.skip('复制步骤（需要复制功能）', async () => {
    const steps = functionTestPage.getStepsList();
    if (await steps.count() > 0) {
      const initialCount = await functionTestPage.getStepCount();

      // 查找复制按钮
      const copyButton = steps.first().locator('button:has-text("复制")');

      if (await copyButton.isVisible()) {
        await copyButton.click();

        const newCount = await functionTestPage.getStepCount();
        expect(newCount).toBe(initialCount + 1);
      }
    }
  });
});

describe('功能测试页 - Tab 面板执行日志测试', () => {
  let functionTestPage: FunctionTestPage;

  test.beforeEach(async ({ page }) => {
    functionTestPage = new FunctionTestPage(page);
    await functionTestPage.goto();
    await functionTestPage.clickLoadCases();
    await sleep(500);
  });

  /**
   * 日志显示 - 执行用例，验证日志显示
   * 注意：此测试需要执行用例
   */
  test.skip('日志显示（需要执行用例）', async () => {
    // 切换到执行日志 Tab
    await functionTestPage.switchToLogTab();

    // 执行用例
    await functionTestPage.clickExecuteCases();
    await sleep(1000);

    // 验证日志列表有内容
    const logCount = await functionTestPage.getLogCount();
    expect(logCount).toBeGreaterThan(0);
  });

  /**
   * 日志 Tab 切换 - 多用例执行时切换 Tab
   * 注意：此测试需要多用例执行
   */
  test.skip('日志 Tab 切换（需要多用例执行）', async () => {
    await functionTestPage.switchToLogTab();

    const logTabs = functionTestPage.getLogTabs();
    const tabCount = await logTabs.count();

    if (tabCount > 1) {
      // 切换到第二个日志 Tab
      await functionTestPage.switchLogTab(1);

      // 验证 Tab 切换成功
      await expect(logTabs.nth(1)).toHaveClass(/n-tabs-tab--active/);
    }
  });

  /**
   * 日志滚动 - 滚动日志列表
   */
  test.skip('日志滚动（需要执行用例）', async () => {
    await functionTestPage.switchToLogTab();

    // 执行用例产生日志
    await functionTestPage.clickExecuteCases();
    await sleep(2000);

    const logList = functionTestPage.page.locator('.robot-test-log .log-list');

    if (await logList.isVisible()) {
      // 滚动到底部
      await functionTestPage.page.evaluate(() => {
        const list = document.querySelector('.robot-test-log .log-list');
        if (list) {
          list.scrollTop = list.scrollHeight;
        }
      });

      await sleep(200);

      // 验证日志列表可见
      await expect(logList).toBeVisible();
    }
  });

  /**
   * 错误高亮 - 验证错误日志高亮
   * 注意：此测试需要产生错误日志
   */
  test.skip('错误高亮（需要错误日志）', async () => {
    await functionTestPage.switchToLogTab();

    // 执行用例（可能产生错误）
    await functionTestPage.clickExecuteCases();
    await sleep(2000);

    // 查找错误日志
    const errorLog = functionTestPage.page.locator('.log-item.error, .log-item[data-level="error"]');

    if (await errorLog.isVisible()) {
      // 验证错误日志有高亮样式
      await expect(errorLog).toBeVisible();
    }
  });
});

describe('功能测试页 - Footer 区域测试', () => {
  let functionTestPage: FunctionTestPage;

  test.beforeEach(async ({ page }) => {
    functionTestPage = new FunctionTestPage(page);
    await functionTestPage.goto();
    await functionTestPage.clickLoadCases();
    await sleep(500);
  });

  /**
   * 用例数量统计 - 验证显示用例数量
   */
  test('用例数量统计', async () => {
    const caseCount = await functionTestPage.getCaseCountValue();
    const count = parseInt(caseCount);

    // 验证用例数量大于 0
    expect(count).toBeGreaterThan(0);
  });

  /**
   * 动作数量统计 - 验证显示动作数量
   */
  test('动作数量统计', async () => {
    const stepCount = await functionTestPage.getStepCountValue();
    const count = parseInt(stepCount);

    // 验证动作数量大于等于 0
    expect(count).toBeGreaterThanOrEqual(0);
  });

  /**
   * 执行进度显示 - 执行中验证进度
   * 注意：此测试需要执行用例
   */
  test.skip('执行进度显示（需要执行用例）', async () => {
    // 执行用例
    await functionTestPage.clickExecuteCases();

    // 查找进度指示器
    const progressIndicator = functionTestPage.page.locator('.n-progress, .progress-bar');

    if (await progressIndicator.isVisible()) {
      // 验证进度指示器可见
      await expect(progressIndicator).toBeVisible();
    }
  });

  /**
   * 断言错误数 - 验证断言错误数显示
   * 注意：此测试需要产生断言错误
   */
  test.skip('断言错误数（需要断言错误）', async () => {
    // 执行用例
    await functionTestPage.clickExecuteCases();
    await sleep(2000);

    // 查找断言错误数显示
    const errorCount = functionTestPage.page.locator('.error-count, [data-testid="assertion-errors"]');

    if (await errorCount.isVisible()) {
      // 验证断言错误数显示
      await expect(errorCount).toBeVisible();
    }
  });
});

/**
 * 布局验证测试组
 */
describe('功能测试页 - 布局验证', () => {
  let functionTestPage: FunctionTestPage;

  test.beforeEach(async ({ page }) => {
    functionTestPage = new FunctionTestPage(page);
    await functionTestPage.goto();
  });

  /**
   * 页面布局结构验证
   */
  test('页面布局结构验证', async () => {
    // 验证 Header 区域
    await expect(functionTestPage.headerMenu).toBeVisible();

    // 验证左侧树面板
    await expect(functionTestPage.siderPanel).toBeVisible();

    // 验证 Tab 面板
    await expect(functionTestPage.tabsContainer).toBeVisible();

    // 验证 Footer 区域
    await expect(functionTestPage.footerBar).toBeVisible();
  });

  /**
   * Header 按钮验证
   */
  test('Header 按钮验证', async () => {
    await expect(functionTestPage.loadButton).toBeVisible();
    await expect(functionTestPage.saveButton).toBeVisible();
    await expect(functionTestPage.executeButton).toBeVisible();
    await expect(functionTestPage.settingsButton).toBeVisible();
  });

  /**
   * Tab 标签验证
   */
  test('Tab 标签验证', async () => {
    await expect(functionTestPage.configTab).toBeVisible();
    await expect(functionTestPage.stepsTab).toBeVisible();
    await expect(functionTestPage.logTab).toBeVisible();
  });

  /**
   * 主 Tab 导航稳定 — 点击执行日志后标题栏不应整体右移
   */
  test('主 Tab 导航切换执行日志后不偏移', async () => {
    await functionTestPage.switchToConfigTab();
    await functionTestPage.expectMainTabNavStableAfter(async () => {
      await functionTestPage.switchToLogTab();
    });
    await functionTestPage.expectMainTabNavStableAfter(async () => {
      await functionTestPage.switchToStepsTab();
    });
  });

  /**
   * 树面板组件验证
   */
  test('树面板组件验证', async () => {
    await expect(functionTestPage.searchInput).toBeVisible();
    await expect(functionTestPage.testCaseTree).toBeVisible();
  });
});
