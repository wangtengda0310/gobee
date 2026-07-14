/**
 * 协议重放页 Page Object
 */
import { Page, Locator, expect } from '@playwright/test';
import { BasePage } from './BasePage';
import { sleep } from '../utils/helpers';

/**
 * goto() 可选参数
 */
export interface GotoOptions {
  /** 跳过清空重放结果（默认 false，即默认执行清空） */
  skipClearResults?: boolean;
}

export class ProtoTestPage extends BasePage {
  // 页签
  readonly tabPacket: Locator;
  readonly tabTestcase: Locator;

  // 目标服务
  readonly replayServerInput: Locator;
  readonly replayHttpInput: Locator;
  readonly replayOpenIDInput: Locator;
  readonly maxConcurrencyInput: Locator;
  readonly settingsButton: Locator;

  // 账号序号范围
  readonly rangeStartInput: Locator;
  readonly rangeEndInput: Locator;

  // 录制按钮
  readonly startRecordButton: Locator;
  readonly stopRecordButton: Locator;

  // 顶部重放/多选按钮
  readonly startReplayButton: Locator;
  readonly multiSelectButton: Locator;

  // 消息表格
  readonly messageTable: Locator;

  // 重放面板
  readonly replayPanel: Locator;
  readonly replayRetryButton: Locator;
  readonly replayCountInput: Locator;
  readonly replayStopButton: Locator;
  readonly ntfDisplayArea: Locator;
  readonly replayStatusTag: Locator;

  // 多选
  readonly batchCancelButton: Locator;
  readonly batchCountText: Locator;

  // 测试用例
  readonly loadCaseButton: Locator;
  readonly executeCaseButton: Locator;
  readonly newCaseButton: Locator;
  readonly deleteCaseButton: Locator;
  readonly caseSelect: Locator;

  // 文件信息 / 录制进度
  readonly fileInfoRow: Locator;
  readonly recordProgressTag: Locator;

  // 卡片编辑器
  readonly reqCardEditor: Locator;
  readonly formatButton: Locator;
  readonly applyButton: Locator;
  /** 测试用例页签：步骤描述输入框 */
  readonly caseDescriptInput: Locator;
  readonly fieldCardContainer: Locator;

  // 新输入组件选择器
  readonly rangeInputLabel: Locator;
  readonly rangeInputStart: Locator;
  readonly rangeInputStep: Locator;
  readonly rangeInputEnd: Locator;
  readonly enumSelectLabel: Locator;
  readonly enumSelectInput: Locator;
  readonly comboSelectLabel: Locator;
  readonly comboSelectInput: Locator;

  // 重放结果页签
  readonly tabReplayResult: Locator;
  readonly replayResultTable: Locator;
  readonly sourceLabel: Locator;
  readonly resultSelector: Locator;

  constructor(page: Page) {
    super(page);

    // 页签（使用精确文本匹配，避免匹配到容器元素）
    this.tabPacket = page.locator('div').filter({ hasText: /^发包改包$/ }).first();
    this.tabTestcase = page.locator('div').filter({ hasText: /^测试用例$/ }).first();

    // 目标服务（两个页签各有输入框，使用 first() 选择发包改包页签的输入框）
    this.replayServerInput = page.locator('input[placeholder*="TCP"]').first();
    this.replayHttpInput = page.locator('input[placeholder*="HTTP"]').first();
    this.replayOpenIDInput = page.locator('input[placeholder*="登录"]').first();

    // 设置按钮（target-service-config.vue 中打开重放设置抽屉的按钮）
    // 用 data-testid 定位，避免与 header 全局"设置"导航按钮冲突
    // （button:visible 设置 .first() 会误命中 header 那个，导致点错进设置页而非打开抽屉）
    this.settingsButton = page.locator('[data-testid="target-service-settings-btn"]').first();

    // 账号序号范围（target-service-config.vue 中的 n-input-number）
    this.rangeStartInput = page.locator('input[placeholder*="起始"]').first();
    this.rangeEndInput = page.locator('input[placeholder*="终止"]').first();

    // 最大并发（target-service-config.vue 中的 n-input-number）
    this.maxConcurrencyInput = page.locator('input[placeholder*="不限"]').first();

    // 录制按钮（两个页签各有 protocol-content，使用 first() 选择默认页签）
    this.startRecordButton = page.locator('button').filter({ hasText: '开始录制' }).first();
    this.stopRecordButton = page.locator('button').filter({ hasText: '停止录制' }).first();

    // 顶部按钮
    this.startReplayButton = page.locator('button').filter({ hasText: '开始重放' }).first();
    this.multiSelectButton = page.locator('button:visible').filter({ hasText: /^多选$/ }).first();

    // 消息表格（可见页签内的表格）
    this.messageTable = page.locator('.n-data-table:visible');

    // 重放面板
    this.replayPanel = page.locator('text=重放控制').first();
    this.replayRetryButton = page.locator('button').filter({ hasText: '重发' }).first();
    this.replayCountInput = page.locator('.n-input-number input').first();
    this.replayStopButton = page.locator('button').filter({ hasText: '停止' }).last();
    this.ntfDisplayArea = page.locator('text=Ntf:');
    this.replayStatusTag = page.locator('text=正在重放').or(page.locator('text=重放完成'));

    // 多选
    this.batchCancelButton = page.locator('button').filter({ hasText: '取消多选' }).first();
    this.batchCountText = page.locator('text=已选择');

    // 测试用例（测试用例页签按钮，切换后使用 first()）
    this.loadCaseButton = page.locator('button').filter({ hasText: '加载用例' }).first();
    this.executeCaseButton = page.locator('button').filter({ hasText: '执行用例' }).first();
    this.newCaseButton = page.locator('button').filter({ hasText: '新增模块' }).first();
    this.deleteCaseButton = page.locator('button').filter({ hasText: '删除模块' }).first();
    this.caseSelect = page.locator('.n-base-selection').first();

    // 文件信息 / 录制进度
    this.fileInfoRow = page.locator('text=版本:').first();
    this.recordProgressTag = page.locator('text=录制进度').first();

    // 卡片编辑器
    this.reqCardEditor = page.locator('text=Payload 字段');
    this.formatButton = page.locator('button').filter({ hasText: '格式化' });
    this.applyButton = page.locator('button').filter({ hasText: '应用' });
    this.caseDescriptInput = page.locator('[data-testid="case-descript-input"]:visible');
    this.fieldCardContainer = page.locator('.n-card');

    // 新输入组件选择器
    this.rangeInputLabel = page.locator('text=起始值:');
    this.rangeInputStart = page.locator('.input-row').filter({ hasText: '起始值:' }).locator('input').first();
    this.rangeInputStep = page.locator('.input-row').filter({ hasText: '步长:' }).locator('input').first();
    this.rangeInputEnd = page.locator('.input-row').filter({ hasText: '终值:' }).locator('input').first();
    this.enumSelectLabel = page.locator('text=枚举值:');
    this.enumSelectInput = page.locator('div').filter({ hasText: '枚举值:' }).locator('.n-base-selection').first();
    this.comboSelectLabel = page.locator('text=组合:');
    this.comboSelectInput = page.locator('div').filter({ hasText: '组合:' }).locator('.n-base-selection').first();

    // 重放结果页签
    this.tabReplayResult = page.locator('div').filter({ hasText: /^重放结果$/ }).first();
    this.replayResultTable = page.locator('.n-data-table');
    this.sourceLabel = page.locator('text=来源:').first();
    this.resultSelector = page.locator('text=重放结果:').locator('..').locator('.n-base-selection').first();
  }

  async goto(options?: GotoOptions): Promise<void> {
    const { skipClearResults = false } = options ?? {};

    // 清理可能的遗留模态框和抽屉遮罩
    // 逐 mask 独立检查（而非 OR 条件），两次 Escape 后兜底点击遮罩关闭
    for (const sel of ['.n-modal-mask', '.n-drawer-mask']) {
      const mask = this.page.locator(sel).first();
      if (await mask.isVisible()) {
        await this.page.keyboard.press('Escape');
        await sleep(300);
        if (await mask.isVisible()) {
          await this.page.keyboard.press('Escape'); // 处理嵌套层级
          await sleep(300);
        }
        if (await mask.isVisible()) {
          await mask.click({ position: { x: 5, y: 5 }, force: true }); // 兜底
          await sleep(200);
        }
      }
    }

    // 导航到协议重放页（Wails 使用 createMemoryHistory，必须点击菜单按钮触发路由）
    await this.page.locator('#layout-header button:has-text("Proto测试")').click();
    // 注：sleep(800) 是务实选择——v-show 使三个页签的 DOM 同时存在，
    // waitForSelector({state:'attached'}) 无法区分当前激活页签，不能作为导航确认信号
    await sleep(800);

    // 清理多选模式（如果已激活且可见）
    const cancelMultiSelectBtn = this.page.locator('button').filter({ hasText: '取消多选' }).first();
    if (await cancelMultiSelectBtn.isVisible()) {
      await cancelMultiSelectBtn.click();
      await sleep(300);
    }

    // 重置到发包改包页签（确保每个测试开始时状态一致）
    await this.clickTabPacket();
    await sleep(300);

    // 清空重放结果历史（避免上一个测试遗留的结果干扰来源判断）
    if (!skipClearResults) {
      await this.clickTabReplayResult();
      // 使用前置断言：清空按钮始终在重放结果页签中，若不可见则页签切换失败
      const clearBtn = this.page.locator('button').filter({ hasText: '清空' }).first();
      await expect(clearBtn).toBeVisible({ timeout: 5000 });
      await clearBtn.click({ force: true });
      await sleep(300);
    }

    // 切回发包改包页签
    await this.clickTabPacket();
    await sleep(300);
  }

  // ==================== 页签操作 ====================
  async clickTabPacket(): Promise<void> {
    await this.tabPacket.click();
    await sleep(300);
  }

  async clickTabTestcase(): Promise<void> {
    await this.tabTestcase.click();
    await sleep(1000);
    // 等待测试用例页签的关键元素出现（允许 hidden — v-show 下元素可能被标记为 hidden）
    try {
      await this.page.waitForSelector('button:has-text("加载用例")', { timeout: 5000, state: 'attached' });
    } catch {
      // fallback: 即使检测失败也继续
    }
  }

  // ==================== 目标服务操作 ====================
  async openSettings(): Promise<void> {
    await this.settingsButton.click();
    await sleep(300);
  }

  async setReplayServerAddr(addr: string): Promise<void> {
    await this.replayServerInput.evaluate((el: HTMLInputElement, val: string) => {
      const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, 'value')!.set!;
      setter.call(el, val);
      el.dispatchEvent(new Event('input', { bubbles: true }));
    }, addr);
    await sleep(200);
  }

  async setReplayHttpAddr(addr: string): Promise<void> {
    await this.replayHttpInput.evaluate((el: HTMLInputElement, val: string) => {
      const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, 'value')!.set!;
      setter.call(el, val);
      el.dispatchEvent(new Event('input', { bubbles: true }));
    }, addr);
    await sleep(200);
  }

  async setReplayOpenID(openID: string): Promise<void> {
    await this.replayOpenIDInput.evaluate((el: HTMLInputElement, val: string) => {
      const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, 'value')!.set!;
      setter.call(el, val);
      el.dispatchEvent(new Event('input', { bubbles: true }));
    }, openID);
    await sleep(200);
  }

  // ==================== 账号序号范围操作 ====================

  /**
   * 检查起始序号输入框是否可见
   */
  async isAccountRangeStartVisible(): Promise<boolean> {
    return await this.rangeStartInput.isVisible();
  }

  /**
   * 检查终止序号输入框是否可见
   */
  async isAccountRangeEndVisible(): Promise<boolean> {
    return await this.rangeEndInput.isVisible();
  }

  /**
   * 设置起始序号
   */
  async setAccountRangeStart(n: number): Promise<void> {
    await this.rangeStartInput.clear();
    await this.rangeStartInput.fill(String(n));
    await sleep(200);
  }

  /**
   * 设置终止序号
   */
  async setAccountRangeEnd(n: number): Promise<void> {
    await this.rangeEndInput.clear();
    await this.rangeEndInput.fill(String(n));
    await sleep(200);
  }

  /**
   * 验证账号范围组件的所有元素可见
   */
  async expectAccountRangeVisible(): Promise<void> {
    await expect(this.rangeStartInput).toBeVisible();
    await expect(this.rangeEndInput).toBeVisible();
  }

  // ==================== 录制操作 ====================
  async clickStartRecord(): Promise<void> {
    await this.startRecordButton.click();
    await sleep(500);
  }

  async clickStopRecord(): Promise<void> {
    await this.stopRecordButton.click();
    await sleep(500);
  }

  // ==================== 多选操作 ====================
  async clickMultiSelect(): Promise<void> {
    await this.multiSelectButton.click();
    await sleep(500);
  }

  async clickCancelSelect(): Promise<void> {
    await this.batchCancelButton.click();
    await sleep(500);
  }

  // ==================== 表格操作 ====================
  getTableRows(): Locator {
    return this.messageTable.locator('tbody tr');
  }

  async getTableRowCount(): Promise<number> {
    return await this.getTableRows().count();
  }

  async clickTableRow(index: number): Promise<void> {
    const rows = this.getTableRows();
    await rows.nth(index).click();
    await sleep(200);
  }

  // ==================== 重放操作 ====================
  async clickRetryReplay(): Promise<void> {
    await this.replayRetryButton.click();
    await sleep(500);
  }

  async setRepeatCount(count: number): Promise<void> {
    // 使用 force: true 因为输入框可能在隐藏的 v-show 页签中
    await this.replayCountInput.click({ force: true });
    await this.replayCountInput.fill('');
    await this.replayCountInput.fill(String(count));
    await sleep(200);
  }

  // ==================== 测试用例操作 ====================
  async clickLoadCase(): Promise<void> { await this.loadCaseButton.dispatchEvent('click'); await sleep(500); }
  async clickExecuteCase(): Promise<void> { await this.executeCaseButton.dispatchEvent('click'); await sleep(500); }
  async clickNewCase(): Promise<void> { await this.newCaseButton.dispatchEvent('click'); await sleep(500); }
  async clickDeleteCase(): Promise<void> { await this.deleteCaseButton.dispatchEvent('click'); await sleep(500); }

  async selectCaseFromDropdown(name: string): Promise<void> {
    await this.caseSelect.click();
    await sleep(300);
    // 直接定位「可见的」目标选项。Naive UI 关闭后的 menu 仍留 DOM（其选项不可见），
    // 用 :visible 过滤残留，避免 strict mode 多元素冲突
    const option = this.page.locator('.n-base-select-option:visible').filter({ hasText: name }).first();
    await option.waitFor({ state: 'visible', timeout: 3000 });
    await option.click();
    await sleep(500);
  }

  // ==================== 验证方法 ====================
  async expectPacketTabVisible(): Promise<void> { await expect(this.tabPacket).toBeVisible(); }
  async expectTestcaseTabVisible(): Promise<void> { await expect(this.tabTestcase).toBeVisible(); }
  async expectTargetServiceVisible(): Promise<void> { await expect(this.replayServerInput).toBeVisible(); }
  async expectRecordButtonsVisible(): Promise<void> {
    await expect(this.startRecordButton).toBeVisible({ timeout: 15000 });
    await expect(this.stopRecordButton).toBeVisible();
  }
  async expectMessageTableVisible(): Promise<void> { await expect(this.messageTable).toBeVisible(); }
  async expectReplayPanelVisible(): Promise<void> { await expect(this.replayPanel).toBeVisible(); }

  // ==================== 卡片编辑器操作 ====================
  async clickFormatButton(): Promise<void> {
    await this.formatButton.first().click();
    await sleep(300);
  }

  async clickApplyButton(): Promise<void> {
    await this.applyButton.first().click();
    await sleep(500);
  }

  /** 断言统一步骤保存 toast（测试用例页签） */
  async expectStepSavedToast(): Promise<void> {
    await expect(this.page.locator('.n-message').filter({ hasText: '已保存' }).first()).toBeVisible({ timeout: 5000 });
  }

  /** 断言步骤顺序保存 toast（测试用例页签） */
  async expectOrderSavedToast(): Promise<void> {
    await expect(this.page.locator('.n-message').filter({ hasText: '顺序已保存' }).first()).toBeVisible({ timeout: 5000 });
  }

  /** 顺序变更提示栏 */
  get orderDirtyBar(): Locator {
    return this.page.locator('[data-testid="order-dirty-bar"]:visible');
  }

  /** 保存步骤顺序 */
  async clickSaveOrder(): Promise<void> {
    await this.orderDirtyBar.locator('[data-testid="save-order-btn"]').click();
    await sleep(500);
  }

  /** 还原步骤顺序 */
  async clickRevertOrder(): Promise<void> {
    await this.orderDirtyBar.locator('[data-testid="revert-order-btn"]').click();
    await sleep(500);
  }

  /** 当前可见表格中的拖动手柄数量 */
  async getDragHandleCount(): Promise<number> {
    return await this.page.locator('[data-testid="row-drag-handle"]:visible').count();
  }

  /** 将表格第 fromIndex 行拖到第 toIndex 行（仅拖动手柄） */
  async dragTableRow(fromIndex: number, toIndex: number): Promise<void> {
    const rows = this.getTableRows();
    await rows.nth(fromIndex).waitFor({ state: 'visible' });
    const source = rows.nth(fromIndex).locator('[data-testid="row-drag-handle"]');
    await source.waitFor({ state: 'visible' });
    const target = rows.nth(toIndex);
    await source.dragTo(target);
    await sleep(300);
  }

  async isCardEditorVisible(): Promise<boolean> {
    return await this.reqCardEditor.count() > 0;
  }

  async isNtfDisplayVisible(): Promise<boolean> {
    return await this.ntfDisplayArea.count() > 0;
  }

  async getFieldInputCount(): Promise<number> {
    return await this.fieldCardContainer.locator('input').count();
  }

  // ==================== 重放相关辅助方法 ====================

  /**
   * 获取表格中指定行的文本内容
   * @param rowIndex 行索引（0-based）
   */
  async getRowText(rowIndex: number): Promise<string> {
    const rows = this.getTableRows();
    const count = await rows.count();
    if (rowIndex >= count) {
      throw new Error(`行索引 ${rowIndex} 超出范围，总行数: ${count}`);
    }
    return await rows.nth(rowIndex).textContent() || '';
  }

  /**
   * 等待表格行数达到指定值
   * @param expectedCount 期望的行数
   * @param timeout 超时时间（毫秒）
   */
  async waitForTableRowCount(expectedCount: number, timeout = 5000): Promise<void> {
    const startTime = Date.now();
    while (Date.now() - startTime < timeout) {
      const currentCount = await this.getTableRowCount();
      if (currentCount >= expectedCount) {
        return;
      }
      await sleep(200);
    }
    throw new Error(`等待表格行数达到 ${expectedCount} 超时，当前行数: ${await this.getTableRowCount()}`);
  }

  /**
   * 获取重放状态标签文本
   */
  async getReplayStatusText(): Promise<string> {
    const tag = this.replayStatusTag.first();
    if (await tag.count() > 0) {
      return await tag.textContent() || '';
    }
    return '';
  }

  /**
   * 检查重放是否正在进行
   */
  async isReplayRunning(): Promise<boolean> {
    const statusText = await this.getReplayStatusText();
    return statusText.includes('正在重放') || statusText.includes('正在执行');
  }

  /**
   * 等待重放完成
   * @param timeout 超时时间（毫秒）
   */
  async waitForReplayComplete(timeout = 10000): Promise<void> {
    const startTime = Date.now();
    while (Date.now() - startTime < timeout) {
      const running = await this.isReplayRunning();
      if (!running) {
        return;
      }
      await sleep(500);
    }
    throw new Error('等待重放完成超时');
  }

  // ==================== 新输入组件辅助方法 ====================

  /**
   * 检查范围输入组件是否可见
   */
  async isRangeInputVisible(): Promise<boolean> {
    return await this.rangeInputLabel.count() > 0;
  }

  /**
   * 检查枚举值选择组件是否可见
   */
  async isEnumSelectVisible(): Promise<boolean> {
    return await this.enumSelectLabel.count() > 0;
  }

  /**
   * 检查组合选择组件是否可见
   */
  async isComboSelectVisible(): Promise<boolean> {
    return await this.comboSelectLabel.count() > 0;
  }

  /**
   * 设置范围输入值
   */
  async setRangeInputValues(start: number, step: number, end: number): Promise<void> {
    await this.rangeInputStart.click();
    await this.rangeInputStart.fill(String(start));
    await sleep(100);

    await this.rangeInputStep.click();
    await this.rangeInputStep.fill(String(step));
    await sleep(100);

    await this.rangeInputEnd.click();
    await this.rangeInputEnd.fill(String(end));
    await sleep(100);
  }

  /**
   * 获取范围输入的当前值
   */
  async getRangeInputValues(): Promise<{ start: string; step: string; end: string }> {
    const start = await this.rangeInputStart.inputValue();
    const step = await this.rangeInputStep.inputValue();
    const end = await this.rangeInputEnd.inputValue();
    return { start, step, end };
  }

  /**
   * 添加枚举值
   */
  async addEnumValue(value: string): Promise<void> {
    await this.enumSelectInput.click();
    await sleep(300);

    // 输入新值并按回车创建标签
    const searchInput = this.page.locator('.n-base-select-menu__input').first();
    await searchInput.fill(value);
    await sleep(200);
    await this.page.keyboard.press('Enter');
    await sleep(200);

    // 关闭下拉框
    await this.page.keyboard.press('Escape');
    await sleep(200);
  }

  /**
   * 获取枚举值标签数量
   */
  async getEnumTagCount(): Promise<number> {
    // 枚举值标签在 .n-base-selection 内部
    const tags = this.enumSelectInput.locator('.n-tag');
    return await tags.count();
  }

  /**
   * 添加组合值
   */
  async addComboValue(value: string): Promise<void> {
    await this.comboSelectInput.click();
    await sleep(300);

    // 输入新值并按回车创建标签
    const searchInput = this.page.locator('.n-base-select-menu__input').first();
    await searchInput.fill(value);
    await sleep(200);
    await this.page.keyboard.press('Enter');
    await sleep(200);

    // 关闭下拉框
    await this.page.keyboard.press('Escape');
    await sleep(200);
  }

  /**
   * 获取组合值标签数量
   */
  async getComboTagCount(): Promise<number> {
    const tags = this.comboSelectInput.locator('.n-tag');
    return await tags.count();
  }

  /**
   * 移除枚举值标签
   */
  async removeEnumTag(index: number): Promise<void> {
    const tags = this.enumSelectInput.locator('.n-tag');
    const count = await tags.count();
    if (index >= count) return;

    const tag = tags.nth(index);
    const closeIcon = tag.locator('.n-tag__close').first();
    await closeIcon.click();
    await sleep(200);
  }

  /**
   * 移除组合值标签
   */
  async removeComboTag(index: number): Promise<void> {
    const tags = this.comboSelectInput.locator('.n-tag');
    const count = await tags.count();
    if (index >= count) return;

    const tag = tags.nth(index);
    const closeIcon = tag.locator('.n-tag__close').first();
    await closeIcon.click();
    await sleep(200);
  }

  /**
   * 根据字段名查找字段卡片
   */
  getFieldCardByName(fieldName: string): Locator {
    return this.page.locator('.field-item').filter({ hasText: fieldName });
  }

  /**
   * 获取字段项的组件选择下拉菜单
   */
  getFieldItemDropdown(fieldItem: Locator): Locator {
    return fieldItem.locator('.n-select');
  }

  /**
   * 获取字段项的原始值只读输入框
   */
  getFieldItemOriginalInput(fieldItem: Locator): Locator {
    return fieldItem.locator('.n-input[readonly]');
  }

  /**
   * 切换字段组件类型
   */
  async switchFieldType(fieldItem: Locator, typeName: string): Promise<void> {
    const dropdown = this.getFieldItemDropdown(fieldItem);
    await dropdown.click();
    await sleep(300);

    const option = this.page.locator('.n-base-select-option').filter({ hasText: typeName });
    await option.click();
    await sleep(300);
  }

  /**
   * 检查字段是否显示原始值模式
   */
  async isFieldOriginalMode(fieldItem: Locator): Promise<boolean> {
    const originalLabel = fieldItem.locator('text=原始值:');
    const readonlyInput = fieldItem.locator('.n-input[readonly]');
    return await originalLabel.count() > 0 && await readonlyInput.count() > 0;
  }

  /**
   * 检查字段是否显示范围输入模式
   */
  async isFieldRangeMode(fieldItem: Locator): Promise<boolean> {
    const startLabel = fieldItem.locator('text=起始值:');
    const stepLabel = fieldItem.locator('text=步长:');
    const endLabel = fieldItem.locator('text=终值:');
    return await startLabel.count() > 0 && await stepLabel.count() > 0 && await endLabel.count() > 0;
  }

  /**
   * 检查字段是否显示枚举选择模式
   */
  async isFieldEnumMode(fieldItem: Locator): Promise<boolean> {
    const enumLabel = fieldItem.locator('text=枚举值:');
    return await enumLabel.count() > 0;
  }

  /**
   * 检查字段是否显示组合选择模式
   */
  async isFieldComboMode(fieldItem: Locator): Promise<boolean> {
    const comboLabel = fieldItem.locator('text=组合:');
    return await comboLabel.count() > 0;
  }

  // ==================== 重放结果页签操作 ====================

  async clickTabReplayResult(): Promise<void> {
    await this.tabReplayResult.click();
    await sleep(500);
  }

  async getReplayResultRowCount(): Promise<number> {
    // 重放结果页签的表格行数
    const resultTable = this.page.locator('.n-data-table').last();
    return await resultTable.locator('tbody tr').count();
  }

  async getResultSourceLabel(): Promise<string> {
    const label = this.sourceLabel;
    if (await label.count() > 0) {
      const text = await label.textContent() || '';
      return text.trim();
    }
    return '';
  }

  async getResultSelectorCount(): Promise<number> {
    // 获取重放结果选择器中的结果数量
    const selector = this.resultSelector;
    if (await selector.count() === 0) return 0;
    await selector.click({ force: true });
    await sleep(500);
    const options = await this.page.locator('.n-base-select-option').count();
    await this.page.keyboard.press('Escape');
    await sleep(200);
    return options;
  }

  // ==================== JSON 编辑器操作 ====================

  /**
   * 获取 JSON 编辑器的 textarea 元素
   * （paired-payload-editor 中 JSON 模式的输入框）
   */
  getJsonEditorTextarea(): Locator {
    // paired-payload-editor 中 JSON 模式使用 n-input type="textarea"
    // 定位方式：在 paired-payload-editor 区域内找 textarea 输入框
    // 左侧是 Req 编辑器（可编辑），右侧是 Ack 编辑器（只读）
    return this.page.locator('textarea').first();
  }

  /**
   * 获取 JSON 编辑器当前内容
   */
  async getJsonEditorValue(): Promise<string> {
    const textarea = this.getJsonEditorTextarea();
    return await textarea.inputValue();
  }

  /**
   * 设置 JSON 编辑器内容
   */
  async setJsonEditorValue(value: string): Promise<void> {
    const textarea = this.getJsonEditorTextarea();
    await textarea.click();
    await sleep(200);
    // 全选并替换
    await this.page.keyboard.press('Control+a');
    await sleep(100);
    await textarea.fill(value);
    await sleep(200);
  }

  /**
   * 检查"应用"按钮是否可用（有修改时启用）
   */
  async isApplyButtonEnabled(): Promise<boolean> {
    const applyBtn = this.applyButton.first();
    if (await applyBtn.count() === 0) return false;
    return await applyBtn.isEnabled();
  }

  /**
   * 获取重放结果页签表格中指定行的文本
   */
  async getReplayResultRowText(rowIndex: number): Promise<string> {
    const resultTable = this.page.locator('.n-data-table').last();
    const rows = resultTable.locator('tbody tr');
    const count = await rows.count();
    if (rowIndex >= count) {
      throw new Error(`重放结果行索引 ${rowIndex} 超出范围，总行数: ${count}`);
    }
    return await rows.nth(rowIndex).textContent() || '';
  }

  /**
   * 获取重放结果页签表格最后一行的文本
   */
  async getLastReplayResultRowText(): Promise<string> {
    const resultTable = this.page.locator('.n-data-table').last();
    const rows = resultTable.locator('tbody tr');
    const count = await rows.count();
    if (count === 0) return '';
    return await rows.last().textContent() || '';
  }
}
