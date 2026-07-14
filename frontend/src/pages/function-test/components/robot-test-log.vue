<script setup lang="ts">
import {logCache, LogLevel, LogMsg, LogType} from "../composables/RobotTestLog";
import {h, nextTick, onMounted, onUnmounted, ref, watch} from "vue";
import {DebugLog} from "../composables/Option";

const isAssetWrong = (msg: LogMsg) => msg.Level == LogLevel.ERROR && msg.Type == LogType.ASSET
const isAssetMaybeWrong = (msg: LogMsg) => msg.Level == LogLevel.WARN && msg.Type == LogType.ASSET
const isAssetOk = (msg: LogMsg) => msg.Level == LogLevel.INFO && msg.Type == LogType.ASSET
const isError = (msg: LogMsg) => msg.Level == LogLevel.ERROR && msg.Type == LogType.None

// 使用 Map 来管理引用
const scrollbarRefs = ref(new Map<string, any>())
const contentRefs = ref(new Map<string, HTMLElement>())

let autoScroll = true

const wheel = (e: WheelEvent) => {
  if (e.deltaY < 0 && autoScroll) {
    autoScroll = false
  }
}

onMounted(() => {
  document.addEventListener('wheel', wheel)
})
onUnmounted(() => {
  document.removeEventListener('wheel', wheel)
})

/**
 * 监听日志缓存变化，自动切换标签页并滚动到底部
 *
 * 注意：defaultVal 更新逻辑必须位于 autoScroll 检查之前。
 * 原因：用户向上滚动日志后 autoScroll 变为 false，若此时执行新用例，
 * 需要确保 defaultVal 先更新到新用例，否则 n-tabs 的 value 指向旧用例名
 *（可能已不在 logCache 中），导致面板渲染为空。
 */
watch(logCache, () => {
  const keys = Object.keys(logCache)
  // 当当前标签页失效（为空或对应key已不存在）时，自动切换到第一个可用标签页
  // 注意：此逻辑不受 autoScroll 影响，确保切换用例后标签页始终正确显示
  if (!defaultVal.value || !keys.includes(defaultVal.value)) {
    defaultVal.value = keys.length > 0 ? keys[0] : undefined
  }
  if (!autoScroll) return
  nextTick(() => {
    scrollToBottom()
  })
}, {deep: true})

const scrollToBottom = () => {
  const activeTabName = defaultVal.value || Object.keys(logCache)[0]

  const scrollbar = scrollbarRefs.value.get(activeTabName)
  const container = contentRefs.value.get(activeTabName)

  if (scrollbar && container) {
    scrollbar.scrollTo({
      top: container.scrollHeight,
      behavior: 'smooth'
    })
  }
}

// 设置 ref 的函数
const setScrollbarRef = (el: any, key: string) => {
  if (el) {
    scrollbarRefs.value.set(key, el)
  }
}

const setContentRef = (el: HTMLElement, key: string) => {
  if (el) {
    contentRefs.value.set(key, el)
  }
}

const defaultVal = ref<string>()

const switchTab = (k: string) => {
  defaultVal.value = k
}
</script>

<template>
  <n-tabs style="height: 100%" type="bar" size="small"
          :value="defaultVal"
          @update:value="switchTab"
  >
    <!--flex: 1占用剩余空间 min-height:0 防止溢出-->
    <n-tab-pane
        type="bar"
        style="flex: 1; min-height: 0;"
        v-for="k in Object.keys(logCache).sort()"
        :scrollable="{
          prev: true,
          next: true
        }"
        :key="k"
        :name="k"
        :tab="h('div', {style: `color: ${logCache[k].find(l=>l.msg.Level >= LogLevel.ERROR) ? '#ff3f3f' : 'white'}`}, k)"
    >
      <n-scrollbar
          :ref="(el: any) => setScrollbarRef(el, k)"
          style="max-height: 100%"
          :x-scrollable="DebugLog"
          :y-placement="DebugLog ? 'left' : 'right'"
      >
        <!-- 普通模式：显示所有日志，按 ID 格式化不同显示 -->
        <div v-if="!(DebugLog)" :ref="(el: any) => setContentRef(el, k)">
          <div v-for="log in logCache[k]"
               :style="{color: isAssetWrong(log.msg) ? '#ff4141' : isAssetMaybeWrong(log.msg) ? '#ffc74d' : isAssetOk(log.msg) ? '#66ff5c' : isError(log.msg) ? '#be5cff' : ''}">
            [{{ log.msg.Time.substring(0, 26) }}],
            ID[{{ log.msg.ID }}],
            Case[{{ log.msg.Case }}],
            name[{{ log.msg.RobotName }}],
            [{{ LogLevel[log.msg.Level] }}],
            {{ log.msg.Msg }}
          </div>
        </div>
        <!-- DebugLog 模式：单行不换行 -->
        <div v-else :ref="(el: any) => setContentRef(el, k)">
          <div v-for="log in logCache[k]"
               :style="{color: isAssetWrong(log.msg) ? '#ff4141' : isAssetMaybeWrong(log.msg) ? '#ffc74d' : isAssetOk(log.msg) ? '#66ff5c' : isError(log.msg) ? '#be5cff' : ''}"
               style="text-wrap: nowrap"
          >
            [{{ log.msg.Time.substring(0, 26) }}],
            ID[{{ log.msg.ID }}],
            Case[{{ log.msg.Case }}],
            name[{{ log.msg.RobotName }}],
            [{{ LogLevel[log.msg.Level] }}],
            {{ log.msg.Msg }}
          </div>
        </div>
      </n-scrollbar>
    </n-tab-pane>
  </n-tabs>
</template>

<style scoped>

</style>
