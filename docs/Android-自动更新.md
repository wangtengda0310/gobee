# Android APK 自动更新

> rain-qa-func(Wails v3) Android 端自动更新方案（**已实施** 2026-07-10）。桌面端用 Wails 自带 updater,Android 自建。
> 相关:[Android APK 构建](Android-APK构建.md) | [Android 运行时调试](Android-运行时调试.md)

## 一、Wails v3 updater 不适用 Android（机制层面）

Wails v3 自带 updater（`pkg/updater` + `providers/github`）挂在 `*application.App` 单例（`app.Updater.Init/CheckAndInstall`），机制是 **"Swap the running binary and relaunch"**（替换运行中可执行文件 + 重启）。

**不适用 Android 的理由**（机制层面，非 bindings）：
1. Android **无"可执行文件"概念**,应用更新 = APK 重装（系统安装器）,与 binary swap 是两套机制
2. Android 命中 `updater_notdarwin.go` + `helper_unix.go`（编译通过）,但 spawn helper 替换 libwails.so 无意义 + 沙箱受限
3. 教程/guide 全篇桌面 build(darwin/windows/linux),无 Android 章节

⚠️ **纠错**：`frontend/bindings/.../updater/` 的"空 Updater 类"是正常的（updater 不经前端 bindings,由后端 `app.Updater` 触发）。"bindings 空"**不是**"不支持"的标志。

来源:[Self-Updating Wails App 教程](https://v3.wails.io/tutorials/04-self-update-a-wails-app/) | [Updater Guide](https://v3.wails.io/guides/updater/)

## 二、Android 自升级标准方案（非 Play 分发）

rain-qa-func 自分发 APK,用业界标准"下载 APK + 触发系统安装器"（FB/Amazon 先例）。关键约束：
- **"自动"上限**：自动检查+下载+提示,但**安装必须用户确认**（非 rooted 不可静默）
- **Android 8+ 双重授权**：Manifest `REQUEST_INSTALL_PACKAGES` + 用户在"安装未知应用"授权本应用
- **安全**：下载后校验 SHA256（防劫持）

参考:[Unity 自升级讨论](https://discussions.unity.com/t/solved-self-updating-on-android-without-google-store/667547) | [Stack Overflow](https://stackoverflow.com/questions/3987534/self-upgrading-own-apk-via-net-programmatically-on-android)

## 三、本项目实施（已完成）

### 架构
```
前端 settings "检查更新"按钮
  │ window.wails.getAppVersionCode()  (Java 桥读 build.gradle versionCode,单一版本源)
  ▼
Go UpdateService.CheckUpdate(vc) → GET https://itsnot.fun/rain-qa-func/latest.json → 对比 versionCode
  │ 有新版 → 前端显示"下载并安装"
  ▼
Go DownloadApk(info) → files/updates/ (边下边算 SHA256 + Event "updateProgress" 推进度)
  │ 下载完 + SHA256 校验(若 latest.json 提供 sha256)
  ▼
window.wails.installApk(path) (Java 桥 ACTION_VIEW + FileProvider) → 系统安装器(用户确认)
```

### 代码索引
| 层 | 文件 | 职责 |
|---|---|---|
| 服务端 | itsnot.fun nginx `/root/itsnot.fun/nginx/html/rain-qa-func/` | 托管 latest.json + APK(HTTPS letsencrypt) |
| Manifest | [AndroidManifest.xml](../build/android/app/src/main/AndroidManifest.xml) | `REQUEST_INSTALL_PACKAGES` 权限 |
| FileProvider | [file_paths.xml](../build/android/app/src/main/res/xml/file_paths.xml) | `files-path updates/` |
| Java 桥 | [WailsJSBridge.java](../build/android/app/src/main/java/com/wails/app/WailsJSBridge.java) | `installApk` + `getAppVersionCode` @JavascriptInterface |
| Go | [update/service.go](../backend/pkg/settings/update/service.go) | `CheckUpdate` + `DownloadApk`(SHA256 + updateProgress Event) |
| 注册 | [wails.go](../cmd/rain-qa-func/wails.go) | `updateService.InitWithApp` + `RegisterService` |
| 前端逻辑 | [use-update.ts](../frontend/src/pages/settings/composables/use-update.ts) | `checkUpdate` + `downloadAndInstall` |
| UI | [settings/index.vue](../frontend/src/pages/settings/index.vue) | "应用更新"卡片(仅 Android 显示,`isAndroid=!!window.wails?.installApk`) |
| 上传脚本 | [publish-update.ps1](../build/android/scripts/publish-update.ps1) | build 后算 SHA256 + 生成 latest.json + scp |

### 版本协议 latest.json
```json
{
  "versionCode": 2,           // 整数,递增,对比用(源自 build.gradle)
  "versionName": "1.1",       // 展示名
  "apkUrl": "https://itsnot.fun/rain-qa-func/rain-qa-func-2.apk",
  "sha256": "<hex>",          // APK SHA256(空则跳过校验)
  "releaseNotes": "更新说明"
}
```

### 下载优化（断点续传 + 分块 + 超时 + 重试，2026-07-13）
原全量 `http.Get` 三大问题:**无超时**(连接挂起卡死)、**无续传**(中断重头)、**无重试**(弱网一次失败即放弃)。重构 DownloadApk 为:

| 优化 | 实现 | 解决 |
|---|---|---|
| **分块下载** | 2MB/块,逐块 Range 请求 | 弱网容错,小块易成功 |
| **单块超时** | `context.WithTimeout(30s/块)` | 防卡死(无数据 30s 判超时) |
| **断点续传** | `.part` 文件存已下部分,Range `bytes=offset-` 续传 | 中断后不重头 |
| **失败重试** | 单块失败重试 5 次,线性退避(1s/2s/3s...) | 弱网抖动恢复 |
| **SHA256** | 下载完整后整体重算 `.part`(非边下边算) | 避免 hash 状态序列化,简单可靠 |
| **系统干净** | `cleanOldUpdates` 清旧 versionCode APK + .part,仅留当前版本 | 不累积旧版本 |

代码:[update/service.go](../backend/pkg/settings/update/service.go) `DownloadApk` + `downloadChunk` + `headApk` + `cleanOldUpdates` + `sha256File`。nginx 已支持 Range(`Accept-Ranges: bytes`,206 Partial Content)。

日志(serverLog 面板):`[update] 断点续传: 已下 X/Y`(恢复)、`[update] 块下载重试 N`(重试)、`[update] 下载完成`。

### 待优化:系统 DownloadManager(根本解,前端页面适配完成后讨论) ⏸️

当前应用内 Go 分块下载(versionCode 6)虽加了断点续传+超时+重试,但用户实测仍"意外中断"。根因有二:
- **A 后台冻结**:Android Doze/后台限制冻结 WebView 进程 → Go 下载 goroutine 被中断(切应用/熄屏后中断)
- **B 前台也断**:Go http 栈在 Android 网络栈上的稳定性问题(亮屏看着也中断)

而**浏览器下载稳**(用户验证),因用系统 DownloadManager(系统级进程,不受应用后台限制 + 成熟续传/重试/通知栏)。

**结论**:应用内下载在 Android 架构上不可靠(进程级限制),需换系统级机制。

**方案(待讨论实施)**:换系统 DownloadManager(Java 桥)
- Java `WailsJSBridge` 加 `downloadApk(url)`:`DownloadManager.enqueue` + 下载完成广播接收器 → `installApk`
- Go `DownloadApk` 简化(仅查版本)或前端直接调 Java 桥(下载下放给系统)
- 进度:系统通知栏天然显示;或轮询 `DownloadManager.query`(`COLUMN_BYTES_DOWNLOADED`)
- 系统干净:`DownloadManager` 下到指定路径(私有目录),安装后清

**代价**:进度条不如应用内精细(系统百分比非实时,但有通知栏兜底);Java 代码增(广播接收器需注册/反注册)。

**优先级**:前端页面适配工作完成后再讨论实施。

### 操作流程（发布新版本）
1. 改 `build/android/app/build.gradle` 递增 `versionCode` + 改 `versionName`
2. 构建 APK（git bash）：`wails3 task android:compile:go:shared` + `gradlew.bat assembleDebug`（详见 [Android-APK构建.md](Android-APK构建.md)）
3. 发布：`pwsh build/android/scripts/publish-update.ps1 -Notes "更新说明"`（自动算 SHA256 + 生成 latest.json + scp 到 itsnot.fun）
4. 用户端：应用启动/手动 → 检查更新 → 发现新版 → 下载 → 系统安装器确认

### 端到端验证（2026-07-10，完整升级链）
- 服务端 itsnot.fun latest.json versionCode=2 + 真 APK（61MB + SHA256）
- 模拟器装 versionCode=1（旧版）→ 检查更新 → 发现 versionCode=2
- 下载 61MB ~130s（updateProgress 事件 7%→100%，~0.5MB/s）+ SHA256 校验通过
- installApk 触发:`ACTION_VIEW dat=content://com.wails.app.fileprovider/... typ=application/vnd.android.package-archive` → 系统包安装器（com.google.android.packageinstaller）
- Android 8+ 首次:Play Protect 扫描 → 授权 → 安装
- **APK 升级成功:`versionCode 1→2, versionName 1.0→1.1`,启动后 `getAppVersionCode=2`** ✅

### 真机 arm64 验证（2026-07-10,生产可用确认）
- fat APK(arm64-v8a + x86_64)发布到 itsnot.fun
- **真机(arm64)实测**:装 versionCode 3 → 设置"应用更新"→ 检查更新 → 发现 versionCode 4 → 下载 60MB → installApk → 授权"安装未知应用"(com.wails.app)→ 系统安装器 → **成功升级 versionCode 4**
- 真机 arm64 .so 运行正常 + installApk(FileProvider + ACTION_VIEW) + 首次授权流程 全链通过 ✅

## 四、证书续期（letsencrypt，itsnot.fun）

**背景**:itsnot.fun 原 DigiCert 证书 2026-02-04 过期,Go net/http 验证证书失败(`x509: certificate has expired`)。curl `-sk`/WebFetch 忽略证书,掩盖了问题。

**修复**(ssh root@itsnot.fun):
```bash
yum install -y epel-release certbot
certbot certonly --webroot -w /root/itsnot.fun/nginx/html -d itsnot.fun \
  --non-interactive --agree-tos --register-unsafely-without-email
# 替换 nginx 容器证书(挂载 host /root/itsnot.fun/nginx/conf/conf.d/)
cp /etc/letsencrypt/live/itsnot.fun/fullchain.pem /root/itsnot.fun/nginx/conf/conf.d/itsnot.fun.pem
cp /etc/letsencrypt/live/itsnot.fun/privkey.pem   /root/itsnot.fun/nginx/conf/conf.d/itsnot.fun.key
podman exec nginx nginx -s reload
```
**自动续期 hook** `/etc/letsencrypt/renewal-hooks/deploy/nginx-itsnot.sh`:续期后自动 cp + reload。
有效期 90 天,certbot 后台自动续期。webroot 模式(nginx 80 提供 `/.well-known/`,不停服)。

## 五、Wails updater guide 借鉴（对比与可选增强）

读 [Updater Guide](https://v3.wails.io/guides/updater/) 后对比。updater 全桌面(无 Android 章节),其设计可借鉴:

| updater 机制 | 本项目对应 | 借鉴价值 |
|---|---|---|
| Provider 接口(Check/Download) | UpdateService.CheckUpdate/DownloadApk(简化直 HTTP) | 未来多源(GitLab Release/备用源)可重构 Provider 链 |
| **Verification Ed25519 签名** | DownloadApk SHA256 digest-only | **可选升级**:Ed25519 签名防篡改(需 wails3 updater genkey + publish 脚本签名) |
| Events(CheckStarted/DownloadProgress/Verifying/Error 等) | 仅 updateProgress | **可选扩展**:加 Error/DownloadComplete 事件提升诊断 |
| State machine(checking/downloading/verifying/...) | 简单 checking/downloading | 可扩展(前端状态更丰富) |
| WindowNone + 自定义 UI | settings "应用更新"卡片 | 已用(等同 headless + 自定义 UI) |
| Skip version | 无 | **可选**:用户跳过某版 |
| helper swap(binary rename) | Java installApk(APK 安装) | **不适用**(Android 无 binary 概念,用系统安装器) |

**当前 MVP 评估**:功能完整(检查 + 下载 + SHA256 校验 + 安装)。签名/事件扩展/skip 作为后续优化。

**关键差异**:updater swap(替换运行中 binary + relaunch)是桌面机制。Android 用系统安装器重装 APK(用户确认),不能复用 swap。

## 六、进度

- [x] Wails updater 调查(不适用 Android)
- [x] Android 自升级方案调研(搜索)
- [x] 服务端 itsnot.fun nginx 托管
- [x] Manifest 权限 + FileProvider paths
- [x] Java installApk + getAppVersionCode
- [x] Go UpdateService + 前端 UI + 上传脚本
- [x] 证书续期(letsencrypt + 自动 hook)
- [x] 端到端 CDP 验证(CheckUpdate 通,btnCount=2 发现新版)
- [x] **完整升级链验证**(versionCode 1→2,下载 61MB + SHA256 + installApk + 系统安装器,启动 getAppVersionCode=2 确认)
- [ ] 可选:Ed25519 签名 / 事件扩展 / skip version
