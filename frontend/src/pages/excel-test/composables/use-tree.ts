/**
 * Excel 树形结构管理
 *
 * 管理 Excel 文件树的展开、选中、拖拽和渲染逻辑
 * 基础状态从 shared 导入，保留 Excel 业务特有的拖拽规则和右键菜单
 */
import {DropdownOption, NTag, TreeDropInfo, TreeOption} from "naive-ui";
import {h, VNode} from "vue";
import {dataRef} from "./use-tree-and-history";
import {DialogApiInjection} from "naive-ui/es/dialog/src/DialogProvider";
import {MessageApiInjection} from "naive-ui/es/message/src/MessageProvider";
import {excelCheckManagerListMap, excelCheckMap, nowCheckData, nowSheetData, excelTableRulesMap, nowTableRules} from "./use-excel-check-data";
import {tableCheckResults} from "./use-excel-check-log";
import {SheetHeader, SheetType} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio";
import {
    createManagerToOneOption,
    createExecCateOption,
    createDividerOption
} from "../../../shared/composables/use-tree-context-menu";
import {openExcelFile} from "@shared/composables/use-open-excel";
import {
    optionsRef,
    showDropdownRef,
    xRef,
    yRef
} from "../../../shared/composables/use-tree-dropdown";
import {startExcelRulesForSheets} from "./func";
import {createTreeState} from "../../../shared/composables/use-tree-state";
import {findSiblingsAndIndex, markDropModified} from "../../../shared/composables/use-tree-utils";

// 创建独立的树形状态实例，避免与 function-test 页面共享状态
const treeState = createTreeState()
export const {expandedKeysRef, checkedKeysRef, checkStrategy, cascade, handleExpandedKeysChange, handleCheckedKeysChange} = treeState

/** 处理节点拖拽（Excel 特有规则：禁止拖入 Excel 节点内） */
export function handleDrop({node, dragNode, dropPosition}: TreeDropInfo) {
    if (dropPosition === 'inside' || node.levelType === 'Excel') return

    const [dragNodeSiblings, dragNodeIndex] = findSiblingsAndIndex(
        dragNode,
        dataRef
    )

    if (dragNodeSiblings === null || dragNodeIndex === null)
        return

    // 记录拖拽源分类
    let fromCate = ""
    dataRef.forEach((data: TreeOption & ExtraExcelTreeOption) => {
        if (data.children?.find(c => c.label == dragNode.label) && data.label) {
            fromCate = data.label
        }
    })

    dragNodeSiblings.splice(dragNodeIndex, 1)
    if (dropPosition === 'before' && (node.levelType == dragNode.levelType)) {
        const [nodeSiblings, nodeIndex] = findSiblingsAndIndex(node, dataRef)
        if (nodeSiblings === null || nodeIndex === null)
            return
        nodeSiblings.splice(nodeIndex, 0, dragNode)
    } else if (dropPosition === 'after' && (node.levelType == dragNode.levelType)) {
        const [nodeSiblings, nodeIndex] = findSiblingsAndIndex(node, dataRef)
        if (nodeSiblings === null || nodeIndex === null)
            return
        nodeSiblings.splice(nodeIndex + 1, 0, dragNode)
    }

    markDropModified(dataRef, fromCate, node)
    Object.assign(dataRef, Array.from(dataRef))
}

/** Excel 树节点额外属性 */
export type ExtraExcelTreeOption = {
    uuid?: string
    fullPath?: string
    levelType?: 'Excel' | 'Sheet'
    modified?: boolean
    // case信息
    sheetType?: SheetType,
    sheetHeader?: SheetHeader
    sheetError?: string,
}

/** 处理节点加载 */
export function handleLoad(node: TreeOption & ExtraExcelTreeOption) {
    return new Promise<void>((resolve, reject) => {
        if (!node.levelType) {
            reject()
        }
    })
}

/**
 * 获取节点属性
 * @param option 节点选项
 * @param dialog 对话框 API
 * @param message 消息 API
 */
export function nodeProps({option}: {
    option: TreeOption & ExtraExcelTreeOption
}, dialog: DialogApiInjection, message: MessageApiInjection) {
    return {
        // 左键点击
        onClick() {
            // 仅处理Sheet
            if (option.levelType != 'Sheet') return
            // 仅处理没有错误的Sheet
            if (option.sheetError != '') return

            if (!option.label) return
            // 如果规则数组里有
            let sheetRule = excelCheckMap.value.get(option.label)
            if (!sheetRule) {
                sheetRule = {}
                excelCheckMap.value.set(option.label, sheetRule)
            }

            console.log('[Click Option]', option)
            console.log('[Click Check]', excelCheckMap.value)
            nowSheetData.value = option
            // 填充不存在的列规则
            option.sheetHeader?.Col.forEach(c => {
                if (!c || c.AttrName == '') return
                if (!sheetRule[c.AttrName]) {
                    sheetRule[c.AttrName] = {
                        PropName: c.AttrName,
                        PropType: c.AttrType,
                        PropRules: []
                    }
                }
            })
            // 填充不存在的负责人
            if (!excelCheckManagerListMap.value.has(option.label)) {
                excelCheckManagerListMap.value.set(option.label, {
                    QA: [],
                    Designer: [],
                    Programmer: []
                })
            }
            // 加载表级规则
            nowTableRules.value = excelTableRulesMap.value.get(option.label) || []
            console.log('[Click SheetRule]', sheetRule)
            console.log('[Click SheetManager]', excelCheckManagerListMap.value)
            console.log('[Click TableRules]', nowTableRules.value)
            // 当前checkData赋予
            nowCheckData.value = sheetRule
        },
        // 右键菜单
        onContextmenu(e: MouseEvent): void {
            e.preventDefault()
            if (option.levelType === 'Excel') {
                // 收集该 Excel 下所有 Sheet 的负责人数据
                const sheetChildren = option.children || []

                // 构建用于负责人归一的子节点数组
                // 需要从 excelCheckManagerListMap 中获取每个 Sheet 的负责人
                const managerChildren = sheetChildren.map(child => {
                    const sheetName = child.label || ''
                    const managerList = excelCheckManagerListMap.value.get(sheetName)
                    return {
                        label: sheetName,
                        modified: false,
                        QA: managerList?.QA || [],
                        Designer: managerList?.Designer || [],
                        Programmer: managerList?.Programmer || [],
                    }
                })

                const menuOptions: DropdownOption[] = [
                    createManagerToOneOption(dialog, message, managerChildren as any, 'QA', () => {
                        // 同步回 excelCheckManagerListMap
                        managerChildren.forEach(child => {
                            const sheetName = child.label
                            const existing = excelCheckManagerListMap.value.get(sheetName)
                            if (existing) {
                                existing.QA = child.QA
                            }
                        })
                        message.success('QA负责人归一完成')
                    }),
                    createManagerToOneOption(dialog, message, managerChildren as any, 'Designer', () => {
                        managerChildren.forEach(child => {
                            const sheetName = child.label
                            const existing = excelCheckManagerListMap.value.get(sheetName)
                            if (existing) {
                                existing.Designer = child.Designer
                            }
                        })
                        message.success('策划负责人归一完成')
                    }),
                    createManagerToOneOption(dialog, message, managerChildren as any, 'Programmer', () => {
                        managerChildren.forEach(child => {
                            const sheetName = child.label
                            const existing = excelCheckManagerListMap.value.get(sheetName)
                            if (existing) {
                                existing.Programmer = child.Programmer
                            }
                        })
                        message.success('程序负责人归一完成')
                    }),
                    createDividerOption(),
                    createExecCateOption(message, option.label, () => {
                        // 执行该 Excel 下所有 Sheet 的检查
                        const sheetNames = sheetChildren
                            .filter(child => child.sheetError === '')
                            .map(child => child.label)
                            .filter((name): name is string => !!name)
                        if (sheetNames.length > 0) {
                            startExcelRulesForSheets(sheetNames, message)
                        } else {
                            message.error('该 Excel 下没有可检查的 Sheet')
                        }
                    }),
                    createDividerOption(),
                    {
                        title: '打开Excel文件',
                        key: 'open-excel-file',
                        props: {
                            onClick() {
                                const filePath = option.fullPath
                                if (filePath) {
                                    // 通过后端打开 Excel 文件（传空sheetName，直接打开文件路径）
                                    openExcelFile(message, filePath)
                                } else {
                                    message.error('未找到Excel文件路径')
                                }
                            }
                        }
                    },
                ]
                optionsRef.value = menuOptions as typeof optionsRef.value
                showDropdownRef.value = true
                xRef.value = e.clientX
                yRef.value = e.clientY
            }
            console.log(e.clientX, e.clientY)
        }
    }
}

/**
 * 渲染节点标签
 * @param option 节点选项
 */
export function renderLabel({option}: { option: TreeOption & ExtraExcelTreeOption }) {
    const sheetNum = option.children?.length || 0
    const level = option.levelType
    if (level == 'Excel') {
        const tags: VNode[] = []
        const wrongSheetNum = option.children?.reduce((acc: number, cur: (TreeOption & ExtraExcelTreeOption)) => {
            if (cur.sheetError != '') {
                acc++
            }
            return acc
        }, 0)
        tags.push(h(NTag, {
            type: "info",
            style: "marginLeft: 10px",
            size: "small"
        }, () => `${sheetNum - (wrongSheetNum || 0)}`))
        if (wrongSheetNum != 0) {
            tags.push(h(NTag, {type: "error", style: "marginLeft: 2px", size: "small"}, () => `${wrongSheetNum}`))
        }
        const children = [
            h("div", {}, [`${option.label}`, ...tags]),
        ]
        return h("div", {}, children)
    } else if (level == 'Sheet') {
        let ruleNum = 0
        if (option.label) {
            const rules = excelCheckMap.value.get(option.label)
            if (rules) {
                ruleNum = Object.values(rules).reduce((acc, cur) => {
                    acc += cur?.PropRules.length || 0
                    return acc
                }, 0)
            }
        }

        // 检查表级规则是否有失败的
        let tableRuleFailedNum = 0
        if (option.label) {
            const results = tableCheckResults.value.filter(r => r.sheetName === option.label)
            tableRuleFailedNum = results.filter(r => !r.ok).length
        }

        const tags: VNode[] = []
        tags.push(h(NTag, {type: "success", style: "marginLeft: 10px", size: "small"}, () => `${ruleNum}`))
        if (tableRuleFailedNum > 0) {
            tags.push(h(NTag, {type: "error", style: "marginLeft: 2px", size: "small"}, () => `${tableRuleFailedNum}表级错误`))
        }

        const hasErr = option.sheetError != ''
        const children = [
            h("div", {style: `color: ${hasErr ? 'red' : ''}`}, [`${option.label}${hasErr ? '(' + option.sheetError + ')' : ''}`, ...tags]),
        ]
        return h("div", {}, children)
    }
}
