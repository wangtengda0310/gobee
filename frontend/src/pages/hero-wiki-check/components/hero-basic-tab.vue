/**
 * 武将基础信息标签页
 *
 * 展示武将基础属性、战斗属性、特殊属性、使用限制
 */
<script setup lang="ts">
import {HeroCompleteData} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/herowiki_def/models";
import {ContainerFilled, HeartFilled, HeartOutlined, ToolFilled} from '@vicons/antd';

defineProps<{
  heroWiki: HeroCompleteData
}>()
</script>

<template>
  <n-grid :cols="24" :x-gap="16" :y-gap="16">
    <!-- 左侧：基础属性 -->
    <n-gi :span="12">
      <n-card title="基础属性" size="small" :bordered="false">
        <n-descriptions label-placement="left" :column="1">
          <n-descriptions-item label="武将ID">
            <n-text>{{ heroWiki.Basic?.Id }}</n-text>
          </n-descriptions-item>
          <n-descriptions-item label="性别">
            <n-text>{{ heroWiki.Basic?.Gender }}</n-text>
          </n-descriptions-item>
          <n-descriptions-item label="排除的身份枚举">
            <n-text>{{ heroWiki.Basic?.ExcludeIdentity }}</n-text>
          </n-descriptions-item>
          <n-descriptions-item label="不可以使用的房间模式">
            <n-text>{{ heroWiki.Basic?.NotUseModeType }}</n-text>
          </n-descriptions-item>
          <n-descriptions-item label="类型">
            <n-text>{{ heroWiki.Basic?.HeroType }}</n-text>
          </n-descriptions-item>
          <n-descriptions-item label="枚举">
            <n-tag :bordered="false" size="small">{{ heroWiki.Basic?.EHeroId }}</n-tag>
          </n-descriptions-item>
          <n-descriptions-item label="开放日期">
            <n-text :type="heroWiki.Basic?.IsOpen ? 'success' : 'warning'">
              {{ heroWiki.Basic?.OpenDate || '未指定' }}
            </n-text>
          </n-descriptions-item>
          <n-descriptions-item label="所属扩展包">
            <n-tag :bordered="false" type="info" size="small">
              {{ heroWiki.Basic?.BelongExpansionPack || '基础包' }}
            </n-tag>
          </n-descriptions-item>
          <n-descriptions-item label="是否为新增武将">
            <n-tag :bordered="false" :type="heroWiki.Basic?.IsNewHero ? 'success' : 'warning'" size="small">
              {{ heroWiki.Basic?.IsNewHero }}
            </n-tag>
          </n-descriptions-item>
          <n-descriptions-item label="是否招募产出">
            <n-tag :bordered="false" :type="heroWiki.Basic?.IsGacha ? 'success' : 'warning'" size="small">
              {{ heroWiki.Basic?.IsGacha }}
            </n-tag>
          </n-descriptions-item>
        </n-descriptions>
      </n-card>
    </n-gi>

    <!-- 右侧：战斗属性 -->
    <n-gi :span="12">
      <n-card title="战斗属性" size="small" :bordered="false">
        <n-grid :cols="2" :x-gap="12">
          <n-gi>
            <n-statistic label="体力上限" :tabular="true">
              <template #prefix>
                <n-icon>
                  <heart-filled/>
                </n-icon>
              </template>
              <n-number-animation :from="0" :to="heroWiki.Basic?.HpLimit"/>
            </n-statistic>
          </n-gi>
          <n-gi>
            <n-statistic label="手牌上限" :tabular="true">
              <template #prefix>
                <n-icon>
                  <container-filled/>
                </n-icon>
              </template>
              <n-number-animation :from="0" :to="heroWiki.Basic?.HandLimit"/>
            </n-statistic>
          </n-gi>
          <n-gi>
            <n-statistic label="装备上限" :tabular="true">
              <template #prefix>
                <n-icon>
                  <tool-filled/>
                </n-icon>
              </template>
              <n-number-animation :from="0" :to="heroWiki.Basic?.EquipLimit"/>
            </n-statistic>
          </n-gi>
          <n-gi>
            <n-statistic label="初始体力" :tabular="true">
              <template #prefix>
                <n-icon>
                  <heart-outlined/>
                </n-icon>
              </template>
              <n-number-animation :from="0" :to="heroWiki.Basic?.Point"/>
            </n-statistic>
          </n-gi>
        </n-grid>
      </n-card>
    </n-gi>

    <!-- 特殊属性 -->
    <n-gi :span="24">
      <n-card title="特殊属性" size="small" :bordered="false">
        <n-space>
          <n-tag v-if="heroWiki.Basic?.IsAlwaysZhuGong" type="warning" :bordered="false">
            常驻主公
          </n-tag>
          <n-tag v-if="heroWiki.Basic?.CanMelt" type="success" :bordered="false">
            可熔炼
          </n-tag>
          <n-tag v-if="heroWiki.Basic?.IsNewHero" type="info" :bordered="false">
            新英雄
          </n-tag>
          <n-tag v-if="heroWiki.Basic?.IsGacha" type="primary" :bordered="false">
            抽卡获取
          </n-tag>
          <n-tag v-if="heroWiki.Basic?.MeltName && heroWiki.Basic?.MeltName.length > 0"
                 type="success" :bordered="false">
            熔炼名：{{ heroWiki.Basic?.MeltName.join('、') }}
          </n-tag>
        </n-space>
      </n-card>
    </n-gi>

    <!-- 限制信息 -->
    <n-gi :span="24"
          v-if="heroWiki.Basic?.ExcludeIdentity?.length && (heroWiki.Basic?.ExcludeIdentity?.length > 0 || heroWiki.Basic?.NotUseModeType?.length > 0)">
      <n-card title="使用限制" size="small" :bordered="false">
        <n-space vertical>
          <n-space v-if="heroWiki.Basic?.ExcludeIdentity?.length > 0">
            <span class="label">禁用身份：</span>
            <n-tag v-for="id in heroWiki.Basic?.ExcludeIdentity" :key="id" size="small">
              身份{{ id }}
            </n-tag>
          </n-space>
          <n-space v-if="heroWiki.Basic?.NotUseModeType?.length > 0">
            <span class="label">禁用模式：</span>
            <n-tag v-for="mode in heroWiki.Basic?.NotUseModeType" :key="mode" size="small">
              模式{{ mode }}
            </n-tag>
          </n-space>
        </n-space>
      </n-card>
    </n-gi>
  </n-grid>
</template>

<style scoped>
.label {
  font-weight: bold;
  color: #666;
  min-width: 80px;
}
</style>
