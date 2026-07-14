<!--
  FieldAnchorNav - 右侧锚点导航
  展示表级规则和每个字段的导航链接，点击跳转到对应卡片
-->
<script setup lang="ts">
defineProps<{
  /** 字段列数据 */
  columns: any[]
  /** 列规则数据（用于显示规则数量） */
  checkData: { [key: string]: any }
  /** 表级规则数量 */
  tableRuleCount: number
}>()
</script>

<template>
  <n-anchor :show-rail="false" :show-background="true" :bound="114" style="margin: 10px 0">
    <!-- 表级规则导航 -->
    <n-anchor-link href="#tableRules">
      <template #title>
        <div style="display: flex; align-items: center">
          <div style="white-space: pre-line; line-height: 1.4; display: flex">
            <div style="flex: 0 0 25px; color: #FF6B6B">
              0.
            </div>
            <div style="flex: 1 1 0; color: #FF6B6B">
              表级规则
            </div>
          </div>
          <n-tag style="margin-left: 5px" type="warning" size="small">
            {{ tableRuleCount }}
          </n-tag>
        </div>
      </template>
    </n-anchor-link>
    <!-- 列级规则导航 -->
    <n-anchor-link v-for="(col, index) in columns"
                   :href="'#'+(col?.AttrName ? col.AttrName : '空'+index)">
      <template #title>
        <div style="display: flex; align-items: center">
          <div style="white-space: pre-line; line-height: 1.4; display: flex">
            <div style="flex: 0 0 25px">
              {{ (index + 1) + '. ' }}
            </div>
            <div style="flex: 1 1 0">
              {{ (col?.AttrCHS ? col.AttrCHS : (col?.AttrName ? col.AttrName : '空')) }}
            </div>
          </div>
          <n-tag style="margin-left: 5px" type="warning">
            {{
              col && checkData[col.AttrName] && checkData[col.AttrName]?.PropRules ? checkData[col.AttrName]?.PropRules.length : 0
            }}
          </n-tag>
        </div>
      </template>
    </n-anchor-link>
  </n-anchor>
</template>