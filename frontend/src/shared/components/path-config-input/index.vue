<script setup lang="ts">
/**
 * PathConfigInput - 可复用的配置路径输入框组件
 *
 * @module shared/components/path-config-input
 * @description
 * 用于统一多个页面的路径配置输入界面，包含两个输入框：
 * - 第一个：Excel/配表目录（固定用途）
 * - 第二个：可动态配置用途
 *
 * 支持通过文件夹选择对话框选择路径。
 *
 * @example 基础使用（label模式）
 * ```vue
 * <PathConfigInput
 *   v-model:excel-dir="excelDir"
 *   v-model:second-value="caseDir"
 *   excel-label="配表"
 *   second-label="用例"
 * />
 * ```
 *
 * @example placeholder模式（无label）
 * ```vue
 * <PathConfigInput
 *   v-model:excel-dir="excelDir"
 *   v-model:second-value="jsonPath"
 *   excel-label=""
 *   second-label=""
 *   excel-placeholder="Excel 配置目录路径"
 *   second-placeholder="JSON 文件路径"
 * />
 * ```
 *
 * @example 带持久化回调
 * ```vue
 * <PathConfigInput
 *   v-model:excel-dir="excelDir"
 *   v-model:second-value="caseDir"
 *   :on-save="saveConfig"
 * />
 * ```
 *
 * @example flex布局模式（输入框自适应宽度）
 * ```vue
 * <PathConfigInput
 *   v-model:excel-dir="excelDir"
 *   v-model:second-value="cardDir"
 *   excel-label="配表位置"
 *   second-label="Card文件夹位置"
 *   layout="flex"
 * />
 * ```
 *
 * @props
 * - excelDir: string - Excel目录路径（v-model:excel-dir）
 * - excelLabel?: string - 第一个输入框的label，默认"配表"
 * - excelPlaceholder?: string - 第一个输入框的placeholder
 * - secondValue: string - 第二个输入框的值（v-model:second-value）
 * - secondLabel?: string - 第二个输入框的label
 * - secondPlaceholder?: string - 第二个输入框的placeholder
 * - onSave?: () => void - 失焦时的保存回调
 * - inputWidth?: string - 输入框宽度，默认"180px"（仅inline模式有效）
 * - size?: 'small' | 'medium' | 'large' - 尺寸，默认"small"
 * - layout?: 'inline' | 'flex' - 布局模式，默认"inline"
 */

import { Dialogs } from '@wailsio/runtime'
import { FolderOpenOutline } from '@vicons/ionicons5'

interface Props {
  // Excel 目录配置
  excelDir: string
  excelLabel?: string
  excelPlaceholder?: string

  // 第二个输入框配置
  secondValue: string
  secondLabel?: string
  secondPlaceholder?: string

  // 持久化回调
  onSave?: () => void | Promise<void>

  // 布局配置
  inputWidth?: string
  size?: 'small' | 'medium' | 'large'
  layout?: 'inline' | 'flex'
}

const props = withDefaults(defineProps<Props>(), {
  excelLabel: '配表',
  excelPlaceholder: '',
  secondLabel: '',
  secondPlaceholder: '',
  inputWidth: '180px',
  size: 'small',
  layout: 'inline'
})

const emit = defineEmits<{
  'update:excelDir': [value: string]
  'update:secondValue': [value: string]
}>()

// 处理失焦事件
const handleBlur = async () => {
  if (props.onSave) {
    try {
      await props.onSave()
    } catch (err) {
      console.error('保存配置失败:', err)
    }
  }
}

// 打开文件夹选择对话框
const selectFolder = async (type: 'excel' | 'second') => {
  try {
    const title = type === 'excel'
      ? `选择${props.excelLabel || '配表'}目录`
      : `选择${props.secondLabel || '目录'}`

    const result = await Dialogs.OpenFile({
      Title: title,
      CanChooseDirectories: true,
      CanChooseFiles: false
    })

    // OpenFile 返回 string[] 或 string，取第一个有效路径
    const path = Array.isArray(result) ? result[0] : result
    if (path && typeof path === 'string') {
      if (type === 'excel') {
        emit('update:excelDir', path)
      } else {
        emit('update:secondValue', path)
      }
      // 选择后触发保存
      await handleBlur()
    }
  } catch (err) {
    console.error('选择文件夹失败:', err)
  }
}
</script>

<template>
  <!-- inline模式：使用 n-space 水平排列 -->
  <n-space v-if="layout === 'inline'" align="center" :size="10" :wrap="false">
    <n-input-group>
      <n-input-group-label v-if="excelLabel" :size="size">
        {{ excelLabel }}
      </n-input-group-label>
      <n-input
        :value="excelDir"
        :placeholder="excelPlaceholder"
        :size="size"
        :style="{ width: inputWidth }"
        @update:value="(v: string) => emit('update:excelDir', v)"
        @blur="handleBlur"
      />
      <n-button :size="size" @click="selectFolder('excel')">
        <n-icon :size="14">
          <FolderOpenOutline />
        </n-icon>
      </n-button>
    </n-input-group>

    <n-input-group>
      <n-input-group-label v-if="secondLabel" :size="size">
        {{ secondLabel }}
      </n-input-group-label>
      <n-input
        :value="secondValue"
        :placeholder="secondPlaceholder"
        :size="size"
        :style="{ width: inputWidth }"
        @update:value="(v: string) => emit('update:secondValue', v)"
        @blur="handleBlur"
      />
      <n-button :size="size" @click="selectFolder('second')">
        <n-icon :size="14">
          <FolderOpenOutline />
        </n-icon>
      </n-button>
    </n-input-group>
  </n-space>

  <!-- flex模式：输入框自适应宽度（兼容HeroVoiceResourceCheck.vue） -->
  <template v-else-if="layout === 'flex'">
    <span style="flex: 0 0 70px">{{ excelLabel || '配表' }}:</span>
    <n-input-group style="flex: 1 1 0">
      <n-input
        style="flex: 1 1 0"
        :value="excelDir"
        :placeholder="excelPlaceholder"
        :size="size"
        @update:value="(v: string) => emit('update:excelDir', v)"
        @blur="handleBlur"
      />
      <n-button :size="size" @click="selectFolder('excel')">
        <n-icon :size="14">
          <FolderOpenOutline />
        </n-icon>
      </n-button>
    </n-input-group>
    <span style="flex: 0 0 115px">{{ secondLabel }}:</span>
    <n-input-group style="flex: 1 1 0">
      <n-input
        style="flex: 1 1 0"
        :value="secondValue"
        :placeholder="secondPlaceholder"
        :size="size"
        @update:value="(v: string) => emit('update:secondValue', v)"
        @blur="handleBlur"
      />
      <n-button :size="size" @click="selectFolder('second')">
        <n-icon :size="14">
          <FolderOpenOutline />
        </n-icon>
      </n-button>
    </n-input-group>
  </template>
</template>
