<!-- 一般卡牌类资产段 — DrawCard/DisCard/PlayCard/GiveCard/CardEnhance/XianBaTouChou -->
<script setup lang="ts">
import {Asset} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/function-test"
import {AssetEnum} from "../../composables/StepActionsAndAssetsSelect"
import {optActionTypeOptions} from "../../composables/AssetProtoOptions"
import {
  cardSelectFilter,
  createTagOnlyNumber,
  excelCardsSelectFallbackOption,
  excelCardsSelectOptionFromInit,
  excelSkillSelectOption
} from "../../composables/HeroAndCardsAndSkillsSelect"
import {excelCardsMap} from "../../config/Card"

const props = defineProps<{
  asset: Asset
  assetList: { [key: string]: any }
}>()

const emit = defineEmits<{
  (e: 'update'): void
}>()
</script>

<template>
  <div class="normalCardTypeAsset">
    <!-- DrawCard 额外行动类型/值 -->
    <div v-if="props.asset.msgName == AssetEnum.DrawCard">
      <span>行动类型:</span>
      <n-select v-model:value="props.assetList.ActionType" :options="optActionTypeOptions"
                @update:value="emit('update')"
                clearable filterable
                placeholder="行动类型"/>
      <span>行动值:</span>
      <n-select v-model:value="props.assetList.ActionValue" :options="props.assetList.ActionType == 1 ? [{
                    label: '摸牌',
                    value: 1
                  },{
                    label: '弃牌',
                    value: 2
                  },{
                    label: '出牌',
                    value: 3
                  },] : props.assetList.ActionType == 2 || props.assetList.ActionType == 3 ? excelSkillSelectOption : []"
                @update:value="emit('update')"
                clearable filterable tag
                @create="createTagOnlyNumber"
                placeholder="行动值"/>
    </div>
    <div>
      <span>卡牌数:</span>
      <n-input-number v-model:value="props.assetList.Count" min="0" max="999999"
                      @update:value="emit('update')"
                      placeholder="卡牌数"/>
      <span>卡:</span>
      <n-select v-model:value="props.assetList.Cards"
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
      <span>不期望卡:</span>
      <n-select v-model:value="props.assetList.UnexpectCards"
                :options="excelCardsSelectOptionFromInit"
                :fallback-option="excelCardsSelectFallbackOption"
                :filter="cardSelectFilter"
                @create="createTagOnlyNumber($event, excelCardsMap, 'cards')"
                tag filterable multiple clearable
                @update:value="emit('update')" placeholder="不期望卡"/>
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
  </div>
</template>
