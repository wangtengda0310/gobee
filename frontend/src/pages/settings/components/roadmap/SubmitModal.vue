<script setup lang="ts">
/**
 * 提交新建议弹窗组件
 *
 * 功能：
 * - 表单输入（标题、描述、优先级）
 * - 表单验证
 * - 提交新建议
 */
import { ref, computed } from 'vue'
import { NModal, NForm, NFormItem, NInput, NRadioGroup, NRadio, NButton, useMessage } from 'naive-ui'
import type { Priority } from '../../config/roadmap-types'

const emit = defineEmits<{
  close: []
  submit: [data: { title: string; description: string; priority: Priority }]
}>()

const message = useMessage()

// 表单数据
const formData = ref({
  title: '',
  description: '',
  priority: 'medium' as Priority
})

// 表单验证
const formErrors = computed(() => {
  const errors: { title?: string; description?: string } = {}
  if (!formData.value.title.trim()) {
    errors.title = '请输入功能标题'
  } else if (formData.value.title.length > 50) {
    errors.title = '标题不能超过50个字符'
  }
  if (!formData.value.description.trim()) {
    errors.description = '请输入功能描述'
  } else if (formData.value.description.length < 10) {
    errors.description = '描述至少需要10个字符'
  }
  return errors
})

const isValid = computed(() => Object.keys(formErrors.value).length === 0)

// 提交
function handleSubmit() {
  if (!isValid.value) {
    message.warning('请完善表单信息')
    return
  }

  emit('submit', {
    title: formData.value.title.trim(),
    description: formData.value.description.trim(),
    priority: formData.value.priority
  })

  // 重置表单
  formData.value = {
    title: '',
    description: '',
    priority: 'medium'
  }
}

// 关闭
function handleClose() {
  emit('close')
}
</script>

<template>
  <n-modal
    :show="true"
    @update:show="handleClose"
    preset="card"
    title="提交新功能建议"
    style="width: 500px; max-width: 90vw"
    :bordered="false"
    :mask-closable="true"
  >
    <n-form label-placement="top">
      <n-form-item label="功能标题" :feedback="formErrors.title" :validation-status="formErrors.title ? 'error' : undefined">
        <n-input
          v-model:value="formData.title"
          placeholder="请输入功能标题"
          maxlength="50"
          show-count
        />
      </n-form-item>

      <n-form-item label="功能描述" :feedback="formErrors.description" :validation-status="formErrors.description ? 'error' : undefined">
        <n-input
          v-model:value="formData.description"
          type="textarea"
          placeholder="请详细描述功能需求，支持 Markdown 格式"
          :autosize="{ minRows: 4, maxRows: 10 }"
        />
      </n-form-item>

      <n-form-item label="优先级建议">
        <n-radio-group v-model:value="formData.priority">
          <n-radio value="low">低</n-radio>
          <n-radio value="medium">中</n-radio>
          <n-radio value="high">高</n-radio>
        </n-radio-group>
      </n-form-item>
    </n-form>

    <template #footer>
      <div class="modal-footer">
        <n-button @click="handleClose">取消</n-button>
        <n-button type="primary" :disabled="!isValid" @click="handleSubmit">提交</n-button>
      </div>
    </template>
  </n-modal>
</template>

<style scoped>
.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>
