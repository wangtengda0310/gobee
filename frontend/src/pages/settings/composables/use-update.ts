// use-update - Android APK 自动更新前端逻辑
//
// 流程:检查更新(Java 桥读 versionCode → Go CheckUpdate 对比服务端) →
// 下载(Go DownloadApk + updateProgress 事件) → 安装(Java 桥 installApk 调系统安装器)。
// 仅 Android 端可用(WailsJSBridge 注入 window.wails);桌面端用 wails3 自带 updater。
// 详见 docs/Android-自动更新.md。
import { ref } from 'vue'
import { Events } from '@wailsio/runtime'
import { CheckUpdate, DownloadApk } from '@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/settings/update/updateservice'
import type { UpdateInfo } from '@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/settings/update/models'

// Android Java 桥(WailsJSBridge 注入 window.wails,@JavascriptInterface 暴露)。
// 桌面端无此对象(isAndroid=false,UI 隐藏更新功能)。
declare global {
  interface Window {
    wails?: {
      getAppVersionCode: () => number
      installApk: (path: string) => void
    }
  }
}

const isAndroid = typeof window !== 'undefined' && !!window.wails?.installApk
const checking = ref(false)
const downloading = ref(false)
const progress = ref(0)
const updateInfo = ref<UpdateInfo | null>(null)
const errorMsg = ref('')

export function useUpdate() {
  // checkUpdate:读当前 versionCode(Java 桥,build.gradle 单一源) → Go CheckUpdate 对比服务端 latest.json
  async function checkUpdate() {
    errorMsg.value = ''
    if (!isAndroid) {
      errorMsg.value = '自动更新仅在 Android 端可用(桌面用 wails3 自带 updater)'
      return
    }
    checking.value = true
    try {
      const current = window.wails!.getAppVersionCode()
      const info = await CheckUpdate(current)
      updateInfo.value = info
      if (!info) errorMsg.value = '已是最新版本'
    } catch (e: any) {
      // Wails bindings reject 的 e 可能是 Error/string/对象,统一提取可读 message
      errorMsg.value = e instanceof Error ? e.message
        : (typeof e === 'string' ? e : (e?.message || JSON.stringify(e)))
    } finally {
      checking.value = false
    }
  }

  // downloadAndInstall:Go DownloadApk 下载(监听 updateProgress 事件) → Java installApk 触发系统安装器
  async function downloadAndInstall() {
    if (!updateInfo.value || !isAndroid) return
    errorMsg.value = ''
    downloading.value = true
    progress.value = 0
    const off = Events.On('updateProgress', (e: any) => {
      progress.value = e.data?.percent ?? 0
    })
    try {
      const path = await DownloadApk(updateInfo.value)
      off()
      // 触发系统安装器(用户需点"安装"确认,非 rooted 不可静默)
      window.wails!.installApk(path)
    } catch (e) {
      errorMsg.value = String(e)
    } finally {
      downloading.value = false
    }
  }

  return { isAndroid, checking, downloading, progress, updateInfo, errorMsg, checkUpdate, downloadAndInstall }
}
