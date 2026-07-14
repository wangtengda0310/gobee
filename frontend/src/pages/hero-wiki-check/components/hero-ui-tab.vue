/**
 * 武将详细信息标签页
 *
 * 展示武将简介、专属卡牌、新手展示技能、胜率信息
 */
<script setup lang="ts">
import {HeroCompleteData} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/herowiki_def/models";
import {BookFilled} from '@vicons/antd';

defineProps<{
  heroWiki: HeroCompleteData
}>()
</script>

<template>
  <n-grid :cols="24" :x-gap="16">
    <n-gi :span="24">
      <n-card :bordered="false">
        <template #header>
          <div class="section-title">武将简介</div>
        </template>

        <n-space vertical :size="16">
          <!-- 短介绍 -->
          <n-alert v-if="heroWiki.UI?.ShortIntroduction" type="info" :bordered="false">
            <template #icon>
              <n-icon>
                <book-filled/>
              </n-icon>
            </template>
            <span style="font-style: italic;">"{{ heroWiki.UI.ShortIntroduction }}"</span>
          </n-alert>

          <!-- 长介绍 -->
          <n-card v-if="heroWiki.UI?.LongIntroduction" size="small" :bordered="true">
            <template #header>
              <n-text strong>详细介绍</n-text>
            </template>
            <n-text depth="2">{{ heroWiki.UI.LongIntroduction }}</n-text>
          </n-card>

          <!-- 其他描述信息 -->
          <n-descriptions label-placement="left" :column="2" bordered>
            <n-descriptions-item label="考据">
              {{ heroWiki.UI?.Evidence || '无' }}
            </n-descriptions-item>
            <n-descriptions-item label="评价">
              {{ heroWiki.UI?.Evaluation || '无' }}
            </n-descriptions-item>
            <n-descriptions-item label="文案">
              {{ heroWiki.UI?.CopyWriter || '无' }}
            </n-descriptions-item>
            <n-descriptions-item label="技能设计师">
              {{ heroWiki.UI?.SkillDesigner || '无' }}
            </n-descriptions-item>
            <n-descriptions-item label="获取方式">
              <n-tag :bordered="false" type="success">{{ heroWiki.UI?.GetWay || '未知' }}</n-tag>
            </n-descriptions-item>
            <n-descriptions-item label="武将定位">
              <n-tag :bordered="false" type="info">{{ heroWiki.UI?.Position || '未知' }}</n-tag>
            </n-descriptions-item>
          </n-descriptions>
        </n-space>
      </n-card>
    </n-gi>

    <!-- 专属卡牌 -->
    <n-gi :span="24" v-if="heroWiki.UI?.ExclusiveCard?.length">
      <n-card title="专属卡牌" size="small" :bordered="false">
        <n-space>
          <n-tag v-for="card in heroWiki.UI.ExclusiveCard" :key="card" type="warning" :bordered="false">
            卡牌 {{ card }}
          </n-tag>
        </n-space>
      </n-card>
    </n-gi>

    <!-- 新手展示技能标签 -->
    <n-gi :span="24" v-if="heroWiki.UI?.NewbieShowSkillTag?.length">
      <n-card title="新手展示技能" size="small" :bordered="false">
        <n-space>
          <n-tag v-for="tag in heroWiki.UI.NewbieShowSkillTag" :key="tag" type="info" :bordered="false">
            技能标签 {{ tag }}
          </n-tag>
        </n-space>
      </n-card>
    </n-gi>

    <!-- 胜率信息 -->
    <n-gi :span="24" v-if="heroWiki.UI">
      <n-card title="胜率信息" size="small" :bordered="false">
        <n-space>
          <n-statistic label="2v2胜率" :value="heroWiki.UI.WinningRateIn2v2" unit="%"/>
          <n-divider vertical/>
          <n-statistic label="显示优先级" :value="heroWiki.UI.WinRateShowPriority"/>
        </n-space>
      </n-card>
    </n-gi>
  </n-grid>
</template>

<style scoped>
.section-title {
  font-size: 16px;
  font-weight: bold;
  margin-bottom: 12px;
  padding-left: 8px;
  border-left: 4px solid #18a058;
}
</style>
