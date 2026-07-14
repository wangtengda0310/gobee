---
name: android-layout-adaptation
description: rain-qa-func(Wails v3 Vue3 Naive UI) 前端移动端布局适配。触发:Android/手机/移动端/窄屏布局适配、sider/抽屉/导航/按钮触控、白边/拖出屏幕/半个字/对人类不可用、PathConfigInput/共享组件窄屏、响应式 vs 触屏适配选型。策略:分层判定(pointer:coarse 交互层 + .is-mobile UA Android 布局层),不做 max-width 响应式。涵盖 sider 折叠/锚点收起/抽屉满屏/PathConfigInput 自适应 + 白边排查 + box-sizing 坑 + Naive UI sider transform 非 overlay。详见 docs/Android-前端适配.md + frontend/docs/layout/android-adaptation.md。
---

# Android 前端布局适配（rain-qa-func）

rain-qa-func 原桌面布局(Wails v3 + Vue3 + Naive UI)适配 Android 手机(触屏 360×640)。**策略:分层判定,A(pointer:coarse) 为主 + 平台检测(.is-mobile),不做 max-width 响应式**(QA 用户高频窄窗 + 目标单一 + 真瓶颈是数据供给)。

## 触发
- Android/手机/移动端布局问题(窄屏溢出/截断/不可用)
- sider/抽屉/导航/按钮触控适配
- "白边"、"拖出屏幕"、"半个字"、"对人类不可用"
- PathConfigInput/共享组件窄屏适配
- 响应式 vs 触屏适配选型

## 适配策略（分层,PC 零影响）

| 层 | 判定 | 触发设备 | 改动 | PC/触屏笔记本影响 |
|---|---|---|---|---|
| **交互层** | `@media(pointer:coarse)` / `matchMedia('(pointer: coarse)')` | 所有触屏 | 按钮增大(themeOverrides)、导航 44px 横滚、抽屉满屏 100vw、overflowX 兜底 | 鼠标 PC 零影响;触屏笔记本按钮变大(无害) |
| **布局层** | `.is-mobile` class(`navigator.userAgent` 含 Android) | **仅 Android 手机** | sider 折叠/overlay、锚点收起、PathConfigInput 自适应 | 任何 PC(含触屏笔记本)零影响 |

**why 分层**:`pointer:coarse` 误判触屏笔记本(CSS-Tricks)。交互增强(按钮大)对触屏笔记本无害,但布局重排(sider→overlay)破坏其桌面多栏 → 布局层用 `.is-mobile`(UA Android)精确锁定。

**why 不用 max-width**:QA 用户高频与游戏/Excel/IDE 并排,窗口常落 800-1200px,max-width 断点会触发 sider 抽屉化/导航变汉堡,破坏桌面肌肉记忆。

## 常见适配模式

| 问题 | 模式 |
|---|---|
| sider 多栏占满窄屏 | `:default-collapsed="isTouchDevice"`(交互层)+ `collapsed-width=50` |
| 锚点栏(120/150px)吃宽度 | `.is-mobile` 下 `v-show="!isMobile"` 收起 |
| 抽屉 >视口,X 屏外 | `.n-drawer-content-wrapper{max-width:100vw}`(**全局非 scoped CSS**,n-drawer teleport body,scoped 的 data-v 不生效) |
| **抽屉 width prop(400/500/700)溢出** ⭐ | **`.n-drawer` 本身也要 `width:100vw!important`**(width prop 设 .n-drawer 元素,仅覆盖 wrapper 不够:drawer 仍 700px,placement=right 时 x=360-700=-340 左半屏外,关闭按钮不可见) |
| PathConfigInput 固定宽溢出 | inline `:wrap="isMobile"` + input-group `flex:1 1 100%` + input `flex:1 1 0;width:auto;minWidth:0` |
| 导航 8 按钮挤窄屏 | `@media(pointer:coarse)` `.idea-icon-button` min44+nowrap+**flex:1 1 auto** + `#layout-header` **flex-wrap:wrap**(换行 2 行全显,非 overflow-x:auto 横滚出屏) |
| **footer(n-layout-footer)多 statistic 溢出** ⭐ | footer 移动端 `overflow-x:auto` + `flex-wrap:nowrap` + 子项 `flex-shrink:0`(内容横滚,footer 本身 360 不超出屏;FooterCaseLogStatistic n-progress min300+用例 min250+错误 min200=750px 典型) |
| 横向白边 | 逐层 scrollWidth 定位(见 [android-debug](../android-debug/SKILL.md) "白边排查")+ 源头 `box-sizing:border-box` |
| **header 换行后 content 高度** ⭐ | `#layout` 改 `display:flex;flex-direction:column`(移动端),`#layout-content` `flex:1 1 0;height:auto!important`(原 height:calc 固定 68px 头尾不够换行后 header) |

## 关键坑

| 坑 | 说明 |
|---|---|
| **Naive UI sider collapse-mode="transform" 非 overlay** | 源码确认 transform 仍 push content(容器 max-width transition),仅内容固定宽防重排。真 overlay 需 `n-drawer` 替代 sider |
| **status-bar box-sizing** | `width:100%`+`padding:0 12px` content-box 溢出 24px → `#layout-footer` sw>cw → `#layout` overflow-x:auto 可横滚 → 左拖出右侧白边。加 `box-sizing:border-box` |
| **n-drawer teleport body** | scoped CSS 不生效,必须全局 `<style>`(非 scoped) |
| **dist embed libwails.so** | 前端改必须 `compile:go:shared`(见 [android-build](../android-build/SKILL.md)),否则 APK 用旧 dist(@media 旧 data-v hash 不匹配) |
| **#layout overflow-x 兜底** | 各页适配无溢出后,`@media(pointer:coarse)` `html/body/#layout{overflow-x:hidden}`(原 auto 允许溢出时横滚出白边) |

## 适配检查清单（改前端后必须）

1. `pnpm run build:dev`(dist)
2. **`compile:go:shared`**(embed dist → libwails.so,ARCH 视目标)
3. `gradlew assembleDebug`
4. `adb install -r` + [android-debug](../android-debug/SKILL.md) CDP 验证(`matchMedia('(pointer:coarse)').matches` + 各层 scrollWidth + 元素 rect)
5. **PC 回归**:`wails3 dev` 确认 pointer:fine + UA 非 Android → 无 .is-mobile,布局不变

详见 [frontend/docs/layout/android-adaptation.md](../../../frontend/docs/layout/android-adaptation.md)(规范 + 各页状态 + 坑表) + [docs/Android-前端适配.md](../../../docs/Android-前端适配.md)(数据供给课题)。

## 自迭代

每次使用后:是否新适配模式/坑?是否新页需适配?如是,更新本 skill + android-adaptation.md,并记 `.learnings/LEARNINGS.md`。
