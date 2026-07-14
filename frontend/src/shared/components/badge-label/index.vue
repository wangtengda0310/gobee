<script setup lang="ts">
/**
 * 带规则角标的字段 Label 组件
 *
 * 封装 renderFieldWithBadge 调用，通过 props 直接传入 ruleCoverage。
 * 也支持通过 provide('ruleCoverage') 注入（无需传 coverage prop）。
 *
 * 用法: <BadgeLabel sheet="商品表|ShopGood" field="Id" label="商品ID" />
 * 或:   <BadgeLabel :coverage="ruleCoverage" sheet="商品表|ShopGood" field="Id" label="商品ID" />
 */
import {computed, inject, type Ref} from "vue"
import {renderFieldWithBadge} from "@shared/composables/use-rule-badge"
import type {RuleCoverageData} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/activity-wiki-check/models.js"

const props = defineProps<{
  /** Sheet名（如 "商品表|ShopGood"） */
  sheet: string
  /** 字段名（如 "Id"） */
  field: string
  /** 显示文本（如 "商品ID"） */
  label: string
  /** 规则覆盖数据（可选，不传则从 provide 注入） */
  coverage?: RuleCoverageData | null
}>()

// 优先使用 prop，其次从 provide 注入
const injectedCoverage = inject<Ref<RuleCoverageData | null> | RuleCoverageData | null>('ruleCoverage', null)

const vnode = computed(() => {
  // prop 优先，其次 inject（处理 ref 和原始值两种情况）
  let data: RuleCoverageData | null = props.coverage ?? null
  if (!data && injectedCoverage != null) {
    data = (injectedCoverage as any)?.value ?? injectedCoverage
  }
  return renderFieldWithBadge(data, props.sheet, props.field, props.label)
})

</script>

<template>
  <component :is="vnode" />
</template>
