/**
 * 活动面板组件
 *
 * 显示单个活动的详细信息，包括基础信息、抽奖配置、掉落规则、次数奖励等
 * 使用 Tabs 组织内容，支持条件渲染各 Tab
 */
<script setup lang="ts">
import {computed, provide, ref} from "vue";
import {useMessage} from "naive-ui";
import BadgeLabel from "@shared/components/badge-label/index.vue";
import {ActivityCompleteData} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/activitywiki_def/models.js";
import {RuleCoverageData} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/activity-wiki-check/models.js";
import {DrawSkinDiff} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/draw_skin/models.js";
import {DrawPetDiff} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/draw_pet/models.js";
import {formatArray, formatBoolean, formatItemArray} from "@shared/composables/use-format-utils";
import {renderTabWithBadge} from "@shared/composables/use-rule-badge";
import {GridOutline as TableIcon} from "@vicons/ionicons5";
import {openExcelBySheet} from "@shared/composables/use-open-excel";
import DrawSkinCard from "./DrawSkinCard.vue";
import DrawPetCard from "./DrawPetCard.vue";
import DropRuleCard from "./DropRuleCard.vue";
import ResourcePreview from "./ResourcePreview.vue"
import {inject} from "vue"
import type {useResourceCheck} from "../composables/use-resource-check"
const props = defineProps<{
  seq: number
  activityData: ActivityCompleteData
  excelDir: string
  ruleCoverage: RuleCoverageData | null
}>()

// 向子组件（BadgeLabel、DrawSkinCard）注入 ruleCoverage
provide('ruleCoverage', computed(() => props.ruleCoverage))

// 从父组件注入资源检查能力
const resourceCheck = inject<ReturnType<typeof useResourceCheck>>('resourceCheck')

const message = useMessage()
const activeTab = ref('basic');

// 页签名称 → Excel Sheet名 映射
const tabToSheetMap: Record<string, string> = {
  'basic': '活动表|Activity',
  'drawSkin': '皮肤抽奖|DrawSkin',
  'dropRule': '掉落规则表|DropRule',
  'timesRewards': '限时皮肤次数奖|LimitSkinTimesReward',
  'shop': '商店表|Shop',
  'shopGoods': '商品表|ShopGood',
  'heroSkinCollition': '英雄皮肤收藏|HeroSkinCollition',
  'itemHeroSkin': '武将皮肤展示表|ItemHeroSkin',
  'heroSkinItem': '英雄皮肤|HeroSkinItem',
  'heroSkinSpine': '英雄皮肤Spine|HeroSkinSpine',
  'drawPet': '结缘亭|DrawPet',
  'pet': '灵宠|Pet',
  'petAudio': '灵宠音效|PetAudio',
  'accumulatedRecharge': '累充奖励表|AccumulatedRechargeReward',
}

const handleOpenExcel = async () => {
  const sheetName = tabToSheetMap[activeTab.value]
  if (!sheetName) {
    return
  }
  await openExcelBySheet(message, sheetName, props.excelDir)
}

// 构建三期DrawSkin展示数据
interface DrawSkinPeriod {
  title: string
  drawSkin: DrawSkinDiff
  highlight: boolean
  periodLabel: string
}

// 构建三期DrawPet展示数据
interface DrawPetPeriod {
  title: string
  drawPet: DrawPetDiff
  highlight: boolean
  periodLabel: string
}

const getDrawSkinPeriods = computed(() => {
  const periods: DrawSkinPeriod[] = []
  if (props.activityData.PrevDrawSkin) {
    periods.push({
      title: '上一期 DrawSkin',
      drawSkin: props.activityData.PrevDrawSkin,
      highlight: false,
      periodLabel: '上一期'
    })
  }
  if (props.activityData.DrawSkin) {
    periods.push({
      title: '当前期 DrawSkin (活动关联)',
      drawSkin: props.activityData.DrawSkin,
      highlight: true,
      periodLabel: '当前关联'
    })
  }
  if (props.activityData.NextDrawSkin) {
    periods.push({
      title: '下一期 DrawSkin',
      drawSkin: props.activityData.NextDrawSkin,
      highlight: false,
      periodLabel: '下一期'
    })
  }
  return periods
})

const getDrawPetPeriods = computed(() => {
  const periods: DrawPetPeriod[] = []
  if (props.activityData.PrevDrawPet) {
    periods.push({
      title: '上一期 DrawPet',
      drawPet: props.activityData.PrevDrawPet,
      highlight: false,
      periodLabel: '上一期'
    })
  }
  if (props.activityData.DrawPet) {
    periods.push({
      title: '当前期 DrawPet (活动关联)',
      drawPet: props.activityData.DrawPet,
      highlight: true,
      periodLabel: '当前关联'
    })
  }
  if (props.activityData.NextDrawPet) {
    periods.push({
      title: '下一期 DrawPet',
      drawPet: props.activityData.NextDrawPet,
      highlight: false,
      periodLabel: '下一期'
    })
  }
  return periods
})

// 页签角标渲染（BadgeLabel 组件通过 inject 获取 ruleCoverage，无需手动传递）
const tBadge = (sheet: string, label: string) => renderTabWithBadge(props.ruleCoverage, sheet, label)

// 获取活动类型样式
const getActivityTypeStyle = (type: string) => {
  const styles: Record<string, { color: string; bgColor: string }> = {
    'ActTypeTest': {color: '#d32f2f', bgColor: '#ffebee'},
    'ActTypeGacha': {color: '#7b1fa2', bgColor: '#f3e5f5'},
    'ActTypeLogin': {color: '#1976d2', bgColor: '#e3f2fd'},
    'ActTypeRecharge': {color: '#388e3c', bgColor: '#e8f5e9'},
    'ActTypeAccumulatedRecharge': {color: '#f57c00', bgColor: '#fff3e0'},
  };
  return styles[type] || {color: '#666', bgColor: '#f5f5f5'};
}
</script>

<template>
  <n-card
      :id="'ActivityId:' + props.activityData.Basic?.Id"
      :segmented="{
      content: true,
      footer: true
    }"
      hoverable
      class="activity-card"
  >
    <template #header>
      <div style="display: flex; align-items: center; gap: 8px;">
        <span style="font-size: 18px; font-weight: bold;">{{ seq + 1 }}. {{ props.activityData.Basic?.Name || '未命名活动' }}</span>
        <span :style="{
          padding: '2px 8px',
          borderRadius: '4px',
          fontSize: '14px',
          backgroundColor: getActivityTypeStyle(props.activityData.Basic?.ActivityType || '').bgColor,
          color: getActivityTypeStyle(props.activityData.Basic?.ActivityType || '').color,
          border: '1px solid ' + getActivityTypeStyle(props.activityData.Basic?.ActivityType || '').color,
        }">{{ props.activityData.Basic?.ActivityType || '未知类型' }}</span>
        <span style="padding: 2px 8px; border-radius: 4px; font-size: 14px; background-color: #e3f2fd; color: #1565c0;">{{ props.activityData.Basic?.EActivityId || '' }}</span>
      </div>
    </template>
    <template #header-extra>
      <div style="display: flex; align-items: center; gap: 8px;">
        <n-button
            size="small"
            type="primary"
            ghost
            @click="handleOpenExcel"
        >
          <template #icon>
            <n-icon><TableIcon /></n-icon>
          </template>
          打开Excel
        </n-button>
        <div class="activity-id-badge">ID: {{ props.activityData.Basic?.Id }}</div>
      </div>
    </template>

    <div class="activity-content">
      <!-- 使用Tabs组织内容 -->
      <n-tabs v-model:value="activeTab" type="line" animated>
        <!-- 基础信息页签 -->
        <n-tab-pane name="basic"
                    :tab="() => tBadge('活动表|Activity', '基础信息')">
          <n-grid :cols="24" :x-gap="16" :y-gap="16">
            <n-gi :span="12">
              <n-card title="活动基础属性" size="small" :bordered="false">
                <n-descriptions label-placement="left" :column="1">
                  <n-descriptions-item>
                    <template #label><BadgeLabel sheet="活动表|Activity" field="Id" label="活动ID" /></template>
                    <n-tag :bordered="false" size="small">{{ props.activityData.Basic?.Id }}</n-tag>
                  </n-descriptions-item>
                  <n-descriptions-item>
                    <template #label><BadgeLabel sheet="活动表|Activity" field="ActivityType" label="活动类型" /></template>
                    <n-tag :bordered="false" type="primary" size="small">{{ props.activityData.Basic?.ActivityType }}</n-tag>
                  </n-descriptions-item>
                  <n-descriptions-item>
                    <template #label><BadgeLabel sheet="活动表|Activity" field="ActivityPrefabType" label="活动预制类型" /></template>
                    <n-text>{{ props.activityData.Basic?.ActivityPrefabType || '无' }}</n-text>
                  </n-descriptions-item>
                  <n-descriptions-item>
                    <template #label><BadgeLabel sheet="活动表|Activity" field="BelongId" label="所属活动ID" /></template>
                    <n-text>{{ props.activityData.Basic?.BelongId || '无' }}</n-text>
                  </n-descriptions-item>
                  <n-descriptions-item>
                    <template #label><BadgeLabel sheet="活动表|Activity" field="ShowTab" label="是否显示页签" /></template>
                    <n-tag :bordered="false" :type="props.activityData.Basic?.ShowTab ? 'success' : 'warning'" size="small">
                      {{ formatBoolean(props.activityData.Basic?.ShowTab) }}
                    </n-tag>
                  </n-descriptions-item>
                  <n-descriptions-item>
                    <template #label><BadgeLabel sheet="活动表|Activity" field="Weight" label="排序权重" /></template>
                    <n-text>{{ props.activityData.Basic?.Weight ?? 0 }}</n-text>
                  </n-descriptions-item>
                </n-descriptions>
              </n-card>
            </n-gi>

            <n-gi :span="12">
              <n-card title="时间配置" size="small" :bordered="false">
                <n-descriptions label-placement="left" :column="1">
                  <n-descriptions-item>
                    <template #label><BadgeLabel sheet="活动表|Activity" field="StartTime" label="开始时间" /></template>
                    <n-text :type="props.activityData.Basic?.StartTime ? 'success' : 'warning'">
                      {{ props.activityData.Basic?.StartTime || '未配置' }}
                    </n-text>
                  </n-descriptions-item>
                  <n-descriptions-item>
                    <template #label><BadgeLabel sheet="活动表|Activity" field="EndTime" label="结束时间" /></template>
                    <n-text :type="props.activityData.Basic?.EndTime ? 'success' : 'warning'">
                      {{ props.activityData.Basic?.EndTime || '未配置' }}
                    </n-text>
                  </n-descriptions-item>
                  <n-descriptions-item>
                    <template #label><BadgeLabel sheet="活动表|Activity" field="TimeType" label="时间类型" /></template>
                    <n-text>{{ props.activityData.Basic?.TimeType ?? '无' }}</n-text>
                  </n-descriptions-item>
                  <n-descriptions-item>
                    <template #label><BadgeLabel sheet="活动表|Activity" field="RewardTime" label="奖励开始时间" /></template>
                    <n-text :type="props.activityData.Basic?.RewardTime ? 'success' : 'warning'">
                      {{ props.activityData.Basic?.RewardTime || '未配置' }}
                    </n-text>
                  </n-descriptions-item>
                  <n-descriptions-item>
                    <template #label><BadgeLabel sheet="活动表|Activity" field="RewardEndTime" label="奖励结束时间" /></template>
                    <n-text :type="props.activityData.Basic?.RewardEndTime ? 'success' : 'warning'">
                      {{ props.activityData.Basic?.RewardEndTime || '未配置' }}
                    </n-text>
                  </n-descriptions-item>
                </n-descriptions>
              </n-card>
            </n-gi>

            <n-gi :span="24" v-if="props.activityData.Basic?.CustomParma && props.activityData.Basic?.CustomParma.length > 0">
              <n-card title="自定义参数" size="small" :bordered="false">
                <n-descriptions label-placement="left" :column="2" bordered>
                  <n-descriptions-item>
                    <template #label><BadgeLabel sheet="活动表|Activity" field="CustomParma" label="CustomParma" /></template>
                    {{ formatArray(props.activityData.Basic?.CustomParma) }}
                  </n-descriptions-item>
                </n-descriptions>
              </n-card>
            </n-gi>

            <n-gi :span="24" v-if="props.activityData.Basic?.CustomParma2">
              <n-card title="自定义参数2" size="small" :bordered="false">
                <n-descriptions label-placement="left" :column="2" bordered>
                  <n-descriptions-item>
                    <template #label><BadgeLabel sheet="活动表|Activity" field="CustomParma2" label="CustomParma2" /></template>
                    {{ props.activityData.Basic?.CustomParma2 }}
                  </n-descriptions-item>
                </n-descriptions>
              </n-card>
            </n-gi>

            <!-- 关联关系链可视化 -->
            <n-gi :span="24">
              <n-card title="关联关系链" size="small" :bordered="false">
                <n-steps :current="999" size="small">
                  <n-step title="活动表|Activity" :description="`Id = ${props.activityData.Basic?.Id}`" />
                  <n-step v-if="props.activityData.DrawSkin" title="皮肤抽奖|DrawSkin"
                          :description="`Id = ${props.activityData.DrawSkin?.Id}`" />
                  <n-step v-if="props.activityData.DrawPet" title="结缘亭|DrawPet"
                          :description="`Id = ${props.activityData.DrawPet?.Id}`" />
                  <n-step v-if="props.activityData.DropRule" title="掉落规则|DropRule"
                          :description="`Id = ${props.activityData.DropRule?.Id}`" />
                  <n-step v-if="props.activityData.DropGroups?.length > 0" title="掉落组|DropGroup"
                          :description="`共 ${props.activityData.DropGroups?.length} 组`" />
                  <n-step v-if="props.activityData.DropItems?.length > 0" title="掉落项|DropItem"
                          :description="`共 ${props.activityData.DropItems?.length} 项`" />
                  <n-step v-if="props.activityData.Shop" title="商店|Shop"
                          :description="`ShopType = ${props.activityData.Shop?.ShopType}`" />
                  <n-step v-if="props.activityData.TimesRewards?.length > 0" title="次数奖励|LimitSkinTimesReward"
                          :description="`共 ${props.activityData.TimesRewards?.length} 条`" />
                  <n-step v-if="props.activityData.AccumulatedRecharges?.length > 0" title="累充奖励|AccumulatedRechargeReward"
                          :description="`共 ${props.activityData.AccumulatedRecharges?.length} 档`" />
                </n-steps>
              </n-card>
            </n-gi>
          </n-grid>
        </n-tab-pane>

        <!-- 抽奖配置页签 -->
        <n-tab-pane
            name="drawSkin"
            v-if="props.activityData.DrawSkin"
            :tab="() => tBadge('皮肤抽奖|DrawSkin', '抽奖配置')"
        >
          <n-grid :cols="24" :x-gap="16" :y-gap="16">
            <n-gi :span="24">
              <n-alert type="info" :show-icon="false" size="small">
                <template #header>
                  <n-text strong>关联说明</n-text>
                </template>
                <n-text code size="small">
                  Activity.CustomParma[0] → DrawSkin.Id = {{ props.activityData.DrawSkin?.Id }}
                </n-text>
              </n-alert>
            </n-gi>
            <n-gi :span="24" v-for="period in getDrawSkinPeriods" :key="period.title">
              <DrawSkinCard
                :title="period.title"
                :draw-skin="period.drawSkin"
                :highlight="period.highlight"
                :period-label="period.periodLabel"
              />
            </n-gi>
          </n-grid>
        </n-tab-pane>

        <!-- 掉落规则页签 -->
        <n-tab-pane
            v-if="props.activityData.DropRule || props.activityData.TenDropRule"
            name="dropRule"
            :tab="() => tBadge('掉落规则表|DropRule', '掉落规则')"
        >
          <n-grid :cols="24" :x-gap="16" :y-gap="16">
            <n-gi :span="24">
              <n-alert type="info" :show-icon="false" size="small">
                <template #header><n-text strong>关联说明</n-text></template>
                <template v-if="props.activityData.DrawPet">
                  <div style="margin-bottom: 4px;"><n-text code size="small">DrawPet.OnceDropRule({{ props.activityData.DrawPet?.OnceDropRule || '无' }}) → DropRule.Id = {{ props.activityData.DropRule?.Id || '无' }}</n-text></div>
                  <div><n-text code size="small">DrawPet.TenDropRule({{ props.activityData.DrawPet?.TenDropRule || '无' }}) → DropRule.Id = {{ props.activityData.TenDropRule?.Id || '无' }}</n-text></div>
                </template>
                <template v-else>
                  <div style="margin-bottom: 4px;"><n-text code size="small">DrawSkin.OnceDropRule({{ props.activityData.DrawSkin?.OnceDropRule || '无' }}) → DropRule.Id = {{ props.activityData.DropRule?.Id || '无' }}</n-text></div>
                  <div><n-text code size="small">DrawSkin.TenDropRule({{ props.activityData.DrawSkin?.TenDropRule || '无' }}) → DropRule.Id = {{ props.activityData.TenDropRule?.Id || '无' }}</n-text></div>
                </template>
              </n-alert>
            </n-gi>
            <!-- 单抽掉落规则 -->
            <n-gi :span="24" v-if="props.activityData.DropRule">
              <DropRuleCard
                title="单抽掉落规则 (OnceDropRule)"
                card-class="drawpet-card--current"
                :rule="props.activityData.DropRule"
                :groups="props.activityData.DropGroups || []"
                :items="props.activityData.DropItems || []"
              />
            </n-gi>
            <!-- 十连掉落规则 -->
            <n-gi :span="24" v-if="props.activityData.TenDropRule">
              <DropRuleCard
                title="十连掉落规则 (TenDropRule)"
                :rule="props.activityData.TenDropRule"
                :groups="props.activityData.TenDropGroups || []"
                :items="props.activityData.TenDropItems || []"
              />
            </n-gi>
          </n-grid>
        </n-tab-pane>

        <!-- 次数奖励页签 -->
        <n-tab-pane
            name="timesRewards"
            v-if="props.activityData.TimesRewards && props.activityData.TimesRewards.length > 0"
            :tab="() => tBadge('限时皮肤次数奖|LimitSkinTimesReward', '次数奖励')"
        >
          <n-grid :cols="24" :x-gap="16" :y-gap="16">
            <n-gi :span="24">
              <n-alert type="info" :show-icon="false" size="small">
                <template #header><n-text strong>关联说明</n-text></template>
                <n-text code size="small">LimitSkinTimesReward.ActIdStr == Activity.EActivityId = {{ props.activityData.Basic?.EActivityId }}</n-text>
              </n-alert>
            </n-gi>
            <n-gi :span="24">
              <n-card title="累计抽卡奖励" size="small" :bordered="false">
                <n-table :bordered="true" :single-line="false" size="small">
                  <thead><tr><th>奖励ID</th><th>活动ID(数字)</th><th>活动ID(枚举)</th><th>抽卡次数</th><th>奖励道具</th></tr></thead>
                  <tbody>
                  <tr v-for="(reward, idx) in props.activityData.TimesRewards" :key="reward?.Id || idx">
                    <template v-if="reward">
                      <td>{{ reward.Id }}</td>
                      <td>{{ reward.ActId }}</td>
                      <td><n-tag size="small" :bordered="false">{{ reward.ActIdStr || '-' }}</n-tag></td>
                      <td><n-tag size="small" type="success" :bordered="false">{{ reward.DrawTimes || 0 }} 次</n-tag></td>
                      <td>
                        <n-tag v-for="(cfg, cidx) in reward.Reward" :key="cidx" size="small" :bordered="false">{{ cfg?.ItemId }}×{{ cfg?.Count }}</n-tag>
                        <span v-if="!reward.Reward || reward.Reward.length === 0">-</span>
                      </td>
                    </template>
                  </tr>
                  </tbody>
                </n-table>
              </n-card>
            </n-gi>
          </n-grid>
        </n-tab-pane>

        <!-- 商店页签 -->
        <n-tab-pane name="shop" v-if="props.activityData.Shop"
            :tab="() => tBadge('商店表|Shop', '商店配置')">
          <n-grid :cols="24" :x-gap="16" :y-gap="16">
            <n-gi :span="24">
              <n-alert type="info" :show-icon="false" size="small">
                <template #header><n-text strong>关联说明</n-text></template>
                <n-text code size="small">Shop.ShopType == "{{ props.activityData.Shop?.ShopType }}"</n-text>
              </n-alert>
            </n-gi>
            <n-gi :span="24">
              <n-card title="商店信息" size="small" :bordered="false">
                <n-descriptions label-placement="left" :column="2" bordered>
                  <n-descriptions-item>
                    <template #label><BadgeLabel sheet="商店表|Shop" field="ShopType" label="商店类型" /></template>
                    <n-tag :bordered="false" type="primary">{{ props.activityData.Shop?.ShopType }}</n-tag>
                  </n-descriptions-item>
                  <n-descriptions-item>
                    <template #label><BadgeLabel sheet="商店表|Shop" field="Name" label="商店名称" /></template>
                    <n-text strong>{{ props.activityData.Shop?.Name || '未命名' }}</n-text>
                  </n-descriptions-item>
                  <n-descriptions-item>
                    <template #label><BadgeLabel sheet="商店表|Shop" field="ShopName" label="显示名称" /></template>
                    <n-text>{{ props.activityData.Shop?.ShopName || '-' }}</n-text>
                  </n-descriptions-item>
                  <n-descriptions-item>
                    <template #label><BadgeLabel sheet="商店表|Shop" field="UseCurrency" label="使用货币" /></template>
                    <n-tag :bordered="false" type="info">{{ formatArray(props.activityData.Shop?.UseCurrency) }}</n-tag>
                  </n-descriptions-item>
                  <n-descriptions-item>
                    <template #label><BadgeLabel sheet="商店表|Shop" field="OpenTime" label="开启时间" /></template>
                    <n-text>{{ props.activityData.Shop?.OpenTime || '-' }}</n-text>
                  </n-descriptions-item>
                  <n-descriptions-item>
                    <template #label><BadgeLabel sheet="商店表|Shop" field="CloseTime" label="关闭时间" /></template>
                    <n-text>{{ props.activityData.Shop?.CloseTime || '-' }}</n-text>
                  </n-descriptions-item>
                  <n-descriptions-item>
                    <template #label><BadgeLabel sheet="商店表|Shop" field="IsLimitedShop" label="限时商店" /></template>
                    <n-tag :bordered="false" :type="props.activityData.Shop?.IsLimitedShop ? 'success' : 'default'">{{ formatBoolean(props.activityData.Shop?.IsLimitedShop) }}</n-tag>
                  </n-descriptions-item>
                  <n-descriptions-item>
                    <template #label><BadgeLabel sheet="商店表|Shop" field="IsDynamicsShop" label="动态商店" /></template>
                    <n-tag :bordered="false" :type="props.activityData.Shop?.IsDynamicsShop ? 'success' : 'default'">{{ formatBoolean(props.activityData.Shop?.IsDynamicsShop) }}</n-tag>
                  </n-descriptions-item>
                  <n-descriptions-item>
                    <template #label><BadgeLabel sheet="商店表|Shop" field="InMainShopOrder" label="主界面商店排序" /></template>
                    <n-text>{{ props.activityData.Shop?.InMainShopOrder ?? '-' }}</n-text>
                  </n-descriptions-item>
                </n-descriptions>
              </n-card>
            </n-gi>
          </n-grid>
        </n-tab-pane>

        <!-- 商品页签 -->
        <n-tab-pane name="shopGoods" v-if="props.activityData.ShopGoods && props.activityData.ShopGoods.length > 0"
            :tab="() => tBadge('商品表|ShopGood', '商品配置')">
          <n-grid :cols="24" :x-gap="16" :y-gap="16">
            <n-gi :span="24" v-for="(goods, idx) in props.activityData.ShopGoods" :key="goods?.Id || idx">
              <n-card :title="'商品: ' + (goods?.Name || '未命名') + ' (ID: ' + goods?.Id + ')'" size="small" :bordered="false">
                <n-descriptions label-placement="left" :column="2" bordered>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="商品表|ShopGood" field="Id" label="商品ID" /></template><n-tag :bordered="false">{{ goods?.Id }}</n-tag></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="商品表|ShopGood" field="Name" label="名称" /></template><n-text strong>{{ goods?.Name || '-' }}</n-text></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="商品表|ShopGood" field="Desc" label="描述" /></template><n-text>{{ goods?.Desc || '-' }}</n-text></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="商品表|ShopGood" field="ShopType" label="商店类型" /></template><n-tag :bordered="false" type="primary">{{ goods?.ShopType || '-' }}</n-tag></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="商品表|ShopGood" field="MainItem" label="主道具" /></template>
                    <n-tag v-for="(cfg, cidx) in goods?.Item" :key="cidx" size="small" :bordered="false">{{ cfg?.ItemId }}×{{ cfg?.Count }}</n-tag>
                    <span v-if="!goods?.Item || goods.Item.length === 0">-</span>
                  </n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="商品表|ShopGood" field="ExtraItems" label="额外道具" /></template>
                    <n-tag v-for="(cfg, cidx) in goods?.ExtraItem" :key="cidx" size="small" :bordered="false">{{ cfg?.ItemId }}×{{ cfg?.Count }}</n-tag>
                    <span v-if="!goods?.ExtraItem || goods.ExtraItem.length === 0">-</span>
                  </n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="商品表|ShopGood" field="CostId" label="货币ID" /></template><n-tag :bordered="false" type="info">{{ goods?.CostId || '-' }}</n-tag></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="商品表|ShopGood" field="Price" label="价格" /></template><n-text type="success" strong>{{ goods?.Price ?? 0 }}</n-text></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="商品表|ShopGood" field="OldPrice" label="原价" /></template><n-text>{{ goods?.OldPrice || '-' }}</n-text></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="商品表|ShopGood" field="Discount" label="折扣" /></template><n-text>{{ goods?.Discount || '-' }}</n-text></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="商品表|ShopGood" field="LimitType" label="限制类型" /></template><n-text>{{ goods?.LimitType ?? '-' }}</n-text></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="商品表|ShopGood" field="LimitCount" label="限购次数" /></template><n-text>{{ (goods?.LimitCount ?? 0) > 0 ? goods!.LimitCount + '次' : '不限' }}</n-text></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="商品表|ShopGood" field="MaxBuyCount" label="最大购买次数" /></template><n-text>{{ goods?.MaxBuyCount || '-' }}</n-text></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="商品表|ShopGood" field="Pos" label="位置" /></template><n-text>{{ goods?.Pos ?? '-' }}</n-text></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="商品表|ShopGood" field="IsNew" label="是否新品" /></template><n-tag :bordered="false" :type="goods?.IsNew ? 'success' : 'default'">{{ formatBoolean(goods?.IsNew) }}</n-tag></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="商品表|ShopGood" field="Countdown" label="倒计时" /></template><n-text>{{ goods?.Countdown || '-' }}</n-text></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="商品表|ShopGood" field="IsHomePage" label="是否在首页" /></template><n-tag :bordered="false" :type="goods?.IsHomePage ? 'success' : 'default'">{{ formatBoolean(goods?.IsHomePage) }}</n-tag></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="商品表|ShopGood" field="Icon" label="图标" /></template><ResourcePreview :value="goods?.Icon" :status="resourceCheck?.getStatus(goods?.Icon || '')" /></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="商品表|ShopGood" field="OnShelfTime" label="上架时间" /></template><n-text>{{ goods?.OnShelfTime || '-' }}</n-text></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="商品表|ShopGood" field="OffShelfTime" label="下架时间" /></template><n-text>{{ goods?.OffShelfTime || '-' }}</n-text></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="商品表|ShopGood" field="RechargeAndriod" label="安卓充值" /></template><n-text>{{ goods?.RechargeAndriod || '-' }}</n-text></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="商品表|ShopGood" field="RechargeIOS" label="iOS充值" /></template><n-text>{{ goods?.RechargeIOS || '-' }}</n-text></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="商品表|ShopGood" field="RechargeMulti" label="多平台充值" /></template><n-text>{{ goods?.RechargeMulti || '-' }}</n-text></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="商品表|ShopGood" field="RechargeMultiGroup" label="多平台充值组" /></template><n-text>{{ goods?.RechargeMultiGroup || '-' }}</n-text></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="商品表|ShopGood" field="RechargeMultiBeginTime" label="多平台充值开始时间" /></template><n-text>{{ goods?.RechargeMultiBeginTime || '-' }}</n-text></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="商品表|ShopGood" field="RechargeMultiEndTime" label="多平台充值结束时间" /></template><n-text>{{ goods?.RechargeMultiEndTime || '-' }}</n-text></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="商品表|ShopGood" field="PreSelOutShopId" label="预售出商店ID" /></template><n-text>{{ goods?.PreSelOutShopId || '-' }}</n-text></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="商品表|ShopGood" field="HideWhenSelOut" label="售出时隐藏" /></template><n-tag :bordered="false" :type="goods?.HideWhenSelOut ? 'warning' : 'default'">{{ formatBoolean(goods?.HideWhenSelOut) }}</n-tag></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="商品表|ShopGood" field="LimitZeroWhenHasItem" label="有道具时限购为0" /></template><n-tag :bordered="false" :type="goods?.LimitZeroWhenHasItem ? 'warning' : 'default'">{{ formatBoolean(goods?.LimitZeroWhenHasItem) }}</n-tag></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="商品表|ShopGood" field="IconDisplayType" label="图标显示类型" /></template><n-text>{{ goods?.IconDisplayType ?? '-' }}</n-text></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="商品表|ShopGood" field="ShowItemCount" label="显示道具数量" /></template><n-tag :bordered="false" :type="goods?.ShowItemCount ? 'success' : 'default'">{{ formatBoolean(goods?.ShowItemCount) }}</n-tag></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="商品表|ShopGood" field="ShowQuality" label="显示品质" /></template><n-tag :bordered="false" :type="goods?.ShowQuality ? 'success' : 'default'">{{ formatBoolean(goods?.ShowQuality) }}</n-tag></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="商品表|ShopGood" field="IconInBuyWindow" label="购买窗口图标" /></template><n-text>{{ goods?.IconInBuyWindow || '-' }}</n-text></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="商品表|ShopGood" field="RechargeMultiGoodsDes" label="多平台商品描述" /></template><n-text>{{ goods?.RechargeMultiGoodsDes || '-' }}</n-text></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="商品表|ShopGood" field="RewardID" label="奖励ID" /></template><n-text>{{ goods?.RewardID || '-' }}</n-text></n-descriptions-item>
                </n-descriptions>
              </n-card>
            </n-gi>
          </n-grid>
        </n-tab-pane>

        <!-- 皮肤收藏册页签 -->
        <n-tab-pane name="heroSkinCollition" v-if="props.activityData.HeroSkinCollition"
            :tab="() => tBadge('英雄皮肤收藏|HeroSkinCollition', '皮肤收藏')">
          <n-grid :cols="24" :x-gap="16" :y-gap="16">
            <n-gi :span="24">
              <n-card title="皮肤收藏册信息" size="small" :bordered="false">
                <n-descriptions label-placement="left" :column="2" bordered>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="英雄皮肤收藏|HeroSkinCollition" field="Type" label="收藏册类型" /></template><n-tag :bordered="false" type="primary">{{ props.activityData.HeroSkinCollition?.Type }}</n-tag></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="英雄皮肤收藏|HeroSkinCollition" field="Name" label="收藏册名称" /></template><n-text strong>{{ props.activityData.HeroSkinCollition?.Name || '未命名' }}</n-text></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="英雄皮肤收藏|HeroSkinCollition" field="NameImg" label="名称图片" /></template><ResourcePreview :value="props.activityData.HeroSkinCollition?.NameImg" :status="resourceCheck?.getStatus(props.activityData.HeroSkinCollition?.NameImg || '')" /></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="英雄皮肤收藏|HeroSkinCollition" field="NameBg" label="名称底图" /></template><ResourcePreview :value="props.activityData.HeroSkinCollition?.NameBg" :status="resourceCheck?.getStatus(props.activityData.HeroSkinCollition?.NameBg || '')" /></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="英雄皮肤收藏|HeroSkinCollition" field="Desc" label="诗词文案" /></template><n-text>{{ props.activityData.HeroSkinCollition?.Desc || '-' }}</n-text></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="英雄皮肤收藏|HeroSkinCollition" field="Weight" label="排序权重" /></template><n-text>{{ props.activityData.HeroSkinCollition?.Weight ?? 0 }}</n-text></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="英雄皮肤收藏|HeroSkinCollition" field="OpenDate" label="开放时间" /></template><n-text>{{ props.activityData.HeroSkinCollition?.OpenDate || '-' }}</n-text></n-descriptions-item>
                </n-descriptions>
              </n-card>
            </n-gi>
          </n-grid>
        </n-tab-pane>

        <!-- 武将皮肤展示页签 -->
        <n-tab-pane name="itemHeroSkin" v-if="props.activityData.ItemHeroSkin"
            :tab="() => tBadge('武将皮肤展示表|ItemHeroSkin', '皮肤展示')">
          <n-grid :cols="24" :x-gap="16" :y-gap="16">
            <n-gi :span="24">
              <n-card title="皮肤展示信息" size="small" :bordered="false">
                <n-descriptions label-placement="left" :column="2" bordered>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="武将皮肤展示表|ItemHeroSkin" field="SkinItemId" label="皮肤道具ID" /></template><n-tag :bordered="false">{{ props.activityData.ItemHeroSkin?.SkinItemId }}</n-tag></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="武将皮肤展示表|ItemHeroSkin" field="Path" label="路径" /></template><ResourcePreview :value="props.activityData.ItemHeroSkin?.Path" :status="resourceCheck?.getStatus(props.activityData.ItemHeroSkin?.Path || '')" /></n-descriptions-item>
                </n-descriptions>
              </n-card>
            </n-gi>
          </n-grid>
        </n-tab-pane>

        <!-- 英雄皮肤页签 -->
        <n-tab-pane name="heroSkinItem" v-if="props.activityData.HeroSkinItem"
            :tab="() => tBadge('英雄皮肤|HeroSkinItem', '皮肤信息')">
          <n-grid :cols="24" :x-gap="16" :y-gap="16">
            <n-gi :span="24">
              <n-card title="皮肤信息" size="small" :bordered="false">
                <n-descriptions label-placement="left" :column="2" bordered>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="英雄皮肤|HeroSkinItem" field="SkinItemId" label="皮肤道具ID" /></template><n-tag :bordered="false">{{ props.activityData.HeroSkinItem?.SkinItemId }}</n-tag></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="英雄皮肤|HeroSkinItem" field="HeroId" label="英雄ID" /></template><n-tag :bordered="false" type="primary">{{ props.activityData.HeroSkinItem?.HeroId }}</n-tag></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="英雄皮肤|HeroSkinItem" field="Name" label="皮肤名称" /></template><n-text strong>{{ props.activityData.HeroSkinItem?.Name || '未命名' }}</n-text></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="英雄皮肤|HeroSkinItem" field="SkinType" label="皮肤类型" /></template><n-text>{{ props.activityData.HeroSkinItem?.SkinType || '-' }}</n-text></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="英雄皮肤|HeroSkinItem" field="RailyType" label="品质类型" /></template><n-text>{{ props.activityData.HeroSkinItem?.RailyType || '-' }}</n-text></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="英雄皮肤|HeroSkinItem" field="GetWay" label="获取途径" /></template><n-text>{{ props.activityData.HeroSkinItem?.GetWay || '-' }}</n-text></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="英雄皮肤|HeroSkinItem" field="SkinPinYin" label="皮肤拼音" /></template><n-text>{{ props.activityData.HeroSkinItem?.SkinPinYin || '-' }}</n-text></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="英雄皮肤|HeroSkinItem" field="SeatSpecialImg" label="座位特殊图片" /></template><ResourcePreview :value="props.activityData.HeroSkinItem?.SeatSpecialImg" :status="resourceCheck?.getStatus(props.activityData.HeroSkinItem?.SeatSpecialImg || '')" /></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="英雄皮肤|HeroSkinItem" field="HeroUIExtraIcons" label="英雄UI额外图标" /></template>
                    <n-tag v-for="(icon, iidx) in props.activityData.HeroSkinItem?.HeroUIExtraIcons" :key="iidx" size="small" :bordered="false">{{ icon }}</n-tag>
                    <span v-if="!props.activityData.HeroSkinItem?.HeroUIExtraIcons || props.activityData.HeroSkinItem.HeroUIExtraIcons.length === 0">-</span>
                  </n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="英雄皮肤|HeroSkinItem" field="VoiceLines" label="语音Lines" /></template>
                    <n-tag v-for="(line, lidx) in props.activityData.HeroSkinItem?.Lines" :key="lidx" size="small" :bordered="false">{{ line }}</n-tag>
                    <span v-if="!props.activityData.HeroSkinItem?.Lines || props.activityData.HeroSkinItem.Lines.length === 0">-</span>
                  </n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="英雄皮肤|HeroSkinItem" field="AppearVoice" label="出场语音" /></template>
                    <n-tag v-for="(line, didx) in props.activityData.HeroSkinItem?.DebutLines" :key="didx" size="small" :bordered="false">{{ line }}</n-tag>
                    <span v-if="!props.activityData.HeroSkinItem?.DebutLines || props.activityData.HeroSkinItem.DebutLines.length === 0">-</span>
                  </n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="英雄皮肤|HeroSkinItem" field="KillVoice" label="击杀语音" /></template>
                    <n-tag v-for="(line, kidx) in props.activityData.HeroSkinItem?.KillLines" :key="kidx" size="small" :bordered="false">{{ line }}</n-tag>
                    <span v-if="!props.activityData.HeroSkinItem?.KillLines || props.activityData.HeroSkinItem.KillLines.length === 0">-</span>
                  </n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="英雄皮肤|HeroSkinItem" field="DeathVoice" label="死亡语音" /></template>
                    <n-tag v-for="(line, ddidx) in props.activityData.HeroSkinItem?.DeadLines" :key="ddidx" size="small" :bordered="false">{{ line }}</n-tag>
                    <span v-if="!props.activityData.HeroSkinItem?.DeadLines || props.activityData.HeroSkinItem.DeadLines.length === 0">-</span>
                  </n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="英雄皮肤|HeroSkinItem" field="HeroAudio" label="英雄音频" /></template>
                    <n-tag v-for="(audio, aidx) in props.activityData.HeroSkinItem?.HeroAudio" :key="aidx" size="small" :bordered="false">{{ audio }}</n-tag>
                    <span v-if="!props.activityData.HeroSkinItem?.HeroAudio || props.activityData.HeroSkinItem.HeroAudio.length === 0">-</span>
                  </n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="英雄皮肤|HeroSkinItem" field="LinesDubbed" label="语音配音状态" /></template><n-text>{{ props.activityData.HeroSkinItem?.LinesDubbed || '-' }}</n-text></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="英雄皮肤|HeroSkinItem" field="OriginalArtDesigner" label="原画设计师" /></template><n-text>{{ props.activityData.HeroSkinItem?.OriginalArtDesigner || '-' }}</n-text></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="英雄皮肤|HeroSkinItem" field="CollitionType" label="收藏类型" /></template><n-text>{{ props.activityData.HeroSkinItem?.CollitionType || '-' }}</n-text></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="英雄皮肤|HeroSkinItem" field="IsOpen" label="是否开放" /></template><n-tag :bordered="false" :type="props.activityData.HeroSkinItem?.IsOpen ? 'success' : 'warning'">{{ formatBoolean(props.activityData.HeroSkinItem?.IsOpen) }}</n-tag></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="英雄皮肤|HeroSkinItem" field="OpenDate" label="开放时间" /></template><n-text>{{ props.activityData.HeroSkinItem?.OpenDate || '-' }}</n-text></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="英雄皮肤|HeroSkinItem" field="CollitionTagImg" label="收藏标签图片" /></template><ResourcePreview :value="props.activityData.HeroSkinItem?.CollitionTagImg" :status="resourceCheck?.getStatus(props.activityData.HeroSkinItem?.CollitionTagImg || '')" /></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="英雄皮肤|HeroSkinItem" field="KillShowTime" label="击杀显示时间" /></template><n-text>{{ props.activityData.HeroSkinItem?.KillShowTime ?? '-' }}</n-text></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="英雄皮肤|HeroSkinItem" field="BodyOffset" label="身体偏移" /></template>
                    <n-tag v-for="(val, bidx) in props.activityData.HeroSkinItem?.BodyOffset" :key="bidx" size="small" :bordered="false">{{ val }}</n-tag>
                    <span v-if="!props.activityData.HeroSkinItem?.BodyOffset || props.activityData.HeroSkinItem.BodyOffset.length === 0">-</span>
                  </n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="英雄皮肤|HeroSkinItem" field="HasOutFrameIcon" label="是否有外框图标" /></template><n-tag :bordered="false" :type="props.activityData.HeroSkinItem?.HasOutFrameIcon ? 'success' : 'default'">{{ formatBoolean(props.activityData.HeroSkinItem?.HasOutFrameIcon) }}</n-tag></n-descriptions-item>
                </n-descriptions>
              </n-card>
            </n-gi>
          </n-grid>
        </n-tab-pane>

        <!-- 英雄皮肤Spine页签 -->
        <n-tab-pane name="heroSkinSpine" v-if="props.activityData.HeroSkinSpine"
            :tab="() => tBadge('英雄皮肤Spine|HeroSkinSpine', 'Spine配置')">
          <n-grid :cols="24" :x-gap="16" :y-gap="16">
            <n-gi :span="24">
              <n-card title="皮肤Spine配置" size="small" :bordered="false">
                <n-descriptions label-placement="left" :column="2" bordered>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="英雄皮肤|HeroSkinItem" field="SkinItemId" label="皮肤道具ID" /></template><n-tag :bordered="false">{{ props.activityData.HeroSkinSpine?.SkinItemId }}</n-tag></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="英雄皮肤Spine|HeroSkinSpine" field="IsHasSeatSpine" label="座位Spine" /></template><n-tag :bordered="false" :type="props.activityData.HeroSkinSpine?.IsHasSeatSpine ? 'success' : 'default'">{{ formatBoolean(props.activityData.HeroSkinSpine?.IsHasSeatSpine) }}</n-tag></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="英雄皮肤Spine|HeroSkinSpine" field="IsHasBookSpine" label="图鉴Spine" /></template><n-tag :bordered="false" :type="props.activityData.HeroSkinSpine?.IsHasBookSpine ? 'success' : 'default'">{{ formatBoolean(props.activityData.HeroSkinSpine?.IsHasBookSpine) }}</n-tag></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="英雄皮肤Spine|HeroSkinSpine" field="IsHasMainBgSpine" label="主界面Spine" /></template><n-tag :bordered="false" :type="props.activityData.HeroSkinSpine?.IsHasMainBgSpine ? 'success' : 'default'">{{ formatBoolean(props.activityData.HeroSkinSpine?.IsHasMainBgSpine) }}</n-tag></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="英雄皮肤Spine|HeroSkinSpine" field="MainBgFx" label="主背景特效" /></template><ResourcePreview :value="props.activityData.HeroSkinSpine?.MainBgFx" :status="resourceCheck?.getStatus(props.activityData.HeroSkinSpine?.MainBgFx || '')" /></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="英雄皮肤Spine|HeroSkinSpine" field="IsHasSeatKillSpine" label="击杀Spine" /></template><n-tag :bordered="false" :type="props.activityData.HeroSkinSpine?.IsHasSeatKillSpine ? 'success' : 'default'">{{ formatBoolean(props.activityData.HeroSkinSpine?.IsHasSeatKillSpine) }}</n-tag></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="英雄皮肤Spine|HeroSkinSpine" field="KillFxId" label="击杀特效ID" /></template><n-text>{{ props.activityData.HeroSkinSpine?.KillFxId ?? '-' }}</n-text></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="英雄皮肤Spine|HeroSkinSpine" field="IsHasCollitionBgSpine" label="收藏册Spine" /></template><n-tag :bordered="false" :type="props.activityData.HeroSkinSpine?.IsHasCollitionBgSpine ? 'success' : 'default'">{{ formatBoolean(props.activityData.HeroSkinSpine?.IsHasCollitionBgSpine) }}</n-tag></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="英雄皮肤Spine|HeroSkinSpine" field="IsHasCollitionCardBgSpine" label="收藏卡片背景Spine" /></template><n-tag :bordered="false" :type="props.activityData.HeroSkinSpine?.IsHasCollitionCardBgSpine ? 'success' : 'default'">{{ formatBoolean(props.activityData.HeroSkinSpine?.IsHasCollitionCardBgSpine) }}</n-tag></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="英雄皮肤Spine|HeroSkinSpine" field="IsHasCollitionCardBgSpineDuplicate" label="收藏卡片背景Spine(重复)" /></template><n-tag :bordered="false" :type="props.activityData.HeroSkinSpine?.IsHasCollitionCardBgSpineDuplicate ? 'success' : 'default'">{{ formatBoolean(props.activityData.HeroSkinSpine?.IsHasCollitionCardBgSpineDuplicate) }}</n-tag></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="英雄皮肤Spine|HeroSkinSpine" field="SpineAnimAudio" label="Spine动画音效" /></template><ResourcePreview :value="props.activityData.HeroSkinSpine?.SpineAnimAudio" :status="resourceCheck?.getStatus(props.activityData.HeroSkinSpine?.SpineAnimAudio || '')" /></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="英雄皮肤Spine|HeroSkinSpine" field="KillAudio" label="击杀音效" /></template><ResourcePreview :value="props.activityData.HeroSkinSpine?.KillAudio" :status="resourceCheck?.getStatus(props.activityData.HeroSkinSpine?.KillAudio || '')" /></n-descriptions-item>
                </n-descriptions>
              </n-card>
            </n-gi>
          </n-grid>
        </n-tab-pane>

        <!-- 结缘亭页签 -->
        <n-tab-pane name="drawPet" v-if="props.activityData.DrawPet"
            :tab="() => tBadge('结缘亭|DrawPet', '抽奖配置')">
          <n-grid :cols="24" :x-gap="16" :y-gap="16">
            <n-gi :span="24">
              <n-alert type="info" :show-icon="false" size="small">
                <template #header><n-text strong>关联说明</n-text></template>
                <n-text code size="small">
                  Activity.CustomParma[0] → DrawPet.Id = {{ props.activityData.DrawPet?.Id }}
                  <span v-if="!props.activityData.Basic?.CustomParma || props.activityData.Basic?.CustomParma.length === 0">（CustomParma为空，通过 DrawPet.ActivityId 反向关联）</span>
                </n-text>
              </n-alert>
            </n-gi>
            <n-gi :span="24" v-for="period in getDrawPetPeriods" :key="period.title">
              <DrawPetCard
                :title="period.title"
                :draw-pet="period.drawPet"
                :highlight="period.highlight"
                :period-label="period.periodLabel"
              />
            </n-gi>
          </n-grid>
        </n-tab-pane>

        <!-- 灵宠页签：结缘亭活动始终显示 -->
        <n-tab-pane name="pet" v-if="props.activityData.DrawPet"
            :tab="() => tBadge('灵宠|Pet', '灵宠信息')">
          <n-grid :cols="24" :x-gap="16" :y-gap="16">
            <n-gi :span="24">
              <n-alert type="info" :show-icon="false" size="small">
                <template #header><n-text strong>关联说明</n-text></template>
                <n-text code size="small">DrawPet.BigAwardItemId({{ props.activityData.DrawPet?.BigAwardItemId || '无' }}) → Pet.Id</n-text>
              </n-alert>
            </n-gi>
            <!-- 数据缺失提示 -->
            <n-gi :span="24" v-if="!props.activityData.DrawPet?.BigAwardItemId || props.activityData.DrawPet.BigAwardItemId <= 0">
              <n-alert type="warning" size="small">
                <template #header>关联数据缺失</template>
                DrawPet.BigAwardItemId 为空或为 0，无法关联灵宠数据。请检查「结缘亭|DrawPet」表中该记录的 BigAwardItemId 列。
              </n-alert>
            </n-gi>
            <n-gi :span="24" v-else-if="!props.activityData.Pets || props.activityData.Pets.length === 0">
              <n-alert type="warning" size="small">
                <template #header>关联数据缺失</template>
                DrawPet.BigAwardItemId = {{ props.activityData.DrawPet?.BigAwardItemId }}，但在「灵宠|Pet」表中未找到 Id={{ props.activityData.DrawPet?.BigAwardItemId }} 的记录。请检查灵宠表数据。
              </n-alert>
            </n-gi>
            <n-gi :span="24" v-for="(pet, idx) in props.activityData.Pets" :key="pet?.Id || idx">
              <n-card :title="'灵宠: ' + (pet?.Name || '未命名')" size="small" :bordered="false">
                <n-descriptions label-placement="left" :column="2" bordered>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="灵宠|Pet" field="Id" label="灵宠ID" /></template><n-tag :bordered="false">{{ pet?.Id }}</n-tag></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="灵宠|Pet" field="Name" label="名称" /></template><n-text strong>{{ pet?.Name || '未命名' }}</n-text></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="灵宠|Pet" field="WuXingType" label="五行类型" /></template><n-text>{{ pet?.WuXingType || '无' }}</n-text></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="灵宠|Pet" field="PrefabPath" label="预制路径" /></template><ResourcePreview :value="pet?.PrefabPath" :status="resourceCheck?.getStatus(pet?.PrefabPath || '')" /></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="灵宠|Pet" field="Skills" label="技能" /></template>
                    <n-tag :bordered="false" type="info" v-for="(skill, sidx) in pet?.Skills" :key="sidx">{{ skill?.Key }}:{{ skill?.Value }}</n-tag>
                    <span v-if="!pet?.Skills || pet.Skills.length === 0">-</span>
                  </n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="灵宠|Pet" field="BattleAttrWeights" label="战斗属性权重" /></template>
                    <n-tag :bordered="false" type="info" v-for="(attr, aidx) in pet?.BattleAttrWeights" :key="aidx">{{ attr?.Key }}:{{ attr?.Value }}</n-tag>
                    <span v-if="!pet?.BattleAttrWeights || pet.BattleAttrWeights.length === 0">-</span>
                  </n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="灵宠|Pet" field="HighBattleAttrs" label="高阶战斗属性" /></template><n-tag :bordered="false" type="info">{{ formatArray(pet?.HighBattleAttrs) }}</n-tag></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="灵宠|Pet" field="UpgradeLikeItems" label="喜爱升级道具" /></template><n-tag :bordered="false" type="success">{{ formatArray(pet?.UpgradeLikeItems) }}</n-tag></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="灵宠|Pet" field="UpgradeHateItems" label="厌恶升级道具" /></template><n-tag :bordered="false" type="warning">{{ formatArray(pet?.UpgradeHateItems) }}</n-tag></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="灵宠|Pet" field="SquareHeadIcon" label="方形头像" /></template><ResourcePreview :value="pet?.SquareHeadIcon" :status="resourceCheck?.getStatus(pet?.SquareHeadIcon || '')" /></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="灵宠|Pet" field="HeadIcon" label="头像" /></template><ResourcePreview :value="pet?.HeadIcon" :status="resourceCheck?.getStatus(pet?.HeadIcon || '')" /></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="灵宠|Pet" field="Silhouette" label="剪影" /></template><ResourcePreview :value="pet?.Silhouette" :status="resourceCheck?.getStatus(pet?.Silhouette || '')" /></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="灵宠|Pet" field="PopBg" label="弹窗背景" /></template><ResourcePreview :value="pet?.PopBg" :status="resourceCheck?.getStatus(pet?.PopBg || '')" /></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="灵宠|Pet" field="PopIcon" label="弹窗图标" /></template><ResourcePreview :value="pet?.PopIcon" :status="resourceCheck?.getStatus(pet?.PopIcon || '')" /></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="灵宠|Pet" field="PopTitle" label="弹窗标题" /></template><ResourcePreview :value="pet?.PopTitle" :status="resourceCheck?.getStatus(pet?.PopTitle || '')" /></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="灵宠|Pet" field="PetWeekTaskBg" label="周任务背景" /></template><ResourcePreview :value="pet?.PetWeekTaskBg" :status="resourceCheck?.getStatus(pet?.PetWeekTaskBg || '')" /></n-descriptions-item>
<n-descriptions-item >
                    <template #label><BadgeLabel sheet="灵宠|Pet" field="InfoTextID" label="信息文本ID" /></template><n-text>{{ pet?.InfoTextID || '无' }}</n-text></n-descriptions-item>
                </n-descriptions>
              </n-card>
            </n-gi>
          </n-grid>
        </n-tab-pane>

        <!-- 灵宠音效页签：结缘亭活动始终显示 -->
        <n-tab-pane name="petAudio" v-if="props.activityData.DrawPet"
            :tab="() => tBadge('灵宠音效|PetAudio', '音效配置')">
          <n-grid :cols="24" :x-gap="16" :y-gap="16">
            <n-gi :span="24">
              <n-alert type="info" :show-icon="false" size="small">
                <template #header><n-text strong>关联说明</n-text></template>
                <n-text code size="small">PetAudio.ItemCfgId == Pet.Id</n-text>
              </n-alert>
            </n-gi>
            <!-- 数据缺失提示 -->
            <n-gi :span="24" v-if="!props.activityData.Pets || props.activityData.Pets.length === 0">
              <n-alert type="warning" size="small">
                <template #header>关联数据缺失</template>
                灵宠数据(Pet)为空，无法关联音效数据。请先检查「灵宠」页签的缺失原因。
              </n-alert>
            </n-gi>
            <n-gi :span="24" v-else-if="!props.activityData.PetAudios || props.activityData.PetAudios.length === 0">
              <n-alert type="warning" size="small">
                <template #header>关联数据缺失</template>
                灵宠数据存在但未找到对应的音效记录。请检查「灵宠音效|PetAudio」表中 ItemCfgId={{ props.activityData.Pets?.[0]?.Id }} 的记录。
              </n-alert>
            </n-gi>
            <template v-else>
              <n-gi :span="24">
                <n-card title="灵宠音效列表" size="small" :bordered="false">
                  <n-table :bordered="true" :single-line="false" size="small">
                    <thead><tr><th>动画状态</th><th>道具配置ID</th><th>大厅音效</th><th>灵宠窗口音效</th></tr></thead>
                    <tbody>
                    <tr v-for="(audio, idx) in props.activityData.PetAudios" :key="audio?.AnimationState || idx">
                      <template v-if="audio">
                        <td>{{ audio.AnimationState }}</td><td>{{ audio.ItemCfgId }}</td>
                        <td>{{ audio.LobbyAudio || '-' }}</td><td>{{ audio.PetWindowAudio || '-' }}</td>
                      </template>
                    </tr>
                    </tbody>
                  </n-table>
                </n-card>
              </n-gi>
            </template>
          </n-grid>
        </n-tab-pane>

        <!-- activity-wiki-dev: 新增页签 - 累充活动奖励档位表格 -->
        <n-tab-pane name="accumulatedRecharge" v-if="props.activityData.AccumulatedRecharges && props.activityData.AccumulatedRecharges.length > 0"
            :tab="() => tBadge('累充奖励表|AccumulatedRechargeReward', '累充奖励')">
          <n-grid :cols="24" :x-gap="16" :y-gap="16">
            <n-gi :span="24">
              <n-alert type="info" :show-icon="false" size="small">
                <template #header><n-text strong>关联说明</n-text></template>
                <n-text code size="small">AccumulatedRechargeReward.ActId == Activity.EActivityId = {{ props.activityData.Basic?.EActivityId }}</n-text>
              </n-alert>
            </n-gi>
            <n-gi :span="24">
              <n-card title="累充奖励档位" size="small" :bordered="false">
                <n-table :bordered="true" :single-line="false" size="small">
                  <thead><tr><th>奖励ID</th><th>活动ID</th><th>累充金额</th><th>奖励物品</th></tr></thead>
                  <tbody>
                  <tr v-for="(reward, idx) in props.activityData.AccumulatedRecharges" :key="reward?.Id || idx">
                    <template v-if="reward">
                      <td>{{ reward.Id }}</td>
                      <td><n-tag size="small" :bordered="false">{{ reward.ActId || '-' }}</n-tag></td>
                      <td><n-tag size="small" type="success" :bordered="false">{{ reward.RechargeNum || 0 }}</n-tag></td>
                      <td>
                        <n-tag v-for="(cfg, cidx) in reward.RewardItems" :key="cidx" size="small" :bordered="false">{{ cfg?.ItemId }}×{{ cfg?.Count }}</n-tag>
                        <span v-if="!reward.RewardItems || reward.RewardItems.length === 0">-</span>
                      </td>
                    </template>
                  </tr>
                  </tbody>
                </n-table>
              </n-card>
            </n-gi>
          </n-grid>
        </n-tab-pane>
      </n-tabs>
    </div>
  </n-card>
</template>

<style scoped>
.activity-card {
  width: calc(100% - 2px);
  margin-bottom: 16px;
  transition: all 0.3s ease;
}

.activity-card:hover {
  transform: translateX(2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.activity-id-badge {
  padding: 4px 8px;
  background-color: #f0f0f0;
  border-radius: 4px;
  font-size: 12px;
  color: #666;
}

.activity-content {
  min-height: 300px;
}

:deep(.n-card) {
  transition: all 0.3s ease;
}

:deep(.n-card .n-card-header) {
  font-weight: bold;
}

:deep(.n-tag) {
  margin-right: 4px;
  margin-bottom: 4px;
}
</style>
