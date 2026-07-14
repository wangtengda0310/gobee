/**
 * 武将机器人行为标签页
 *
 * 展示机器人行为列表，每个行为含行为、目标数量、目标类型等
 */
<script setup lang="ts">
import {HeroCompleteData} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/herowiki_def/models";
import {formatArray} from "./hero-panel-utils";

defineProps<{
  heroWiki: HeroCompleteData
}>()
</script>

<template>
  <n-collapse>
    <n-collapse-item
        v-for="(action, index) in heroWiki.RobotActions"
        :key="index"
        :title="`行为 ${action?.Id}`"
        :name="index"
    >
      <n-descriptions label-placement="left" :column="2" bordered>
        <n-descriptions-item label="行为">
          {{ formatArray(action?.Action) }}
        </n-descriptions-item>
        <n-descriptions-item label="目标数量">
          {{ formatArray(action?.TargetNum) }}
        </n-descriptions-item>
        <n-descriptions-item label="目标类型">
          {{ formatArray(action?.TargetType) }}
        </n-descriptions-item>
        <n-descriptions-item label="卡牌数量">
          {{ formatArray(action?.CardNum) }}
        </n-descriptions-item>
        <n-descriptions-item label="卡牌来源">
          {{ formatArray(action?.CardFromType) }}
        </n-descriptions-item>
        <n-descriptions-item label="转化技能">
          {{ formatArray(action?.TransCardSkill) }}
        </n-descriptions-item>
      </n-descriptions>
    </n-collapse-item>
  </n-collapse>
  <n-empty v-if="!heroWiki.RobotActions?.length" description="暂无机器人行为数据"/>
</template>
