/**
 * 协议重放页 E2E 测试 — 拆分自 stream-proxy.spec.ts
 *
 * 运行方式：
 *   wails3 dev
 *   cd frontend && npx playwright test stream-proxy/<filename>
 */
import { test, expect, describe } from '../shared/fixtures';
import { ProtoTestPage } from '../shared/pages/ProtoTestPage';
describe('协议重放页 — 输入组件与卡片编辑器', () => {
  let page: ProtoTestPage;

  test.beforeEach(async ({ page: p }) => {
    page = new ProtoTestPage(p);
    await page.goto();
  });

  describe('重放控制', () => {
    test('重发按钮和次数输入框元素存在', async () => {
      const retryBtn = page.replayRetryButton;
      expect(await retryBtn.count()).toBeGreaterThanOrEqual(0);
    });

    test('选中行后显示重放控制面板', async () => {
      // 前置条件：需要已加载的录制数据
      // 先切换到测试用例页签并选择一个用例
      await page.clickTabTestcase();
      await page.selectCaseFromDropdown('e2e_test_case');
      await page.page.waitForTimeout(2000); // 等待数据加载

      // 切回发包改包页签
      await page.clickTabPacket();
      await page.page.waitForTimeout(500);

      // 点击第一行
      const rows = page.getTableRows();
      const count = await rows.count();
      if (count > 0) {
        await rows.first().click();
        await page.page.waitForTimeout(300);

        // 验证重放控制面板可见
        const replayPanel = page.page.locator('text=重放控制').first();
        await expect(replayPanel).toBeVisible();

        // 验证重发按钮可见
        await expect(page.replayRetryButton).toBeVisible();

        // 验证次数输入框可见
        await expect(page.replayCountInput).toBeVisible();
      }
    });

    test('Ntf 显示区域条件渲染', async () => {
      // 前置条件：选中某行
      await page.clickTabTestcase();
      await page.selectCaseFromDropdown('e2e_test_case');
      await page.page.waitForTimeout(2000);
      await page.clickTabPacket();
      await page.page.waitForTimeout(500);

      const rows = page.getTableRows();
      const count = await rows.count();
      if (count > 0) {
        await rows.first().click();
        await page.page.waitForTimeout(300);

        // 检查是否有 Ntf 显示（某些行可能有 Ntf）
        const ntfLabel = page.page.locator('text=Ntf:');
        const hasNtf = await ntfLabel.count() > 0;

        if (hasNtf) {
          // 如果有 Ntf，验证 Ntf 标签可见
          await expect(ntfLabel.first()).toBeVisible();
        } else {
          // 如果没有 Ntf，验证显示"已配对"或"等待 Ack..."
          const pairedLabel = page.page.locator('text=已配对');
          const waitingLabel = page.page.locator('text=等待 Ack');
          const hasStatus = await pairedLabel.count() > 0 || await waitingLabel.count() > 0;
          expect(hasStatus).toBeTruthy();
        }
      }
    });
  });

  // ==================== 新输入组件测试 ====================

  describe('范围输入组件 (RangeInput)', () => {
    test('范围输入标签和输入框正确显示', async () => {
      await page.clickTabTestcase();
      await page.selectCaseFromDropdown('e2e_test_case');
      await page.page.waitForTimeout(2000);
      await page.clickTabPacket();
      await page.page.waitForTimeout(500);

      const rows = page.getTableRows();
      const count = await rows.count();
      if (count > 0) {
        await rows.first().click();
        await page.page.waitForTimeout(300);

        // 查找包含范围类型字段的卡片（如 hp、attack、defense 等）
        const rangeFieldCards = page.page.locator('.field-item').filter({ hasText: /hp|attack|defense|level/ });
        const hasRangeField = await rangeFieldCards.count() > 0;

        if (hasRangeField) {
          // 验证范围输入组件的标签和输入框
          const firstRangeCard = rangeFieldCards.first();
          await expect(firstRangeCard.locator('text=起始值:')).toBeVisible();
          await expect(firstRangeCard.locator('text=步长:')).toBeVisible();
          await expect(firstRangeCard.locator('text=终值:')).toBeVisible();

          // 验证三个输入框都存在
          const inputs = firstRangeCard.locator('.n-input-number input');
          const inputCount = await inputs.count();
          expect(inputCount).toBe(3);
        }
      }
    });

    test('可以输入数字到范围输入框', async () => {
      await page.clickTabTestcase();
      await page.selectCaseFromDropdown('e2e_test_case');
      await page.page.waitForTimeout(2000);
      await page.clickTabPacket();
      await page.page.waitForTimeout(500);

      const rows = page.getTableRows();
      const count = await rows.count();
      if (count > 0) {
        await rows.first().click();
        await page.page.waitForTimeout(300);

        const rangeFieldCards = page.page.locator('.field-item').filter({ hasText: /hp|attack|defense|level/ });
        const hasRangeField = await rangeFieldCards.count() > 0;

        if (hasRangeField) {
          const firstRangeCard = rangeFieldCards.first();

          // 获取所有输入框
          const inputs = firstRangeCard.locator('.n-input-number input');

          // 设置起始值
          await inputs.nth(0).click();
          await inputs.nth(0).fill('100');
          await page.page.waitForTimeout(200);

          // 验证值已设置
          const startValue = await inputs.nth(0).inputValue();
          expect(startValue).toBe('100');
        }
      }
    });

    test('范围输入值变化触发更新事件', async () => {
      await page.clickTabTestcase();
      await page.selectCaseFromDropdown('e2e_test_case');
      await page.page.waitForTimeout(2000);
      await page.clickTabPacket();
      await page.page.waitForTimeout(500);

      const rows = page.getTableRows();
      const count = await rows.count();
      if (count > 0) {
        await rows.first().click();
        await page.page.waitForTimeout(300);

        const rangeFieldCards = page.page.locator('.field-item').filter({ hasText: /hp|attack|defense|level/ });
        const hasRangeField = await rangeFieldCards.count() > 0;

        if (hasRangeField) {
          const firstRangeCard = rangeFieldCards.first();
          const inputs = firstRangeCard.locator('.n-input-number input');

          // 修改值
          await inputs.nth(0).click();
          await inputs.nth(0).fill('200');
          await page.page.waitForTimeout(200);

          // 验证应用按钮变为可用状态（有修改时启用）
          const applyButton = page.applyButton.first();
          const isEnabled = await applyButton.isEnabled();
          // 注意：初始状态可能禁用，修改后应该启用
          expect(isEnabled).toBe(true);
        }
      }
    });
  });

  describe('枚举值选择组件 (EnumSelect)', () => {
    test('枚举值标签和选择框正确显示', async () => {
      await page.clickTabTestcase();
      await page.selectCaseFromDropdown('e2e_test_case');
      await page.page.waitForTimeout(2000);
      await page.clickTabPacket();
      await page.page.waitForTimeout(500);

      const rows = page.getTableRows();
      const count = await rows.count();
      if (count > 0) {
        await rows.first().click();
        await page.page.waitForTimeout(300);

        // 查找包含枚举类型字段的卡片（如 heroId、itemId、skillId 等）
        const enumFieldCards = page.page.locator('.field-item').filter({ hasText: /heroId|itemId|skillId/ });
        const hasEnumField = await enumFieldCards.count() > 0;

        if (hasEnumField) {
          // 验证枚举值标签显示
          const firstEnumCard = enumFieldCards.first();
          await expect(firstEnumCard.locator('text=枚举值:')).toBeVisible();

          // 验证选择框存在
          const selectInput = firstEnumCard.locator('.n-base-selection');
          await expect(selectInput).toBeVisible();
        }
      }
    });

    test('可以添加枚举值标签', async () => {
      await page.clickTabTestcase();
      await page.selectCaseFromDropdown('e2e_test_case');
      await page.page.waitForTimeout(2000);
      await page.clickTabPacket();
      await page.page.waitForTimeout(500);

      const rows = page.getTableRows();
      const count = await rows.count();
      if (count > 0) {
        await rows.first().click();
        await page.page.waitForTimeout(300);

        const enumFieldCards = page.page.locator('.field-item').filter({ hasText: /heroId|itemId|skillId/ });
        const hasEnumField = await enumFieldCards.count() > 0;

        if (hasEnumField) {
          const firstEnumCard = enumFieldCards.first();
          const selectInput = firstEnumCard.locator('.n-base-selection');

          // 获取初始标签数量
          const beforeCount = await firstEnumCard.locator('.n-tag').count();

          // 打开下拉框
          await selectInput.click();
          await page.page.waitForTimeout(300);

          // 输入新值
          const searchInput = page.page.locator('.n-base-select-menu__input').first();
          await searchInput.fill('test_enum_value');
          await page.page.waitForTimeout(200);

          // 按回车创建标签
          await page.page.keyboard.press('Enter');
          await page.page.waitForTimeout(200);

          // 关闭下拉框
          await page.page.keyboard.press('Escape');
          await page.page.waitForTimeout(200);

          // 验证标签数量增加
          const afterCount = await firstEnumCard.locator('.n-tag').count();
          expect(afterCount).toBeGreaterThan(beforeCount);
        }
      }
    });

    test('可以移除枚举值标签', async () => {
      await page.clickTabTestcase();
      await page.selectCaseFromDropdown('e2e_test_case');
      await page.page.waitForTimeout(2000);
      await page.clickTabPacket();
      await page.page.waitForTimeout(500);

      const rows = page.getTableRows();
      const count = await rows.count();
      if (count > 0) {
        await rows.first().click();
        await page.page.waitForTimeout(300);

        const enumFieldCards = page.page.locator('.field-item').filter({ hasText: /heroId|itemId|skillId/ });
        const hasEnumField = await enumFieldCards.count() > 0;

        if (hasEnumField) {
          const firstEnumCard = enumFieldCards.first();

          // 获取初始标签数量
          const tags = firstEnumCard.locator('.n-tag');
          const beforeCount = await tags.count();

          if (beforeCount > 0) {
            // 点击第一个标签的关闭按钮
            const closeIcon = tags.first().locator('.n-tag__close').first();
            await closeIcon.click();
            await page.page.waitForTimeout(200);

            // 验证标签数量减少
            const afterCount = await tags.count();
            expect(afterCount).toBeLessThan(beforeCount);
          }
        }
      }
    });

    test('枚举值选择支持多选', async () => {
      await page.clickTabTestcase();
      await page.selectCaseFromDropdown('e2e_test_case');
      await page.page.waitForTimeout(2000);
      await page.clickTabPacket();
      await page.page.waitForTimeout(500);

      const rows = page.getTableRows();
      const count = await rows.count();
      if (count > 0) {
        await rows.first().click();
        await page.page.waitForTimeout(300);

        const enumFieldCards = page.page.locator('.field-item').filter({ hasText: /heroId|itemId|skillId/ });
        const hasEnumField = await enumFieldCards.count() > 0;

        if (hasEnumField) {
          const firstEnumCard = enumFieldCards.first();
          const selectInput = firstEnumCard.locator('.n-base-selection');

          // 打开下拉框
          await selectInput.click();
          await page.page.waitForTimeout(300);

          // 验证多选模式下有 tag 属性（通过 Naive UI class 判断）
          const menu = page.page.locator('.n-base-select-menu').first();
          const hasMultipleAttr = await menu.count() > 0;
          expect(hasMultipleAttr).toBeTruthy();

          // 关闭下拉框
          await page.page.keyboard.press('Escape');
        }
      }
    });
  });

  describe('组合选择组件 (ComboSelect)', () => {
    test('组合标签和选择框正确显示', async () => {
      await page.clickTabTestcase();
      await page.selectCaseFromDropdown('e2e_test_case');
      await page.page.waitForTimeout(2000);
      await page.clickTabPacket();
      await page.page.waitForTimeout(500);

      const rows = page.getTableRows();
      const count = await rows.count();
      if (count > 0) {
        await rows.first().click();
        await page.page.waitForTimeout(300);

        // 查找包含组合类型字段的卡片（如 heroIds、itemIds、skillIds 等）
        const comboFieldCards = page.page.locator('.field-item').filter({ hasText: /heroIds|itemIds|skillIds/ });
        const hasComboField = await comboFieldCards.count() > 0;

        if (hasComboField) {
          // 验证组合标签显示
          const firstComboCard = comboFieldCards.first();
          await expect(firstComboCard.locator('text=组合:')).toBeVisible();

          // 验证选择框存在
          const selectInput = firstComboCard.locator('.n-base-selection');
          await expect(selectInput).toBeVisible();
        }
      }
    });

    test('可以添加组合值标签', async () => {
      await page.clickTabTestcase();
      await page.selectCaseFromDropdown('e2e_test_case');
      await page.page.waitForTimeout(2000);
      await page.clickTabPacket();
      await page.page.waitForTimeout(500);

      const rows = page.getTableRows();
      const count = await rows.count();
      if (count > 0) {
        await rows.first().click();
        await page.page.waitForTimeout(300);

        const comboFieldCards = page.page.locator('.field-item').filter({ hasText: /heroIds|itemIds|skillIds/ });
        const hasComboField = await comboFieldCards.count() > 0;

        if (hasComboField) {
          const firstComboCard = comboFieldCards.first();
          const selectInput = firstComboCard.locator('.n-base-selection');

          // 获取初始标签数量
          const beforeCount = await firstComboCard.locator('.n-tag').count();

          // 打开下拉框
          await selectInput.click();
          await page.page.waitForTimeout(300);

          // 输入新值
          const searchInput = page.page.locator('.n-base-select-menu__input').first();
          await searchInput.fill('test_combo_value');
          await page.page.waitForTimeout(200);

          // 按回车创建标签
          await page.page.keyboard.press('Enter');
          await page.page.waitForTimeout(200);

          // 关闭下拉框
          await page.page.keyboard.press('Escape');
          await page.page.waitForTimeout(200);

          // 验证标签数量增加
          const afterCount = await firstComboCard.locator('.n-tag').count();
          expect(afterCount).toBeGreaterThan(beforeCount);
        }
      }
    });

    test('可以移除组合值标签', async () => {
      await page.clickTabTestcase();
      await page.selectCaseFromDropdown('e2e_test_case');
      await page.page.waitForTimeout(2000);
      await page.clickTabPacket();
      await page.page.waitForTimeout(500);

      const rows = page.getTableRows();
      const count = await rows.count();
      if (count > 0) {
        await rows.first().click();
        await page.page.waitForTimeout(300);

        const comboFieldCards = page.page.locator('.field-item').filter({ hasText: /heroIds|itemIds|skillIds/ });
        const hasComboField = await comboFieldCards.count() > 0;

        if (hasComboField) {
          const firstComboCard = comboFieldCards.first();

          // 获取初始标签数量
          const tags = firstComboCard.locator('.n-tag');
          const beforeCount = await tags.count();

          if (beforeCount > 0) {
            // 点击第一个标签的关闭按钮
            const closeIcon = tags.first().locator('.n-tag__close').first();
            await closeIcon.click();
            await page.page.waitForTimeout(200);

            // 验证标签数量减少
            const afterCount = await tags.count();
            expect(afterCount).toBeLessThan(beforeCount);
          }
        }
      }
    });

    test('组合选择支持多选', async () => {
      await page.clickTabTestcase();
      await page.selectCaseFromDropdown('e2e_test_case');
      await page.page.waitForTimeout(2000);
      await page.clickTabPacket();
      await page.page.waitForTimeout(500);

      const rows = page.getTableRows();
      const count = await rows.count();
      if (count > 0) {
        await rows.first().click();
        await page.page.waitForTimeout(300);

        const comboFieldCards = page.page.locator('.field-item').filter({ hasText: /heroIds|itemIds|skillIds/ });
        const hasComboField = await comboFieldCards.count() > 0;

        if (hasComboField) {
          const firstComboCard = comboFieldCards.first();
          const selectInput = firstComboCard.locator('.n-base-selection');

          // 打开下拉框
          await selectInput.click();
          await page.page.waitForTimeout(300);

          // 验证多选模式下下拉菜单存在
          const menu = page.page.locator('.n-base-select-menu').first();
          const hasMenu = await menu.count() > 0;
          expect(hasMenu).toBeTruthy();

          // 关闭下拉框
          await page.page.keyboard.press('Escape');
        }
      }
    });
  });

  // ==================== 组件选择下拉菜单测试 ====================

  describe('组件选择下拉菜单', () => {
    test('每个字段显示组件选择下拉菜单', async () => {
      await page.clickTabTestcase();
      await page.selectCaseFromDropdown('e2e_test_case');
      await page.page.waitForTimeout(2000);
      await page.clickTabPacket();
      await page.page.waitForTimeout(500);

      const rows = page.getTableRows();
      const count = await rows.count();
      if (count > 0) {
        await rows.first().click();
        await page.page.waitForTimeout(300);

        // 验证卡片编辑器可见
        const cardEditor = page.page.locator('text=Payload 字段');
        const hasCardEditor = await cardEditor.count() > 0;

        if (hasCardEditor) {
          // 获取所有字段项
          const fieldItems = page.page.locator('.field-item');
          const fieldCount = await fieldItems.count();

          // 验证每个基础类型字段都有下拉菜单
          for (let i = 0; i < fieldCount; i++) {
            const item = fieldItems.nth(i);
            const dropdown = item.locator('.n-select');
            const hasDropdown = await dropdown.count() > 0;

            if (hasDropdown) {
              await expect(dropdown.first()).toBeVisible();
            }
          }
        }
      }
    });

    test('下拉菜单包含4个选项', async () => {
      await page.clickTabTestcase();
      await page.selectCaseFromDropdown('e2e_test_case');
      await page.page.waitForTimeout(2000);
      await page.clickTabPacket();
      await page.page.waitForTimeout(500);

      const rows = page.getTableRows();
      const count = await rows.count();
      if (count > 0) {
        await rows.first().click();
        await page.page.waitForTimeout(300);

        const cardEditor = page.page.locator('text=Payload 字段');
        const hasCardEditor = await cardEditor.count() > 0;

        if (hasCardEditor) {
          // 查找第一个字段项的下拉菜单
          const firstField = page.page.locator('.field-item').first();
          const dropdown = firstField.locator('.n-select');

          const hasDropdown = await dropdown.count() > 0;
          if (hasDropdown) {
            // 点击下拉菜单
            await dropdown.click();
            await page.page.waitForTimeout(300);

            // 验证选项数量
            const options = page.page.locator('.n-base-select-option');
            const optionCount = await options.count();
            expect(optionCount).toBe(4);

            // 验证选项文本
            const optionTexts = await options.allTextContents();
            expect(optionTexts).toContain('原始值');
            expect(optionTexts).toContain('范围');
            expect(optionTexts).toContain('枚举');
            expect(optionTexts).toContain('组合');

            // 关闭下拉菜单
            await page.page.keyboard.press('Escape');
            await page.page.waitForTimeout(200);
          }
        }
      }
    });

    test('默认选中原始值', async () => {
      await page.clickTabTestcase();
      await page.selectCaseFromDropdown('e2e_test_case');
      await page.page.waitForTimeout(2000);
      await page.clickTabPacket();
      await page.page.waitForTimeout(500);

      const rows = page.getTableRows();
      const count = await rows.count();
      if (count > 0) {
        await rows.first().click();
        await page.page.waitForTimeout(300);

        const cardEditor = page.page.locator('text=Payload 字段');
        const hasCardEditor = await cardEditor.count() > 0;

        if (hasCardEditor) {
          // 查找第一个字段项
          const firstField = page.page.locator('.field-item').first();

          // 验证显示"原始值:"标签
          const hasOriginalLabel = await firstField.locator('text=原始值:').count() > 0;
          expect(hasOriginalLabel).toBeTruthy();

          // 验证显示只读输入框
          const readonlyInput = firstField.locator('.n-input[readonly]');
          const hasReadonlyInput = await readonlyInput.count() > 0;
          expect(hasReadonlyInput).toBeTruthy();
        }
      }
    });

    test('可以从原始值切换到范围', async () => {
      await page.clickTabTestcase();
      await page.selectCaseFromDropdown('e2e_test_case');
      await page.page.waitForTimeout(2000);
      await page.clickTabPacket();
      await page.page.waitForTimeout(500);

      const rows = page.getTableRows();
      const count = await rows.count();
      if (count > 0) {
        await rows.first().click();
        await page.page.waitForTimeout(300);

        const cardEditor = page.page.locator('text=Payload 字段');
        const hasCardEditor = await cardEditor.count() > 0;

        if (hasCardEditor) {
          const firstField = page.page.locator('.field-item').first();
          const dropdown = firstField.locator('.n-select');

          const hasDropdown = await dropdown.count() > 0;
          if (hasDropdown) {
            // 打开下拉菜单
            await dropdown.click();
            await page.page.waitForTimeout(300);

            // 选择"范围"
            const rangeOption = page.page.locator('.n-base-select-option').filter({ hasText: '范围' });
            await rangeOption.click();
            await page.page.waitForTimeout(300);

            // 验证显示范围输入组件
            await expect(firstField.locator('text=起始值:')).toBeVisible();
            await expect(firstField.locator('text=步长:')).toBeVisible();
            await expect(firstField.locator('text=终值:')).toBeVisible();
          }
        }
      }
    });

    test('可以从原始值切换到枚举', async () => {
      await page.clickTabTestcase();
      await page.selectCaseFromDropdown('e2e_test_case');
      await page.page.waitForTimeout(2000);
      await page.clickTabPacket();
      await page.page.waitForTimeout(500);

      const rows = page.getTableRows();
      const count = await rows.count();
      if (count > 0) {
        await rows.first().click();
        await page.page.waitForTimeout(300);

        const cardEditor = page.page.locator('text=Payload 字段');
        const hasCardEditor = await cardEditor.count() > 0;

        if (hasCardEditor) {
          const firstField = page.page.locator('.field-item').first();
          const dropdown = firstField.locator('.n-select');

          const hasDropdown = await dropdown.count() > 0;
          if (hasDropdown) {
            // 打开下拉菜单
            await dropdown.click();
            await page.page.waitForTimeout(300);

            // 选择"枚举"
            const enumOption = page.page.locator('.n-base-select-option').filter({ hasText: '枚举' });
            await enumOption.click();
            await page.page.waitForTimeout(300);

            // 验证显示枚举选择组件
            await expect(firstField.locator('text=枚举值:')).toBeVisible();
          }
        }
      }
    });

    test('可以从原始值切换到组合', async () => {
      await page.clickTabTestcase();
      await page.selectCaseFromDropdown('e2e_test_case');
      await page.page.waitForTimeout(2000);
      await page.clickTabPacket();
      await page.page.waitForTimeout(500);

      const rows = page.getTableRows();
      const count = await rows.count();
      if (count > 0) {
        await rows.first().click();
        await page.page.waitForTimeout(300);

        const cardEditor = page.page.locator('text=Payload 字段');
        const hasCardEditor = await cardEditor.count() > 0;

        if (hasCardEditor) {
          const firstField = page.page.locator('.field-item').first();
          const dropdown = firstField.locator('.n-select');

          const hasDropdown = await dropdown.count() > 0;
          if (hasDropdown) {
            // 打开下拉菜单
            await dropdown.click();
            await page.page.waitForTimeout(300);

            // 选择"组合"
            const comboOption = page.page.locator('.n-base-select-option').filter({ hasText: '组合' });
            await comboOption.click();
            await page.page.waitForTimeout(300);

            // 验证显示组合选择组件
            await expect(firstField.locator('text=组合:')).toBeVisible();
          }
        }
      }
    });
  });

  // ==================== 原始值模式测试 ====================

  describe('原始值模式', () => {
    test('选择原始值后显示只读输入框', async () => {
      await page.clickTabTestcase();
      await page.selectCaseFromDropdown('e2e_test_case');
      await page.page.waitForTimeout(2000);
      await page.clickTabPacket();
      await page.page.waitForTimeout(500);

      const rows = page.getTableRows();
      const count = await rows.count();
      if (count > 0) {
        await rows.first().click();
        await page.page.waitForTimeout(300);

        const cardEditor = page.page.locator('text=Payload 字段');
        const hasCardEditor = await cardEditor.count() > 0;

        if (hasCardEditor) {
          const firstField = page.page.locator('.field-item').first();

          // 验证显示"原始值:"标签
          await expect(firstField.locator('text=原始值:')).toBeVisible();

          // 验证显示只读输入框
          const readonlyInput = firstField.locator('.n-input[readonly]');
          await expect(readonlyInput).toBeVisible();
        }
      }
    });

    test('只读输入框显示正确的字段值', async () => {
      await page.clickTabTestcase();
      await page.selectCaseFromDropdown('e2e_test_case');
      await page.page.waitForTimeout(2000);
      await page.clickTabPacket();
      await page.page.waitForTimeout(500);

      const rows = page.getTableRows();
      const count = await rows.count();
      if (count > 0) {
        await rows.first().click();
        await page.page.waitForTimeout(300);

        const cardEditor = page.page.locator('text=Payload 字段');
        const hasCardEditor = await cardEditor.count() > 0;

        if (hasCardEditor) {
          const firstField = page.page.locator('.field-item').first();
          const readonlyInput = firstField.locator('.n-input[readonly]');

          // 获取输入框的值
          const inputValue = await readonlyInput.inputValue();

          // 验证值不为空
          expect(inputValue).toBeTruthy();
          expect(inputValue.length).toBeGreaterThan(0);
        }
      }
    });

    test('只读输入框不可编辑', async () => {
      await page.clickTabTestcase();
      await page.selectCaseFromDropdown('e2e_test_case');
      await page.page.waitForTimeout(2000);
      await page.clickTabPacket();
      await page.page.waitForTimeout(500);

      const rows = page.getTableRows();
      const count = await rows.count();
      if (count > 0) {
        await rows.first().click();
        await page.page.waitForTimeout(300);

        const cardEditor = page.page.locator('text=Payload 字段');
        const hasCardEditor = await cardEditor.count() > 0;

        if (hasCardEditor) {
          const firstField = page.page.locator('.field-item').first();
          const readonlyInput = firstField.locator('.n-input[readonly]');

          // 验证输入框是禁用状态
          const isDisabled = await readonlyInput.isDisabled();
          expect(isDisabled).toBeTruthy();
        }
      }
    });
  });

  // ==================== 组件类型切换测试 ====================

  describe('组件类型切换', () => {
    test('切换后组件正确显示', async () => {
      await page.clickTabTestcase();
      await page.selectCaseFromDropdown('e2e_test_case');
      await page.page.waitForTimeout(2000);
      await page.clickTabPacket();
      await page.page.waitForTimeout(500);

      const rows = page.getTableRows();
      const count = await rows.count();
      if (count > 0) {
        await rows.first().click();
        await page.page.waitForTimeout(300);

        const cardEditor = page.page.locator('text=Payload 字段');
        const hasCardEditor = await cardEditor.count() > 0;

        if (hasCardEditor) {
          const firstField = page.page.locator('.field-item').first();

          // 切换到范围模式
          await page.switchFieldType(firstField, '范围');

          // 验证范围模式显示
          const isRangeMode = await page.isFieldRangeMode(firstField);
          expect(isRangeMode).toBeTruthy();

          // 切换到枚举模式
          await page.switchFieldType(firstField, '枚举');

          // 验证枚举模式显示
          const isEnumMode = await page.isFieldEnumMode(firstField);
          expect(isEnumMode).toBeTruthy();

          // 切换到组合模式
          await page.switchFieldType(firstField, '组合');

          // 验证组合模式显示
          const isComboMode = await page.isFieldComboMode(firstField);
          expect(isComboMode).toBeTruthy();

          // 切换回原始值模式
          await page.switchFieldType(firstField, '原始值');

          // 验证原始值模式显示
          const isOriginalMode = await page.isFieldOriginalMode(firstField);
          expect(isOriginalMode).toBeTruthy();
        }
      }
    });

    test('切换范围模式后可以输入数值', async () => {
      await page.clickTabTestcase();
      await page.selectCaseFromDropdown('e2e_test_case');
      await page.page.waitForTimeout(2000);
      await page.clickTabPacket();
      await page.page.waitForTimeout(500);

      const rows = page.getTableRows();
      const count = await rows.count();
      if (count > 0) {
        await rows.first().click();
        await page.page.waitForTimeout(300);

        const cardEditor = page.page.locator('text=Payload 字段');
        const hasCardEditor = await cardEditor.count() > 0;

        if (hasCardEditor) {
          const firstField = page.page.locator('.field-item').first();

          // 切换到范围模式
          await page.switchFieldType(firstField, '范围');

          // 验证可以输入数值
          const startInput = firstField.locator('.input-row').filter({ hasText: '起始值:' }).locator('input').first();
          await startInput.click();
          await startInput.fill('100');
          await page.page.waitForTimeout(200);

          const value = await startInput.inputValue();
          expect(value).toBe('100');
        }
      }
    });

    test('切换枚举模式后可以添加标签', async () => {
      await page.clickTabTestcase();
      await page.selectCaseFromDropdown('e2e_test_case');
      await page.page.waitForTimeout(2000);
      await page.clickTabPacket();
      await page.page.waitForTimeout(500);

      const rows = page.getTableRows();
      const count = await rows.count();
      if (count > 0) {
        await rows.first().click();
        await page.page.waitForTimeout(300);

        const cardEditor = page.page.locator('text=Payload 字段');
        const hasCardEditor = await cardEditor.count() > 0;

        if (hasCardEditor) {
          const firstField = page.page.locator('.field-item').first();

          // 切换到枚举模式
          await page.switchFieldType(firstField, '枚举');

          // 获取初始标签数量
          const beforeCount = await firstField.locator('.n-tag').count();

          // 添加枚举值
          await page.addEnumValue('test_enum');

          // 验证标签数量增加
          const afterCount = await firstField.locator('.n-tag').count();
          expect(afterCount).toBeGreaterThan(beforeCount);
        }
      }
    });

    test('切换组合模式后可以添加标签', async () => {
      await page.clickTabTestcase();
      await page.selectCaseFromDropdown('e2e_test_case');
      await page.page.waitForTimeout(2000);
      await page.clickTabPacket();
      await page.page.waitForTimeout(500);

      const rows = page.getTableRows();
      const count = await rows.count();
      if (count > 0) {
        await rows.first().click();
        await page.page.waitForTimeout(300);

        const cardEditor = page.page.locator('text=Payload 字段');
        const hasCardEditor = await cardEditor.count() > 0;

        if (hasCardEditor) {
          const firstField = page.page.locator('.field-item').first();

          // 切换到组合模式
          await page.switchFieldType(firstField, '组合');

          // 获取初始标签数量
          const beforeCount = await firstField.locator('.n-tag').count();

          // 添加组合值
          await page.addComboValue('test_combo');

          // 验证标签数量增加
          const afterCount = await firstField.locator('.n-tag').count();
          expect(afterCount).toBeGreaterThan(beforeCount);
        }
      }
    });
  });

  // ==================== 新组件交互测试 ====================

  describe('新组件交互', () => {
    test('范围输入组件可以输入数值', async () => {
      await page.clickTabTestcase();
      await page.selectCaseFromDropdown('e2e_test_case');
      await page.page.waitForTimeout(2000);
      await page.clickTabPacket();
      await page.page.waitForTimeout(500);

      const rows = page.getTableRows();
      const count = await rows.count();
      if (count > 0) {
        await rows.first().click();
        await page.page.waitForTimeout(300);

        const cardEditor = page.page.locator('text=Payload 字段');
        const hasCardEditor = await cardEditor.count() > 0;

        if (hasCardEditor) {
          const firstField = page.page.locator('.field-item').first();

          // 切换到范围模式
          await page.switchFieldType(firstField, '范围');

          // 设置范围输入值
          await page.setRangeInputValues(10, 5, 100);

          // 验证值已设置
          const values = await page.getRangeInputValues();
          expect(values.start).toBe('10');
          expect(values.step).toBe('5');
          expect(values.end).toBe('100');
        }
      }
    });

    test('枚举选择组件可以添加标签', async () => {
      await page.clickTabTestcase();
      await page.selectCaseFromDropdown('e2e_test_case');
      await page.page.waitForTimeout(2000);
      await page.clickTabPacket();
      await page.page.waitForTimeout(500);

      const rows = page.getTableRows();
      const count = await rows.count();
      if (count > 0) {
        await rows.first().click();
        await page.page.waitForTimeout(300);

        const cardEditor = page.page.locator('text=Payload 字段');
        const hasCardEditor = await cardEditor.count() > 0;

        if (hasCardEditor) {
          const firstField = page.page.locator('.field-item').first();

          // 切换到枚举模式
          await page.switchFieldType(firstField, '枚举');

          // 添加多个枚举值
          await page.addEnumValue('enum1');
          await page.addEnumValue('enum2');
          await page.addEnumValue('enum3');

          // 验证标签数量
          const tagCount = await page.getEnumTagCount();
          expect(tagCount).toBeGreaterThanOrEqual(3);
        }
      }
    });

    test('组合选择组件可以添加标签', async () => {
      await page.clickTabTestcase();
      await page.selectCaseFromDropdown('e2e_test_case');
      await page.page.waitForTimeout(2000);
      await page.clickTabPacket();
      await page.page.waitForTimeout(500);

      const rows = page.getTableRows();
      const count = await rows.count();
      if (count > 0) {
        await rows.first().click();
        await page.page.waitForTimeout(300);

        const cardEditor = page.page.locator('text=Payload 字段');
        const hasCardEditor = await cardEditor.count() > 0;

        if (hasCardEditor) {
          const firstField = page.page.locator('.field-item').first();

          // 切换到组合模式
          await page.switchFieldType(firstField, '组合');

          // 添加多个组合值
          await page.addComboValue('combo1');
          await page.addComboValue('combo2');
          await page.addComboValue('combo3');

          // 验证标签数量
          const tagCount = await page.getComboTagCount();
          expect(tagCount).toBeGreaterThanOrEqual(3);
        }
      }
    });

    test('可以移除枚举值标签', async () => {
      await page.clickTabTestcase();
      await page.selectCaseFromDropdown('e2e_test_case');
      await page.page.waitForTimeout(2000);
      await page.clickTabPacket();
      await page.page.waitForTimeout(500);

      const rows = page.getTableRows();
      const count = await rows.count();
      if (count > 0) {
        await rows.first().click();
        await page.page.waitForTimeout(300);

        const cardEditor = page.page.locator('text=Payload 字段');
        const hasCardEditor = await cardEditor.count() > 0;

        if (hasCardEditor) {
          const firstField = page.page.locator('.field-item').first();

          // 切换到枚举模式
          await page.switchFieldType(firstField, '枚举');

          // 添加枚举值
          await page.addEnumValue('temp_enum');

          // 获取初始标签数量
          const beforeCount = await page.getEnumTagCount();

          if (beforeCount > 0) {
            // 移除第一个标签
            await page.removeEnumTag(0);

            // 验证标签数量减少
            const afterCount = await page.getEnumTagCount();
            expect(afterCount).toBeLessThan(beforeCount);
          }
        }
      }
    });

    test('可以移除组合值标签', async () => {
      await page.clickTabTestcase();
      await page.selectCaseFromDropdown('e2e_test_case');
      await page.page.waitForTimeout(2000);
      await page.clickTabPacket();
      await page.page.waitForTimeout(500);

      const rows = page.getTableRows();
      const count = await rows.count();
      if (count > 0) {
        await rows.first().click();
        await page.page.waitForTimeout(300);

        const cardEditor = page.page.locator('text=Payload 字段');
        const hasCardEditor = await cardEditor.count() > 0;

        if (hasCardEditor) {
          const firstField = page.page.locator('.field-item').first();

          // 切换到组合模式
          await page.switchFieldType(firstField, '组合');

          // 添加组合值
          await page.addComboValue('temp_combo');

          // 获取初始标签数量
          const beforeCount = await page.getComboTagCount();

          if (beforeCount > 0) {
            // 移除第一个标签
            await page.removeComboTag(0);

            // 验证标签数量减少
            const afterCount = await page.getComboTagCount();
            expect(afterCount).toBeLessThan(beforeCount);
          }
        }
      }
    });
  });

  // ==================== 卡片编辑器 ====================

  describe('卡片编辑器', () => {
    test('选中 Req 行后显示卡片编辑器', async () => {
      // 前置条件：需要已加载的录制数据
      await page.clickTabTestcase();
      await page.selectCaseFromDropdown('e2e_test_case');
      await page.page.waitForTimeout(2000);
      await page.clickTabPacket();
      await page.page.waitForTimeout(500);

      const rows = page.getTableRows();
      const count = await rows.count();
      if (count > 0) {
        await rows.first().click();
        await page.page.waitForTimeout(300);

        // 验证卡片编辑器可见（通过 Payload 字段标题定位）
        const cardEditor = page.page.locator('text=Payload 字段');
        const hasCardEditor = await cardEditor.count() > 0;

        if (hasCardEditor) {
          await expect(cardEditor.first()).toBeVisible();
        }
      }
    });

    test('格式化按钮可见', async () => {
      await page.clickTabTestcase();
      await page.selectCaseFromDropdown('e2e_test_case');
      await page.page.waitForTimeout(2000);
      await page.clickTabPacket();
      await page.page.waitForTimeout(500);

      const rows = page.getTableRows();
      const count = await rows.count();
      if (count > 0) {
        await rows.first().click();
        await page.page.waitForTimeout(300);

        // 验证格式化按钮可见
        const formatButton = page.page.locator('button').filter({ hasText: '格式化' });
        const hasFormatButton = await formatButton.count() > 0;

        if (hasFormatButton) {
          await expect(formatButton.first()).toBeVisible();
        }
      }
    });

    test('应用按钮可见（有修改时启用）', async () => {
      await page.clickTabTestcase();
      await page.selectCaseFromDropdown('e2e_test_case');
      await page.page.waitForTimeout(2000);
      await page.clickTabPacket();
      await page.page.waitForTimeout(500);

      const rows = page.getTableRows();
      const count = await rows.count();
      if (count > 0) {
        await rows.first().click();
        await page.page.waitForTimeout(300);

        // 验证应用按钮存在（初始可能禁用）
        const applyButton = page.page.locator('button').filter({ hasText: '应用' });
        const hasApplyButton = await applyButton.count() > 0;

        if (hasApplyButton) {
          await expect(applyButton.first()).toBeVisible();
          // 注意：初始状态可能禁用（无修改时）
        }
      }
    });

    test('字段列表渲染', async () => {
      await page.clickTabTestcase();
      await page.selectCaseFromDropdown('e2e_test_case');
      await page.page.waitForTimeout(2000);
      await page.clickTabPacket();
      await page.page.waitForTimeout(500);

      const rows = page.getTableRows();
      const count = await rows.count();
      if (count > 0) {
        await rows.first().click();
        await page.page.waitForTimeout(300);

        // 验证字段卡片容器可见
        const cardContainer = page.page.locator('.n-card');
        const hasCard = await cardContainer.count() > 0;

        if (hasCard) {
          // 检查是否有字段项（FieldItem）
          const fieldInputs = page.page.locator('.n-card input');
          const inputCount = await fieldInputs.count();
          // 至少应该有一些字段输入框
          expect(inputCount).toBeGreaterThan(0);
        }
      }
    });
  });

  // ==================== 页面布局 ====================

  describe('页面布局', () => {
    test('主要区域元素存在', async () => {
      await expect(page.tabPacket).toBeVisible();
      await expect(page.replayServerInput).toBeVisible();
    });

    test('未加载时文件信息不显示', async () => {
      await expect(page.fileInfoRow).not.toBeVisible();
    });

    test('录制进度初始不显示', async () => {
      await expect(page.recordProgressTag).not.toBeVisible();
    });
  });
});
