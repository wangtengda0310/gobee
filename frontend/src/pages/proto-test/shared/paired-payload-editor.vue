<template>
  <div v-if="entry" style="flex: 0 0 300px; display: flex; flex-direction: column;">
    <!-- 配对行：测试用例仅 Req；录制/重放为 Req + Ack 双栏 -->
    <template v-if="entry.type === 'pair'">
      <div :style="isTestcaseVariant ? 'display: flex; flex-direction: column; min-height: 0;' : 'display: flex; gap: 8px; height: 100%;'">
        <!-- 左侧 / 唯一：Req（可编辑） -->
        <div :style="isTestcaseVariant ? 'flex: 1; display: flex; flex-direction: column; min-width: 0; min-height: 0;' : 'flex: 1; display: flex; flex-direction: column; min-width: 0;'">
          <!-- 工具栏：协议名 · 描述(用例) · 操作按钮 -->
          <div style="display: flex; align-items: center; gap: 8px; margin-bottom: 4px; flex-wrap: wrap;">
            <div style="font-size: 13px; color: var(--n-text-color-2); white-space: nowrap; flex-shrink: 0;">
              {{ formatMsgLabel(entry.req) }}
            </div>
            <n-input
              v-if="isTestcaseVariant && entry.req"
              ref="descriptInputRef"
              :value="descript"
              placeholder="步骤描述"
              size="small"
              style="flex: 1; min-width: 140px;"
              :input-props="{ 'data-testid': 'case-descript-input' } as any"
              @update:value="emit('update:descript', $event)"
            />
            <div style="display: flex; gap: 8px; flex-shrink: 0; margin-left: auto;">
              <n-button size="tiny" @click="toggleReqEditMode">
                {{ reqEditMode === 'card' ? 'JSON模式' : '卡片模式' }}
              </n-button>
              <template v-if="showHeaderActions">
                <n-button size="tiny" @click="handleFormat">格式化</n-button>
                <n-button v-if="props.showReqApply" size="tiny" type="primary" :loading="applyingReq" :disabled="!canApplyReq" @click="handleApply">应用</n-button>
              </template>
            </div>
          </div>

          <!-- 卡片编辑模式 -->
          <req-card-editor
            ref="reqCardEditorRef"
            v-if="reqEditMode === 'card' && entry.req"
            :entry="entry.req"
            :show-apply="!isTestcaseVariant && props.showReqApply"
            :extra-apply-enabled="props.extraApplyEnabled"
            :hide-toolbar="isTestcaseVariant"
            :compact-title="isTestcaseVariant"
            @apply="handleCardApply"
            @config-change="onCardConfigChange"
            style="flex: 1; overflow: hidden;"
          />

          <!-- JSON 编辑模式 -->
          <template v-else-if="reqEditMode === 'json'">
            <n-input
              type="textarea"
              v-model:value="reqJsonEditValue"
              :rows="12"
              :status="reqInputStatus"
              style="font-family: monospace; font-size: 12px;"
            />
            <div v-if="!isValidReq" style="font-size: 12px; color: var(--n-error-color); margin-top: 4px;">
              JSON 格式错误
            </div>
          </template>
        </div>

        <!-- 右侧：Ack（只读，测试用例页签不展示） -->
        <div v-if="!isTestcaseVariant" style="flex: 1; display: flex; flex-direction: column; min-width: 0;">
          <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 4px;">
            <div style="font-size: 13px; color: var(--n-text-color-2);">
              <template v-if="entry.ack">
                {{ formatMsgLabel(entry.ack) }}
              </template>
              <template v-else>
                <span style="color: var(--n-text-color-3); font-style: italic;">等待响应...</span>
              </template>
            </div>
          </div>
          <n-input
            v-if="entry.ack"
            type="textarea"
            :value="formatJSON(entry.ack.payload)"
            :rows="12"
            readonly
            style="font-family: monospace; font-size: 12px;"
          />
          <div v-else style="flex: 1; display: flex; align-items: center; justify-content: center; border: 1px dashed var(--n-border-color); border-radius: var(--n-border-radius); color: var(--n-text-color-3); font-size: 13px;">
            等待响应...
          </div>
        </div>
      </div>
    </template>

    <!-- 单行：Ntf/Ack 只读展示 -->
    <template v-else>
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 4px;">
        <div style="font-size: 13px; color: var(--n-text-color-2);">
          {{ formatMsgLabel(displayMsg) }}
        </div>
      </div>
      <n-input
        type="textarea"
        :value="formatJSON(displayMsg?.payload)"
        :rows="12"
        readonly
        style="font-family: monospace; font-size: 12px;"
      />
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { NInput, NButton, useMessage } from 'naive-ui'
import ReqCardEditor from './req-card-editor.vue'
import type { PairedEntry } from './composables/use-paired-messages'
import type { RecordEntryView } from '@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/models'

const props = withDefaults(defineProps<{
  entry: PairedEntry | null
  showReqApply?: boolean
  /** 测试用例页签：描述等外部字段有未保存修改时也启用「应用」 */
  extraApplyEnabled?: boolean
  /** default=录制双栏；testcase=仅 Req + 描述工具栏 */
  variant?: 'default' | 'testcase'
  /** 测试用例步骤描述（variant=testcase） */
  descript?: string
  /** 拦截模式下内存中的编辑 payload，优先于 entry.req.payload */
  reqPayloadOverride?: Record<string, any> | null
}>(), {
  showReqApply: true,
  extraApplyEnabled: false,
  variant: 'default',
  descript: '',
  reqPayloadOverride: null,
})

const emit = defineEmits<{
  apply: [index: number, payload: Record<string, any>]
  'config-change': []
  'update:descript': [value: string]
}>()

const isTestcaseVariant = computed(() => props.variant === 'testcase')
const showHeaderActions = computed(() => isTestcaseVariant.value || reqEditMode.value === 'json')

// req-card-editor ref
const reqCardEditorRef = ref<InstanceType<typeof ReqCardEditor> | null>(null)
const descriptInputRef = ref<InstanceType<typeof NInput> | null>(null)
const cardChangeTick = ref(0)

function onCardConfigChange() {
  cardChangeTick.value++
  emit('config-change')
}

function formatMsgLabel(msg: RecordEntryView | null | undefined): string {
  if (!msg) return ''
  return `${msg.msg_name} (${msg.msg_id})`
}

function focusDescriptInput() {
  descriptInputRef.value?.focus()
}

// 代理暴露：收集 req 编辑器中所有字段的4态数据
function getFieldFourStates(): Record<string, any> {
  return reqCardEditorRef.value?.buildFourStatePayload() ?? {}
}

function getCurrentReqPayload(): Record<string, any> | null {
  if (!props.entry || props.entry.type !== 'pair' || !props.entry.req) return null

  if (reqEditMode.value === 'json') {
    try {
      return JSON.parse(reqJsonEditValue.value)
    } catch {
      return null
    }
  }
  return reqCardEditorRef.value?.buildCurrentPayload() ?? null
}

function hasCardPayloadChanges(): boolean {
  if (!props.entry?.req || !reqCardEditorRef.value) return false
  try {
    const current = reqCardEditorRef.value.buildCurrentPayload()
    return JSON.stringify(current) !== JSON.stringify(props.entry.req.payload)
  } catch {
    return false
  }
}

// 检测字段配置变更（input_type 从 original 变为 range/enum/combo/variable）
function hasFieldConfigChanges(): boolean {
  if (!reqCardEditorRef.value) return false
  try {
    const states = reqCardEditorRef.value.buildFourStatePayload()
    return Object.values(states).some((s: any) => s?.input_type && s.input_type !== 'original')
  } catch {
    return false
  }
}

function hasReqPayloadChanges(): boolean {
  if (reqEditMode.value === 'card') {
    cardChangeTick.value // 依赖卡片变更 tick
    return hasCardPayloadChanges()
  }
  return hasChangesReq.value
}

defineExpose({ getFieldFourStates, getCurrentReqPayload, hasReqPayloadChanges, focusDescriptInput })

const reqEditMode = ref<'card' | 'json'>('json')
const reqJsonEditValue = ref('')

function toggleReqEditMode() {
  if (reqEditMode.value === 'card') {
    if (reqCardEditorRef.value) {
      const payload = reqCardEditorRef.value.buildCurrentPayload()
      reqJsonEditValue.value = JSON.stringify(payload, null, 2)
    }
    reqEditMode.value = 'json'
  } else {
    reqEditMode.value = 'card'
  }
}

// 检测 req 是否持有非 original 的字段配置（variable/range/enum/combo）。
// 选中该类 Req 时默认进入卡片模式——这些字段只能在卡片模式下编辑/查看，
// 若默认 JSON 模式会让用户看不到已配置的迭代/变量信息。
function reqHasFieldConfig(req: PairedEntry['req']): boolean {
  if (!req?.field_values) return false
  return Object.values(req.field_values).some(
    (fv: any) => fv?.input_type && fv.input_type !== 'original'
  )
}

function handleCardApply(index: number, payload: Record<string, any>) {
  emit('apply', index, payload)
}

const message = useMessage()
const applyingReq = ref(false)

const displayMsg = computed(() => {
  if (!props.entry) return null
  return props.entry.ntf ?? props.entry.ack ?? null
})

function resolveReqPayload(req: PairedEntry['req']): Record<string, any> | undefined {
  if (!req) return undefined
  return props.reqPayloadOverride ?? req.payload
}

watch(
  () => [props.entry, props.reqPayloadOverride] as const,
  ([val]) => {
    if (!val) {
      reqJsonEditValue.value = ''
      return
    }
    if (val.type === 'pair') {
      const payload = resolveReqPayload(val.req)
      reqJsonEditValue.value = payload ? formatJSON(payload) : ''
    } else {
      reqJsonEditValue.value = ''
    }
    // 有字段配置的 Req 默认卡片模式，否则 JSON 模式
    reqEditMode.value = val.type === 'pair' && reqHasFieldConfig(val.req) ? 'card' : 'json'
    cardChangeTick.value = 0
  },
  { immediate: true, deep: true },
)

const isValidReq = computed(() => {
  if (!reqJsonEditValue.value) return true
  try {
    JSON.parse(reqJsonEditValue.value)
    return true
  } catch {
    return false
  }
})
const reqInputStatus = computed(() => isValidReq.value ? undefined : 'error')

const hasChangesReq = computed(() => {
  const baseline = resolveReqPayload(props.entry?.req ?? null)
  if (!baseline) return false
  return reqJsonEditValue.value !== JSON.stringify(baseline, null, 2)
})

const canApplyReq = computed(() => {
  if (reqEditMode.value === 'json' && !isValidReq.value) return false
  if (props.extraApplyEnabled) return true
  if (reqEditMode.value === 'json') return hasChangesReq.value
  cardChangeTick.value
  return hasCardPayloadChanges() || hasFieldConfigChanges()
})

function formatJSON(obj: Record<string, any> | undefined): string {
  if (!obj) return ''
  try {
    return JSON.stringify(obj, null, 2)
  } catch {
    return String(obj)
  }
}

function handleFormat() {
  if (reqEditMode.value === 'card') {
    message.success('已格式化')
    return
  }
  if (!isValidReq.value) {
    message.warning('JSON 格式错误，无法格式化')
    return
  }
  reqJsonEditValue.value = formatJSON(JSON.parse(reqJsonEditValue.value))
  message.success('已格式化')
}

async function handleApply() {
  if (!props.entry?.req || !canApplyReq.value) return

  applyingReq.value = true
  try {
    let payload: Record<string, any>
    if (reqEditMode.value === 'json') {
      if (!isValidReq.value) return
      payload = JSON.parse(reqJsonEditValue.value)
    } else {
      const fromCard = reqCardEditorRef.value?.buildCurrentPayload()
      if (!fromCard) return
      payload = fromCard
    }
    emit('apply', props.entry.req.index, payload)
  } catch (e: any) {
    message.error('应用失败: ' + (e.message || e))
  } finally {
    applyingReq.value = false
  }
}
</script>
