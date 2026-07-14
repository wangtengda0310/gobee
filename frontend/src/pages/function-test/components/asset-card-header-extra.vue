<!-- 断言卡片 header-extra：与动作卡片一致的拖动 + 智能描述 + 描述输入 -->
<script setup lang="ts">
import {computed} from "vue"
import {nowCaseData} from "../composables/use-case-data"
import {useAssetAiDesc} from "../composables/use-asset-ai-desc"

const props = defineProps<{
  stepIndex: number
  assetIndex: number
}>()

const {cachedDesc, applyAiDesc, updateDesc} = useAssetAiDesc(props.stepIndex, props.assetIndex)

const desc = computed({
  get: () => nowCaseData.value?.caseSteps?.[props.stepIndex]?.assets?.[props.assetIndex]?.desc ?? "",
  set: (value: string) => updateDesc(value),
})
</script>

<template>
  <div class="cardHeaderExtra" data-testid="asset-card-header-extra">
    <n-button class="custom-drag-handle-asset" type="info" dashed>
      拖动
    </n-button>
    <n-tooltip trigger="hover" :style="{ maxWidth: '700px' }">
      <template #trigger>
        <n-button secondary type="success" class="aiDesc" @click="applyAiDesc">
          应用智能描述->
        </n-button>
      </template>
      {{ cachedDesc }}
    </n-tooltip>
    <n-input v-model:value="desc" :placeholder="cachedDesc"/>
  </div>
</template>

<style scoped>
.cardHeaderExtra {
  display: flex;
  width: 100%;
  gap: 10px;
}

.cardHeaderExtra > :nth-child(1) {
  flex: 0 0 auto;
}

.cardHeaderExtra > :nth-child(2) {
  flex: 0 0 auto;
}

.cardHeaderExtra > :nth-child(3) {
  flex: 1 0 500px;
}
</style>
