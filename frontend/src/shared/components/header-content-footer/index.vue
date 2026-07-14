<!--
  HeaderContentFooter - 通用四区域页面布局组件

  提供标准的 Header + Sider + Content + Footer 布局结构。
  全部通过 slot 传入内容，不持有业务状态。

  布局结构：
  ┌───────────────────────────────────────────────┐
  │  Header (34px, slot#header)                    │
  ├──────────┬────────────────────────────────────┤
  │          │                                     │
  │  Sider   │  Content (slot#content)             │
  │  (240px) │                                     │
  │          │                                     │
  ├──────────┴────────────────────────────────────┤
  │  Footer (64px, slot#footer)                    │
  └───────────────────────────────────────────────┘
-->
<script setup lang="ts">
import {ref} from 'vue'

defineProps<{
  /** 顶部高度，默认 34px */
  headerHeight?: string
  /** 底部高度，默认 64px */
  footerHeight?: string
  /** 侧边栏宽度，默认 240px */
  siderWidth?: number
  /** 是否反转颜色 */
  inverted?: boolean
}>()

/** 侧边栏折叠状态 */
const siderCollapsed = ref(false)
</script>

<template>
  <div class="header-content-footer">
    <n-layout position="absolute">
      <!-- 顶部区域 -->
      <n-layout-header
        :style="{height: headerHeight ?? '34px', display: 'flex', alignItems: 'center'}"
        :inverted="inverted ?? false"
        bordered
      >
        <slot name="header"/>
      </n-layout-header>

      <!-- 主体区域 -->
      <n-layout
        position="absolute"
        :style="{top: headerHeight ?? '34px', bottom: footerHeight ?? '64px'}"
        has-sider
      >
        <!-- 侧边栏 -->
        <n-layout-sider
          v-if="$slots.sider"
          bordered
          show-trigger
          collapse-mode="width"
          :collapsed-width="50"
          :width="siderWidth ?? 240"
          :native-scrollbar="false"
          :inverted="inverted ?? false"
          :show-collapsed-content="false"
          v-model:collapsed="siderCollapsed"
        >
          <slot name="sider"/>
        </n-layout-sider>

        <!-- 内容区域 -->
        <n-layout>
          <slot name="content"/>
        </n-layout>
      </n-layout>

      <!-- 弹窗容器（不在布局流中） -->
      <slot name="modals"/>

      <!-- 底部区域 -->
      <n-layout-footer
        class="hcf-footer"
        position="absolute"
        :inverted="inverted ?? false"
        bordered
        :style="{height: footerHeight ?? '64px', padding: '0 20px', display: 'flex', alignItems: 'center', gap: '20px'}"
      >
        <slot name="footer"/>
      </n-layout-footer>
    </n-layout>
  </div>
</template>

<style scoped>
.header-content-footer {
  position: relative;
  width: 100%;
  height: 100%;
  background-color: #2b2b2b;
  color: white;
}

.hcf-footer :deep(.n-statistic .n-statistic-value) {
  margin-top: 0;
}
</style>
