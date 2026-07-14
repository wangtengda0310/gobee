<!--
  DateParams - 日期类型检查规则参数组件
-->
<script setup lang="ts">
import {NInput} from "naive-ui"
import StandardBaseRow from '../../base/StandardBaseRow.vue'

defineProps<{ params: { [p: string]: string } }>()
</script>

<template>
  <div style="flex: 1 1 0; display: flex; flex-direction: column; justify-content: space-between; gap: 10px">
    <StandardBaseRow :params="params" :extra-defaults="{'format': '2006-01-02'}" />
    <div style="display: flex; gap: 10px; align-items: center">
      <div>日期格式:</div>
      <n-input
        style="flex: 1 1 0"
        placeholder="格式，如2006-01-02"
        :value="params['format']"
        @update:value="params['format'] = $event"
      />
    </div>
  </div>
</template>
