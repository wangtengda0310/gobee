<template>
  <div v-if="msg" style="flex: 0 0 300px; display: flex; flex-direction: column;">
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 4px;">
      <div style="font-size: 13px; color: var(--n-text-color-2);">
        {{ msg.msg_name }} (MsgID={{ msg.msg_id }}, SeqID={{ msg.seq_id }})
      </div>
      <div style="display: flex; gap: 8px;">
        <n-button size="tiny" @click="handleFormat">格式化</n-button>
        <n-button size="tiny" type="primary" :loading="applying" :disabled="!isValid || !hasChanges" @click="handleApply">应用</n-button>
      </div>
    </div>
    <n-input
      type="textarea"
      v-model:value="editValue"
      :rows="12"
      :status="inputStatus"
      style="font-family: monospace; font-size: 12px;"
    />
    <div v-if="!isValid" style="font-size: 12px; color: var(--n-error-color); margin-top: 4px;">
      JSON 格式错误
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { NInput, NButton, useMessage } from 'naive-ui'
import type { RecordEntryView, RecordFileData } from '@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/models'

const props = defineProps<{
  msg: RecordEntryView | null
  recordData: RecordFileData | null
  filePath: string
}>()

const emit = defineEmits<{
  apply: [index: number, payloadJSON: string]
}>()

const message = useMessage()
const editValue = ref('')
const applying = ref(false)

// 当选中消息变化时，同步编辑值
watch(() => props.msg, (val) => {
  if (val) {
    editValue.value = formatJSON(val.payload)
  } else {
    editValue.value = ''
  }
}, { immediate: true })

// JSON 合法性验证
const isValid = computed(() => {
  try {
    JSON.parse(editValue.value)
    return true
  } catch {
    return false
  }
})

const inputStatus = computed(() => isValid.value ? undefined : 'error')

// 是否有修改
const hasChanges = computed(() => {
  if (!props.msg) return false
  return editValue.value !== JSON.stringify(props.msg.payload)
})

function formatJSON(payload: Record<string, any>): string {
  try {
    return JSON.stringify(payload, null, 2)
  } catch {
    return JSON.stringify(payload)
  }
}

function handleFormat() {
  if (!isValid.value) {
    message.warning('JSON 格式错误，无法格式化')
    return
  }
  // editValue 是用户编辑中的 JSON 字符串，解析后重新格式化
  try {
    const parsed = JSON.parse(editValue.value)
    editValue.value = JSON.stringify(parsed, null, 2)
  } catch {
    // JSON.parse 已通过 isValid 验证，理论上不会走到这里
  }
  message.success('已格式化')
}

async function handleApply() {
  if (!props.msg || !isValid.value) return
  applying.value = true
  try {
    emit('apply', props.msg.index, editValue.value)
    message.success('已应用修改')
  } catch (e: any) {
    message.error('应用失败: ' + (e.message || e))
  } finally {
    applying.value = false
  }
}
</script>
