<!--
  AllBaseRow - 全基础行

  包含：允许注释 + 空N行截断 + 不能为空 + 唯一不重复 + 仅中文 + 从1开始自增
  allBaseParams 规则专用
-->
<script setup lang="ts">
import {NCheckbox, NInputNumber} from "naive-ui"
import {ERuleParam} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"

const props = defineProps<{
  params: Record<string, string>
}>()

// 初始化默认值
const params = props.params || {}
if (!(ERuleParam.ALLOW_COMMIT in params)) params[ERuleParam.ALLOW_COMMIT] = 'true'
if (!(ERuleParam.BREAK_LINE in params)) params[ERuleParam.BREAK_LINE] = '3'
if (!(ERuleParam.ALLOW_EMPTY in params)) params[ERuleParam.ALLOW_EMPTY] = 'true'
if (!('unique' in params)) params['unique'] = 'false'
if (!('chsOnly' in params)) params['chsOnly'] = 'false'
if (!('increase' in params)) params['increase'] = 'false'
</script>

<template>
  <div style="display: flex; gap: 10px; align-items: center">
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
    <n-checkbox
      style="flex: 0 0 90px"
      :checked="params[ERuleParam.ALLOW_EMPTY] !== 'true'"
      @update:checked="params[ERuleParam.ALLOW_EMPTY] = (!($event as boolean)).toString()"
    >
      不能为空
    </n-checkbox>
    <n-checkbox
      style="flex: 0 0 110px"
      :checked="params['unique'] === 'true'"
      @update:checked="params['unique'] = ($event as boolean).toString()"
    >
      唯一不重复
    </n-checkbox>
    <n-checkbox
      style="flex: 0 0 80px"
      :checked="params['chsOnly'] === 'true'"
      @update:checked="params['chsOnly'] = ($event as boolean).toString()"
    >
      仅中文
    </n-checkbox>
    <n-checkbox
      style="flex: 0 0 120px"
      :checked="params['increase'] === 'true'"
      @update:checked="params['increase'] = ($event as boolean).toString()"
    >
      从1开始自增
    </n-checkbox>
  </div>
</template>
