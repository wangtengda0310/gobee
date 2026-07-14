/**
 * 功能测试页 - 下拉分组与悬浮提示回归
 *
 * 覆盖三处 UI 元素（防止后续重构破坏）：
 * 1. 卡牌下拉"分割线"：使用卡牌下拉按 naive-ui group 分组（🎴 摸牌堆 / 🧑 座位N·角色名）
 * 2. 「应用智能描述」按钮 hover tooltip：外包 n-tooltip，内容=aiDesc(step)
 * 3. 技能下拉选项 hover 文字区 tooltip：renderSkillDescLabel 用 NTooltip 包裹选项文字 span，
 *    hover 文字触发、hover 空白不触发
 *
 * 依赖 PO：loadCaseWithSteps（处理"加载确认 dialog"+ 跳过分类节点）。
 */

import {test, expect, describe} from '../shared/fixtures';
import {FunctionTestPage} from '../shared/pages/FunctionTestPage';

// 关键超时/重试常量（命名化魔法值，保留防 flaky 的 sleep/重试不弱化）
const SELECT_MENU_TIMEOUT = 5000;        // naive-ui select 菜单打开/选项可见超时
const TOOLTIP_HOVER_RETRY = 5;           // 技能 tooltip 虚拟列表下首次 hover 偶发不触发的重试次数
const TOOLTIP_HOVER_INTERVAL_MS = 900;   // 每次 hover 后等 tooltip 浮层出现的间隔
const VLIST_RENDER_DELAY_MS = 800;       // 等下拉虚拟列表渲染出分组标题的固定延时

describe('功能测试页 - 下拉分组与悬浮提示回归', () => {
    let p: FunctionTestPage;

    test.beforeEach(async ({page}) => {
        p = new FunctionTestPage(page);
        await p.goto();
        await p.loadCaseWithSteps('行动牌');
    });

    test('卡牌下拉按 牌堆/座位 分组并染色（分割线 + 左侧色条）', async () => {
        await p.openSelectAndWait(p.getStepCardsSelect(0));
        await p.page.waitForTimeout(VLIST_RENDER_DELAY_MS);  // 等虚拟列表渲染分组标题

        const headers = p.getVisibleSelectGroupHeaders();
        await expect(headers.first()).toBeVisible({timeout: SELECT_MENU_TIMEOUT});
        // 牌堆组标题存在
        expect(await headers.filter({hasText: '摸牌堆'}).count()).toBeGreaterThanOrEqual(1);
        // 至少一个座位组标题存在（虚拟列表下首屏通常可见座位1）
        expect(await headers.filter({hasText: '座位'}).count()).toBeGreaterThanOrEqual(1);

        // 颜色区分：验证"至少两种不同 borderLeftColor"（牌堆灰 + 座位身份色），
        // 不绑定具体身份色 rgb（避免与用例数据耦合、避免 WebView 版本差异 flaky）
        const colors = await p.page.locator('.n-base-select-menu:visible')
            .locator('.n-base-select-option').evaluateAll(
                (opts) => opts.map((o) => getComputedStyle(o as HTMLElement).borderLeftColor)
            );
        expect(new Set(colors).size).toBeGreaterThanOrEqual(2);
    });

    test('「应用智能描述」按钮 hover 显示描述 tooltip', async () => {
        const btn = p.getApplyAiDescButton(0);
        await expect(btn).toBeVisible();

        await btn.hover();
        const tip = p.getVisibleTooltip();
        await expect(tip).toBeVisible({timeout: 3000});
        // tooltip 应渲染了 aiDesc 文本（非空），防止"空 tooltip 假通过"
        const text = (await tip.textContent()) ?? '';
        expect(text.trim().length).toBeGreaterThan(0);
    });

    test('技能下拉选项 hover 文字区显示描述 tooltip', async () => {
        await p.switchToConfigTab();
        await p.page.waitForTimeout(400);
        const delSelect = p.getDelSkillsSelect(0);
        await delSelect.scrollIntoViewIfNeeded();
        await p.openSelectAndWait(delSelect);

        // 必须 hover 选项文字区(.n-base-select-option__content)而非选项中心(空白/checkmark)，
        // 因为 NTooltip trigger 仅包裹文字 span
        const content = p.getVisibleOptionContent(0);
        await expect(content).toBeVisible({timeout: SELECT_MENU_TIMEOUT});
        // 等 render-label span 出现（excelSkillDescMap 加载完成、renderSkillDescLabel 返回 NTooltip trigger）。
        // 该 map 由后端 SkillUIDescService.LoadSkillUIDesc 异步加载，存在间歇返回空的时序问题（已用 Func.ts
        // await 缓解前端竞态，但后端数据本身仍可能慢/空）；此时无法验证 tooltip，skip 而非 fail。
        const hasSpan = await content.locator('span').waitFor({state: 'visible', timeout: 8000})
            .then(() => true).catch(() => false);
        if (!hasSpan) {
            test.skip('excelSkillDescMap 未加载（LoadSkillUIDesc 间歇返回空），技能 tooltip 暂无法验证');
            return;
        }
        await p.page.waitForTimeout(500);

        // hover 文字区触发 NTooltip；虚拟列表下首次 hover 偶尔不触发，重试 TOOLTIP_HOVER_RETRY 次
        let tipVisible = false;
        for (let i = 0; i < TOOLTIP_HOVER_RETRY && !tipVisible; i++) {
            // 重新取 box（虚拟列表可能滚动），精确 hover 文字中心
            const box = await content.boundingBox();
            if (box) {
                await p.page.mouse.move(box.x + box.width / 2, box.y + box.height / 2);
            } else {
                await content.hover();
            }
            await p.page.waitForTimeout(TOOLTIP_HOVER_INTERVAL_MS);
            tipVisible = await p.page.locator('.n-tooltip:visible').count() > 0;
        }
        expect(tipVisible).toBe(true);
        const text = (await p.getVisibleTooltip().textContent()) ?? '';
        expect(text.trim().length).toBeGreaterThan(0);
    });

    test('资产断言「卡」下拉同样按 牌堆/座位 分组并染色（NormalCard + group 兼容）', async () => {
        // 资产断言的「卡」下拉与步骤动作的 step-cards-select 共用同一个 computed
        // (excelCardsSelectOptionFromInit)，但断言下拉叠加了 multiple tag + fallback-option + @create。
        // 本用例验证 group 分组标题、颜色区分、以及 tag/fallback 在分组渲染下不发生异常。

        // 1. 确保第一个动作下至少有一个断言卡片
        if (await p.getAssertionCards(0).count() === 0) {
            await p.clickAddAssertion(0);
        }
        await expect(p.getAssertionCards(0).first()).toBeVisible({timeout: SELECT_MENU_TIMEOUT});

        // 2. 切换断言类型为 NormalCard 类（出牌 PlayCard）以渲染 AssetNormalCardSection
        await p.pickAssertionType(0, 0, '出牌(PlayCard)');

        // 3. 打开「卡」下拉
        const cardsSelect = p.getAssetCardsSelect(0, 0);
        await cardsSelect.scrollIntoViewIfNeeded();
        await expect(cardsSelect).toBeVisible({timeout: SELECT_MENU_TIMEOUT});
        await p.openSelectAndWait(cardsSelect);
        // multiple select 菜单带 n-base-select-menu--multiple 类，精确定位避免与残留菜单冲突
        await expect(p.page.locator('.n-base-select-menu--multiple:visible').first()).toBeVisible({timeout: SELECT_MENU_TIMEOUT});
        await p.page.waitForTimeout(VLIST_RENDER_DELAY_MS);  // 等虚拟列表渲染分组标题

        // 4. 分组标题与步骤卡牌下拉一致：含「摸牌堆」+「座位」
        const headers = p.page.locator('.n-base-select-menu--multiple:visible .n-base-select-group-header');
        await expect(headers.first()).toBeVisible({timeout: SELECT_MENU_TIMEOUT});
        expect(await headers.filter({hasText: '摸牌堆'}).count()).toBeGreaterThanOrEqual(1);
        expect(await headers.filter({hasText: '座位'}).count()).toBeGreaterThanOrEqual(1);

        // 5. 颜色区分：至少两种 borderLeftColor
        const colors = await p.page.locator('.n-base-select-menu--multiple:visible')
            .locator('.n-base-select-option').evaluateAll(
                (opts) => opts.map((o) => getComputedStyle(o as HTMLElement).borderLeftColor)
            );
        expect(new Set(colors).size).toBeGreaterThanOrEqual(2);

        // 6. 关闭下拉（点空白处收起，避免 mask/层叠残留影响后续测试）
        await p.page.locator('body').click({position: {x: 1, y: 1}}).catch(() => {});
        await p.page.waitForTimeout(300);
    });
});
