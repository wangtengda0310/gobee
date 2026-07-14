<!--
  CustomParams - 自定义规则参数组件

  动态 key-value 编辑器，支持添加/删除参数
  不使用 StandardBaseRow 和 initDefaults
-->
<script setup lang="ts">
import {ref, computed} from "vue"
import {NButton, NInput} from "naive-ui"
import {Add, Remove} from '@vicons/ionicons5'
import {Icon} from '@vicons/utils'

const props = defineProps<{ params: { [p: string]: string } }>()

const newKey = ref('')
const newVal = ref('')
// 响应式追踪：delete/add params 属性后手动递增触发 computed 重计算
const paramsVersion = ref(0)

const paramEntries = computed(() => {
  paramsVersion.value
  return Object.entries(props.params || {})
})

const deleteParam = (key: string) => {
  delete props.params[key]
  paramsVersion.value++
}

const addParam = () => {
  if (!newKey.value) return
  if (newKey.value in (props.params || {})) return
  props.params[newKey.value] = newVal.value
  paramsVersion.value++
  newKey.value = ''
  newVal.value = ''
}
</script>

<template>
  <div style="flex: 1 1 0; display: flex; flex-direction: column; justify-content: space-between; gap: 10px">
    <!-- 已有参数列表 -->
    <div
      v-for="([key, value]) in paramEntries"
      :key="key"
      style="display: flex; justify-content: space-between; gap: 10px"
    >
      <n-input
        style="flex: 1 1 0"
        placeholder="参数名"
        :value="key"
        disabled
      />
      <n-input
        style="flex: 3 1 0"
        placeholder="参数值"
        :value="value"
        @update:value="params[key] = $event"
      />
      <n-button
        style="flex: 0 0 0"
        type="error"
        @click="deleteParam(key)"
      >
        <Icon><Remove /></Icon>
      </n-button>
    </div>

    <!-- 添加新参数 -->
    <div style="display: flex; justify-content: space-between; gap: 10px">
      <n-input
        style="flex: 1 1 0"
        placeholder="新参数名"
        :value="newKey"
        @update:value="newKey = $event"
      />
      <n-input
        style="flex: 3 1 0"
        placeholder="新参数值"
        :value="newVal"
        @update:value="newVal = $event"
      />
      <n-button
        style="flex: 0 0 0"
        type="info"
        @click="addParam"
      >
        <Icon><Add /></Icon>
      </n-button>
    </div>
  </div>
</template>
