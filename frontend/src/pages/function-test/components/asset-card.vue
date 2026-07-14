<!-- 资产卡片容器组件 — 根据 msgName 类型渲染对应的子组件 -->
<script setup lang="ts">
import {computed, Ref, ref} from "vue"
import {nowCaseData, updateNtf} from "../composables/use-case-data"
import {
  AssetEnum,
  assetSelectOption,
  chsAndEngSelectFilter
} from "../composables/StepActionsAndAssetsSelect"
import {assetList2AssetMap, assetMap2AssetList, normalCardAssetValueType} from "../composables/AssetMapTrans"
import {Asset} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/function-test"

// 子组件
import AssetAckSection from "./asset-sections/AssetAckSection.vue"
import AssetUpdatePropertySection from "./asset-sections/AssetUpdatePropertySection.vue"
import AssetUpdateHeroSkillSection from "./asset-sections/AssetUpdateHeroSkillSection.vue"
import AssetRoomGameActionSection from "./asset-sections/AssetRoomGameActionSection.vue"
import AssetNormalCardSection from "./asset-sections/AssetNormalCardSection.vue"
import AssetReplaceCardSection from "./asset-sections/AssetReplaceCardSection.vue"
import AssetCommonHpChangeSection from "./asset-sections/AssetCommonHpChangeSection.vue"
import AssetAttrChangeSection from "./asset-sections/AssetAttrChangeSection.vue"
import AssetChangeCountrySection from "./asset-sections/AssetChangeCountrySection.vue"
import AssetSkillTriggerSection from "./asset-sections/AssetSkillTriggerSection.vue"
import AssetEquipChangeSection from "./asset-sections/AssetEquipChangeSection.vue"

const props = defineProps<{
  stepIndex: number
  assetIndex: number
}>()

const step = computed(() => {
  return nowCaseData.value?.caseSteps?.[props.stepIndex]
})

const asset = computed(() => {
  return step.value?.assets?.[props.assetIndex]
})

// Ack 类型的布尔值（成功/错误）
const ackBool = ref<boolean>((step.value && asset.value) ? (!!(asset.value.attr['Result'] && asset.value.attr['Result'] == '0')) : false)

// 初始化不定长map, 暂定List
let assetList = ref<{ [key: string]: any }>({})

const init = () => {
  if (asset.value && assetList.value)
    assetMap2AssetList(asset as Ref<Asset>, assetList)
}
// 仅初始化一次
init()

// 历史原因，兼容旧case格式
const updateAssetListToAssetMap = () => {
  if (asset.value)
    assetList2AssetMap(asset as Ref<Asset>, assetList, ackBool)
}

/** 子组件更新后统一调用的处理函数：同步 assetList 到 assetMap 并通知 */
const handleSectionUpdate = () => {
  updateAssetListToAssetMap()
  updateNtf()
}
</script>

<template>
  <div v-if="step && asset" class="assetCard">
    <!-- 断言类型选择 -->
    <div class="assetTypeRow" data-testid="asset-type-row">
      <n-select class="assetTypeSelect"
                v-model:value="asset.msgName"
                :options="assetSelectOption(step)"
                filterable
                :filter="chsAndEngSelectFilter" @update:value="handleSectionUpdate"/>
    </div>

    <!-- Ack 类型 -->
    <AssetAckSection
      v-if="asset.msgName.endsWith('Ack')"
      :assetList="assetList"
      :ackBool="ackBool"
      @update="handleSectionUpdate"
      @update:ackBool="(v: boolean) => { ackBool = v; handleSectionUpdate() }"
    />

    <!-- UpdateProperty 类型 -->
    <AssetUpdatePropertySection
      v-if="asset.msgName == AssetEnum.UpdateProperty"
      :assetList="assetList"
      @update="handleSectionUpdate"
    />

    <!-- UpdateHeroSkill 类型 -->
    <AssetUpdateHeroSkillSection
      v-if="asset.msgName == AssetEnum.UpdateHeroSkill"
      :assetList="assetList"
      @update="handleSectionUpdate"
    />

    <!-- 非 Ack 和 UpdateProperty 的类型组 -->
    <div v-if="!asset.msgName.endsWith('Ack') && asset.msgName != AssetEnum.UpdateProperty"
         class="notAckOrUpdatePropertyTypeAssets">

      <!-- RoomGameActionAsset 类型 -->
      <AssetRoomGameActionSection
        v-if="asset.msgName == AssetEnum.RoomGameActionAsset"
        :assetList="assetList"
        @update="handleSectionUpdate"
      />

      <!-- 一般卡牌类断言 -->
      <AssetNormalCardSection
        v-if="normalCardAssetValueType(asset)"
        :asset="asset"
        :assetList="assetList"
        @update="handleSectionUpdate"
      />

      <!-- 替换卡牌类断言 -->
      <AssetReplaceCardSection
        v-if="asset.msgName == AssetEnum.ReplaceCard"
        :assetList="assetList"
        @update="handleSectionUpdate"
      />

      <!-- 生命变化断言 -->
      <AssetCommonHpChangeSection
        v-if="asset.msgName == AssetEnum.CommonHpChange"
        :assetList="assetList"
        @update="handleSectionUpdate"
      />

      <!-- 属性变化断言 -->
      <AssetAttrChangeSection
        v-if="asset.msgName == AssetEnum.AttrChange"
        :assetList="assetList"
        @update="handleSectionUpdate"
      />

      <!-- 势力改变断言 -->
      <AssetChangeCountrySection
        v-if="asset.msgName == AssetEnum.ChangeCountry"
        :assetList="assetList"
        @update="handleSectionUpdate"
      />

      <!-- 技能触发断言 -->
      <AssetSkillTriggerSection
        v-if="asset.msgName == AssetEnum.SkillTrigger"
        :assetList="assetList"
        @update="handleSectionUpdate"
      />

      <!-- 装备改变断言 -->
      <AssetEquipChangeSection
        v-if="asset.msgName == AssetEnum.EquipChange"
        :assetList="assetList"
        @update="handleSectionUpdate"
      />
    </div>
  </div>
</template>

<style scoped>
/* 长度初始化1344 */
.assetCard {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

/*第一行, 决定断言类型*/
.assetTypeRow {
  display: flex;
  gap: 10px
}

.assetTypeSelect {
  flex: 0 0 200px;
}

/*Ack类型断言*/
.ackTypeAsset {
  display: flex;
  gap: 10px;
  align-items: center
}

.ackTypeAsset > :nth-child(1) {
  flex: 0 0 60px;
}

.ackTypeAsset > :nth-child(2) {
  flex: 0 0 60px;
}

.ackTypeAsset > :nth-child(3) {
  flex: 1 1 900px;
}

/*UpdatePropertyNtf断言*/
.updatePropertyTypeAsset {
  display: flex;
  gap: 10px;
  align-items: center
}

.updatePropertyTypeAsset > :nth-child(1) {
  flex: 0 0 40px;
}

.updatePropertyTypeAsset > :nth-child(2) {
  flex: 0 0 200px;
}

.updatePropertyTypeAsset > :nth-child(3) {
  flex: 0 0 40px;
}

.updatePropertyTypeAsset > :nth-child(4) {
  flex: 0 0 200px;
}

.updatePropertyTypeAsset > :nth-child(5) {
  flex: 0 0 25px;
}

.updatePropertyTypeAsset > :nth-child(6) {
  flex: 1 1 0;
}

.updatePropertyTypeAsset > :nth-child(7) {
  flex: 0 0 80px;
}

/*UpdateHeroSkill断言*/
.updateHeroSkillTypeAsset {
  display: flex;
  gap: 10px;
  flex-direction: column;
}

.updateHeroSkillTypeAsset > :nth-child(1) {
  display: flex;
  gap: 10px;
  align-items: center
}

.updateHeroSkillTypeAsset > :nth-child(1) > :nth-child(1) {
  flex: 0 0 60px;
}

.updateHeroSkillTypeAsset > :nth-child(1) > :nth-child(2) {
  flex: 1 1 0;
}

.updateHeroSkillTypeAsset > :nth-child(1) > :nth-child(3) {
  flex: 0 0 80px;
}

.updateHeroSkillTypeAsset > :nth-child(n+2) {
  display: flex;
  gap: 10px;
  align-items: center
}

.updateHeroSkillTypeAsset > :nth-child(n+2) > :nth-child(1) {
  flex: 0 0 180px;
}

.updateHeroSkillTypeAsset > :nth-child(n+2) > :nth-child(2) {
  flex: 0 0 200px;
}

.updateHeroSkillTypeAsset > :nth-child(n+2) > :nth-child(3) {
  flex: 1 1 0;
}

.updateHeroSkillTypeAsset > :nth-child(n+2) > :nth-child(4) {
  flex: 0 0 80px
}

/*非Ack和UpdatePropertyNtf类型断言*/
.notAckOrUpdatePropertyTypeAssets {
  display: flex;
  flex-direction: column;
  gap: 5px
}

/*各种断言默认样式*/
.notAckOrUpdatePropertyTypeAssets > :nth-child(n) {
  display: flex;
  gap: 10px;
}

/*动作类型断言*/
.roomGameActionTypeAsset {
  align-items: center
}

.roomGameActionTypeAsset > :nth-child(1) {
  flex: 0 0 65px;
}

.roomGameActionTypeAsset > :nth-child(2) {
  flex: 0 0 200px;
}

.roomGameActionTypeAsset > :nth-child(3) {
  flex: 0 0 25px;
}

.roomGameActionTypeAsset > :nth-child(4) {
  flex: 0 0 200px;
}

.roomGameActionTypeAsset > :nth-child(5) {
  flex: 0 0 40px;
}

.roomGameActionTypeAsset > :nth-child(6) {
  flex: 1 1 0;
}

.roomGameActionTypeAsset > :nth-child(7) {
  flex: 0 0 150px;
}

.roomGameActionTypeAsset > :nth-child(8) {
  flex: 0 0 100px;
}

.roomGameActionTypeAsset > :nth-child(9) {
  flex: 0 0 80px;
}

/*一般卡牌类断言*/
.normalCardTypeAsset {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.normalCardTypeAsset > :first-child {
  align-items: center;
  display: flex;
  gap: 10px;
}

.normalCardTypeAsset > :first-child > :nth-child(1) {
  flex: 0 0 65px;
}

.normalCardTypeAsset > :first-child > :nth-child(2) {
  flex: 0 0 200px;
}

.normalCardTypeAsset > :first-child > :nth-child(3) {
  flex: 0 0 55px;
}

.normalCardTypeAsset > :first-child > :nth-child(4) {
  flex: 0 0 200px;
}

.normalCardTypeAsset > :last-child {
  align-items: center;
  display: flex;
  gap: 10px;
}

.normalCardTypeAsset > :last-child > :nth-child(1) {
  flex: 0 0 65px;
}

.normalCardTypeAsset > :last-child > :nth-child(2) {
  flex: 0 0 200px;
}

.normalCardTypeAsset > :last-child > :nth-child(3) {
  flex: 0 0 25px;
}

.normalCardTypeAsset > :last-child > :nth-child(4) {
  flex: 1 0 0;
}

.normalCardTypeAsset > :last-child > :nth-child(5) {
  flex: 0 0 100px;
}

.normalCardTypeAsset > :last-child > :nth-child(6) {
  flex: 0 0 65px;
}

.normalCardTypeAsset > :last-child > :nth-child(7) {
  flex: 1 0 0;
}

.normalCardTypeAsset > :last-child > :nth-child(8) {
  flex: 0 0 80px;
}

/*替换卡牌类断言*/
.replaceCardTypeAsset {
  flex-direction: column;
}

.replaceCardTypeAsset .replaceCardRow {
  display: flex;
  gap: 10px;
  align-items: center;
}

.replaceCardTypeAsset .replaceCardRow > :nth-child(1) {
  flex: 0 0 130px;
}

.replaceCardTypeAsset .replaceCardRow > :nth-child(2) {
  flex: 1 0 0;
}

.replaceCardTypeAsset .replaceCardRow > :nth-child(3) {
  flex: 0 0 100px;
}

.replaceCardTypeAsset .replaceCardToRows {
  display: flex;
  gap: 10px;
  flex-direction: column;
  margin-bottom: 10px;
  margin-top: 10px;
  color: #8effa1
}

.replaceCardTypeAsset .replaceCardToRows > :nth-child(n) {
  display: flex;
  align-items: center;
  gap: 10px
}

.replaceCardTypeAsset .replaceCardToRows > :nth-child(n) > :nth-child(1) {
  flex: 0 0 60px;
}

.replaceCardTypeAsset .replaceCardToRows > :nth-child(n) > :nth-child(2) {
  flex: 0 0 auto;
}

.replaceCardTypeAsset .replaceCardToRows > :nth-child(n) > :nth-child(3) {
  flex: 0 0 100px;
}

.replaceCardTypeAsset .replaceCardToRows > :nth-child(n) > :nth-child(4) {
  flex: 1 0 0;
}

.replaceCardTypeAsset .replaceCardToRows > :nth-child(n) > :nth-child(5) {
  flex: 0 0 35px;
}

.replaceCardTypeAsset .replaceCardToRows > :nth-child(n) > :nth-child(6) {
  flex: 0 0 100px;
}

.replaceCardTypeAsset .replaceCardToRows > :nth-child(n) > :nth-child(7) {
  flex: 0 0 35px;
}

.replaceCardTypeAsset .replaceCardToRows > :nth-child(n) > :nth-child(8) {
  flex: 0 0 100px;
}

/*生命变化断言*/
.commonHpChangeTypeAsset {
  align-items: center;
}

.commonHpChangeTypeAsset > :nth-child(1) {
  flex: 0 0 65px;
}

.commonHpChangeTypeAsset > :nth-child(2) {
  flex: 1 1 0;
}

.commonHpChangeTypeAsset > :nth-child(3) {
  flex: 0 0 65px;
}

.commonHpChangeTypeAsset > :nth-child(4) {
  flex: 1 1 0;
}

.commonHpChangeTypeAsset > :nth-child(5) {
  flex: 0 0 65px;
}

.commonHpChangeTypeAsset > :nth-child(6) {
  flex: 1 1 0;
}

.commonHpChangeTypeAsset > :nth-child(7) {
  flex: 0 0 65px;
}

.commonHpChangeTypeAsset > :nth-child(8) {
  flex: 1 1 0;
}

.commonHpChangeTypeAsset > :nth-child(9) {
  flex: 0 0 80px;
}

/*属性变化断言*/
.attrChangeTypeAsset {
  align-items: center;
}

.attrChangeTypeAsset > :not(:last-child) > :nth-child(1) {
  text-align: left;
}

.attrChangeTypeAsset > :last-child {
  flex: 0 0 80px;
}

/*势力改变断言*/
.changeCountryTypeAsset {
  align-items: center;
}

.changeCountryTypeAsset > :nth-child(1) {
  flex: 0 0 65px;
}

.changeCountryTypeAsset > :nth-child(2) {
  flex: 1 0 0;

}

.changeCountryTypeAsset > :nth-child(3) {
  flex: 0 0 65px;
}

.changeCountryTypeAsset > :nth-child(4) {
  flex: 1 0 0;
}

/*技能触发断言*/
.skillTriggerTypeAsset {
  align-items: center;
}

.skillTriggerTypeAsset > :nth-child(1) {
  flex: 0 0 40px;
}

.skillTriggerTypeAsset > :nth-child(2) {
  flex: 1 0 0;
}

.skillTriggerTypeAsset > :nth-child(3) {
  flex: 0 0 65px;
}

.skillTriggerTypeAsset > :nth-child(4) {
  flex: 1 0 0;
}

.skillTriggerTypeAsset > :nth-child(5) {
  flex: 0 0 100px;
}

.skillTriggerTypeAsset > :nth-child(6) {
  flex: 0 0 40px;
}

.skillTriggerTypeAsset > :nth-child(7) {
  flex: 1 0 0;
}

.skillTriggerTypeAsset > :nth-child(8) {
  flex: 0 0 80px;
}

/*装备改变断言*/
.equipChangeTypeAsset {
  align-items: center;
}

.equipChangeTypeAsset > :nth-child(1) {
  flex: 0 0 65px;
}

.equipChangeTypeAsset > :nth-child(2) {
  flex: 0 0 200px;
}

.equipChangeTypeAsset > :nth-child(3) {
  flex: 0 0 65px;
}

.equipChangeTypeAsset > :nth-child(4) {
  flex: 0 0 120px;
}

.equipChangeTypeAsset > :nth-child(5) {
  flex: 0 0 65px;
}

.equipChangeTypeAsset > :nth-child(6) {
  flex: 1 1 0;
}

.equipChangeTypeAsset > :nth-child(7) {
  flex: 0 0 100px;
}

.equipChangeTypeAsset > :nth-child(8) {
  flex: 0 0 80px;
}
</style>
