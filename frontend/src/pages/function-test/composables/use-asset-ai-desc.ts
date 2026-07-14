/**
 * 资产智能描述生成 composable
 *
 * 根据资产类型和当前数据，生成可读的中文描述文字。
 * 提取自 asset-card.vue 的 aiDesc() 函数。
 */
import {Asset, Step} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/function-test";
import {actionsSelectOption, AssetEnum, assetSelectOption, isAssetIsAck} from "./StepActionsAndAssetsSelect";
import {nowCaseData} from "./use-case-data";
import {excelHeroMap} from "@shared/config/hero";
import {excelCardsMap} from "../config/Card";
import {excelSkillMap} from "../config/Skill";
import {countryMap} from "../config/ECountry";

/**
 * 获取当前 step 对应的武将名称
 */
const getHeroName = (step: Step): string => {
    if (!nowCaseData.value?.initYanWu?.customHeroes?.length) return ""
    const heroId = nowCaseData.value.initYanWu.customHeroes[
        (step.robotIdx - 1) % nowCaseData.value.initYanWu.customHeroes.length
        ]?.heroId
    return heroId ? (excelHeroMap.value[heroId]?.Name || "") : ""
}

/**
 * 生成一般卡牌类（DrawCard/DisCard/PlayCard/GiveCard/CardEnhance/XianBaTouChou）的描述
 * @param verb 动词，如"抽到"、"弃掉"、"打出"等
 */
const generateCardDesc = (assetList: { [key: string]: any }, verb: string): string => {
    return (assetList["Count"] ? verb + assetList["Count"] + "张牌" : "")
        + (assetList["Count"] && assetList["Cards"] ? "牌为" : !assetList["Count"] && assetList["Cards"] ? verb : "")
        + (assetList["Cards"] ? `[${assetList["Cards"].toString()}]` : "")
}

/**
 * 生成替换卡牌段的描述（ChangeCards 或 DrawCards）
 */
const generateReplaceCardSegment = (
    assetList: { [key: string]: any },
    cardListKey: string,
    toCardsKey: string,
    toPointsKey: string,
    toAttrTypesKey: string,
    randomKey: string
): string => {
    const fromCards = assetList[cardListKey]
    if (!fromCards || fromCards.length === 0) return ""

    return fromCards.map((c: number) => {
        let toCards = ""
        if (assetList[toCardsKey]?.[c]?.length > 0) {
            toCards = assetList[toCardsKey][c].length === 1
                ? assetList[toCardsKey][c][0]
                : assetList[toCardsKey][c].map((c: number) => `${excelSkillMap.value[c]?.SkillName}(${c})`).join(",") + "(包含)"
        }
        let toPoints = ""
        if (assetList[toPointsKey]?.[c]?.length > 0) {
            toPoints = assetList[toPointsKey][c].length === 1
                ? assetList[toPointsKey][c][0]
                : assetList[toPointsKey][c].map((c: number) => `${c}`).join(",") + "(包含)"
        }
        let toAttrTypes = ""
        if (assetList[toAttrTypesKey]?.[c]?.length > 0) {
            toAttrTypes = assetList[toAttrTypesKey][c].length === 1
                ? assetList[toAttrTypesKey][c][0]
                : assetList[toAttrTypesKey][c].map((c: number) => `${c}`).join(",") + "(包含)"
        }
        return (toCards || toPoints || toAttrTypes)
            ? `[${excelCardsMap.value[c]?.Name}(${c}) 替换为: [${toCards}]${toPoints ? ', 点数:[' + toPoints + ']' : ''}${toAttrTypes ? ', 属性[' + toAttrTypes + ']' : ''} ]`
            : `替换${excelCardsMap.value[c]?.Name}(${c})`
    }).join("; ") + (assetList[randomKey] ? "(任意)" : "")
}

/**
 * 生成资产的智能描述文字
 *
 * @param asset 当前资产对象
 * @param assetList 资产的列表形式数据（响应式）
 * @param step 当前步骤
 * @returns 可读的中文描述字符串
 */
export function generateAiDesc(
    asset: Asset,
    assetList: { [key: string]: any },
    step: Step
): string {
    const action = step.action
    const assetName = asset.msgName

    const stepActionChs = actionsSelectOption.find(o => o.value == action)
    const assetActionChs = assetSelectOption(step).find(o => o.value == assetName)
    const heroName = getHeroName(step)

    if (isAssetIsAck(asset)) {
        return stepActionChs ? heroName + stepActionChs.label + "成功" : ""
    }

    if (assetName == AssetEnum.DrawCard) {
        return assetActionChs ? heroName + generateCardDesc(assetList, "抽到") : ""
    }
    if (assetName == AssetEnum.DisCard) {
        return assetActionChs ? heroName + generateCardDesc(assetList, "弃掉") : ""
    }
    if (assetName == AssetEnum.PlayCard) {
        return assetActionChs ? heroName + generateCardDesc(assetList, "打出") : ""
    }
    if (assetName == AssetEnum.GiveCard) {
        return assetActionChs ? heroName + generateCardDesc(assetList, "给") : ""
    }
    if (assetName == AssetEnum.CommonHpChange) {
        return assetActionChs ? heroName
            + (assetList["HpSrc"] ? "受到来自" + assetList["HpSrc"] + "的伤害" : "")
            + (assetList["ChangeHp"] ? ", 体力变化:" + assetList["ChangeHp"] : "")
            + (assetList["CurHp"] ? `, 当前体力: [${assetList["CurHp"]}]` : "")
            + (assetList["MaxHp"] ? `, 最大体力: [${assetList["MaxHp"]}]` : "")
            : ""
    }
    if (assetName == AssetEnum.ReplaceCard) {
        const cc = generateReplaceCardSegment(assetList, "ReplaceCards_ChangeCards",
            "ReplaceCards_ChangeCardsToCards", "ReplaceCards_ChangeCardsToPoints",
            "ReplaceCards_ChangeCardsToAttrTypes", "ReplaceCards_ChangeCardsRandom")
        const dc = generateReplaceCardSegment(assetList, "ReplaceCards_DrawCards",
            "ReplaceCards_DrawCardsToCards", "ReplaceCards_DrawCardsToPoints",
            "ReplaceCards_DrawCardsToAttrTypes", "ReplaceCards_DrawCardsRandom")

        return assetActionChs ? heroName
            + (cc ? "替换卡牌:" + cc : "")
            + (dc && cc ? ", 同时生成卡牌:" + dc : dc ? "生成卡牌" + dc : "")
            : ""
    }
    if (assetName == AssetEnum.AttrChange) {
        return assetActionChs ? heroName
            + (assetList["ShaCount"] ? ", 出杀次数: " + assetList["ShaCount"] : "")
            + (assetList["HandLimit"] ? ", 手牌上限: " + assetList["HandLimit"] : "")
            + (assetList["EquipLimit"] ? ", 装备上限: " + assetList["EquipLimit"] : "")
            + (assetList["AttackRange"] ? ", 攻击范围: " + assetList["AttackRange"] : "")
            + (assetList["DistIncr"] ? ", 距离增加: " + assetList["DistIncr"] : "")
            + (assetList["DistDecr"] ? ", 距离减少: " + assetList["DistDecr"] : "")
            + (assetList["MaxHp"] ? ", 最大生命: " + assetList["MaxHp"] : "")
            + (assetList["CurHp"] ? ", 当前生命: " + assetList["CurHp"] : "")
            : ""
    }
    if (assetName == AssetEnum.ChangeCountry) {
        return assetActionChs ? heroName
            + (countryMap[assetList["MainCountry"]] ? ", 主要势力: " + countryMap[assetList["MainCountry"]] : "")
            + (countryMap[assetList["ExtraCountry"]] ? ", 额外势力: " + countryMap[assetList["ExtraCountry"]] : "")
            : ""
    }
    if (assetName == AssetEnum.CardEnhance) {
        return assetActionChs ? heroName + generateCardDesc(assetList, "强化") : ""
    }
    if (assetName == AssetEnum.SkillTrigger) {
        return assetActionChs ? heroName
            + (assetList["DestSeatIds"] ? `发动技能${assetList["ActionValue"] && excelSkillMap.value[assetList["ActionValue"]] ? excelSkillMap.value[assetList["ActionValue"]].SkillName : ''}, 目标座位号: [${assetList["DestSeatIds"]}]` : "")
            + (assetList["Random"] && assetList["DestSeatIds"] ? "(包含)" : "")
            + (assetList["Param"] ? `, 参数[${assetList["Param"].toString()}]` : "")
            : ""
    }
    if (assetName == AssetEnum.XianBaTouChou) {
        return assetActionChs ? heroName + generateCardDesc(assetList, "先拔头筹拿到") : ""
    }
    if (assetName == AssetEnum.EquipChange) {
        const re = assetList["RemoveEquip"] ? assetList["RemoveEquip"].map((eid: number) => {
            return excelCardsMap.value[eid]?.Name ? `[${excelCardsMap.value[eid]?.Name}](${eid})` : eid
        }).join(" ") : ""

        return assetActionChs ? heroName
            + (assetList["AddEquip"] ? `获得装备[${excelCardsMap.value[Number(assetList["AddEquip"])]?.Name}](${assetList["AddEquip"]})` : "")
            + (assetList["Count"] ? `, 失去${assetList["Count"]}件装备` : "")
            + (assetList["RemoveEquip"] ? `, 失去装备[${re}]` : "")
            : ""
    }

    return "请输入描述"
}
