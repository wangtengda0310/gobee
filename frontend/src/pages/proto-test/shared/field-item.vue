<template>
  <div class="field-item" :style="{ paddingLeft: `${depth * 16}px` }">
    <!-- 标签行 -->
    <div class="field-label">
      <n-button
        v-if="isObject"
        text
        size="tiny"
        @click="toggleExpanded"
        style="padding: 0; margin-right: 4px;"
      >
        <template #icon>
          <n-icon>
            <ChevronForwardOutline v-if="!expanded" />
            <ChevronDownOutline v-else />
          </n-icon>
        </template>
      </n-button>
      <span class="label-text">{{ String(fieldKey) }}</span>
      <n-tag size="tiny" :type="getTypeTagType(value)" style="margin-left: 8px;">
        {{ getValueType(value) }}
      </n-tag>
    </div>

    <!-- 值编辑区 -->
    <div v-if="!isObject || expanded" class="field-value">
      <!-- 对象和数组类型不支持组件选择 -->
      <template v-if="isObject || isArray">
        <!-- 嵌套对象 -->
        <div v-if="isObject" class="nested-fields">
          <field-item
            v-for="(subValue, subKey) in value"
            :key="String(subKey)"
            :field-key="String(subKey)"
            :value="subValue"
            :path="[...path, subKey]"
            :depth="depth + 1"
            @update:value="handleNestedUpdate"
          />
        </div>

        <!-- 数组 -->
        <div v-else-if="isArray" class="array-items">
          <div
            v-for="(item, index) in value"
            :key="index"
            class="array-item"
          >
            <div class="array-item-header">
              <span class="array-index">[{{ index }}]</span>
              <n-button size="tiny" text type="error" @click="handleArrayRemove(Number(index))">
                <template #icon>
                  <n-icon><TrashOutline /></n-icon>
                </template>
              </n-button>
            </div>
            <field-item
              :field-key="String(index)"
              :value="item"
              :path="[...path, index]"
              :depth="depth + 1"
              @update:value="handleArrayUpdate"
            />
          </div>
          <n-button size="tiny" dashed @click="handleArrayAdd">
            <template #icon>
              <n-icon><AddOutline /></n-icon>
            </template>
            添加项
          </n-button>
        </div>
      </template>

      <!-- 基础类型：支持组件选择 -->
      <template v-else>
        <div class="input-selector">
          <!-- 组件类型选择下拉菜单 -->
          <n-select
            v-model:value="inputType"
            :options="inputTypeOptions"
            size="small"
            style="width: 120px; margin-right: 8px;"
          />

          <!-- 原始值显示（只读，使用 :value 单向绑定） -->
          <div v-if="inputType === 'original'" class="original-value">
            <span class="value-label">原始值:</span>
            <n-input
              :value="String(value ?? '')"
              size="small"
              readonly
              style="flex: 1;"
            />
          </div>

          <!-- 范围输入组件 -->
          <range-input
            v-if="inputType === 'range'"
            v-model="rangeValue"
          />

          <!-- 枚举值选择组件 -->
          <enum-select
            v-if="inputType === 'enum'"
            v-model="enumValue"
          />

          <!-- 组合选择组件 -->
          <combo-select
            v-if="inputType === 'combo'"
            v-model="comboValue"
          />

          <!-- 动态变量选择组件 -->
          <variable-select
            v-if="inputType === 'variable'"
            v-model="varName"
            :msg-name="msgName"
            @change="emit('change')"
          />
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { NInput, NInputNumber, NSwitch, NButton, NTag, NIcon, NSelect } from 'naive-ui'
import {
  ChevronForwardOutline,
  ChevronDownOutline,
  TrashOutline,
  AddOutline,
} from '@vicons/ionicons5'
import RangeInput from './range-input.vue'
import EnumSelect from './enum-select.vue'
import ComboSelect from './combo-select.vue'
import VariableSelect from './variable-select.vue'

// 输入组件类型
export type InputComponentType = 'original' | 'range' | 'enum' | 'combo' | 'variable'

// 4态完整数据（一次性提交给后端）
// 注意：字段名必须与 Go FieldValues 的 JSON tag 一致（snake_case）
export interface FieldFourState {
  range_value: { start: number; step: number; end: number }
  enum_value: any[]
  combo_value: any[]
  input_type: InputComponentType
  /** 动态变量名（ShortName），仅 input_type="variable" 时有值，与 Go FieldValues JSON tag 对齐 */
  variable_name?: string
}

// 字段更新事件类型
interface FieldUpdateEvent {
  path: (string | number)[]
  value: any
}

// 输入组件选项（变量模式作为正式功能始终开放）
const inputTypeOptions = computed(() => [
  { label: '原始值', value: 'original' },
  { label: '范围', value: 'range' },
  { label: '枚举', value: 'enum' },
  { label: '组合', value: 'combo' },
  { label: '变量', value: 'variable' },
])

const props = withDefaults(defineProps<{
  fieldKey: string | number
  value: any
  path: (string | number)[]
  depth?: number
  /** 持久化的字段配置（从 entry.field_values[fieldKey] 传入），用于恢复 input_type 等状态 */
  initialState?: {
    input_type?: string
    variable_name?: string
    range_value?: { start: number; step: number; end: number }
    enum_value?: any[]
    combo_value?: any[]
  } | null
  /** 当前 Req 的 msg_name，透传给 variable-select 按 AvailableReqs 过滤变量 */
  msgName?: string
}>(), {
  depth: 0,
  initialState: null,
  msgName: '',
})

const emit = defineEmits<{
  'update:value': [event: FieldUpdateEvent]
  'change': []
}>()

// 展开状态
const expanded = ref(true)

// 当 entry 切换时（选择不同消息），initialState 会变化，需要恢复状态
watch(() => props.initialState, (state) => {
  if (state) {
    const restoredRange = state.range_value
    const restoredRangeIsValid = restoredRange && typeof restoredRange.start === 'number' && typeof restoredRange.end === 'number' && typeof restoredRange.step === 'number'
    if (restoredRangeIsValid) rangeValue.value = { ...restoredRange! }
    if (Array.isArray(state.enum_value)) enumValue.value = [...state.enum_value]
    if (Array.isArray(state.combo_value)) comboValue.value = [...state.combo_value]
    if (state.variable_name) varName.value = state.variable_name
    if (state.input_type) inputType.value = state.input_type as InputComponentType
  } else {
    // 无持久化状态，重置为默认
    inputType.value = 'original'
    varName.value = ''
  }
}, { immediate: false })

// 4个完全独立的本地状态，初始化后互不干扰
// 如果有持久化的 initialState，从中恢复；否则使用默认值
const restoredRange = props.initialState?.range_value
const restoredRangeIsValid = restoredRange && typeof restoredRange.start === 'number' && typeof restoredRange.end === 'number' && typeof restoredRange.step === 'number'

const rangeValue = ref<{ start: number; step: number; end: number }>(
  restoredRangeIsValid ? { ...restoredRange! } : { start: 0, step: 1, end: 10 }
)
const enumValue = ref<any[]>(Array.isArray(props.initialState?.enum_value) ? [...(props.initialState!.enum_value!)] : [])
const comboValue = ref<any[]>(Array.isArray(props.initialState?.combo_value) ? [...(props.initialState!.combo_value!)] : [])
// 变量模式选中的变量名（ShortName）
const varName = ref(props.initialState?.variable_name ?? '')

// 输入组件类型：从持久化状态恢复，默认原始值
const inputType = ref<InputComponentType>(
  (props.initialState?.input_type as InputComponentType) || 'original'
)

// 4态值完全独立：切换类型时各值互不干扰，仅在首次切换到某类型时做初始化
watch(inputType, (newType, oldType) => {
  if (newType === oldType) return

  switch (newType) {
    case 'range':
      // 首次切换到范围模式时，基于原始值初始化（仅当范围值还是默认值时）
      if (rangeValue.value.start === 0 && rangeValue.value.end === 10) {
        const numValue = Number(props.value) || 0
        rangeValue.value = { start: numValue, step: 1, end: numValue + 10 }
      }
      break
    // enum 和 combo 的值由各自的子组件维护，切换时保持已有值不变
  }
  // 通知父组件重新检查变更状态
  emit('change')
})

// 监听子组件值变化，通知父组件重新检查变更状态
watch([rangeValue, enumValue, comboValue, varName], () => {
  emit('change')
}, { deep: true })

// 类型检测
const type = computed(() => {
  if (props.value === null) return 'null'
  return typeof props.value
})

const isPrimitive = computed(() => {
  const t = type.value
  return t === 'string' || t === 'number' || t === 'boolean'
})

const isObject = computed(() => {
  return type.value === 'object' && Array.isArray(props.value) === false && props.value !== null
})

const isArray = computed(() => {
  return Array.isArray(props.value)
})

// 获取当前字段在"单条发送"语义下的有效值
// 卡片模式不直接编辑字段值：range/enum/combo 是迭代配置（通过 getFourState() 提交给后端展开），
// 单条 payload 一律使用原始类型的原始值，避免逗号串（enum/combo）、
// 范围对象（range）、String() 类型污染（original）泄漏进发送的消息
function getActiveValue(): any {
  return props.value
}

// 获取完整的4态数据（snake_case 键名与 Go FieldValues JSON tag 对齐）
function getFourState(): FieldFourState {
  const state: FieldFourState = {
    range_value: { ...rangeValue.value },
    enum_value: [...enumValue.value],
    combo_value: [...comboValue.value],
    input_type: inputType.value,
  }
  if (inputType.value === 'variable' && varName.value) {
    state.variable_name = varName.value
  }
  return state
}

// 切换展开
function toggleExpanded() {
  expanded.value = !expanded.value
}

// 获取值类型显示
function getValueType(val: any): string {
  if (val === null) return 'null'
  if (Array.isArray(val)) return `array[${val.length}]`
  return typeof val
}

// 获取标签类型
function getTypeTagType(val: any): 'default' | 'info' | 'success' | 'warning' | 'error' {
  if (val === null) return 'default'
  if (Array.isArray(val)) return 'info'
  const t = typeof val
  if (t === 'string') return 'success'
  if (t === 'number') return 'warning'
  if (t === 'boolean') return 'error'
  return 'default'
}

// 处理嵌套字段更新（直接透传）
function handleNestedUpdate(event: FieldUpdateEvent) {
  emit('update:value', event)
}

// 处理数组项更新
function handleArrayUpdate(event: FieldUpdateEvent) {
  const newValue = [...props.value]
  const lastIndex = event.path.length - 1
  const index = Number(event.path[lastIndex])
  newValue[index] = event.value
  emit('update:value', { path: props.path, value: newValue })
}

// 处理数组添加
function handleArrayAdd() {
  const newValue = [...props.value, null]
  emit('update:value', { path: props.path, value: newValue })
}

// 处理数组删除
function handleArrayRemove(index: number) {
  const numericIndex = Number(index)
  const newValue = props.value.filter((_: any, i: number) => i !== numericIndex)
  emit('update:value', { path: props.path, value: newValue })
}

// 暴露方法供父组件调用
defineExpose({
  getActiveValue,
  getFourState,
})
</script>

<style scoped>
.field-item {
  border: 1px solid var(--n-border-color);
  border-radius: var(--n-border-radius);
  padding: 8px;
  margin-bottom: 8px;
  background-color: var(--n-card-color);
}

.field-label {
  display: flex;
  align-items: center;
  margin-bottom: 8px;
  font-size: 13px;
  font-weight:  500;
}

.label-text {
  color: var(--n-text-color-1);
}

.field-value {
  margin-left: 20px;
}

.input-selector {
  display: flex;
  align-items: center;
  gap: 8px;
}

.original-value {
  display: flex;
  align-items: center;
  flex: 1;
  gap: 8px;
}

.value-label {
  font-size: 13px;
  color: var(--n-text-color-2);
  white-space: nowrap;
  min-width: 60px;
}

.nested-fields {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.array-items {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.array-item {
  border: 1px dashed var(--n-border-color);
  border-radius: var(--n-border-radius);
  padding: 8px;
  background-color: var(--n-color-modal);
}

.array-item-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.array-index {
  font-size: 12px;
  color: var(--n-text-color-2);
  font-weight: 500;
}
</style>
