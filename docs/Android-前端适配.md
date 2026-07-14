# Android 前端/数据适配

> rain-qa-func(Wails v3) 跑在 Android 上时，前端页面与"数据供给"的适配现状、难点与候选方案。基于 2026-07-09 在 `Small_Phone`(x86_64, 视口 360×640) 模拟器的 CDP 实测。调试方法见 [Android 运行时调试](Android-运行时调试.md)。

## 一、前端体检实测（4 个核心页）

通过 CDP（[`cdp_eval.ps1`](../build/android/scripts/cdp_eval.ps1)）逐页扫描主导航触达的 4 个核心页：

| 页面 | 横向溢出 | 表格 | 触控目标(<36px) |
|---|---|---|---|
| 战斗测试 | 否 | 0 | 13/13 |
| 配表测试 | 否 | 0（无数据，空页） | 13/13 |
| 武将Wiki检查 | 否 | 0（无数据，空页） | 12/12 |
| Proto测试 | 否（文档级） | 3 表，1 个 975px | 21/21 |

**关键测量值**：
- 视口 `360×640`，dpr `2`，基础字号 `14px`
- 4 页 `documentElement.scrollWidth == 360`，**均无横向溢出**（框架弹性好，不是硬编码桌面布局）
- Proto 表格 `n-data-table` 975px：父容器链 `wrapper(ovX:hidden,336px) → header(ovX:scroll)` —— Naive UI 表格**内部横滚**，内容可见、不裁剪
- 触控目标：主导航 `button.idea-icon-button` 25–28px、页内按钮最小 18px、多数 28–34px（**普遍 <44px 推荐值**）
- 主导航是**自定义 `button.idea-icon-button`**（图标+文字），**不是** Naive UI `n-menu`（排查导航点击时勿认成 n-menu）

## 二、适配全景与优先级

| 课题 | 性质 | 影响 | 工作量 |
|---|---|---|---|
| **数据供给** | 跨层（后端路径 + 前端 UI + Android 权限） | 配表/Wiki/游戏配置等"需数据"的功能在手机喂不进数据 → **可用性瓶颈** | 大，需规划 |
| **触控目标偏小** | 纯前端、全局 | 操作偏困难 | 小 |
| 表格横滚体验 | 次要 | 宽表查看不便（但可用） | 小 |

**核心判断**：后端代码可用（已证，见 [Android 运行时调试](Android-运行时调试.md)）、表格组件可用（已证），**手机上真正的瓶颈是"怎么把 Excel/资源喂给应用 + 路径配置"**，而非代码兼容或表格。

## 三、数据供给课题（关键难点）

### 现状事实（实测）

1. **配置文件位置依赖 CWD**：[appconfig.init](../backend/pkg/common/appconfig/appconfig.go:30) 用 `os.Getwd()` + `.rain-qa-func.json` 定位配置。Android c-shared 进程的 CWD 不确定（`adb shell ls /proc/<pid>/cwd` 因权限看不到），且 `/data/data/com.wails.app/files/` 下**不存在** `.rain-qa-func.json`。
2. **无外部存储权限**：应用上下文 `ls /sdcard` → `Permission denied`。推到 `/sdcard` 的数据应用读不到。
3. **待查**：首页能显示"1 条用例"（`TestCaseService.LoadTestCaseList` 成功返回 1 条），但 `files/` 下无 `cases` 目录——用例在 Android 的实际读取位置未确认（可能 CWD 指向了某可达目录，或运行时生成）。理清这点是数据供给方案的前置。

### 影响

配表检查 `CheckAllExcelRules`、游戏配置 `InitExcel`、Wiki 检查等功能都依赖"配置文件里存的 Excel/资源目录路径"。在 Android 上配置写入位置不明 + 无法读外部存储 → **这些功能缺数据无法端到端验证/使用**。

### 候选方案（待选型）

| 方案 | 做法 | 优缺 |
|---|---|---|
| A. 配置改写到应用私有目录 | 把 `appconfig` 的路径从 `CWD/.rain-qa-func.json` 改为 `os.UserConfigDir()` 或 Android `files/` 下；提供"导入 Excel"把数据拷进私有目录 | 不需额外权限；但要解决"数据怎么进私有目录"（adb push 调试 / 应用内导入） |
| B. 申请外部存储权限 | Android `MANAGE_EXTERNAL_STORAGE` 或运行时权限，读 `/sdcard` | 用户可放数据到 `/sdcard`；但权限敏感、审核受限 |
| C. SAF 文件选择 | 前端路径配置用 Android Storage Access Framework 让用户选目录 | 体验好但实现复杂（需 Wails/原生桥） |
| D. 内置示例数据 | APK 内打包一份测试配表（`go:embed`），提供"加载示例" | 零配置可演示，但不解决用户真实数据 |

**建议**：先做 A（配置路径去 CWD 化，写入私有目录）+ 理清 cases 读取机制，这是最小可用前提；B/C 作为用户体验增强。

### 实施进展（2026-07-10）

**步骤 1（配置去 CWD 化）已完成并验证**：根因为 Android c-shared 进程 CWD=`/`（实测 readlink）且无 HOME/TMPDIR（实测 environ），`os.UserConfigDir/UserHomeDir` 全失败。修复 [appconfig.configDir](../backend/pkg/common/appconfig/appconfig.go) —— GOOS=android 时用 `/data/data/<pkg>/files`（包名取自 `/proc/self/cmdline`，回退硬编码）。模拟器验证：关闭配置抽屉触发 `SaveConfig`，`/data/data/com.wails.app/files/.rain-qa-func.json` 成功创建（含 `function_test` section），桌面行为不变（`go build ./...` 通过）。

**待续（步骤 2，数据导入）**：配置默认值仍是相对路径（`cases/fight_cases`、`../rain-robot/...`），Android CWD=`/` 下不可达；需解决"Excel/资源如何进手机"——存储权限 / 应用内导入 / SAF，以及默认路径的 Android 适配。

**步骤 2 端到端打通（2026-07-10）**：adb push resources 到私有目录 + 配置指向，game 成功加载（GetAllHeroCfg 返回 219 武将）。两个关键卡点：
1. **.bytes 版本须与 go.mod rain-robot proto 匹配**：本机 `D:\work\rain-robot`（05-11）.bytes 与 go.mod rain-robot（05-27）proto 不一致 → `proto: wrong wireType` → game 加载失败。正确来源：`go mod 缓存 rain-robot@<go.mod版本>/.../resources`（与 proto 同版本）。
2. **serverLog 启动期日志丢失**：启动期 log（app.Run 前 Emit）前端连前丢失，无法排查。已加历史缓存 + `serverLogHistoryRequest` 事件回放解决（[serverlog/service.go](../backend/pkg/settings/serverlog/service.go)），InitExcel 延后到 serverLog 后配套。

**真机导入仍待续**：当前仅 adb push 可达私有目录；真机用户方案（外部存储权限 / 应用内导入 / wails 补 file chooser）未实现。注意 wails android **未实现 onShowFileChooser**，`<input type=file>` 不可用。

## 四、触控目标

全局图标/按钮触控目标下限 `44px`（[Material 推荐最小触控尺寸](https://m3.material.io/foundations/accessible-design/accessibility-basics)）。

### 实施进展（2026-07-10）

用 `@media (pointer: coarse)` / `window.matchMedia('(pointer: coarse)')` 判定触屏——基于**物理输入设备**（触屏=coarse / 鼠标=fine），非 UA 字符串、非屏幕宽度，**PC 鼠标任何窗口宽度都不触发**。

| 组件 | 改动 | 状态 |
|---|---|---|
| 主导航 `idea-icon-button` | [App.vue](../frontend/src/App.vue) `@media (pointer: coarse)` min-height/min-width 44px | ✅ 模拟器实测（28→44px） |
| 标准 n-button（配置卡/对话框） | App.vue themeOverrides `{Button:{heightMedium:'40px',heightSmall:'38px',heightTiny:'34px'}}` | ✅ 编译进 dist |
| 横向 `n-menu` 工具栏（function-test 加载/保存用例等） | 被 `n-layout-header`(34px) 压缩，不受 Button themeOverrides 影响；需 Menu itemHeight + header 高度 | ❌ 待续 |

**关键坑**：改前端后必须 `wails3 task android:compile:go:shared`——前端 dist 经 `go:embed` 进 libwails.so，仅 `vite build + gradle` 不够（dist 没重新 embed，CSS 旧 data-v hash 不匹配，@media 不生效，实测踩坑）。

## 五、实施检查清单（适配改动后）

1. 改前端 → `pnpm exec vite build --minify false --mode development` 重建 dist
2. 改 Go → `wails3 task android:compile:go:shared` 重建 libwails.so
3. `gradlew.bat assembleDebug` 重打 APK
4. `adb install -r` + 用 [`cdp_eval.ps1`](../build/android/scripts/cdp_eval.ps1) 复测触控目标/溢出

（构建全流程见 [Android APK 构建](Android-APK构建.md)）

## 相关文档

- [Android 运行时调试](Android-运行时调试.md) — CDP/logcat/截图/端口 + 后端可用性结论
- [Android APK 构建](Android-APK构建.md) — 构建/安装/重打流程
