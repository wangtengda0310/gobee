<script setup lang="ts">
import {addCateModalData, showAddCateModal} from "../../composables/Modals";
import {ref} from "vue";
import {JsonsDir} from "../../composables/Option";
import {dataRef} from "../../composables/use-case-data";

const bodyStyle = {
  width: '600px'
}
const segmented = {
  content: 'soft',
  footer: 'soft'
} as const

const cateName = ref("")

const confirm = () => {
  if (cateName.value) {
    const savePath = JsonsDir.value + '/' + cateName.value + '.json'
    if (addCateModalData.value)
      dataRef.splice(addCateModalData.value.index + 1, 0, {
        label: cateName.value,
        key: crypto.randomUUID(),
        isLeaf: true,
        fullPath: savePath,
        levelType: "Categories",
        modified: true,
        children: []
      })
    showAddCateModal.value = false
  }
}
</script>

<template>
  <n-modal
      v-model:show="showAddCateModal"
      transform-origin="center"
      class="custom-card"
      preset="card"
      :style="bodyStyle"
      title="添加分类"
      size="huge"
      :bordered="false"
      :segmented="segmented"
      :show-mask="false"
  >
    <template #header-extra>
    </template>
    <div>
      {{ addCateModalData.text }}
    </div>
    <div style="display: flex; align-items: center">
      <span style="width: 100px">分类名: </span>
      <n-input v-model:value="cateName" placeholder="分类名"/>
    </div>
    <template #footer>
      <div style="display: flex; align-items: center; justify-content: right">
        <n-button type="success" @click="confirm">
          添加分类
        </n-button>
      </div>
    </template>
  </n-modal>
</template>

<style scoped>

</style>