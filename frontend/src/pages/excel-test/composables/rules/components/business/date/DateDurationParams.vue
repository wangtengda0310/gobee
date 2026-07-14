<!--
  DateDurationParams - 日期跨度检查规则参数组件
-->
<script setup lang="ts">
import {NInput} from "naive-ui"
import StandardBaseRow from '../../../base/StandardBaseRow.vue'

const props = defineProps<{ params: { [p: string]: string } }>()
</script>

<template>
  <div style="flex: 1 1 0; display: flex; flex-direction: column; justify-content: space-between; gap: 10px">
    <StandardBaseRow :params="params" :extra-defaults="{startDateCol: '', endDateCol: ''}" />
    <div style="display: flex; gap: 10px; align-items: center">
      <div>开始日期列:</div>
      <n-input
        style="flex: 1 1 0"
        placeholder="列名"
        :value="params['startDateCol']"
        @update:value="params['startDateCol'] = $event"
      />
      <div>结束日期列:</div>
      <n-input
        style="flex: 1 1 0"
        placeholder="列名"
        :value="params['endDateCol']"
        @update:value="params['endDateCol'] = $event"
      />
    </div>
  </div>
</template>
