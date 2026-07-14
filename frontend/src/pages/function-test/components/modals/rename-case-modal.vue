<script setup lang="ts">
import {modifyCaseNameModalData, showModifyCaseNameModal} from "../../composables/Modals";

const bodyStyle = {
  width: '600px'
}
const segmented = {
  content: 'soft',
  footer: 'soft'
} as const

const modCaseName = () => {
  if (modifyCaseNameModalData.value.newName && modifyCaseNameModalData.value.caseRef) {
    modifyCaseNameModalData.value.caseRef.label = modifyCaseNameModalData.value.newName
    modifyCaseNameModalData.value.caseRef.modified = true
    modifyCaseNameModalData.value.newName = ""
    showModifyCaseNameModal.value = false
  } else {
    console.log("没找到对应case或新名称为空")
  }
}
</script>

<template>
  <n-modal
      v-model:show="showModifyCaseNameModal"
      transform-origin="center"
      preset="card"
      :style="bodyStyle"
      title="修改用例名"
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
        <n-input v-model:value="modifyCaseNameModalData.newName" placeholder="输入一个Case名称"/>
      </div>
    </div>
    <template #footer>
      <div style="display: flex; align-items: center; justify-content: right">
        <n-button type="success" @click="modCaseName">
          修改名称
        </n-button>
      </div>
    </template>
  </n-modal>
</template>

<style scoped>

</style>