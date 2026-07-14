# Android 移动端布局适配规范

> 指导 rain-qa-func 前端在 Android(触屏)的布局适配。全局体检/数据供给见 [docs/Android-前端适配.md](../../../../docs/Android-前端适配.md)，构建/调试见 [docs/Android-运行时调试.md](../../../../docs/Android-运行时调试.md)。

## 一、适配策略（2026-07-10 三方分析综合决策）

经架构成本 / 产品 UX / 行业实践三方分析，采用 **A（pointer:coarse）为主 + 平台检测（`.is-mobile`）锁定布局层**，**不做 max-width 响应式断点**。

### 分层判定

| 层 | 判定信号 | 触发设备 | 改动内容 | PC / 触屏笔记本影响 |
|---|---|---|---|---|
| **交互增强** | `@media (pointer: coarse)` | 所有触屏 | 按钮增大(themeOverrides 40px)、导航 44px+横滚、抽屉满屏 100vw、overflowX 兜底 | 鼠标 PC 零影响；触屏笔记本按钮变大（无害） |
| **布局重排** | `.is-mobile` class（UA 含 Android） | **仅 Android 手机** | sider→overlay 浮层、锚点栏收起、PC-only 区隐藏 | 任何 PC（含触屏笔记本）零影响 |

### 为什么不全做 max-width 响应式
1. **PC 窄窗破坏**：QA 用户高频与游戏/Excel/IDE 并排，窗口常落 800–1200px，max-width 断点会触发 sider 抽屉化/导航变汉堡，破坏桌面肌肉记忆
2. **目标单一**：只适配 Android 手机，max-width 多断点优势用不上
3. **真瓶颈是数据供给**：布局再好，喂不进 Excel 也空（见 docs/Android-前端适配.md）
4. **E2E 翻倍**：双布局（桌面/窄屏）测试矩阵膨胀

### 为什么分两层（pointer + .is-mobile）
`pointer:coarse` 会误判**触屏笔记本**（CSS-Tricks 指出：pointer 只描述主输入设备）。交互增强（按钮大）对触屏笔记本无害，但布局重排（sider→overlay）会破坏其桌面多栏 → 布局层改用 `.is-mobile`（UA 含 Android）精确锁定真移动设备。

## 二、判定实现

### 交互层（CSS @media，[App.vue](../../../src/App.vue)）
```css
@media (pointer: coarse) {
  .idea-icon-button { min-height:44px; min-width:44px; flex-shrink:0; white-space:nowrap; width:auto; }
  #layout-header { overflow-x:auto; flex-wrap:nowrap; }
  /* 非 scoped <style>:n-drawer teleport 到 body,scoped 的 data-v 不生效 */
  .n-drawer-content-wrapper { max-width:100vw !important; width:100vw !important; }
  html, body, #layout { overflow-x:auto !important; }  /* 兜底 */
}
```
```ts
// Naive UI themeOverrides(按钮高度),桌面 undefined 完全不变
const isTouchDevice = window.matchMedia('(pointer: coarse)').matches
const themeOverrides = isTouchDevice ? { Button: { heightMedium:'40px', heightSmall:'38px', heightTiny:'34px' } } : undefined
```

### 布局层（JS，App.vue 入口挂 `.is-mobile`）
```ts
// UA 含 Android → 移动端布局(仅真移动设备,触屏笔记本不触发)。作为 pointer 的精确补充。
const isMobile = typeof navigator !== 'undefined' && /android/i.test(navigator.userAgent)
if (isMobile) document.documentElement.classList.add('is-mobile')
```
组件内读：`const isMobile = document.documentElement.classList.contains('is-mobile')`

## 三、各页适配

### sider 多栏页（function-test / excel-test）
| 层 | 改动 | 状态 |
|---|---|---|
| 交互层 | sider `:default-collapsed="isTouchDevice"` 折叠到 50px | ✅ 已做（CDP 验证无溢出，够用） |
| 布局层(.is-mobile) | sider 折叠 50 已够。**注**：Naive UI `collapse-mode="transform"` 经源码（[layout-sider.cssr.mjs](../../../../frontend/node_modules/naive-ui/es/layout/src/styles/layout-sider.cssr.mjs)）确认**非真 overlay**——sider 容器 max-width transition 仍 push content，transform 仅让内容固定宽度防重排。真 overlay 需 `n-drawer` 替代 sider（sider 树/搜索内容搬 drawer + v-if 二选一，工作量大），**推迟**。当前优先收 content 内锚点栏（收益 >> 50px） | ⏸ 推迟 |

### 锚点栏页（宽度"小偷"）
| 页 | 锚点宽度 | 布局层(.is-mobile) | 状态 |
|---|---|---|---|
| function-test steps-panel | 120px（[index.vue:169](../../../src/pages/function-test/index.vue)） | 触屏收起（顶部下拉/FAB） | 🔲 待做 |
| hero-wiki-check | anchor-container 150px（[index.vue:147](../../../src/pages/hero-wiki-check/index.vue)） | 触屏隐藏/移入抽屉 | 🔲 待做 |
| activity-wiki-check | 同 hero-wiki | 同 | 🔲 待做 |

### shared 组件（PathConfigInput）

[PathConfigInput](../../../src/shared/components/path-config-input/index.vue) inline 模式原 `:wrap="false"` + `input-width` 固定（如 wiki 传 280px），两个输入框并排 ~600px > 360 溢出——既撑爆配置区（第二个输入框屏外），又触发 `overflow-x:auto` 兜底让横向可拖出白边（问题根因）。

移动端（`.is-mobile`）改为：`:wrap="isMobile"`（换行，两个输入框上下各占一行）+ 输入框 `flex:1 1 0; width:auto; minWidth:0`（自适应）。PC `isMobile=false` 保持 `:wrap=false` + 固定宽，完全不变。

涉及页：hero-wiki-check / activity-wiki（input-width 280px）、excel-test、hero-voice-resource-check。CDP 实测（2026-07-10）：input1W=input2W=193，cardW=360，scrollW=360 无溢出。

### PC-only 标签（重编辑，手机只读）
用例配置 / 用例步骤 / 配表规则 / Proto Payload 编辑 —— 360px 装不下重编排，即使响应式也不可用。**定位：手机看结果/监控/轻操作，编排用 PC**。文档标注，不做全响应式。

## 四、已适配状态（2026-07-10）

| 页 | 交互层(pointer) | 布局层(.is-mobile) | 状态 |
|---|---|---|---|
| 全局 App.vue | 导航/抽屉/overflowX/themeOverrides | .is-mobile 挂载(UA Android) | ✅交互 / ✅布局(CDP:cls=is-mobile,coarse=true,vw=360) |
| function-test | sider 折叠 50 | steps 锚点(120px)收起 | ✅交互 / ✅布局(CDP:siderCollapsed=true,siderW=50,contentW=310,scrollW=360 无溢出) |
| excel-test | sider 折叠 50 | sider overlay 推迟 | ✅交互 / ⏸推迟 |
| hero-wiki-check | — | 锚点收起 | ✅布局(CDP:mainW=360 全宽,scrollW=360 无溢出) |
| activity-wiki | — | 锚点收起 | ✅布局(同 hero-wiki 结构,isMobile 逻辑同) |
| proto-test | — | — | ✅ 表格 975px Naive UI 内部横滚可用 |
| settings | 抽屉满屏 | — | ✅ |

## 五、适配检查清单（改前端后必须）
1. `pnpm exec vite build --minify false --mode development`（dist）
2. **`wails3 task android:compile:go:shared`**（dist 经 `go:embed` 进 libwails.so，**仅 vite+gradle 不够**）
3. `gradlew.bat assembleDebug`（APK）
4. `adb install -r` + [`cdp_eval.ps1`](../../../../build/android/scripts/cdp_eval.ps1) 验证：
   - `matchMedia('(pointer: coarse)').matches`（交互层）
   - `document.documentElement.className`（布局层，确认含 `is-mobile`）
   - 目标元素 `getBoundingClientRect()`（sider 宽/抽屉宽/按钮文字 scrollWidth≤clientWidth）
5. **PC 回归**：`wails3 dev` 确认 pointer:fine + UA 非 Android → 无 `.is-mobile`、布局完全不变

## 六、关键坑

| 坑 | 说明 |
|---|---|
| **dist embed libwails.so** | 前端改必须 `compile:go:shared`，否则 APK 用旧 dist（@media 旧 data-v hash 不匹配，不生效）—— 实测踩坑 |
| **n-drawer teleport body** | scoped CSS 不生效，必须全局 `<style>` |
| **n-layout position:absolute** | 脱离 body flow，body overflowX 兜底对 n-layout 内多栏无效 → 必须 sider 折叠/overlay |
| **Naive UI sider transform 非 overlay** | 源码确认 `collapse-mode="transform"` 仍 push content（容器 max-width transition），真 overlay 需 `n-drawer` 替代 sider |
| **PathConfigInput 固定宽溢出** | inline `:wrap=false` + `input-width` 固定（280px）×2 > 视口，撑爆配置区（第二个输入框屏外）。移动端改 `:wrap` + `flex` 自适应 |
| **#layout 横滚白边（status-bar padding）** | `.status-bar` `width:100%`+`padding:0 12px` content-box 溢出 24px → `#layout-footer` sw(384)>cw(360) → `#layout` `overflow-x:auto` 可横滚 → 用户**左拖出右侧白边**（只左拖出白边、右拖恢复是典型 scrollLeft 现象）。修：`.status-bar` 加 `box-sizing:border-box` + 各页适配无溢出后 `#layout` `overflow-x:auto`→`hidden`（导航 8 按钮横滚由 `#layout-header` 自身 `overflow-x:auto` 独立处理，不受影响） |
| **pointer 触屏笔记本误判** | 交互层无害，**布局层必须用 `.is-mobile`（UA Android）而非 pointer** |
| **go.mod rain-robot .bytes 版本** | 须与 proto 对齐（见 LRN-20260710-002）；go.mod 升级后须重 push 匹配 resources |
| **serverLog 启动期日志** | app.Run 前 Emit 前端连前丢失，已加历史回放（`serverLogHistoryRequest`） |
