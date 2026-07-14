import {ref} from "vue";
import {TreeOption} from "naive-ui";
import {dataRef, ExtraCaseTreeOption, normalizeCaseSteps, nowCaseData} from "./use-case-data";
import {nowRunningCase} from "./RobotTestLog";
import {resolveCaseInTreeAndExpand} from "./Tree";

/** 左侧用例树选中 key（与 nowCaseData 同步） */
export const selectedCaseKeysRef = ref<Array<string | number>>([])

/** 执行日志 Tab 当前索引（对应 nowRunningCase 下标） */
export const activeLogCaseIndexRef = ref<number>()

export function findCaseInDataRef(
    predicate: (c: TreeOption & ExtraCaseTreeOption) => boolean
): (TreeOption & ExtraCaseTreeOption) | undefined {
    for (const cate of dataRef) {
        const found = cate.children?.find(predicate)
        if (found) return found as TreeOption & ExtraCaseTreeOption
    }
    return undefined
}

/**
 * 按本次运行用例列表下标切换当前用例，同步树选中、用例配置/步骤与日志 Tab。
 */
export function selectCase(index: number | string) {
    const normalizedIndex = Number(index)
    if (!Number.isFinite(normalizedIndex)) return

    const batch = nowRunningCase.value
    if (normalizedIndex < 0 || normalizedIndex >= batch.length) return

    const batchOption = batch[normalizedIndex]
    if (!batchOption) return

    const option = resolveCaseInTreeAndExpand(batchOption)

    activeLogCaseIndexRef.value = normalizedIndex
    normalizeCaseSteps(option?.caseSteps)
    nowCaseData.value = option
    if (option.key != null) {
        selectedCaseKeysRef.value = [option.key]
    }
}

export function handleTreeSelectedKeysUpdate(keys: Array<string | number>) {
    selectedCaseKeysRef.value = keys
}

/**
 * 左侧树点击：优先映射到 nowRunningCase 索引，否则仅切换编辑用例（非本次批量）。
 */
export function selectCaseByOption(option: TreeOption & ExtraCaseTreeOption) {
    const resolved = resolveCaseInTreeAndExpand(option)
    const batchIndex = nowRunningCase.value.findIndex(
        c => c.key === resolved.key || c.label === resolved.label
    )
    if (batchIndex >= 0) {
        selectCase(batchIndex)
        return
    }

    nowCaseData.value = resolved
    normalizeCaseSteps(resolved?.caseSteps)
    if (resolved.key != null) {
        selectedCaseKeysRef.value = [resolved.key]
    }
    activeLogCaseIndexRef.value = undefined
}
