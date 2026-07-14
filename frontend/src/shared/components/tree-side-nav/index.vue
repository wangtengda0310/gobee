<!--
  TreeSideNav - 带搜索的树形导航侧边栏

  通用左侧导航组件，封装搜索框 + n-tree + 右键菜单。
  参考 excel-test/index.vue 的 sider 区域封装。
-->
<script setup lang="ts">
import {type TreeOption, type TreeDropInfo, type DropdownOption} from 'naive-ui'

defineProps<{
  /** 树数据 */
  treeData: TreeOption[]
  /** 搜索关键词 */
  pattern?: string
  /** 展开的节点 */
  expandedKeys?: string[]
  /** 选中的节点 */
  checkedKeys?: string[]
  /** 选中策略 */
  checkStrategy?: 'all' | 'parent' | 'child'
  /** 级联选中 */
  cascade?: boolean
  /** 是否可拖拽 */
  draggable?: boolean
  /** 是否显示加载状态 */
  loading?: boolean
  /** 节点属性函数 */
  nodeProps?: (info: { option: TreeOption }) => Record<string, any>
  /** 自定义标签渲染 */
  renderLabel?: (info: { option: TreeOption }) => any
  /** 右键菜单选项 */
  dropdownOptions?: DropdownOption[]
  /** 右键菜单是否显示 */
  dropdownShow?: boolean
  /** 右键菜单 X 坐标 */
  dropdownX?: number
  /** 右键菜单 Y 坐标 */
  dropdownY?: number
  /** 是否显示不相关节点 */
  showIrrelevantNodes?: boolean
}>()

const emit = defineEmits<{
  'update:pattern': [value: string]
  'update:expandedKeys': [value: string[]]
  'update:checkedKeys': [value: string[]]
  'drop': [info: TreeDropInfo]
  'node-click': [node: TreeOption]
  'dropdown-select': [key: string | number, option: DropdownOption]
  'dropdown-clickoutside': [e: MouseEvent]
  'load': [node: TreeOption]
}>()
</script>

<template>
  <div class="tree-side-nav">
    <!-- 搜索区域 -->
    <div class="search-area">
      <n-input
        :value="pattern"
        placeholder="搜索"
        @update:value="(v: string) => emit('update:pattern', v)"
      />
      <!-- 搜索框下方的额外控件（如开关） -->
      <slot name="search-extra"/>
    </div>

    <!-- 树区域 -->
    <div class="tree-area">
      <n-spin :show="loading ?? false">
        <n-tree
          block-line
          :draggable="draggable ?? false"
          :data="treeData"
          :pattern="pattern"
          :expanded-keys="expandedKeys"
          :checked-keys="checkedKeys"
          :check-strategy="checkStrategy"
          :cascade="cascade"
          :show-irrelevant-nodes="showIrrelevantNodes"
          :node-props="nodeProps"
          :render-label="renderLabel"
          selectable
          expand-on-click
          :on-load="(node: TreeOption) => emit('load', node)"
          @drop="(info: TreeDropInfo) => emit('drop', info)"
          @update:checked-keys="(keys: string[]) => emit('update:checkedKeys', keys)"
          @update:expanded-keys="(keys: string[]) => emit('update:expandedKeys', keys)"
        />
      </n-spin>

      <!-- 右键菜单 -->
      <n-dropdown
        trigger="manual"
        placement="bottom-start"
        :show="dropdownShow ?? false"
        :options="dropdownOptions"
        :x="dropdownX ?? 0"
        :y="dropdownY ?? 0"
        @select="(key: string | number, option: DropdownOption) => emit('dropdown-select', key, option)"
        @clickoutside="(e: MouseEvent) => emit('dropdown-clickoutside', e)"
      />
    </div>
  </div>
</template>

<style scoped>
.tree-side-nav {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.search-area {
  padding: 8px 12px;
  border-bottom: 1px solid #333;
}

.tree-area {
  flex: 1;
  overflow: auto;
  padding: 0 4px;
}
</style>
