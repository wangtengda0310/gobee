<!-- 普通 <script> 块：定义需要 export 的模块级变量（Vue <script setup> 不允许 export） -->
<script lang="ts">
import {createMemoryHistory, createRouter} from "vue-router";

// 路由实例，export 供 main.ts 导入并注册到 app
export const router = createRouter({
  history: createMemoryHistory(),
  routes: [
    {path: '/', redirect: '/Test'},
    {path: '/Settings', component: () => import("@/pages/settings/index.vue")},
    {path: '/Home', component: () => import("@/pages/llm/index.vue")},
    {path: '/Test', component: () => import("@/pages/function-test/index.vue")},
    {path: '/Excel', component: () => import("@/pages/excel-test/index.vue")},
    {path: '/HeroRes', component: () => import("@/pages/hero-voice-resource-check/index.vue")},
    {path: '/HeroWikiRes', component: () => import("@/pages/hero-wiki-check/index.vue")},
    {path: '/ActivityWiki', component: () => import("@/pages/activity-wiki-check/index.vue")},
    {path: '/ProtoTest', component: () => import("@/pages/proto-test/index.vue")},
  ],
})
</script>

<!-- <script setup> 块：组件逻辑（响应式状态、菜单配置、布局计算） -->
<script setup lang="ts">
/**
 * App - 应用根组件
 *
 * 合并了原 app/index.vue 和 layouts/normal-layout/index.vue
 * 提供应用的标准布局结构：
 * - Header: 导航按钮区域
 * - Content: 主内容区域（带 Naive UI 主题配置）
 * - Footer: 状态栏区域
 */
import {PropType, ref, computed, watch, onBeforeUnmount} from "vue";
import {useRoute} from "vue-router";
import {darkTheme, useThemeVars} from 'naive-ui'
import StatusBar from '@shared/components/status-bar/index.vue'
import InterceptNotification from '@/pages/settings/components/intercept-notification.vue'
import ServerLogPanel from '@/pages/settings/components/server-log-panel.vue'
import { useServerLogs, type ServerLogEntry } from '@/pages/settings/composables/use-server-logs'
const route = useRoute()
const themeVars = useThemeVars()

const props = defineProps({
  headerTitle: {
    type: String,
    default: "标题",
  },
  sideOption: {
    type: Object as PropType<{
      title: string
    }>,
    default: {
      title: '测试'
    }
  }
})

const emits = defineEmits<{
  update: (next: void) => void
}>()

const showRoutePanel = ref(true)

const headerHeight = ref(!showRoutePanel.value ? "0px" : "40px")

// 菜单项配置
const menuItems = [
  { label: '设置', path: '/Settings' },
  { label: 'AI助手', path: '/Home' },
  { label: '战斗测试', path: '/Test' },
  { label: '配表测试', path: '/Excel' },
  { label: '武将资源检查', path: '/HeroRes' },
  { label: '武将Wiki检查', path: '/HeroWikiRes' },
  { label: '活动Wiki', path: '/ActivityWiki' },
  { label: 'Proto测试', path: '/ProtoTest' },
]
const footerHeight = ref("28px")

// 服务端日志状态栏功能
const { latestLog, statusBarLogEnabled, logPanelVisible, getServerLogColor } = useServerLogs()

// 100ms 防抖的最新日志（避免高频日志频繁更新 DOM）
const throttledLog = ref<ServerLogEntry | null>(null)
let throttleTimer: ReturnType<typeof setTimeout> | null = null

watch(latestLog, (newLog) => {
    if (!newLog) {
        throttledLog.value = null
        return
    }
    if (throttleTimer) clearTimeout(throttleTimer)
    throttleTimer = setTimeout(() => {
        throttledLog.value = newLog
    }, 100)
})

onBeforeUnmount(() => {
    if (throttleTimer) {
        clearTimeout(throttleTimer)
        throttleTimer = null
    }
})

// 内容区高度：总高度 - header - footer
const contentHeight = computed(() => {
  const header = showRoutePanel.value ? 40 : 0
  const footer = 28
  return `calc(100% - ${header + footer}px)`
})
</script>

<template>
  <div id="layout">
    <div id="layout-header" v-if="showRoutePanel">
      <button
        v-for="item in menuItems"
        :key="item.path"
        class="idea-icon-button"
        :class="{ active: route.path === item.path }"
        @click="router.push(item.path)"
      >
        {{ item.label }}
      </button>
    </div>
    <div id="layout-content">
      <n-config-provider id="layout-theme-config" :theme="darkTheme">
        <n-modal-provider>
          <n-dialog-provider>
            <n-message-provider>
              <RouterView v-slot="{ Component }">
                <keep-alive>
                  <component :is="Component" />
                </keep-alive>
              </RouterView>
            </n-message-provider>
          </n-dialog-provider>
        </n-modal-provider>
      </n-config-provider>
    </div>
    <div id="layout-footer">
      <StatusBar>
        <template #custom-info>
          <div
            v-if="statusBarLogEnabled"
            class="status-log-area"
            @click="logPanelVisible = !logPanelVisible"
          >
            <template v-if="throttledLog">
              <span
                class="log-dot"
                :style="{ backgroundColor: getServerLogColor(throttledLog.level, throttledLog.isManual) }"
              ></span>
              <span class="status-log-text">{{ throttledLog.message }}</span>
            </template>
            <span v-else class="status-log-text status-log-placeholder">等待日志...</span>
          </div>
        </template>
      </StatusBar>
    </div>
    <!-- 全局消息劫持通知条 -->
    <InterceptNotification />
    <!-- 全局服务端日志面板 -->
    <ServerLogPanel v-model:visible="logPanelVisible" />
  </div>
</template>

<style scoped>
* {
  font-family:
    /* 英文字体部分（等宽） */
      'Cascadia Code', 'Cascadia Mono',
      'JetBrains Mono', 'Fira Code',
      'SF Mono', Monaco,
      'Courier New', Courier,

        /* 中文字体部分（等宽优先） */
      'LXGW WenKai Mono', '霞鹜文楷等宽',
      'Sarasa Mono SC', '更纱黑体 Mono',
      'Source Han Mono', '思源等宽',

        /* 备选优秀中文字体（非等宽） */
      'MiSans',
      'HarmonyOS Sans',
      'PingFang SC',
      'Microsoft YaHei UI',
      'Noto Sans SC',
      'Source Han Sans SC',

        /* 最后回退 */
      monospace, sans-serif;
}

#layout {
  position: relative;
  width: 100%;
  height: 100%;
}

#layout-header {
  position: relative;
  height: v-bind(headerHeight);

  text-align: center;

  display: flex;
  justify-content: left;
  align-items: center;
  gap: 5px;

  color: white;

  background: linear-gradient(to right, #2d2f32 0%, #2c2f30 100%);
  box-sizing: border-box;
  border-bottom: 1px solid #1e1e1e;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.4);
}

#layout-content {
  position: relative;
  height: v-bind(contentHeight);
  display: flex;
  background-color: #2b2b2b;
}

#layout-footer {
  position: relative;
  height: v-bind(footerHeight);
  background: linear-gradient(to right, #1e1e1e 0%, #252525 100%);
  border-top: 1px solid #333;
  box-sizing: border-box;
}

#layout-theme-config {
  position: relative;
  width: 100%;
}

.idea-icon-button {
  position: relative;
  height: 28px;
  border-radius: 6px;
  border: none;
  background-color: #2b2b2b;
  color: #cccccc;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.1),
  0 1px 3px rgba(0, 0, 0, 0.4);
}

.idea-icon-button:hover {
  background-color: #3c3c3c;
  color: #ffffff;
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.15),
  0 2px 5px rgba(0, 0, 0, 0.5);
}

.idea-icon-button:active {
  background-color: #363636;
  transform: scale(1.05);
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.1),
  0 1px 2px rgba(0, 0, 0, 0.4);
}

/* 当前路由激活状态 */
.idea-icon-button.active {
  color: v-bind('themeVars.primaryColor');
  box-shadow: inset 0 0 0 1px v-bind('themeVars.primaryColor'),
  0 1px 3px rgba(0, 0, 0, 0.4);
}

.idea-icon-button.active:hover {
  color: v-bind('themeVars.primaryColor');
  background-color: #3c3c3c;
}

.status-log-area {
  cursor: pointer;
  padding: 0 8px;
  border-radius: 3px;
  transition: background-color 0.15s ease;
  display: flex;
  align-items: center;
  height: 100%;
  min-width: 0;
  overflow: hidden;
}

.status-log-area:hover {
  background-color: rgba(255, 255, 255, 0.08);
}

.log-dot {
  display: inline-block;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  margin-right: 6px;
  flex-shrink: 0;
}

.status-log-text {
  font-size: 12px;
  font-family: monospace;
  line-height: 28px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: #aaa;
}

.status-log-placeholder {
  color: #555;
  font-style: italic;
}
</style>
