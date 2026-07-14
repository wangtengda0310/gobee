<!-- UpdateProperty 类型资产段 — 目标/类型/值/期望切换 -->
<script setup lang="ts">
import {
  propertyValueOptions,
  updatePropertyOptions
} from "../../composables/AssetProtoOptions"
import {engAndSelectFilterSpecial} from "../../composables/StepActionsAndAssetsSelect"

const props = defineProps<{
  assetList: { [key: string]: any }
}>()

const emit = defineEmits<{
  (e: 'update'): void
}>()
</script>

<template>
  <div class="updatePropertyTypeAsset">
    <span>目标:</span>
    <n-select v-model:value="props.assetList.IdType" :options="updatePropertyOptions" filterable
              :filter="engAndSelectFilterSpecial" @update:value="emit('update')"
              clearable
              placeholder="目标"/>
    <span>类型:</span>
    <n-select v-model:value="props.assetList.PropID" :options="propertyValueOptions" filterable
              :filter="engAndSelectFilterSpecial" @update:value="emit('update')"
              clearable
              placeholder="类型"/>
    <span>值:</span>
    <n-input-number v-model:value="props.assetList.PropValue"
                    @update:value="emit('update')"
                    placeholder="值"/>
    <n-switch v-model:value="props.assetList.UnExpect"
              @update:value="emit('update')"
              :round="false">
      <template #checked>
        不期望
      </template>
      <template #unchecked>
        期望
      </template>
    </n-switch>
  </div>
</template>
