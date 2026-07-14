<!-- CommonHpChange 类型资产段 — 伤害来源/体力变化/当前体力/最大体力 -->
<script setup lang="ts">
import {hpSrcOptions} from "../../composables/AssetProtoOptions"

const props = defineProps<{
  assetList: { [key: string]: any }
}>()

const emit = defineEmits<{
  (e: 'update'): void
}>()
</script>

<template>
  <div class="commonHpChangeTypeAsset">
    <span>伤害来源:</span>
    <n-select v-model:value="props.assetList.HpSrc" :options="hpSrcOptions" clearable
              @update:value="emit('update')" placeholder="来源"/>
    <span>体力变化:</span>
    <n-input-number v-model:value="props.assetList.ChangeHp" min="-999999" max="999999"
                    @update:value="emit('update')" placeholder="体力变化"
                    clearable/>
    <span>当前体力:</span>
    <n-input-number v-model:value="props.assetList.CurHp" min="-999999" max="999999"
                    @update:value="emit('update')" placeholder="体力" clearable/>
    <span>最大体力:</span>
    <n-input-number v-model:value="props.assetList.MaxHp" min="-999999" max="999999"
                    @update:value="emit('update')" placeholder="最大体力"
                    clearable/>
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
