<!--
  StandardBaseRow - 标准基础三参数行

  包含：允许空值(Checkbox) + 允许注释(Checkbox,disabled) + 空N行截断(InputNumber)
  自动初始化默认值，无需外部手动调用 initDefaults
-->
<script setup lang="ts">
import {NCheckbox, NInputNumber} from "naive-ui"
import {ERuleParam} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"

const props = withDefaults(defineProps<{
  params: Record<string, string>
  /** 额外参数默认值（如 {format: '2006-01-02'}） */
  extraDefaults?: Record<string, string>
  /** 允许空值默认值（默认 'true'） */
  allowEmpty?: string
}>(), {
  extraDefaults: () => ({}),
  allowEmpty: 'true',
})

// 初始化标准三参数默认值
const params = props.params || {}
if (!(ERuleParam.ALLOW_EMPTY in params)) params[ERuleParam.ALLOW_EMPTY] = props.allowEmpty
if (!(ERuleParam.ALLOW_COMMIT in params)) params[ERuleParam.ALLOW_COMMIT] = 'true'
if (!(ERuleParam.BREAK_LINE in params)) params[ERuleParam.BREAK_LINE] = '3'

// 初始化额外参数默认值
for (const [key, value] of Object.entries(props.extraDefaults)) {
  if (!(key in params)) params[key] = value
}
</script>

<template>
  <div style="display: flex; gap: 10px; align-items: center">
    <n-checkbox
      style="flex: 0 0 100px"
      :checked="params[ERuleParam.ALLOW_EMPTY] === 'true'"
      @update:checked="params[ERuleParam.ALLOW_EMPTY] = ($event as boolean).toString()"
    >
      允许空值
    </n-checkbox>
    <n-checkbox
      style="flex: 0 0 100px"
      disabled
      :checked="params[ERuleParam.ALLOW_COMMIT] === 'true'"
      @update:checked="params[ERuleParam.ALLOW_COMMIT] = ($event as boolean).toString()"
    >
      允许注释
    </n-checkbox>
    <div>空N行截断:</div>
    <n-input-number
      style="flex: 0 0 150px"
      placeholder="行数"
      :min="1"
      :max="5"
      :value="Number(params[ERuleParam.BREAK_LINE]) || 3"
      @update:value="params[ERuleParam.BREAK_LINE] = ($event?.toString() || '3')"
    />
  </div>
</template>
