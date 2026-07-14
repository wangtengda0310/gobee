<script setup lang="ts">
/**
 * StatusBar - 应用状态栏组件
 *
 * @module shared/components/status-bar
 * @description
 * 显示应用版本信息和最近提交记录
 * - 左侧：应用名称
 * - 中间：自定义插槽（可扩展）
 * - 右侧：版本发布提示（hover 显示详细提交记录）
 */
import { ref, computed, onMounted } from 'vue'
import { NTooltip, NTag } from 'naive-ui'
import { VersionService } from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/settings"

interface CommitInfo {
  hash: string
  message: string
  author: string
  date: string
}

interface BuildInfo {
  commitHash: string
  commitMsg: string
  buildTime: string
}

const commits = ref<CommitInfo[]>([])
const buildInfo = ref<BuildInfo | null>(null)
const currentDir = ref<string>('')
const loading = ref(false)

// 加载 commit 信息和构建信息
const loadCommits = async () => {
  loading.value = true
  try {
    // 并行获取提交信息、构建信息和当前目录
    const [commitResult, buildResult, dirResult] = await Promise.all([
      VersionService.GetRecentCommits(5),
      VersionService.GetBuildInfo(),
      VersionService.GetCurrentDirectory()
    ])
    commits.value = commitResult
    buildInfo.value = buildResult
    currentDir.value = dirResult
  } catch (err) {
    console.error('获取 commit 信息失败:', err)
    commits.value = [{ hash: 'N/A', message: '无法获取版本信息', author: '', date: '' }]
  } finally {
    loading.value = false
  }
}

// 生成 tooltip 内容
const tooltipContent = computed(() => {
  if (commits.value.length === 0) return '加载中...'
  return commits.value
})

// 最新提交的短哈希
const latestHash = computed(() => {
  if (commits.value.length > 0 && commits.value[0].hash !== 'N/A') {
    return commits.value[0].hash
  }
  return '...'
})

// 是否显示构建时间（生产构建时才有）
const hasBuildTime = computed(() => {
  return buildInfo.value?.buildTime && buildInfo.value.buildTime !== ''
})

// 从右侧截取路径，显示路径尾部内容
const MAX_PATH_DISPLAY = 30
const displayDir = computed(() => {
  const dir = currentDir.value || 'rain-qa-func'
  if (dir.length <= MAX_PATH_DISPLAY) return dir
  return '...' + dir.slice(dir.length - MAX_PATH_DISPLAY)
})

onMounted(() => {
  loadCommits()
})
</script>

<template>
  <div class="status-bar">
    <!-- 左侧：应用信息 -->
    <div class="status-left">
      <n-tooltip trigger="hover" placement="top-start">
        <template #trigger>
          <n-tag size="small" :bordered="false" type="info" class="dir-tag">
            {{ displayDir }}
          </n-tag>
        </template>
        <span>{{ currentDir || 'rain-qa-func' }}</span>
      </n-tooltip>
    </div>

    <!-- 中间：自定义信息区（可扩展） -->
    <div class="status-center">
      <slot name="custom-info"></slot>
    </div>

    <!-- 右侧：版本发布提示 -->
    <div class="status-right">
      <n-tooltip trigger="hover" placement="top-end" :style="{ maxWidth: '600px' }">
        <template #trigger>
          <span class="version-tip">
            最近更新: {{ latestHash }}
            <span v-if="hasBuildTime" class="build-time-badge">已构建</span>
          </span>
        </template>
        <div class="commit-list">
          <!-- 显示构建信息（生产构建时才有） -->
          <div v-if="hasBuildTime" class="build-info">
            <span class="build-label">构建时间:</span>
            <span class="build-value">{{ buildInfo?.buildTime }}</span>
          </div>
          <div v-for="commit in commits" :key="commit.hash" class="commit-item">
            <span class="commit-hash">{{ commit.hash }}</span>
            <span class="commit-msg">{{ commit.message }}</span>
            <span class="commit-meta">{{ commit.author }} · {{ commit.date }}</span>
          </div>
        </div>
      </n-tooltip>
    </div>
  </div>
</template>

<style scoped>
.status-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  height: 100%;
  padding: 0 12px;
  user-select: none;
}

.status-left, .status-right {
  display: flex;
  align-items: center;
}

.status-center {
  display: flex;
  align-items: center;
  flex: 1;
  min-width: 0;
  overflow: hidden;
}

/* 这里需要留出足够边界，不然内容会显示不全 */
.status-right {
  padding-right: 2em;
  flex-shrink: 0;
}

.version-tip {
  cursor: pointer;
  opacity: 0.7;
  font-size: 12px;
  color: #999;
  transition: all 0.2s ease;
}

.version-tip:hover {
  opacity: 1;
  color: #fff;
}

.commit-list {
  font-size: 12px;
  max-height: 300px;
  overflow-y: auto;
}

.commit-item {
  padding: 6px 0;
  border-bottom: 1px solid #333;
  display: flex;
  gap: 8px;
  align-items: center;
}

.commit-item:last-child {
  border-bottom: none;
}

.commit-hash {
  color: #58a6ff;
  font-family: monospace;
  flex-shrink: 0;
  min-width: 50px;
}

.commit-msg {
  color: #ccc;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 300px;
}

.commit-meta {
  color: #666;
  flex-shrink: 0;
  font-size: 11px;
  min-width: 100px;
  text-align: right;
}

/* 构建时间徽章 */
.build-time-badge {
  margin-left: 6px;
  padding: 1px 4px;
  background: #18a058;
  border-radius: 3px;
  font-size: 10px;
  color: #fff;
}

/* 构建信息区域 */
.build-info {
  padding: 6px 0 10px 0;
  margin-bottom: 6px;
  border-bottom: 1px solid #444;
  display: flex;
  gap: 8px;
  align-items: center;
}

.build-label {
  color: #888;
  font-size: 11px;
}

.build-value {
  color: #18a058;
  font-family: monospace;
  font-size: 12px;
}

/* 目录标签：截取逻辑在 JS 层完成，此处仅保证不换行 */
.dir-tag {
  white-space: nowrap;
}
</style>
