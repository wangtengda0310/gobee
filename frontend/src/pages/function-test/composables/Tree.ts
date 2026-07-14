/**
 * 用例树形结构管理
 *
 * 管理用例树的展开、选中、拖拽和渲染逻辑
 * 基础状态从 shared 导入，保留用例业务特有的拖拽规则和右键菜单
 */
import {h} from "vue";
import {NTag, TreeDropInfo, TreeOption} from 'naive-ui'
import {JsonCaseService, Step} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/function-test";
import {optionsRef, showDropdownRef, xRef, yRef} from "../../../shared/composables/use-tree-dropdown";
import {
    addCaseModalData,
    addCateModalData,
    modifyCaseNameModalData,
    modifyCateNameModalData,
    showAddCaseModal,
    showAddCateModal,
    showModifyCaseNameModal,
    showModifyCateNameModal
} from "./Modals";
import {nowCaseData, ExtraCaseTreeOption, DragAttr, dataRef, showCasesDesc} from "./use-case-data";
import {ActionEnum} from "./StepActionsAndAssetsSelect";
import {JsonsDir} from "./Option";
import {DialogApiInjection} from "naive-ui/es/dialog/src/DialogProvider";
import {MessageApiInjection} from "naive-ui/es/message/src/MessageProvider";
import {copyCase, execCateCases} from "./Func";
import {
    createManagerToOneOption,
    createExecCateOption,
    createDividerOption
} from "../../../shared/composables/use-tree-context-menu";
import {createTreeState} from "../../../shared/composables/use-tree-state";
import {findSiblingsAndIndex, markDropModified} from "../../../shared/composables/use-tree-utils";

// 创建独立的树形状态实例，避免与 excel-test 页面共享状态
const treeState = createTreeState()
export const {expandedKeysRef, checkedKeysRef, checkStrategy, cascade, handleExpandedKeysChange, handleCheckedKeysChange} = treeState

/** 处理节点拖拽（用例特有规则：允许 Case 拖入 Categories 内） */
export function handleDrop({node, dragNode, dropPosition}: TreeDropInfo) {
    if (dragNode.levelType == 'Categories') return;

    if (
        !(dropPosition === 'inside' && (node.levelType == 'Categories' && dragNode.levelType == 'Cases'))
        && !(dropPosition === 'before' && (node.levelType == dragNode.levelType))
        && !(dropPosition === 'after' && (node.levelType == dragNode.levelType))
    ) return

    const [dragNodeSiblings, dragNodeIndex] = findSiblingsAndIndex(
        dragNode,
        dataRef
    )
    if (dragNodeSiblings === null || dragNodeIndex === null)
        return

    // 记录拖拽源分类
    let fromCate = ""
    dataRef.forEach((data: TreeOption & ExtraCaseTreeOption) => {
        if (data.children?.find(c => c.label == dragNode.label) && data.label) {
            fromCate = data.label
        }
    })

    dragNodeSiblings.splice(dragNodeIndex, 1)
    if (dropPosition === 'inside' && (node.levelType == 'Categories' && dragNode.levelType == 'Cases')) {
        if (node.children) {
            node.children.unshift(dragNode)
        } else {
            node.children = [dragNode]
        }
    } else if (dropPosition === 'before' && (node.levelType == dragNode.levelType)) {
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

export function handleLoad(node: TreeOption & ExtraCaseTreeOption) {
    return new Promise<void>((resolve, reject) => {
        if (!node.levelType) {
            reject()
        }

        if (node.levelType == 'Categories' && node.fullPath) {
            resolve()
        } else if (node.levelType == 'Cases') {
            reject()
        }
    })
}

export function nodeProps({option}: {
    option: TreeOption & ExtraCaseTreeOption
}, dialog: DialogApiInjection, message: MessageApiInjection) {
    return {
        // 左键
        onClick() {
            console.log('[Click]', option)
            // 如果层级属性存在且层级是Cases且steps未加载到内存且有文件路径的时候会从文件加载到内存
            if (option.levelType && option.levelType == "Cases") {
                nowCaseData.value = option
            } else if (option.levelType && option.levelType == "Categories") {
                //
            }
        },
        // 右键 https://www.naiveui.com/zh-CN/dark/components/tree#node-props.vue
        onContextmenu(e: MouseEvent): void {
            if (option.levelType && option.levelType == "Categories") {
                optionsRef.value = [
                    {
                        title: '添加分类',
                        key: 'add-cate',
                        props: {
                            onClick(payload) {
                                console.log(payload)
                                addCateModalData.value.text = "是否添加一个分类?"
                                addCateModalData.value.index = dataRef.findIndex(d => d.label == option.label)
                                showAddCateModal.value = true
                            },
                        },
                    },
                    {
                        title: '修改分类名',
                        key: 'rename-cate',
                        props: {
                            onClick(payload) {
                                console.log(dataRef, option)
                                modifyCateNameModalData.value = {
                                    cateRef: option,
                                    newName: option.label + '(新)'
                                }
                                showModifyCateNameModal.value = true
                            },
                        }
                    },
                    {
                        title: '添加用例',
                        key: 'add-inside-case',
                        props: {
                            onClick(payload) {
                                console.log(dataRef, option)
                                addCaseModalData.value = {
                                    cate: option.label ?? '位置分类',
                                    index: dataRef.find(d => {
                                        return d.label == option.label
                                    })?.children?.findIndex(d => {
                                        return d.key == option.key
                                    }) ?? -1
                                }
                                showAddCaseModal.value = true
                            },
                        }
                    },
                    createManagerToOneOption(
                        dialog,
                        message,
                        option.children as any,
                        'caseManager',
                        () => {
                            // 同步完成后设置 modified 标志
                            option.children?.forEach(c => c.modified = true)
                        }
                    ),
                    createDividerOption(),
                    createExecCateOption(message, option.label, () => execCateCases(message, option, true), "exec-cate-mem", "执行分类"),
                    createExecCateOption(message, option.label, () => execCateCases(message, option, false), "exec-cate-file", "执行分类(从本地文件)"),
                    createDividerOption(),
                    {
                        title: '删除分类',
                        key: 'del-cate',
                        props: {
                            onClick(payload) {
                                dialog.warning({
                                    title: '删除分类',
                                    content: '你要删除这个分类吗?',
                                    positiveText: '确定',
                                    negativeText: '取消',
                                    draggable: true,
                                    onPositiveClick: () => {
                                        console.log(payload)
                                        if (JsonsDir.value && option.label)
                                            JsonCaseService.DeleteJSONFile(JsonsDir.value + '/' + option.label + '.json').then(res => {
                                                const index = dataRef.findIndex(d => d.label == option.label)
                                                if (-1 != index) {
                                                    dataRef.splice(index, 1)
                                                }
                                            })
                                    },
                                    onNegativeClick: () => {
                                    }
                                })
                            }
                        },
                    },
                ]
                showDropdownRef.value = true
                xRef.value = e.clientX
                yRef.value = e.clientY
            } else if (option.levelType && option.levelType == "Cases") {
                optionsRef.value = [
                    {
                        title: '添加用例',
                        key: 'add-case',
                        props: {
                            onClick(payload) {
                                console.log(dataRef, option)
                                addCaseModalData.value = {
                                    cate: option.category ?? '位置分类',
                                    index: dataRef.find(d => {
                                        return d.label == option.category
                                    })?.children?.findIndex(d => {
                                        return d.key == option.key
                                    }) ?? -1
                                }
                                showAddCaseModal.value = true
                            },
                        }
                    },
                    {
                        title: '复制用例',
                        key: 'copy-case',
                        props: {
                            onClick(payload) {
                                console.log(dataRef, option)
                                copyCase(option)
                            },
                        }
                    },
                    {
                        title: '重命名用例',
                        key: 'rename-case',
                        props: {
                            onClick(payload) {
                                console.log(dataRef, option)
                                modifyCaseNameModalData.value.caseRef = option
                                modifyCaseNameModalData.value.newName = option.label + "(新)"
                                showModifyCaseNameModal.value = true
                            },
                        }
                    },
                    {
                        title: '添加行动',
                        key: 'add-step',
                        props: {
                            onClick(payload) {
                                console.log(dataRef, option)
                                if (!option.caseSteps) {
                                    option.caseSteps = []
                                }
                                option.caseSteps.push(Object.assign(new Step({
                                    desc: "",
                                    action: ActionEnum.Sleep,
                                    sleepTime: 0,
                                }), {uuid: crypto.randomUUID()}))
                                option.modified = true
                            },
                        }
                    },
                    {
                        title: '删除用例',
                        key: 'del-case',
                        props: {
                            onClick(payload) {
                                dialog.warning({
                                    title: '删除用例',
                                    content: '你要删除这个用例吗?',
                                    positiveText: '确定',
                                    negativeText: '取消',
                                    draggable: true,
                                    onPositiveClick: () => {
                                        console.log(dataRef, option)
                                        const cate = dataRef.find(d => {
                                            return d.label == option.category
                                        })
                                        const index = cate?.children?.findIndex(d => {
                                            return d.key == option.key
                                        }) ?? -1
                                        if (cate && index != -1) {
                                            cate.modified = true
                                            cate.children?.splice(index, 1)
                                            cate.isLeaf = !cate.children || cate.children?.length == 0
                                        }
                                    },
                                    onNegativeClick: () => {
                                    }
                                })
                            },
                        }
                    },
                ]
                showDropdownRef.value = true
                xRef.value = e.clientX
                yRef.value = e.clientY
            }
            console.log(e.clientX, e.clientY)
            e.preventDefault()
        }
    }
}

export function renderLabel({option}: { option: TreeOption & ExtraCaseTreeOption }) {
    const caseNum = option.children?.length || 0
    const level = option.levelType
    const tag = h(NTag, {type: "info", style: "marginLeft: 10px", size: "small"}, () => `${caseNum}`)
    if (option.modified) {
        const children = [
            h("div", {}, level == 'Categories' ? [`*${option.label}`, tag] : [`*${option.label}`]),
        ]
        if (option.levelType == 'Cases' && showCasesDesc.value) children.push(h("div", {style: {fontSize: "10px"}}, `${option.caseDesc}`))
        return h("div", {style: {color: '#49ff3d'}}, children)
    } else {
        const children = [
            h("div", {}, level == 'Categories' ? [`${option.label}`, tag] : [`${option.label}`]),
        ]
        if (option.levelType == 'Cases' && showCasesDesc.value) children.push(h("div", {style: {fontSize: "10px"}}, `${option.caseDesc}`))
        return h("div", {}, children)
    }
}
