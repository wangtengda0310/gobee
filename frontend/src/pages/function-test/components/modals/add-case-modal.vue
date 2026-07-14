<script setup lang="ts">
import {addCaseModalData, showAddCaseModal} from "../../composables/Modals";
import {newCaseData} from "../../composables/use-case-data";
import {dataRef} from "../../composables/use-case-data";

const bodyStyle = {
  width: '600px'
}
const segmented = {
  content: 'soft',
  footer: 'soft'
} as const

const addCase = () => {
  const cate = dataRef.find(d => d.label == addCaseModalData.value.cate);
  if (cate && cate.label && addCaseModalData.value.case) {
    cate.children?.splice(addCaseModalData.value.index + 1, 0, {
      ...newCaseData(cate.label,
          addCaseModalData.value.case,
          cate.fullPath),
      modified: true,
    })
    if (cate.children && cate.children.length > 0) {
      cate.isLeaf = false
    }
    showAddCaseModal.value = false
  } else {
    console.log(addCaseModalData.value)
  }
}
</script>

<template>
  <n-modal
      v-model:show="showAddCaseModal"
      transform-origin="center"
      preset="card"
      :style="bodyStyle"
      title="添加用例"
      size="huge"
      :bordered="false"
      :segmented="segmented"
      :show-mask="false"
  >
    <template #header-extra>

    </template>
    <div style="display: flex; flex-direction: column; gap: 10px">
      <div style="display: flex; align-items: center">
        <span style="width: 80px">用例名:</span>
        <n-input v-model:value="addCaseModalData.case" placeholder="输入一个Case名称"/>
      </div>
    </div>
    <template #footer>
      <div style="display: flex; align-items: center; justify-content: right">
        <n-button type="success" @click="addCase">
          添加用例
        </n-button>
      </div>
    </template>
  </n-modal>
</template>

<style scoped>

</style>