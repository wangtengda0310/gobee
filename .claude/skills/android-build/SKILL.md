---
name: android-build
description: 构建 rain-qa-func(Wails v3) Android APK。触发:构建/重建/装机 Android APK、compile:go:shared、gradle assembleDebug、adb install、arm64 真机/x86_64 模拟器 APK、fat APK、发新版本到 itsnot.fun。涵盖 vite→compile:go:shared→gradle→install 完整链 + NDK git bash 坑(非 PowerShell)+ .cmd CC + go-task shell + arm64/x86_64 + publish-update。详见 docs/Android-APK构建.md。
---

# Android APK 构建（rain-qa-func）

构建 rain-qa-func(Wails v3 桌面应用) 为 Android APK。**前端 dist 经 go:embed 进 libwails.so**,所以前端改必须 `compile:go:shared`(仅 vite+gradle 不够)。

## 触发
- 构建/重建/装机 Android APK("装到模拟器/手机"、"发新版本")
- `compile:go:shared` 报错(NDK/uname/host OS)
- arm64 真机 / x86_64 模拟器 / fat APK
- 发布更新到 itsnot.fun

## 完整构建链（worktree 根）

```bash
# 1. 前端 dist
cd frontend && pnpm run build:dev && cd ..

# 2. Go .so(必须 git bash!PowerShell 丢 env,见坑1)
export PATH="/c/Users/<u>/go/bin:/c/Users/<u>/AppData/Local/Programs/Git/usr/bin:$PATH"
export ANDROID_HOME="D:/Android/Sdk" ANDROID_NDK_HOME="D:/Android/Sdk/ndk/26.3.11579264"
wails3 task android:compile:go:shared              # 默认 x86_64(模拟器)
wails3 task android:compile:go:shared ARCH=arm64   # 真机 arm64(fat APK 需两个都编)

# 3. Gradle APK(PowerShell/cmd,不能用 task assemble:apk,见坑5)
powershell.exe -NoProfile -Command "cd build/android; \$env:JAVA_HOME='C:/Program Files/Android/Android Studio/jbr'; .\gradlew.bat assembleDebug"
# 产物: build/android/app/build/outputs/apk/debug/app-debug.apk

# 4. 装机+重启
ADB=/d/Android/Sdk/platform-tools/adb.exe
$ADB install -r build/android/app/build/outputs/apk/debug/app-debug.apk
$ADB shell am force-stop com.wails.app; sleep 2
$ADB shell monkey -p com.wails.app -c android.intent.category.LAUNCHER 1
```

## 关键坑（速查）

| # | 现象 | 根因 | 解法 |
|---|---|---|---|
| 1 | `compile:go:shared` 报 `NDK not found`(PowerShell,env/PATH 都对) | PowerShell `$env:ANDROID_NDK_HOME` 经 wails3→go-task→bash 链路**未继承** | **git bash 直接跑**(同 shell 继承 env),非 PATH 问题 |
| 2 | `uname not found` / `Unsupported host OS` | PATH 缺 Git\usr\bin | PATH 加 `/c/Users/<u>/AppData/Local/Programs/Git/usr/bin` |
| 3 | 真机装失败 `INSTALL_FAILED_NO_MATCHING_ABIS` | APK 缺 arm64(默认只编 x86_64) | `compile:go:shared ARCH=arm64` + gradle(fat APK arm64+x86_64) |
| 4 | 前端改了 APK 没变 | dist 经 go:embed 进 libwails.so,仅 vite+gradle 不重 embed | 改前端后必须 `compile:go:shared` |
| 5 | `assemble:apk` 报 `'.' 不是命令` | go-task Windows 把该内联脚本交 cmd,`./gradlew` 不识别 | Gradle 直接 `gradlew.bat`(PowerShell/cmd) |
| 6 | wails 模块/CLI/runtime 不一致(android 编译错) | 三者版本脱节 | 对齐 alpha2.117(CLI+模块) + @wailsio/runtime alpha.97 |
| 7 | `linux_cgo.c` 找 gtk | alpha2.109 漏 !android 守卫 | 用 alpha2.117+(上游已修),勿降级 |

完整 10 坑表见 [docs/Android-APK构建.md](../../../docs/Android-APK构建.md)。

## 发新版本到 itsnot.fun（自动更新源）

```bash
# 1. 递增 build.gradle versionCode + versionName
# 2. 构建(上)
# 3. publish(算 SHA256 + 生成 latest.json + scp)—— 注意 PS5.1 GBK 坑,脚本需 ASCII
pwsh build/android/scripts/publish-update.ps1 -Notes "更新说明"
# 或手动(避脚本坑): sha256sum APK → 生成 latest.json → scp APK+json 到 itsnot.fun
```

详见 [Android-自动更新.md](../../../docs/Android-自动更新.md)。

## 自迭代

每次使用后:是否新坑(版本/工具链/架构)?是否 build 流程变化?如是,更新本 skill + docs/Android-APK构建.md,并记 `.learnings/LEARNINGS.md`。
