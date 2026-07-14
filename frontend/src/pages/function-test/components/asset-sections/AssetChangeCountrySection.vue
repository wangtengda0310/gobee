<!-- ChangeCountry 类型资产段 — 主要势力/额外势力 -->
<script setup lang="ts">
import {countryOptions} from "../../composables/AssetProtoOptions"
import {createTagOnlyNumber} from "../../composables/HeroAndCardsAndSkillsSelect"
import {countryMap} from "../../config/ECountry"

const props = defineProps<{
  assetList: { [key: string]: any }
}>()

const emit = defineEmits<{
  (e: 'update'): void
}>()
</script>

<template>
  <div class="changeCountryTypeAsset">
    <span>主要势力:</span>
    <n-select v-model:value="props.assetList.MainCountry"
              @create="createTagOnlyNumber($event, countryMap, 'normal')"
              tag filterable multiple clearable
              :options="countryOptions"
              @update:value="emit('update')" placeholder="势力ID"/>
    <span>额外势力:</span>
    <n-select v-model:value="props.assetList.ExtraCountry"
              @create="createTagOnlyNumber($event, countryMap, 'normal')"
              tag filterable multiple clearable
              :options="countryOptions"
              @update:value="emit('update')" placeholder="势力ID"/>
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
