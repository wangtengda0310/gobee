<!--
  RegexFormatSelect - 正则格式选择器

  统一的正则提取格式选择组件，包含预设正则选项和自定义正则输入。
  被 AllBaseParams、EnumParams 等需要正则提取的规则参数组件复用。

  使用 params.regexMode 前端参数区分"自定义"和"无特定格式"：
  - regexMode 不存在 → 匹配预设选项，无匹配则回退到"无特定格式"
  - regexMode = 'custom' → 自定义模式，显示手动输入框
-->
<script setup lang="ts">
import {computed} from "vue"
import {NInput, NSelect} from "naive-ui"
import {ERuleParam} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
import {regexOptions, findOptionByRegex} from '../chain-reference-params'

/** 前端专用参数名，用于标记自定义正则模式（不影响后端） */
const REGEX_MODE_KEY = 'regexMode'

const props = defineProps<{ params: { [p: string]: string } }>()

const isCustom = computed(() => props.params[REGEX_MODE_KEY] === 'custom')

const selectedOption = computed(() => {
    if (isCustom.value) return -1
    return findOptionByRegex(props.params[ERuleParam.PATTERN], props.params[ERuleParam.GROUPS])
})

/** 切换选项时，预设选项直接设置 regex/groups，自定义选项标记模式并清空 */
function onSelect(_v: number, option: any) {
    if (option.value === -1) {
        props.params[REGEX_MODE_KEY] = 'custom'
        props.params[ERuleParam.PATTERN] = ''
        props.params[ERuleParam.GROUPS] = ''
    } else {
        delete props.params[REGEX_MODE_KEY]
        props.params[ERuleParam.PATTERN] = option.regex || ''
        props.params[ERuleParam.GROUPS] = option.groups || ''
    }
}
</script>

<template>
  <div style="display: flex; flex-direction: column; gap: 6px">
    <div style="display: flex; gap: 10px; align-items: center">
      <div style="flex: 0 0 auto; white-space: nowrap">正则格式:</div>
      <n-select
        style="flex: 1 1 0"
        placeholder="选择正则提取格式"
        :value="selectedOption"
        :options="regexOptions"
        @update:value="onSelect"
      />
    </div>
    <!-- 自定义正则输入（仅选中"自定义"时显示） -->
    <template v-if="isCustom">
      <div style="display: flex; gap: 10px; align-items: center">
        <div style="flex: 0 0 auto; white-space: nowrap">正则表达式:</div>
        <n-input
          style="flex: 1 1 0"
          placeholder="输入正则表达式（如 \{(\\d+);\\d+\}）"
          :value="params[ERuleParam.PATTERN] || ''"
          @update:value="params[ERuleParam.PATTERN] = $event"
        />
      </div>
      <div style="display: flex; gap: 10px; align-items: center">
        <div style="flex: 0 0 auto; white-space: nowrap">捕获组:</div>
        <n-input
          style="flex: 0 0 80px"
          placeholder="1"
          :value="params[ERuleParam.GROUPS] || ''"
          @update:value="params[ERuleParam.GROUPS] = $event"
        />
      </div>
    </template>
  </div>
</template>
