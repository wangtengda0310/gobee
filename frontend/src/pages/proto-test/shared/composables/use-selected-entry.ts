// use-selected-entry.ts — 三个页签组件共享的"选中条目"计算逻辑
// 消除 packet-tab / testcase-tab / replay-result-tab 中 selectedPairedEntry 的重复定义

import { computed, type ComputedRef, type Ref } from 'vue'
import { buildPairedEntries, type PairedEntry } from './use-paired-messages'
import type { RecordEntryView } from '@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/models'

/**
 * 根据 selectedIndex 从配对消息列表中找到对应的 PairedEntry。
 *
 * @param selectedIndex 当前选中行索引（null 表示未选中）
 * @param messages      消息列表 computed
 * @param recordedAt    录制时间 computed
 * @returns pairedMessages 和 selectedPairedEntry 两个 computed
 */
export function useSelectedEntry(
  selectedIndex: Ref<number | null>,
  messages: ComputedRef<RecordEntryView[]>,
  recordedAt: ComputedRef<string | undefined>
) {
  // 配对消息列表
  const pairedMessages = computed(() => buildPairedEntries(messages.value, recordedAt.value))

  // 根据 selectedIndex 查找对应的 PairedEntry
  const selectedPairedEntry = computed<PairedEntry | null>(() => {
    if (selectedIndex.value === null) return null
    for (const entry of pairedMessages.value) {
      if (entry.type === 'pair') {
        if (entry.req?.index === selectedIndex.value || entry.ack?.index === selectedIndex.value) {
          return entry
        }
      } else {
        if (entry.ntf?.index === selectedIndex.value || entry.ack?.index === selectedIndex.value) {
          return entry
        }
      }
    }
    return null
  })

  return { pairedMessages, selectedPairedEntry }
}
