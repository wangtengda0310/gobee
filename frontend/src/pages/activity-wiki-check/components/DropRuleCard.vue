<!--
  DropRuleCard - 掉落规则卡片组件

  统一展示单抽/十连掉落规则、掉落组和掉落项。
  通过 props 接收数据，消除单抽/十连之间的模板复制。
-->
<script setup lang="ts">
import {NCard, NDescriptions, NDescriptionsItem, NTag, NText, NTable} from "naive-ui"
import BadgeLabel from "@shared/components/badge-label/index.vue"
import {formatArray, formatBoolean} from "@shared/composables/use-format-utils"
import type {DropRuleDiff} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/drop_rule/models.js"
import type {DropGroupDiff} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/drop_group/models.js"
import type {DropItemDiff} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/drop_item/models.js"

defineProps<{
  /** 卡片标题 */
  title: string
  /** 卡片 CSS class */
  cardClass?: string
  /** 掉落规则数据 */
  rule: DropRuleDiff | null
  /** 掉落组列表 */
  groups: (DropGroupDiff | null)[]
  /** 掉落项列表 */
  items: (DropItemDiff | null)[]
}>()
</script>

<template>
  <n-card :title="title" size="small" :bordered="false" :class="cardClass">
    <n-descriptions label-placement="left" :column="2" bordered>
      <n-descriptions-item>
        <template #label><BadgeLabel sheet="掉落规则表|DropRule" field="Id" label="规则ID" /></template>
        <n-tag :bordered="false">{{ rule?.Id }}</n-tag>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel sheet="掉落规则表|DropRule" field="Name" label="规则名称" /></template>
        <n-text strong>{{ rule?.Name || '未命名' }}</n-text>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel sheet="掉落规则表|DropRule" field="DropCount" label="掉落次数" /></template>
        <n-text>{{ rule?.Count || 0 }}</n-text>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel sheet="掉落规则表|DropRule" field="DropGroup" label="掉落组" /></template>
        <n-tag :bordered="false" type="info">{{ formatArray(rule?.DropGroup) }}</n-tag>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel sheet="掉落规则表|DropRule" field="SmallGuarantee" label="小保底" /></template>
        <n-text>{{ rule?.EnsureSmall || '无' }}</n-text>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel sheet="掉落规则表|DropRule" field="SmallGuaranteeGroup" label="小保底组" /></template>
        <n-tag :bordered="false" type="info">{{ formatArray(rule?.EnsureSmallGroup) }}</n-tag>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel sheet="掉落规则表|DropRule" field="BigGuarantee" label="大保底" /></template>
        <n-text>{{ rule?.EnsureBig || '无' }}</n-text>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel sheet="掉落规则表|DropRule" field="BigGuaranteeGroup" label="大保底组" /></template>
        <n-tag :bordered="false" type="info">{{ formatArray(rule?.EnsureBigGroup) }}</n-tag>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel sheet="掉落规则表|DropRule" field="GuaranteeItemCount" label="保底道具数量" /></template>
        <n-text>{{ rule?.EnsureItemCount || 0 }}</n-text>
      </n-descriptions-item>
      <n-descriptions-item>
        <template #label><BadgeLabel sheet="掉落规则表|DropRule" field="GuaranteeItemId" label="保底道具ID" /></template>
        <n-tag :bordered="false" type="error">{{ rule?.EnsureItemID || '无' }}</n-tag>
      </n-descriptions-item>
      <n-descriptions-item label="道具存在性检查">
        <n-tag :bordered="false" :type="rule?.ItemCheckExist ? 'success' : 'default'">{{ formatBoolean(rule?.ItemCheckExist) }}</n-tag>
      </n-descriptions-item>
    </n-descriptions>
    <!-- 掉落组 -->
    <n-table v-if="groups.length > 0" :bordered="true" :single-line="false" size="small" style="margin-top: 12px;">
      <thead><tr><th>组ID</th><th>名称</th><th>权重</th><th>权重递增</th><th>去重</th><th>有效期开始</th><th>有效期结束</th></tr></thead>
      <tbody>
      <tr v-for="(group, idx) in groups" :key="group?.Id || idx">
        <template v-if="group">
          <td>{{ group.Id }}</td><td>{{ group.Name || '-' }}</td>
          <td>{{ group.Weight || 0 }}</td><td>{{ group.WeightInc || 0 }}</td>
          <td>{{ formatBoolean(group.Deduplication) }}</td>
          <td>{{ group.ValidDate || '-' }}</td><td>{{ group.ExpireDate || '-' }}</td>
        </template>
      </tr>
      </tbody>
    </n-table>
    <!-- 掉落项 -->
    <n-table v-if="items.length > 0" :bordered="true" :single-line="false" size="small" style="margin-top: 12px;">
      <thead><tr><th>项ID</th><th>名称</th><th>掉落组</th><th>道具</th><th>权重</th><th>权重递增</th><th>去重</th><th>检查存在</th><th>排除已有</th><th>必出</th><th>替换组</th><th>有效期</th></tr></thead>
      <tbody>
      <tr v-for="(item, idx) in items" :key="item?.Id || idx">
        <template v-if="item">
          <td>{{ item.Id }}</td><td>{{ item.Name || '-' }}</td><td>{{ item.DropGroup || '-' }}</td>
          <td>
            <n-tag v-for="(cfg, cidx) in item.Item" :key="cidx" size="small" :bordered="false">{{ cfg?.ItemId }}×{{ cfg?.Count }}</n-tag>
            <span v-if="!item.Item || item.Item.length === 0">-</span>
          </td>
          <td>{{ item.Weight || 0 }}</td>
          <td>{{ item.WeightInc || 0 }}</td>
          <td>{{ formatBoolean(item.Deduplication) }}</td>
          <td>{{ formatBoolean(item.CheckExist) }}</td>
          <td>{{ formatBoolean(item.ExcludeExist) }}</td>
          <td>{{ formatBoolean(item.MustHave) }}</td>
          <td>{{ item.ReplaceGroup || '-' }}</td>
          <td>{{ item.ValidDate || '-' }}~{{ item.ExpireDate || '-' }}</td>
        </template>
      </tr>
      </tbody>
    </n-table>
  </n-card>
</template>
