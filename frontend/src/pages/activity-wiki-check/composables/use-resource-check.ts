/**
 * 资源路径检查 composable
 *
 * 封装后端 CheckResourcePaths API，
 * 管理资源存在状态和图片预览数据，支持批量检查和缓存。
 */
import {ref, type Ref} from "vue"
import {CheckResourcePaths} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/activity-wiki-check/activitywikicheckservice.js"

/** 资源检查结果 */
export interface ResourceStatus {
  /** 原始路径 */
  path: string
  /** 文件是否存在 */
  exists: boolean
  /** base64 data URL（仅图片文件存在时） */
  previewUrl: string
}

/**
 * 资源路径检查 composable
 *
 * 提供批量检查资源路径是否存在、获取预览的能力。
 * 检查结果按路径缓存，避免重复请求。
 */
export function useResourceCheck(clientPath: Ref<string>) {
  /** 资源路径 → 检查结果的缓存 */
  const cache = ref<Map<string, ResourceStatus>>(new Map())

  /** 是否正在检查 */
  const checking = ref(false)

  /**
   * 批量检查资源路径
   *
   * 仅检查缓存中不存在的路径，已有缓存直接使用。
   * 检查完成后更新缓存。
   */
  const checkPaths = async (paths: string[]) => {
    // 过滤空值和已缓存的路径
    const uncached = paths.filter(p => p && !cache.value.has(p))
    if (uncached.length === 0 || !clientPath.value) return

    checking.value = true
    try {
      const results = await CheckResourcePaths(uncached, clientPath.value)
      if (results) {
        for (const r of results) {
          if (r) {
            cache.value.set(r.path, {
              path: r.path,
              exists: r.exists,
              previewUrl: r.previewUrl || '',
            })
          }
        }
      }
    } catch (err) {
      console.error('资源检查失败:', err)
    } finally {
      checking.value = false
    }
  }

  /** 获取指定路径的缓存状态 */
  const getStatus = (path: string): ResourceStatus | undefined => {
    return cache.value.get(path)
  }

  /** 清空缓存 */
  const clearCache = () => {
    cache.value.clear()
  }

  return {
    cache,
    checking,
    checkPaths,
    getStatus,
    clearCache,
  }
}
