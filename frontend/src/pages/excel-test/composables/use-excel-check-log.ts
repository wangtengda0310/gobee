/**
 * Excel 检查日志状态管理
 *
 * 管理列级检查结果和表级检查结果
 */
import {ref} from "vue";
import {ColCheckResult, TableCheckResult} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule";

/** 列级检查结果 */
export const checkLog = ref<(ColCheckResult | null)[]>()

/** 表级检查结果 - 不包含 null */
export const tableCheckResults = ref<TableCheckResult[]>([])

/** 是否正在执行检查 */
export const checking = ref(false)
