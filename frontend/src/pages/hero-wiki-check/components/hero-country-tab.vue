/**
 * 武将国家信息标签页
 *
 * 展示国家基本信息和图标资源
 */
<script setup lang="ts">
import {HeroCompleteData} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/herowiki_def/models";
import {getCountryStyle} from "./hero-panel-utils";

defineProps<{
  heroWiki: HeroCompleteData
}>()
</script>

<template>
  <n-grid :cols="24" :x-gap="16">
    <n-gi :span="24">
      <n-card :bordered="false">
        <template #header>
          <div class="section-title">
            <span :style="{ color: getCountryStyle(heroWiki.Country?.Name ?? '').color }">
              {{ heroWiki.Country?.Name }}
            </span>
          </div>
        </template>

        <n-descriptions label-placement="left" :column="2" bordered>
          <n-descriptions-item label="势力王名">
            {{ heroWiki.Country?.KingName || '未知' }}
          </n-descriptions-item>
          <n-descriptions-item label="个人信息旗帜图路径">
            {{ heroWiki.Country?.PlayerInfoFlagImagePath || '未知' }}
          </n-descriptions-item>
          <n-descriptions-item label="个人信息旗帜字路径">
            {{ heroWiki.Country?.PlayerInfoFlagTextPath || '未知' }}
          </n-descriptions-item>
          <n-descriptions-item label="旗帜是否为白底">
            {{ heroWiki.Country?.IsWhiteBg || '未知' }}
          </n-descriptions-item>
          <n-descriptions-item label="势力可选开关">
            {{ heroWiki.Country?.IsOpen || '未知' }}
          </n-descriptions-item>
          <n-descriptions-item label="势力名字选中美术字路径">
            {{ heroWiki.Country?.NameArtSelectedFontPath || '未知' }}
          </n-descriptions-item>
          <n-descriptions-item label="势力名字未选中美术字路径">
            {{ heroWiki.Country?.NameArtUnselectedFontPath || '未知' }}
          </n-descriptions-item>
          <n-descriptions-item label="排序优先级">
            {{ heroWiki.Country?.SortPriority || '默认' }}
          </n-descriptions-item>
          <n-descriptions-item label="状态">
            <n-badge :type="heroWiki.Country?.IsOpen ? 'success' : 'error'"
                     :value="heroWiki.Country?.IsOpen ? '已开放' : '未开放'"/>
          </n-descriptions-item>
          <n-descriptions-item label="旗帜背景">
            {{ heroWiki.Country?.IsWhiteBg ? '白色' : '彩色' }}
          </n-descriptions-item>
        </n-descriptions>
      </n-card>
    </n-gi>

    <n-gi :span="24">
      <n-card title="图标资源" size="small" :bordered="false">
        <n-descriptions label-placement="left" :column="3">
          <n-descriptions-item label="公会图标">
            {{ heroWiki.Country?.GuildIcon || '无' }}
          </n-descriptions-item>
          <n-descriptions-item label="势力小图标(顶)">
            {{ heroWiki.Country?.ShiLiSmallTopIcon || '无' }}
          </n-descriptions-item>
          <n-descriptions-item label="势力小图标(底)">
            {{ heroWiki.Country?.ShiLiSmallBottomIcon || '无' }}
          </n-descriptions-item>
          <n-descriptions-item label="座位旗帜图路径">
            {{ heroWiki.Country?.SeatFlagImagePath || '无' }}
          </n-descriptions-item>
        </n-descriptions>
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
