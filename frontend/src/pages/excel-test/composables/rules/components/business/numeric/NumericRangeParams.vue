<!--
  NumericRangeParams - 数值范围检查规则参数组件
-->
<script setup lang="ts">
import {NInput} from "naive-ui"
import StandardBaseRow from '../../../base/StandardBaseRow.vue'

const props = defineProps<{ params: { [p: string]: string } }>()
</script>

<template>
  <div style="flex: 1 1 0; display: flex; flex-direction: column; justify-content: space-between; gap: 10px">
    <StandardBaseRow :params="params" :extra-defaults="{minValue: '', maxValue: ''}" />
    <div style="display: flex; gap: 10px; align-items: center">
      <div>最小值:</div>
      <n-input
        style="flex: 1 1 0"
        placeholder="数值"
        :value="params['minValue']"
        @update:value="params['minValue'] = $event"
      />
      <div>最大值:</div>
      <n-input
        style="flex: 1 1 0"
        placeholder="数值"
        :value="params['maxValue']"
        @update:value="params['maxValue'] = $event"
      />
    </div>
  </div>
</template>
