<script setup lang="ts">
import BadgeLabel from "@shared/components/badge-label/index.vue";

const SHEET = '赛季战令表|SeasonPass'

const props = defineProps<{
  title: string
  seasonPass: any
  highlight: boolean
  periodLabel: string
}>()
</script>

<template>
  <n-card
      :title="title"
      size="small"
      :bordered="false"
      :class="highlight ? 'seasonpass-card--current' : 'seasonpass-card--other'"
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
        <template #label><BadgeLabel :sheet="SHEET" field="Id" label="战令ID" /></template>
        <n-tag :bordered="false">{{ props.seasonPass?.Id }}</n-tag>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel :sheet="SHEET" field="Name" label="战令名称" /></template>
        <n-text strong>{{ props.seasonPass?.Name || '未命名' }}</n-text>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel :sheet="SHEET" field="LevelLimit" label="等级上限" /></template>
        <n-tag :bordered="false" type="primary">{{ props.seasonPass?.LevelLimit }}</n-tag>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel :sheet="SHEET" field="ExpPerLevel" label="每级经验" /></template>
        <n-text>{{ props.seasonPass?.ExpPerLevel }}</n-text>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel :sheet="SHEET" field="WeekExpLimit" label="每周经验上限" /></template>
        <n-text>{{ props.seasonPass?.WeekExpLimit }}</n-text>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel :sheet="SHEET" field="LoopExp" label="循环经验" /></template>
        <n-text>{{ props.seasonPass?.LoopExp }}</n-text>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel :sheet="SHEET" field="StartTime" label="开始时间" /></template>
        <n-text :type="props.seasonPass?.StartTime ? 'success' : 'warning'">
          {{ props.seasonPass?.StartTime || '未配置' }}
        </n-text>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel :sheet="SHEET" field="EndTime" label="结束时间" /></template>
        <n-text :type="props.seasonPass?.EndTime ? 'success' : 'warning'">
          {{ props.seasonPass?.EndTime || '未配置' }}
        </n-text>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel :sheet="SHEET" field="ExtTime" label="延长结束时间" /></template>
        <n-text>{{ props.seasonPass?.ExtTime || '无' }}</n-text>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel :sheet="SHEET" field="HeroId" label="武将ID" /></template>
        <n-tag :bordered="false" type="primary">{{ props.seasonPass?.HeroId || '-' }}</n-tag>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel :sheet="SHEET" field="SkinId" label="皮肤ID" /></template>
        <n-tag :bordered="false" type="primary">{{ props.seasonPass?.SkinId || '-' }}</n-tag>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel :sheet="SHEET" field="MailId" label="邮件ID" /></template>
        <n-tag :bordered="false">{{ props.seasonPass?.MailId || '-' }}</n-tag>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel :sheet="SHEET" field="ExpShowItemId" label="经验展示道具" /></template>
        <n-tag :bordered="false">{{ props.seasonPass?.ExpShowItemId || '-' }}</n-tag>
      </n-descriptions-item>
    </n-descriptions>
  </n-card>
</template>

<style scoped>
.seasonpass-card--current {
  border-left: 4px solid var(--n-primary-color, #2080f0) !important;
  background-color: var(--n-primary-color-hover, rgba(32, 128, 240, 0.08)) !important;
}
.seasonpass-card--other {
  border-left: 4px solid var(--n-border-color, rgba(255, 255, 255, 0.09)) !important;
  background-color: var(--n-action-color, rgba(255, 255, 255, 0.06)) !important;
}
</style>
