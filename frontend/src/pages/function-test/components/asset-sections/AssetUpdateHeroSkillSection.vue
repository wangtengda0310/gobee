<!-- UpdateHeroSkill 类型资产段 — UUID管理 + CardParamMap + 技能状态 -->
<script setup lang="ts">
import {VueDraggable} from "vue-draggable-plus"
import {randomUUID} from "../../composables/use-case-data"
import {createTagOnlyNumber} from "../../composables/HeroAndCardsAndSkillsSelect"

const props = defineProps<{
  assetList: { [key: string]: any }
}>()

const emit = defineEmits<{
  (e: 'update'): void
}>()
</script>

<template>
  <div class="updateHeroSkillTypeAsset">
    <div>
      <span>唯一ID:</span>
      <n-select v-model:value="props.assetList.SkillUUid"
                @create="createTagOnlyNumber($event)"
                tag filterable multiple clearable
                @update:value="(newUUIDs)=>{
                  const oldValue = [...(props.assetList.SkillUUid || [])];
                  props.assetList.SkillUUid = Array.isArray(newUUIDs) ? newUUIDs : [];

                  // 确保 CardParamMap 存在
                  if (!props.assetList.CardParamMap) {
                    props.assetList.CardParamMap = {};
                  }

                  // 备份旧数据
                  const oldMap = props.assetList.CardParamMap;

                  // 创建新映射
                  const newMap = {};

                  // 只添加当前存在的 UUID 的数据
                  props.assetList.SkillUUid.forEach(uuid => {
                    // 如果旧数据中有这个 UUID，保留它；否则初始化新的
                    newMap[uuid] = oldMap[uuid] || {
                      CardParamList: [],
                      IsInvalid: false
                    };
                  });

                  // 更新
                  props.assetList.CardParamMap = newMap;

                  emit('update');
                }"
                placeholder="UUID"/>
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
    <div v-for="(val, uuid) in props.assetList.CardParamMap as {[key:number]:{CardParamList:number[], IsInvalid:boolean}}">
      <span>UUID({{ uuid }})传参:</span>
      <n-select @create="createTagOnlyNumber($event)"
                tag filterable clearable placeholder="UUID"
                @update:value="(value)=>{if (Number.isFinite(value)) val.CardParamList.push(value); emit('update')}"/>
      <VueDraggable v-model="val.CardParamList" :animation="150">
        <n-tag style="cursor: pointer"
               v-for="(item, index) in val.CardParamList"
               :key="randomUUID()" closable @close="()=>{val.CardParamList.splice(index, 1); emit('update')}">
          {{
            `[${index + 1}] ` + `${item}`
          }}
        </n-tag>
      </VueDraggable>
      <span>技能状态:</span>
      <n-switch v-model:value="val.IsInvalid"
                @update:value="emit('update')"
                :round="false">
        <template #checked>
          不可用
        </template>
        <template #unchecked>
          可用
        </template>
      </n-switch>
    </div>
  </div>
</template>
