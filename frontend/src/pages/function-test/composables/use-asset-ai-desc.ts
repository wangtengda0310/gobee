import {computed, ref, watch} from "vue"
import {nowCaseData, updateNtf} from "./use-case-data"
import {Asset, InitYanWu, JsonCaseService, Step} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/function-test"

const cachedDescByUuid = ref<Record<string, string>>({})

export function useAssetAiDesc(stepIndex: number, assetIndex: number) {
    const step = computed(() => nowCaseData.value?.caseSteps?.[stepIndex])
    const asset = computed(() => step.value?.assets?.[assetIndex])

    const assetUuid = computed(() => (asset.value as { uuid?: string } | undefined)?.uuid)

    const fetchAiDesc = (): Promise<string> => {
        const a = asset.value
        const s = step.value
        if (!s || !a || !nowCaseData.value?.initYanWu) {
            return Promise.resolve("请输入描述")
        }
        const assetCopy = new Asset(a)
        const stepCopy = new Step(s)
        const initYanWuCopy = new InitYanWu(nowCaseData.value.initYanWu)
        return JsonCaseService.GenerateAssetDesc(assetCopy, stepCopy, initYanWuCopy)
    }

    const refreshCachedDesc = () => {
        const uuid = assetUuid.value
        if (!uuid) return
        fetchAiDesc().then(desc => {
            cachedDescByUuid.value[uuid] = desc
        })
    }

    // 仅在 asset 的 uuid 或断言类型(msgName) 变化时重新生成建议描述。
    // 刻意不 deep 监听 asset/step：否则 watch 回调链路会同步回写 asset.attr，
    // 而 step 是各断言组件共享的深层对象，任一 asset.attr 变化都会让所有断言组件的 watch 重入 → 死循环 + 内存泄漏。
    // 刻意不 immediate：切换用例/加载用例时 N 个断言组件会批量挂载，immediate 会让每个组件各发一次
    // GenerateAssetDesc RPC → 后端 N 次全量配表查询，造成明显卡顿。首次挂载显示缓存或"请输入描述"，需新生成点「应用智能描述」。
    watch(
        () => [assetUuid.value, asset.value?.msgName],
        () => refreshCachedDesc()
    )

    const cachedDesc = computed(() => {
        const uuid = assetUuid.value
        return (uuid && cachedDescByUuid.value[uuid]) || "请输入描述"
    })

    const applyAiDesc = () => {
        if (!asset.value) return
        fetchAiDesc().then(desc => {
            if (asset.value) {
                asset.value.desc = desc
                const uuid = assetUuid.value
                if (uuid) {
                    cachedDescByUuid.value[uuid] = desc
                }
                updateNtf()
            }
        })
    }

    const updateDesc = (value: string) => {
        if (asset.value) {
            asset.value.desc = value
            updateNtf()
        }
    }

    return {cachedDesc, applyAiDesc, updateDesc}
}
