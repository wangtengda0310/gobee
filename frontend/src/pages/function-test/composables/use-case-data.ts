import {reactive, ref} from "vue";
import {TreeOption} from "naive-ui";
import {Asset, CustomHero, InitYanWu, Step} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/function-test";
import {ActionEnum, AssetEnum} from "./StepActionsAndAssetsSelect";

// 额外的用例属性
export type DragAttr = {
    uuid?: string
}

export type ExtraCaseTreeOption = {
    uuid?: string
    fullPath?: string
    levelType?: 'Categories' | 'Cases' | 'Step'
    category?: string
    caseDesc?: string
    caseManager?: string
    initYanWu?: InitYanWu
    modified?: boolean
    caseSteps?: (Step & DragAttr)[]
}

/** 生成随机UUID */
export const randomUUID = () => crypto.randomUUID()

// 正在查看的Case的详细
export const nowCaseData = ref<TreeOption & ExtraCaseTreeOption>()

/** 用例树数据源（从 TreeAndHistory.ts 合并） */
export const dataRef = reactive<(TreeOption & ExtraCaseTreeOption)[]>(createData() as (TreeOption & ExtraCaseTreeOption)[])

/** 是否显示用例描述（从 TreeSearch.ts 内联） */
export const showCasesDesc = ref(true)

// case内修改时的通知
export const updateNtf = () => {
    if (nowCaseData.value)
        nowCaseData.value.modified = true
}

export function newCaseData(cate: string, caseName: string, fullPath?: string,): (TreeOption & ExtraCaseTreeOption) {
    return {
        label: caseName,
        key: crypto.randomUUID(),
        isLeaf: true,
        levelType: 'Cases',
        fullPath,
        category: cate,
        initYanWu: new InitYanWu({
            presentId: 0,
            cardPile: 0,
            cards: [],
            operateTimeMs: 30000,
            customHeroes: [
                {
                    ...new CustomHero({
                        identity: 1,
                        color: 1,
                        heroId: 10001,
                        initCards: [],
                        initEquips: [],
                        augurCards: [],
                        addSkills: [],
                        delSkills: [],
                        exEquips: [],
                        skillCardsMap: {},
                    }),
                    uuid: crypto.randomUUID()
                } as CustomHero & { uuid: string },
                {
                    ...new CustomHero({
                        identity: 1,
                        color: 2,
                        heroId: 10002,
                        initCards: [],
                        initEquips: [],
                        augurCards: [],
                        addSkills: [],
                        delSkills: [],
                        exEquips: [],
                        skillCardsMap: {},
                    }),
                    uuid: crypto.randomUUID()
                } as CustomHero & { uuid: string },
            ]
        }),
        caseDesc: "用例描述",
        caseSteps: [
            {
                ...new Step({
                    id: 1,
                    desc: "",
                    robotIdx: 1,
                    action: ActionEnum.Sleep,
                    sleepTime: 0,
                    assets: [
                        {
                            ...new Asset({
                                msgName: AssetEnum.UpdateProperty,
                                desc: "",
                                attr: {
                                    "IdType": "User",
                                    "PropID": "CircleDoHurtValue",
                                    "PropValue": "2",
                                }
                            }),
                            uuid: crypto.randomUUID()
                        } as any,
                        {
                            ...new Asset({
                                msgName: AssetEnum.ReplaceCard,
                                desc: "",
                                attr: {}
                            }),
                            uuid: crypto.randomUUID()
                        } as any
                    ]
                }),
                uuid: crypto.randomUUID(),
            } as Step & { uuid: string }
        ]
    }
}

export function createData(): (TreeOption & ExtraCaseTreeOption)[] {
    const data: (TreeOption & ExtraCaseTreeOption)[] = [
        {
            label: "测试类别",
            key: crypto.randomUUID(),
            isLeaf: false,
            fullPath: "",
            levelType: "Categories",
            modified: false,
            children: [
                newCaseData("测试类别", "新用例", "",)
            ]
        },
    ]

    // 给每个step来一个uuid
    data.forEach(cate => {
        if (cate.children) {
            cate.children.forEach((c: TreeOption & ExtraCaseTreeOption) => {
                if (c.caseSteps) {
                    c.caseSteps.forEach(s => {
                        if (s.assets) {
                            s.assets.forEach(a => {
                                a['uuid'] = crypto.randomUUID()
                            })
                        }
                        s['uuid'] = crypto.randomUUID()
                    })
                }
            })
        }
    })

    return data
    // return []
}