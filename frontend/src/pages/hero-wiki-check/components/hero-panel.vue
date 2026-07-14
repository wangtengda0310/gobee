/**
 * 武将面板组件
 *
 * 显示单个武将的详细信息，包括基础属性、技能、皮肤、成就等
 * 支持显示 diff 变化标记和已删除武将的特殊样式
 *
 * 各标签页内容已拆分为独立子组件：
 * - hero-basic-tab: 基础信息
 * - hero-ui-tab: 详细信息
 * - hero-gacha-tab: 卡池
 * - hero-skills-tab: 技能
 * - hero-skins-tab: 皮肤
 * - hero-achievements-tab: 成就
 * - hero-recommend-tab: 推荐布阵
 * - hero-robot-tab: 机器人行为
 * - hero-country-tab: 国家信息
 */
<script setup lang="ts">
import {HeroDiff} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/hero";
import {computed, h, ref} from "vue";
import {DiffIndexMap} from "../composables/hero-wiki.types";
import {DataContainer} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/diff";
import {HeroCompleteData} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/herowiki_def/models";
import {NBadge, NIcon} from "naive-ui";
import {GamepadFilled, InfoFilled} from "@vicons/material";
// 从antd导入图标（仅用于 tab 渲染函数）
import {
  BookFilled,
  SkinFilled,
  FireFilled,
  FlagFilled,
  RobotFilled,
  StarFilled,
  TrophyFilled
} from '@vicons/antd';
import HeroDiffDisplay from "./hero-diff-display.vue";
import {getCountryStyle, getHeroTypeStyle} from "./hero-panel-utils";
// 标签页子组件
import HeroBasicTab from "./hero-basic-tab.vue";
import HeroUiTab from "./hero-ui-tab.vue";
import HeroGachaTab from "./hero-gacha-tab.vue";
import HeroSkillsTab from "./hero-skills-tab.vue";
import HeroSkinsTab from "./hero-skins-tab.vue";
import HeroAchievementsTab from "./hero-achievements-tab.vue";
import HeroRecommendTab from "./hero-recommend-tab.vue";
import HeroRobotTab from "./hero-robot-tab.vue";
import HeroCountryTab from "./hero-country-tab.vue";

const props = defineProps<{
  seq: number
  heroInfo: HeroDiff
  diffExcels: DataContainer
  diffIndexMap: DiffIndexMap
  // 可选：直接传入武将的完整 Wiki 数据（用于删除武将显示）
  heroWikiData?: HeroCompleteData | null
  // 可选：标记是否为删除的武将
  isRemoved?: boolean
}>()

const activeTab = ref('basic');

const heroWiki = computed(() => {
  // 如果直接传入了 heroWikiData，直接使用
  if (props.heroWikiData) {
    return props.heroWikiData
  }
  // 否则从 diffExcels 中获取
  return props.diffExcels.HeroWikiDiff?.Heroes[props.heroInfo.EHeroId]
})
</script>

<template>
  <n-card
      :id="props.isRemoved ? ('HeroId:removed-' + props.heroInfo.EHeroId) : ('HeroId:' + props.heroInfo.Id)"
      :title="() => h('div', { style: 'display: flex; align-items: center; gap: 8px;' }, [
        h('span', { style: props.isRemoved ? 'font-size: 18px; font-weight: bold; color: #ff6b6b; text-decoration: line-through;' : 'font-size: 18px; font-weight: bold;' }, `${props.seq + 1}. ${props.heroInfo.Name}`),
        props.isRemoved ? h('span', {
          style: {
            padding: '2px 8px',
            borderRadius: '4px',
            fontSize: '12px',
            backgroundColor: 'rgba(255, 107, 107, 0.2)',
            color: '#ff6b6b',
            border: '1px solid #ff6b6b',
          }
        }, '已删除') : null,
        h('span', {
          style: {
            padding: '2px 8px',
            borderRadius: '4px',
            fontSize: '14px',
            backgroundColor: getCountryStyle(props.heroInfo.Country).bgColor,
            color: getCountryStyle(props.heroInfo.Country).color,
            border: `1px solid ${getCountryStyle(props.heroInfo.Country).color}`,
          }
        }, props.heroInfo.Country),
        h('span', {
          style: {
            padding: '2px 8px',
            borderRadius: '4px',
            fontSize: '14px',
            backgroundColor: getHeroTypeStyle(props.heroInfo.EHeroType).bgColor,
            color: getHeroTypeStyle(props.heroInfo.EHeroType).color,
          }
        }, props.heroInfo.EHeroType),
        !props.isRemoved ? h('div', { style: 'display: flex; align-items: center; gap: 4px;' }, [
          h(NBadge, {
            type: props.heroInfo.IsOpen ? 'success' : 'error',
            processing: !props.heroInfo.IsOpen,
            class: 'hero-status-badge'
          }, () => props.heroInfo.IsOpen ? '已开放' : '未开放'),
          h(HeroDiffDisplay, {
            diffExcels: props.diffExcels,
            heroId: props.heroInfo.EHeroId,
            heroName: props.heroInfo.Name
          })
        ]) : null
      ].filter(Boolean))"
      :segmented="{
      content: true,
      footer: true
    }"
      hoverable
      class="hero-card"
      :class="{ 'has-diff': props.diffExcels.HeroWikiDiffResult?.HeroesDiff[props.heroInfo.EHeroId], 'removed-hero': props.isRemoved }"
  >
    <template #header-extra>
      <div class="hero-id-badge" :style="props.isRemoved ? 'color: #ff6b6b;' : ''">ID: {{ props.heroInfo.Id }}</div>
    </template>

    <div v-if="heroWiki" class="hero-content">
      <!-- 使用Tabs组织内容，各标签页内容由子组件渲染 -->
      <n-tabs v-model:value="activeTab" type="line" animated>
        <!-- 基础信息页签 -->
        <n-tab-pane name="basic"
                    :tab="() => h('div', [h(NIcon,{component: InfoFilled}), '基础信息'])">
          <HeroBasicTab :hero-wiki="heroWiki"/>
        </n-tab-pane>

        <!-- UI信息页签 -->
        <n-tab-pane name="ui" :tab="() => h('div', [h(NIcon,{component: BookFilled}), '详细信息'])">
          <HeroUiTab :hero-wiki="heroWiki"/>
        </n-tab-pane>

        <!-- 卡池信息 -->
        <n-tab-pane name="gacha" :tab="() => h('div', [h(NIcon,{component: StarFilled}), '卡池'])">
          <HeroGachaTab :hero-wiki="heroWiki"/>
        </n-tab-pane>

        <!-- 技能信息页签 -->
        <n-tab-pane name="skills" :tab="() => h('div', [h(NIcon,{component: FireFilled}), '技能'])">
          <HeroSkillsTab :hero-wiki="heroWiki"/>
        </n-tab-pane>

        <!-- 皮肤信息页签 -->
        <n-tab-pane name="skins" :tab="() => h('div', [h(NIcon,{component: SkinFilled}), '皮肤'])">
          <HeroSkinsTab :hero-wiki="heroWiki"/>
        </n-tab-pane>

        <!-- 成就信息页签 -->
        <n-tab-pane name="achievements" :tab="() => h('div', [h(NIcon,{component: TrophyFilled}), '成就'])">
          <HeroAchievementsTab :hero-wiki="heroWiki"/>
        </n-tab-pane>

        <!-- 推荐布阵页签 -->
        <n-tab-pane name="recommend" :tab="() => h('div', [h(NIcon,{component: GamepadFilled}), '推荐布阵'])">
          <HeroRecommendTab :hero-wiki="heroWiki"/>
        </n-tab-pane>

        <!-- 机器人行为页签 -->
        <n-tab-pane name="robot" :tab="() => h('div', [h(NIcon,{component: RobotFilled}), '机器人行为'])">
          <HeroRobotTab :hero-wiki="heroWiki"/>
        </n-tab-pane>

        <!-- 国家信息页签 -->
        <n-tab-pane name="country" :tab="() => h('div', [h(NIcon,{component: FlagFilled}), '国家'])">
          <HeroCountryTab :hero-wiki="heroWiki"/>
        </n-tab-pane>
      </n-tabs>
    </div>

    <!-- 无数据时的提示 -->
    <div v-else class="no-data">
      <n-empty description="暂无详细数据"/>
    </div>
  </n-card>
</template>

<style scoped>
.hero-card {
  width: calc(100% - 2px);
  margin-bottom: 16px;
  transition: all 0.3s ease;
}

.hero-card:hover {
  transform: translateX(2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.hero-id-badge {
  padding: 4px 8px;
  background-color: #f0f0f0;
  border-radius: 4px;
  font-size: 12px;
  color: #666;
}

.hero-content {
  min-height: 400px;
}

.section-title {
  font-size: 16px;
  font-weight: bold;
  margin-bottom: 12px;
  padding-left: 8px;
  border-left: 4px solid #18a058;
}

.no-data {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 200px;
}

:deep(.n-statistic) {
  text-align: center;
}

:deep(.n-statistic .n-statistic-label) {
  font-size: 14px;
  color: #666;
}

:deep(.n-statistic .n-statistic-value) {
  font-size: 24px;
  font-weight: bold;
}

:deep(.n-card) {
  transition: all 0.3s ease;
}

:deep(.n-card .n-card-header) {
  font-weight: bold;
}

:deep(.n-tag) {
  margin-right: 4px;
  margin-bottom: 4px;
}

.hero-status-badge {
  margin-left: 8px;
}

.hero-card.has-diff {
  border: 2px solid #faad14;
  position: relative;
}

.hero-card.has-diff::before {
  content: '';
  position: absolute;
  top: -2px;
  right: -2px;
  width: 0;
  height: 0;
  border-style: solid;
  border-width: 0 20px 20px 0;
  border-color: transparent #faad14 transparent transparent;
}

/* 删除武将样式 */
.hero-card.removed-hero {
  border: 2px solid #ff6b6b;
  background: linear-gradient(135deg, rgba(255, 107, 107, 0.05) 0%, rgba(255, 107, 107, 0.1) 100%);
  position: relative;
}

.hero-card.removed-hero::before {
  content: '';
  position: absolute;
  top: -2px;
  right: -2px;
  width: 0;
  height: 0;
  border-style: solid;
  border-width: 0 20px 20px 0;
  border-color: transparent #ff6b6b transparent transparent;
}
</style>
