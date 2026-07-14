/**
 * 活动Wiki页面测试
 * 对应 src/pages/activity-wiki-check/index.vue
 */

import { test, expect, describe } from '../shared/fixtures';
import { ActivityWikiPage } from '../shared/pages/ActivityWikiPage';

describe('活动Wiki页面测试', () => {
  let wikiPage: ActivityWikiPage;

  test.beforeEach(async ({ page }) => {
    wikiPage = new ActivityWikiPage(page);
    await wikiPage.goto();
  });

  describe('页面加载', () => {
    test('页面加载 - 显示配置区域', async () => {
      await wikiPage.expectConfigAreaVisible();
    });

    test('Excel目录配置 - 输入Excel目录路径', async () => {
      const testPath = '/test/excel/path';
      await wikiPage.setExcelDir(testPath);
    });

    test('历史JSON配置 - 输入历史JSON路径', async () => {
      const testPath = '/test/history.json';
      await wikiPage.setOldJsonPath(testPath);
    });
  });

  describe('累充活动卡片', () => {
    // 需要后端支持
    test.skip('执行检查 - 显示活动卡片列表', async () => {
      await wikiPage.clickExecuteCheck();
      await wikiPage.waitForCheckComplete();
      const cards = wikiPage.getActivityCards();
      const count = await cards.count();
      expect(count).toBeGreaterThan(0);
    });

    test.skip('累充活动卡片 - 显示盈仓嘉礼卡片', async () => {
      await wikiPage.clickExecuteCheck();
      await wikiPage.waitForCheckComplete();
      const cards = wikiPage.getAccumulatedRechargeCards();
      const count = await cards.count();
      expect(count).toBeGreaterThan(0);
    });

    test.skip('累充活动类型标签 - 显示 ActTypeAccumulatedRecharge 标签（橙色）', async () => {
      await wikiPage.clickExecuteCheck();
      await wikiPage.waitForCheckComplete();
      const cards = wikiPage.getAccumulatedRechargeCards();
      const firstCard = cards.first();
      const tag = firstCard.locator('span').filter({hasText: 'ActTypeAccumulatedRecharge'});
      await expect(tag).toBeVisible();
    });
  });

  describe('累充奖励页签', () => {
    test.skip('页签显示 - 累充奖励页签可见', async () => {
      await wikiPage.clickExecuteCheck();
      await wikiPage.waitForCheckComplete();
      await wikiPage.clickAccumulatedRechargeCard(0);
      await wikiPage.expectAccumulatedRechargeTabVisible();
    });

    test.skip('关联说明 - 显示关联说明提示', async () => {
      await wikiPage.clickExecuteCheck();
      await wikiPage.waitForCheckComplete();
      await wikiPage.clickAccumulatedRechargeCard(0);
      const alert = wikiPage.page.locator('.n-alert').filter({hasText: 'AccumulatedRechargeReward.ActId'});
      await expect(alert).toBeVisible();
    });

    test.skip('奖励表格 - 显示累充档位表格', async () => {
      await wikiPage.clickExecuteCheck();
      await wikiPage.waitForCheckComplete();
      await wikiPage.clickAccumulatedRechargeCard(0);
      await wikiPage.expectAccumulatedRechargeTableHasData();
    });

    test.skip('奖励数据 - 每行包含ID、金额、奖励物品', async () => {
      await wikiPage.clickExecuteCheck();
      await wikiPage.waitForCheckComplete();
      await wikiPage.clickAccumulatedRechargeCard(0);
      const table = wikiPage.getAccumulatedRechargeTable();
      const rows = table.locator('tbody tr');
      const firstRow = rows.first();
      // 验证行内有内容
      const cells = firstRow.locator('td');
      const count = await cells.count();
      expect(count).toBeGreaterThanOrEqual(3);
    });

    test.skip('页签切换 - 切换到基础信息后可切回', async () => {
      await wikiPage.clickExecuteCheck();
      await wikiPage.waitForCheckComplete();
      await wikiPage.clickAccumulatedRechargeCard(0);
      // 切到基础信息
      await wikiPage.page.locator('.n-tabs-tab').filter({hasText: '基础信息'}).click();
      // 切回累充奖励
      await wikiPage.page.locator('.n-tabs-tab').filter({hasText: '累充奖励'}).click();
      await wikiPage.expectAccumulatedRechargeTableHasData();
    });
  });

  describe('打开Excel', () => {
    test.skip('打开Excel按钮 - 累充奖励页签下点击打开Excel', async () => {
      await wikiPage.clickExecuteCheck();
      await wikiPage.waitForCheckComplete();
      await wikiPage.clickAccumulatedRechargeCard(0);
      // 切到累充奖励页签
      await wikiPage.page.locator('.n-tabs-tab').filter({hasText: '累充奖励'}).click();
      // 点击打开Excel按钮
      const btn = wikiPage.getButton('打开Excel');
      await expect(btn).toBeVisible();
    });
  });
});
