<script setup lang="ts">
import {computed, provide, ref} from "vue";
import {useMessage} from "naive-ui";
import BadgeLabel from "@shared/components/badge-label/index.vue";
import {RuleCoverageData} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/activity-wiki-check/models.js";
import {SeasonPassCompleteData} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/activitywiki_def/models.js";
import {formatBoolean, formatItemArray} from "@shared/composables/use-format-utils";
import {renderTabWithBadge} from "@shared/composables/use-rule-badge";
import {GridOutline as TableIcon} from "@vicons/ionicons5";
import {openExcelBySheet} from "@shared/composables/use-open-excel";
import SeasonPassPeriodCard from "./season-pass-period-card.vue";

const props = defineProps<{
  seq: number
  seasonPassData: SeasonPassCompleteData
  excelDir: string
  ruleCoverage: RuleCoverageData | null
}>()

provide('ruleCoverage', computed(() => props.ruleCoverage))

const activeTab = ref('basic');

const tabToSheetMap: Record<string, string> = {
  'basic': '赛季战令表|SeasonPass',
  'bags': '战令礼包表|SeasonPassBag',
  'rewards': '赛季战令奖励表|SeasonPassReward',
  'tasks': '战令任务表|SeasonPassTask',
}

const message = useMessage()

const handleOpenExcel = async () => {
  const sheetName = tabToSheetMap[activeTab.value]
  if (!sheetName) {
    return
  }
  await openExcelBySheet(message, sheetName, props.excelDir)
}

const tBadge = (sheet: string, label: string) => renderTabWithBadge(props.ruleCoverage, sheet, label)

// 构建三期战令展示数据
interface SeasonPassPeriod {
  title: string
  seasonPass: any
  highlight: boolean
  periodLabel: string
}

const getSeasonPassPeriods = computed(() => {
  const periods: SeasonPassPeriod[] = []
  if ((props.seasonPassData as any).PrevSeasonPass) {
    periods.push({
      title: '上一期战令',
      seasonPass: (props.seasonPassData as any).PrevSeasonPass,
      highlight: false,
      periodLabel: '上一期'
    })
  }
  if (props.seasonPassData.Basic) {
    periods.push({
      title: '本期战令 (当前关联)',
      seasonPass: props.seasonPassData.Basic,
      highlight: true,
      periodLabel: '当前关联'
    })
  }
  if ((props.seasonPassData as any).NextSeasonPass) {
    periods.push({
      title: '下一期战令',
      seasonPass: (props.seasonPassData as any).NextSeasonPass,
      highlight: false,
      periodLabel: '下一期'
    })
  }
  return periods
})
</script>

<template>
  <n-card
      :id="'SeasonPassId:' + props.seasonPassData.Basic?.Id"
      :segmented="{ content: true, footer: true }"
      hoverable
      class="season-pass-card"
  >
    <template #header>
      <div style="display: flex; align-items: center; gap: 8px;">
        <span style="font-size: 18px; font-weight: bold;">{{ seq + 1 }}. {{ props.seasonPassData.Basic?.Name || '未命名战令' }}</span>
        <span style="padding: 2px 8px; border-radius: 4px; font-size: 14px; background-color: #e3f2fd; color: #1565c0;">战令</span>
      </div>
    </template>
    <template #header-extra>
      <div style="display: flex; align-items: center; gap: 8px;">
        <n-button size="small" type="primary" ghost @click="handleOpenExcel">
          <template #icon><n-icon><TableIcon /></n-icon></template>
          打开Excel
        </n-button>
        <div class="season-pass-id-badge">ID: {{ props.seasonPassData.Basic?.Id }}</div>
      </div>
    </template>

    <div class="season-pass-content">
      <n-tabs v-model:value="activeTab" type="line" animated>
        <!-- 基础信息页签 - 参考丹青阁抽奖配置页签布局：纵向卡片排列 -->
        <n-tab-pane name="basic" :tab="() => tBadge('赛季战令表|SeasonPass', '基础信息')">
          <n-grid :cols="24" :x-gap="16" :y-gap="16">
            <!-- 三期战令纵向排列 -->
            <n-gi :span="24" v-for="period in getSeasonPassPeriods" :key="period.title">
              <SeasonPassPeriodCard
                :title="period.title"
                :season-pass="period.seasonPass"
                :highlight="period.highlight"
                :period-label="period.periodLabel"
              />
            </n-gi>

            <!-- 关联关系链可视化 -->
            <n-gi :span="24">
              <n-card title="关联关系链" size="small" :bordered="false">
                <n-steps :current="999" size="small">
                  <n-step title="赛季战令表|SeasonPass" :description="`时间匹配 → Id = ${props.seasonPassData.Basic?.Id}`" />
                  <n-step v-if="props.seasonPassData.Bags && props.seasonPassData.Bags.length > 0"
                      title="战令礼包表|SeasonPassBag"
                      :description="`SeasonPassId = ${props.seasonPassData.Basic?.Id}，共 ${props.seasonPassData.Bags.length} 个`" />
                  <n-step v-if="props.seasonPassData.Rewards && props.seasonPassData.Rewards.length > 0"
                      title="赛季战令奖励表|SeasonPassReward"
                      :description="`SeasonPassId = ${props.seasonPassData.Basic?.Id}，共 ${props.seasonPassData.Rewards.length} 级`" />
                  <n-step v-if="props.seasonPassData.Tasks && props.seasonPassData.Tasks.length > 0"
                      title="战令任务表|SeasonPassTask"
                      :description="`SeasonPassId = ${props.seasonPassData.Basic?.Id}，共 ${props.seasonPassData.Tasks.length} 个`" />
                </n-steps>
              </n-card>
            </n-gi>

            <!-- 无数据提示 -->
            <n-gi :span="24" v-if="getSeasonPassPeriods.length === 0">
              <n-empty description="无战令数据" size="small" />
            </n-gi>
          </n-grid>
        </n-tab-pane>

        <!-- 礼包页签 - 只显示本期 -->
        <n-tab-pane name="bags" v-if="props.seasonPassData.Bags && props.seasonPassData.Bags.length > 0"
            :tab="() => tBadge('战令礼包表|SeasonPassBag', '战令礼包')">
          <n-grid :cols="24" :x-gap="16" :y-gap="16">
            <!-- 关联说明 -->
            <n-gi :span="24">
              <n-alert type="info" :show-icon="false" size="small">
                <template #header><n-text strong>关联说明</n-text></template>
                <n-text code size="small">SeasonPassBag.SeasonPassId == SeasonPass.Id = {{ props.seasonPassData.Basic?.Id }}</n-text>
              </n-alert>
            </n-gi>
            <n-gi :span="24" v-for="(bag, idx) in props.seasonPassData.Bags" :key="bag?.Id || idx">
              <n-card :title="'礼包: ' + (bag?.Name || '未命名') + ' (ID: ' + bag?.Id + ')'" size="small" :bordered="false">
                <n-descriptions label-placement="left" :column="3" bordered>
                  <n-descriptions-item>
                    <template #label><BadgeLabel sheet="战令礼包表|SeasonPassBag" field="Id" label="礼包ID" /></template>
                    <n-tag :bordered="false">{{ bag?.Id }}</n-tag>
                  </n-descriptions-item>
                  <n-descriptions-item>
                    <template #label><BadgeLabel sheet="战令礼包表|SeasonPassBag" field="BagType" label="礼包类型" /></template>
                    <n-tag :bordered="false" type="primary">{{ bag?.BagType || '-' }}</n-tag>
                  </n-descriptions-item>
                  <n-descriptions-item>
                    <template #label><BadgeLabel sheet="战令礼包表|SeasonPassBag" field="IsUnlockTask" label="解锁任务" /></template>
                    <n-tag :bordered="false" :type="bag?.IsUnlockTask ? 'success' : 'default'">{{ formatBoolean(bag?.IsUnlockTask) }}</n-tag>
                  </n-descriptions-item>
                  <n-descriptions-item>
                    <template #label><BadgeLabel sheet="战令礼包表|SeasonPassBag" field="UnlockHighlReward" label="解锁高级奖励" /></template>
                    <n-tag :bordered="false" :type="bag?.UnlockHighlReward ? 'success' : 'default'">{{ formatBoolean(bag?.UnlockHighlReward) }}</n-tag>
                  </n-descriptions-item>
                  <n-descriptions-item>
                    <template #label><BadgeLabel sheet="战令礼包表|SeasonPassBag" field="PassExtWeeklyExpLimit" label="额外周经验" /></template>
                    <n-text>{{ bag?.PassExtWeeklyExpLimit || 0 }}</n-text>
                  </n-descriptions-item>
                  <n-descriptions-item>
                    <template #label><BadgeLabel sheet="战令礼包表|SeasonPassBag" field="AddLevel" label="增加等级" /></template>
                    <n-text>{{ bag?.AddLevel || 0 }}</n-text>
                  </n-descriptions-item>
                  <n-descriptions-item>
                    <template #label><BadgeLabel sheet="战令礼包表|SeasonPassBag" field="FirstReward" label="首购奖励" /></template>
                    <n-tag v-for="(cfg, cidx) in bag?.FirstReward" :key="cidx" size="small" :bordered="false">{{ cfg?.ItemId }}×{{ cfg?.Count }}</n-tag>
                    <span v-if="!bag?.FirstReward || bag.FirstReward.length === 0">-</span>
                  </n-descriptions-item>
                  <n-descriptions-item>
                    <template #label><BadgeLabel sheet="战令礼包表|SeasonPassBag" field="UpReward" label="升级奖励" /></template>
                    <n-tag v-for="(cfg, cidx) in bag?.UpReward" :key="cidx" size="small" :bordered="false">{{ cfg?.ItemId }}×{{ cfg?.Count }}</n-tag>
                    <span v-if="!bag?.UpReward || bag.UpReward.length === 0">-</span>
                  </n-descriptions-item>
                  <n-descriptions-item>
                    <template #label><BadgeLabel sheet="战令礼包表|SeasonPassBag" field="ShowReward" label="展示奖励" /></template>
                    <n-tag v-for="(cfg, cidx) in bag?.ShowReward" :key="cidx" size="small" :bordered="false">{{ cfg?.ItemId }}×{{ cfg?.Count }}</n-tag>
                    <span v-if="!bag?.ShowReward || bag.ShowReward.length === 0">-</span>
                  </n-descriptions-item>
                </n-descriptions>
              </n-card>
            </n-gi>
          </n-grid>
        </n-tab-pane>

        <!-- 奖励页签 - 只显示本期 -->
        <n-tab-pane name="rewards" v-if="props.seasonPassData.Rewards && props.seasonPassData.Rewards.length > 0"
            :tab="() => tBadge('赛季战令奖励表|SeasonPassReward', '等级奖励')">
          <n-grid :cols="24" :x-gap="16" :y-gap="16">
            <!-- 关联说明 -->
            <n-gi :span="24">
              <n-alert type="info" :show-icon="false" size="small">
                <template #header><n-text strong>关联说明</n-text></template>
                <n-text code size="small">SeasonPassReward.SeasonPassId == SeasonPass.Id = {{ props.seasonPassData.Basic?.Id }}</n-text>
              </n-alert>
            </n-gi>
            <n-gi :span="24">
              <n-card title="等级奖励列表" size="small" :bordered="false">
                <n-table :bordered="true" :single-line="false" size="small">
                  <thead>
                    <tr>
                      <th>等级</th>
                      <th>普通奖励</th>
                      <th>高级奖励</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="(reward, idx) in props.seasonPassData.Rewards" :key="reward?.Id || idx">
                      <template v-if="reward">
                        <td><n-tag size="small" type="primary" :bordered="false">{{ reward.Level }}</n-tag></td>
                        <td>
                          <n-tag v-for="(cfg, cidx) in reward.NormalReward" :key="cidx" size="small" :bordered="false">{{ cfg?.ItemId }}×{{ cfg?.Count }}</n-tag>
                          <span v-if="!reward.NormalReward || reward.NormalReward.length === 0">-</span>
                        </td>
                        <td>
                          <n-tag v-for="(cfg, cidx) in reward.HighReward" :key="cidx" size="small" type="success" :bordered="false">{{ cfg?.ItemId }}×{{ cfg?.Count }}</n-tag>
                          <span v-if="!reward.HighReward || reward.HighReward.length === 0">-</span>
                        </td>
                      </template>
                    </tr>
                  </tbody>
                </n-table>
              </n-card>
            </n-gi>
          </n-grid>
        </n-tab-pane>

        <!-- 任务页签 - 只显示本期 -->
        <n-tab-pane name="tasks" v-if="props.seasonPassData.Tasks && props.seasonPassData.Tasks.length > 0"
            :tab="() => tBadge('战令任务表|SeasonPassTask', '战令任务')">
          <n-grid :cols="24" :x-gap="16" :y-gap="16">
            <!-- 关联说明 -->
            <n-gi :span="24">
              <n-alert type="info" :show-icon="false" size="small">
                <template #header><n-text strong>关联说明</n-text></template>
                <n-text code size="small">SeasonPassTask.SeasonPassId == SeasonPass.Id = {{ props.seasonPassData.Basic?.Id }}</n-text>
              </n-alert>
            </n-gi>
            <n-gi :span="24">
              <n-card title="任务列表" size="small" :bordered="false">
                <n-table :bordered="true" :single-line="false" size="small">
                  <thead>
                    <tr>
                      <th>任务ID</th>
                      <th>名称</th>
                      <th>类别</th>
                      <th>子类型</th>
                      <th>完成条件</th>
                      <th>奖励</th>
                      <th>经验</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="(task, idx) in props.seasonPassData.Tasks" :key="task?.Id || idx">
                      <template v-if="task">
                        <td>{{ task.Id }}</td>
                        <td><n-text strong>{{ task.Name || '-' }}</n-text></td>
                        <td><n-tag size="small" :bordered="false">{{ task.Class || '-' }}</n-tag></td>
                        <td><n-tag size="small" :bordered="false">{{ task.SubType || '-' }}</n-tag></td>
                        <td><n-text code size="small">{{ task.CompleteCond || '-' }}</n-text></td>
                        <td>
                          <n-tag v-for="(cfg, cidx) in task.Reward" :key="cidx" size="small" :bordered="false">{{ cfg?.ItemId }}×{{ cfg?.Count }}</n-tag>
                          <span v-if="!task.Reward || task.Reward.length === 0">-</span>
                        </td>
                        <td><n-tag size="small" type="success" :bordered="false">{{ task.PassExp || 0 }}</n-tag></td>
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
.season-pass-card {
  width: calc(100% - 2px);
  margin-bottom: 16px;
  transition: all 0.3s ease;
}

.season-pass-card:hover {
  transform: translateX(2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.season-pass-id-badge {
  padding: 4px 8px;
  background-color: #f0f0f0;
  border-radius: 4px;
  font-size: 12px;
  color: #666;
}

.season-pass-content {
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
