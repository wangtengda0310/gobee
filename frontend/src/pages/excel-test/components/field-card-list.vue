<!--
  FieldCardList - 字段展示区域
  展示当前 Sheet 的所有字段卡片，每个卡片包含列级规则配置
  字段卡片在有校验规则时，鼠标悬停显示执行检查按钮
-->
<script setup lang="ts">
import {h} from "vue";
import {EColRule} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule";
import ColumnRuleList from "@shared/components/column-rule-list/index.vue";
import {EColType} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio";
import {startExcelRulesForColumn} from "../composables/func";
import {useMessage} from "naive-ui";

const props = defineProps<{
  /** 当前 Sheet 数据 */
  sheetData: any
  /** 当前列规则数据 */
  checkData: { [key: string]: any }
  /** 级联选择器选项 */
  ruleOptionsList: any[]
  /** 规则参数组件映射 */
  ruleComponentMap: Map<EColRule, any>
}>()

const message = useMessage()

/**
 * 添加列规则
 * @param attrName 列属性名
 */
const addColRule = (attrName: string) => {
  if (!props.checkData[attrName]) return
  if (props.checkData[attrName].PropRules) {
    props.checkData[attrName].PropRules.push({
      Type: EColRule.ALL_BASE,
      Uuid: crypto.randomUUID(),
      DisplayName: "",
      Enabled: true,
      Params: {}
    })
  } else {
    console.error("不存在列")
  }
}

/**
 * 删除列规则
 * @param attrName 列属性名
 * @param index 规则索引
 */
const delColRule = (attrName: string, index: number) => {
  if (!props.checkData[attrName]) return
  if (props.checkData[attrName].PropRules && index < props.checkData[attrName].PropRules.length && index >= 0 && props.checkData[attrName].PropRules.length > 0) {
    props.checkData[attrName].PropRules.splice(index, 1)
  } else {
    console.error("删除报错")
  }
}

/**
 * 执行单个列的检查
 * @param attrName 列属性名
 */
const handleCheckColumn = (attrName: string) => {
  const sheetName = props.sheetData?.label
  if (!sheetName) return
  const colRule = props.checkData[attrName]
  if (!colRule) return
  startExcelRulesForColumn(sheetName, attrName, colRule, message)
}

/**
 * 判断列是否有校验规则
 * @param attrName 列属性名
 */
const hasRules = (attrName: string): boolean => {
  const colRule = props.checkData[attrName]
  return colRule != null && colRule.PropRules != null && colRule.PropRules.length > 0
}
</script>

<template>
  <!-- 列级规则卡片 -->
  <n-card v-for="(col, index) in sheetData.sheetHeader?.Col"
          :title="()=>{
                  return h('div', {style: 'display: flex; gap: 10px;'}, [
                    h('span', {style: 'color: #80FF00; min-width: 28px'}, (Number(index)+1) + '.'),
                    h('span', {style: 'min-width: 200px'}, (col?.AttrCHS ? col.AttrCHS : '空')),
                    h('span', {style: 'color: #DE3163; min-width: 200px'}, (col?.AttrName ? col.AttrName : '空')),
                    h('span', {style: 'color: #FF7F50; min-width: 200px'}, col?.AttrType),
                  ]
                )}"
          :id="(col?.AttrName ? col.AttrName : '空'+index)"
          header-style="font-size: 18px; padding: 10px"
          class="field-card">
    <template #header-extra>
      <div style="font-size:18px; display: flex; align-items: center; gap: 8px;">
        <n-button
            v-if="col?.AttrName && hasRules(col.AttrName)"
            class="check-column-btn"
            type="primary"
            size="small"
            @click="() => handleCheckColumn(col.AttrName)">
          <template #icon>
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="16" height="16">
              <g fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M5 12l5 5l9 -9"/>
              </g>
            </svg>
          </template>
          执行检查
        </n-button>
      </div>
    </template>

    <div
        v-if="col?.AttrName && checkData[col.AttrName] && (col.ColStatus == EColType.NORMAL || col.ColStatus == EColType.ENUM)">
      <!-- 使用公用列级规则列表组件 -->
      <ColumnRuleList
          :model-value="checkData[col.AttrName]!.PropRules"
          :rule-options="ruleOptionsList"
          :rule-components="ruleComponentMap"
          @update:model-value="(v) => { checkData[col.AttrName]!.PropRules = v }"
          @add-rule="() => addColRule(col.AttrName)"
          @delete-rule="(index) => delColRule(col.AttrName, index)"
      />
    </div>
    <div v-else-if="col?.ColStatus == EColType.EMPTY">
      空列
    </div>
    <div v-else-if="col?.ColStatus == EColType.COMMENT">
      注释列
    </div>
    <div v-else-if="col?.ColStatus == EColType.ERROR">
      列错误: {{ col?.Error }}
    </div>
    <div v-else-if="col?.AttrName == ''">
      属性名为空
    </div>
    <div v-else>
      未知错误
    </div>
  </n-card>
</template>

<style scoped>
/* 字段卡片悬停时显示执行检查按钮 */
.field-card :deep(.check-column-btn) {
  opacity: 0;
  transition: opacity 0.2s ease;
}

.field-card:hover :deep(.check-column-btn) {
  opacity: 1;
}
</style>