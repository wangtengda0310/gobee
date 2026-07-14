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
import {computed, h, nextTick, ref} from "vue";
import {CustomHero} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/function-test";
import {canUseIdentityOption, excelIdentityList, getIdentityColorHex, getSeatColorHex} from "../config/Identity";
import {excelSkillMap} from "../config/Skill";
import {excelSkillDescMap} from "../config/SkillUI";
import {excelCardsMap} from "../config/Card";
import {NTag, NTooltip, useDialog} from "naive-ui";
import {excelHeroMap} from "@shared/config/hero";

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

// 渲染「删除技能」下拉框的已选标签：用 n-tooltip 包裹，悬浮展示该技能的描述文案
// 无描述文案时返回默认 tag，不显示空 tooltip
// naive-ui render-tag 签名为单参数对象：{ option, handleClose }
const renderSkillDescTag = ({option, handleClose}: { option: { label: string; value: number }; handleClose: () => void }) => {
  const desc = excelSkillDescMap.value[option.value]
  const tag = h(NTag, {
    closable: true,
    onClose: () => {
      handleClose()
    }
  }, {default: () => option.label})
  if (!desc) {
    return tag
  }
  return h(NTooltip, {trigger: "hover", placement: "top", style: {maxWidth: '500px'}}, {
    trigger: () => h('span', {style: 'display:inline-block'}, [tag]),
    default: () => desc
  })
}

// 渲染「删除/增加技能」下拉框的选项 label：n-tooltip 包裹，悬浮显示该技能的描述文案。
// 复用 excelSkillDescMap；无描述时返回纯 label（保持默认外观，不显示空 tooltip）。
// 取舍：使用 render-label 后，filterable 搜索的关键词高亮会失效（过滤本身仍正常）。
// naive-ui n-select render-label 签名：(option: SelectOption) => VNodeChild
const renderSkillDescLabel = (option: { label: string; value: number }) => {
  const desc = excelSkillDescMap.value[option.value]
  if (!desc) {
    return option.label
  }
  return h(NTooltip, {trigger: "hover", placement: "right", style: {maxWidth: '500px'}}, {
    trigger: () => h('span', {style: 'display: inline-block; width: 100%'}, [option.label]),
    default: () => desc
  })
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
  addSkills: false,
  cardSkill: false,
  skillCards: nowCaseData.value?.children?.map(c => false) || [],
})

// 将卡牌ID格式化为按钮显示文本（简化版，避免内容过挤）
// 格式：卡牌名(卡牌ID)，如 桃(1047)；查不到配置时回退为纯ID
const formatCardLabel = (cardId: number): string => {
  const card = excelCardsMap.value[cardId]
  if (!card) {
    return String(cardId)
  }
  return `${card.Name}(${card.Id})`
}

// 将技能ID格式化为按钮显示文本（与卡牌按钮统一的 名(ID) 样式，用于删除技能区）
// 格式：技能名(技能ID)，如 咆哮(1047)；查不到配置时回退为纯ID
const formatSkillLabel = (skillId: number): string => {
  const skill = excelSkillMap.value[skillId]
  if (!skill) {
    return String(skillId)
  }
  return `${skill.SkillName}(${skill.Id})`
}

// 摸牌堆顺序按钮的座位身份色：摸牌堆每 2 张一组对应一个座位的摸牌（座位轮流摸 2 张），
// 映射 index → 座位号 floor(index/2) % 座位数 + 1 → 该座位的身份阵营色（getIdentityColorHex 4 色）。
// 用于顺序调整模式下给按钮加边框，直观区分各座位的摸牌归属。
const cardSeatColor = (index: number): string | undefined => {
  const heroes = nowCaseData.value?.initYanWu?.customHeroes
  if (!heroes?.length) return undefined
  const seatNo = Math.floor(index / 2) % heroes.length + 1
  return getSeatColorHex(heroes, seatNo)
}

// 摸牌堆顺序按钮的 style：组间（每 2 张）留 8px 间隔 + 设自定义 --seat-border-color（座位身份色）。
// naive-ui 的 --n-border-color 由其 inline 注入根元素、会覆盖外部 :style，故改用自定义 variable，
// 再由 <style> 里的 :deep(.n-button__border) 引用，作用到边框；未设时回退 naive-ui 默认色。
const cardSeatStyle = (index: number): Record<string, string> => {
  const color = cardSeatColor(index)
  return {
    marginLeft: index > 0 && index % 2 === 0 ? '8px' : '',
    '--seat-border-color': color || ''
  }
}
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
            <!--摸牌堆顺序按钮：两两分组，组内保持原有紧凑间距，组间留 8px 间隔，便于区分第几回合摸牌（每回合摸2张）-->
            <VueDraggable v-if="orderMod.cards && nowCaseData.initYanWu.cards.length > 0"
                          v-model="nowCaseData.initYanWu.cards"
                          :animation="150">
              <n-button
                  v-for="(item, index) in nowCaseData.initYanWu.cards"
                  :key="item"
                  :style="cardSeatStyle(index)">
                {{ `[${index + 1}] ` + formatCardLabel(item) }}
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
                    header-style="font-size: 16px; padding: 10px"
                    style="margin-bottom: 10px" data-testid="hero-config-card" closable @close="()=>{delBot(index); updateNtf()}"
            >
              <!--座位标题旁的圆点 = 该座位身份阵营色（按 IdentityClass 大类：主红/忠金/反绿/内蓝），来源 config/Identity.ts getIdentityColorHex-->
              <template #header>
                <span style="display: inline-flex; align-items: center; gap: 8px">
                  座位 {{ index + 1 }}
                  <span v-if="getIdentityColorHex(bot.identity)"
                        :title="`身份阵营色 ${getIdentityColorHex(bot.identity)}`"
                        :style="{ display: 'inline-block', width: '12px', height: '12px', borderRadius: '50%', backgroundColor: getIdentityColorHex(bot.identity), border: '1px solid rgba(0,0,0,0.3)', boxShadow: '0 0 2px rgba(0,0,0,0.3)' }"></span>
                </span>
              </template>
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
                    {{ `[${index + 1}] ` + formatCardLabel(item) }}
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
                    {{ `[${index + 1}] ` + formatCardLabel(item) }}
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
                    {{ `[${index + 1}] ` + formatCardLabel(item) }}
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
                    {{ `[${index + 1}] ` + formatCardLabel(item) }}
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
                  <n-tooltip v-for="(item, sIndex) in bot.delSkills" :key="item" trigger="hover" placement="top" :style="{maxWidth: '500px'}">
                    <template #trigger>
                      <n-button>
                        {{ `[${sIndex + 1}] ` + formatSkillLabel(item) }}
                      </n-button>
                    </template>
                    {{ excelSkillDescMap[item] || '暂无技能描述' }}
                  </n-tooltip>
                </VueDraggable>
                <div style="display: flex; align-items: center">
                  <span style="flex: 0 0 100px">删除技能</span>
                  <n-select style="flex: 1 1 auto" v-model:value="bot.delSkills"
                            :options="excelHeroInitSkillSelectOption(bot.heroId)"
                            data-testid="del-skills-select"
                            multiple filterable :filter="chsAndNumSelectFilter" placeholder="技能"
                            :render-tag="renderSkillDescTag"
                            :render-label="renderSkillDescLabel"
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
                <!--增加技能区（展示逻辑与删除技能一致：已选技能以标签内嵌于多选下拉框；顺序调整时按钮悬浮提示技能描述）-->
                <VueDraggable v-if="orderMod.addSkills && bot.addSkills.length > 0"
                              v-model="bot.addSkills" :animation="150">
                  <n-tooltip v-for="(item, sIndex) in bot.addSkills" :key="item" trigger="hover" placement="top" :style="{maxWidth: '500px'}">
                    <template #trigger>
                      <n-button>
                        {{ `[${sIndex + 1}] ` + formatSkillLabel(item) }}
                      </n-button>
                    </template>
                    {{ excelSkillDescMap[item] || '暂无技能描述' }}
                  </n-tooltip>
                </VueDraggable>
                <div style="display: flex; align-items: center">
                  <span style="flex: 0 0 100px">增加技能</span>
                  <n-select style="flex: 1 1 auto" v-model:value="bot.addSkills"
                            :options="excelSkillSelectOption"
                            data-testid="add-skills-select"
                            multiple filterable :filter="chsAndNumSelectFilter" placeholder="技能"
                            :render-tag="renderSkillDescTag"
                            :render-label="renderSkillDescLabel"
                            @update:value="()=>{updateNtf()}"/>
                  <n-switch style="flex: 0 0 100px; margin-left: 10px" v-model:value="orderMod.addSkills"
                            :round="false">
                    <template #checked>
                      顺序调整
                    </template>
                    <template #unchecked>
                      固定顺序
                    </template>
                  </n-switch>
                </div>
                <div style="display: flex; align-items: center">
                  <span style="flex: 0 0 100px">技能牌区</span>
                  <n-select style="flex: 0 0 200px;" :options="excelSkillSelectOption"
                            filterable :filter="chsAndNumSelectFilter" placeholder="技能" clearable
                            @update:value="(value)=>{if (!bot.skillCardsMap) {bot.skillCardsMap = {}} if (Number.isFinite(value)) {bot.skillCardsMap[value] = []} updateNtf()}"/>
                  <div style="flex: 1 1 auto">
                    <n-tag style="cursor: pointer"
                           v-for="(item, index) in computed(()=>Object.keys(bot.skillCardsMap)).value"
                           :key="item + '-' + index" closable
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
/*摸牌堆顺序按钮：按座位身份色覆盖 naive-ui 边框色。naive-ui .n-button__border / .n-button__state-border
  引用 --n-border-color(-hover/pressed)，但该 variable 由 naive-ui inline 注入根元素、覆盖外部 :style；
  改用自定义 --seat-border-color 经 :deep 作用到边框元素，未设时回退 naive-ui 默认色（不影响其他按钮）。
  同时覆盖 state-border，避免 hover/拖拽时闪回 naive-ui 默认浅绿色。*/
:deep(.n-button__border),
:deep(.n-button__state-border) {
  border-color: var(--seat-border-color, var(--n-border-color));
}
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