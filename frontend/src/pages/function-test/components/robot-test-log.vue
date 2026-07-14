<script setup lang="ts">
import {
  getCaseLogOmittedCount,
  getCaseLogsForDisplay,
  logCacheRevision,
  LogLevel,
  LogMsg,
  LogType,
  isAssertionFailureLog,
  nowRunningCase,
  nowRunningCaseStats
} from "../composables/RobotTestLog";
import {computed, h, nextTick, onMounted, onUnmounted, watch} from "vue";
import {DebugLog} from "../composables/Option";
import {ExtraCaseTreeOption} from "../composables/use-case-data";
import {activeLogCaseIndexRef, selectCase} from "../composables/use-case-selection";
import {getSeatColorHex} from "../config/Identity";
import {TreeOption} from "naive-ui";

const isAssetWrong = (msg: LogMsg) => isAssertionFailureLog(msg)
const isAssetMaybeWrong = (msg: LogMsg) => msg.Level == LogLevel.WARN && msg.Type == LogType.ASSET
const isAssetOk = (msg: LogMsg) => msg.Level == LogLevel.INFO && msg.Type == LogType.ASSET
const isError = (msg: LogMsg) => msg.Level == LogLevel.ERROR && msg.Type == LogType.None

const stepInfoMaps = computed(() =>
    nowRunningCase.value.map(c => {
      const map = new Map<number, { desc: string, robotIdx: number }>()
      c.caseSteps?.forEach(s => {
        if (s.id === undefined) return
        map.set(s.id, {desc: s.desc || '', robotIdx: s.robotIdx})
      })
      return map
    })
)

const getStepInfo = (
    caseIndex: number,
    logId: number
): { desc: string, robotIdx: number } | null => {
  return stepInfoMaps.value[caseIndex]?.get(logId) ?? null
}

const tabRefKey = (index: number) => String(index)

const scrollbarRefs = new Map<string, any>()
const contentRefs = new Map<string, HTMLElement>()

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

const logTabValue = computed(() => {
  const len = nowRunningCase.value.length
  if (len === 0) return null
  const idx = Number(activeLogCaseIndexRef.value ?? 0)
  return idx >= 0 && idx < len ? idx : 0
})

const onLogTabUpdate = (value: string | number) => {
  selectCase(value)
}

watch(logCacheRevision, () => {
  const batchLen = nowRunningCase.value.length
  if (batchLen > 0) {
    const idx = activeLogCaseIndexRef.value
    if (idx === undefined || idx < 0 || idx >= batchLen) {
      selectCase(0)
    }
  }
  if (!autoScroll) return
  nextTick(() => {
    scrollToBottom()
  })
})

const scrollToBottom = () => {
  const activeIndex = logTabValue.value ?? 0
  const refKey = tabRefKey(activeIndex)

  const scrollbar = scrollbarRefs.get(refKey)
  const container = contentRefs.get(refKey)

  if (scrollbar && container) {
    scrollbar.scrollTo({
      top: container.scrollHeight,
      behavior: 'auto'
    })
  }
}

const setScrollbarRef = (el: any, key: string) => {
  if (el) {
    scrollbarRefs.set(key, el)
  }
}

const setContentRef = (el: HTMLElement, key: string) => {
  if (el) {
    contentRefs.set(key, el)
  }
}

const tabLabel = (caseOption: TreeOption & ExtraCaseTreeOption, index: number) => {
  const label = caseOption.label ?? `用例${index + 1}`
  const hasError = nowRunningCaseStats.value[index]?.hasAnyAssertionFailure === true
  return h('div', {style: `color: ${hasError ? '#ff3f3f' : 'white'}`}, label)
}

const logsForCase = (caseOption: TreeOption & ExtraCaseTreeOption) => {
  const label = caseOption.label
  if (!label) return []
  return getCaseLogsForDisplay(label)
}

const omittedLogCount = (caseOption: TreeOption & ExtraCaseTreeOption) => {
  const label = caseOption.label
  if (!label) return 0
  return getCaseLogOmittedCount(label)
}

const logLineStyle = (msg: LogMsg) => ({
  color: isAssetWrong(msg) ? '#ff4141'
      : isAssetMaybeWrong(msg) ? '#ffc74d'
          : isAssetOk(msg) ? '#66ff5c'
              : isError(msg) ? '#be5cff'
                  : ''
})
</script>

<template>
  <div class="robot-test-log-root">
    <n-empty v-if="nowRunningCase.length === 0" description="暂无执行中的用例，请先运行用例"/>
    <n-tabs
        v-else
        class="robot-test-log-tabs"
        style="height: 100%"
        type="line"
        size="small"
        :scrollable="nowRunningCase.length > 4"
        :value="logTabValue"
        @update:value="onLogTabUpdate"
    >
      <n-tab-pane
          v-for="(caseOption, index) in nowRunningCase"
          style="height: 100%; min-height: 0; overflow: hidden"
          :key="caseOption.key ?? index"
          :name="index"
          :tab="tabLabel(caseOption, index)"
      >
        <template v-if="logTabValue === index">
          <n-scrollbar
              :ref="(el: any) => setScrollbarRef(el, tabRefKey(index))"
              style="max-height: 100%"
              :x-scrollable="DebugLog"
              :y-placement="DebugLog ? 'left' : 'right'"
          >
            <div v-if="omittedLogCount(caseOption) > 0" class="log-omitted-hint">
              已省略较早的 {{ omittedLogCount(caseOption) }} 条日志（仅展示最近 {{ logsForCase(caseOption).length }} 条）
            </div>
            <div v-if="!(DebugLog)" :ref="(el: any) => setContentRef(el, tabRefKey(index))" class="log-lines">
              <div
                  v-for="log in logsForCase(caseOption)"
                  :key="log.seq"
                  class="log-line"
                  :style="logLineStyle(log.msg)"
              >
                [{{ log.msg.Time.substring(0, 26) }}],
                <span :style="{color: getSeatColorHex(caseOption.initYanWu?.customHeroes, getStepInfo(index, log.msg.ID)?.robotIdx || 0)}">动作[{{ log.msg.ID }}], Step[{{ getStepInfo(index, log.msg.ID)?.desc || log.msg.Case }}]</span>,
                name[{{ log.msg.RobotName }}],
                [{{ LogLevel[log.msg.Level] }}],
                {{ log.msg.Msg }}
              </div>
            </div>
            <div v-else :ref="(el: any) => setContentRef(el, tabRefKey(index))" class="log-lines">
              <div
                  v-for="log in logsForCase(caseOption)"
                  :key="log.seq"
                  class="log-line log-line--nowrap"
                  :style="logLineStyle(log.msg)"
              >
                [{{ log.msg.Time.substring(0, 26) }}],
                <span :style="{color: getSeatColorHex(caseOption.initYanWu?.customHeroes, getStepInfo(index, log.msg.ID)?.robotIdx || 0)}">动作[{{ log.msg.ID }}], Step[{{ getStepInfo(index, log.msg.ID)?.desc || log.msg.Case }}]</span>,
                name[{{ log.msg.RobotName }}],
                [{{ LogLevel[log.msg.Level] }}],
                {{ log.msg.Msg }}
              </div>
            </div>
          </n-scrollbar>
        </template>
        <div v-else class="log-inactive-placeholder">
          {{ (caseOption.label && getCaseLogsForDisplay(caseOption.label).length) || 0 }} 条日志（切换到此 Tab 查看）
        </div>
      </n-tab-pane>
    </n-tabs>
  </div>
</template>

<style scoped>
.robot-test-log-root {
  height: 100%;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.robot-test-log-tabs {
  flex: 1;
  min-height: 0;
  overflow: hidden;
}

.robot-test-log-tabs :deep(.n-tabs-nav) {
  padding-left: 0;
}

.robot-test-log-tabs :deep(.n-tabs-pane-wrapper) {
  flex: 1;
  min-height: 0;
  overflow: hidden;
}

.log-lines {
  font-size: 12px;
  line-height: 1.5;
}

.log-line {
  margin-bottom: 2px;
}

.log-line--nowrap {
  text-wrap: nowrap;
}

.log-omitted-hint {
  color: #999;
  font-size: 12px;
  margin-bottom: 6px;
}

.log-inactive-placeholder {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #888;
  font-size: 13px;
}
</style>
