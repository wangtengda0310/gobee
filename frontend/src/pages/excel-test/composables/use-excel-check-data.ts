/**
 * Excel 检查数据状态管理
 *
 * 管理当前选中的 Sheet 数据、检查规则映射和表级规则状态
 */
import {TreeOption} from "naive-ui";
import {ref} from "vue";
import {ExtraExcelTreeOption} from "./use-tree";
import {ManagerList, SheetColRule, TableRule, TableRuleMeta} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule";
import {ExcelCheckService} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/excel-test";

/** 配置面板显示状态（原 modals.ts） */
export const showExcelCaseOptionModal = ref(false)

/** 当前选中的 Sheet 数据 */
export const nowSheetData = ref<TreeOption & ExtraExcelTreeOption>()

/** Excel 检查规则映射 - key 为 Sheet 名称 */
export const excelCheckMap = ref<Map<string, { [key: string]: SheetColRule | null }>>(new Map<string, {
    [key: string]: SheetColRule | null
}>())

/** 负责人列表映射 - key 为 Sheet 名称 */
export const excelCheckManagerListMap = ref<Map<string, ManagerList>>(new Map<string, ManagerList>())

/** 当前检查数据 */
export const nowCheckData = ref<{ [key: string]: SheetColRule | null }>({})

/** 表级规则映射 - key 为 Sheet 名称 */
export const excelTableRulesMap = ref<Map<string, TableRule[]>>(new Map<string, TableRule[]>())

/** 当前 Sheet 的表级规则 */
export const nowTableRules = ref<TableRule[]>([])

/** 所有表级规则元数据（从后端获取） */
export const allTableRuleMetas = ref<TableRuleMeta[]>([])

/** 是否已加载表级规则元数据 */
export const tableRuleMetasLoaded = ref(false)

/** 加载所有表级规则元数据 */
export async function loadAllTableRuleMetas() {
    if (tableRuleMetasLoaded.value) {
        return
    }
    try {
        const result = await ExcelCheckService.GetAllTableRuleMetas()
        allTableRuleMetas.value = (result || []) as TableRuleMeta[]
        tableRuleMetasLoaded.value = true
    } catch (e) {
        console.error("加载表级规则元数据失败:", e)
    }
}

/** 根据 Sheet 名称加载适用的表级规则元数据（后端过滤） */
export async function loadTableRuleMetasForSheet(sheetName: string) {
    if (!sheetName) {
        return
    }
    try {
        const result = await ExcelCheckService.GetTableRuleMetasForSheet(sheetName)
        allTableRuleMetas.value = (result || []) as TableRuleMeta[]
        tableRuleMetasLoaded.value = true
    } catch (e) {
        console.error("加载表级规则元数据失败:", e)
    }
}

/** 创建空数据树 */
export function createData(): (TreeOption & ExtraExcelTreeOption)[] {
    return []
}
