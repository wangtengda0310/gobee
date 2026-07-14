/**
 * 路线图抽屉面板 E2E 测试
 * 对应 src/pages/settings/components/roadmap-panel.vue
 * 路线图已迁移到设置页面抽屉，非独立页面
 */

import { test, expect, describe } from '../shared/fixtures';

describe('路线图抽屉面板测试', () => {
	test.beforeEach(async ({ roadmapPage }) => {
		// 导航到设置页面
		await roadmapPage.goto();
		// 打开路线图抽屉
		await roadmapPage.openDrawer();
		await roadmapPage.waitForLoading();
	});

	test.afterEach(async ({ roadmapPage }) => {
		// 只清理弹窗状态，不关闭抽屉（beforeEach 会处理导航和抽屉状态）
		await roadmapPage.cleanup();
	});

	describe('页面加载', () => {
		test('抽屉打开 - 显示筛选栏和列表', async ({ roadmapPage }) => {
			await roadmapPage.expectPageLoaded();
		});

		test('抽屉打开 - 默认显示项目列表', async ({ roadmapPage }) => {
			await roadmapPage.expectItemsVisible();
		});

		test('筛选栏 - 包含状态选择器', async ({ roadmapPage }) => {
			await expect(roadmapPage.statusSelect).toBeVisible();
		});

		test('筛选栏 - 包含排序选择器', async ({ roadmapPage }) => {
			await expect(roadmapPage.sortSelect).toBeVisible();
		});

		test('筛选栏 - 包含搜索框', async ({ roadmapPage }) => {
			await expect(roadmapPage.searchInput).toBeVisible();
		});

		test('提交按钮 - 提交新建议按钮可见', async ({ roadmapPage }) => {
			await expect(roadmapPage.submitButton).toBeVisible();
		});
	});

	describe('筛选和排序', () => {
		test('状态筛选 - 按规划中筛选', async ({ roadmapPage }) => {
			await roadmapPage.selectStatusFilter('规划中');
			const items = roadmapPage.getRoadmapItems();
			const count = await items.count();
			if (count > 0) {
				await expect(items.first().locator('.n-tag')).toContainText('规划中');
			}
		});

		test('状态筛选 - 按已完成筛选', async ({ roadmapPage }) => {
			await roadmapPage.selectStatusFilter('已完成');
			const items = roadmapPage.getRoadmapItems();
			const count = await items.count();
			if (count > 0) {
				await expect(items.first().locator('.n-tag')).toContainText('已完成');
			}
		});

		test('状态筛选 - 全部状态显示所有项目', async ({ roadmapPage }) => {
			await roadmapPage.selectStatusFilter('全部状态');
			await roadmapPage.expectItemsVisible();
		});

		test('搜索 - 输入关键词过滤', async ({ roadmapPage }) => {
			await roadmapPage.search('战斗测试');
			const items = roadmapPage.getRoadmapItems();
			const count = await items.count();
			if (count > 0) {
				await expect(items.first()).toContainText('战斗');
			}
		});

		test('搜索 - 无结果时显示空状态', async ({ roadmapPage }) => {
			await roadmapPage.search('不存在的功能xyz123');
			await roadmapPage.expectEmptyStateVisible();
		});
	});

	describe('项目详情', () => {
		test('点击项目 - 打开详情弹窗', async ({ roadmapPage }) => {
			const items = roadmapPage.getRoadmapItems();
			await expect(items.first()).toBeVisible();
			await roadmapPage.clickItem(items.first());
			await roadmapPage.expectDetailModalVisible();
		});

		test('详情弹窗 - 显示功能描述', async ({ roadmapPage }) => {
			const items = roadmapPage.getRoadmapItems();
			await roadmapPage.clickItem(items.first());
			const modal = roadmapPage.getDetailModal();
			await expect(modal.locator('h4:has-text("功能描述")')).toBeVisible();
		});

		test('详情弹窗 - 显示投票区域', async ({ roadmapPage }) => {
			const items = roadmapPage.getRoadmapItems();
			await roadmapPage.clickItem(items.first());
			const modal = roadmapPage.getDetailModal();
			await expect(modal.locator('h4:has-text("投票")')).toBeVisible();
		});

		test('详情弹窗 - 显示评论区', async ({ roadmapPage }) => {
			const items = roadmapPage.getRoadmapItems();
			await roadmapPage.clickItem(items.first());
			const modal = roadmapPage.getDetailModal();
			await expect(modal.locator('h4:has-text("评论区")')).toBeVisible();
		});

		test('详情弹窗 - 点击关闭按钮关闭', async ({ roadmapPage }) => {
			const items = roadmapPage.getRoadmapItems();
			await roadmapPage.clickItem(items.first());
			await roadmapPage.closeDetailModal();
			// 使用 waitFor 等待弹窗消失，避免严格模式问题
			await roadmapPage.getDetailModal().waitFor({ state: 'hidden', timeout: 5000 });
		});
	});

	describe('投票功能', () => {
		// 投票会修改持久化数据，在共享 WebView2 实例上无法隔离，skip
		test.skip('列表页投票 - 点击支持按钮', async () => {});
		test.skip('列表页投票 - 点击反对按钮', async () => {});
		test.skip('详情页投票 - 支持功能', async () => {});
	});

	describe('评论功能', () => {
		// 评论会修改持久化数据，在共享 WebView2 实例上无法隔离，skip
		test.skip('添加评论 - 在详情页添加评论', async () => {});

		test('添加评论 - 空评论不发送', async ({ roadmapPage }) => {
			const items = roadmapPage.getRoadmapItems();
			await roadmapPage.clickItem(items.first());
			const modal = roadmapPage.getDetailModal();
			const sendBtn = modal.locator('button:has-text("发送")');
			await expect(sendBtn).toBeDisabled();
		});
	});

	describe('提交新建议', () => {
		// 提交会修改持久化数据，在共享 WebView2 实例上无法隔离，skip
		test.skip('打开提交弹窗', async () => {});
		test.skip('提交弹窗 - 填写并提交', async () => {});

		test('提交弹窗 - 取消按钮关闭弹窗', async ({ roadmapPage }) => {
			await roadmapPage.clickSubmitSuggestion();
			await roadmapPage.expectSubmitModalVisible();
			await roadmapPage.clickModalCancel();
			await roadmapPage.expectSubmitModalHidden();
		});

		test('提交弹窗 - 空标题不能提交', async ({ roadmapPage }) => {
			await roadmapPage.clickSubmitSuggestion();
			await roadmapPage.fillSuggestionDescription('描述内容');
			const modal = roadmapPage.getSubmitModal();
			const submitBtn = modal.locator('button:has-text("提交")');
			await expect(submitBtn).toBeDisabled();
			await roadmapPage.clickModalCancel();
		});
	});

	describe('导航集成', () => {
		test('从设置页面打开路线图抽屉', async ({ page, roadmapPage }) => {
			// 关闭当前抽屉
			await roadmapPage.closeDrawer();
			// 验证设置页面仍然可见
			await expect(roadmapPage.settingsPage).toBeVisible();
			// 重新打开抽屉
			await roadmapPage.openDrawer();
			// 验证抽屉内容
			await roadmapPage.expectPageLoaded();
		});
	});
});
