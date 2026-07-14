/**
 * 武将卡池标签页
 *
 * 展示掉落配置
 */
<script setup lang="ts">
import {HeroCompleteData} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/herowiki_def/models";
import DropDisplay from "./drop-display.vue";

defineProps<{
  heroWiki: HeroCompleteData
}>()
</script>

<template>
  <n-grid>
    <n-gi :span="24" v-if="heroWiki.DropInfo">
      <n-card title="掉落配置" size="small" :bordered="false">
        <template #header-extra>
          <n-tag :bordered="false" type="success" size="small">掉落系统</n-tag>
        </template>
        <DropDisplay :drop-info="heroWiki.DropInfo"/>
      </n-card>
    </n-gi>
  </n-grid>
</template>
