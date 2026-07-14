/**
 * 通用树搜索状态管理
 *
 * 管理树节点的搜索过滤功能，
 * excel-test 和 function-test 页面共用
 */
import {ref} from "vue";

/** 搜索关键词 */
export const pattern = ref('')

/** 是否显示不相关节点 */
export const showIrrelevantNodes = ref(false)
