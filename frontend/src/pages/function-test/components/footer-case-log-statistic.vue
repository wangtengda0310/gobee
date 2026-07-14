<script setup lang="ts">
import {computed} from "vue";
import {nowRunningCase, nowRunningCaseError, nowRunningCaseProcess} from "../composables/RobotTestLog";

// 进度计算
const process = computed(() => {
  return nowRunningCaseProcess.value.length > 0 ? 100 * nowRunningCaseProcess.value.reduce((acc, cur) => {
    acc += cur
    return acc
  }) / nowRunningCaseProcess.value.length : 0
})
</script>

<template>
  <div>
    <div style="display: flex; flex-direction: row; width: 100%; justify-content: center; align-items: center">
      <n-progress
          v-if="nowRunningCase.length > 0"
          type="line"
          :percentage="process"
          indicator-placement="inside"
          :processing="process != 100"
          style="min-width: 300px; margin-right: 10px"
      />
      <div style="min-width: 250px">当前运行用例:
        {{ nowRunningCase.length == 0 ? '未运行' : nowRunningCase.length == 1 ? nowRunningCase[0].label : '并行' }}
      </div>
      <div style="min-width: 200px">断言错误数目: {{
          nowRunningCaseError.reduce((acc, cur) => {
            acc += cur;
            return acc
          }, 0)
        }}
      </div>
    </div>
  </div>
</template>

<style scoped>

</style>