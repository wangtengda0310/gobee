<template>
  <div class="range-input">
    <div class="input-row">
      <label class="input-label">起始值:</label>
      <n-input-number
        v-model:value="localValue.start"
        size="small"
        :show-button="false"
        placeholder="起始值"
        class="input-field"
        @update:value="handleChange"
      />
    </div>
    <div class="input-row">
      <label class="input-label">步长:</label>
      <n-input-number
        v-model:value="localValue.step"
        size="small"
        :show-button="false"
        placeholder="步长"
        class="input-field"
        @update:value="handleChange"
      />
    </div>
    <div class="input-row">
      <label class="input-label">终值:</label>
      <n-input-number
        v-model:value="localValue.end"
        size="small"
        :show-button="false"
        placeholder="终值"
        class="input-field"
        @update:value="handleChange"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { NInputNumber } from 'naive-ui'

/**
 * 范围输入类型
 */
export interface RangeInput {
  start: number
  step: number
  end: number
}

const props = withDefaults(defineProps<{
  modelValue: RangeInput
}>(), {
  modelValue: () => ({ start: 0, step: 1, end: 0 })
})

const emit = defineEmits<{
  'update:modelValue': [value: RangeInput]
}>()

// 本地值
const localValue = ref<RangeInput>({ ...props.modelValue })

// 监听 modelValue 变化
watch(() => props.modelValue, (val) => {
  localValue.value = { ...val }
}, { deep: true })

// 处理值变化
function handleChange() {
  emit('update:modelValue', { ...localValue.value })
}
</script>

<style scoped>
.range-input {
  display: flex;
  gap: 8px;
  align-items: center;
}

.input-row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
}

.input-label {
  font-size: 13px;
  color: var(--n-text-color-2);
  white-space: nowrap;
  min-width: 40px;
}

.input-field {
  flex: 1;
  min-width: 0;
}
</style>
