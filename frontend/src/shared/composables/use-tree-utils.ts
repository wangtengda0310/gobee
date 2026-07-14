/**
 * 通用树形操作工具函数
 *
 * 提供树节点的查找和拖拽基础逻辑，
 * excel-test 和 function-test 页面共用
 */
import type {TreeOption, TreeDropInfo} from "naive-ui";

/**
 * 查找节点的兄弟节点和索引
 * @param node 目标节点
 * @param nodes 搜索的节点列表
 * @returns [兄弟节点数组, 索引] 或 [null, null]
 */
export function findSiblingsAndIndex(
    node: TreeOption,
    nodes?: TreeOption[]
): [TreeOption[], number] | [null, null] {
    if (!nodes)
        return [null, null]
    for (let i = 0; i < nodes.length; ++i) {
        const siblingNode = nodes[i]
        if (siblingNode.key === node.key)
            return [nodes, i]
        const [siblings, index] = findSiblingsAndIndex(node, siblingNode.children)
        if (siblings && index !== null)
            return [siblings, index]
    }
    return [null, null]
}

/**
 * 标记拖拽涉及的分类节点为已修改
 * @param dataRef 树根数据
 * @param fromCateLabel 拖拽源分类 label
 * @param targetNode 拖拽目标节点
 */
export function markDropModified<T extends { label?: string; modified?: boolean; children?: T[] }>(
    dataRef: T[],
    fromCateLabel: string,
    targetNode: TreeOption,
) {
    dataRef.forEach((data) => {
        if (data.label == fromCateLabel) data.modified = true
        if (data.label == targetNode.label || data.children?.find(c => c.label == targetNode.label)) {
            data.modified = true
        }
    })
}
