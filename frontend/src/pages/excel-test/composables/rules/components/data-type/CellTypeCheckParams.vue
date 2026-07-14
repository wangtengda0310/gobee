<!--
  CellTypeCheckParams - 单元格类型检查规则参数组件

  检测字符串列中是否存在数值格式单元格
-->
<script setup lang="ts">
import StandardBaseRow from '../../base/StandardBaseRow.vue'

defineProps<{ params: { [p: string]: string } }>()
</script>

<template>
  <div style="flex: 1 1 0; display: flex; flex-direction: column; justify-content: space-between; gap: 10px">
    <StandardBaseRow :params="params" />
    <div style="color: #999; font-size: 12px; margin-top: 5px">
      <p>说明: 检测字符串列中是否存在数值格式单元格（如 "abc" 列出现 123）</p>
      <p>文本格式对应 CellTypeSharedString(7) / CellTypeInlineString(5)，数值格式对应 CellTypeUnset(0) / CellTypeNumber(6)</p>
    </div>
  </div>
</template>
