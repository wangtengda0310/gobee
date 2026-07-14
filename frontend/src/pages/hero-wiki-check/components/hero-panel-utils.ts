/**
 * 武将面板工具函数
 *
 * 包含武将 Wiki 检查页面特有的样式映射函数，
 * 通用格式化函数从 shared 导入
 */
export {formatArray, formatBoolean} from "@shared/composables/use-format-utils"

/** 国家名称 -> 颜色样式映射（文字色、背景色） */
export const getCountryStyle = (country: string) => {
  const styles: Record<string, { color: string; bgColor: string }> = {
    '魏': {color: '#b3442c', bgColor: '#fef2e9'},
    '蜀': {color: '#2e7d32', bgColor: '#e8f5e9'},
    '吴': {color: '#1565c0', bgColor: '#e3f2fd'},
    '群': {color: '#6d4c41', bgColor: '#efebe9'},
    '晋': {color: '#7b1fa2', bgColor: '#f3e5f5'},
    '神': {color: '#c2185b', bgColor: '#fce4ec'},
  };
  return styles[country] || {color: '#666', bgColor: '#f5f5f5'};
}

/** 英雄类型 -> 颜色样式映射（文字色、背景色） */
export const getHeroTypeStyle = (type: string) => {
  const styles: Record<string, { color: string; bgColor: string }> = {
    '力量': {color: '#d32f2f', bgColor: '#ffebee'},
    '敏捷': {color: '#388e3c', bgColor: '#e8f5e9'},
    '智力': {color: '#1976d2', bgColor: '#e3f2fd'},
    '全能': {color: '#f57c00', bgColor: '#fff3e0'},
  };
  return styles[type] || {color: '#666', bgColor: '#f5f5f5'};
}
