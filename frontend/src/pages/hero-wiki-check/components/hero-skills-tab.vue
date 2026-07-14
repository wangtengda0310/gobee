/**
 * 武将技能标签页
 *
 * 展示技能列表，每个技能含基础信息、文本、标签、台词、熔炼、关联技能、Buff
 */
<script setup lang="ts">
import {HeroCompleteData} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/herowiki_def/models";
import {formatArray} from "./hero-panel-utils";
import BuffDisplay from "./buff-display.vue";

defineProps<{
  heroWiki: HeroCompleteData
}>()
</script>

<template>
  <n-collapse>
    <n-collapse-item
        v-for="(skill, index) in heroWiki.Skills"
        :key="index"
        :title="`${index + 1}. ${skill?.Basic?.SkillName || skill?.UI?.SkillName || '未知技能'}`"
        :name="index"
    >
      <template #header-extra>
        <n-space>
          <n-tag v-if="skill?.Basic?.IsFromOther" size="small" type="warning">衍生</n-tag>
          <n-tag v-if="skill?.Basic?.IsFromAura" size="small" type="info">光环</n-tag>
          <n-tag v-if="skill?.Melt?.CanMelt" size="small" type="success">可熔炼</n-tag>
        </n-space>
      </template>

      <n-grid :cols="24" :x-gap="16">
        <!-- 技能基础信息 -->
        <n-gi :span="24">
          <n-card size="small" :bordered="false">
            <n-descriptions label-placement="left" :column="2" bordered>
              <n-descriptions-item label="技能ID">
                <n-tag :bordered="false">{{ skill?.Basic?.Id || skill?.UI?.Id }}</n-tag>
              </n-descriptions-item>
              <n-descriptions-item label="Enum ID">
                <n-tag :bordered="false" type="info">{{ skill?.Basic?.ESkillId }}</n-tag>
              </n-descriptions-item>
              <n-descriptions-item label="技能类型">
                <n-tag :bordered="false" type="primary">{{ skill?.Basic?.SkillType || '未知' }}</n-tag>
              </n-descriptions-item>
              <n-descriptions-item label="技能来源">
                {{ skill?.Basic?.SkillFromType || '默认' }}
              </n-descriptions-item>
              <n-descriptions-item label="转换卡牌类型">
                {{ skill?.Basic?.TransCard || '默认' }}
              </n-descriptions-item>
              <n-descriptions-item label="计数器公式">
                {{ skill?.Basic?.CounterFormula || '默认' }}
              </n-descriptions-item>
              <n-descriptions-item label="重置计数类型">
                {{ skill?.Basic?.ResetCounterType }}
              </n-descriptions-item>
              <n-descriptions-item label="重置使用次数类型">
                {{ skill?.Basic?.ResetTimesType }}
              </n-descriptions-item>
              <n-descriptions-item label="限定次数">
                {{ skill?.Basic?.SkillLimitTimes || '无限制' }}
              </n-descriptions-item>
              <n-descriptions-item label="游戏限制次数">
                {{ skill?.Basic?.TotalLimitTimes || '无限制' }}
              </n-descriptions-item>
              <n-descriptions-item label="触发时机">
                {{ skill?.Basic?.TriggerCondition || '无限制' }}
              </n-descriptions-item>
              <n-descriptions-item label="技能效果">
                {{ skill?.Basic?.SkillEffect || '' }}
              </n-descriptions-item>
              <n-descriptions-item label="空节点等待时间">
                {{ skill?.Basic?.EmptyWaitTime || '' }}
              </n-descriptions-item>
              <n-descriptions-item label="响应时显示点数与花色">
                {{ skill?.Basic?.ShowPointAndAttr || '' }}
              </n-descriptions-item>
              <n-descriptions-item label="互斥技能ID列表">
                {{ skill?.Basic?.MutexSkill || '' }}
              </n-descriptions-item>
              <n-descriptions-item label="战斗标记牌类型">
                {{ skill?.Basic?.BattleCardClass || '' }}
              </n-descriptions-item>
              <n-descriptions-item label="大模型区域判定参数">
                {{ skill?.Basic?.AIJudgeArea || '' }}
              </n-descriptions-item>
              <n-descriptions-item label="阵亡类型">
                {{ skill?.Basic?.DeadType || '' }}
              </n-descriptions-item>
              <n-descriptions-item label="初始属性">
                {{ skill?.Basic?.InitPro || '' }}
              </n-descriptions-item>
              <n-descriptions-item label="光环技能ID">
                {{ skill?.Basic?.MagicSkillID || '' }}
              </n-descriptions-item>
              <n-descriptions-item label="是否自动选中所有其他玩家">
                {{ skill?.Basic?.IsAutoSelAllOther || '' }}
              </n-descriptions-item>
              <n-descriptions-item label="是否禁止复制（刻写）">
                {{ skill?.Basic?.IsForbidCopy || '' }}
              </n-descriptions-item>
              <n-descriptions-item label="是否禁止转移">
                {{ skill?.Basic?.IsForbidTrans || '' }}
              </n-descriptions-item>
              <n-descriptions-item label="是否禁止废除">
                {{ skill?.Basic?.IsForbidDestroy || '' }}
              </n-descriptions-item>
            </n-descriptions>
          </n-card>
        </n-gi>

        <!-- 技能文本 -->
        <n-gi :span="24">
          <n-card size="small" title="技能描述" :bordered="false">
            <n-space vertical>
              <n-alert type="info" :bordered="false">
                {{ skill?.UI?.SkillText || '暂无技能描述' }}
              </n-alert>
              <n-text v-if="skill?.UI?.ShortSkillText" depth="3" style="font-style: italic;">
                简版：{{ skill?.UI.ShortSkillText }}
              </n-text>
            </n-space>
          </n-card>
          <n-card size="small" title="技能典故" :bordered="false">
            <n-space vertical>
              <n-alert type="warning" :bordered="false">
                {{ skill?.UI?.Allusion || '暂无技能典故' }}
              </n-alert>
              <n-text v-if="skill?.UI?.BattleSkillStep" depth="3" style="font-style: italic;">
                战斗内阶段显示：{{ skill?.UI.BattleSkillStep }}
              </n-text>
            </n-space>
          </n-card>
          <n-card size="small" title="UI台词音效" :bordered="false">
            <n-space vertical>
              <n-text v-if="skill?.UI?.PlayCardAudio" depth="2">
                出牌音效：{{ skill?.UI.PlayCardAudio }}
              </n-text>
              <n-text v-if="skill?.UI?.IdentityLine" depth="2">
                身份技能台词：{{ skill?.UI.IdentityLine }}
              </n-text>
              <n-text v-if="skill?.UI?.Audio" depth="2">
                打牌时触发技能的台词：{{ skill?.UI.Audio }}
              </n-text>
              <n-text v-if="skill?.UI?.SpecialAudio" depth="2">
                打牌时触发技能的台词：{{ skill?.UI.SpecialAudio }}
              </n-text>
            </n-space>
          </n-card>
        </n-gi>

        <!-- 技能标签 -->
        <n-gi :span="24" v-if="skill?.Tags?.length">
          <n-card size="small" title="技能标签" :bordered="false">
            <n-space>
              <n-popover v-for="tag in skill?.Tags" :key="tag?.SkillTag" trigger="hover">
                <template #trigger>
                  <n-tag :bordered="false" :color="{ color: tag?.TagColor }">
                    {{ tag?.TagName }}
                  </n-tag>
                </template>
                <span>{{ tag?.TagDes }}</span>
              </n-popover>
            </n-space>
          </n-card>
        </n-gi>

        <!-- 技能台词 -->
        <n-gi :span="24" v-if="skill?.Lines?.length">
          <n-card size="small" title="技能台词" :bordered="false">
            <n-timeline>
              <n-timeline-item v-for="(line, idx) in skill?.Lines" :key="idx" type="success">
                <template #header>
                  台词组 {{ line?.Id }} 皮肤Id: {{ line?.SkinId }}
                </template>
                <template v-if="line?.SkillFirstLine?.length">
                  第一段台词：{{ formatArray(line?.SkillFirstLine) }}
                </template>
                <template v-if="line?.SkillSecondLine?.length">
                  第二段台词：{{ formatArray(line?.SkillSecondLine) }}
                </template>
                <template v-if="line?.SkillThirdLine?.length">
                  第三段台词：{{ formatArray(line?.SkillThirdLine) }}
                </template>
                <template v-if="line?.SkillForthLine?.length">
                  第四段台词：{{ formatArray(line?.SkillForthLine) }}
                </template>
                <template v-if="line?.SpecialAudio">
                  <n-tag size="small" :bordered="false" type="warning">
                    特殊音效: {{ line?.SpecialAudio }}
                  </n-tag>
                </template>
              </n-timeline-item>
            </n-timeline>
          </n-card>
        </n-gi>

        <!-- 技能熔炼信息 -->
        <n-gi :span="24" v-if="skill?.Melt">
          <n-card size="small" title="熔炼信息" :bordered="false">
            <n-space>
              <n-statistic label="熔炼战力" :value="skill?.Melt.MeltPower || 0"/>
              <n-divider vertical/>
              <n-badge :value="skill?.Melt.CanMelt ? '可熔炼' : '不可熔炼'"
                       :type="skill?.Melt.CanMelt ? 'success' : 'error'"/>
            </n-space>
          </n-card>
        </n-gi>

        <!-- 关联技能 -->
        <n-gi :span="24" v-if="skill?.UI?.RelatedSkill?.length">
          <n-card size="small" title="关联技能" :bordered="false">
            <n-space>
              <n-tag v-for="rel in skill?.UI.RelatedSkill" :key="rel" type="primary">
                {{ rel }}
              </n-tag>
            </n-space>
          </n-card>
        </n-gi>

        <!-- Buff -->
        <n-gi :span="24" v-if="skill?.Basic?.Buff?.length">
          <n-card size="small" title="关联Buff" :bordered="false">
            <BuffDisplay
                :buffs="skill.Basic.Buff"
                :maxDisplay="5"
            />
          </n-card>
        </n-gi>
      </n-grid>
    </n-collapse-item>
  </n-collapse>
</template>
