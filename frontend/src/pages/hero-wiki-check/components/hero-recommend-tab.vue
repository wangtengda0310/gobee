/**
 * 武将推荐布阵标签页
 *
 * 展示各级别推荐布阵信息
 */
<script setup lang="ts">
import {HeroCompleteData} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/herowiki_def/models";
import {formatArray} from "./hero-panel-utils";

defineProps<{
  heroWiki: HeroCompleteData
}>()
</script>

<template>
  <n-card v-if="heroWiki.RecommendBd" :bordered="false">
    <n-steps :current="6" :status="'process'">
      <n-step title="壹地(1-4)" :description="formatArray(heroWiki.RecommendBd.Level1, '暂无')"/>
      <n-step title="贰北(2-8)" :description="formatArray(heroWiki.RecommendBd.Level2, '暂无')"/>
      <n-step title="叁西(3-12)" :description="formatArray(heroWiki.RecommendBd.Level3, '暂无')"/>
      <n-step title="肆南(4-16)" :description="formatArray(heroWiki.RecommendBd.Level4, '暂无')"/>
      <n-step title="伍东(5-20)" :description="formatArray(heroWiki.RecommendBd.Level5, '暂无')"/>
      <n-step title="陆天(6-24)" :description="formatArray(heroWiki.RecommendBd.Level6, '暂无')"/>
    </n-steps>
  </n-card>
  <n-empty v-else description="暂无推荐布阵数据"/>
</template>
