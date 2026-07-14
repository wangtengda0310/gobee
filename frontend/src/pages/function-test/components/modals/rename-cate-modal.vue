<script setup lang="ts">
import {modifyCateNameModalData, showModifyCateNameModal} from "../../composables/Modals";

const bodyStyle = {
  width: '600px'
}
const segmented = {
  content: 'soft',
  footer: 'soft'
} as const

const modCaseName = () => {
  if (modifyCateNameModalData.value.newName && modifyCateNameModalData.value.cateRef) {
    modifyCateNameModalData.value.cateRef.label = modifyCateNameModalData.value.newName
    modifyCateNameModalData.value.cateRef.modified = true
    modifyCateNameModalData.value.newName = ""
    showModifyCateNameModal.value = false
  } else {
    console.log("没找到对应case或新名称为空")
  }
}
</script>

<template>
  <n-modal
      v-model:show="showModifyCateNameModal"
      transform-origin="center"
      preset="card"
      :style="bodyStyle"
      title="修改分类名"
      size="huge"
      :bordered="false"
      :segmented="segmented"
      :show-mask="false"
  >
    <template #header-extra>

    </template>
    <div style="display: flex; flex-direction: column; gap: 10px">
      <div style="display: flex; align-items: center">
        <span style="width: 80px">分类名:</span>
        <n-input v-model:value="modifyCateNameModalData.newName" placeholder="输入一个分类名称"/>
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