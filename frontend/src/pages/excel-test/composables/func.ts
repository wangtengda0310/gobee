/**
 * Excel 检查功能函数
 *
 * 提供加载、保存和执行 Excel 规则检查的核心功能
 */
import {ExcelCheckService} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/excel-test";
import {dataRef} from "./use-tree-and-history";
import {excelCheckManagerListMap, excelCheckMap, excelTableRulesMap, nowCheckData, nowSheetData, nowTableRules} from "./use-excel-check-data";
import {ManagerList, SheetColRule, SheetRule, TableRule, ColCheckResult, TableCheckResult, ColRule} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule";
import {checkLog, tableCheckResults, checking} from "./use-excel-check-log";
import {ExcelCaseDir, ExcelResourceDir, ClientPath} from "./option";
import {MessageApiInjection} from "naive-ui/es/message/src/MessageProvider";
import {ref} from "vue";

/** 当前激活的页签名称，执行检查时自动切换到"执行日志" */
export const activeTab = ref('option')

// ==================== 通用辅助函数 ====================

/**
 * 收集所有有配置数据的 Sheet 名称（列级规则 + 表级规则 + 负责人）
 */
const collectAllSheetKeys = (): string[] => {
    const allKeys = new Set<string>()
    excelCheckMap.value.forEach((_, key) => allKeys.add(key))
    excelTableRulesMap.value.forEach((_, key) => allKeys.add(key))
    excelCheckManagerListMap.value.forEach((_, key) => allKeys.add(key))
    return Array.from(allKeys)
}

/**
 * 构建 SheetRule 列表
 *
 * 同时将全局 ClientPath 注入到所有 RESOURCE 规则的 params 中，
 * 使得后端 RESOURCE 规则可以通过 clientPath 参数查找资源文件
 */
const buildSheetRules = (sheetNames: string[]): SheetRule[] => {
    return sheetNames.map(key => {
        const rules = excelCheckMap.value.get(key) || {}
        // 深拷贝规则，将 clientPath 注入到 RESOURCE 规则的 params 中
        const injectedRules: { [k: string]: SheetColRule | null } = {}
        for (const [attrName, colRule] of Object.entries(rules)) {
            const rule = colRule as SheetColRule | null | undefined
            if (!rule) {
                injectedRules[attrName] = null
                continue
            }
            const needsInjection = rule.PropRules?.some((r: ColRule | null) => r?.Type === 'RESOURCE')
            if (!needsInjection) {
                injectedRules[attrName] = rule
                continue
            }
            // 深拷贝并注入 clientPath
            injectedRules[attrName] = {
                PropName: rule.PropName,
                PropType: rule.PropType,
                PropRules: rule.PropRules.map((r: ColRule | null) => {
                    if (!r || r.Type !== 'RESOURCE') return r
                    return { ...r, Params: { ...r.Params, clientPath: ClientPath.value } }
                })
            }
        }
        return {
            sheet: key,
            managerList: excelCheckManagerListMap.value.get(key) ?? {
                QA: [],
                Designer: [],
                Programmer: []
            },
            rules: injectedRules,
            tableRules: excelTableRulesMap.value.get(key) || []
        }
    })
}

/**
 * 统一的检查前置准备：清空结果、设置状态、切换到日志页签
 */
const prepareCheck = () => {
    checkLog.value = []
    tableCheckResults.value = []
    checking.value = true
    activeTab.value = 'report'
}

/**
 * 统一的检查后置处理：写入结果、统计失败数、显示消息
 */
const handleCheckResult = (
    colResults: (ColCheckResult | null)[],
    tableResults: TableCheckResult[],
    message: MessageApiInjection | undefined,
    context: string
) => {
    checkLog.value = colResults
    tableCheckResults.value = tableResults

    const failCount = colResults.filter(r => r && !r.Ok).length +
        tableResults.filter(r => r && !r.ok).length

    if (failCount > 0) {
        message?.error(`${context}完成，发现 ${failCount} 个错误`)
    } else {
        message?.success(`${context}完成，未发现错误`)
    }
}

// ==================== 加载/保存 ====================

/**
 * 加载 Excel 文件和规则配置
 *
 * 从配置目录加载所有 Excel 文件结构，并加载对应的检查规则
 */
export const loadExcelsAndExcelRules = () => {
    // 清空检查日志，避免旧日志干扰
    checkLog.value = []
    tableCheckResults.value = []

    ExcelCheckService.GetAllExcelsConcurrent(ExcelResourceDir.value).then(res => {
        Object.assign(dataRef, res.map(excelFile => {
            if (!excelFile) {
                console.error('excel为空')
                return
            }
            const path = excelFile.Path.replaceAll(/[\\/]/g, "/").split('/')
            const name = path[path.length - 1].replace('".xlsx', '')
            return {
                label: name,
                key: crypto.randomUUID(),
                isLeaf: false,
                // 额外标注
                fullPath: excelFile.Path,
                levelType: 'Excel',
                children: excelFile.Sheets.map(s => {
                    if (!s) {
                        console.error('sheet为空')
                        return
                    }
                    return {
                        label: s.Name,
                        key: crypto.randomUUID(),
                        isLeaf: true,
                        levelType: 'Sheet',
                        // case信息
                        sheetType: s.SheetType,
                        sheetHeader: s.Header,
                        sheetError: s.Error,
                    }
                })
            }
        }).sort((a, b) => {
            if (!a) return -1
            if (!b) return 1
            // 先分割一下
            return a.label.localeCompare(b.label)
        }))
        console.log("excel读取", res)
    })
    ExcelCheckService.GetAllExcelRules(ExcelCaseDir.value).then(res => {
        // 先转为map[sheet]rule看看, 保持对r.ColRule的引用, r.Sheet r.FileName为不可更改
        excelCheckMap.value = new Map<string, { [key: string]: SheetColRule | null }>()
        excelCheckManagerListMap.value = new Map<string, ManagerList>()
        excelTableRulesMap.value = new Map<string, TableRule[]>()
        res.map(r => {
            if (!r) {
                console.error('sheet为空')
                return
            }
            excelCheckMap.value.set(r.sheet, r.rules as { [key: string]: SheetColRule | null })
            excelCheckManagerListMap.value.set(r.sheet, r.managerList || {
                QA: [],
                Designer: [],
                Programmer: []
            })
            // 加载表级规则 - 过滤掉 null
            excelTableRulesMap.value.set(r.sheet, (r.tableRules || []).filter((tr): tr is TableRule => tr !== null))
        })
        console.log("excel check读取", excelCheckMap.value)
        console.log("excel table rules读取", excelTableRulesMap.value)
        // 刷新当前选中 Sheet 的引用，避免 nowCheckData/nowTableRules 仍指向旧 Map 中的对象
        if (nowSheetData.value?.label) {
            const sheetName = nowSheetData.value.label
            nowCheckData.value = excelCheckMap.value.get(sheetName) || {}
            nowTableRules.value = excelTableRulesMap.value.get(sheetName) || []
        }
    }).catch(err => {
        console.log(err)
    })
}

/**
 * 保存 Excel 规则配置
 *
 * 将当前内存中的规则配置保存到用例目录
 */
export const saveExcelRules = () => {
    console.log(excelCheckMap.value)
    const checkList = buildSheetRules(collectAllSheetKeys())
    console.log(checkList)
    ExcelCheckService.SaveAllExcelRules(ExcelCaseDir.value, checkList).then(res => {
        console.log(res)
    }).catch(err => {
        console.log(err)
    })
}

// ==================== 统一检查入口 ====================

/**
 * Excel 检查执行入口（统一实现）
 *
 * 三种模式共用同一套前置准备、结果处理逻辑，避免各入口实现差异导致 bug：
 * - 'all'    : 检查所有已配置的 Sheet（菜单"执行检查"）
 * - 'sheets' : 检查指定 Sheet 列表（右键"执行分类"）
 * - 'column' : 检查单个列的列级规则（字段卡片悬浮"执行检查"按钮）
 *
 * 所有模式都会：
 * 1. 防重入检查（checking.value）
 * 2. 清空旧结果（checkLog + tableCheckResults）
 * 3. 切换到执行日志页签
 * 4. 调用对应后端接口
 * 5. 写入新结果并显示消息
 */
export const runExcelCheck = (options: {
    mode: 'all' | 'sheets' | 'column'
    sheetNames?: string[]
    sheetName?: string
    attrName?: string
    colRule?: SheetColRule
}, message?: MessageApiInjection) => {
    if (checking.value) {
        console.log('[Check] 正在执行中，忽略重复请求')
        return
    }

    const {mode, sheetNames, sheetName, attrName, colRule} = options

    // 入参校验
    if (mode === 'column') {
        if (!sheetName || !attrName || !colRule) {
            message?.error('参数缺失：sheetName/attrName/colRule')
            return
        }
        if (!colRule.PropRules || colRule.PropRules.length === 0) {
            message?.warning('该字段没有配置校验规则')
            return
        }
    } else if (mode === 'sheets') {
        if (!sheetNames || sheetNames.length === 0) {
            message?.error('没有指定要检查的 Sheet')
            return
        }
    }

    // 统一前置准备
    prepareCheck()

    switch (mode) {
        case 'all': {
            const checkList = buildSheetRules(collectAllSheetKeys())
            console.log('[Check All]', checkList)

            ExcelCheckService.CheckAllExcelRules(ExcelResourceDir.value, checkList).then(res => {
                console.log('[Check Result]', res)
                if (res) {
                    handleCheckResult(
                        res.colResults ?? [],
                        (res.tableResults ?? []).filter((r): r is TableCheckResult => r !== null),
                        message,
                        '检查'
                    )
                }
            }).catch(err => {
                console.log('[Check Err]', err)
                message?.error('检查执行失败: ' + err)
            }).finally(() => {
                checking.value = false
            })
            break
        }

        case 'sheets': {
            const checkList = buildSheetRules(sheetNames!)
            if (checkList.length === 0) {
                message?.warning('指定的 Sheet 没有配置检查规则')
                checking.value = false
                return
            }

            console.log('[Check Sheets]', sheetNames, checkList)
            ExcelCheckService.CheckAllExcelRules(ExcelResourceDir.value, checkList).then(res => {
                console.log('[Check Result]', res)
                if (res) {
                    // 只保留指定 Sheet 的结果
                    const filteredCol = (res.colResults ?? []).filter(
                        (r): r is ColCheckResult => r != null && sheetNames!.includes(r.SheetName || '')
                    )
                    const filteredTable = (res.tableResults ?? []).filter(
                        (r): r is TableCheckResult => r != null && sheetNames!.includes(r.sheetName || '')
                    )
                    handleCheckResult(filteredCol, filteredTable, message, '分类检查')
                }
            }).catch(err => {
                console.log('[Check Err]', err)
                message?.error('检查执行失败: ' + err)
            }).finally(() => {
                checking.value = false
            })
            break
        }

        case 'column': {
            console.log('[Check Column]', sheetName, attrName)
            ExcelCheckService.CheckSingleColumn(ExcelResourceDir.value, sheetName!, attrName!, colRule!).then(result => {
                console.log('[Check Column Result]', result)
                if (result) {
                    // 单列检查不涉及表级规则，tableResults 传空数组
                    handleCheckResult([result], [], message, `字段 [${attrName}] 检查`)
                }
            }).catch(err => {
                console.log('[Check Column Err]', err)
                message?.error('字段检查执行失败: ' + err)
            }).finally(() => {
                checking.value = false
            })
            break
        }
    }
}

// ==================== 兼容旧入口（委托到统一入口）====================

/**
 * 执行所有 Excel 规则检查（菜单"执行检查"）
 * @deprecated 使用 runExcelCheck({mode: 'all'})
 */
export const startExcelRules = () => {
    runExcelCheck({mode: 'all'})
}

/**
 * 执行指定 Sheet 列表的规则检查（右键"执行分类"）
 * @deprecated 使用 runExcelCheck({mode: 'sheets', sheetNames})
 */
export const startExcelRulesForSheets = (sheetNames: string[], message?: MessageApiInjection) => {
    runExcelCheck({mode: 'sheets', sheetNames}, message)
}

/**
 * 执行单个列的规则检查（字段卡片悬浮"执行检查"按钮）
 * @deprecated 使用 runExcelCheck({mode: 'column', sheetName, attrName, colRule})
 */
export const startExcelRulesForColumn = (sheetName: string, attrName: string, colRule: SheetColRule, message?: MessageApiInjection) => {
    runExcelCheck({mode: 'column', sheetName, attrName, colRule}, message)
}
