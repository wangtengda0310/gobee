/**
 * 通用树形状态管理
 *
 * 提供树组件的展开、选中、级联等基础状态的工厂函数。
 * 每个页面调用 createTreeState() 获取独立的状态实例，避免跨页面状态污染。
 */
import {ref} from "vue";

/** 树形状态集合 */
export interface TreeState {
    /** 展开的节点 key 列表 */
    expandedKeysRef: ReturnType<typeof ref<string[]>>
    /** 选中的节点 key 列表 */
    checkedKeysRef: ReturnType<typeof ref<string[]>>
    /** 选中策略 */
    checkStrategy: ReturnType<typeof ref<'all' | 'parent' | 'child'>>
    /** 级联选中 */
    cascade: ReturnType<typeof ref<boolean>>
    /** 处理展开节点变化 */
    handleExpandedKeysChange: (keys: string[]) => void
    /** 处理选中节点变化 */
    handleCheckedKeysChange: (keys: string[]) => void
}

/**
 * 创建独立的树形状态实例
 * 每个页面应调用此函数获取各自的状态，避免状态共享冲突
 */
export function createTreeState(): TreeState {
    const expandedKeysRef = ref<string[]>([])
    const checkedKeysRef = ref<string[]>([])
    const checkStrategy = ref<'all' | 'parent' | 'child'>('all')
    const cascade = ref(true)

    function handleExpandedKeysChange(keys: string[]) {
        expandedKeysRef.value = keys
    }

    function handleCheckedKeysChange(keys: string[]) {
        checkedKeysRef.value = keys
    }

    return {
        expandedKeysRef,
        checkedKeysRef,
        checkStrategy,
        cascade,
        handleExpandedKeysChange,
        handleCheckedKeysChange,
    }
}
