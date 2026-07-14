<!--
  WeightSumParams - 权重总和检查规则参数组件
-->
<script setup lang="ts">
import {NInput} from "naive-ui"
import StandardBaseRow from '../../../base/StandardBaseRow.vue'

const props = defineProps<{ params: { [p: string]: string } }>()
</script>

<template>
  <div style="flex: 1 1 0; display: flex; flex-direction: column; justify-content: space-between; gap: 10px">
    <StandardBaseRow :params="params" :extra-defaults="{targetSum: '100', weightCols: ''}" />
    <div style="display: flex; gap: 10px; align-items: center">
      <div>目标总和:</div>
      <n-input
        style="flex: 1 1 0"
        placeholder="数值"
        :value="params['targetSum']"
        @update:value="params['targetSum'] = $event"
      />
      <div>权重列:</div>
      <n-input
        style="flex: 1 1 0"
        placeholder="列名,逗号分隔"
        :value="params['weightCols']"
        @update:value="params['weightCols'] = $event"
      />
    </div>
  </div>
</template>
