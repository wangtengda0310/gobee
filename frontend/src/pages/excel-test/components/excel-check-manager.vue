<!--
  ExcelCheckManager - Excel 检查负责人管理

  用于配置每个 Sheet 的负责人信息（QA、策划、程序）
-->
<script setup lang="ts">
import {excelCheckManagerListMap, nowSheetData} from "../composables/use-excel-check-data";
import {GiantLabel} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule";

const QAData: { label: string, value: GiantLabel }[] = [
  {
    label: '唐方达',
    value: {
      Name: '唐方达'
    }
  }
]
const DesignerData = []
const ProgrammerData = []
</script>

<template>
  <div id="ExcelCheckManager">
    <div class="row">
      <div class="row_title">负责QA:</div>
      <n-select v-if="excelCheckManagerListMap.get(nowSheetData?.label ?? '')?.QA"
                style="flex: 1 1 0"
                v-model:value="excelCheckManagerListMap.get(nowSheetData?.label ?? '')!.QA"
                multiple :options="QAData"
                placeholder="请选择负责该表的QA"
      />
    </div>
    <div class="row">
      <div class="row_title">负责策划:</div>
      <n-select v-if="excelCheckManagerListMap.get(nowSheetData?.label ?? '')?.Designer"
                v-model:value="excelCheckManagerListMap.get(nowSheetData?.label ?? '')!.Designer"
                multiple :options="DesignerData"
                placeholder="请选择负责该表的策划"
      />
    </div>
    <div class="row">
      <div class="row_title">负责程序:</div>
      <n-select v-if="excelCheckManagerListMap.get(nowSheetData?.label ?? '')?.Programmer"
                v-model:value="excelCheckManagerListMap.get(nowSheetData?.label ?? '')!.Programmer"
                multiple :options="ProgrammerData"
                placeholder="请选择负责该表的程序"
      />
    </div>
  </div>
</template>

<style scoped>
#ExcelCheckManager {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  gap: 10px;
  margin: 10px;
}

.row {
  display: flex;
  align-items: center
}

.row_title {
  flex: 0 0 80px
}
</style>
