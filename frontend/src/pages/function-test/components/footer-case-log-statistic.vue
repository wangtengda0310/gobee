<script setup lang="ts">
import {
  nowRunningCase,
  nowRunningCaseAssertionErrorTotal,
  nowRunningCaseBatchProgress,
  nowRunningCaseCompletedCount,
  nowRunningCaseTotalCount,
} from "../composables/RobotTestLog";
</script>

<template>
  <div>
    <div style="display: flex; flex-direction: row; width: 100%; justify-content: center; align-items: center">
      <n-progress
          v-if="nowRunningCase.length > 0"
          data-testid="batch-case-progress"
          type="line"
          :percentage="nowRunningCaseBatchProgress"
          indicator-placement="inside"
          :processing="nowRunningCaseCompletedCount < nowRunningCaseTotalCount"
          style="min-width: 300px; margin-right: 10px"
      >
        <template #default>
          {{ nowRunningCaseCompletedCount }}/{{ nowRunningCaseTotalCount }}
        </template>
      </n-progress>
      <div style="min-width: 250px">当前运行用例:
        {{ nowRunningCase.length == 0 ? '未运行' : nowRunningCase.length == 1 ? nowRunningCase[0].label : '并行' }}
      </div>
      <div style="min-width: 200px">断言错误数目: {{ nowRunningCaseAssertionErrorTotal }}</div>
    </div>
  </div>
</template>

<style scoped>

</style>