/**
 * 国家/阵营配置数据
 *
 * @module config/ECountry
 * @description
 * 定义游戏中的国家/阵营映射关系
 */
export const countryMap = {
    0: "CoNone",
    1: "Qin",
    2: "XiChu",
    3: "XiHan",
    4: "DongHan",
    5: "Huang",
    6: "CaoWei",
    7: "Shu",
    8: "SunWu",
    9: "Fei",
    10: "ZhangChu",
    11: "Wei",
    12: "Chu",
    13: "Qi",
    14: "Zhao",
    15: "Han",
    16: "Yan",
}

export const countryList = Object.keys(countryMap).map(k => {
    return [1 << Number(k), countryMap[Number(k)]]
})
