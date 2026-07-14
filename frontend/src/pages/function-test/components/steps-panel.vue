<script setup lang="ts">
import {nowCaseData, normalizeCaseSteps, updateNtf, DragAttr} from "../composables/use-case-data";
import {SortableEvent, VueDraggable} from 'vue-draggable-plus'
import {
  ActionEnum,
  actionsSelectOption,
  AssetEnum,
  isStepsHasAck
} from "../composables/StepActionsAndAssetsSelect";
import {excelHeroMap} from "@shared/config/hero";
import {excelCardsMap} from "../config/Card";
import {computed, ref, watch} from "vue";
import {excelSkillMap} from "../config/Skill";
import {getSeatColorHex} from "../config/Identity";
import {Asset, Step} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/function-test";
import AssetCard from "./asset-card.vue";
import AssetCardHeaderExtra from "./asset-card-header-extra.vue";
import {useDialog} from 'naive-ui'
import {
  chsAndNumSelectFilter,
  createTagOnlyNumber,
  excelCardsSelectFallbackOption,
  excelCardsSelectOptionFromInit,
  excelSkillSelectOption,
  excelSkillSelectOptionFromInit
} from "../composables/HeroAndCardsAndSkillsSelect";

const onStart = (e: SortableEvent) => {
  console.log(e)
}

const onEnd = () => {
  normalizeCaseSteps(nowCaseData.value?.caseSteps)
  bumpStepListVersion()
  updateNtf()
}

const stepListVersion = ref(0)
const bumpStepListVersion = () => {
  stepListVersion.value++
}

watch(
    () => nowCaseData.value?.caseSteps,
    (steps) => normalizeCaseSteps(steps),
    {immediate: true}
)

const stepKey = (step: Step & DragAttr) => {
  if (!step.uuid) {
    step.uuid = crypto.randomUUID()
  }
  return step.uuid
}

/** 按当前数组位置显示序号，避免拖拽后 DOM 仍显示旧 index */
const stepDisplayNumber = (step: Step & DragAttr) => {
  const steps = nowCaseData.value?.caseSteps
  if (!steps) return 0
  const idx = steps.findIndex(s => s.uuid === step.uuid)
  return idx < 0 ? 0 : idx + 1
}

const replaceCaseSteps = (steps: (Step & DragAttr)[]) => {
  if (!nowCaseData.value) return
  normalizeCaseSteps(steps)
  nowCaseData.value.caseSteps = steps
  bumpStepListVersion()
}

// 座位下拉选项：显示"武将名 座位(n)"，如"赵云 座位(1)"。武将名取自 excelHeroMap，查不到回退 heroId。
// 放在 computed 而非 template 内联，避免 :options 属性里的多行模板字符串导致 vue 编译失败。
const seatOptions = computed(() => {
    return (nowCaseData.value?.initYanWu?.customHeroes ?? []).map((hero, index) => {
        const heroName = excelHeroMap.value[hero.heroId]?.Name || hero.heroId
        return {
            label: `${heroName} 座位(${index + 1})`,
            value: index + 1
        }
    })
})

const targetSelectOptions = computed(() => [
    {
        label: '全部或自己(0)',
        value: 0,
        chs: '全部或自己'
    },
    ...seatOptions.value,
])

const updateTargetsId = (step: Step, value: number[] | null) => {
    step.targetsId = value && value.length > 0 ? value : undefined
    updateNtf()
}

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
      const steps = nowCaseData.value?.caseSteps
      if (!steps) return
      replaceCaseSteps(steps.filter((_, i) => i !== index) as (Step & DragAttr)[])
      updateNtf()
    },
    onNegativeClick: () => {
    }
  })
}
const addStep = (index: number) => {
  const steps = [...(nowCaseData.value?.caseSteps ?? [])] as (Step & DragAttr)[]
  const newStep = Object.assign(new Step({
    desc: "",
    action: ActionEnum.Sleep,
    robotIdx: 1,
    sleepTime: 0,
    timeout: 1
  }), {
    uuid: crypto.randomUUID(),
  }) as Step & DragAttr
  replaceCaseSteps([
    ...steps.slice(0, index + 1),
    newStep,
    ...steps.slice(index + 1),
  ])
}

const copyStep = (index: number) => {
  const steps = nowCaseData.value?.caseSteps
  if (!steps?.[index]) return
  const newStep = JSON.parse(JSON.stringify(steps[index])) as Step & DragAttr
  newStep.uuid = crypto.randomUUID()
  newStep.assets?.forEach(asset => {
    (asset as Asset & DragAttr).uuid = crypto.randomUUID()
  })
  replaceCaseSteps([
    ...steps.slice(0, index + 1),
    newStep,
    ...steps.slice(index + 1),
  ] as (Step & DragAttr)[])
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
  <div v-if="nowCaseData?.caseSteps" data-testid="steps-panel">
    <VueDraggable
        :key="stepListVersion"
        v-model="nowCaseData.caseSteps as any"
        @start="onStart"
        @end="onEnd"
        :animation="150"
        :scroll="true"
        :scroll-sensitivity="300"
        :scroll-speed="20"
        handle=".custom-drag-handle"
        class="sort-target"
    >
        <n-card header-style="font-size: 18px; padding: 10px" v-for="(step, index) in nowCaseData.caseSteps"
                :key="stepKey(step)" :id="'动作 ' + stepDisplayNumber(step)"
                data-testid="action-step-card"
                closable @close="()=>{delStep(index); updateNtf()}"
        >
          <!--动作卡片标题旁圆点 = step.robotIdx 对应座位的身份阵营色（响应座位下拉改变），config/Identity.ts getSeatColorHex-->
          <template #header>
            <span style="display: inline-flex; align-items: center; gap: 8px">
              动作 {{ stepDisplayNumber(step) }}
              <span v-if="getSeatColorHex(nowCaseData.initYanWu?.customHeroes, step.robotIdx)"
                    :title="`座位${step.robotIdx} 身份阵营色 ${getSeatColorHex(nowCaseData.initYanWu?.customHeroes, step.robotIdx)}`"
                    :style="{ display: 'inline-block', width: '12px', height: '12px', borderRadius: '50%', backgroundColor: getSeatColorHex(nowCaseData.initYanWu?.customHeroes, step.robotIdx), border: '1px solid rgba(0,0,0,0.3)', boxShadow: '0 0 2px rgba(0,0,0,0.3)' }"></span>
            </span>
          </template>
          <template #header-extra>
            <div class="cardHeaderExtra" data-testid="action-step-header-extra">
              <n-button class="custom-drag-handle" type="info" dashed>
                拖动
              </n-button>
              <n-tooltip trigger="hover">
                <template #trigger>
                  <n-button secondary type="success" class="aiDesc" data-testid="apply-ai-desc-btn" @click="applyAiDesc(step)">
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
              <n-select style="width: 160px" v-model:value="step.robotIdx"
                        @create="createTagOnlyNumber($event)"
                        :consistent-menu-width="false"
                        :options="seatOptions" @update:value="()=>{updateNtf()}" tag filterable/>
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
              <n-select
                  v-if="step.action == ActionEnum.UseHeroSkill
                           || (step.confirm && step.action == ActionEnum.OptRoomAction)
                           || step.action == ActionEnum.PlayCard"
                  style="width: 315px"
                  :value="step.targetsId ?? []"
                  multiple
                  tag
                  filterable
                  placeholder="选择目标"
                  clearable
                  :consistent-menu-width="false"
                  :options="targetSelectOptions"
                  @update:value="(value) => updateTargetsId(step, value)"
              />
              <!--cards-->
              <n-select v-if="step.action == ActionEnum.UseHeroSkill
                           || (step.confirm && step.action == ActionEnum.OptRoomAction)
                           || step.action == ActionEnum.DisCard
                           || step.action == ActionEnum.PlayCard"
                        style="width: 430px" v-model:value="step.cards" data-testid="step-cards-select"
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
                        data-testid="asset-card"
                        closable
                        @close="()=>{delAsset(step.assets, aindex); /*updateNtf()*/}"
                        :key="(asset as Asset & DragAttr).uuid">
                  <template #header>
                    断言 {{ aindex + 1 }}
                  </template>
                  <template #header-extra>
                    <AssetCardHeaderExtra :step-index="index" :asset-index="aindex"/>
                  </template>
                  <AssetCard v-if="asset" :step-index="index" :asset-index="aindex"/>
                </n-card>
              </VueDraggable>
            </div>
            <div>
              <n-button strong secondary type="primary" @click="()=>{addAsset(step); updateNtf()}">
                新增断言
              </n-button>
              <n-button data-testid="add-step-btn" strong secondary type="primary" @click="()=>{addStep(index); updateNtf()}">
                新增
              </n-button>
            </div>
          </div>
        </n-card>
    </VueDraggable>
  </div>
  <n-empty v-else description="当前用例暂无动作步骤"/>
</template>

<style scoped>
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