<!-- EquipChange 类型资产段 — 新增装备/移除数量/移除装备/包含任意/期望 -->
<script setup lang="ts">
import {
  cardSelectFilter,
  createTagOnlyNumber,
  excelCardsSelectFallbackOption,
  excelCardsSelectOptionFromInit
} from "../../composables/HeroAndCardsAndSkillsSelect"
import {excelCardsMap} from "../../config/Card"

const props = defineProps<{
  assetList: { [key: string]: any }
}>()

const emit = defineEmits<{
  (e: 'update'): void
}>()
</script>

<template>
  <div class="equipChangeTypeAsset">
    <span>新增装备:</span>
    <n-select v-model:value="props.assetList.AddEquip"
              :options="excelCardsSelectOptionFromInit"
              :fallback-option="excelCardsSelectFallbackOption"
              :filter="cardSelectFilter"
              @create="createTagOnlyNumber($event, excelCardsMap, 'cards')"
              tag filterable clearable
              @update:value="emit('update')" placeholder="卡"/>
    <span>移除数量:</span>
    <n-input-number v-model:value="props.assetList.Count" min="0" max="999999"
                    @update:value="emit('update')"
                    placeholder="卡牌数"/>
    <span>移除装备:</span>
    <n-select v-model:value="props.assetList.RemoveEquip"
              :options="excelCardsSelectOptionFromInit"
              :fallback-option="excelCardsSelectFallbackOption"
              :filter="cardSelectFilter"
              @create="createTagOnlyNumber($event, excelCardsMap, 'cards')"
              tag filterable multiple clearable
              @update:value="emit('update')" placeholder="卡"/>
    <n-switch v-model:value="props.assetList.Random"
              @update:value="emit('update')"
              :round="false">
      <template #checked>
        包含任意
      </template>
      <template #unchecked>
        严格匹配
      </template>
    </n-switch>
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
