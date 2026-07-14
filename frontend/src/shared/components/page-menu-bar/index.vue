<!--
  PageMenuBar - 页面菜单栏共用组件

  通用顶部菜单栏，支持水平菜单和右侧扩展区域。
  参考 excel-test/index.vue 的 header 区域封装。
-->
<script setup lang="ts">
import {NMenu, type MenuOption} from 'naive-ui'

defineProps<{
  /** 菜单选项 */
  options: MenuOption[]
  /** 当前激活的菜单项 */
  activeKey?: string | null
  /** 是否反转颜色 */
  inverted?: boolean
}>()

const emit = defineEmits<{
  'update:activeKey': [value: string | null]
}>()
</script>

<template>
  <div class="page-menu-bar">
    <n-menu
      mode="horizontal"
      :options="options"
      :value="activeKey ?? undefined"
      :inverted="inverted ?? false"
      @update:value="(val: string) => emit('update:activeKey', val)"
    />
    <div v-if="$slots.right" class="right-area">
      <slot name="right"/>
    </div>
  </div>
</template>

<style scoped>
.page-menu-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  height: 100%;
}

.right-area {
  display: flex;
  align-items: center;
  padding-right: 12px;
  flex-shrink: 0;
}
</style>
