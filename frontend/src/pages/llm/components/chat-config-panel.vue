<script setup lang="ts">
/**
 * 聊天配置面板组件
 * 用于配置 API 提供商、API Key、模型等参数
 */
import { ref, watch } from 'vue'
import { NDrawer, NDrawerContent, NForm, NFormItem, NRadioGroup, NRadioButton, NInput, NSelect, NButton, NSpace, useMessage } from 'naive-ui'
import type { ChatConfig } from '../composables/chat.types'
import { defaultChatConfig, anthropicModels, openaiModels } from '../composables/chat.types'
import { ChatService } from '@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/settings/home'

const props = defineProps<{
  show: boolean
}>()

const emit = defineEmits<{
  (e: 'update:show', value: boolean): void
}>()

const message = useMessage()

// 本地配置状态
const localConfig = ref<ChatConfig>(JSON.parse(JSON.stringify(defaultChatConfig)))
const showAnthropicKey = ref(false)
const showOpenAIKey = ref(false)

// 监听 show 变化，加载配置
watch(() => props.show, async (newVal) => {
  if (newVal) {
    try {
      const config = await ChatService.GetConfig()
      if (config) {
        localConfig.value = config as ChatConfig
      }
    } catch (e) {
      console.error('加载配置失败:', e)
    }
  }
})

// 从 Claude Code 导入配置
async function handleImportFromClaudeCode() {
  try {
    const config = await ChatService.ImportFromClaudeCode()
    if (config) {
      localConfig.value = config as ChatConfig
      message.success('已从 Claude Code 导入配置')
    }
  } catch (e: any) {
    message.error('导入失败: ' + e.message)
  }
}

// 保存配置
async function handleSave() {
  try {
    await ChatService.SaveConfig(localConfig.value as any)
    message.success('配置已保存')
    emit('update:show', false)
  } catch (e: any) {
    message.error('保存失败: ' + e.message)
  }
}

// 取消
function handleCancel() {
  emit('update:show', false)
}
</script>

<template>
  <n-drawer
    :show="show"
    @update:show="emit('update:show', $event)"
    :width="400"
    placement="right"
  >
    <n-drawer-content title="API 配置" closable>
      <n-form label-placement="left" label-width="80">
        <!-- 提供商选择 -->
        <n-form-item label="提供商">
          <n-radio-group v-model:value="localConfig.provider">
            <n-radio-button value="anthropic">
              Anthropic
            </n-radio-button>
            <n-radio-button value="openai">
              OpenAI
            </n-radio-button>
          </n-radio-group>
        </n-form-item>

        <!-- Anthropic 配置 -->
        <template v-if="localConfig.provider === 'anthropic'">
          <n-form-item label="API Key">
            <n-input
              v-model:value="localConfig.anthropicConfig.apiKey"
              :type="showAnthropicKey ? 'text' : 'password'"
              placeholder="sk-ant-..."
            >
              <template #suffix>
                <n-button
                  text
                  @click="showAnthropicKey = !showAnthropicKey"
                >
                  {{ showAnthropicKey ? '🙈' : '👁️' }}
                </n-button>
              </template>
            </n-input>
          </n-form-item>
          <n-form-item label="Base URL">
            <n-input
              v-model:value="localConfig.anthropicConfig.baseUrl"
              placeholder="https://api.anthropic.com/v1"
            />
          </n-form-item>
          <n-form-item label="模型">
            <n-select
              v-model:value="localConfig.anthropicConfig.model"
              :options="anthropicModels"
              filterable
              tag
              placeholder="选择或输入模型名"
            />
          </n-form-item>
        </template>

        <!-- OpenAI 配置 -->
        <template v-if="localConfig.provider === 'openai'">
          <n-form-item label="API Key">
            <n-input
              v-model:value="localConfig.openaiConfig.apiKey"
              :type="showOpenAIKey ? 'text' : 'password'"
              placeholder="sk-..."
            >
              <template #suffix>
                <n-button
                  text
                  @click="showOpenAIKey = !showOpenAIKey"
                >
                  {{ showOpenAIKey ? '🙈' : '👁️' }}
                </n-button>
              </template>
            </n-input>
          </n-form-item>
          <n-form-item label="Base URL">
            <n-input
              v-model:value="localConfig.openaiConfig.baseUrl"
              placeholder="https://api.openai.com/v1"
            />
          </n-form-item>
          <n-form-item label="模型">
            <n-select
              v-model:value="localConfig.openaiConfig.model"
              :options="openaiModels"
              filterable
              tag
              placeholder="选择或输入模型名"
            />
          </n-form-item>
        </template>

        <!-- 系统提示 -->
        <n-form-item label="系统提示">
          <n-input
            v-model:value="localConfig.systemPrompt"
            type="textarea"
            :rows="3"
            placeholder="你是一个有帮助的AI助手。"
          />
        </n-form-item>
      </n-form>

      <template #footer>
        <n-space>
          <n-button @click="handleImportFromClaudeCode">
            从 Claude Code 导入
          </n-button>
          <n-button @click="handleCancel">取消</n-button>
          <n-button type="primary" @click="handleSave">保存</n-button>
        </n-space>
      </template>
    </n-drawer-content>
  </n-drawer>
</template>

<style scoped>
/* 配置面板样式 */
</style>
