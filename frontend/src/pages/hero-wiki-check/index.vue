/**
 * 武将 Wiki 检查页面
 *
 * 用于对比 Excel 配置中的武将数据变化
 */
<script setup lang="ts">
import {useHeroWikiCheck} from "./composables/use-hero-wiki";
import HeroPanel from "./components/hero-panel.vue";
import PathConfigInput from "@shared/components/path-config-input/index.vue";
import type {HeroCompleteData} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/herowiki_def/models";

// 布局层:App.vue 入口按 UA(Android)挂 .is-mobile。读取以收起锚点栏(150px),回收 content 宽度。
// 仅真移动设备;PC 不触发。
const isMobile = typeof document !== 'undefined' && document.documentElement.classList.contains('is-mobile')

const {
  excelDir, oldJsonPath, isLoading, isSaving, errorMsg,
  diffExcels, transMap,
  hasDiffResult, diffSummary,
  searchName, filterCountry, filterIsNewHero, filterIsGacha, filterIsOpen, filterDiffType,
  countryOptions, setDiffTypeFilter,
  filteredHeroes, filteredAnchors,
  removedHeroesDetail,
  runCheck, saveResult,
  saveConfig,
} = useHeroWikiCheck()
</script>

<template>
  <div id="HeroWikiResCheck">
    <div style="display: flex; height: 100%;">
      <div><!--可以放点侧边栏--></div>
      <div class="main-content">
        <!-- 配置区域 -->
        <n-card class="config-card" size="small">
          <n-space align="center" wrap>
            <PathConfigInput
              v-model:excel-dir="excelDir"
              v-model:second-value="oldJsonPath"
              excel-label=""
              second-label=""
              excel-placeholder="Excel 配置目录路径"
              second-placeholder="历史数据 JSON 文件路径(可选)"
              input-width="280px"
              :on-save="saveConfig"
            />
            <n-button type="primary" :loading="isLoading" @click="runCheck">
              执行检查
            </n-button>
            <n-button
              type="info"
              :loading="isSaving"
              :disabled="!diffExcels"
              @click="saveResult"
            >
              保存结果
            </n-button>
            <n-text v-if="errorMsg" type="error">{{ errorMsg }}</n-text>
          </n-space>
        </n-card>
        <template v-if="diffExcels">
          <!-- 顶部统计区域 (固定) -->
          <n-card v-if="hasDiffResult" class="global-diff-summary" size="small">
            <n-space align="center">
              <n-tag type="info" :bordered="false">
                总变化: {{ diffSummary?.TotalChanges || 0 }}
              </n-tag>
              <n-tag
                type="success"
                :bordered="false"
                :class="{ 'clickable-tag': diffSummary?.AddedHeroes?.length, 'active-tag': filterDiffType === 'added' }"
                @click="diffSummary?.AddedHeroes?.length && setDiffTypeFilter('added')"
              >
                新增: {{ diffSummary?.AddedHeroes?.length || 0 }}
              </n-tag>
              <n-tag
                type="error"
                :bordered="false"
                :class="{ 'clickable-tag': diffSummary?.RemovedHeroes?.length, 'active-tag': filterDiffType === 'removed' }"
                @click="diffSummary?.RemovedHeroes?.length && setDiffTypeFilter('removed')"
              >
                删除: {{ diffSummary?.RemovedHeroes?.length || 0 }}
              </n-tag>
              <n-tag
                type="warning"
                :bordered="false"
                :class="{ 'clickable-tag': diffSummary?.ModifiedHeroes?.length, 'active-tag': filterDiffType === 'modified' }"
                @click="diffSummary?.ModifiedHeroes?.length && setDiffTypeFilter('modified')"
              >
                修改: {{ diffSummary?.ModifiedHeroes?.length || 0 }}
              </n-tag>
            </n-space>
          </n-card>
          <!-- 筛选条件区域 (固定) -->
          <n-card class="filter-card" size="small">
            <n-space align="center" wrap>
              <n-input
                v-model:value="searchName"
                placeholder="搜索武将名称"
                clearable
                style="width: 200px"
              />
              <n-select
                v-model:value="filterCountry"
                multiple
                placeholder="势力"
                clearable
                style="width: 150px"
                :options="countryOptions"
              />
              <n-checkbox :checked="filterIsNewHero === true" @update:checked="filterIsNewHero = $event ? true : null">
                新武将
              </n-checkbox>
              <n-checkbox :checked="filterIsGacha === true" @update:checked="filterIsGacha = $event ? true : null">
                抽卡武将
              </n-checkbox>
              <n-checkbox :checked="filterIsOpen === true" @update:checked="filterIsOpen = $event ? true : null">
                已开放
              </n-checkbox>
              <n-text depth="3" style="margin-left: auto">
                筛选结果: {{ filteredHeroes.length }} / {{ diffExcels?.HeroDiff?.length || 0 }}
              </n-text>
            </n-space>
          </n-card>
          <!-- 武将面板列表 (可滚动) -->
          <div class="hero-list-container">
            <n-scrollbar>
              <!-- 正常武将列表 -->
              <HeroPanel v-for="(v,k) in filteredHeroes" :seq="k" :hero-info="v"
                         :diff-excels="diffExcels" :diff-index-map="transMap()" :key="v.EHeroId"/>
              <!-- 删除的武将列表 -->
              <template v-if="removedHeroesDetail.length > 0 && (filterDiffType === null || filterDiffType === 'removed')">
                <n-divider style="margin: 24px 0;">
                  <n-text type="error" style="font-size: 14px;">已删除的武将 ({{ removedHeroesDetail.length }})</n-text>
                </n-divider>
                <HeroPanel
                  v-for="(hero, idx) in removedHeroesDetail"
                  :seq="filteredHeroes.length + idx"
                  :hero-info="hero.heroInfo"
                  :diff-excels="diffExcels"
                  :diff-index-map="transMap()"
                  :hero-wiki-data="hero.heroWikiData"
                  :is-removed="true"
                  :key="'removed-' + hero.eHeroId"
                />
              </template>
            </n-scrollbar>
          </div>
        </template>
      </div>
      <div v-if="diffExcels" v-show="!isMobile" class="anchor-container">
        <n-scrollbar>
          <!-- 正常武将导航 -->
          <n-anchor :show-rail="false" :bound="114" :show-background="true">
            <n-anchor-link v-for="(anchor, k) in filteredAnchors"
                           :href="'#HeroId:'+anchor.hero.Id">
              <template #title>
                <div style="display: flex; align-items: center">
                  <div style="white-space: pre-line; line-height: 1.4; display: flex">
                    <div style="flex: 0 0 35px">
                      {{ (k + 1) + '. ' }}
                    </div>
                    <div style="flex: 1 1 0"
                         :style="{color: (diffExcels.HeroWikiDiffResult?.HeroesDiff[anchor.hero.EHeroId]?.ChangeCount || diffExcels.HeroWikiDiffResult?.HeroesDiff[anchor.hero.EHeroId]?.ChangeCount == 0) ? 'yellow' : 'inherit'}">
                      {{ anchor.hero.Name }}
                    </div>
                  </div>
                </div>
              </template>
            </n-anchor-link>
          </n-anchor>
          <!-- 删除武将导航 -->
          <template v-if="removedHeroesDetail.length > 0 && (filterDiffType === null || filterDiffType === 'removed')">
            <n-divider style="margin: 12px 0">
              <n-text depth="3" style="font-size: 11px;">已删除</n-text>
            </n-divider>
            <n-anchor :show-rail="false" :bound="114" :show-background="true">
              <n-anchor-link v-for="(hero, idx) in removedHeroesDetail"
                             :href="'#HeroId:removed-'+hero.eHeroId">
                <template #title>
                  <div style="display: flex; align-items: center">
                    <div style="white-space: pre-line; line-height: 1.4; display: flex">
                      <div style="flex: 0 0 35px">
                        {{ (filteredHeroes.length + idx + 1) + '. ' }}
                      </div>
                      <div style="flex: 1 1 0; color: #ff6b6b;">
                        {{ hero.name }}
                      </div>
                    </div>
                  </div>
                </template>
              </n-anchor-link>
            </n-anchor>
          </template>
        </n-scrollbar>
      </div>
    </div>
  </div>
</template>
<style scoped>
#HeroWikiResCheck {
  position: absolute;
  width: 100%;
  height: 100%;
  box-sizing: border-box;
  color: white;
}
.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}
.global-diff-summary {
  flex-shrink: 0;
  margin-bottom: 8px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}
.config-card {
  flex-shrink: 0;
  margin-bottom: 8px;
  background: #363636;
}
.filter-card {
  flex-shrink: 0;
  margin-bottom: 8px;
  background: #363636;
}
.hero-list-container {
  flex: 1;
  overflow: hidden;
}
.anchor-container {
  flex-shrink: 0;
  width: 150px;
  max-height: 100%;
}
/* 可点击标签样式 */
.clickable-tag {
  cursor: pointer;
  transition: transform 0.2s, opacity 0.2s;
}
.clickable-tag:hover {
  transform: scale(1.05);
  opacity: 0.9;
}
/* 激活状态标签 */
.active-tag {
  box-shadow: 0 0 8px rgba(255, 255, 255, 0.6);
}
</style>
