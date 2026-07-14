<!--
  AllBaseParams - 基础规则参数组件

  使用全基础行（不能为空+允许注释+截断+唯一+中文+自增）+ 正则格式选择
-->
<script setup lang="ts">
import AllBaseRow from '../base/AllBaseRow.vue'
import RegexFormatSelect from './RegexFormatSelect.vue'

defineProps<{ params: { [p: string]: string } }>()
</script>

<template>
  <div style="flex: 1 1 0; display: flex; flex-direction: column; justify-content: space-between; gap: 10px">
    <AllBaseRow :params="params" />
    <RegexFormatSelect :params="params" />
  </div>
</template>
