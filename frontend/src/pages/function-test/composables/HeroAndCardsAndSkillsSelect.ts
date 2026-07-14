import {excelSkillList, excelSkillMap} from "../config/Skill";
import {cardTypeMap, excelCardsList, excelCardsMap} from "../config/Card";
import {excelHeroList, excelHeroMap} from "@shared/config/hero";
import {nowCaseData} from "./use-case-data";
import {computed} from "vue";
import {Step} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/function-test";
import {CardsTemplate} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-robot/project/xcard/xcard_excel/excel";
import type {SelectGroupOption, SelectOption} from "naive-ui";
import {getIdentityColorHex} from "../config/Identity";

/** 牌堆组色条：公共抽牌堆，无所属角色 */
const DECK_BAR_COLOR = '#888888';
/** 未映射身份的回退色条（与牌堆灰区分） */
const UNKNOWN_IDENTITY_BAR_COLOR = '#BBBBBB';

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

/**
 * 按"演武初始化"已分配的卡牌来源生成分组下拉选项（供步骤动作/资产断言选牌使用）。
 *
 * 流程：取 initYanWu 的 cards 与各 customHero 的私有牌 → 按来源分组 → 跳过空组。
 * 分组顺序与含义：
 *   1. 「🎴 摸牌堆」：initYanWu.cards，即公共抽牌堆里的牌
 *   2. 「🧑 座位N·角色名」：每个 customHero 的 initCards(初始手牌) +
 *      augurCards(卜卦牌) + exEquips(ex装备) + skillCardsMap(技能关联牌)，
 *      座位号 = 数组下标 + 1，与步骤里"座位(N)"标签一致
 *
 * 叶子 option 结构与原扁平实现保持一致（label/value/chs/point/type），
 * 以保证 cardSelectFilter / fallback-option / tag@create / multiple 全部兼容；
 * step.cards 等存储字段仍为扁平卡牌 id 数组，旧用例数据不受影响。
 */
export const excelCardsSelectOptionFromInit = computed((): SelectGroupOption[] => {
    if (!nowCaseData.value?.initYanWu) return []

    const {cards, customHeroes} = nowCaseData.value.initYanWu

    // 把卡牌 id 列表转成叶子选项：过滤 undefined → 查表 → 跳过未配置的 id → 映射为 option
    // barColor 给选项左侧竖条着色（牌堆灰、角色手牌按身份阵营色），与分组标题共同区分来源
    const toCardOptions = (ids: (number | undefined)[], barColor: string): SelectOption[] =>
        ids.filter((id): id is number => id !== undefined)
            .map(id => {
                const card = excelCardsMap.value[id]
                if (!card) {
                    console.warn(`Card with ID ${id} not found in excelCardsMap`)
                    return null
                }
                return {
                    label: `${card.Name}(${id})`,
                    value: id,
                    chs: card.Name,
                    point: card.Point,
                    type: card.AttrType,
                    style: {borderLeft: `4px solid ${barColor}`}
                }
            })
            .filter(Boolean) as SelectOption[]

    const groups: SelectGroupOption[] = []

    // 组1：摸牌堆（左侧灰竖条 —— 公共牌堆，无所属角色）
    const deckOptions = toCardOptions(cards, DECK_BAR_COLOR)
    if (deckOptions.length) {
        groups.push({type: 'group', label: '🎴 摸牌堆', key: 'deck', children: deckOptions})
    }

    // 组2..N：每个座位角色（查不到武将名时回退显示"座位N"）
    customHeroes.forEach((hero, idx) => {
        if (!hero) return
        const heroName = excelHeroMap.value[hero.heroId]?.Name ?? `座位${idx + 1}`
        // 左侧竖条 = 该角色身份阵营色（主红/忠金/反绿/内蓝），未映射身份回退浅灰（与牌堆灰区分）
        const barColor = getIdentityColorHex(hero.identity) ?? UNKNOWN_IDENTITY_BAR_COLOR
        const heroCardIds = [
            ...hero.initCards,
            ...hero.augurCards,
            ...hero.exEquips,
            ...Object.values(hero.skillCardsMap).flat()
        ]
        const heroOptions = toCardOptions(heroCardIds, barColor)
        if (heroOptions.length) {
            groups.push({
                type: 'group',
                label: `🧑 座位${idx + 1}·${heroName}`,
                key: `hero-${idx}`,
                children: heroOptions
            })
        }
    })

    return groups
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
