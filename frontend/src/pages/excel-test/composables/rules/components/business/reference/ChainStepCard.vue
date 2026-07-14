<!--
  ChainStepCard - 关系链步骤卡片组件

  单个步骤的参数编辑卡片，左右链共用。
  通过 side 属性区分来源链(left)和目标链(right)的显示差异。

  参数顺序：
  - 左链第一步：title 正则 过滤 nextCol(与下一步匹配)
  - 右链第一步：title 表 preCol(与左边比较) 正则 过滤 nextCol(与下一步匹配)
  - 中间步骤：title preCol(与上一步匹配) 正则 过滤 nextCol(与下一步匹配)
  - 最后一步(左)：title preCol(与上一步匹配) 正则 过滤 nextCol(与右边匹配)
  - 最后一步(右)：title preCol(与上一步匹配) 正则 过滤 nextCol(与左边匹配)
-->
<script setup lang="ts">
import {computed} from "vue"
import {NButton, NInput} from "naive-ui"
import {type ChainStep} from '@/pages/excel-test/composables/rules/chain-reference-params'
import RegexFormatRow from '../../RegexFormatRow.vue'
import FilterConditionRow from '../../FilterConditionRow.vue'

const props = defineProps<{
  step: ChainStep
  index: number
  total: number
  side: 'left' | 'right'
}>()

const emit = defineEmits<{
  update: [field: string, value: string]
}>()

const isLeft = computed(() => props.side === 'left')
const isFirst = computed(() => props.index === 0)
const isLast = computed(() => props.index === props.total - 1)
const hasMultiple = computed(() => props.total > 1)

// 左侧第一步不显示表名，右侧始终显示
const showSheet = computed(() => isLeft.value ? !isFirst.value : true)

// 第一步的查找列描述（仅右链第一步显示）
const firstPreColLabel = computed(() => isLeft.value ? '' : '与左边比较')

// 提取列描述：根据位置和 side 决定
const extractLabel = computed(() => {
  if (isLast.value) {
    return isLeft.value ? '与右边匹配' : '与左边匹配'
  }
  return '与下一步匹配'
})

function onInsert() { emit('update', '__insert__', String(props.index)) }
function onAppend() { emit('update', '__append__', String(props.index)) }
function onRemove() { emit('update', '__remove__', String(props.index)) }

function set(field: string, value: string) {
  emit('update', field, value)
}
</script>

<template>
  <div class="step-card">
    <!-- 悬停操作: 顶部插入 -->
    <div class="step-card-actions-top">
      <n-button size="tiny" quaternary @click="onInsert">向上插入</n-button>
    </div>
    <!-- 悬停操作: 底部追加 -->
    <div class="step-card-actions-bottom">
      <n-button size="tiny" quaternary @click="onAppend">向后追加</n-button>
    </div>

    <!-- 标题行 -->
    <div style="display: flex; align-items: center; margin-bottom: 6px">
      <div style="font-weight: bold; font-size: 13px">步骤{{ index + 1 }}</div>
      <template v-if="showSheet">
        <div style="flex: 0 0 25px; font-size: 12px; margin-left: 8px">表:</div>
        <n-input style="flex: 1 1 0" size="small" placeholder="表名" :value="step.sheet" @update:value="set('sheet', $event)" />
      </template>
      <div style="flex: 1"></div>
      <n-button size="tiny" type="error" quaternary @click="onRemove">x</n-button>
    </div>

    <!-- 右链第一步: preCol 与左边比较 -->
    <div v-if="isFirst && !isLeft" style="display: flex; gap: 8px; align-items: center; margin-bottom: 4px">
      <div style="flex: 0 0 25px; font-size: 12px">列:</div>
      <n-input style="flex: 1 1 0" size="small" placeholder="列名" :value="step.preCol" @update:value="set('preCol', $event)" />
      <div style="font-size: 11px; color: rgba(255,255,255,0.5); white-space: nowrap">{{ firstPreColLabel }}</div>
    </div>

    <!-- 非第一步: preCol 与上一步匹配 -->
    <div v-if="!isFirst" style="display: flex; gap: 8px; align-items: center; margin-bottom: 4px">
      <div style="flex: 0 0 25px; font-size: 12px">列:</div>
      <n-input style="flex: 1 1 0" size="small" placeholder="列名" :value="step.preCol" @update:value="set('preCol', $event)" />
      <div style="font-size: 11px; color: rgba(255,255,255,0.5); white-space: nowrap">与上一步匹配</div>
    </div>

    <!-- 正则提取 + 是否为数组（共用组件，v-model 模式绑定 step 字段） -->
    <div style="margin-bottom: 4px">
      <RegexFormatRow
        :pattern="step.pattern"
        :groups="step.groups"
        :is-array="step.isArray"
        :show-is-array="true"
        @update:pattern="set('pattern', $event)"
        @update:groups="set('groups', $event)"
        @update:is-array="set('isArray', $event)"
      />
    </div>

    <!-- 过滤条件（共用组件，v-model 模式绑定 step 字段） -->
    <div style="margin-bottom: 4px">
      <FilterConditionRow
        :filter-col="step.filterCol"
        :filter-val="step.filterVal"
        :filter-is-array="step.filterIsArray"
        :filter-mode="step.filterMode"
        :filter-days="step.filterDays"
        label="过滤:"
        @update:filter-col="set('filterCol', $event)"
        @update:filter-val="set('filterVal', $event)"
        @update:filter-is-array="set('filterIsArray', $event)"
        @update:filter-mode="set('filterMode', $event)"
        @update:filter-days="set('filterDays', $event)"
      />
    </div>

    <!-- nextCol: 所有步骤都显示（单步时隐藏，因为没有下一步） -->
    <div v-if="hasMultiple" style="display: flex; gap: 8px; align-items: center">
      <div style="flex: 0 0 25px; font-size: 12px">列</div>
      <n-input style="flex: 1 1 0" size="small" placeholder="字段名" :value="step.nextCol" @update:value="set('nextCol', $event)" />
      <div style="font-size: 11px; color: rgba(255,255,255,0.5); white-space: nowrap">{{ extractLabel }}</div>
    </div>
  </div>
</template>

<style scoped>
.step-card {
  position: relative;
  border: 1px solid rgba(128,128,128,0.3);
  border-radius: 4px;
  padding: 8px;
}
.step-card-actions-top {
  position: absolute;
  top: -12px;
  left: 50%;
  transform: translateX(-50%);
  display: none;
  z-index: 1;
}
.step-card-actions-bottom {
  position: absolute;
  bottom: -12px;
  left: 50%;
  transform: translateX(-50%);
  display: none;
  z-index: 1;
}
.step-card:hover .step-card-actions-top,
.step-card:hover .step-card-actions-bottom {
  display: block;
}
</style>
