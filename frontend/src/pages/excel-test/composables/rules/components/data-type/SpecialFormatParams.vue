<!--
  SpecialFormatParams - 特殊格式检查规则参数组件

  标准三参数行 + 是否为数组复选框
-->
<script setup lang="ts">
import {NCheckbox} from "naive-ui"
import {ERuleParam} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
import StandardBaseRow from '../../base/StandardBaseRow.vue'

defineProps<{ params: { [p: string]: string } }>()
</script>

<template>
  <div style="flex: 1 1 0; display: flex; flex-direction: column; justify-content: space-between; gap: 10px">
    <div style="display: flex; gap: 10px; align-items: center">
      <StandardBaseRow :params="params" :extra-defaults="{'isArray': 'false'}" />
      <n-checkbox
        style="flex: 0 0 100px"
        :checked="params['isArray'] === 'true'"
        @update:checked="params['isArray'] = ($event as boolean).toString()"
      >
        是否为数组
      </n-checkbox>
    </div>
  </div>
</template>
