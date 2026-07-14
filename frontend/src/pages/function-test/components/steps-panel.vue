<script setup lang="ts">
import {nowCaseData, updateNtf} from "../composables/use-case-data";
import {SortableEvent, VueDraggable} from 'vue-draggable-plus'
import {
  ActionEnum,
  actionsSelectOption,
  AssetEnum,
  isStepsHasAck
} from "../composables/StepActionsAndAssetsSelect";
import {excelHeroMap} from "@shared/config/hero";
import {excelCardsMap} from "../config/Card";
import {computed, nextTick, ref} from "vue";
import {excelSkillMap} from "../config/Skill";
import {Asset, Step} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/function-test";
import AssetCard from "./asset-card.vue";
import {useDialog} from 'naive-ui'
import {DragAttr} from "../composables/use-case-data";
import {
  chsAndNumSelectFilter,
  createTagOnlyNumber,
  excelCardsSelectFallbackOption,
  excelCardsSelectOptionFromInit,
  excelSkillSelectOption,
  excelSkillSelectOptionFromInit
} from "../composables/HeroAndCardsAndSkillsSelect";
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
  console.log(nowCaseData.value)
  updateNtf()
}

const drag = ref(false)

const addAsset = (step: Step) => {
  if (!step.assets)
    step.assets = [];
  (step.assets as (Asset & DragAttr)[]).push({
    ...new Asset({
      msgName: AssetEnum.UpdateProperty,
      desc: "",
      attr: {
        "IdType": "User",
        "PropID": "CircleDoHurtValue",
        "PropValue": "2",
      }
    }),
    uuid: crypto.randomUUID()
  })
}

const delAsset = (asset: Asset[] | undefined, index: number) => {
  if (asset) {
    dialog.warning({
      title: '删除断言',
      content: '你要删除这个断言吗?',
      positiveText: '确定',
      negativeText: '取消',
      draggable: true,
      onPositiveClick: () => {
        asset.splice(index, 1)
      },
      onNegativeClick: () => {
      }
    })
    updateNtf()
  }
}

const delStep = (index: number) => {
  dialog.warning({
    title: '删除行动',
    content: '你要删除这个行动吗?',
    positiveText: '确定',
    negativeText: '取消',
    draggable: true,
    onPositiveClick: () => {
      nowCaseData.value?.caseSteps?.splice(index, 1)
    },
    onNegativeClick: () => {
    }
  })
}
const addStep = (index: number) => {
  const newStep = new Step({
    desc: "",
    action: ActionEnum.Sleep,
    robotIdx: 1,
    sleepTime: 0,
    timeout: 1
  })
  nowCaseData.value?.caseSteps?.splice(index + 1, 0, Object.assign(newStep, {
    uuid: crypto.randomUUID(),
  }))
}

const copyStep = (index: number) => {
  if (nowCaseData.value?.caseSteps) {
    const newStep = JSON.parse(JSON.stringify(nowCaseData.value.caseSteps[index]))
    newStep.uuid = crypto.randomUUID()
    nowCaseData.value?.caseSteps?.splice(index + 1, 0, newStep)
  }
}

function aiDesc(step: Step) {
  if (!nowCaseData.value) return '请输入描述'
  const heroId = (step.robotIdx && nowCaseData.value.initYanWu) ? nowCaseData.value.initYanWu.customHeroes[(step.robotIdx - 1) % nowCaseData.value.initYanWu.customHeroes.length]?.heroId : 0
  const heroName = heroId ? excelHeroMap.value[heroId]?.Name : heroId
  switch (step.action) {
    case ActionEnum.Sleep:
      return heroName + `(${step.robotIdx})`
          + '等待' + `${step.sleepTime}秒`
    case ActionEnum.PlayCard:
      return heroName + `(${step.robotIdx})`
          + ((step.targetsId && step.targetsId.length > 0) ? `对[${step.targetsId?.map(id => (nowCaseData.value?.initYanWu?.customHeroes[id - 1]?.heroId ? excelHeroMap.value[nowCaseData.value.initYanWu?.customHeroes[id - 1]?.heroId]?.Name : nowCaseData.value?.initYanWu?.customHeroes[id - 1]?.heroId) + '(' + id + ')').join(',')}]` : '')
          + '使用牌' + `[${step.cards?.map(id => excelCardsMap.value[id]?.Name ? excelCardsMap.value[id]?.Name + '(' + id + ')' : id).join(',')}]`
          + ((step.transCardSkill && step.transCardSkill > 0) ? `, 当作牌${excelCardsMap.value[step.transCardSkill]?.Name ? excelCardsMap.value[step.transCardSkill]?.Name : step.transCardSkill}打出` : '')
    case ActionEnum.DisCard:
      return heroName + `(${step.robotIdx})`
          + ((step.cards && step.cards.length > 0) ? `弃牌[${step.cards?.map(id => excelCardsMap.value[id]?.Name ? excelCardsMap.value[id]?.Name + '(' + id + ')' : id).join(',')}]` : '自动弃牌')
    case ActionEnum.OptRoomAction:
      const confirm = step.confirm ? ((step.targetsId && step.targetsId.length > 0) ? `对[${step.targetsId?.map(id => (nowCaseData.value?.initYanWu?.customHeroes[id - 1]?.heroId ? excelHeroMap.value[nowCaseData.value.initYanWu?.customHeroes[id - 1]?.heroId]?.Name : nowCaseData.value?.initYanWu?.customHeroes[id - 1]?.heroId) + '(' + id + ')').join(',')}]` : '')
          + ((step.cards && step.cards.length > 0) ? `, 使用牌[${step.cards?.map(id => excelCardsMap.value[id]?.Name ? excelCardsMap.value[id]?.Name + '(' + id + ')' : id).join(',')}]` : '')
          + ((step.transCardSkill && step.transCardSkill > 0) ? `, 当作牌${excelSkillMap.value[step.transCardSkill]?.SkillName ? excelSkillMap.value[step.transCardSkill]?.SkillName : step.transCardSkill}打出` : '') : ''
      return heroName + `(${step.robotIdx})`
          + (step.confirm ? '确认' : '取消')
          + confirm
    case ActionEnum.UseHeroSkill:
      return heroName + `(${step.robotIdx})`
          + ((step.targetsId && step.targetsId.length > 0) ? `对[${step.targetsId?.map(id => (nowCaseData.value?.initYanWu?.customHeroes[id - 1]?.heroId ? excelHeroMap.value[nowCaseData.value.initYanWu?.customHeroes[id - 1]?.heroId]?.Name : nowCaseData.value?.initYanWu?.customHeroes[id - 1]?.heroId) + '(' + id + ')').join(',')}]` : '')
          + ((step.heroSkillUuid && step.heroSkillUuid > 0) ? `, 使用技能[${excelSkillMap.value[step.heroSkillUuid]?.SkillName ? excelSkillMap.value[step.heroSkillUuid]?.SkillName : step.heroSkillUuid}]` : '')
          + ((step.cards && step.cards.length > 0) ? `, 使用牌[${step.cards?.map(id => excelCardsMap.value[id]?.Name ? excelCardsMap.value[id]?.Name + '(' + id + ')' : id).join(',')}]` : '')
          + ((step.transCardSkill && step.transCardSkill > 0) ? `, 当作牌${excelSkillMap.value[step.transCardSkill]?.SkillName ? excelSkillMap.value[step.transCardSkill]?.SkillName : step.transCardSkill}打出` : '')
    case ActionEnum.PlayCardOver:
      return heroName + `(${step.robotIdx})` + '结束打牌'
    case ActionEnum.OnlyAsset:
      return heroName + `(${step.robotIdx})` + '断言'
  }
  return '请输入描述'
}

function applyAiDesc(step: Step) {
  updateNtf()
  step.desc = aiDesc(step)
}

function clearOptOtherVal(v: boolean, step: Step) {
  if (!v) {
    step.heroSkillUuid = undefined
    step.targetsId = undefined
    step.cards = undefined
    step.transCardSkill = undefined
    console.log(v, step)
  } else {
    console.log(v, step)
  }
}

function clearActionOtherVal(step: Step) {
  // 如果这个step没有asset, 那么给他自动增加一个默认asset
  if (!step.assets || step.assets.length == 0) {
    if (isStepsHasAck(step)) {
      step.assets = [{
        uuid: crypto.randomUUID(),
        ...new Asset({
          msgName: step.action + "Ack",
          desc: "",
          attr: {}
        })
      } as any]
    } else if (step.action == ActionEnum.OnlyAsset) {
      step.assets = [{
        uuid: crypto.randomUUID(),
        ...new Asset({
          msgName: AssetEnum.UpdateProperty,
          desc: "",
          attr: {}
        })
      } as any]
    }
  } else {
    if (isStepsHasAck(step)) {
      const find = step.assets.find(a => a.msgName.endsWith("Ack"));
      if (find) {
        find.msgName = step.action + "Ack"
      }
    }
  }

  // TODO 可能还要删除其他属性，比如cards等
  // "id"?: number;
  // "desc"?: string;
  // "robotIdx": number;
  // "action": string;
  // "targetsId"?: number[];
  // "cards"?: number[];
  // "cardsConcat"?: number[];
  // "confirm": boolean;
  // "heroSkillUuid"?: number;
  // "transCardSkill"?: number;
  // "timeout"?: number;
  // "sleepTime"?: number;
  // "assets"?: Asset[];

  step.desc = undefined
  step.targetsId = undefined
  step.cards = undefined
  step.cardsConcat = undefined
  step.confirm = false
  step.heroSkillUuid = undefined
  step.transCardSkill = undefined
  step.timeout = undefined
  step.sleepTime = undefined
  step.assets = undefined

}

const dialog = useDialog()
</script>

<template>
  <div v-if="nowCaseData">
    <VueDraggable
        v-model="nowCaseData.caseSteps as any"
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
        <n-card header-style="font-size: 18px; padding: 10px" v-for="(step, index) in nowCaseData.caseSteps"
                :key="step.uuid" :title="'动作 ' + (index + 1)" :id="'动作 ' + (index + 1)"
                closable @close="()=>{delStep(index); updateNtf()}"
        >
          <template #header-extra>
            <div class="cardHeaderExtra">
              <n-button class="custom-drag-handle" type="info" dashed>
                拖动
              </n-button>
              <n-tooltip trigger="hover">
                <template #trigger>
                  <n-button secondary type="success" class="aiDesc" @click="applyAiDesc(step)">
                    应用智能描述->
                  </n-button>
                </template>
                {{ aiDesc(step) }}
              </n-tooltip>
              <n-input v-model:value="step.desc" :placeholder="aiDesc(step)" @update:value="()=>{updateNtf()}"/>
              <n-button strong secondary type="info" @click="()=>copyStep(index)">
                复制
              </n-button>
            </div>
          </template>
          <div class="stepCardContainer">
            <div>
              <n-select style="width: 105px" v-model:value="step.action" :options="actionsSelectOption"
                        @update:value="()=>{clearActionOtherVal(step); updateNtf();}"/>
              <!--根据Action这里渲染不同的Select来选择不同的玩意儿-->
              <!--武将及座位-->
              <n-select style="width: 120px" v-model:value="step.robotIdx"
                        @create="createTagOnlyNumber($event)"
                        :options="nowCaseData.initYanWu?.customHeroes.map((hero, index)=>{
                          // 显示英雄id和(座位号)
                          return {
                            label: `座位(${index+1})`,
                            value: index + 1
                          }
                        })" @update:value="()=>{updateNtf()}" tag filterable/>
              <!--Sleep-->
              <n-input-number v-if="step.action == ActionEnum.Sleep" style="width: 110px"
                              v-model:value="step.sleepTime" @update:value="()=>{updateNtf()}"
                              round max="600" min="0" placeholder="1">
                <template #suffix>秒</template>
              </n-input-number>
              <!--PlayCard\DisCard\OptRoomAction\UseHeroSkill\PlayCardOver-->
              <!--confirm-->
              <n-switch v-if="step.action == ActionEnum.OptRoomAction" v-model:value="step.confirm" :round="false"
                        @update:value="(v)=>{updateNtf(); clearOptOtherVal(v, step)}">
                <template #checked>确认</template>
                <template #unchecked>取消</template>
              </n-switch>
              <!--heroSkillUuid-->
              <n-select
                  v-if="(step.confirm && step.action == ActionEnum.OptRoomAction) || (step.action == ActionEnum.UseHeroSkill)"
                  style="width: 180px" v-model:value="step.heroSkillUuid"
                  :options="excelSkillSelectOptionFromInit(step)"
                  @update:value="(v)=>{step.heroSkillUuid = Number.isFinite(Number(v)) ? Number(v) : 0;updateNtf()}"
                  @create="createTagOnlyNumber($event, excelSkillMap, 'skills')"
                  clearable tag filterable
                  placeholder="使用技能"
              />
              <!--targetsId-->
              <div v-if="step.action == ActionEnum.UseHeroSkill
                           || (step.confirm && step.action == ActionEnum.OptRoomAction)
                           || step.action == ActionEnum.PlayCard" style="display: flex; align-items: center; width: 315px">
                <n-select style="flex: 0 0 140px;margin-right: 5px" tag filterable placeholder="选择目标" clearable
                          :options="[
                            {
                              label: '全部或自己(0)',
                              value: 0,
                              chs: '全部或自己'
                            },
                            ...nowCaseData.initYanWu?.customHeroes.map((hero, index)=>{
                              // 显示英雄id和(座位号)
                              return {
                                label: `座位(${index+1})`,
                                value: index + 1, // 表示座位号
                              }}) || [],
                          ]"
                          @update:value="(value)=>{if (!step.targetsId) {step.targetsId = []} if (Number.isFinite(Number(value))) {step.targetsId.push(Number(value))} updateNtf()}"/>
                <div style="flex: 1 1 0">
                  <n-tag style="cursor: pointer"
                         v-for="(item, index) in step.targetsId"
                         :key="randomUUID()" closable
                         @close="()=>{step.targetsId?.splice(index, 1); updateNtf()}">
                    {{ `座位(${item})` }}
                  </n-tag>
                </div>
              </div>
              <!--cards-->
              <n-select v-if="step.action == ActionEnum.UseHeroSkill
                           || (step.confirm && step.action == ActionEnum.OptRoomAction)
                           || step.action == ActionEnum.DisCard
                           || step.action == ActionEnum.PlayCard"
                        style="width: 430px" v-model:value="step.cards"
                        :options="excelCardsSelectOptionFromInit"
                        :fallback-option="excelCardsSelectFallbackOption"
                        multiple @update:value="()=>{updateNtf()}"
                        tag filterable
                        placeholder="使用卡牌"
                        clearable
                        @create="createTagOnlyNumber($event, excelCardsMap, 'cards')"
              />
              <!--cards-concat-->
              <n-tooltip trigger="hover">
                <template #trigger>
                  <div>
                    <n-dynamic-tags
                        v-if="step.action == ActionEnum.UseHeroSkill
                          || (step.confirm && step.action == ActionEnum.OptRoomAction)"
                        v-model:value="computed(()=>{return step.cardsConcat?.map(n=>n.toString())}).value"
                        @create="(label:string)=>{return Number(label).toString()}"
                        @update:value="(newArr)=>{step.cardsConcat = newArr.filter(n=>Number.isFinite(Number(n))).map(n=>Number(n)); updateNtf(); console.log(step.cardsConcat)}"
                    />
                  </div>
                </template>
                添加额外参数
              </n-tooltip>
              <!--transCardSkill(把xxx当作xxx打出)-->
              <n-select
                  v-if="step.action == ActionEnum.UseHeroSkill
                     || (step.confirm && step.action == ActionEnum.OptRoomAction)
                     || step.action == ActionEnum.PlayCard"
                  style="width: 160px" v-model:value="step.transCardSkill"
                  :options="excelSkillSelectOption"
                  @create="createTagOnlyNumber"
                  clearable tag filterable :filter="chsAndNumSelectFilter"
                  placeholder="当xx牌打出" min="0" max="999999" @update:value="()=>{updateNtf()}"
              />
              <!--超时时间-->
              <n-input-number
                  style="width: 160px" v-model:value="step.timeout"
                  placeholder="超时时间" min="0" max="9999" @update:value="()=>{updateNtf()}">
                <template #suffix>
                  s
                </template>
              </n-input-number>
            </div>
            <div v-if="step.assets">
              <VueDraggable
                  v-model="step.assets as any"
                  @start="onStart"
                  @end="onEnd"
                  :animation="150"
                  :scroll="true"
                  :scroll-sensitivity="300"
                  :scroll-speed="20"
                  handle=".custom-drag-handle-asset"
                  class="steps"
              >
                <n-card header-style="font-size: 16px; padding: 10px" v-for="(asset, aindex) in step.assets"
                        :title="'断言 ' + (aindex+1)" closable
                        @close="()=>{delAsset(step.assets, aindex); /*updateNtf()*/}"
                        :key="(asset as Asset & DragAttr).uuid">
                  <template #header-extra>
                    <div class="custom-drag-handle-asset" style="cursor: pointer; min-width: 100px">
                      <n-button type="info" dashed>
                        拖动
                      </n-button>
                    </div>
                  </template>
                  <AssetCard v-if="asset" :step-index="index" :asset-index="aindex"/>
                </n-card>
              </VueDraggable>
            </div>
            <div>
              <n-button strong secondary type="primary" @click="()=>{addAsset(step); updateNtf()}">
                新增断言
              </n-button>
              <n-button strong secondary type="primary" @click="()=>{addStep(index); updateNtf()}">
                新增
              </n-button>
            </div>
          </div>
        </n-card>
      </TransitionGroup>
    </VueDraggable>
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

.sort-target {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.cardHeaderExtra {
  display: flex;
  width: 100%;
  gap: 10px;
}

.cardHeaderExtra > :nth-child(1) {
  flex: 0 0 auto;
}

.cardHeaderExtra > :nth-child(2) {
  flex: 0 0 auto;
}

.cardHeaderExtra > :nth-child(3) {
  flex: 1 0 500px
}

.cardHeaderExtra > :nth-child(4) {
  flex: 0 0 auto;
}

.stepCardContainer {
  display: flex;
  flex-direction: column;
  gap: 10px
}

.stepCardContainer > :first-child {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  align-items: center
}

.steps {
  display: flex;
  flex-direction: column;
  gap: 5px
}

.stepCardContainer > :last-child {
  display: flex;
  justify-content: right;
  gap: 10px
}
</style>