<!--
  RegexParams - 正则表达式检查规则参数组件
-->
<script setup lang="ts">
import {NInput} from "naive-ui"
import {ERuleParam} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
import StandardBaseRow from '../../../base/StandardBaseRow.vue'

const props = defineProps<{ params: { [p: string]: string } }>()
</script>

<template>
  <div style="flex: 1 1 0; display: flex; flex-direction: column; justify-content: space-between; gap: 10px">
    <StandardBaseRow :params="params" :extra-defaults="{[ERuleParam.PATTERN]: ''}" />
    <div style="display: flex; gap: 10px; align-items: center">
      <div>正则表达式:</div>
      <n-input
        style="flex: 1 1 0"
        placeholder="正则表达式"
        :value="params[ERuleParam.PATTERN]"
        @update:value="params[ERuleParam.PATTERN] = $event"
      />
    </div>
  </div>
</template>
