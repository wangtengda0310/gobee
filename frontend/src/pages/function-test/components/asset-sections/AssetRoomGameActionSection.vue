<!-- RoomGameActionAsset 类型资产段 — 动作类型/值/传参/包含任意/期望切换 -->
<script setup lang="ts">
import {VueDraggable} from "vue-draggable-plus"
import {randomUUID} from "../../composables/use-case-data"
import {optActionTypeOptions} from "../../composables/AssetProtoOptions"

const props = defineProps<{
  assetList: { [key: string]: any }
}>()

const emit = defineEmits<{
  (e: 'update'): void
}>()
</script>

<template>
  <div class="roomGameActionTypeAsset">
    <span>动作类型:</span>
    <n-select v-model:value="props.assetList.ActionType" :options="optActionTypeOptions"
              @update:value="emit('update')"
              clearable filterable
              placeholder="ActionType"/>
    <span>值:</span>
    <n-input-number v-model:value="props.assetList.ActionValue" min="-999999999" max="999999999"
                    @update:value="emit('update')" placeholder="ActionValue"
                    clearable/>
    <span>传参:</span>
    <VueDraggable v-model="props.assetList.Params" :animation="150">
      <n-tag style="cursor: pointer"
             v-for="(item, index) in props.assetList.Params"
             :key="randomUUID()" closable
             @close="()=>{props.assetList.Params.splice(index, 1); emit('update')}">
        {{ item }}
      </n-tag>
    </VueDraggable>
    <n-select tag filterable placeholder="生成参数" clearable
              @update:value="(value)=>{if (!props.assetList.Params) {props.assetList.Params = []} if (Number.isFinite(Number(value))) {props.assetList.Params.push(Number(value))} emit('update')}"/>
    <n-switch v-model:value="props.assetList.ParamsRandom"
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
