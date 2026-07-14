/**
 * 身份配置数据
 *
 * @module config/Identity
 * @description
 * 定义游戏中的身份映射关系和颜色配置
 *
 * @依赖
 * - CaseData: 当前用例数据（用于身份颜色选择）
 * - Hero: 英雄配置（用于显示英雄名称）
 */
import {nowCaseData} from "../composables/use-case-data";
import {excelHeroMap} from "@shared/config/hero";

export const excelIdentityMap = {
    1: "主公",
    2: "先主",
    3: "潜龙",
    4: "明主",
    5: "忠臣",
    6: "储君",
    7: "上将",
    8: "禁军",
    9: "反贼",
    10: "盟主",
    11: "军师",
    12: "先锋",
    13: "内奸",
    14: "黄巾",
    15: "刺客",
    16: "伪帝",
    24: "友方",
    25: "敌方",
}

export const excelIdentityColorMap = {
    1: [1, 2, 3, 4, 5, 6, 7, 8,],
    2: [9],
    3: [17],
    4: [25],
    5: [1, 2, 3, 4, 5, 6, 7, 9, 17, 25],
    6: [1, 2, 3, 4, 5, 6, 7, 17, 25],
    7: [1, 2, 3, 4, 5, 6, 7, 9, 17, 25],
    8: [1, 2, 3, 4, 5, 6, 7, 9, 17, 25],
    9: [65],
    10: [65],
    11: [65],
    12: [65],
    13: [97],
    14: [105],
    15: [113],
    16: [121],
    24: [73, 74, 75, 76, 77, 78, 79, 80],
    25: [81, 82, 83, 84, 85, 86, 87, 88],
}

export const excelIdentityList = Object.keys(excelIdentityMap).map((k) => {
    return [Number(k), excelIdentityMap[Number(k)]]
})

export const canUseIdentityOption = (identity: number, color: number) => {
    const option = excelIdentityColorMap[identity]?.map(cur => {
            // 遍历该身份的阵营
            let hasColor = false
            let heroWhoHasColor = null
            for (let customHero of nowCaseData.value?.initYanWu?.customHeroes || []) {
                hasColor = customHero?.color == cur && customHero?.identity == identity
                if (hasColor) {
                    // 同身份的人有了相同的颜色,
                    if (color != cur) heroWhoHasColor = (customHero?.heroId) ? excelHeroMap.value[customHero?.heroId]?.Name : customHero?.heroId
                    break
                }
            }
            return {
                label: '颜色' + '(' + cur + ')' + (heroWhoHasColor ? '->' + heroWhoHasColor : ''),
                value: cur,
                // disabled: hasColor // 放开颜色选择限制
            }
        }
    )
    return option
}
