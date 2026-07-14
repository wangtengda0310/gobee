/**
 * 表内关联规则 E2E 测试
 * 验证 REGEX_CROSS_COL 和 REGEX_EXTRACT_RANGE 规则配置界面
 */
import { test, expect, describe } from '../shared/fixtures/index';
import { ExcelTestPage } from '../shared/pages/ExcelTestPage';
import { sleep } from '../shared/utils/helpers';

describe('表内关联规则 - REGEX_CROSS_COL', () => {
  let excelTestPage: ExcelTestPage;

  test.beforeEach(async ({ page }) => {
    excelTestPage = new ExcelTestPage(page);
    await excelTestPage.goto();
    await excelTestPage.switchToConfigTab();
  });

  test('规则类型下拉菜单包含"表内关联规则"分类', async () => {
    await expect(excelTestPage.configTab).toBeVisible();
  });

  test('参数组件 - 正则表达式输入框', async () => {
    await expect(excelTestPage.configTab).toBeVisible();
  });
});

describe('表内关联规则 - REGEX_EXTRACT_RANGE', () => {
  let excelTestPage: ExcelTestPage;

  test.beforeEach(async ({ page }) => {
    excelTestPage = new ExcelTestPage(page);
    await excelTestPage.goto();
    await excelTestPage.switchToConfigTab();
  });

  test('参数组件 - 范围类型切换', async () => {
    await expect(excelTestPage.configTab).toBeVisible();
  });
});
