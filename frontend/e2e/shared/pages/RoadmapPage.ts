/**
 * 路线图抽屉面板 Page Object
 * 对应 src/pages/settings/components/roadmap-panel.vue
 * 路线图已迁移到设置页面抽屉，非独立页面
 */

import { Page, Locator, expect } from '@playwright/test';
import { BasePage, Route, resolveRoute } from './BasePage';
import { sleep } from '../utils/helpers';

/**
 * 路线图抽屉面板 Page Object
 */
export class RoadmapPage extends BasePage {
	// 设置页面元素
	readonly settingsPage: Locator;
	readonly openRoadmapButton: Locator;

	// 抽屉面板元素
	readonly drawer: Locator;
	readonly drawerContent: Locator;
	readonly filterBar: Locator;
	readonly statusSelect: Locator;
	readonly sortSelect: Locator;
	readonly searchInput: Locator;
	readonly submitButton: Locator;
	readonly itemList: Locator;
	readonly emptyState: Locator;
	readonly loadingSpinner: Locator;

	constructor(page: Page) {
		super(page);

		// 设置页面
		this.settingsPage = page.locator('#settings-page');
		this.openRoadmapButton = page.locator('button:has-text("查看开发路线图")');

		// 抽屉面板（n-drawer 渲染后出现在 body 下）
		this.drawer = page.locator('.n-drawer');
		this.drawerContent = page.locator('.n-drawer-content');
		this.filterBar = page.locator('.n-drawer .filter-bar');
		this.statusSelect = page.locator('.n-drawer .filter-bar .n-select').first();
		this.sortSelect = page.locator('.n-drawer .filter-bar .n-select').nth(1);
		this.searchInput = page.locator('.n-drawer .filter-bar input');
		this.submitButton = page.locator('.n-drawer button:has-text("提交新建议")');
		this.itemList = page.locator('.n-drawer .item-list');
		this.emptyState = page.locator('.n-drawer .empty-state');
		this.loadingSpinner = page.locator('.n-drawer .n-spin');
	}

	/**
	 * 导航到设置页面并打开路线图抽屉
	 * 注意：CDP 模式下 page.goto 不会触发 Wails 路由切换，需通过菜单点击导航
	 */
	async goto(): Promise<void> {
		// 先检查当前是否已在设置页面
		const hasSettingsCard = await this.page.locator('.setting-card').count() > 0;
		if (!hasSettingsCard) {
			// 通过导航菜单切换到设置页面
			const settingsBtn = this.page.locator('#layout-header button:has-text("设置")');
			await settingsBtn.click();
			await sleep(800);
		}
	}

	/**
	 * 打开路线图抽屉
	 */
	async openDrawer(): Promise<void> {
		// 如果抽屉已打开则跳过
		const isOpen = await this.drawer.isVisible().catch(() => false);
		if (isOpen) return;

		await this.openRoadmapButton.click();
		await sleep(500);
		await expect(this.drawer).toBeVisible();
	}

	/**
	 * 关闭路线图抽屉
	 */
	async closeDrawer(): Promise<void> {
		const isOpen = await this.drawer.isVisible().catch(() => false);
		if (!isOpen) return;

		// 点击抽屉遮罩层或关闭按钮关闭
		await this.page.evaluate(() => {
			const mask = document.querySelector('.n-drawer-mask');
			if (mask) (mask as HTMLElement).click();
		});
		await sleep(300);
	}

	// ==================== 元素获取 ====================

	/**
	 * 获取路线图项目列表
	 */
	getRoadmapItems(): Locator {
		return this.page.locator('.n-drawer .roadmap-item');
	}

	/**
	 * 获取指定索引的项目
	 */
	getRoadmapItem(index: number): Locator {
		return this.getRoadmapItems().nth(index);
	}

	/**
	 * 获取项目的标题
	 */
	getItemTitle(item: Locator): Locator {
		return item.locator('.item-title');
	}

	/**
	 * 获取项目的状态标签
	 */
	getItemStatusTag(item: Locator): Locator {
		return item.locator('.n-tag');
	}

	/**
	 * 获取项目的投票按钮
	 */
	getItemVoteButtons(item: Locator): Locator {
		return item.locator('.item-actions button');
	}

	/**
	 * 获取项目的评论数
	 */
	getItemCommentCount(item: Locator): Locator {
		return item.locator('.comments');
	}

	// ==================== 交互操作 ====================

	/**
	 * 点击项目查看详情
	 */
	async clickItem(item: Locator): Promise<void> {
		await item.click();
		await sleep(300);
	}

	/**
	 * 点击项目内的投票按钮（支持/反对）
	 */
	async clickVoteButton(item: Locator, type: 'up' | 'down'): Promise<void> {
		const buttons = this.getItemVoteButtons(item);
		const button = type === 'up'
			? buttons.filter({ hasText: /支持|已支持/ }).first()
			: buttons.filter({ hasText: /反对|已反对/ }).first();
		await button.click();
		await sleep(200);
	}

	/**
	 * 选择状态筛选
	 */
	async selectStatusFilter(status: string): Promise<void> {
		await this.page.evaluate((targetStatus) => {
			const selectTrigger = document.querySelector('.n-drawer .filter-bar .n-select .n-base-selection');
			if (selectTrigger) {
				(selectTrigger as HTMLElement).click();
			}
			setTimeout(() => {
				const options = document.querySelectorAll('.n-base-select-option');
				for (const opt of options) {
					if (opt.textContent?.trim() === targetStatus) {
						(opt as HTMLElement).click();
						break;
					}
				}
			}, 100);
		}, status);
		await sleep(400);
	}

	/**
	 * 选择排序方式
	 */
	async selectSortBy(sortType: string): Promise<void> {
		await this.page.evaluate((targetSort) => {
			const selects = document.querySelectorAll('.n-drawer .filter-bar .n-select');
			if (selects.length >= 2) {
				const sortSelect = selects[1].querySelector('.n-base-selection');
				if (sortSelect) {
					(sortSelect as HTMLElement).click();
				}
			}
			setTimeout(() => {
				const options = document.querySelectorAll('.n-base-select-option');
				for (const opt of options) {
					if (opt.textContent?.trim() === targetSort) {
						(opt as HTMLElement).click();
						break;
					}
				}
			}, 100);
		}, sortType);
		await sleep(400);
	}

	/**
	 * 输入搜索关键词
	 */
	async search(keyword: string): Promise<void> {
		await this.searchInput.fill(keyword);
		await sleep(300);
	}

	/**
	 * 点击提交新建议按钮
	 */
	async clickSubmitSuggestion(): Promise<void> {
		await this.page.evaluate(() => {
			const btns = document.querySelectorAll('.n-drawer button');
			for (const b of btns) {
				if (b.textContent?.includes('提交新建议')) {
					b.click();
					return 'clicked';
				}
			}
			return 'not found';
		});
		await sleep(500);
	}

	// ==================== 弹窗操作 ====================

	/**
	 * 获取提交弹窗
	 */
	getSubmitModal(): Locator {
		return this.page.locator('.n-modal:has-text("提交新功能建议")');
	}

	/**
	 * 在提交弹窗中填写标题
	 */
	async fillSuggestionTitle(title: string): Promise<void> {
		const modal = this.getSubmitModal();
		const input = modal.locator('input[placeholder="请输入功能标题"]');
		await input.fill(title);
		await sleep(100);
	}

	/**
	 * 在提交弹窗中填写描述
	 */
	async fillSuggestionDescription(description: string): Promise<void> {
		const modal = this.getSubmitModal();
		const textarea = modal.locator('textarea');
		await textarea.fill(description);
		await sleep(100);
	}

	/**
	 * 在提交弹窗中选择优先级
	 */
	async selectSuggestionPriority(priority: '低' | '中' | '高'): Promise<void> {
		const modal = this.getSubmitModal();
		const radio = modal.locator(`.n-radio:has-text("${priority}")`);
		await radio.click();
		await sleep(100);
	}

	/**
	 * 点击提交弹窗的提交按钮
	 */
	async clickModalSubmit(): Promise<void> {
		const modal = this.getSubmitModal();
		const submitBtn = modal.locator('button:has-text("提交")');
		await submitBtn.click();
		await sleep(300);
	}

	/**
	 * 点击提交弹窗的取消按钮
	 */
	async clickModalCancel(): Promise<void> {
		const modal = this.getSubmitModal();
		const cancelBtn = modal.locator('button:has-text("取消")');
		await cancelBtn.click();
		await sleep(200);
	}

	/**
	 * 获取详情弹窗（排除提交弹窗）
	 */
	getDetailModal(): Locator {
		// 详情弹窗包含"功能描述"标题，提交弹窗包含"提交新功能建议"标题
		return this.page.locator('.n-modal').filter({ hasText: '功能描述' }).filter({ hasNotText: '提交新功能建议' });
	}

	/**
	 * 在详情弹窗中填写评论
	 */
	async fillComment(content: string): Promise<void> {
		const modal = this.getDetailModal();
		const textarea = modal.locator('textarea');
		await textarea.fill(content);
		await sleep(100);
	}

	/**
	 * 在详情弹窗中点击发送评论
	 */
	async clickSendComment(): Promise<void> {
		const modal = this.getDetailModal();
		const sendBtn = modal.locator('button:has-text("发送")');
		await sendBtn.click();
		await sleep(300);
	}

	/**
	 * 在详情弹窗中点击投票
	 */
	async clickDetailVote(type: 'up' | 'down'): Promise<void> {
		const modal = this.getDetailModal();
		const button = type === 'up'
			? modal.locator('button:has-text("支持")').first()
			: modal.locator('button:has-text("反对")').first();
		await button.click();
		await sleep(200);
	}

	/**
	 * 关闭详情弹窗
	 */
	async closeDetailModal(): Promise<void> {
		await this.page.evaluate(() => {
			const btns = document.querySelectorAll('.n-modal button');
			for (const b of btns) {
				if (b.textContent?.trim() === '关闭') {
					(b as HTMLButtonElement).click();
					return 'clicked';
				}
			}
			return 'not found';
		});
		await sleep(200);
	}

	// ==================== 验证方法 ====================

	/**
	 * 验证抽屉已打开且内容可见
	 */
	async expectPageLoaded(): Promise<void> {
		await expect(this.drawer).toBeVisible();
		await expect(this.filterBar).toBeVisible();
	}

	/**
	 * 验证项目列表不为空
	 */
	async expectItemsVisible(): Promise<void> {
		const items = this.getRoadmapItems();
		await expect(items.first()).toBeVisible();
	}

	/**
	 * 验证项目数量
	 */
	async expectItemCount(count: number): Promise<void> {
		const items = this.getRoadmapItems();
		await expect(items).toHaveCount(count);
	}

	/**
	 * 验证空状态显示
	 */
	async expectEmptyStateVisible(): Promise<void> {
		await expect(this.emptyState).toBeVisible();
	}

	/**
	 * 验证提交弹窗显示
	 */
	async expectSubmitModalVisible(): Promise<void> {
		await expect(this.getSubmitModal()).toBeVisible();
	}

	/**
	 * 验证提交弹窗隐藏
	 */
	async expectSubmitModalHidden(): Promise<void> {
		await expect(this.getSubmitModal()).toBeHidden();
	}

	/**
	 * 验证详情弹窗显示
	 */
	async expectDetailModalVisible(): Promise<void> {
		await expect(this.getDetailModal()).toBeVisible();
	}

	/**
	 * 验证项目包含指定文本
	 */
	async expectItemContainsText(index: number, text: string): Promise<void> {
		const item = this.getRoadmapItem(index);
		await expect(item).toContainText(text);
	}

	// ==================== 等待和清理 ====================

	/**
	 * 等待加载完成
	 */
	async waitForLoading(): Promise<void> {
		try {
			await this.loadingSpinner.waitFor({ state: 'visible', timeout: 1000 });
			await this.loadingSpinner.waitFor({ state: 'hidden', timeout: 10000 });
		} catch {
			// 没有加载状态，直接返回
		}
	}

	/**
	 * 清理页面状态 —— 测试后调用，避免状态污染后续测试
	 * 使用 JS evaluate 强制关闭所有弹窗，避免 Playwright click 被遮挡
	 */
	async cleanup(): Promise<void> {
		// 强制关闭所有 modal（提交弹窗、详情弹窗等）
		await this.page.evaluate(() => {
			// 点击所有 modal 的关闭按钮（n-base-close 类）
			const closeBtns = document.querySelectorAll('.n-modal-container button.n-base-close');
			closeBtns.forEach((btn) => (btn as HTMLElement).click());
			// 也找文本为"关闭"或"取消"的按钮
			const allBtns = document.querySelectorAll('.n-modal-container button');
			allBtns.forEach((btn) => {
				const text = btn.textContent?.trim();
				if (text === '关闭' || text === '取消') {
					(btn as HTMLElement).click();
				}
			});
			// 触发 Escape 关闭任何剩余浮层
			document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }));
		});
		await sleep(500);

		// 清空搜索框
		await this.page.evaluate(() => {
			const inputs = document.querySelectorAll('.n-drawer .filter-bar input');
			inputs.forEach((input) => {
				const el = input as HTMLInputElement;
				el.value = '';
				el.dispatchEvent(new Event('input', { bubbles: true }));
				el.dispatchEvent(new Event('change', { bubbles: true }));
			});
		});
		await sleep(200);

		// 重置状态筛选为"全部状态"
		const statusText = await this.statusSelect.textContent().catch(() => '');
		if (statusText && !statusText.includes('全部状态')) {
			await this.selectStatusFilter('全部状态');
		}

		// 重置排序为"按时间排序"
		const sortText = await this.sortSelect.textContent().catch(() => '');
		if (sortText && !sortText.includes('按时间排序')) {
			await this.selectSortBy('按时间排序');
		}

		// 再次关闭下拉菜单
		await this.page.evaluate(() => {
			document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }));
		});
		await sleep(300);
	}
}
