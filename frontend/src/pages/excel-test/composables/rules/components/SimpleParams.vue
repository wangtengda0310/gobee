<!--
  SimpleParams - 通用简单参数组件

  仅包含标准三参数行（允许空值、允许注释、空N行截断），
  用于没有额外参数行的规则类型
-->
<script setup lang="ts">
import StandardBaseRow from '../base/StandardBaseRow.vue'

defineProps<{
  params: { [p: string]: string }
  defaults?: { allowEmpty?: string }
}>()
</script>

<template>
  <div style="flex: 1 1 0; display: flex; flex-direction: column; justify-content: space-between; gap: 10px">
    <StandardBaseRow :params="params" :allow-empty="defaults?.allowEmpty ?? 'true'" />
  </div>
</template>
