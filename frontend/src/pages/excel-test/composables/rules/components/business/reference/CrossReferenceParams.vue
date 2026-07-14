<!--
  CrossReferenceParams - 跨表引用检查规则参数组件

  五行布局：基础参数 → 行条件过滤 → 正则格式(RegexFormatRow) → 比较操作 → 关联表/关联列
-->
<script setup lang="ts">
import {NCheckbox, NInput, NInputNumber, NSelect} from "naive-ui"
import {ERuleParam} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
import {compareOpOptions} from '../../../chain-reference-params'
import RegexFormatRow from '../../RegexFormatRow.vue'
import FilterConditionRow from '../../FilterConditionRow.vue'

const props = defineProps<{ params: { [p: string]: string } }>()

// 初始化默认值
const params = props.params || {}
if (!(ERuleParam.ALLOW_EMPTY in params)) params[ERuleParam.ALLOW_EMPTY] = 'false'
if (!(ERuleParam.EXCEPTS in params)) params[ERuleParam.EXCEPTS] = ''
if (!(ERuleParam.ALLOW_COMMIT in params)) params[ERuleParam.ALLOW_COMMIT] = 'true'
if (!(ERuleParam.BREAK_LINE in params)) params[ERuleParam.BREAK_LINE] = '3'
if (!('refSheet' in params)) params['refSheet'] = ''
if (!('refCol' in params)) params['refCol'] = ''
if (!(ERuleParam.FILTER_COL in params)) params[ERuleParam.FILTER_COL] = ''
if (!(ERuleParam.FILTER_VAL in params)) params[ERuleParam.FILTER_VAL] = ''
if (!('compareOp' in params)) params['compareOp'] = ''
if (!('matchCol' in params)) params['matchCol'] = ''
if (!('matchRefCol' in params)) params['matchRefCol'] = ''
</script>

<template>
  <div style="flex: 1 1 0; display: flex; flex-direction: column; justify-content: space-between; gap: 10px">
    <!-- 第一行: 基础参数 -->
    <div style="display: flex; gap: 10px; align-items: center">
      <n-checkbox
        style="flex: 0 0 100px"
        :checked="params[ERuleParam.ALLOW_EMPTY] === 'true'"
        @update:checked="params[ERuleParam.ALLOW_EMPTY] = ($event as boolean).toString()"
      >
        允许空值
      </n-checkbox>
      <n-checkbox
        style="flex: 0 0 100px"
        disabled
        :checked="params[ERuleParam.ALLOW_COMMIT] === 'true'"
        @update:checked="params[ERuleParam.ALLOW_COMMIT] = ($event as boolean).toString()"
      >
        允许注释
      </n-checkbox>
      <div>空N行截断:</div>
      <n-input-number
        style="flex: 0 0 150px"
        placeholder="行数"
        :min="1" :max="5"
        :value="Number(params[ERuleParam.BREAK_LINE]) || 3"
        @update:value="params[ERuleParam.BREAK_LINE] = ($event?.toString() || '3')"
      />
      <div>排除检查某些特殊值:</div>
      <n-select
        style="flex: 1 1 0"
        placeholder="值"
        filterable multiple tag
        :value="params[ERuleParam.EXCEPTS] ? params[ERuleParam.EXCEPTS].split(',') : []"
        @update:value="params[ERuleParam.EXCEPTS] = ($event as string[]).join(',')"
      />
    </div>
    <!-- 第二行: 行条件过滤（共用组件，params 模式） -->
    <FilterConditionRow :params="params" label="仅当列:" />
    <!-- 第三行: 正则格式(共用组件，内部处理 pattern/groups/isArray 初始化和 UI) -->
    <RegexFormatRow :params="params" />
    <!-- 同行关联参数 -->
    <div style="display: flex; gap: 10px; align-items: center">
      <div>匹配当前表字段:</div>
      <n-input
        style="flex: 1 1 0"
        placeholder="留空则任意匹配"
        :value="params['matchCol']"
        @update:value="params['matchCol'] = $event"
      />
      <div>匹配列:</div>
      <n-input
        style="flex: 1 1 0"
        placeholder="参照表中匹配列(默认同左)"
        :value="params['matchRefCol']"
        @update:value="params['matchRefCol'] = $event"
      />
    </div>
    <!-- 第四行: 比较操作 -->
    <div style="display: flex; gap: 10px; align-items: center">
      <div>比较操作:</div>
      <n-select
        style="flex: 1 1 0"
        placeholder="默认精确匹配"
        :value="params['compareOp'] || ''"
        :options="compareOpOptions"
        @update:value="params['compareOp'] = $event"
      />
    </div>
    <!-- 第五行: 关联表 + 关联列 -->
    <div style="display: flex; gap: 10px; align-items: center">
      <div>关联表:</div>
      <n-input
        style="flex: 3 1 0"
        placeholder="XX表|XX"
        :value="params['refSheet']"
        @update:value="params['refSheet'] = $event"
      />
      <div>关联列:</div>
      <n-input
        style="flex: 3 1 0"
        placeholder="列名"
        :value="params['refCol']"
        @update:value="params['refCol'] = $event"
      />
    </div>
  </div>
</template>
