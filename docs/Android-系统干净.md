# Android 系统干净（运行时 + 卸载）

> rain-qa-func Android 端"不留垃圾历史数据"的设计与实施。CLAUDE.md "系统干净约束"的具体落地。
> 相关:[Android-自动更新.md](Android-自动更新.md)(下载优化) | [Android-数据加载.md](Android-数据加载.md)(C/A/B/D 方案)

## ⚠️ Rewind 恢复线索（会话回退后先读本节，避免重复工作）

**最近状态(2026-07-13)**:
- versionCode **8** 已上传 itsnot.fun(含 C 方案 + 下载优化断点续传 + headApk 超时 + emitProgress + reload 修复)
- 用户正在**手机浏览器下载 versionCode 8** 测试(因 6→7 应用内下载卡 0%,改浏览器诊断网络 vs 应用)
- **待用户反馈**:① 浏览器下载 8 顺不顺(区分网络/服务器 vs 应用逻辑) ② 装 8 后加载示例能否见 10 用例
- 本文档同批提交了 **启动清 APK**(versionCode 9 待构建上传) + CLAUDE.md 加"禁止私有目录外文件操作"

**已完成(勿重复)**:
- C 方案内置示例数据(`backend/pkg/exampledata/`,独立包)— versionCode 5
- 下载优化(分块 2MB + 30s 超时 + 断点续传 .part + 重试 5 次 + cleanOldUpdates)— versionCode 6
- 加载示例后 reload 前端(Option.ts GetConfig 模块级一次性,改配置需 reload 刷新用例树)— versionCode 7
- headApk 加 15s 超时 + emitProgress(0) + 诊断日志(修 0% 不动)— versionCode 8
- 启动清 APK(CleanInstalledApks,wails.go 启动调)— versionCode 9(本次)
- CLAUDE.md 加"禁止私有目录外文件操作"约束

**未完成(rewind 后接着做)**:
- [ ] versionCode 9 构建 + 上传(启动清 APK 已写码,未构建)
- [ ] 用户反馈 versionCode 8 测试结果(浏览器下载 + 加载示例 10 用例)
- [ ] 若下载仍卡 → 诊断 serverLog `[update]` 日志(headApk/块下载重试)定位
- [ ] picker cache 清理机制(OpenFileDialog 用后清,A/B 数据导入方案时补)
- [ ] 系统干净检验脚本(adb shell du 重复操作前后对比)

## 一、两个维度

| 维度 | 谁负责 | 机制 |
|---|---|---|
| **卸载干净** | Android 系统 | 卸载自动删私有目录 `/data/data/<pkg>/` 全部(files/cache/databases)。应用无卸载回调,不需写码。**前提:数据全在私有目录** |
| **运行时干净** | 我们代码 | 运行中/长期使用,文件不无限累积(有清理机制) |

## 二、涉及文件(私有目录内,卸载全清)

| 文件 | 位置 | 大小 | 清理机制 | 状态 |
|---|---|---|---|---|
| APK 下载 | `updates/rain-qa-func-<vc>.apk` | 63MB/个 | `cleanOldUpdates`(每次下载清旧版本) + `CleanInstalledApks`(启动清已装的) | ✅ |
| .part 续传 | `updates/*.part` | ≤63MB | `cleanOldUpdates`(新版本清旧)+ 完成后 rename .apk | ✅ |
| 示例数据 | `example-data/` | 1.4M | `LoadExampleData` 先 RemoveAll 再释放(多次加载不累积) | ✅ |
| picker cache | `cache/wails-picker/` | 视文件 | 待清理(A/B 数据导入用 OpenFileDialog 时补) | ⚠️ 待办 |
| 配置 | `.rain-qa-func.json` | 小 | section 覆盖写(路径切换不残留) | ✅ |
| serverLog | 内存 | - | maxHistory=500(既有) | ✅ |

## 三、已实施

1. **`cleanOldUpdates`**([update/service.go](../backend/pkg/settings/update/service.go)):每次 DownloadApk 清旧 versionCode 的 APK + .part,仅留当前版本
2. **`CleanInstalledApks`**(update/service.go):应用启动调(wails.go),清 `updates/*.apk`(已装的),保留 `.part`(续传未完成)。方案 🅱️:安装失败可重试(APK 未删),装成功下次启动清
3. **`LoadExampleData` 先 RemoveAll**([exampledata/service.go](../backend/pkg/exampledata/service.go)):多次加载示例不累积
4. **serverLog maxHistory=500**(既有):日志缓存上限

APK 清理方案选 🅱️(安装失败可重试) + 启动清(装成功下次启动清),非 🅰️(安装后立即清,失败要重下 63MB)。

## 四、禁止私有目录外文件操作(CLAUDE.md 约束)

所有文件读写必须在私有目录(`/data/data/<pkg>/files|cache`)。禁止写 `/sdcard`(卸载残留)。
- 用户数据导入 → 释放到私有目录(C/A/B 方案)
- 不用 D(MANAGE_EXTERNAL_STORAGE 写 /sdcard)——破坏卸载干净

## 五、检验方法(判定符合预期)

```bash
# 1. 运行时:重复操作前后对比私有目录大小(不应单调增长)
adb shell du -sh /data/data/com.wails.app/files/
#   重复"加载示例"5 次 → example-data/ 大小恒定(非 ×5)
#   重复"更新"到不同版本 → updates/ 始终 ≤1 APK(当前版本,启动清后为 0)

# 2. 卸载:卸载后私有目录应不存在
adb shell pm uninstall com.wails.app
adb shell ls /data/data/com.wails.app      # 应: No such file or directory

# 3. 启动清验证:更新装成功后重启应用 → updates/*.apk 应被清
adb shell ls /data/data/com.wails.app/files/updates/
```

判定标准(CLAUDE.md):同一操作重复 N 次,存储不单调增长;卸载无残留。

## 六、待办

- [ ] versionCode 9 构建 + 上传(启动清 APK)
- [ ] picker cache 清理(OpenFileDialog 用后清,A/B 方案时补)
- [ ] 检验脚本 + 实测记录基线
