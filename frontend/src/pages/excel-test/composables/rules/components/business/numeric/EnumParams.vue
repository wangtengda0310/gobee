<!--
  EnumParams - 枚举检查规则参数组件
  支持直接枚举检查 + 正则提取后枚举检查
-->
<script setup lang="ts">
import {NSelect} from "naive-ui"
import {ERuleParam} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
import StandardBaseRow from '../../../base/StandardBaseRow.vue'
import RegexFormatSelect from '../../RegexFormatSelect.vue'

const props = defineProps<{ params: { [p: string]: string } }>()
</script>

<template>
  <div style="flex: 1 1 0; display: flex; flex-direction: column; justify-content: space-between; gap: 10px">
    <StandardBaseRow :params="params" :extra-defaults="{[ERuleParam.ENUMS]: ''}" />
    <RegexFormatSelect :params="params" />
    <!-- 枚举值 -->
    <div style="display: flex; gap: 10px; align-items: center">
      <div>枚举值:</div>
      <n-select
        style="flex: 1 1 0"
        placeholder="枚举值"
        filterable multiple tag
        :value="params[ERuleParam.ENUMS] ? params[ERuleParam.ENUMS].split(',') : []"
        @update:value="params[ERuleParam.ENUMS] = ($event as string[]).join(',')"
      />
    </div>
  </div>
</template>
