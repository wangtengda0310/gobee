<!-- SkillTrigger 类型资产段 — 技能/目标/包含任意/传参/期望 -->
<script setup lang="ts">
import {excelSkillSelectOption} from "../../composables/HeroAndCardsAndSkillsSelect"
import {nowCaseData} from "../../composables/use-case-data"
import {excelHeroMap} from "@shared/config/hero"

const props = defineProps<{
  assetList: { [key: string]: any }
}>()

const emit = defineEmits<{
  (e: 'update'): void
}>()
</script>

<template>
  <div class="skillTriggerTypeAsset">
    <span>技能:</span>
    <n-select v-model:value="props.assetList.ActionValue" filterable clearable
              :options="excelSkillSelectOption"
              @update:value="emit('update')" placeholder="技能"/>
    <span>技能目标:</span>
    <n-select v-model:value="props.assetList.DestSeatIds" filterable multiple clearable
              :options="nowCaseData && nowCaseData.initYanWu && nowCaseData.initYanWu.customHeroes ? nowCaseData.initYanWu.customHeroes.map((h,i)=>{return {label: excelHeroMap[h.heroId]?.Name + `(${i+1})`, value: i+1}}) : []"
              @update:value="emit('update')" placeholder="技能目标"/>
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
    <span>传参:</span>
    <n-input-number v-model:value="props.assetList.Param" clearable min="0" max="999999"
                    @update:value="emit('update')" placeholder="参数"/>
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
