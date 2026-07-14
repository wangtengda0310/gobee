# Android 适配任务复盘（2026-07-10）

> rain-qa-func(Wails v3 桌面 QA 工具)扩展 Android 端的适配任务复盘。聚焦踩坑根因、可复用经验、技能提取建议。
> 相关功能/构建/调试文档见 [Android-APK构建](Android-APK构建.md) / [Android-运行时调试](Android-运行时调试.md) / [Android-前端适配](Android-前端适配.md) / [Android-自动更新](Android-自动更新.md)。

## 一、任务范围与产出

| 方向 | 产出 |
|---|---|
| 前端布局适配 | 分层判定(pointer 交互层 + .is-mobile 布局层)、sider 折叠、锚点收起、PathConfigInput 自适应、白边修复 |
| APK 自动更新 | 自建 HTTP 更新源(itsnot.fun nginx)、Java installApk、Go UpdateService、前端 UI、上传脚本、letsencrypt 证书 |
| 文档/记忆 | 4 篇 Android docs + 4 个 memory 沉淀 + 本复盘 |

worktree-android 分支提交序列:14a3121(布局) → a3ea4d2(docs memory) → 6dae1f0(自动更新调研) → 1447cf7(自动更新实现) → (证书 + 文档 + 复盘)。

## 二、关键决策（含理由）

1. **分层适配:交互层 pointer:coarse + 布局层 .is-mobile(UA Android)**,不做 max-width 响应式
   - 理由:QA 用户高频窄窗(800-1200px)并排工作,max-width 会破坏桌面;目标单一(Android);真瓶颈是数据供给
   - 修正:pointer:coarse 误判触屏笔记本 → 布局层用 .is-mobile(UA)精确锁定

2. **APK 自动更新自建,不用 Wails updater**
   - 理由:Wails updater 机制是 binary swap(桌面专用);Android = APK 重装(系统安装器)
   - 教程/guide 全篇桌面,无 Android 章节;updater_notdarwin.go 命中但 spawn 替换 libwails.so 无意义

3. **更新源用 itsnot.fun 自建 HTTP(非 Play/市场)**
   - 理由:复用现有服务器(ssh 免密)+ nginx(已托管文件)+ 正式证书;零额外基础设施

4. **证书续期用 letsencrypt/certbot(非阿里云控制台)**
   - 理由:免费 + 自动续期(certbot 后台 + deploy-hook);webroot 模式不停 nginx;不依赖云控制台

## 三、踩坑根因与通用教训

### 1. 证书过期被 curl -sk / WebFetch 掩盖 ⭐ 最耗时坑
- **现象**:Go CheckUpdate http.Get 失败,但 curl/WebFetch 正常(HTTPS 200)
- **根因**:itsnot.fun DigiCert 证书 2026-02-04 过期。curl `-sk`/WebFetch 忽略证书;**Go net/http 默认验证证书** → 失败
- **通用教训**:HTTPS 调试时,Go/Android(默认验证证书)与 curl -sk(忽略)行为不同。**证书问题用不带 -k 的 curl + openssl s_client 验证**,别被 -sk 的 200 误导。Go 错误信息("x509: certificate has expired")是关键信号。

### 2. Wails updater "bindings 空 = 不支持" 误判
- **误判**:看 frontend/bindings/.../updater/ 空 Updater 类,推断 Android 不支持
- **实际**:updater 挂 app.Updater(后端触发),不经前端 bindings,bindings 空正常
- **通用教训**:Wails v3 service 方法经 bindings 暴露,但 updater 这类"app 单例 + 后端触发"机制不经 bindings。"bindings 空"不是"不支持"标志,要看机制(本例:binary swap 桌面专用)。

### 3. Naive UI sider collapse-mode="transform" 非 overlay
- **误判**:以为 transform = overlay 浮层(回收 sider 宽度)
- **实际**:源码确认 transform 仍 push content(容器 max-width transition),仅内容固定宽度防重排
- **通用教训**:组件库 prop 名(如 transform)≠ 直觉语义。**改 CSS/布局前看组件源码**(node_modules/naive-ui/es/...),别靠 prop 名猜。

### 4. #layout 横滚白边根因(status-bar box-sizing)
- **现象**:向左拖应用出右侧白边(只左拖、右拖恢复 = scrollLeft)
- **排查**:CDP 测各层 scrollWidth,发现 #layout sw=384(>360),溯源到 #layout-footer sw=384 → status-bar width:100% + padding content-box 溢出 24px
- **修复**:status-bar box-sizing:border-box + 各页适配后 #layout overflow-x:auto→hidden
- **通用教训**:横向白边/可拖出 = 某 overflow-x:auto 元素有子元素溢出。**CDP 逐层测 scrollWidth > clientWidth 定位溢出源**,别只看 documentElement(可能被子 overflow 裁剪掩盖)。

### 5. PathConfigInput 固定宽溢出(第二个输入框屏外)
- **根因**:inline :wrap=false + input-width=280px(×2=560>360)
- **通用教训**:共享组件(input/button)的固定宽(prop 传入)在窄屏可能溢出。移动端用 :wrap + flex 自适应。

### 6. compile:go:shared PowerShell env 丢失
- **现象**:PowerShell 跑 wails3 task compile:go:shared 报 NDK not found(即便 env/PATH 对)
- **根因**:PowerShell $env:ANDROID_NDK_HOME 经 wails3→go-task→bash 链路未继承到 bash 子进程
- **修复**:改 git bash 直接跑(同 shell 继承 env)
- **通用教训**:Windows 多 shell 链路(PowerShell→wails3→go-task→bash)环境变量可能丢。**优先在最终 shell(git bash)直接 export**,避免跨 shell 继承。

### 7. cdp_eval.ps1 -JsFile 中文 GBK 乱码
- **根因**:PowerShell 5.1 按 GBK 读 BOM-less UTF-8,中文字面量损坏
- **修复**:CDP JS 全 ASCII(导航用 menuItems index,非中文文本匹配;var 用 IIFE 包裹)
- **通用教训**:Windows PowerShell 5.1 脚本含中文,默认 GBK 读取。**跨 shell 传中文用文件(UTF-8 BOM)或转 ASCII**(index/英文)。

## 四、方法论收获

1. **CDP 量化验证优于截图**:
   - 适配效果用 getBoundingClientRect() + scrollWidth/clientWidth 量化,而非"看起来对"
   - -Js/-JsFile + cdp_eval.ps1 自动化,ASCII JS 避免 GBK 坑

2. **debug 顺序:CDP(logcat/serverLog)→ 根因,不盲改**:
   - 证书坑:Go log(serverLog 面板)+ bindings methodID(logcat)→ CheckUpdate 被调但 http.Get 失败 → 读完整错误(tls x509 expired)→ 根因
   - 白边:逐层 scrollWidth 定位

3. **文档同步强制**:CLAUDE.md "每次代码改同步文档"。本轮 4 篇 Android docs(构建/调试/适配/更新)+ memory 沉淀,新会话读 docs 不踩坑

## 五、后续优化（未做，记录待续）

| 项 | 优先级 | 说明 |
|---|---|---|
| 真机数据加载(战斗/配表测试) | **高** | Wails Android 无 onShowFileChooser,数据喂不进。方案:go:embed 示例数据 + share intent/SAF 导入。**功能可用性的真瓶颈** |
| ~~自动更新真机下载+安装验证~~ | ✅ 完成 | 模拟器 versionCode 1→2 升级链(2026-07-10) + **真机 arm64 versionCode 3→4 升级(用户实测)**:下载 60MB + SHA256 + installApk + 系统安装器 + 授权,全链通过 |
| Ed25519 签名(防篡改) | 低 | 当前 SHA256 digest-only;加签名需 wails3 updater genkey + publish 签名 |
| 启动自动检查更新 | 低 | 当前手动按钮;CheckInterval 后台检查 |
| 技能整理 | 中 | android-build / android-cdp-debug / android-layout-adaptation 三个 skill(本轮重复工作) |
| n-menu 水平工具栏触控 | 低 | 战斗/配表测试 34px header 菜单压扁,文档标"待续" |

## 六、技能提取建议

本轮重复工作模式适合提取为 skills(.claude/skills/):
1. **android-build**:vite → compile:go:shared(git bash NDK env)→ gradlew → adb install 的完整构建链 + 踩坑(NDK env、go-task shell、.cmd CC)
2. **android-cdp-debug**:cdp_eval.ps1 + ASCII JS 规范(menuItems index/IIFE) + verify 脚本模板 + 各层 scrollWidth 排查法
3. **android-layout-adaptation**:分层判定(pointer + .is-mobile)+ 各页适配模式(sider 折叠/锚点收起/抽屉满屏/PathConfigInput 自适应)+ 白边排查 + box-sizing 坑

这三个 skill 能让后续 Android 工作(以及新会话)直接复用流程,避免重新摸索构建/调试/适配的坑。
