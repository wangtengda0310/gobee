<!--
  PinYinChsParams - 拼音中文检查规则参数组件
-->
<script setup lang="ts">
import {NInput} from "naive-ui"
import StandardBaseRow from '../../../base/StandardBaseRow.vue'

const props = defineProps<{ params: { [p: string]: string } }>()
</script>

<template>
  <div style="flex: 1 1 0; display: flex; flex-direction: column; justify-content: space-between; gap: 10px">
    <StandardBaseRow :params="params" :extra-defaults="{chsCol: ''}" />
    <div style="display: flex; gap: 10px; align-items: center">
      <div>中文列:</div>
      <n-input
        style="flex: 1 1 0"
        placeholder="中文列名"
        :value="params['chsCol']"
        @update:value="params['chsCol'] = $event"
      />
    </div>
  </div>
</template>
