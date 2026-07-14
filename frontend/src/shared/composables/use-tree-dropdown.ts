/**
 * 通用树节点下拉菜单管理
 *
 * 管理树节点的右键下拉菜单状态，
 * excel-test 和 function-test 页面共用
 */
import type {DropdownOption} from 'naive-ui'
import {ref} from 'vue'

/** 下拉菜单选项类型 */
export type TreeDDOption = {}

/** 下拉菜单显示状态 */
export const showDropdownRef = ref(false)

/** 下拉菜单选项 */
export const optionsRef = ref<(DropdownOption & TreeDDOption)[]>([])

/** 下拉菜单 X 坐标 */
export const xRef = ref(0)

/** 下拉菜单 Y 坐标 */
export const yRef = ref(0)

/** 处理下拉菜单选项选择 */
export function handleSelect(key: string | number, option: DropdownOption) {
    showDropdownRef.value = false
}

/** 处理点击下拉菜单外部 */
export function handleClickOutside(e: MouseEvent) {
    showDropdownRef.value = false
}
