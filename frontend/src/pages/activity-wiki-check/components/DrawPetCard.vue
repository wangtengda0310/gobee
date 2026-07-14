/**
 * DrawPet 卡片子组件
 *
 * 展示单期结缘亭抽奖池信息，
 * 支持当前期（高亮）和上/下期（普通）两种样式
 */
<script setup lang="ts">
import {DrawPetDiff} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/draw_pet/models.js";
import {formatItemArray} from "@shared/composables/use-format-utils";
import BadgeLabel from "@shared/components/badge-label/index.vue";
import ResourcePreview from "./ResourcePreview.vue";
import {inject} from "vue";
import type {useResourceCheck} from "../composables/use-resource-check";

const resourceCheck = inject<ReturnType<typeof useResourceCheck>>('resourceCheck')

const SHEET = '结缘亭|DrawPet'

const props = defineProps<{
  title: string
  drawPet: DrawPetDiff
  highlight: boolean
  periodLabel: string
}>()
</script>

<template>
  <n-card
      :title="title"
      size="small"
      :bordered="false"
      :class="highlight ? 'drawpet-card--current' : 'drawpet-card--other'"
  >
    <template #header-extra>
      <n-tag v-if="highlight" type="primary" size="small" :bordered="false">
        当前关联
      </n-tag>
      <n-tag v-else size="small" :bordered="false">
        {{ periodLabel }}
      </n-tag>
    </template>
    <n-descriptions label-placement="left" :column="2" bordered>
      <n-descriptions-item>
        <template #label><BadgeLabel :sheet="SHEET" field="Id" label="抽奖池ID" /></template>
        <n-tag :bordered="false">{{ props.drawPet?.Id }}</n-tag>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel :sheet="SHEET" field="Name" label="抽奖池名称" /></template>
        <n-text strong>{{ props.drawPet?.Name || '未命名' }}</n-text>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel :sheet="SHEET" field="EnsureNewer" label="新手保底次数(M0)" /></template>
        <n-text>{{ props.drawPet?.EnsureNewer || 0 }}</n-text>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel :sheet="SHEET" field="EnsureCount" label="正常保底次数(M)" /></template>
        <n-text>{{ props.drawPet?.EnsureCount || 0 }}</n-text>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel :sheet="SHEET" field="FirstEnsureNewer" label="新手保底触发次数(X)" /></template>
        <n-text>{{ props.drawPet?.FirstEnsureNewer || 0 }}</n-text>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel :sheet="SHEET" field="OnceDropRule" label="单抽掉落规则" /></template>
        <n-tag :bordered="false" type="info">{{ props.drawPet?.OnceDropRule || '无' }}</n-tag>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel :sheet="SHEET" field="TenDropRule" label="十连掉落规则" /></template>
        <n-tag :bordered="false" type="info">{{ props.drawPet?.TenDropRule || '无' }}</n-tag>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel :sheet="SHEET" field="OnceItemCost" label="单抽消耗道具" /></template>
        <n-tag :bordered="false" type="warning">{{ formatItemArray(props.drawPet?.OnceItemCost) }}</n-tag>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel :sheet="SHEET" field="TenItemCost" label="十连消耗道具" /></template>
        <n-tag :bordered="false" type="warning">{{ formatItemArray(props.drawPet?.TenItemCost) }}</n-tag>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel :sheet="SHEET" field="DailyLimit" label="每日抽卡限制" /></template>
        <n-text>{{ props.drawPet?.DailyLimit || '无' }}</n-text>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel :sheet="SHEET" field="ActivityId" label="活动ID" /></template>
        <n-tag :bordered="false">{{ props.drawPet?.ActivityId?.join(', ') || '无' }}</n-tag>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel :sheet="SHEET" field="BigAwardCount" label="大奖保底次数" /></template>
        <n-text type="success" strong>{{ props.drawPet?.BigAwardCount || 0 }}</n-text>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel :sheet="SHEET" field="BigAwardItemId" label="大奖道具ID" /></template>
        <n-tag :bordered="false" type="error">{{ props.drawPet?.BigAwardItemId || '无' }}</n-tag>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel :sheet="SHEET" field="BigAwards" label="大奖道具" /></template>
        <n-tag :bordered="false" type="success">{{ formatItemArray(props.drawPet?.BigAwards) }}</n-tag>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel :sheet="SHEET" field="PartnerItem" label="伙伴道具" /></template>
        <n-tag :bordered="false">{{ props.drawPet?.PartnerItem ? props.drawPet.PartnerItem.ItemId + '×' + props.drawPet.PartnerItem.Count : '无' }}</n-tag>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel :sheet="SHEET" field="Byproducts" label="副产物" /></template>
        <n-tag :bordered="false" type="info">{{ props.drawPet?.Byproducts?.join(', ') || '无' }}</n-tag>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel :sheet="SHEET" field="StartTime" label="开始时间" /></template>
        <n-text>{{ props.drawPet?.StartTime || '无' }}</n-text>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel :sheet="SHEET" field="EndTime" label="结束时间" /></template>
        <n-text>{{ props.drawPet?.EndTime || '无' }}</n-text>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel :sheet="SHEET" field="ProbabilityContents" label="概率公示内容" /></template>
        <n-tag :bordered="false" type="info">{{ props.drawPet?.ProbabilityContents?.join(', ') || '无' }}</n-tag>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel :sheet="SHEET" field="ProbabilityPercents" label="概率公示百分比" /></template>
        <n-tag :bordered="false" type="info">{{ props.drawPet?.ProbabilityPercents?.join(', ') || '无' }}</n-tag>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel :sheet="SHEET" field="DrawPetTitleIcon" label="标题图标" /></template>
        <ResourcePreview :value="props.drawPet?.DrawPetTitleIcon" :status="resourceCheck?.getStatus(props.drawPet?.DrawPetTitleIcon || '')" />
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel :sheet="SHEET" field="DrawPetTitleDescBg" label="标题描述背景" /></template>
        <ResourcePreview :value="props.drawPet?.DrawPetTitleDescBg" :status="resourceCheck?.getStatus(props.drawPet?.DrawPetTitleDescBg || '')" />
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel :sheet="SHEET" field="DrawRuleContent" label="抽奖规则内容" /></template>
        <ResourcePreview :value="props.drawPet?.DrawRuleContent" :status="resourceCheck?.getStatus(props.drawPet?.DrawRuleContent || '')" />
      </n-descriptions-item>
    </n-descriptions>
  </n-card>
</template>

<style scoped>
.drawpet-card--current {
  border-left: 4px solid var(--n-primary-color, #2080f0) !important;
  background-color: var(--n-primary-color-hover, rgba(32, 128, 240, 0.08)) !important;
}
.drawpet-card--other {
  border-left: 4px solid var(--n-border-color, rgba(255, 255, 255, 0.09)) !important;
  background-color: var(--n-action-color, rgba(255, 255, 255, 0.06)) !important;
}
</style>
