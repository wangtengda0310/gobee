<!--
  ColumnRuleList - 列级校验规则列表

  通用组件：展示单个列的所有列级校验规则，支持拖拽排序、添加/删除规则、规则类型选择和参数配置。
  通过 props 接收规则列表和规则配置，通过 emits 通知父组件规则变更。
-->
<script setup lang="ts">
import { type Component, type VNode, h, nextTick, ref } from "vue"
import { NCascader, NButton, NCard } from "naive-ui"
import { SortableEvent, VueDraggable } from "vue-draggable-plus"
import { DragHandleFilled } from "@vicons/material"
import { Icon } from "@vicons/utils"
import { ColRule } from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
import { EColRule } from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
import type { CascaderOption } from "naive-ui"

/** 规则参数组件映射值类型：SFC 组件或注入默认值的包装函数 */
type RuleComponentEntry = Component | ((props: { params: { [p: string]: string } }) => VNode)

const props = defineProps<{
  /** 当前列的规则列表（v-model） */
  modelValue: (ColRule | null)[]
  /** 级联选择器选项 */
  ruleOptions: CascaderOption[]
  /** 规则参数组件映射 */
  ruleComponents: Map<EColRule, RuleComponentEntry>
}>()

const emit = defineEmits<{
  /** 更新规则列表 */
  (e: "update:modelValue", value: (ColRule | null)[]): void
  /** 添加规则 */
  (e: "add-rule"): void
  /** 删除规则 */
  (e: "delete-rule", index: number): void
  /** 拖拽开始 */
  (e: "drag-start", event: SortableEvent): void
  /** 拖拽结束 */
  (e: "drag-end", event: SortableEvent): void
}>()

/** 拖拽状态 */
const drag = ref(false)

/** 拖拽开始事件处理 */
const onStart = (e: SortableEvent) => {
  drag.value = true
  emit("drag-start", e)
}

/** 拖拽结束事件处理 */
const onEnd = (e: SortableEvent) => {
  nextTick(() => {
    drag.value = false
  })
  emit("drag-end", e)
}

/** 删除规则 */
const delColRule = (index: number) => {
  emit("delete-rule", index)
}

/**
 * 根据规则类型渲染对应的参数配置组件
 */
const renderRuleParams = (ruleType: EColRule, params: { [p: string]: string }) => {
  const entry = props.ruleComponents.get(ruleType)
  if (!entry) {
    return h("div", { class: "text-gray-400" }, "未找到对应的规则额外参数配置")
  }

  // 包装函数（withDefaults）直接调用，SFC 组件用 h() 渲染
  if (typeof entry === 'function') {
    return (entry as (p: { params: { [p: string]: string } }) => ReturnType<typeof h>)({ params })
  }
  return h(entry, { params })
}
</script>

<template>
  <div style="display: flex; flex-direction: column; gap: 10px; justify-content: space-between">
    <!-- 规则列表 - 支持拖拽排序 -->
    <VueDraggable
      :model-value="modelValue"
      @update:model-value="(v) => emit('update:modelValue', v as (ColRule | null)[])"
      @start="onStart"
      @end="onEnd"
      :animation="150"
      :scroll="true"
      :scroll-sensitivity="300"
      :scroll-speed="20"
      handle=".custom-drag-handle"
      style="display: flex; flex-direction: column; gap: 10px; justify-content: space-between"
    >
      <NCard
        v-for="(v, i) in modelValue"
        :key="v?.Uuid"
        :title="() =>
          h('div', { style: 'display: flex; margin-right: 100px' }, [
            h('div', { style: 'flex: 0 0 100px' }, ['规则' + (i + 1)]),
            h(NCascader, {
              checkStrategy: 'child',
              value: v?.Type,
              'onUpdate:value': (nk: EColRule) => {
                if (v) {
                  v.Type = nk
                  v.Params = {}
                }
              },
              options: ruleOptions,
              placeholder: '选择一个规则',
            }),
          ])
        "
        closable
        @close="delColRule(i)"
        size="small"
      >
        <template #header-extra>
          <NButton class="custom-drag-handle" style="flex: 0 0 80px; margin-right: 50px" type="info" dashed>
            <Icon>
              <DragHandleFilled />
            </Icon>
          </NButton>
        </template>

        <!-- 规则参数区域：用渲染函数避免函数式组件 props 名不匹配 -->
        <div v-if="v" style="display: flex; justify-content: space-between; gap: 10px">
          <component :is="renderRuleParams(v.Type, v.Params as { [p: string]: string })" />
        </div>
      </NCard>
    </VueDraggable>

    <!-- 添加规则按钮插槽 -->
    <div style="margin-top: 10px; display: flex; justify-content: right">
      <slot name="add-button">
        <NButton @click="() => emit('add-rule')">
          <template #icon>
            <svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" viewBox="0 0 24 24">
              <g fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <circle cx="12" cy="12" r="9"></circle>
                <path d="M9 12h6"></path>
                <path d="M12 9v6"></path>
              </g>
            </svg>
          </template>
        </NButton>
      </slot>
    </div>
  </div>
</template>

<style scoped>
/* 拖拽动画效果 */
.fade-move,
.fade-enter-active,
.fade-leave-active {
  transition: all 0.5s cubic-bezier(0.55, 0, 0.1, 1);
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: scaleY(0.01) translate(30px, 0);
}

.fade-leave-active {
  position: absolute;
}
</style>
