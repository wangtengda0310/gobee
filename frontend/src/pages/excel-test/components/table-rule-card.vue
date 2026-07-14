<!-- TableRuleCard - 表级校验规则卡片 -->
<script setup lang="ts">
import { watch } from 'vue'
import TableRulePanel from '@shared/components/table-rule-panel/index.vue'
import {
	nowSheetData,
	nowTableRules,
	excelTableRulesMap,
	allTableRuleMetas,
	loadTableRuleMetasForSheet,
	tableRuleMetasLoaded
} from '../composables/use-excel-check-data'

watch(() => nowSheetData.value?.label, (sheetName) => {
	if (sheetName) {
		loadTableRuleMetasForSheet(sheetName)
	}
}, { immediate: true })

const onTableRulesUpdate = (rules: typeof nowTableRules.value) => {
	nowTableRules.value = rules
	syncTableRulesToMap()
}

const syncTableRulesToMap = () => {
	const sheetName = nowSheetData.value?.label
	if (sheetName) {
		excelTableRulesMap.value.set(sheetName, [...nowTableRules.value])
	}
}
</script>

<template>
	<!-- 表级校验规则 -->
	<TableRulePanel
		:rule-metas="allTableRuleMetas"
		:rules="nowTableRules"
		:loaded="tableRuleMetasLoaded"
		@update:rules="onTableRulesUpdate"
	/>
</template>