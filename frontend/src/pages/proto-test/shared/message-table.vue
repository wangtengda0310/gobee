<template>
  <div style="flex: 1; min-height: 0;">
    <n-dropdown
      v-model:show="showContextMenu"
      :x="contextMenuX"
      :y="contextMenuY"
      placement="bottom-start"
      :options="contextMenuOptions"
      @select="handleContextMenuSelect"
    />
    <n-data-table
      :columns="columns"
      :data="displayedMessages"
      :row-props="rowProps"
      :row-class-name="rowClassName"
      :pagination="pagination"
      :bordered="true"
      :single-line="false"
      size="small"
      style="height: 100%"
      flex-height
    />
  </div>
</template>

<script setup lang="ts">
import { h, computed, ref } from 'vue'
import { NDataTable, NTag, NCheckbox, NDropdown, NButton } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import type { RecordEntryView } from '@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/models'
import { buildPairedEntries, type PairedEntry } from './composables/use-paired-messages'

const props = defineProps<{
  messages: RecordEntryView[]
  selectedIndex: number | null
  recordedAt?: string
  selectMode?: boolean
  selectedRowIds?: number[]
  enableAddToCase?: boolean // 发包改包页签启用"增加到用例"选项
  interceptedSeqIds?: Set<number> // 被拦截但未放行的消息 SeqID 集合
  // 页签变体：packet=发包改包（默认，含 Req 过滤按钮），testcase=测试用例（描述列，隐藏账号/时间/SeqID/结果）
  variant?: 'packet' | 'testcase'
  /** 是否允许拖拽调整步骤顺序（仅测试用例页签启用） */
  enableReorder?: boolean
}>()

const emit = defineEmits<{
  select: [index: number]
  reorder: [newOrder: RecordEntryView[]]
  toggleSelectRow: [rowId: number]
  deleteRow: [rowId: number]
  // 传递配对行中的原始消息（req + ack / ntf），而非配对行 ID，避免索引错位
  addToCase: [messages: RecordEntryView[]]
  /** 测试用例页签：双击描述列时聚焦底部描述输入框 */
  focusDescript: []
}>()

const pagination = { pageSize: 50 }

// 拖拽状态
const dragIndex = ref<number | null>(null)
const dragOverIndex = ref<number | null>(null)

// 过滤状态（发包改包页签）
const reqFilterActive = ref(false) // 开启后只显示带有 Req 数据的行
const ackFilterActive = ref(false) // 开启后只显示带有 Ack 数据的行
const ntfFilterActive = ref(false) // 开启后只显示带有 Ntf 数据的行

// 右键菜单状态
const showContextMenu = ref(false)
const contextMenuX = ref(0)
const contextMenuY = ref(0)
const contextMenuRowId = ref<number | null>(null)

// 右键菜单选项
const contextMenuOptions = computed(() => {
  // 发包改包页签：只显示"增加到用例"
  if (props.enableAddToCase) {
    // 根据当前右键行是否有 Req 决定菜单项状态
    const rowId = contextMenuRowId.value
    const entry = rowId !== null ? pairedMessages.value.find(p => p.id === rowId) : null
    const hasReq = !!entry?.req
    // 无 Req 时不显示自定义菜单（onContextmenu 已阻止右键弹窗）
    if (!hasReq) return []
    return [
      {
        label: '增加到用例',
        key: 'addToCase',
        disabled: props.selectMode, // 多选模式下禁用右键增加
      }
    ]
  }

  // 测试用例页签：只显示"删除行"
  return [
    {
      label: '删除行',
      key: 'delete',
      disabled: props.selectMode, // 多选模式下禁用右键删除
    }
  ]
})

// 将原始消息转换为配对条目
const pairedMessages = computed(() => buildPairedEntries(props.messages))

// 表格实际展示的条目（应用 Req/Ack/Ntf 过滤）
const displayedMessages = computed(() => {
  if (props.variant !== 'packet') return pairedMessages.value
  let rows = pairedMessages.value
  if (reqFilterActive.value) {
    rows = rows.filter(row => row.req !== null)
  }
  if (ackFilterActive.value) {
    rows = rows.filter(row => row.ack !== null)
  }
  if (ntfFilterActive.value) {
    rows = rows.filter(row => row.ntf !== null)
  }
  return rows
})

// 格式化时间
function formatTime(row: PairedEntry): string {
  const baseAt = props.recordedAt
  if (!baseAt) return ''
  const t = new Date(new Date(baseAt).getTime() + row.offset_ms)
  return t.toTimeString().slice(0, 8) + '.' + String(t.getMilliseconds()).padStart(3, '0')
}

// 取行的描述信息（测试用例页签"描述"列）
function getDescript(row: PairedEntry): string {
  return row.req?.descript || row.ntf?.descript || row.ack?.descript || ''
}

// 表格列定义（依赖 selectMode / variant / reqFilterActive，需响应式）
const columns = computed<DataTableColumns<PairedEntry>>(() => {
  const isTestcase = props.variant === 'testcase'
  const isPacket = props.variant === 'packet'

  const cols: DataTableColumns<PairedEntry> = [
    // 第一列：多选模式显示"多选"标题；非多选模式显示拖拽手柄
    {
      title: props.selectMode ? '多选' : '',
      key: 'check_drag',
      width: 30,
      render: (row) => {
        if (props.selectMode) {
          const checked = (props.selectedRowIds ?? []).includes(row.id)
          // 用例保存只保留 Req，非 Req 行禁止勾选
          const canSelect = row.req !== null
          return h(NCheckbox, {
            checked,
            disabled: !canSelect,
            title: canSelect ? '' : '仅可选择 Req 消息',
            'onUpdate:checked': () => {
              if (canSelect) emit('toggleSelectRow', row.id)
            },
          })
        }
        if (props.enableReorder) {
          return renderDragHandle(row)
        }
        return null
      }
    },
    { title: '#', key: 'id', width: 45 },
  ]

  if (!isTestcase) {
    cols.push({
      title: '账号',
      key: 'account_id',
      width: 70,
      render: (row) => {
        if (row.type === 'pair') {
          return row.req?.account_id || row.ack?.account_id || '-'
        }
        return row.ntf?.account_id || row.ack?.account_id || '-'
      }
    })
  }

  if (!isTestcase) {
    cols.push({
      title: '时间',
      key: 'time',
      width: 110,
      render: (row) => formatTime(row),
    })
  }

  cols.push({ title: 'MsgID', key: 'msg_id', width: 65 })

  cols.push({
    // 发包改包页签：标题旁显示过滤按钮
    title: !isPacket ? '请求(Req)' : () => h('span', { style: 'display: inline-flex; align-items: center; gap: 4px;' }, [
      '请求(Req)',
      h(NButton, {
        size: 'tiny',
        bordered: true,
        type: reqFilterActive.value ? 'primary' : 'default',
        title: '只显示Req',
        'data-testid': 'req-filter-btn',
        onClick: (e: MouseEvent) => {
          e.stopPropagation()
          reqFilterActive.value = !reqFilterActive.value
        },
      }, { default: () => reqFilterActive.value ? '取消' : '只显示Req' }),
    ]),
    key: 'req_name',
    width: 180,
    render: (row) => {
      if (row.type === 'single') return '-'
      return row.req ? row.req.msg_name : ''
    }
  })

  if (isTestcase) {
    // 测试用例页签："响应(Ack)"改为"描述"，展示用例 JSON 中的 descript
    cols.push({
      title: '描述',
      key: 'descript',
      width: 180,
      render: (row) => {
        const text = getDescript(row)
        return h('span', {
          'data-testid': 'descript-cell',
          style: 'cursor: text;',
          onDblclick: (e: MouseEvent) => {
            e.stopPropagation()
            let targetIndex: number | null = null
            if (row.type === 'pair') {
              targetIndex = row.req?.index ?? row.ack?.index ?? null
            } else {
              targetIndex = row.ntf?.index ?? row.ack?.index ?? null
            }
            if (targetIndex !== null) emit('select', targetIndex)
            emit('focusDescript')
          },
        }, text)
      },
    })
  } else {
    cols.push({
      title: isPacket ? () => h('span', { style: 'display: inline-flex; align-items: center; gap: 4px;' }, [
        '响应(Ack)',
        h(NButton, {
          size: 'tiny',
          bordered: true,
          type: ackFilterActive.value ? 'primary' : 'default',
          title: '只显示Ack',
          'data-testid': 'ack-filter-btn',
          onClick: (e: MouseEvent) => {
            e.stopPropagation()
            ackFilterActive.value = !ackFilterActive.value
          },
        }, { default: () => ackFilterActive.value ? '取消' : '只显示Ack' }),
        h(NButton, {
          size: 'tiny',
          bordered: true,
          type: ntfFilterActive.value ? 'primary' : 'default',
          title: '只显示Ntf',
          'data-testid': 'ntf-filter-btn',
          onClick: (e: MouseEvent) => {
            e.stopPropagation()
            ntfFilterActive.value = !ntfFilterActive.value
          },
        }, { default: () => ntfFilterActive.value ? '取消' : '只显示Ntf' }),
      ]) : '响应(Ack)',
      key: 'ack_name',
      width: 260,
      render: (row) => {
        if (row.type === 'single') {
          return row.ntf ? row.ntf.msg_name : (row.ack ? row.ack.msg_name : '')
        }
        if (row.ack) return row.ack.msg_name
        return h('span', { style: 'color: var(--n-text-color-3); font-style: italic;' }, '(等待中...)')
      }
    })
    cols.push({ title: 'SeqID', key: 'seq_id', width: 60 })
  }

  cols.push({
    title: '方向',
    key: 'direction',
    width: 95,
    render: (row) => {
      const dir = row.direction
      let type: 'info' | 'success' | 'warning' = 'info'
      if (dir === 'C->S,S->C') type = 'warning'
      else if (dir === 'S->C') type = 'success'
      return h(NTag, { size: 'small', type }, { default: () => dir })
    }
  })

  if (!isTestcase) {
    cols.push({
      title: '结果',
      key: 'result',
      width: 60,
      render: (row) => {
        let type: 'success' | 'warning' | 'default' = 'default'
        if (row.result === '成功') type = 'success'
        else if (row.result === '已发') type = 'warning'
        return h(NTag, { size: 'small', type }, { default: () => row.result })
      }
    })
  }

  return cols
})

// 完成行重排（drop 到目标行）
function applyRowDrop(targetRow: PairedEntry) {
  if (dragIndex.value === null || dragIndex.value === targetRow.id) {
    dragIndex.value = null
    dragOverIndex.value = null
    return
  }
  const fromRow = pairedMessages.value.find(p => p.id === dragIndex.value)
  const toRow = pairedMessages.value.find(p => p.id === targetRow.id)
  dragIndex.value = null
  dragOverIndex.value = null
  if (!fromRow || !toRow) return
  const sourceIndices = getRowMessageIndices(fromRow)
  const targetIndices = getRowMessageIndices(toRow)
  const targetPos = Math.min(...targetIndices)
  const remaining = props.messages.filter((_, i) => !sourceIndices.includes(i))
  const sourceMsgs = sourceIndices.map(i => props.messages[i])
  remaining.splice(targetPos, 0, ...sourceMsgs)
  emit('reorder', remaining.map((msg, idx) => { msg.index = idx; return msg }))
}

function renderDragHandle(row: PairedEntry) {
  return h('span', {
    draggable: true,
    'data-testid': 'row-drag-handle',
    style: 'cursor: grab; font-size: 16px; color: var(--n-text-color-3); user-select: none; display: inline-block; touch-action: none;',
    onDragstart: (e: DragEvent) => {
      e.stopPropagation()
      dragIndex.value = row.id
      if (e.dataTransfer) {
        e.dataTransfer.effectAllowed = 'move'
        // 避免浏览器默认拖拽选中整行文本
        e.dataTransfer.setData('text/plain', String(row.id))
      }
    },
    onDragend: (e: DragEvent) => {
      e.stopPropagation()
      dragIndex.value = null
      dragOverIndex.value = null
    },
    onClick: (e: MouseEvent) => e.stopPropagation(),
  }, '⠿')
}

function getRowMessageIndices(row: PairedEntry): number[] {
  const indices: number[] = []
  if (row.req !== null) indices.push(row.req.index)
  if (row.ack !== null) indices.push(row.ack.index)
  if (row.ntf !== null) indices.push(row.ntf.index)
  return indices
}

function isRowSelected(row: PairedEntry): boolean {
  if (props.selectedIndex === null) return false
  if (row.type === 'pair') {
    return row.req?.index === props.selectedIndex || row.ack?.index === props.selectedIndex
  }
  return row.ntf?.index === props.selectedIndex || row.ack?.index === props.selectedIndex
}

function isRowIntercepted(row: PairedEntry): boolean {
  const seqId = row.req?.seq_id ?? row.ack?.seq_id ?? row.ntf?.seq_id
  return seqId !== undefined && (props.interceptedSeqIds?.has(seqId) ?? false)
}

// 高亮所点击的行：通过 row-class-name + CSS 覆盖 td 背景（tr 背景会被 td 背景遮盖）
function rowClassName(row: PairedEntry): string {
  const classes: string[] = []
  if (isRowIntercepted(row)) classes.push('intercepted-row')
  if (isRowSelected(row)) classes.push('selected-row')
  return classes.join(' ')
}

function rowProps(row: PairedEntry) {
  // 多选模式：checkbox 处理勾选，行点击不干扰
  if (props.selectMode) {
    return {
      style: { cursor: 'default' },
    }
  }

  const reorderEnabled = props.enableReorder && !props.selectMode

  // 普通模式：点击行选中；拖动手柄负责拖拽，行本身作为 drop 目标
  return {
    style: {
      cursor: 'pointer',
      opacity: reorderEnabled && dragOverIndex.value === row.id ? 0.5 : 1,
    },
    onClick: () => {
      let targetIndex: number | null = null
      if (row.type === 'pair') {
        targetIndex = row.req?.index ?? row.ack?.index ?? null
      } else {
        targetIndex = row.ntf?.index ?? row.ack?.index ?? null
      }
      if (targetIndex !== null) emit('select', targetIndex)
    },
    onContextmenu: (e: MouseEvent) => {
      // 发包改包页签：无 Req 的行不显示自定义右键菜单（用例只保存 Req）
      if (props.enableAddToCase && row.req === null) return
      e.preventDefault()
      contextMenuRowId.value = row.id
      contextMenuX.value = e.clientX
      contextMenuY.value = e.clientY
      showContextMenu.value = true
    },
    ...(reorderEnabled ? {
      onDragover: (e: DragEvent) => {
        e.preventDefault()
        if (dragIndex.value !== null) dragOverIndex.value = row.id
      },
      onDragleave: () => { if (dragOverIndex.value === row.id) dragOverIndex.value = null },
      onDrop: (e: DragEvent) => {
        e.preventDefault()
        applyRowDrop(row)
      },
    } : {}),
  }
}

// 暴露 scrollToRow 方法供父组件调用
function scrollToRow(index: number) {
  // 通过 data-table 的 row-key 找到对应行并滚动到可视区域
  const tableEl = document.querySelector('.n-data-table tbody')
  if (tableEl) {
    const rows = tableEl.querySelectorAll('tr')
    // 找到对应原始消息索引的行（考虑配对关系，找到包含该索引的配对行）
    for (let i = 0; i < rows.length; i++) {
      const rowData = displayedMessages.value[i]
      if (!rowData) continue
      const indices = getRowMessageIndices(rowData)
      if (indices.includes(index)) {
        rows[i].scrollIntoView({ behavior: 'smooth', block: 'center' })
        break
      }
    }
  }
}

defineExpose({ scrollToRow })

// 右键菜单选择处理
function handleContextMenuSelect(key: string) {
  if (key === 'delete' && contextMenuRowId.value !== null) {
    emit('deleteRow', contextMenuRowId.value)
  } else if (key === 'addToCase' && contextMenuRowId.value !== null) {
    // 从配对行提取实际包含的原始消息（req + ack / ntf），避免索引错位
    const entry = pairedMessages.value.find(p => p.id === contextMenuRowId.value)
    if (entry) {
      const msgs: RecordEntryView[] = []
      if (entry.req) msgs.push(entry.req)
      // 只保存 Req，不保存 Ack/Ntf
      emit('addToCase', msgs)
    }
  }
  showContextMenu.value = false
  contextMenuRowId.value = null
}
</script>

<style scoped>
/* 被拦截但未放行的行：微黄背景 + 左侧橙色标记 */
:deep(tr.intercepted-row td) {
  background-color: rgba(240, 160, 32, 0.10) !important;
}
:deep(tr.intercepted-row td:first-child) {
  box-shadow: inset 3px 0 0 #f0a020;
}
/* 所点击（选中）的行高亮，优先级高于拦截样式 */
:deep(tr.selected-row td) {
  background-color: rgba(32, 128, 240, 0.18) !important;
}
:deep([data-testid="row-drag-handle"]:active) {
  cursor: grabbing;
}
</style>
