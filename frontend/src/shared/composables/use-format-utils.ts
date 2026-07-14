/**
 * 通用格式化工具函数
 *
 * 提供 array/boolean/ItemArray 的格式化显示，
 * hero-wiki-check 和 activity-wiki-check 页面共用
 */

/** 格式化数组显示，空数组或 undefined 时返回默认值 */
export const formatArray = (arr: any[] | undefined, defaultValue: string = '无'): string => {
  if (!arr || arr.length === 0) return defaultValue
  return arr.join(', ')
}

/** 格式化布尔值为文字显示 */
export const formatBoolean = (value: boolean | undefined, trueText: string = '是', falseText: string = '否'): string => {
  return value ? trueText : falseText
}

/** 格式化道具数组显示 (如 [{ItemId:1, Count:2}]) */
export const formatItemArray = (items: { ItemId: number; Count: number }[] | undefined, defaultValue: string = '无'): string => {
  if (!items || items.length === 0) return defaultValue
  return items.map(item => `${item.ItemId}×${item.Count}`).join(', ')
}
