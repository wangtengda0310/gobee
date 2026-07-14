<!--
  ExcelCheckPanel - Excel 检查配置面板（编排层）

  用于配置每个 Sheet 的列级规则和表级规则
  按视觉区域拆分为三个子组件：TableRuleCard、FieldCardList、FieldAnchorNav
-->
<script setup lang="ts">
import {
  nowCheckData,
  nowSheetData,
  nowTableRules
} from "../composables/use-excel-check-data";
import {ruleComponents, ruleOptions} from "../composables/excel-rules-template";
import TableRuleCard from "./table-rule-card.vue";
import FieldCardList from "./field-card-list.vue";
import FieldAnchorNav from "./field-anchor-nav.vue";
</script>

<template>
  <div style="height: 100%" v-if="nowSheetData">
    <div v-if="nowSheetData.sheetHeader?.Col && nowSheetData.sheetHeader?.Col.length > 0"
         style="display: flex; height: 100%">
      <n-scrollbar style="max-height: 100%; flex: 1 1 0;" id="CardsTop">
        <div
            style="display: flex; flex-direction: column; justify-content: space-between; gap: 10px; margin: 10px 10px">

          <!-- 表级校验规则 -->
          <TableRuleCard />

          <!-- 字段卡片列表 -->
          <FieldCardList
              :sheet-data="nowSheetData"
              :check-data="nowCheckData"
              :rule-options-list="ruleOptions"
              :rule-component-map="ruleComponents"
          />

        </div>
      </n-scrollbar>
      <n-scrollbar style="flex: 0 0 auto; max-height: 100%; max-width: 200px;">
        <FieldAnchorNav
            :columns="nowSheetData.sheetHeader?.Col"
            :check-data="nowCheckData"
            :table-rule-count="nowTableRules.length"
        />
      </n-scrollbar>
    </div>
  </div>
</template>
