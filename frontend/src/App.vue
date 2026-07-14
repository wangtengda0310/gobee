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

// 触控设备(pointer: coarse,如 Android/手机触屏)增大 Naive UI 按钮高度。
// 桌面(pointer: fine)不匹配 matchMedia,themeOverrides=undefined,主题完全不变。
// 用 Naive UI 官方 themeOverrides 而非 CSS 全局覆盖,保证按钮内部布局自适应。
const isTouchDevice = typeof window !== 'undefined' && window.matchMedia('(pointer: coarse)').matches
const themeOverrides = isTouchDevice ? {
  // medium 34→40, small 28→38, tiny 22→34;主导航 idea-icon-button 已单独 44px
  Button: { heightMedium: '40px', heightSmall: '38px', heightTiny: '34px' },
} : undefined

// 布局层判定:UA 含 Android → 移动端布局(仅真移动设备)。
// 挂到 <html>.is-mobile,供全局 CSS / 组件 v-if 锁定"布局重排"(锚点栏收起等)。
// 作为 pointer:coarse 的精确补充——pointer 会误判触屏笔记本(CSS-Tricks 指出),
// 交互增强(按钮大)对触屏笔记本无害,但布局重排会破坏其桌面多栏,故布局层必须用 .is-mobile。
// PC(UA 非 Android)不触发,零影响。
const isMobile = typeof navigator !== 'undefined' && /android/i.test(navigator.userAgent)
if (isMobile && typeof document !== 'undefined') {
  document.documentElement.classList.add('is-mobile')
}

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
      <n-config-provider id="layout-theme-config" :theme="darkTheme" :theme-overrides="themeOverrides">
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

/* 触控设备（pointer: coarse，如 Android/手机触屏）适配。
   桌面鼠标（pointer: fine）不匹配此媒体查询,布局完全不变。 */
@media (pointer: coarse) {
  /* 主导航按钮:移动端均分换行(8 按钮 × 文字 > 360 装不下,
     换行 2 行 4 按钮让全显,而非横滚出屏外按钮不可见) */
  .idea-icon-button {
    min-height: 44px;
    min-width: 44px;
    flex: 1 1 auto;
    white-space: nowrap;
    width: auto;
  }
  /* 主导航容器:换行 + height auto(原 overflow-x:auto 横滚导致按钮屏外,
     用户反馈"超出屏幕";换行让所有按钮在屏内可见) */
  #layout-header {
    flex-wrap: wrap;
    height: auto !important;
    gap: 4px;
    padding: 4px 6px;
  }
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

<!-- 全局样式(非 scoped):n-drawer 抽屉 teleport 到 body,scoped 的 data-v 不生效,必须用全局 <style> -->
<style>
@media (pointer: coarse) {
  /* 抽屉移动端满屏:桌面宽度(如 500/700px)在 360 视口溢出,导致右上角关闭按钮 X 在屏外无法点击。
     改为 100vw 后 X 可达;桌面(pointer:fine)不匹配,抽屉宽度不变。
     ⚠️ n-drawer 本身(width prop 设此元素)也须覆盖,否则 wrapper 虽 100vw 但 drawer 仍 700px,
     placement=right 时 x=360-700=-340 左半屏外(关闭按钮在左上角不可见)。 */
  .n-drawer,
  .n-drawer-content-wrapper {
    max-width: 100vw !important;
    width: 100vw !important;
  }
  /* 抽屉内部内容按桌面宽(400/500/700)设计,wrapper 虽 100vw 但 content/body 仍按原宽溢出。
     约束 max-width:100% + overflow-x:auto 兜底,内部宽内容(表格/表单)横滚而非出屏。 */
  .n-drawer-content {
    max-width: 100% !important;
  }
  .n-drawer-body-content {
    max-width: 100% !important;
    overflow-x: auto !important;
  }
  /* 战斗/配表 footer(n-layout-footer.footer):多 statistic + FooterCaseLogStatistic
     (n-progress min300 + 用例 min250 + 错误 min200 = 750px)远超 360 视口。
     移动端 footer 内 overflow-x:auto 横滚(footer 本身 360 不超出屏,内容在 footer 内滑),
     不增高不破坏 content 布局;桌面(pointer:fine)不匹配,footer 原 flex 不变。 */
  .footer {
    overflow-x: auto !important;
    flex-wrap: nowrap !important;
    -webkit-overflow-scrolling: touch;
  }
  .footer > * {
    flex-shrink: 0;
  }
  /* n-tab-pane 溢出 2px(padding/border content-box):box-sizing 修 */
  .n-tab-pane {
    box-sizing: border-box;
  }
  /* 兜底:各页已适配无溢出(sider 折叠/锚点收起/PathConfigInput 自适应),改 hidden 彻底
     消除横向滚动——否则 #layout 因某子元素(如 status-bar padding content-box)溢出而可横滚,
     用户向左拖会出现右侧白边(只左拖出白边、右拖恢复是典型 scrollLeft 现象)。
     桌面(pointer:fine)不匹配,overflow 不变。 */
  html, body, #layout {
    overflow-x: hidden !important;
  }
  /* #layout 改 flex 列布局:header 移动端换行(2 行)增高后,
     content 用 flex:1 自适应剩余高度(原 height:calc 固定 68px 头尾不够)。
     桌面(pointer:fine)不匹配,保持原 position:relative + height:calc 布局。 */
  #layout {
    display: flex;
    flex-direction: column;
  }
  #layout-header {
    flex: 0 0 auto;
  }
  #layout-content {
    flex: 1 1 0;
    height: auto !important;
    min-height: 0;
  }
  #layout-footer {
    flex: 0 0 auto;
  }
}
</style>
