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

/**
 * 身份 → 可选颜色值表，座位区域"颜色下拉框"的选项数据源（init-yanwu-panel.vue 绑定）。
 *
 * color 值语义 = 身份阵营颜色编号（同 color 的玩家属同一颜色阵营），权威定义来自游戏配表
 * IdentityEncodeRule_身份编码规则表（config/excel，字段 IdentityId→Color→Encode）。本表是该配表的复刻，常见取值：
 *   - 主公/忠臣/储君/上将/禁军（主忠系）：1~8、9、17、25
 *   - 反贼/盟主/军师/先锋：65
 *   - 内奸：97、黄巾：105、刺客：113、伪帝：121
 *   - 友方/敌方（千里单骑模式，不在编码规则表）：73~80 / 81~88
 *
 * ⚠️ color 不是 ECountry 国家编号（ECountry 是另一个独立维度，见 config/ECountry.ts）。
 *
 * 下拉框 label 由 canUseIdentityOption 生成为 "颜色(N)"，被同身份角色占用时追加 "->武将名"。
 *
 * @see 游戏配表 IdentityEncodeRule_身份编码规则表（config/excel）color 的权威来源
 * @see backend/pkg/function-test/reverse_translate.go 第716行 hero.Color 解析
 *      ⚠️ 后端误用 countryMap（国家表）翻译 color，语义不符（已知 bug，详见该处注释）。
 */
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

/**
 * 身份 → 身份大类（IdentityClass）映射，复刻游戏配表 IdentityRulesTemplate.json 的 Class 字段。
 *
 * IdentityClass（proto common/enum.proto:120）将所有身份归为 4 大类，是座位「显示底色」的依据：
 *   1=IC_Zhu(主)、2=IC_Zhong(忠)、3=IC_Fan(反)、4=IC_Nei(内)
 * 变体身份归入对应色系：
 *   主公系 1~4 → 主；忠臣系 5~8 → 忠；反贼系 9~12 → 反；内奸系 13~16 → 内。
 * 友方(24)/敌方(25) 等千里单骑身份不在本表，显示底色回退 undefined。
 */
export const identityClassMap: Record<number, number> = {
    1: 1, 2: 1, 3: 1, 4: 1,      // 主公/先主/潜龙/明主
    5: 2, 6: 2, 7: 2, 8: 2,      // 忠臣/储君/上将/禁军
    9: 3, 10: 3, 11: 3, 12: 3,   // 反贼/盟主/军师/先锋
    13: 4, 14: 4, 15: 4, 16: 4,  // 内奸/黄巾/刺客/伪帝
}

/**
 * 身份大类 → 显示色（hex），复刻客户端 ColorUtil.cs 的 GetColorByIdentity 4 色 RGB。
 *
 * 权威来源：D:/work/client/Master/Card/Assets/Scripts/HotUpdate/Logic/Util/ColorUtil.cs
 *   ColorZhuGong    = (0xff,0x48,0x48) = #FF4848（主-红）
 *   ColorZhongCheng = (0xff,0xca,0x59) = #FFCA59（忠-金）
 *   ColorFanZei     = (0x5d,0xd8,0x57) = #5DD857（反-绿）
 *   ColorNeiJian    = (0x6d,0xd1,0xff) = #6DD1FF（内-蓝）
 *
 * ⚠️ 这是按「身份大类」的显示底色，与座位 color 下拉框的值是两个维度：
 *   - color 下拉框值（excelIdentityColorMap，1~8/65/97…）= 同身份内序号，计算公式见客户端
 *     UIUtils.Room.cs:9 GetIdentityColor：主公系 8*identity+count-8、反贼系固定 8*FanZei-7、
 *     黄巾 8*identity-7、内奸/刺客/伪帝 8*identity+count-8。
 *   - 显示底色（本表）= 按 identity 大类着色；多座位时用 color 值做数字角标区分，不改底色。
 */
export const classColorHexMap: Record<number, string> = {
    1: '#FF4848', // 主
    2: '#FFCA59', // 忠
    3: '#5DD857', // 反
    4: '#6DD1FF', // 内
}

/**
 * 取身份对应的显示色（hex）；未映射的身份（友方/敌方/未知）返回 undefined。
 */
export const getIdentityColorHex = (identity: number): string | undefined => {
    const cls = identityClassMap[identity]
    return cls !== undefined ? classColorHexMap[cls] : undefined
}

/**
 * 取某个座位号对应的身份阵营色（hex）。
 * 映射链：座位号 n → customHeroes[n-1].identity → getIdentityColorHex → 阵营 4 色。
 *
 * 用于「按座位号」着色但语义上仍属身份阵营色的场景，与座位卡片（按 identity 直接取色）
 * 保持同一套 4 色体系：
 *   - 动作卡片：step.robotIdx（座位下拉）变色
 *   - 执行日志：log.msg.ID（座位号）的 ID[x] 变色
 *
 * @param customHeroes 座位列表（nowCaseData.initYanWu.customHeroes）
 * @param seatNo 座位号（1-based；超出长度时按取模回绕，与 aiDesc 的 robotIdx 取座位逻辑一致）
 */
export const getSeatColorHex = (
    customHeroes: { identity: number }[] | undefined,
    seatNo: number
): string | undefined => {
    if (!customHeroes?.length || seatNo < 1) return undefined
    const hero = customHeroes[(seatNo - 1) % customHeroes.length]
    return hero ? getIdentityColorHex(hero.identity) : undefined
}
