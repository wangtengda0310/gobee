/**
 * DrawSkin 卡片子组件
 *
 * 用于展示单个 DrawSkin 抽奖池信息，
 * 支持当前期（高亮）和上/下期（普通）两种样式
 * 使用 Naive UI 主题感知的 CSS 变量适配深色/浅色主题
 */
<script setup lang="ts">
import {DrawSkinDiff} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/draw_skin/models.js";
import {formatItemArray} from "@shared/composables/use-format-utils";
import BadgeLabel from "@shared/components/badge-label/index.vue";

const SHEET = '皮肤抽奖|DrawSkin'

const props = defineProps<{
  title: string
  drawSkin: DrawSkinDiff
  highlight: boolean
  periodLabel: string
}>()
</script>

<template>
  <n-card
      :title="title"
      size="small"
      :bordered="false"
      :class="highlight ? 'drawskin-card--current' : 'drawskin-card--other'"
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
        <n-tag :bordered="false">{{ props.drawSkin?.Id }}</n-tag>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel :sheet="SHEET" field="Name" label="抽奖池名称" /></template>
        <n-text strong>{{ props.drawSkin?.Name || '未命名' }}</n-text>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel :sheet="SHEET" field="ActivityId" label="关联活动ID" /></template>
        <n-tag :bordered="false">{{ props.drawSkin?.ActivityId || '无' }}</n-tag>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel :sheet="SHEET" field="OnceDropRule" label="单抽掉落规则" /></template>
        <n-tag :bordered="false" type="info">{{ props.drawSkin?.OnceDropRule || '无' }}</n-tag>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel :sheet="SHEET" field="TenDropRule" label="十连掉落规则" /></template>
        <n-tag :bordered="false" type="info">{{ props.drawSkin?.TenDropRule || '无' }}</n-tag>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel :sheet="SHEET" field="OnceItemCost" label="单抽消耗道具" /></template>
        <n-tag :bordered="false" type="warning">{{ formatItemArray(props.drawSkin?.OnceItemCost) }}</n-tag>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel :sheet="SHEET" field="TenItemCost" label="十连消耗道具" /></template>
        <n-tag :bordered="false" type="warning">{{ formatItemArray(props.drawSkin?.TenItemCost) }}</n-tag>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel :sheet="SHEET" field="BigAwardCount" label="大奖保底次数" /></template>
        <n-text type="success" strong>{{ props.drawSkin?.BigAwardCount || 0 }}</n-text>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel :sheet="SHEET" field="BigAwardItemId" label="大奖道具ID" /></template>
        <n-tag :bordered="false" type="error">{{ props.drawSkin?.BigAwardItemId || '无' }}</n-tag>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel :sheet="SHEET" field="StartTime" label="开始时间" /></template>
        <n-text>{{ props.drawSkin?.StartTime || '无' }}</n-text>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel :sheet="SHEET" field="EndTime" label="结束时间" /></template>
        <n-text>{{ props.drawSkin?.EndTime || '无' }}</n-text>
      </n-descriptions-item>
    </n-descriptions>
  </n-card>
</template>

<style scoped>
.drawskin-card--current {
  border-left: 4px solid var(--n-primary-color, #2080f0) !important;
  background-color: var(--n-primary-color-hover, rgba(32, 128, 240, 0.08)) !important;
}
.drawskin-card--other {
  border-left: 4px solid var(--n-border-color, rgba(255, 255, 255, 0.09)) !important;
  background-color: var(--n-action-color, rgba(255, 255, 255, 0.06)) !important;
}
</style>
