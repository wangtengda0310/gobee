<script setup lang="ts">
import {nowCaseData, updateNtf} from "../composables/use-case-data";
import {SortableEvent, VueDraggable} from "vue-draggable-plus";
import {
  cardSelectFilter,
  chsAndNumSelectFilter,
  excelCardsSelectDynUniqueOption,
  excelHeroesSelectOption,
  excelHeroInitSkillSelectOption,
  excelSkillSelectOption
} from "../composables/HeroAndCardsAndSkillsSelect";
import {computed, nextTick, ref} from "vue";
import {CustomHero} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/function-test";
import {canUseIdentityOption, excelIdentityList} from "../config/Identity";
import {excelSkillMap} from "../config/Skill";
import {useDialog} from "naive-ui";
import {excelHeroMap} from "@shared/config/hero";
import {randomUUID} from "../composables/use-case-data";

const onStart = (e: SortableEvent) => {
  drag.value = true
  console.log(e)
}

const onEnd = (e: SortableEvent) => {
  nextTick(() => {
    drag.value = false
  })
  console.log(e)
  updateNtf()
}

const drag = ref(false)

const addBot = () => {
  const identity = 1
  let color = 1

  const option = canUseIdentityOption(identity, color).filter(o => !o.disabled);

  (nowCaseData.value?.initYanWu?.customHeroes as (CustomHero & { uuid: string })[]).push({
    ...new CustomHero({
      heroId: 10105,
      identity: 1,
      color: option[0].value
    }),
    uuid: crypto.randomUUID()
  })
}

const delBot = (index: number) => {
  dialog.warning({
    title: '删除武将',
    content: '你要删除这个武将吗?',
    positiveText: '确定',
    negativeText: '取消',
    draggable: true,
    onPositiveClick: () => {
      if (nowCaseData.value?.initYanWu?.customHeroes && nowCaseData.value?.initYanWu?.customHeroes.length > 1)
        nowCaseData.value?.initYanWu?.customHeroes.splice(index, 1)
    },
    onNegativeClick: () => {
    }
  })
}

type DragCustomHeros = (CustomHero & { uuid: string })[]

const dialog = useDialog()

const orderMod = ref({
  cards: false,
  initCards: nowCaseData.value?.children?.map(c => false) || [],
  initEquips: nowCaseData.value?.children?.map(c => false) || [],
  exEquips: nowCaseData.value?.children?.map(c => false) || [],
  augurCards: nowCaseData.value?.children?.map(c => false) || [],
  delSkills: nowCaseData.value?.children?.map(c => false) || [],
  addSkills: true,
  cardSkill: false,
  skillCards: nowCaseData.value?.children?.map(c => false) || [],
})
</script>

<template>
  <div v-if="nowCaseData">
    <n-card :title="nowCaseData.label" header-style="font-size: 18px; padding: 10px">
      <template #header-extra>
        <div style="font-size:18px;">
          <n-input v-model:value="nowCaseData.caseDesc" style="min-width: 800px; font-size: 16px"
                   @update:value="()=>{updateNtf()}" placeholder="请输入用例描述"/>
        </div>
      </template>
      <div v-if="nowCaseData.label && nowCaseData.initYanWu">
        <div style="width: 300px;font-size:14px;margin: 10px 0; display: flex; align-items: center; justify-content: center">
          <div style="flex: 0 0 80px">负责人:</div>
          <n-input v-model:value="nowCaseData.caseManager" style="flex: 1 1 0; font-size: 14px"
                   @update:value="()=>{updateNtf()}" placeholder="负责人"/>
        </div>
        <n-card style="margin-bottom: 10px">
          <div style="display: flex; flex-direction: column; gap: 10px; justify-content: space-between">
            <VueDraggable v-if="orderMod.cards && nowCaseData.initYanWu.cards.length > 0"
                          v-model="nowCaseData.initYanWu.cards"
                          :animation="150">
              <n-button
                  v-for="(item, index) in nowCaseData.initYanWu.cards"
                  :key="item">
                {{ `[${index + 1}] ` + item }}
              </n-button>
            </VueDraggable>
            <div style="display: flex; align-items: center">
              <span style="flex: 0 0 100px; font-size: 16px">
                牌堆组
              </span>
              <n-input-number style="flex: 1 1 auto" v-model:value="nowCaseData.initYanWu.cardPile"
                              :min="0" :max="999999" :default-value="0"
                              @update:value="()=>{updateNtf()}"/>
            </div>
            <div style="display: flex; align-items: center">
              <span style="flex: 0 0 100px; font-size: 16px">
                摸牌堆
              </span>
              <n-select style="flex: 1 1 auto" v-model:value="nowCaseData.initYanWu.cards"
                        :options="excelCardsSelectDynUniqueOption"
                        multiple filterable :filter="cardSelectFilter" placeholder="抽牌堆"
                        @update:value="()=>{updateNtf()}"/>
              <n-switch style="flex: 0 0 100px; margin-left: 10px" v-model:value="orderMod.cards" :round="false">
                <template #checked>
                  顺序调整
                </template>
                <template #unchecked>
                  固定顺序
                </template>
              </n-switch>
            </div>
            <div style="display: flex; align-items: center">
              <span style="flex: 0 0 100px; font-size: 16px">
                弃牌堆
              </span>
              <n-select style="flex: 1 1 auto" :options="excelCardsSelectDynUniqueOption"
                        multiple filterable :filter="cardSelectFilter" placeholder="弃牌堆" disabled
                        @update:value="()=>{updateNtf()}"/>
            </div>
          </div>
        </n-card>
        <VueDraggable
            v-if="nowCaseData.initYanWu.customHeroes && nowCaseData.initYanWu.customHeroes.length > 0"
            v-model="nowCaseData.initYanWu.customHeroes"
            @start="onStart"
            @end="onEnd"
            :animation="150"
            :scroll="true"
            :scroll-sensitivity="300"
            :scroll-speed="20"
            target=".sort-target"
            handle=".custom-drag-handle"
        >
          <TransitionGroup
              type="transition"
              tag="div"
              :name="!drag ? 'fade' : undefined"
              class="sort-target"
          >
            <n-card v-for="(bot, index) in nowCaseData.initYanWu.customHeroes as DragCustomHeros" :key="bot.uuid"
                    :title="'座位 ' + (index + 1)" header-style="font-size: 16px; padding: 10px"
                    style="margin-bottom: 10px" closable @close="()=>{delBot(index); updateNtf()}"
            >
              <template #header-extra>
                <div class="custom-drag-handle" style="cursor: pointer">
                  <n-button type="info" dashed>
                    拖动
                  </n-button>
                </div>
              </template>
              <div style="display: flex; flex-direction: column; gap: 10px">
                <!--人物、技能-->
                <!--人物、技能-->
                <div style="display: flex; gap: 10px; flex-wrap: wrap; align-items: center">
                  <n-select style="width: 150px" v-model:value="bot.heroId" :options="excelHeroesSelectOption"
                            filterable :filter="chsAndNumSelectFilter" @update:value="()=>{updateNtf()}"/>
                  <n-select style="width: 150px" v-model:value="bot.identity" :options="excelIdentityList.map(([k,v])=>{
                    return {label: v + '(' + k + ')', value: k }}
                  )" @update:value="(v)=>{
                    updateNtf()
                    if (!bot) return
                    const option = canUseIdentityOption(bot.identity, bot.color).filter(o=>!o.disabled)
                    if (option.length > 0) {
                      bot.color = option[0].value
                    } else {
                      bot.color = -1
                    }
                  }"/>
                  <n-select style="width: 150px" v-model:value="bot.color"
                            :options="canUseIdentityOption(bot.identity, bot.color)"
                            @update:value="()=>{updateNtf()}"/>
                  <span>初始技能: {{
                      excelHeroMap[bot.heroId]?.Skills.map(id => excelSkillMap[id]?.SkillName + '(' + id + ')').join(",")
                    }}</span>
                </div>
                <!--手牌区-->
                <VueDraggable v-if="orderMod.initCards[index] && bot.initCards.length > 0"
                              v-model="bot.initCards" :animation="150">
                  <n-button
                      v-for="(item, index) in bot.initCards"
                      :key="item">
                    {{ `[${index + 1}] ` + item }}
                  </n-button>
                </VueDraggable>
                <div style="display: flex; align-items: center">
                  <span style="flex: 0 0 100px">初始手牌</span>
                  <n-select style="flex: 1 1 auto" v-model:value="bot.initCards"
                            :options="excelCardsSelectDynUniqueOption"
                            multiple filterable :filter="cardSelectFilter" placeholder="手牌"
                            @update:value="()=>{updateNtf()}"/>
                  <n-switch style="flex: 0 0 100px; margin-left: 10px" v-model:value="orderMod.initCards[index]"
                            :round="false">
                    <template #checked>
                      顺序调整
                    </template>
                    <template #unchecked>
                      固定顺序
                    </template>
                  </n-switch>
                </div>
                <!--初始装备(不触发装备效果)-->
                <VueDraggable v-if="orderMod.initEquips[index] && bot.initEquips.length > 0"
                              v-model="bot.initEquips" :animation="150">
                  <n-button
                      v-for="(item, index) in bot.initEquips"
                      :key="item">
                    {{ `[${index + 1}] ` + item }}
                  </n-button>
                </VueDraggable>
                <div style="display: flex; align-items: center">
                  <span style="flex: 0 0 100px">初始装备</span>
                  <n-select style="flex: 1 1 auto" v-model:value="bot.initEquips"
                            :options="excelCardsSelectDynUniqueOption"
                            multiple filterable :filter="cardSelectFilter" placeholder="装备牌"
                            @update:value="()=>{updateNtf()}"/>
                  <n-switch style="flex: 0 0 100px; margin-left: 10px" v-model:value="orderMod.initEquips[index]"
                            :round="false">
                    <template #checked>
                      顺序调整
                    </template>
                    <template #unchecked>
                      固定顺序
                    </template>
                  </n-switch>
                </div>
                <!--穿戴牌区(触发装备效果)-->
                <VueDraggable v-if="orderMod.exEquips[index] && bot.exEquips.length > 0"
                              v-model="bot.exEquips" :animation="150">
                  <n-button
                      v-for="(item, index) in bot.exEquips"
                      :key="item">
                    {{ `[${index + 1}] ` + item }}
                  </n-button>
                </VueDraggable>
                <div style="display: flex; align-items: center">
                  <span style="flex: 0 0 100px">触发装备</span>
                  <n-select style="flex: 1 1 auto" v-model:value="bot.exEquips"
                            :options="excelCardsSelectDynUniqueOption"
                            multiple filterable :filter="cardSelectFilter" placeholder="初始触发装备的装备效果"
                            @update:value="()=>{updateNtf()}"/>
                  <n-switch style="flex: 0 0 100px; margin-left: 10px" v-model:value="orderMod.exEquips[index]"
                            :round="false">
                    <template #checked>
                      顺序调整
                    </template>
                    <template #unchecked>
                      固定顺序
                    </template>
                  </n-switch>
                </div>
                <!--卜卦区-->
                <VueDraggable v-if="orderMod.augurCards[index] && bot.augurCards.length > 0"
                              v-model="bot.augurCards" :animation="150">
                  <n-button
                      v-for="(item, index) in bot.augurCards"
                      :key="item">
                    {{ `[${index + 1}] ` + item }}
                  </n-button>
                </VueDraggable>
                <div style="display: flex; align-items: center">
                  <span style="flex: 0 0 100px">初始卜卦</span>
                  <n-select style="flex: 1 1 auto" v-model:value="bot.augurCards"
                            :options="excelCardsSelectDynUniqueOption"
                            multiple filterable :filter="cardSelectFilter" placeholder="卜卦牌"
                            @update:value="()=>{updateNtf()}"/>
                  <n-switch style="flex: 0 0 100px; margin-left: 10px" v-model:value="orderMod.augurCards[index]"
                            :round="false">
                    <template #checked>
                      顺序调整
                    </template>
                    <template #unchecked>
                      固定顺序
                    </template>
                  </n-switch>
                </div>
                <!--删除技能区-->
                <VueDraggable v-if="orderMod.delSkills[index] && bot.delSkills.length > 0"
                              v-model="bot.delSkills" :animation="150">
                  <n-button
                      v-for="(item, index) in bot.delSkills"
                      :key="item">
                    {{ `[${index + 1}] ` + item }}
                  </n-button>
                </VueDraggable>
                <div style="display: flex; align-items: center">
                  <span style="flex: 0 0 100px">删除技能</span>
                  <n-select style="flex: 1 1 auto" v-model:value="bot.delSkills"
                            :options="excelHeroInitSkillSelectOption(bot.heroId)"
                            multiple filterable :filter="chsAndNumSelectFilter" placeholder="技能"
                            @update:value="()=>{updateNtf()}"/>
                  <n-switch style="flex: 0 0 100px; margin-left: 10px" v-model:value="orderMod.delSkills[index]"
                            :round="false">
                    <template #checked>
                      顺序调整
                    </template>
                    <template #unchecked>
                      固定顺序
                    </template>
                  </n-switch>
                </div>
                <!--增加技能区-->
                <div style="display: flex; align-items: center;">
                  <span style="flex: 0 0 100px">增加技能</span>
                  <n-select style="flex: 0 0 200px;" :options="excelSkillSelectOption"
                            filterable :filter="chsAndNumSelectFilter" placeholder="技能" clearable
                            @update:value="(value)=>{if (!bot.addSkills) {bot.addSkills = []} if (Number.isFinite(value)) {bot.addSkills.push(value)} updateNtf()}"/>
                  <VueDraggable style="flex: 1 1 auto"
                                v-if="orderMod.addSkills"
                                v-model="bot.addSkills" :animation="150">
                    <n-tag style="cursor: pointer"
                           v-for="(item, index) in bot.addSkills"
                           :key="randomUUID()" closable @close="()=>{bot.addSkills.splice(index, 1); updateNtf()}">
                      {{
                        `[${index + 1}] ` + (excelSkillMap[Number(item)]?.SkillName ? excelSkillMap[Number(item)]?.SkillName : item) + `(${item})`
                      }}
                    </n-tag>
                  </VueDraggable>
                </div>
                <div style="display: flex; align-items: center">
                  <span style="flex: 0 0 100px">技能牌区</span>
                  <n-select style="flex: 0 0 200px;" :options="excelSkillSelectOption"
                            filterable :filter="chsAndNumSelectFilter" placeholder="技能" clearable
                            @update:value="(value)=>{if (!bot.skillCardsMap) {bot.skillCardsMap = {}} if (Number.isFinite(value)) {bot.skillCardsMap[value] = []} updateNtf()}"/>
                  <div style="flex: 1 1 auto">
                    <n-tag style="cursor: pointer"
                           v-for="(item, index) in computed(()=>Object.keys(bot.skillCardsMap)).value"
                           :key="randomUUID()" closable
                           @close="()=>{delete bot.skillCardsMap[Number(item)]; updateNtf()}">
                      {{
                        `[${index + 1}] ` + (excelSkillMap[Number(item)]?.SkillName ? excelSkillMap[Number(item)]?.SkillName : item) + `(${item})`
                      }}
                    </n-tag>
                  </div>
                </div>
                <div style="display: flex; align-items: center"
                     v-for="(item, index) in computed(()=>Object.keys(bot.skillCardsMap)).value">
                  <span style="flex: 0 0 100px">{{
                      (excelSkillMap[Number(item)]?.SkillName ? excelSkillMap[Number(item)]?.SkillName : item) + '牌区'
                    }}</span>
                  <n-select style="flex: 1 1 auto" v-model:value="bot.skillCardsMap[Number(item)]"
                            :options="excelCardsSelectDynUniqueOption"
                            multiple filterable :filter="cardSelectFilter" placeholder="技能牌区卡牌"
                            @update:value="()=>{updateNtf()}"/>
                </div>
              </div>
            </n-card>
            <div v-if="nowCaseData.initYanWu.customHeroes.length < 8">
              <n-button strong secondary type="primary" @click="()=>{addBot(); updateNtf()}">
                增加武将
              </n-button>
            </div>
          </TransitionGroup>
        </VueDraggable>
      </div>
    </n-card>
  </div>
</template>

<style scoped>
/*:name自动生成的class*/
.fade-move,
.fade-enter-active,
.fade-leave-active {
  transition: all 0.5s cubic-bezier(0.55, 0, 0.1, 1);
}

/*:name自动生成的class*/
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: scaleY(0.01) translate(30px, 0);
}

/*:name自动生成的class*/
.fade-leave-active {
  position: absolute;
}
</style>