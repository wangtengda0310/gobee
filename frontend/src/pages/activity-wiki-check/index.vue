/**
 * 活动 Wiki 检查页面
 *
 * 用于展示 Excel 配置中的活动数据
 * 包含配置区、筛选区、活动列表和右侧锚点导航
 *
 * 状态和业务逻辑在 composables/use-activity-wiki.ts 中管理
 */
<script setup lang="ts">
import ActivityPanel from "./components/activity-panel.vue";
import SeasonPassPanel from "./components/season-pass-panel.vue";
import PathConfigInput from "@shared/components/path-config-input/index.vue";
import TooltipCheckbox from "@shared/components/tooltip-checkbox/index.vue";
import {useActivityWikiCheck} from "./composables/use-activity-wiki";
import type {ActivityCompleteData} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/activitywiki_def/models.js";
import {provide, type Ref} from "vue";
import type {ResourceStatus} from "./composables/use-resource-check";

const {
  excelDir,
  secondPath,
  clientPath,
  isLoading,
  errorMsg,
  activityWikiDiff,
  ruleCoverage,
  runCheck,
  saveConfig,
  searchName,
  filterActivityType,
  filterShowTab,
  activityTypeOptions,
  filteredActivities,
  filteredAnchors,
  filteredSeasonPasses,
  resourceCheck,
} = useActivityWikiCheck()

// 向所有子组件注入资源检查能力
provide('resourceCheck', resourceCheck)
</script>

<template>
  <div id="ActivityWikiCheck">
    <div style="display: flex; height: 100%;">
      <div><!-- 可以放置侧边栏 --></div>
      <div class="main-content">
        <!-- 配置区域 -->
        <n-card class="config-card" size="small">
          <n-space align="center" wrap>
            <PathConfigInput
              v-model:excel-dir="excelDir"
              v-model:second-value="clientPath"
              excel-label=""
              second-label="客户端路径"
              excel-placeholder="Excel 配置目录路径"
              second-placeholder="客户端 Assets 目录（如 D:/work/client/Master/Card）"
              input-width="280px"
              :on-save="saveConfig"
            />
            <n-button type="primary" :loading="isLoading" @click="runCheck">
              执行检查
            </n-button>
            <n-text v-if="errorMsg" type="error">{{ errorMsg }}</n-text>
          </n-space>
        </n-card>

        <template v-if="activityWikiDiff">
          <!-- 筛选条件区域 -->
          <n-card class="filter-card" size="small">
            <n-space align="center" wrap>
              <n-input
                v-model:value="searchName"
                placeholder="搜索活动名称"
                clearable
                style="width: 200px"
              />
              <n-select
                v-model:value="filterActivityType"
                placeholder="活动类型"
                clearable
                style="width: 150px"
                :options="activityTypeOptions"
              />
              <TooltipCheckbox
                :checked="filterShowTab === true"
                @update:checked="filterShowTab = $event ? true : null"
                label="显示页签"
                tooltip="筛选是否在游戏界面显示页签的活动：勾选后只显示 ShowTab = true 的活动"
              />
              <n-text depth="3" style="margin-left: auto">
                筛选结果: {{ filteredActivities.length }} / {{ Object.keys(activityWikiDiff?.Activities || {}).length }}
              </n-text>
            </n-space>
          </n-card>

          <!-- 活动和战令统一滚动区域 -->
          <div class="activity-list-container">
            <n-scrollbar>
              <ActivityPanel
                v-for="(activity, index) in filteredActivities"
                :seq="index"
                :activity-data="activity as ActivityCompleteData"
                :excel-dir="excelDir"
                :rule-coverage="ruleCoverage"
                :key="activity?.Basic?.Id || index"
              />

              <!-- 战令配置区域（只展示本期一个卡片） -->
              <template v-if="filteredSeasonPasses.length > 0">
                <SeasonPassPanel
                  :seq="0"
                  :season-pass-data="filteredSeasonPasses[0]"
                  :excel-dir="excelDir"
                  :rule-coverage="ruleCoverage"
                  :key="filteredSeasonPasses[0]?.Basic?.Id || 'season-pass'"
                />
              </template>

              <n-empty
                v-if="filteredActivities.length === 0 && filteredSeasonPasses.length === 0"
                description="没有匹配的数据"
              />
            </n-scrollbar>
          </div>
        </template>
      </div>

      <!-- 右侧锚点导航 -->
      <div v-if="activityWikiDiff" class="anchor-container">
        <n-scrollbar>
          <n-anchor :show-rail="false" :bound="114" :show-background="true">
            <n-anchor-link
              v-for="(anchor, index) in filteredAnchors"
              :href="anchor.isSeasonPass ? '#SeasonPassId:' + (filteredSeasonPasses[0]?.Basic?.Id || '') : '#ActivityId:' + (anchor.activity?.Basic?.Id || '')"
            >
              <template #title>
                <div style="display: flex; align-items: center">
                  <div style="white-space: pre-line; line-height: 1.4; display: flex">
                    <div style="flex: 0 0 35px">
                      {{ (index + 1) + '. ' }}
                    </div>
                    <div style="flex: 1 1 0">
                      <template v-if="anchor.isSeasonPass">
                        <span style="color: #e3f2fd;">战令</span>
                      </template>
                      <template v-else>
                        {{ (anchor.activity as ActivityCompleteData)?.Basic?.Name || '未命名' }}
                      </template>
                    </div>
                  </div>
                </div>
              </template>
            </n-anchor-link>
          </n-anchor>
        </n-scrollbar>
      </div>
    </div>
  </div>
</template>

<style scoped>
#ActivityWikiCheck {
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

.activity-list-container {
  flex: 1;
  overflow: hidden;
}

.anchor-container {
  flex-shrink: 0;
  width: 150px;
  max-height: 100%;
}
</style>
