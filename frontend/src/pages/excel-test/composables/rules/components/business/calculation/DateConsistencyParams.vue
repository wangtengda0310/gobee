<!--
  DateConsistencyParams - 日期一致性检查规则参数组件
-->
<script setup lang="ts">
import {NInput} from "naive-ui"
import StandardBaseRow from '../../../base/StandardBaseRow.vue'

const props = defineProps<{ params: { [p: string]: string } }>()
</script>

<template>
  <div style="flex: 1 1 0; display: flex; flex-direction: column; justify-content: space-between; gap: 10px">
    <StandardBaseRow :params="params" :extra-defaults="{dateCol: '', descCol: ''}" />
    <div style="display: flex; gap: 10px; align-items: center">
      <div>日期列:</div>
      <n-input
        style="flex: 1 1 0"
        placeholder="日期配置列名"
        :value="params['dateCol']"
        @update:value="params['dateCol'] = $event"
      />
      <div>描述列:</div>
      <n-input
        style="flex: 1 1 0"
        placeholder="日期描述列名"
        :value="params['descCol']"
        @update:value="params['descCol'] = $event"
      />
    </div>
  </div>
</template>
