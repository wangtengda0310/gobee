/**
 * 用例步骤 Tab — 动作/断言卡片标题行布局
 */

import { test, expect, describe } from '../shared/fixtures';
import { FunctionTestPage } from '../shared/pages/FunctionTestPage';

describe('功能测试页 - 用例步骤卡片布局', () => {
  let functionTestPage: FunctionTestPage;

  test.beforeEach(async ({ page }) => {
    functionTestPage = new FunctionTestPage(page);
    await functionTestPage.goto();
    await functionTestPage.loadCaseWithSteps('行动牌');
  });

  test('动作卡片标题行包含拖动、智能描述与描述输入', async () => {
    const headerExtra = functionTestPage.getActionStepHeaderExtra(0);
    await functionTestPage.expectCardHeaderExtraLayout(headerExtra);
    await expect(headerExtra.locator('button').filter({ hasText: '复制' })).toBeVisible();
  });

  test('已有断言卡片标题行与动作卡片布局一致', async () => {
    const assertionCount = await functionTestPage.getAssertionCards(0).count();
    test.skip(assertionCount === 0, '当前用例第一个动作无断言，跳过');

    const headerExtra = functionTestPage.getAssertionHeaderExtra(0, 0);
    await functionTestPage.expectCardHeaderExtraLayout(headerExtra);
  });

  test('断言正文仅保留类型下拉，智能描述控件在标题行', async () => {
    let assertionCount = await functionTestPage.getAssertionCards(0).count();
    if (assertionCount === 0) {
      await functionTestPage.clickAddAssertion(0);
      assertionCount = await functionTestPage.getAssertionCards(0).count();
    }
    expect(assertionCount).toBeGreaterThan(0);

    const typeRow = functionTestPage.getAssertionTypeRow(0, 0);
    await expect(typeRow).toBeVisible();
    // 正文类型行仅一个类型下拉 n-select
    await expect(typeRow.locator('.n-select')).toHaveCount(1);
    // 智能描述按钮位于标题行(header-extra)，不应出现在正文类型行
    await expect(typeRow.locator('button').filter({ hasText: '应用智能描述' })).toHaveCount(0);
    // 类型行仅含类型下拉，不应有额外的独立描述输入框（n-select 内部的 filter input 不计入：
    // 它属于类型下拉自身，而非独立的断言描述输入框）。这里用 input 数量 == n-select 数量来约束。
    await expect(typeRow.locator('input')).toHaveCount(await typeRow.locator('.n-select').count());

    await functionTestPage.expectCardHeaderExtraLayout(
      functionTestPage.getAssertionHeaderExtra(0, 0)
    );
  });

  test('新增断言后标题行仍保持统一布局', async () => {
    const before = await functionTestPage.getAssertionCards(0).count();
    await functionTestPage.clickAddAssertion(0);
    const after = await functionTestPage.getAssertionCards(0).count();
    expect(after).toBe(before + 1);

    const newIndex = after - 1;
    await expect(await functionTestPage.getAssertionTitle(0, newIndex)).toMatch(/^断言 \d+$/);
    await functionTestPage.expectCardHeaderExtraLayout(
      functionTestPage.getAssertionHeaderExtra(0, newIndex)
    );
  });

  test('动作序号按列表位置连续编号', async () => {
    const initialCount = await functionTestPage.getActionStepCards().count();
    await functionTestPage.clickAddStep(initialCount - 1);
    const newCount = await functionTestPage.getActionStepCards().count();
    expect(newCount).toBe(initialCount + 1);

    const titles: string[] = [];
    for (let i = 0; i < newCount; i++) {
      titles.push(await functionTestPage.getActionStepTitle(i));
    }
    expect(titles).toEqual(
      Array.from({ length: newCount }, (_, i) => `动作 ${i + 1}`)
    );
  });
});
