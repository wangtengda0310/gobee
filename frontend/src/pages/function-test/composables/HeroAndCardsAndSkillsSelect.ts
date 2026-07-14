import {excelSkillList, excelSkillMap} from "../config/Skill";
import {cardTypeMap, excelCardsList, excelCardsMap} from "../config/Card";
import {excelHeroList, excelHeroMap} from "@shared/config/hero";
import {nowCaseData} from "./use-case-data";
import {computed} from "vue";
import {Step} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/function-test";
import {CardsTemplate} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-robot/project/xcard/xcard_excel/excel";

export const cardSelectFilter = (pattern: string, option: any) => {
    // 匹配包含 "·数字" 的模式（如 "abc·123"）
    const pointPatternMatch = pattern.match(/·(\d+)/);

    if (pointPatternMatch && option.point) {
        // 如果包含 "·数字"，需要同时匹配 point 和 chs/value
        const pointNumber = pointPatternMatch[1];
        const remainingPattern = pattern.replace(/·\d+/, '').trim();

        // 检查 point 匹配
        const pointMatch = option.point.toString().startsWith(pointNumber);

        // 检查剩余部分的匹配
        let remainingMatch = true;
        if (remainingPattern) {
            if (/^\d+.*/.test(remainingPattern)) {
                remainingMatch = option.value.toString().startsWith(remainingPattern);
            } else {
                remainingMatch = (option.chs as string).includes(remainingPattern);
            }
        }

        return pointMatch && remainingMatch;
    } else if (/^\d+.*/.test(pattern)) {
        // 如果是数字开头，匹配 option.value
        return option.value.toString().startsWith(pattern);
    } else {
        // 其他情况匹配 option.chs
        return (option.chs as string).includes(pattern);
    }
}

export const chsAndNumSelectFilter = (pattern: string, option: any) => {
    if (/^\d+.*/.test(pattern)) {
        return (option.value as number).toString().startsWith(pattern)
    } else {
        return (option.chs as string).includes(pattern)
    }
}

export const excelSkillSelectOption = computed(() => {
    const result = excelSkillList.value.map(([k, v]) => {
        return {
            label: v + '(' + k + ')',
            value: k,
            chs: v
        }
    })
    console.log('[HeroAndCardsAndSkillsSelect.ts] excelSkillSelectOption 重新计算, 数量=', result.length, '第一条=', result[0])
    return result
})

export const excelSkillSelectOptionFromInit = (step: Step) => {
    if (nowCaseData.value && step.robotIdx && nowCaseData.value.initYanWu) {
        return excelHeroMap.value[nowCaseData.value.initYanWu.customHeroes[(step.robotIdx - 1) % nowCaseData.value.initYanWu.customHeroes.length]?.heroId]?.Skills.filter(s => {
            return !nowCaseData.value!.initYanWu!.customHeroes[(step.robotIdx - 1) % nowCaseData.value!.initYanWu!.customHeroes.length]?.delSkills.includes(s)
        }).concat(...nowCaseData.value.initYanWu?.customHeroes[(step.robotIdx - 1) % nowCaseData.value.initYanWu?.customHeroes.length]?.addSkills).map((skill: number) => {
            // 显示英雄id和(座位号)
            return {
                label: excelSkillMap.value[skill]?.SkillName + `(${skill})`,
                value: skill,
                chs: excelSkillMap.value[skill]?.SkillName
            }
        })
    } else {
        return []
    }
}

export const excelCardsSelectDynUniqueOption = computed(() => {
    const result = excelCardsList.value.map(([k, v, p, t]) => {
        const hasCard = nowCaseData.value && nowCaseData.value.initYanWu && (
            nowCaseData.value.initYanWu.cards.includes(k)
            || nowCaseData.value.initYanWu.customHeroes.find(h => h.initCards.includes(k))
            || nowCaseData.value.initYanWu.customHeroes.find(h => h.initEquips.includes(k))
            || nowCaseData.value.initYanWu.customHeroes.find(h => h.exEquips.includes(k))
            || nowCaseData.value.initYanWu.customHeroes.find(h => h.augurCards.includes(k))
            || nowCaseData.value.initYanWu.customHeroes.find(h => Object.values(h.skillCardsMap).flatMap(cards=>cards).includes(k))
        )
        return {
            label: v + `|·${p}` + `|${cardTypeMap[t]}` + '|(' + k + ')',
            value: k,
            chs: v,
            point: p,
            type: t,
            style: {
                color: hasCard ? 'gray' : '',
                pointerEvents: hasCard ? 'none' : ''
            }
        }
    })
    console.log('[HeroAndCardsAndSkillsSelect.ts] excelCardsSelectDynUniqueOption 重新计算, 数量=', result.length, '第一条=', result[0])
    return result
})

export const excelCardsSelectOption = computed(() => excelCardsList.value.map(([k, v, p, t]) => {
    return {
        label: v + `|·${p}` + `|${cardTypeMap[t]}` + '|(' + k + ')',
        value: k,
        chs: v,
        point: p,
        type: t,
    }
}))

export const excelCardsSelectFallbackOption = (label: string) => {
    // 如果没匹配到的cards value回到这里
    if (Number.isFinite(Number(label))) {
        const id = Number(label)
        const cfg = excelCardsMap.value[id]
        if (cfg)
            return {
                label: cfg.Name + `|·${cfg.Point}` + `|${cardTypeMap[cfg.AttrType]}` + '|(' + cfg.Id + ')[自定义]',
                value: cfg.Id,
                chs: cfg.Name,
                point: cfg.Point,
                type: cfg.AttrType,
            }
        else
            return {
                label,
                value: id
            }
    } else {
        return {
            label: '',
            value: ''
        }
    }
}

export const excelCardsSelectOptionFromInit = computed(() => {
    if (!nowCaseData.value?.initYanWu) return []

    const { cards, customHeroes } = nowCaseData.value.initYanWu

    const allCardIds = [
        ...cards,
        ...customHeroes.flatMap(hero => {
            if (!hero) return []
            return [
                ...hero.initCards,
                ...hero.augurCards,
                ...hero.exEquips,
                ...Object.values(hero.skillCardsMap).flat()
            ]
        })
    ]

    return allCardIds.filter((cardId): cardId is number => cardId !== undefined).map(cardId => {
        const card = excelCardsMap.value[cardId]
        if (!card) {
            console.warn(`Card with ID ${cardId} not found in excelCardsMap`)
            return null
        }

        return {
            label: `${card.Name}(${cardId})`,
            value: cardId,
            chs: card.Name,
            point: card.Point,
            type: card.AttrType
        }
    }).filter(Boolean) // 过滤掉 null
})

export const excelHeroesSelectOption = computed(() => {
    const result = excelHeroList.value.map(([k, v]) => {
        return {
            label: v + '(' + k + ')',
            value: k,
            chs: v
        }
    })
    console.log('[HeroAndCardsAndSkillsSelect.ts] excelHeroesSelectOption 重新计算, 数量=', result.length, '第一条=', result[0])
    return result
})

export const excelHeroInitSkillSelectOption = (heroId: number) => excelHeroMap.value[heroId]?.Skills.map(s => {
    return {
        label: excelSkillMap.value[s]?.SkillName + `(${excelSkillMap.value[s].Id})`,
        value: s,
        chs: excelSkillMap.value[s]?.SkillName
    }
})

export const createTagOnlyNumber = (label: string, useMap?: {
    [p: `${number}`]: CardsTemplate
}, mapType?: 'cards' | 'normal' | 'skills') => {
    if (!isNaN(Number(label))) {
        const numericValue = Number(label)
        if (useMap && useMap[numericValue] && mapType) {

            let chs = ''
            const cardExtra = {}
            switch (mapType) {
                case "cards":
                    // 创建卡牌时额外带上point和type
                    chs = useMap[numericValue]?.Name
                    cardExtra['point'] = useMap[numericValue]?.Point
                    cardExtra['type'] = useMap[numericValue]?.AttrType
                    label = `|·${cardExtra['point']}` + `|${cardTypeMap[cardExtra['type']]}` + '|(' + label + ')'
                    break
                case "normal":
                    chs = useMap[numericValue]
                    break
                case "skills":
                    chs = useMap[numericValue]?.SkillName
                    break
            }

            const option = {
                label: chs ? chs + `${label}[自定义]` : label,
                value: numericValue,
                chs: chs ? chs : '',
                ...cardExtra
            }
            // console.log('useMap option:', option, 'value type:', typeof option.value)
            return option
        } else {
            const option = {
                label: label + '[自定义]',
                value: numericValue,
                chs: label
            }
            // console.log('normal option:', option, 'value type:', typeof option.value)
            return option
        }
    } else return {label: '', value: '', chs: ''}
}
