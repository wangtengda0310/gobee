<!-- ReplaceCard 类型资产段 — 替换卡牌/生成卡牌两组 UI -->
<script setup lang="ts">
import {
  cardSelectFilter,
  createTagOnlyNumber,
  excelCardsSelectFallbackOption,
  excelCardsSelectOption,
  excelSkillSelectOption
} from "../../composables/HeroAndCardsAndSkillsSelect"
import {excelCardsMap} from "../../config/Card"

const props = defineProps<{
  assetList: { [key: string]: any }
}>()

const emit = defineEmits<{
  (e: 'update'): void
}>()

/**
 * 替换卡牌和生成卡牌的List更新时，给两个Map初始化[]
 */
const updateChangeCardToList = (v: number[], assetList: { [key: string]: any }, prefix: string) => {
  const cardKeys = [`${prefix}Cards`, `${prefix}Points`, `${prefix}AttrTypes`];

  cardKeys.forEach(key => {
    // 初始化对象如果不存在
    if (assetList[key] == null) {
      assetList[key] = {};
    }

    // 获取当前所有的cardId
    const currentCardIds = Object.keys(assetList[key]).map(Number);

    // 移除不在新列表v中的card
    currentCardIds.forEach(cardId => {
      if (!v.includes(cardId)) {
        delete assetList[key][cardId];
      }
    });

    // 初始化或保留在新列表v中的card
    v.forEach(c => {
      if (assetList[key][c] == null) {
        assetList[key][c] = [];
      }
    });
  });
}
</script>

<template>
  <div class="replaceCardTypeAsset">
    <!--替换卡牌 UI组-->
    <div>
      <div class="replaceCardRow">
        <span>替换卡牌(唯一ID):</span>
        <n-select v-model:value="props.assetList.ReplaceCards_ChangeCards"
                  :options="excelCardsSelectOption"
                  :fallback-option="excelCardsSelectFallbackOption"
                  :filter="cardSelectFilter"
                  @create="createTagOnlyNumber($event, excelCardsMap, 'cards')"
                  tag filterable multiple clearable
                  @update:value="(v)=>{updateChangeCardToList(v, props.assetList, 'ReplaceCards_ChangeCardsTo'); emit('update');}"
                  placeholder="替换卡牌"/>
        <n-switch v-model:value="props.assetList.ReplaceCards_ChangeCardsRandom"
                  @update:value="emit('update')"
                  :round="false">
          <template #checked>
            包含任意
          </template>
          <template #unchecked>
            严格匹配
          </template>
        </n-switch>
      </div>
      <div class="replaceCardToRows">
        <div v-for="fromCard in props.assetList.ReplaceCards_ChangeCards" class="replaceCardToRow">
          <span>替换卡牌:</span>
          <span>[{{ excelCardsMap[fromCard]?.Name }} {{ fromCard }}]</span>
          <span>替换为(配表ID):</span>
          <n-select v-model:value="props.assetList.ReplaceCards_ChangeCardsToCards[fromCard]"
                    :options="excelSkillSelectOption" filterable clearable multiple
                    @update:value="emit('update')" placeholder="替换卡牌"/>
          <span>点数:</span>
          <n-select v-model:value="props.assetList.ReplaceCards_ChangeCardsToPoints[fromCard]"
                    :options="[1,2,3,4,5,6,7,8].map(n=>{return {label: n, value: n}})"
                    filterable multiple clearable
                    @update:value="emit('update')" placeholder="卡牌点数"/>
          <span>花色:</span>
          <n-select v-model:value="props.assetList.ReplaceCards_ChangeCardsToAttrTypes[fromCard]"
                    :options="['none','Jin','Mu','Shui','Huo','Tu',].map((s, i)=>{return{label: s, value: i}})"
                    filterable multiple clearable
                    @update:value="emit('update')" placeholder="卡牌属性"/>
        </div>
      </div>
    </div>
    <!--生成卡牌 UI组-->
    <div>
      <div class="replaceCardRow">
        <span style="min-width: 70px">生成卡牌(唯一ID):</span>
        <n-select style="flex: 1 1 0" v-model:value="props.assetList.ReplaceCards_DrawCards"
                  :options="excelCardsSelectOption"
                  :fallback-option="excelCardsSelectFallbackOption"
                  @create="createTagOnlyNumber($event, excelCardsMap, 'cards')"
                  tag filterable multiple clearable
                  @update:value="(v)=>{updateChangeCardToList(v, props.assetList, 'ReplaceCards_DrawCardsTo'); emit('update');}"
                  placeholder="替换卡牌"/>
        <n-switch style="width: 100px" v-model:value="props.assetList.ReplaceCards_DrawCardsRandom"
                  @update:value="emit('update')"
                  :round="false">
          <template #checked>
            包含任意
          </template>
          <template #unchecked>
            严格匹配
          </template>
        </n-switch>
      </div>
      <div class="replaceCardToRows">
        <div v-for="fromCard in props.assetList.ReplaceCards_DrawCards">
          <span>替换卡牌:</span>
          <span>[{{ excelCardsMap[fromCard]?.Name }} {{ fromCard }}]</span>
          <span>生成为(配表ID):</span>
          <n-select v-model:value="props.assetList.ReplaceCards_DrawCardsToCards[fromCard]"
                    :options="excelSkillSelectOption" filterable clearable multiple
                    @update:value="emit('update')" placeholder="替换卡牌"/>
          <span>点数:</span>
          <n-select v-model:value="props.assetList.ReplaceCards_DrawCardsToPoints[fromCard]"
                    :options="[1,2,3,4,5,6,7,8].map(n=>{return {label: n, value: n}})"
                    filterable multiple clearable
                    @update:value="emit('update')" placeholder="卡牌点数"/>
          <span>花色:</span>
          <n-select v-model:value="props.assetList.ReplaceCards_DrawCardsToAttrTypes[fromCard]"
                    :options="['none','Jin','Mu','Shui','Huo','Tu',].map((s, i)=>{return{label: s, value: i}})"
                    filterable multiple clearable
                    @update:value="emit('update')" placeholder="卡牌属性"/>
        </div>
      </div>
    </div>
  </div>
</template>
