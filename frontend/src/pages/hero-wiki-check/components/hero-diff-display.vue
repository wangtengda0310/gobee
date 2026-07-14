<!-- HeroDiffDisplay.vue -->
/**
 * 武将差异显示组件
 *
 * 显示武将字段的 diff 变化详情，包括新增、删除、修改等信息
 */
<script setup lang="ts">
import {computed} from 'vue'
import {NIcon, NPopover, NScrollbar, NSpace, NTag, NText, useThemeVars} from 'naive-ui'
import {Add24Filled, Subtract24Filled, Warning24Filled} from '@vicons/fluent'
import {DataContainer} from '@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/diff'

const props = defineProps<{
  diffExcels: DataContainer
  heroId: string
  heroName: string
}>()

// 获取Naive主题变量
const themeVars = useThemeVars()

// 获取当前武将的diff详情
const heroDiff = computed(() => {
  return props.diffExcels.HeroWikiDiffResult?.HeroesDiff[props.heroId]
})

// 判断是否有变化
const hasChanges = computed(() => {
  return heroDiff.value && heroDiff.value.ChangeCount > 0
})

// 获取变化类型对应的样式和图标
const getChangeTypeInfo = (type: string) => {
  const map: Record<string, { type: 'success' | 'error' | 'warning'; icon: any; text: string }> = {
    'added': {type: 'success', icon: Add24Filled, text: '新增'},
    'removed': {type: 'error', icon: Subtract24Filled, text: '删除'},
    'modified': {type: 'warning', icon: Warning24Filled, text: '修改'}
  }
  return map[type] || map['modified']
}

// 格式化值显示
const formatValue = (value: any): string => {
  if (value === null || value === undefined) return 'null'
  if (typeof value === 'object') {
    if (Array.isArray(value)) {
      if (value.length === 0) return '[]'
      return `[${value.join(', ')}]`
    }
    return JSON.stringify(value)
  }
  return String(value)
}

// 获取字段路径的最后一部分
const getLastPathPart = (path: string): string => {
  const parts = path.split('.')
  return parts[parts.length - 1]
}

// 判断值是否为基本类型
const isPrimitive = (value: any): boolean => {
  return value === null ||
      typeof value === 'undefined' ||
      typeof value === 'string' ||
      typeof value === 'number' ||
      typeof value === 'boolean'
}
</script>

<template>
  <div class="hero-diff-display">
    <!-- 汇总徽章 - 使用Naive NTag -->
    <n-popover
        v-if="hasChanges"
        trigger="hover"
        placement="right"
        :width="450"
        scrollable
        style="--n-popover-padding: 0;"
    >
      <template #trigger>
        <n-tag
            :type="getChangeTypeInfo(heroDiff?.ChangeType ?? '').type"
            :bordered="false"
            size="small"
            class="diff-badge"
        >
          <template #icon>
            <n-icon :component="getChangeTypeInfo(heroDiff?.ChangeType ?? '').icon"/>
          </template>
          {{ heroDiff?.ChangeCount }}处变化
        </n-tag>
      </template>

      <!-- Diff详情弹窗 - 完全使用Naive组件和主题变量 -->
      <div class="diff-detail-popup" :style="{ background: themeVars.popoverColor }">
        <div class="diff-header" :style="{ borderBottom: `1px solid ${themeVars.borderColor}` }">
          <n-text strong class="hero-name">{{ heroName }}</n-text>
          <n-tag
              :type="getChangeTypeInfo(heroDiff?.ChangeType ?? '').type"
              size="small"
              :bordered="false"
          >
            {{ getChangeTypeInfo(heroDiff?.ChangeType ?? '').text }}
          </n-tag>
        </div>

        <n-scrollbar style="max-height: 450px;" class="diff-scrollbar">
          <div class="diff-content">
            <!-- 基本字段变化 -->
            <div v-if="heroDiff?.FieldChanges?.length" class="diff-section">
              <div class="section-title" :style="{ borderLeftColor: themeVars.primaryColor }">
                <n-text strong>字段变化</n-text>
              </div>
              <n-space vertical :size="8">
                <div
                    v-for="(change, index) in heroDiff.FieldChanges"
                    :key="index"
                    class="change-item nested"
                    :style="{
                      background: themeVars.cardColor,
                      borderColor: themeVars.borderColor
                    }"
                >
                  <div class="change-field">
                    <n-text strong class="field-name">{{ getLastPathPart(change?.FieldPath ?? '') }}</n-text>
                    <n-text v-if="change?.ValueType" depth="3" class="field-type">
                      ({{ change?.ValueType }})
                    </n-text>
                  </div>
                  <div class="change-values">
                    <div class="old-value">
                      <n-text :style="{ color: themeVars.errorColor }" class="label">旧:</n-text>
                      <div
                          class="value"
                          :style="{
                            background: themeVars.errorColorSuppl,
                            color: themeVars.textColor1
                          }"
                      >
                        <n-text depth="2">{{ formatValue(change?.OldValue) }}</n-text>
                      </div>
                    </div>
                    <div class="new-value">
                      <n-text :style="{ color: themeVars.successColor }" class="label">新:</n-text>
                      <div
                          class="value"
                          :style="{
                            background: themeVars.successColorSuppl,
                            color: themeVars.textColor1
                          }"
                      >
                        <n-text depth="2">{{ formatValue(change?.NewValue) }}</n-text>
                      </div>
                    </div>
                  </div>
                </div>
              </n-space>
            </div>

            <!-- 嵌套结构变化 -->
            <div v-if="heroDiff?.NestedChanges && Object.keys(heroDiff.NestedChanges).length" class="diff-section">
              <div class="section-title" :style="{ borderLeftColor: themeVars.primaryColor }">
                <n-text strong>嵌套结构变化</n-text>
              </div>
              <n-space vertical :size="12">
                <div
                    v-for="(nested, path) in heroDiff.NestedChanges"
                    :key="path"
                    class="nested-item"
                >
                  <div
                      class="nested-header"
                      :style="{
                        background: themeVars.tableHeaderColor,
                        borderBottom: `1px solid ${themeVars.borderColor}`
                      }"
                  >
                    <div class="nested-path-info">
                      <n-text code class="nested-path">{{ path }}</n-text>
                      <n-tag size="tiny" type="info" :bordered="false">
                        {{ nested?.StructType }}
                      </n-tag>
                    </div>
                    <n-badge
                        :value="nested?.FieldCount"
                        :max="99"
                        :type="'info'"
                        show-zero
                    />
                  </div>

                  <div class="nested-changes">
                    <n-space vertical :size="6">
                      <div
                          v-for="(change, idx) in nested?.Changes"
                          :key="idx"
                          class="change-item nested"
                          :style="{
                            background: themeVars.cardColor,
                            borderColor: themeVars.borderColor
                          }"
                      >
                        <div class="change-field">
                          <n-text strong class="field-name">{{ change?.FieldName }}</n-text>
                        </div>
                        <div class="change-values">
                          <div class="old-value">
                            <n-text :style="{ color: themeVars.errorColor }" class="label">旧:</n-text>
                            <div
                                class="value"
                                :style="{
                                  background: themeVars.errorColorSuppl,
                                  color: themeVars.textColor1
                                }"
                            >
                              <n-text depth="2">{{ formatValue(change?.OldValue) }}</n-text>
                            </div>
                          </div>
                          <div class="new-value">
                            <n-text :style="{ color: themeVars.successColor }" class="label">新:</n-text>
                            <div
                                class="value"
                                :style="{
                                  background: themeVars.successColorSuppl,
                                  color: themeVars.textColor1
                                }"
                            >
                              <n-text depth="2">{{ formatValue(change?.NewValue) }}</n-text>
                            </div>
                          </div>
                        </div>
                      </div>
                    </n-space>
                  </div>

                  <!-- 递归显示子嵌套 -->
                  <div v-if="nested?.Children && Object.keys(nested.Children).length" class="nested-children">
                    <n-space vertical :size="8">
                      <div
                          v-for="(child, childPath) in nested.Children"
                          :key="childPath"
                          class="nested-item child"
                          :style="{ borderLeftColor: themeVars.borderColor }"
                      >
                        <div
                            class="nested-header child"
                            :style="{
                              background: themeVars.tableHeaderColor,
                              borderBottom: `1px solid ${themeVars.borderColor}`
                            }"
                        >
                          <div class="nested-path-info">
                            <n-text code class="nested-path">{{ childPath }}</n-text>
                            <n-tag size="tiny" type="info" :bordered="false">
                              {{ child?.StructType }}
                            </n-tag>
                          </div>
                          <n-badge
                              v-if="child?.FieldCount"
                              :value="child?.FieldCount"
                              :max="99"
                              :type="'info'"
                          />
                        </div>

                        <!-- 显示子变化 -->
                        <div v-if="child?.Changes?.length" class="nested-changes">
                          <n-space vertical :size="6">
                            <div
                                v-for="(childChange, childIdx) in child.Changes"
                                :key="childIdx"
                                class="change-item nested"
                                :style="{
                                  background: themeVars.cardColor,
                                  borderColor: themeVars.borderColor
                                }"
                            >
                              <div class="change-field">
                                <n-text strong class="field-name">{{ childChange?.FieldName }}</n-text>
                              </div>
                              <div class="change-values">
                                <div class="old-value">
                                  <n-text :style="{ color: themeVars.errorColor }" class="label">旧:</n-text>
                                  <div
                                      class="value"
                                      :style="{
                                        background: themeVars.errorColorSuppl,
                                        color: themeVars.textColor1
                                      }"
                                  >
                                    <n-text depth="2">{{ formatValue(childChange?.OldValue) }}</n-text>
                                  </div>
                                </div>
                                <div class="new-value">
                                  <n-text :style="{ color: themeVars.successColor }" class="label">新:</n-text>
                                  <div
                                      class="value"
                                      :style="{
                                        background: themeVars.successColorSuppl,
                                        color: themeVars.textColor1
                                      }"
                                  >
                                    <n-text depth="2">{{ formatValue(childChange?.NewValue) }}</n-text>
                                  </div>
                                </div>
                              </div>
                            </div>
                          </n-space>
                        </div>
                      </div>
                    </n-space>
                  </div>
                </div>
              </n-space>
            </div>

            <!-- 无变化时的提示 -->
            <div
                v-if="(!heroDiff?.FieldChanges?.length && (!heroDiff?.NestedChanges || Object.keys(heroDiff.NestedChanges).length === 0))"
                class="diff-section no-changes"
            >
              <n-empty description="无详细变化信息" size="small"/>
            </div>
          </div>
        </n-scrollbar>
      </div>
    </n-popover>
  </div>
</template>

<style scoped>
.hero-diff-display {
  display: inline-block;
  margin-left: 8px;
}

.diff-badge {
  cursor: pointer;
  transition: all 0.3s ease;
}

.diff-badge:hover {
  transform: scale(1.05);
  filter: brightness(1.1);
}

.diff-detail-popup {
  padding: 12px;
  border-radius: 8px;
  min-width: 350px;
}

.diff-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  margin: -12px -12px 12px -12px;
}

.hero-name {
  font-size: 16px;
  font-weight: 600;
}

.diff-scrollbar {
  margin-right: -4px;
}

.diff-content {
  padding-right: 8px;
}

.diff-section {
  margin-bottom: 20px;
}

.section-title {
  margin-bottom: 12px;
  padding-left: 8px;
  border-left-width: 4px;
  border-left-style: solid;
}

.section-title .n-text {
  font-size: 14px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.change-item {
  padding: 10px;
  border-radius: 6px;
  border-width: 1px;
  border-style: solid;
  transition: all 0.2s ease;
}

.change-item:hover {
  transform: scale(1.05);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

.change-item.nested {
  margin-left: 12px;
}

.change-field {
  margin-bottom: 8px;
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.field-name {
  font-size: 13px;
  font-weight: 600;
}

.field-type {
  font-size: 11px;
}

.change-values {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.old-value, .new-value {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.label {
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.value {
  padding: 6px 8px;
  border-radius: 4px;
  font-size: 12px;
  word-break: break-all;
  white-space: pre-wrap;
  max-height: 200px;
  overflow-y: auto;
}

.nested-item {
  margin-bottom: 16px;
}

.nested-item.child {
  margin-left: 24px;
  padding-left: 16px;
  border-left-width: 2px;
  border-left-style: dashed;
}

.nested-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  border-radius: 6px 6px 0 0;
  margin-bottom: 8px;
}

.nested-header.child {
  margin-bottom: 4px;
}

.nested-path-info {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.nested-path {
  font-size: 12px;
  max-width: 300px;
  text-overflow: ellipsis;
}

.nested-changes {
  margin-top: 8px;
}

.nested-children {
  margin-top: 12px;
}

/* 自定义滚动条样式 */
:deep(.n-scrollbar) {
  --n-scrollbar-width: 6px;
}

:deep(.n-scrollbar-rail) {
  background: transparent;
}

:deep(.n-scrollbar-rail:hover) {
  background: rgba(128, 128, 128, 0.1);
}

:deep(.n-scrollbar-bar) {
  border-radius: 3px;
}

/* 代码块样式优化 */
:deep(.n-text.n-text--code) {
  padding: 2px 4px;
  border-radius: 4px;
  font-size: 12px;
}

/* 空状态样式 */
.no-changes {
  display: flex;
  justify-content: center;
  padding: 20px 0;
}

/* 动画效果 */
.change-item-enter-active,
.change-item-leave-active {
  transition: all 0.3s ease;
}

.change-item-enter-from,
.change-item-leave-to {
  opacity: 0;
  transform: scale(1.05);
}


</style>
