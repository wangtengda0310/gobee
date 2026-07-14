/**
 * 通用布局测试
 * 测试应用的通用布局组件：顶部导航栏、页面路由切换、状态栏
 */

import { test, expect, describe } from '../shared/fixtures';
import { resolveRoute } from '../shared/pages/BasePage';

/**
 * 导航按钮配置
 * 根据 normal-layout 文档定义
 */
const NAV_BUTTONS = [
  { label: '设置', pageContent: '飞书通知' },
  { label: 'AI助手', pageContent: 'AI 助手' },
  { label: '战斗测试', pageContent: '加载用例' },
  { label: '配表测试', pageContent: '加载配表' },
  { label: '武将资源检查', pageContent: '配表位置' },
  { label: '武将Wiki检查', pageContent: '执行检查' },
  { label: '路线图', pageContent: '路线图' },
];

describe('通用布局测试', () => {
  /**
   * 测试前准备：导航到首页
   */
  test.beforeEach(async ({ page }) => {
    await page.goto(resolveRoute(page, '/'));
    await page.waitForLoadState('networkidle');
  });

  /**
   * 顶部导航栏渲染验证
   * 验证所有导航按钮都正确渲染
   */
  test('顶部导航栏渲染验证', async ({ page }) => {
    // 验证 header 存在
    const header = page.locator('#layout-header');
    await expect(header).toBeVisible();

    // 验证所有导航按钮都存在
    for (const { label } of NAV_BUTTONS) {
      const button = page.locator(`#layout-header button:has-text("${label}")`);
      await expect(button).toBeVisible();
    }

    // 验证导航按钮数量正确
    const buttons = page.locator('#layout-header button');
    await expect(buttons).toHaveCount(NAV_BUTTONS.length);
  });

  /**
   * 页面路由切换测试 - 设置页面
   * 通过检查页面内容而非URL来验证路由切换
   */
  test('页面路由切换 - 设置页面', async ({ page }) => {
    await page.locator('#layout-header button:has-text("设置")').click();
    // 等待页面内容更新
    await page.waitForTimeout(500);
    // 验证设置页面特有内容出现
    await expect(page.locator('#layout-content')).toContainText('飞书通知');
  });

  /**
   * 页面路由切换测试 - AI助手页面
   */
  test('页面路由切换 - AI助手页面', async ({ page }) => {
    // 先导航到其他页面
    await page.locator('#layout-header button:has-text("设置")').click();
    await page.waitForTimeout(500);

    // 点击 AI助手 按钮
    await page.locator('#layout-header button:has-text("AI助手")').click();
    await page.waitForTimeout(500);
    // 验证 AI 助手页面内容
    await expect(page.locator('#layout-content')).toContainText('AI 助手');
  });

  /**
   * 页面路由切换测试 - 战斗测试页面
   */
  test('页面路由切换 - 战斗测试页面', async ({ page }) => {
    await page.locator('#layout-header button:has-text("战斗测试")').click();
    await page.waitForTimeout(500);
    // 验证战斗测试页面内容
    await expect(page.locator('#layout-content')).toContainText('加载用例');
  });

  /**
   * 页面路由切换测试 - 配表测试页面
   */
  test('页面路由切换 - 配表测试页面', async ({ page }) => {
    await page.locator('#layout-header button:has-text("配表测试")').click();
    await page.waitForTimeout(500);
    // 验证配表测试页面内容
    await expect(page.locator('#layout-content')).toContainText('加载配表');
  });

  /**
   * 页面路由切换测试 - 武将资源检查页面
   */
  test('页面路由切换 - 武将资源检查页面', async ({ page }) => {
    await page.locator('#layout-header button:has-text("武将资源检查")').click();
    await page.waitForTimeout(500);
    // 验证武将资源检查页面内容
    await expect(page.locator('#layout-content')).toContainText('配表位置');
  });

  /**
   * 页面路由切换测试 - 武将Wiki检查页面
   */
  test('页面路由切换 - 武将Wiki检查页面', async ({ page }) => {
    await page.locator('#layout-header button:has-text("武将Wiki检查")').click();
    await page.waitForTimeout(500);
    // 验证 Wiki 检查页面内容
    await expect(page.locator('#layout-content')).toContainText('执行检查');
  });

  /**
   * 状态栏版本信息显示
   * 验证状态栏存在且包含基本信息
   */
  test('状态栏版本信息显示', async ({ page }) => {
    // 验证 footer 存在
    const footer = page.locator('#layout-footer');
    await expect(footer).toBeVisible();

    // 验证状态栏内容
    const statusBar = page.locator('.status-bar');
    await expect(statusBar).toBeVisible();

    // 验证应用名称显示
    await expect(footer).toContainText('rain-qa-func');
  });

  /**
   * 导航按钮悬停效果
   * 验证导航按钮的交互状态
   */
  test('导航按钮悬停效果', async ({ page }) => {
    const firstButton = page.locator('#layout-header button').first();

    // 悬停前验证
    await expect(firstButton).toBeVisible();

    // 悬停
    await firstButton.hover();

    // 验证按钮仍然可见（没有 JS 错误导致按钮消失）
    await expect(firstButton).toBeVisible();
  });

  /**
   * 主内容区域渲染
   * 验证 RouterView 正确渲染内容
   */
  test('主内容区域渲染', async ({ page }) => {
    const content = page.locator('#layout-content');
    await expect(content).toBeVisible();

    // 验证内容区域不为空
    const contentHtml = await content.innerHTML();
    expect(contentHtml.length).toBeGreaterThan(0);
  });

  /**
   * 页面布局结构验证
   * 验证 header、content、footer 的层级关系
   */
  test('页面布局结构验证', async ({ page }) => {
    // 验证三个主要区域都存在
    const header = page.locator('#layout-header');
    const content = page.locator('#layout-content');
    const footer = page.locator('#layout-footer');

    await expect(header).toBeVisible();
    await expect(content).toBeVisible();
    await expect(footer).toBeVisible();

    // 验证顺序：header -> content -> footer
    const headerBox = await header.boundingBox();
    const contentBox = await content.boundingBox();
    const footerBox = await footer.boundingBox();

    // 验证 header 在 content 上方
    expect(headerBox?.y).toBeLessThan(contentBox?.y || 0);

    // 验证 content 在 footer 上方
    expect(contentBox?.y).toBeLessThan(footerBox?.y || 0);
  });

  /**
   * 连续路由切换测试
   * 验证多次切换页面不会出现问题
   */
  test('连续路由切换测试', async ({ page }) => {
    const routes = [
      { label: '设置', content: '飞书通知' },
      { label: 'AI助手', content: 'AI 助手' },
      { label: '战斗测试', content: '加载用例' },
      { label: '配表测试', content: '加载配表' },
    ];

    for (const { label, content } of routes) {
      // 等待任何可能存在的 drawer 关闭
      const drawerMask = page.locator('.n-drawer-mask');
      if (await drawerMask.isVisible()) {
        await drawerMask.click({ position: { x: 10, y: 10 } });
        await page.waitForTimeout(300);
      }

      await page.locator(`#layout-header button:has-text("${label}")`).click();
      await page.waitForTimeout(800);
      await expect(page.locator('#layout-content')).toContainText(content);
    }
  });
});
