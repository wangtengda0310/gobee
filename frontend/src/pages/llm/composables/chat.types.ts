/**
 * 聊天功能类型定义
 */

// 消息角色
export type MessageRole = 'user' | 'assistant'

// 对话消息
export interface ChatMessage {
  role: MessageRole
  content: string
  timestamp: number  // Unix 时间戳（秒）
}

// 对话会话
export interface ChatSession {
  id: string
  messages: ChatMessage[]
  createdAt: number
  updatedAt: number
}

// Anthropic 配置
export interface AnthropicConfig {
  apiKey: string
  baseUrl: string
  model: string
  maxTokens: number
}

// OpenAI 配置
export interface OpenAIConfig {
  apiKey: string
  baseUrl: string
  model: string
  maxTokens: number
}

// 聊天配置
export interface ChatConfig {
  provider: 'anthropic' | 'openai'
  anthropicConfig: AnthropicConfig
  openaiConfig: OpenAIConfig
  systemPrompt: string
}

// 默认配置
export const defaultChatConfig: ChatConfig = {
  provider: 'anthropic',
  anthropicConfig: {
    apiKey: '',
    baseUrl: 'https://api.anthropic.com/v1',
    model: 'claude-3-5-sonnet-20241022',
    maxTokens: 4096
  },
  openaiConfig: {
    apiKey: '',
    baseUrl: 'https://api.openai.com/v1',
    model: 'gpt-4o',
    maxTokens: 4096
  },
  systemPrompt: '你是一个QA测试助手。你可以使用提供的工具来帮助用户检查Excel配表、查询数据、执行测试等。当用户提出与配表相关的问题时，请主动使用工具获取数据后再回答，不要凭猜测回答。'
}

// 模型选项
export const anthropicModels = [
  { label: 'Claude 3.5 Sonnet', value: 'claude-3-5-sonnet-20241022' },
  { label: 'Claude 3.5 Haiku', value: 'claude-3-5-haiku-20241022' },
  { label: 'Claude 3 Opus', value: 'claude-3-opus-20240229' },
  // 智谱AI (Base URL: https://open.bigmodel.cn/api/anthropic/v1)
  { label: 'GLM-5.1 (智谱) 推荐', value: 'glm-5.1' },
  { label: 'GLM-5 (智谱)', value: 'glm-5' },
  { label: 'GLM-4-Flash (智谱)', value: 'glm-4-flash' }
]

export const openaiModels = [
  // OpenAI 模型
  { label: 'GPT-4o', value: 'gpt-4o' },
  { label: 'GPT-4o Mini', value: 'gpt-4o-mini' },
  { label: 'GPT-4 Turbo', value: 'gpt-4-turbo' },
  // 智谱AI 模型 (Base URL: https://open.bigmodel.cn/api/paas/v4)
  { label: 'GLM-4-Flash (智谱)', value: 'glm-4-flash' },
  { label: 'GLM-4-Plus (智谱)', value: 'glm-4-plus' },
  { label: 'GLM-4 (智谱)', value: 'glm-4' }
]
