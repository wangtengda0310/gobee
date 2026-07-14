/**
 * 资源预览组件
 *
 * 根据资源文件存在状态展示三种样式：
 * - 存在（图片文件）：显示缩略图，点击放大
 * - 不存在：红色中划线文字，悬停显示完整路径
 * - 路径为空：灰色占位符
 */
<script setup lang="ts">
import {computed} from "vue"
import type {ResourceStatus} from "../composables/use-resource-check"

const props = defineProps<{
  /** 资源路径值 */
  value: string | undefined | null
  /** 资源检查状态（由 useResourceCheck 提供） */
  status?: ResourceStatus
}>()

const isEmpty = computed(() => !props.value)
const exists = computed(() => props.status?.exists ?? false)
const hasPreview = computed(() => !!props.status?.previewUrl)
</script>

<template>
  <!-- 路径为空 -->
  <span v-if="isEmpty" class="resource-empty">-</span>

  <!-- 资源存在且有预览 -->
  <n-popover v-else-if="hasPreview" trigger="click" placement="left" :width="400">
    <template #trigger>
      <img
        :src="status!.previewUrl"
        :alt="value!"
        class="resource-preview-thumb"
        @click.stop
      />
    </template>
    <div class="resource-preview-full">
      <img :src="status!.previewUrl" :alt="value!" style="width: 100%; max-height: 400px; object-fit: contain;" />
      <div class="resource-preview-path">{{ value }}</div>
    </div>
  </n-popover>

  <!-- 资源存在但无预览（非图片文件） -->
  <n-tooltip v-else-if="exists" trigger="hover">
    <template #trigger>
      <n-text type="success" style="word-break: break-all;">{{ value }}</n-text>
    </template>
    {{ value }}
  </n-tooltip>

  <!-- 资源不存在 -->
  <n-tooltip v-else trigger="hover">
    <template #trigger>
      <span class="resource-missing">{{ value }}</span>
    </template>
    {{ value }}
  </n-tooltip>
</template>

<style scoped>
.resource-empty {
  color: #999;
}

.resource-missing {
  color: #d03050;
  text-decoration: line-through;
  word-break: break-all;
  cursor: help;
}

.resource-preview-thumb {
  max-width: 48px;
  max-height: 48px;
  object-fit: contain;
  border-radius: 4px;
  cursor: pointer;
  border: 1px solid rgba(255, 255, 255, 0.15);
  transition: transform 0.2s ease;
}

.resource-preview-thumb:hover {
  transform: scale(1.5);
  z-index: 10;
  position: relative;
}

.resource-preview-full {
  text-align: center;
}

.resource-preview-path {
  margin-top: 8px;
  font-size: 12px;
  color: #999;
  word-break: break-all;
}
</style>
