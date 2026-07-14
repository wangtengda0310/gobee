# Android APK 构建指南

> 将 rain-qa-func(Wails v3 桌面应用)构建为 Android APK。本文记录可行的构建流程、必装补丁、踩过的坑与已知限制,供后续复现与维护。
>
> 相关：APK 装好后如何在模拟器/真机调试、后端哪些功能可用，见 [Android 运行时调试](Android-运行时调试.md)。

## 成果与边界

- **能构建出** 可安装的 debug APK,并在 Android 模拟器启动、前端 UI 正常渲染。
- **边界**:APK 仅编了 **x86_64** ABI(适配 Small_Phone 模拟器);手机(arm64)需另编 `arm64-v8a`。Windows 专有后端能力(proto 注入 Unity、注册表、`go-ole` COM 等)在 Android 运行期不可用。
- **运行时桥已通**:前端↔后端 wails 运行时调用正常(422 已修复,需 `@wailsio/runtime@alpha.97+`,见「已知问题」)。

## 环境前提

| 组件 | 要求 | 本机实际 |
|---|---|---|
| Android SDK | platform-tools + `platforms;android-35` + `build-tools;35.0.0` | `D:\Android\Sdk` |
| Android NDK | `26.3.11579264`(wails 测试版本) | `D:\Android\Sdk\ndk\26.3.11579264` |
| JDK | **JDK 17–23**(Gradle/AGP 8.7 不支持 JDK 25) | Android Studio JBR `C:\Program Files\Android\Android Studio\jbr`(21.0.10) |
| Wails CLI + 模块 | **CLI 与模块版本必须对齐**(同为 `alpha2.x`) | CLI `alpha2.117` ↔ 模块 `alpha2.117` |
| `@wailsio/runtime`(npm) | **`alpha.97`**(=npm `latest`;旧 `alpha.78` 会导致 422) | `frontend/package.json` 钉 `3.0.0-alpha.97` |
| 模拟器 AVD | x86_64 system-image | `Small_Phone` |

> ⚠️ 本机默认 `JAVA_HOME` 是 JDK 25(Gradle 会拒),构建时须临时指向 JBR 21。

安装 SDK 组件:
```bash
sdkmanager "platforms;android-35" "build-tools;35.0.0" "ndk;26.3.11579264"
```

## 一次性改动(脚手架)

以下改动已在 worktree 落地,提交后即长期生效:

1. **`build/android/`** — 从 `wails3 init -t vanilla` 模板收割的完整 Gradle 工程 + `Taskfile.yml`(`android:*` 任务)。`wails3 update build-assets` 不含 android,须用 vanilla init 模板获取。
2. **根 [`Taskfile.yml`](../Taskfile.yml)** — `includes` 增加 `android: ./build/android/Taskfile.yml`。
3. **`build/android/app/build.gradle`** — `compileSdk`/`targetSdk` 由 34 改 35(本机装的是 platform-35)。
4. **`build/android/Taskfile.yml` 的 `compile:go:shared`** 补 Windows 适配(见踩坑 3/4):
   - host 探测 `case` 加 `MINGW*|MSYS*|CYGWIN*) HOST_TAG="windows-x86_64"`
   - Windows 下 CC/CXX 须带 `.cmd` 后缀(`CLANG_EXT`)
   - `case "{{.ARCH | default .HOST_ARCH}}"`(原模板只由 build 任务传 ARCH,直接调用会空)

## ✅ linux_cgo 守卫(alpha2.117 上游已修)

旧版 `alpha2.109` 的 `linux_cgo.c` / `linux_cgo.h` 曾漏 `&& !android` 守卫(因「android implies linux」,android 构建会把 GTK 文件拉进来找 `<gtk/gtk.h>` 失败),需手改模块缓存(非复现)。

**`alpha2.117` 已在上游修复**——`linux_cgo.c` 现为 `//go:build linux && !android && !gtk3 && !server`,无需任何补丁。故务必使用 `alpha2.117+`,不要再降到 `alpha2.109`。

## 构建命令(完整序列)

```bash
cd <worktree>
export ANDROID_HOME="D:/Android/Sdk"
export ANDROID_NDK_HOME="D:/Android/Sdk/ndk/26.3.11579264"
export JAVA_HOME="C:/Program Files/Android/Android Studio/jbr"   # Gradle 用,JDK21

# 1. 前端 dist(绕过 vue-tsc,见踩坑 6)
cd frontend && pnpm exec vite build --minify false --mode development && cd ..

# 2. Go → libwails.so(NDK c-shared,首次较慢)
#    ⚠️ 必须在 git bash 跑(PowerShell 调 wails3→go-task→bash 链路会丢 $ANDROID_NDK_HOME,
#       报 "Android NDK not found",见踩坑 10)。wails3 在 ~/go/bin,需加入 PATH。
wails3 task android:compile:go:shared

# 3. Gradle 打包 APK(注意:不能用 task runner 的 assemble:apk,见踩坑 5)
cd build/android && cmd //c gradlew.bat assembleDebug && cd ../..
# 产物: build/android/app/build/outputs/apk/debug/app-debug.apk
```

> go-task 在 Windows 的 shell 选择不稳定(同份 Taskfile,`compile:go:shared` 走 sh、`assemble:apk` 走 cmd),`assemble:apk` 的 `./gradlew` 会在 cmd 下报 `'.' 不是命令`。**Gradle 步骤直接用 `gradlew.bat`(PowerShell 或 cmd)最稳**。

## 安装并启动到模拟器

```bash
adb install -r build/android/app/build/outputs/apk/debug/app-debug.apk
adb shell am start -n com.wails.app/.MainActivity
adb logcat -d | grep -E "Wails|AndroidRuntime"   # 看启动日志
```

## 踩坑清单

| # | 现象 | 根因 | 解法 |
|---|---|---|---|
| 1 | `wails3 update build-assets` 不生成 `build/android/` | 它只重生成平台资产清单(darwin/ios/linux/windows),android gradle 不在其中 | 用 `wails3 init -t vanilla` 模板收割 `build/android/` |
| 2 | android 编译报 `undefined: events.Android`/`iosMethodNames`、`runtime.Core` 参数不足 | **wails 模块版本与 CLI 错配**(模块 alpha.96 太旧,android 代码与核心 API 脱节) | 升级模块到与 CLI 同代(当前 `alpha2.117`);升级后桌面 `go vet` 需复核 |
| 3 | `fatal error: 'gtk/gtk.h' file not found`(`linux_cgo.c`) | `alpha2.109` 的 `linux_cgo.c/.h` 漏 `&& !android`,android(implies linux)拉进 GTK | **`alpha2.117` 上游已修**;勿降级到 alpha2.109(见「linux_cgo 守卫」) |
| 4 | `Unsupported host OS` 或 CC 找不到 | 官方 `compile:go:shared` 只支持 Darwin/Linux;Windows NDK clang 是 `.cmd` | 补 Windows host 分支 + CC 带 `.cmd` |
| 5 | `assemble:apk` 报 `'.' 不是内部或外部命令` | go-task 在 Windows 把该内联脚本交给 cmd.exe,`./gradlew` 不识别 | Gradle 步骤改用 `gradlew.bat` 直接跑 |
| 6 | 前端 `vue-tsc` 报 `RunRobotTest Expected 16 arguments, got 15` | **项目既有 bug**:`Func.ts` 未跟上后端 `RunRobotTest` 新增的第 16 参 `maxTimeoutPerCase`([services.go:146](../backend/pkg/function-test/services.go)) | 本次绕过:直接 `vite build`(不跑 vue-tsc)。**需单独修 Func.ts** |
| 7 | Gradle 提示装 build-tools 34 / arm64-v8a 无 .so | AGP 自动补依赖;只编了 x86_64 | 自动安装即可;arm64 需 `compile:go:shared ARCH=arm64` |
| 8 | `task: compile:go:shared` 报 `"uname": executable file not found` / `Unsupported host OS:` | PowerShell session 的 PATH 缺 Git `usr/bin`(uname/ls/sort/tail 所在),go-task 的 sh 脚本无法探测 host | 跑前 `$env:PATH = "C:\Users\<u>\AppData\Local\Programs\Git\usr\bin;" + $env:PATH`;PATH 修复后 `.HOST_ARCH` 正确解析为 amd64(=模拟器 ABI),默认即编 x86_64 |
| 9 | 桌面 `go build ./...` 报 `missing go.sum entry for ... rain-robot` | worktree 的 go.sum 缺 rain-robot(远程 require)条目,与桌面 cgo 无关 | `go mod download git.devcloud.ztgame.com/v-tangfangda/rain-robot` 补 go.sum(不改 go.mod) |
| 10 | `compile:go:shared` 报 `Error: Android NDK not found`(PowerShell 中,即便 PATH/NDK 目录都对) | PowerShell `$env:ANDROID_NDK_HOME` 经 wails3→go-task→bash 链路**未继承**到 bash 子进程,Taskfile `NDK_ROOT="$ANDROID_NDK_HOME"` 读到空 → 落到 `ls $SDK_ROOT/ndk/*` 兜底,SDK_ROOT 模板变量也空 → 报 NDK not found(与踩坑 8 不同:8 是 PATH 缺 uname,本项是 env 继承) | 改用 **git bash** 直接跑(同 shell 继承 env):<br>`export PATH="/c/Users/<u>/go/bin:/c/Users/<u>/AppData/Local/Programs/Git/usr/bin:$PATH"`<br>`export ANDROID_HOME="D:/Android/Sdk"`<br>`export ANDROID_NDK_HOME="D:/Android/Sdk/ndk/26.3.11579264"`<br>`wails3 task android:compile:go:shared` |

## 已知问题(待办)

1. ~~**`/wails/runtime` 422**~~ ✅ **已修复** — 根因是前端 `@wailsio/runtime` 钉在 `alpha.78`,而 alpha2.x 模板用 npm `latest`(=`alpha.97`),协议不匹配。修复:升级 `@wailsio/runtime` → `alpha.97` + 重生成 bindings + 重建 dist/.so/APK。已验证 logcat 无 422、服务调用成功返回。
2. **前端 `Func.ts` 签名过时** — `RunRobotTest` 调用少第 16 参 `maxTimeoutPerCase`(见踩坑 6),`vue-tsc` 报错。本次构建绕过(直接 `vite build`),需单独修。
3. ~~**`linux_cgo.c/.h` 补丁非复现**~~ ✅ **已解决** — 升级到 `alpha2.117` 后上游已带 `!android` 守卫,无需补丁。

## 仅构建 APK 文件(不启模拟器)

`wails3 task android:package` 走 release 路径;debug APK 用上面的 `assembleDebug`。fat APK(arm64+x86_64)需先 `wails3 task android:compile:go:shared ARCH=arm64` 再 `compile:go:shared ARCH=amd64`。
