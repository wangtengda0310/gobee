// variable-select.requirement.ts — 动态变量选择器接口封装
// 职责：获取可用变量列表、选择变量绑定到字段

import { GetAvailableVariables } from '@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/replaycontrolservice'

// ========== DTO ==========

/** 可用变量项 */
export interface VariableItem {
  /** 变量显示名（如 "城池ID"、"当前时间"） */
  display_name: string
  /** 变量短名（用于后端替换，如 "cityId"、"nowTime"） */
  short_name: string
  /** 对哪些 Req 可见（msg_name），undefined/空=对所有 Req 可用 */
  available_reqs?: string[]
}

// ========== Service 接口 ==========
export interface VariableService {
  /** 获取可用变量列表 */
  getAvailableVariables(): Promise<VariableItem[]>
}

// ========== Wails 实现 ==========
export function createWailsVariableService(): VariableService {
  return {
    async getAvailableVariables(): Promise<VariableItem[]> {
      const items = await GetAvailableVariables()
      if (!items) return []
      return items.map(item => ({
        display_name: item.display_name,
        short_name: item.short_name,
        available_reqs: item.available_reqs,
      }))
    },
  }
}

// ========== Mock 实现 ==========
export function createMockVariableService(): VariableService {
  return {
    async getAvailableVariables(): Promise<VariableItem[]> {
      // Mock：返回示例变量列表
      return [
        { display_name: '\u{1F3F0} \u{516C}\u{4F1A}\u{57CE}\u{6218} - \u{57CE}\u{6C60}ID', short_name: 'cityId' },
        { display_name: '\u{1F464} \u{5F53}\u{524D}\u{8D26}\u{53F7}', short_name: 'openid' },
        { display_name: '\u{1F4C5} \u{5F53}\u{524D}\u{65F6}\u{95F4}', short_name: 'nowTime' },
        { display_name: '\u{1F465} \u{4F1A}\u{5458}\u{6570}\u{91CF}', short_name: 'memberCount' },
        { display_name: '\u{2694}\u{FE0F} \u{6218}\u{6597}\u{6B21}\u{6570}', short_name: 'battleCount' },
        { display_name: '\u{1F3AE} \u{73A9}\u{5BB6}\u{7B49}\u{7EA7}', short_name: 'playerLevel' },
      ]
    },
  }
}

// ========== 表单选项格式（供 n-select 使用） ==========
export interface VarOption {
  label: string
  value: string
}

/** 将 VariableItem[] 转换为 n-select options 格式 */
export function toVarOptions(items: VariableItem[]): VarOption[] {
  return items.map(item => ({
    label: item.display_name,
    value: item.short_name,
  }))
}
