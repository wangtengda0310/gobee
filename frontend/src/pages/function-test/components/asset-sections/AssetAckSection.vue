<!-- Ack 类型资产段 — 结果成功/错误切换 + 错误码多选 -->
<script setup lang="ts">
import {errCodeOptions} from "../../composables/AssetProtoOptions"
import {chsAndNumSelectFilter} from "../../composables/HeroAndCardsAndSkillsSelect"

const props = defineProps<{
  assetList: { [key: string]: any }
  ackBool: boolean
}>()

const emit = defineEmits<{
  (e: 'update'): void
  (e: 'update:ackBool', value: boolean): void
}>()
</script>

<template>
  <div class="ackTypeAsset">
    <span>Result: </span>
    <!-- ackBool 通过 emit 双向绑定，因为 n-switch 的 v-model:value 需要可写引用 -->
    <n-switch :value="props.ackBool" @update:value="(v: boolean)=>{emit('update:ackBool', v); emit('update')}" :round="false">
      <template #checked>
        成功
      </template>
      <template #unchecked>
        错误
      </template>
    </n-switch>
    <n-select v-if="!props.ackBool" v-model:value="props.assetList['Result']" :options="errCodeOptions"
              @update:value="emit('update')" filterable multiple
              :filter="chsAndNumSelectFilter" placeholder="选择错误码"/>
  </div>
</template>
