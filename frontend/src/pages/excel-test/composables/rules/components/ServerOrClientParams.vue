<!--
  ServerOrClientParams - 服务端/客户端导出目标规则参数组件
-->
<script setup lang="ts">
import {NRadioGroup, NRadio} from "naive-ui"
import StandardBaseRow from '../base/StandardBaseRow.vue'

defineProps<{ params: { [p: string]: string } }>()
</script>

<template>
  <div style="flex: 1 1 0; display: flex; flex-direction: column; justify-content: space-between; gap: 10px">
    <StandardBaseRow :params="params" :allow-empty="'false'" :extra-defaults="{'target': 'server'}" />
    <div style="display: flex; gap: 10px; align-items: center">
      <div>导出目标:</div>
      <n-radio-group
        :value="params['target']"
        @update:value="params['target'] = $event"
      >
        <n-radio value="server">服务端</n-radio>
        <n-radio value="client">客户端</n-radio>
        <n-radio value="all">全部</n-radio>
      </n-radio-group>
    </div>
  </div>
</template>
