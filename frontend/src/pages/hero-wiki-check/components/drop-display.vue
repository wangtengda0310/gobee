/**
 * 掉落显示组件
 *
 * 显示武将的掉落配置信息，包括直接掉落、保底掉落和掉落组详情
 */
<script setup lang="ts">
import { ref } from 'vue';
import {
  NCollapse,
  NCollapseItem,
  NGrid,
  NGi,
  NTag,
  NSpace,
  NDescriptions,
  NDescriptionsItem,
  NBadge,
  NPopover,
  NScrollbar,
  NEmpty,
  NDivider,
  NText,
  NCard,
  NIcon,
} from 'naive-ui';
import {
} from '@vicons/material';


import {
  GiftFilled,
  CalendarFilled,
  TrophyFilled
} from '@vicons/antd';
import {HeroDropInfo} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/herowiki_def";

const props = defineProps<{
  dropInfo: HeroDropInfo | null;
}>();

// 控制是否显示更多详情
const expandGroups = ref<number[]>([]);

// 格式化日期
const formatDate = (date?: string): string => {
  if (!date || date === '') return '永久';
  return date;
};

// 格式化布尔值
const formatBoolean = (value: boolean | undefined, trueText: string = '是', falseText: string = '否'): string => {
  return value ? trueText : falseText;
};

// 判断是否有效期内
const isValidDate = (validDate?: string, expireDate?: string): boolean => {
  const now = new Date();

  if (validDate && validDate !== '') {
    const valid = new Date(validDate);
    if (now < valid) return false;
  }

  if (expireDate && expireDate !== '') {
    const expire = new Date(expireDate);
    if (now > expire) return false;
  }

  return true;
};
</script>

<template>
  <div class="drop-display-dark">
    <!-- 无数据时显示空状态 -->
    <n-empty v-if="!dropInfo" description="暂无掉落数据" size="small">
      <template #extra>
        <n-text depth="3">该武将暂无掉落配置</n-text>
      </template>
    </n-empty>

    <template v-else>
      <!-- 掉落概览卡片 -->
      <n-card :bordered="false" class="drop-overview-card">
        <template #header>
          <div class="section-header">
            <n-icon size="20" :component="GiftFilled" />
            <span>掉落概览</span>
          </div>
        </template>

        <n-grid :cols="24" :x-gap="16" :y-gap="16">
          <!-- 直接掉落统计 -->
          <n-gi :span="8">
            <div class="stat-card">
              <div class="stat-label">
                <n-icon size="16" :component="GiftFilled" />
                <span>直接掉落规则</span>
              </div>
              <div class="stat-value">{{ dropInfo.DirectDropRules?.length || 0 }}</div>
              <div class="stat-trend">
                涉及组数: {{ dropInfo.DirectDropRules?.reduce((acc, rule) => acc + (rule?.DropGroups?.length || 0), 0) || 0 }}
              </div>
            </div>
          </n-gi>

          <!-- 保底掉落统计 -->
          <n-gi :span="8">
            <div class="stat-card">
              <div class="stat-label">
                <n-icon size="16" :component="TrophyFilled" />
                <span>保底掉落规则</span>
              </div>
              <div class="stat-value">{{ dropInfo.GuaranteeDropRules?.length || 0 }}</div>
              <div class="stat-trend">
                涉及组数: {{ dropInfo.GuaranteeDropRules?.reduce((acc, rule) => acc + (rule?.DropGroups?.length || 0), 0) || 0 }}
              </div>
            </div>
          </n-gi>

          <!-- 掉落组统计 -->
          <n-gi :span="8">
            <div class="stat-card">
              <div class="stat-label">
                <n-icon size="16" :component="CalendarFilled" />
                <span>掉落组总数</span>
              </div>
              <div class="stat-value">{{ dropInfo.DropGroups?.length || 0 }}</div>
              <div class="stat-trend">
                掉落项: {{ dropInfo.DropGroups?.reduce((acc, group) => acc + (group?.DropItems?.length || 0), 0) || 0 }}
              </div>
            </div>
          </n-gi>

          <!-- 类型分布 -->
          <n-gi :span="24" v-if="dropInfo.ByDropType && Object.keys(dropInfo.ByDropType).length > 0">
            <n-divider />
            <div class="type-distribution">
              <span class="distribution-title">掉落类型分布</span>
              <div class="type-tags">
                <n-popover
                    v-for="(info, type) in dropInfo.ByDropType"
                    :key="type"
                    trigger="hover"
                    placement="top"
                >
                  <template #trigger>
                    <div class="type-tag-wrapper">
                      <div class="type-tag" :class="`type-${type}`">
                        <span class="type-name">{{ info?.TypeName || '未知' }}</span>
                        <span class="type-count">{{ info?.TotalCount }}</span>
                      </div>
                    </div>
                  </template>
                  <div class="type-popover">
                    <div class="popover-header">{{ info?.TypeName || '未知' }}</div>
                    <div class="popover-content">
                      <div v-for="rule in info?.DropRules" :key="rule?.RuleId" class="popover-rule">
                        <span class="rule-name">{{ rule?.RuleName || `规则${rule?.RuleId}` }}</span>
                        <span class="rule-count">{{ rule?.DropCount }}次</span>
                      </div>
                    </div>
                  </div>
                </n-popover>
              </div>
            </div>
          </n-gi>
        </n-grid>
      </n-card>

      <!-- 直接掉落规则 -->
      <n-collapse class="drop-collapse">
        <!-- 直接掉落规则 -->
        <n-collapse-item
            v-if="dropInfo.DirectDropRules?.length"
            title="直接掉落规则"
            :name="'direct'"
        >
          <template #header-extra>
            <n-tag :bordered="false" type="info" size="small">
              {{ dropInfo.DirectDropRules.length }}条规则
            </n-tag>
          </template>

          <div class="rules-container">
            <div
                v-for="rule in dropInfo.DirectDropRules"
                :key="rule?.RuleId"
                class="rule-card"
            >
              <div class="rule-header">
                <div class="rule-title">
                  <n-badge type="success" :value="rule?.RuleId" />
                  <span class="rule-name">{{ rule?.RuleName || `掉落规则 ${rule?.RuleId}` }}</span>
                </div>
                <div class="rule-badges">
                  <n-tag v-if="rule?.IsGuarantee" type="warning" :bordered="false" size="small">保底</n-tag>
                  <n-tag type="info" :bordered="false" size="small">掉落{{ rule?.DropCount }}次</n-tag>
                </div>
              </div>

              <div class="rule-content">
                <div class="rule-groups">
                  <span class="groups-label">关联掉落组:</span>
                  <div class="group-tags">
                    <n-tag
                        v-for="groupId in rule?.DropGroups"
                        :key="groupId"
                        size="small"
                        :bordered="false"
                        type="primary"
                        class="group-tag"
                    >
                      Group {{ groupId }}
                    </n-tag>
                  </div>
                </div>
              </div>


            </div>
          </div>
        </n-collapse-item>

        <!-- 保底掉落规则 -->
        <n-collapse-item
            v-if="dropInfo.GuaranteeDropRules?.length"
            title="保底掉落规则"
            :name="'guarantee'"
        >
          <template #header-extra>
            <n-tag :bordered="false" type="warning" size="small">
              {{ dropInfo.GuaranteeDropRules.length }}条规则
            </n-tag>
          </template>

          <div class="rules-container">
            <div
                v-for="rule in dropInfo.GuaranteeDropRules"
                :key="rule?.RuleId"
                class="rule-card guarantee-rule"
            >
              <div class="rule-header">
                <div class="rule-title">
                  <n-badge type="warning" :value="rule?.RuleId" />
                  <span class="rule-name">{{ rule?.RuleName || `保底规则 ${rule?.RuleId}` }}</span>
                </div>
                <div class="rule-badges">
                  <n-tag type="error" :bordered="false" size="small">保底</n-tag>
                  <n-tag type="info" :bordered="false" size="small">掉落{{ rule?.DropCount }}次</n-tag>
                </div>
              </div>

              <div class="rule-content">
                <div class="rule-groups">
                  <span class="groups-label">关联掉落组:</span>
                  <div class="group-tags">
                    <n-tag
                        v-for="groupId in rule?.DropGroups"
                        :key="groupId"
                        size="small"
                        :bordered="false"
                        type="warning"
                        class="group-tag"
                    >
                      Group {{ groupId }}
                    </n-tag>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </n-collapse-item>

        <!-- 掉落组详情 -->
        <n-collapse-item
            v-if="dropInfo.DropGroups?.length"
            title="掉落组详情"
            :name="'groups'"
        >
          <template #header-extra>
            <n-tag :bordered="false" type="success" size="small">
              {{ dropInfo.DropGroups.length }}个掉落组
            </n-tag>
          </template>

          <n-collapse
              class="groups-collapse"
              :expanded-names="expandGroups"
              @update:expanded-names="expandGroups = $event as number[]"
          >
            <n-collapse-item
                v-for="group in dropInfo.DropGroups"
                :key="group?.GroupId"
                :title="`Group ${group?.GroupId} - ${group?.GroupName || '未命名'}`"
                :name="group?.GroupId"
            >
              <template #header-extra>
                <n-space :size="4">
                  <n-tag v-if="group?.Deduplicate" size="tiny" type="info" :bordered="false">去重</n-tag>
                  <n-tag size="tiny" type="primary" :bordered="false">权重{{ group?.Weight }}</n-tag>
                  <n-tag v-if="group?.WeightInc != -1" size="tiny" type="success" :bordered="false">+{{ group?.WeightInc }}</n-tag>
                </n-space>
              </template>

              <n-grid :cols="24" :x-gap="16" :y-gap="12">
                <!-- 组基本信息 -->
                <n-gi :span="24">
                  <n-descriptions label-placement="left" :column="3" size="small" bordered>
                    <n-descriptions-item label="组ID">
                      <n-tag :bordered="false" size="small">{{ group?.GroupId }}</n-tag>
                    </n-descriptions-item>
                    <n-descriptions-item label="组名称">
                      {{ group?.GroupName || '未命名' }}
                    </n-descriptions-item>
                    <n-descriptions-item label="权重">
                      <n-space :size="4">
                        <span>{{ group?.Weight }}</span>
                        <span v-if="group?.WeightInc != -1" class="weight-inc">+{{ group?.WeightInc }}/次</span>
                      </n-space>
                    </n-descriptions-item>
                    <n-descriptions-item label="去重">
                      <n-badge :type="group?.Deduplicate ? 'success' : 'default'" :value="formatBoolean(group?.Deduplicate)" />
                    </n-descriptions-item>
                    <n-descriptions-item label="有效日期">
                      <n-space :size="4" align="center">
                        <n-tag :bordered="false" :type="isValidDate(group?.ValidDate, group?.ExpireDate) ? 'success' : 'error'" size="small">
                          {{ formatDate(group?.ValidDate) }} - {{ formatDate(group?.ExpireDate) }}
                        </n-tag>
                      </n-space>
                    </n-descriptions-item>
                    <n-descriptions-item label="掉落项数量">
                      <n-tag :bordered="false" type="info" size="small">{{ group?.DropItems?.length || 0 }}</n-tag>
                    </n-descriptions-item>
                  </n-descriptions>
                </n-gi>

                <!-- 掉落项列表 -->
                <n-gi :span="24" v-if="group?.DropItems?.length">
                  <div class="drop-items-title">
                    <n-icon size="16" :component="GiftFilled" />
                    <span>掉落项列表</span>
                  </div>

                  <n-scrollbar x-scrollable class="items-scrollbar">
                    <div class="items-grid">
                      <div
                          v-for="item in group?.DropItems"
                          :key="item?.ItemId"
                          class="item-card"
                          :class="{ 'item-active': isValidDate(item?.ValidDate, item?.ExpireDate) }"
                      >
                        <div class="item-header">
                          <span class="item-id">#{{ item?.ItemId }}</span>
                          <span class="item-name">{{ item?.ItemName || `掉落项 ${item?.ItemId}` }}</span>
                        </div>

                        <div class="item-weights">
                          <n-popover trigger="hover" placement="top">
                            <template #trigger>
                              <div class="weight-badge">
                                <span class="weight-label">权重</span>
                                <span class="weight-value">{{ item?.Weight }}</span>
                                <span v-if="item?.WeightInc != -1" class="weight-inc">+{{ item?.WeightInc }}</span>
                              </div>
                            </template>
                            <span>基础权重: {{ item?.Weight }}{{ item?.WeightInc != -1 ? `, 每次递增 ${item?.WeightInc}` : '' }}</span>
                          </n-popover>
                        </div>

                        <div class="item-flags">
                          <n-tag v-if="item?.Deduplicate" size="tiny" type="info" :bordered="false">去重</n-tag>
                          <n-tag v-if="item?.CheckExist" size="tiny" type="warning" :bordered="false">检查存在</n-tag>
                          <n-tag v-if="item?.ExcludeExist" size="tiny" type="error" :bordered="false">排除存在</n-tag>
                          <n-tag v-if="item?.MustHave" size="tiny" type="success" :bordered="false">必须拥有</n-tag>
                        </div>

                        <div class="item-configs">
                          <div
                              v-for="(config, idx) in item?.ItemConfigs"
                              :key="idx"
                              class="config-item"
                              :class="{ 'config-hero': config?.IsHero }"
                          >
                            <n-popover trigger="hover" placement="right">
                              <template #trigger>
                                <div class="config-badge">
                                  <span class="config-type">{{ config?.IsHero ? '👤' : '📦' }}</span>
                                  <span class="config-id">{{ config?.IsHero ? config?.HeroId : config?.ItemId }}</span>
                                  <span class="config-count">x{{ config?.Count }}</span>
                                </div>
                              </template>
                              <div class="config-popover">
                                <div>{{ config?.IsHero ? '武将' : '物品' }}ID: {{ config?.IsHero ? config?.HeroId : config?.ItemId }}</div>
                                <div>数量: {{ config?.Count }}</div>
                              </div>
                            </n-popover>
                          </div>
                        </div>

                        <div class="item-footer">
                          <n-text depth="3" class="item-dates">
                            <n-icon size="12" :component="CalendarFilled" />
                            {{ formatDate(item?.ValidDate) }} ~ {{ formatDate(item?.ExpireDate) }}
                          </n-text>
                          <n-tag v-if="item?.ReplaceGroup" size="tiny" type="primary" :bordered="false">
                            替换组 {{ item?.ReplaceGroup }}
                          </n-tag>
                        </div>
                      </div>
                    </div>
                  </n-scrollbar>
                </n-gi>
              </n-grid>
            </n-collapse-item>
          </n-collapse>
        </n-collapse-item>
      </n-collapse>
    </template>
  </div>
</template>

<style scoped>
.drop-display-dark {
  width: 100%;
}

/* 概览卡片 */
.drop-overview-card {
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid rgba(255, 255, 255, 0.05);
  margin-bottom: 16px;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #e5e7eb;
  font-weight: 500;
}

/* 统计卡片 */
.stat-card {
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 8px;
  padding: 12px;
  transition: all 0.2s ease;
}

.stat-card:hover {
  background: rgba(255, 255, 255, 0.05);
  transform: translateY(-2px);
}

.stat-label {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #9ca3af;
  font-size: 13px;
  margin-bottom: 8px;
}

.stat-value {
  font-size: 28px;
  font-weight: 600;
  color: #f3f4f6;
  margin-bottom: 4px;
}

.stat-trend {
  font-size: 11px;
  color: #6b7280;
}

/* 类型分布 */
.type-distribution {
  padding: 8px 0;
}

.distribution-title {
  font-size: 13px;
  color: #9ca3af;
  margin-bottom: 12px;
  display: block;
}

.type-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.type-tag-wrapper {
  cursor: pointer;
}

.type-tag {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  border-radius: 16px;
  font-size: 12px;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  transition: all 0.2s ease;
}

.type-tag:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.3);
}

.type-tag .type-name {
  color: #e5e7eb;
}

.type-tag .type-count {
  background: rgba(255, 255, 255, 0.1);
  padding: 2px 6px;
  border-radius: 12px;
  font-size: 10px;
  color: #9ca3af;
}

.type-normal .type-count { background: rgba(16, 185, 129, 0.2); color: #10b981; }
.type-guarantee .type-count { background: rgba(245, 158, 11, 0.2); color: #f59e0b; }
.type-special .type-count { background: rgba(139, 92, 246, 0.2); color: #8b5cf6; }

.type-popover {
  min-width: 200px;
}

.popover-header {
  font-weight: 600;
  color: #f3f4f6;
  margin-bottom: 8px;
  padding-bottom: 4px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.popover-rule {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 4px 0;
  font-size: 12px;
}

.popover-rule .rule-name {
  color: #e5e7eb;
}

.popover-rule .rule-count {
  color: #9ca3af;
  font-size: 11px;
}

/* 规则卡片 */
.rules-container {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.rule-card {
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 8px;
  padding: 12px;
  transition: all 0.2s ease;
}

.rule-card:hover {
  background: rgba(255, 255, 255, 0.03);
  border-color: rgba(255, 255, 255, 0.1);
}

.guarantee-rule {
  border-left: 3px solid #f59e0b;
}

.rule-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.rule-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.rule-name {
  font-size: 14px;
  font-weight: 500;
  color: #e5e7eb;
}

.rule-badges {
  display: flex;
  gap: 6px;
}

.rule-groups {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.groups-label {
  font-size: 12px;
  color: #9ca3af;
}

.group-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.group-tag {
  cursor: pointer;
  transition: all 0.2s ease;
}

.group-tag:hover {
  transform: translateY(-1px);
}

/* 掉落项网格 */
.drop-items-title {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #e5e7eb;
  font-size: 14px;
  margin: 16px 0 12px;
}

.items-scrollbar {
  max-height: 500px;
}

.items-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 12px;
  padding: 4px 0;
}

.item-card {
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 8px;
  padding: 12px;
  transition: all 0.2s ease;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.item-card:hover {
  background: rgba(255, 255, 255, 0.04);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}

.item-active {
  border-left: 3px solid #10b981;
}

.item-header {
  display: flex;
  align-items: center;
  gap: 8px;
}

.item-id {
  font-size: 11px;
  color: #6b7280;
  background: rgba(255, 255, 255, 0.05);
  padding: 2px 6px;
  border-radius: 4px;
}

.item-name {
  font-size: 13px;
  font-weight: 500;
  color: #e5e7eb;
}

.item-weights {
  display: flex;
  align-items: center;
  gap: 8px;
}

.weight-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: rgba(255, 255, 255, 0.05);
  padding: 2px 8px;
  border-radius: 12px;
  font-size: 11px;
  cursor: help;
}

.weight-label {
  color: #9ca3af;
}

.weight-value {
  color: #f59e0b;
  font-weight: 500;
}

.weight-inc {
  color: #10b981;
  font-size: 10px;
}

.item-flags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.item-configs {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin: 4px 0;
}

.config-item {
  cursor: pointer;
}

.config-badge {
  display: flex;
  align-items: center;
  gap: 4px;
  background: rgba(255, 255, 255, 0.05);
  padding: 2px 8px;
  border-radius: 12px;
  font-size: 11px;
  border: 1px solid transparent;
  transition: all 0.2s ease;
}

.config-badge:hover {
  background: rgba(255, 255, 255, 0.1);
  border-color: rgba(255, 255, 255, 0.2);
}

.config-hero .config-badge {
  background: rgba(139, 92, 246, 0.15);
  color: #8b5cf6;
}

.config-type {
  font-size: 10px;
}

.config-id {
  font-weight: 500;
}

.config-count {
  color: #9ca3af;
  font-size: 10px;
}

.config-popover {
  font-size: 12px;
  padding: 4px 0;
}

.item-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 4px;
  padding-top: 4px;
  border-top: 1px dashed rgba(255, 255, 255, 0.05);
}

.item-dates {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 10px;
  color: #6b7280;
}

/* 暗黑主题适配 */
:deep(.n-collapse) {
  --n-title-text-color: #e5e7eb;
  --n-title-text-color-hover: #f3f4f6;
  --n-arrow-color: #9ca3af;
  --n-border-color: rgba(255, 255, 255, 0.1);
  background: transparent;
}

:deep(.n-collapse-item) {
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid rgba(255, 255, 255, 0.05);
  margin-bottom: 8px;
  border-radius: 8px;
}

:deep(.n-collapse-item__header) {
  padding: 12px 16px;
  border-radius: 8px;
}

:deep(.n-collapse-item__content) {
  padding: 16px;
  background: rgba(0, 0, 0, 0.2);
}

:deep(.n-descriptions) {
  --n-th-color: rgba(255, 255, 255, 0.05);
  --n-td-color: transparent;
  --n-th-text-color: #9ca3af;
  --n-td-text-color: #e5e7eb;
  --n-border-color: rgba(255, 255, 255, 0.1);
}

:deep(.n-tag) {
  --n-color: rgba(255, 255, 255, 0.05);
  --n-text-color: #e5e7eb;
  --n-border: 1px solid rgba(255, 255, 255, 0.1);
}

:deep(.n-badge) {
  --n-color: #4f46e5;
}

:deep(.n-popover) {
  --n-color: #1e1e1e;
  --n-text-color: #e5e7eb;
  border: 1px solid rgba(255, 255, 255, 0.1);
}
</style>
