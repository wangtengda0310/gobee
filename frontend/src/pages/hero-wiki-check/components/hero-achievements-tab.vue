/**
 * 武将成就标签页
 *
 * 展示成就列表，每个成就含局内成就、成就详情、完成条件
 */
<script setup lang="ts">
import {HeroCompleteData} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/herowiki_def/models";
import {formatArray, formatBoolean} from "./hero-panel-utils";

defineProps<{
  heroWiki: HeroCompleteData
}>()
</script>

<template>
  <n-collapse>
    <n-collapse-item
        v-for="(achievement, index) in heroWiki.Achievements"
        :key="index"
        :title="(()=>{
          if (achievement?.HeroAchieve?.Name) {
            return achievement?.HeroAchieve?.Name + '(Hero)'
          } else if (achievement?.AchieveDetail?.Name) {
            return achievement?.AchieveDetail?.Name + '(Achieve)'
          } else {
            return '未知成就'
          }
        })()"
        :name="index"
    >
      <n-grid :cols="24" :x-gap="16">
        <!-- 英雄成就 -->
        <n-gi :span="12" v-if="achievement?.HeroAchieve">
          <n-card title="局内成就" size="small" :bordered="false">
            <n-descriptions label-placement="left" :column="1">
              <n-descriptions-item label="成就ID">
                {{ achievement?.HeroAchieve.Id }}
              </n-descriptions-item>
              <n-descriptions-item label="是否复用">
                {{ formatBoolean(achievement?.HeroAchieve.IsMult) }}
              </n-descriptions-item>
              <n-descriptions-item label="房间模式">
                {{ formatArray(achievement?.HeroAchieve.Mode) }}
              </n-descriptions-item>
              <n-descriptions-item label="房间人数">
                {{ achievement?.HeroAchieve.MinPlayerNum }}
              </n-descriptions-item>
              <n-descriptions-item label="使用英雄">
                {{ achievement?.HeroAchieve.UseHero }}
              </n-descriptions-item>
              <n-descriptions-item label="身份类型">
                {{ achievement?.HeroAchieve.Class }}
              </n-descriptions-item>
              <n-descriptions-item label="阵营">
                {{ achievement?.HeroAchieve?.Camp }}
              </n-descriptions-item>
              <n-descriptions-item label="身份">
                {{ achievement?.HeroAchieve?.Identity }}
              </n-descriptions-item>
              <n-descriptions-item label="对应钩子">
                {{ achievement?.HeroAchieve?.Hooker }}
              </n-descriptions-item>
              <n-descriptions-item label="对应钩子">
                {{ achievement?.HeroAchieve?.HookerTarget }}
              </n-descriptions-item>
              <n-descriptions-item label="条件参数">
                {{ formatArray(achievement?.HeroAchieve.CondParam) }}
              </n-descriptions-item>
            </n-descriptions>
          </n-card>
        </n-gi>

        <!-- 成就详情 -->
        <n-gi :span="12" v-if="achievement?.AchieveDetail">
          <n-card title="成就详情" size="small" :bordered="false">
            <n-descriptions label-placement="left" :column="1">
              <n-descriptions-item label="成就ID">
                {{ achievement?.AchieveDetail.Id }}
              </n-descriptions-item>
              <n-descriptions-item label="是否隐藏">
                {{ formatBoolean(achievement?.AchieveDetail.IsHide) }}
              </n-descriptions-item>
              <n-descriptions-item label="成就描述">
                {{ achievement?.AchieveDetail.Des }}
              </n-descriptions-item>
              <n-descriptions-item label="完成条件">
                {{ achievement?.AchieveDetail.Condition }}
              </n-descriptions-item>
              <n-descriptions-item label="开放日期">
                {{ achievement?.AchieveDetail.OpenDate || '永久' }}
              </n-descriptions-item>
              <n-descriptions-item label="成就奖励">
                <n-space>
                  <n-tag v-for="reward in achievement?.AchieveDetail.Reward"
                         :key="reward" type="success" size="small">
                    {{ reward }}
                  </n-tag>
                </n-space>
              </n-descriptions-item>
            </n-descriptions>
          </n-card>
        </n-gi>

        <!-- 完成条件 -->
        <n-gi :span="24" v-if="achievement?.AchieveDetail?.ConditionInfo">
          <n-card title="完成条件" size="small" :bordered="false">
            <n-descriptions label-placement="left" :column="2" bordered>
              <n-descriptions-item label="条件描述">
                {{ achievement?.AchieveDetail.ConditionInfo.CondDes }}
              </n-descriptions-item>
              <n-descriptions-item label="完成条件">
                {{ achievement?.AchieveDetail.ConditionInfo.CompleteCond }}
              </n-descriptions-item>
              <n-descriptions-item label="条件参数">
                {{ formatArray(achievement?.AchieveDetail.ConditionInfo.CompleteCondParam) }}
              </n-descriptions-item>
              <n-descriptions-item label="跳转条件">
                {{ achievement?.AchieveDetail.ConditionInfo.JumpCond || '无' }}
              </n-descriptions-item>
              <n-descriptions-item label="跳转参数" :span="2">
                {{ formatArray(achievement?.AchieveDetail.ConditionInfo.JumpParm) }}
              </n-descriptions-item>
            </n-descriptions>
          </n-card>
        </n-gi>
      </n-grid>
    </n-collapse-item>
  </n-collapse>
  <n-empty v-if="!heroWiki.Achievements?.length" description="暂无成就数据"/>
</template>
