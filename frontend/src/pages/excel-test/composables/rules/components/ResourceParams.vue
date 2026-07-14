<!--
  ResourceParams - 资源路径校验规则参数组件

  包含标准三参数行（允许空值、允许注释、空N行截断）
  加上可选的"父文件夹前缀"输入框
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
    <n-input-group>
      <n-input-group-label size="small">前缀</n-input-group-label>
      <n-input
        :value="params?.prefix ?? ''"
        placeholder="父文件夹前缀（可选）"
        size="small"
        style="flex: 1 1 0"
        @update:value="(v: string) => { if (params) params.prefix = v }"
      />
    </n-input-group>
  </div>
</template>
