<template>
  <div class="req-card-editor">
    <!-- 工具栏（测试用例页由 paired-payload-editor 统一提供） -->
    <div v-if="!hideToolbar" style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px;">
      <div v-if="!compactTitle" style="font-size: 13px; color: var(--n-text-color-2);">
        {{ entry?.msg_name }} ({{ entry?.msg_id }})
      </div>
      <div v-else style="flex: 1;" />
      <div style="display: flex; gap: 8px;">
        <n-button size="tiny" @click="handleFormat">格式化</n-button>
        <n-button v-if="showApply" size="tiny" type="primary" :loading="applying" :disabled="!canApply" @click="handleApply">应用</n-button>
      </div>
    </div>

    <!-- 卡片式编辑器 -->
    <div v-if="payload" class="card-container">
      <n-card size="small" :bordered="true">
        <template #header>
          <div style="display: flex; justify-content: space-between; align-items: center;">
            <span style="font-size: 13px; font-weight: 500;">Payload 字段</span>
            <n-tag v-if="!isValid" type="error" size="tiny">格式错误</n-tag>
          </div>
        </template>

        <!-- 字段列表 -->
        <div class="field-list">
          <field-item
            v-for="(value, key) in payload"
            :key="key"
            :ref="setFieldRef(String(key))"
            :field-key="String(key)"
            :value="value"
            :path="[key]"
            :initial-state="entryFieldValues[String(key)]"
            :msg-name="entry?.msg_name ?? ''"
            @update:value="handleFieldUpdate"
            @change="checkChanges"
          />
        </div>
      </n-card>
    </div>

    <!-- 空状态 -->
    <div v-else style="display: flex; align-items: center; justify-content: center; height: 200px; color: var(--n-text-color-3); font-size: 13px;">
      无数据
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, type ComponentPublicInstance } from 'vue'
import { NCard, NButton, NTag, useMessage } from 'naive-ui'
import FieldItem from './field-item.vue'
import type { FieldFourState } from './field-item.vue'
import type { RecordEntryView } from '@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/models'

const props = withDefaults(defineProps<{
  entry: RecordEntryView | null
  showApply?: boolean
  /** 描述等外部字段有未保存修改时也启用「应用」 */
  extraApplyEnabled?: boolean
  /** 隐藏顶部工具栏（由父级 paired-payload-editor 提供） */
  hideToolbar?: boolean
  /** 隐藏协议名，仅保留按钮区 */
  compactTitle?: boolean
}>(), {
  showApply: true,
  extraApplyEnabled: false,
  hideToolbar: false,
  compactTitle: false,
})

const emit = defineEmits<{
  apply: [index: number, payload: Record<string, any>]
  'config-change': []
}>()

const message = useMessage()

// payload 直接引用 entry.payload（唯一数据源，防御 Wails 将 null map 反序列化为 undefined）
const payload = computed(() => props.entry?.payload ?? null)

// 从 entry.field_values 提取当前消息的字段配置（用于恢复 field-item 的持久化状态）
const entryFieldValues = computed(() => props.entry?.field_values ?? {})

// field-item 组件引用（用于收集4态值）
const fieldRefs = ref<Record<string, ComponentPublicInstance & {
  getActiveValue: () => any
  getFourState: () => FieldFourState
}>>({})

// 设置 field-item ref 的安全方法
function setFieldRef(key: string) {
  return (el: any) => {
    if (el && fieldRefs.value) {
      fieldRefs.value[key] = el
    }
  }
}

// 应用状态
const applying = ref(false)

// 变更标记：由子组件值变化时手动触发
const hasChanges = ref(false)

// 监听 entry 变化，重置状态
watch(() => props.entry, () => {
  fieldRefs.value = {}
  hasChanges.value = false
}, { immediate: true })

// 合法性验证
const isValid = computed(() => {
  return payload.value !== null && typeof payload.value === 'object'
})

const canApply = computed(() => {
  if (!isValid.value) return false
  return hasChanges.value || props.extraApplyEnabled
})

// 检查是否有修改
function checkChanges() {
  if (!payload.value) {
    hasChanges.value = false
    emit('config-change')
    return
  }
  try {
    const current = buildCurrentPayload()
    const currentJSON = JSON.stringify(current)
    const origJSON = JSON.stringify(props.entry!.payload)
    hasChanges.value = currentJSON !== origJSON
  } catch {
    hasChanges.value = true
  }
  emit('config-change')
}

// 从所有 field-item 组件收集当前有效值
function buildCurrentPayload(): Record<string, any> {
  if (!payload.value) return {}

  const result: Record<string, any> = {}
  for (const key of Object.keys(payload.value)) {
    const fieldRef = fieldRefs.value[key]
    if (fieldRef && typeof fieldRef.getActiveValue === 'function') {
      result[key] = fieldRef.getActiveValue()
    } else {
      result[key] = payload.value[key]
    }
  }
  return result
}

// 收集所有字段的4态数据
function buildFourStatePayload(): Record<string, FieldFourState> {
  if (!payload.value) return {}

  const result: Record<string, FieldFourState> = {}
  for (const key of Object.keys(payload.value)) {
    const fieldRef = fieldRefs.value[key]
    if (fieldRef && typeof fieldRef.getFourState === 'function') {
      result[key] = fieldRef.getFourState()
    }
  }
  return result
}

// 处理字段更新
function handleFieldUpdate(event: { path: (string | number)[]; value: any }) {
  // 卡片模式下字段变更只用于变更检测，实际值由 buildCurrentPayload 收集
}

// 格式化（卡片模式无实际操作）
function handleFormat() {
  if (!payload.value) {
    message.warning('无数据可格式化')
    return
  }
  message.success('已格式化')
}

// 应用修改
async function handleApply() {
  if (!props.entry || !canApply.value) return

  applying.value = true
  try {
    const newPayload = buildCurrentPayload()
    emit('apply', props.entry.index, newPayload)
    hasChanges.value = false
  } catch (e: any) {
    message.error('应用失败: ' + (e.message || e))
  } finally {
    applying.value = false
  }
}

// 暴露给父组件
defineExpose({
  buildFourStatePayload,
  buildCurrentPayload,
})
</script>

<style scoped>
.req-card-editor {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.card-container {
  flex: 1;
  overflow-y: auto;
}

.field-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
</style>
