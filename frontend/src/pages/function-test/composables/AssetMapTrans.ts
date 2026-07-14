import {AssetEnum} from "./StepActionsAndAssetsSelect";
import {Ref} from "vue";
import {Asset} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/function-test";

const numberArrD = (key: string, asset: Ref<Asset>, assetList: Ref, def?: number[]) => {
    if (asset.value.attr[key] != null)
        assetList.value[key] = asset.value.attr[key] == "" ? [] : asset.value.attr[key].split(" ").filter(n => isFinite(Number(n))).map(c => Number(c)) || []
    else if (def != null)
        assetList.value[key] = def
}

const numberArrU = (key: string, asset: Ref<Asset>, assetList: Ref) => {
    if (assetList.value[key] && Array.isArray(assetList.value[key]))
        asset.value.attr[key] = (assetList.value[key]).filter(n => isFinite(n)).join(" ")
}

const numberArr2D = (key: string, asset: Ref<Asset>, assetList: Ref, def?: number[][]) => {
    if (asset.value.attr[key] != null) {
        const value = asset.value.attr[key];
        if (value === "") {
            assetList.value[key] = [];
        } else {
            // 用逗号分隔外层数组，再用空格分隔内层数组
            const outerArr = value.split(",");
            assetList.value[key] = outerArr.map(innerStr => {
                // 解析内层数字数组
                return innerStr.trim()
                    .split(/\s+/) // 匹配一个或多个空白字符
                    .filter(n => isFinite(Number(n)))
                    .map(c => Number(c));
            }).filter(innerArr => innerArr.length > 0); // 过滤掉空数组
        }
    } else if (def != null) {
        assetList.value[key] = def;
    }
}

const numberArr2DU = (key: string, asset: Ref<Asset>, assetList: Ref) => {
    if (assetList.value[key] && Array.isArray(assetList.value[key])) {
        const arr2D = assetList.value[key];
        // 将二维数组序列化为字符串：内层用空格分隔，外层用逗号分隔
        asset.value.attr[key] = arr2D
            .map(innerArr => {
                // 过滤有效数字并转换为字符串
                return innerArr
                    .filter(n => isFinite(n))
                    .map(n => n.toString())
                    .join(" ");
            })
            .join(",");
    }
}

const boolArrD = (key: string, asset: Ref<Asset>, assetList: Ref, def?: boolean[]) => {
    if (asset.value.attr[key] != null) {
        const value = asset.value.attr[key];
        if (value === "") {
            assetList.value[key] = [];
        } else {
            // 按空格分割，转换为布尔值
            assetList.value[key] = value.split(/\s+/)
                .map(item => item.trim())
                .filter(item => item !== "")  // 过滤空字符串
                .map(item => item.toLowerCase() === "true");
        }
    } else if (def != null) {
        assetList.value[key] = def;
    }
}

const boolArrU = (key: string, asset: Ref<Asset>, assetList: Ref) => {
    if (assetList.value[key] && Array.isArray(assetList.value[key])) {
        const arr = assetList.value[key] as boolean[];
        // 将布尔数组转换为字符串，用空格分隔
        asset.value.attr[key] = arr
            .map(item => item ? "true" : "false")
            .join(" ");
    }
}

const numberD = (key: string, asset: Ref<Asset>, assetList: Ref, def?: number) => {
    if (Number.isFinite(Number(asset.value.attr[key])))
        assetList.value[key] = Number(asset.value.attr[key]) || 0
    else if (def != null)
        assetList.value[key] = def
}

const numberU = (key: string, asset: Ref<Asset>, assetList: Ref) => {
    if (assetList.value[key] != null && Number.isFinite(assetList.value[key]))
        asset.value.attr[key] = assetList.value[key].toString()
}

const strD = (key: string, asset: Ref<Asset>, assetList: Ref, def?: string) => {
    if (asset.value.attr[key] != null)
        assetList.value[key] = asset.value.attr[key]
    else if (def != null)
        assetList.value[key] = def
}

const strU = (key: string, asset: Ref<Asset>, assetList: Ref) => {
    if (assetList.value[key] != null)
        asset.value.attr[key] = assetList.value[key]
}

const boolD = (key: string, asset: Ref<Asset>, assetList: Ref, def?: boolean) => {
    if (asset.value.attr[key] != null)
        assetList.value[key] = asset.value.attr[key] === "true"
    else if (def != null)
        assetList.value[key] = def
}

const boolU = (key: string, asset: Ref<Asset>, assetList: Ref) => {
    if (assetList.value[key] != null)
        asset.value.attr[key] = (assetList.value[key] === true).toString()
}

const ackD = (asset: Ref<Asset>, assetList: Ref) => {
    numberArrD("Result", asset, assetList, [0])
}

const ackU = (asset: Ref<Asset>, assetList: Ref, ackBool: Ref<boolean>) => {
    // Ack类断言 TODO 如果ackBool是false且后面的result没有值会没有Resultkey
    if (ackBool.value)
        asset.value.attr['Result'] = '0'
    else if (!ackBool.value && !('Result' in assetList.value))
        asset.value.attr['Result'] = '0'
    else numberArrU("Result", asset, assetList)
}

const roomGameActionAssetD = (asset: Ref<Asset>, assetList: Ref) => {
    numberD("ActionType", asset, assetList, 0)
    numberD("ActionValue", asset, assetList, 0)
    numberArrD("Params", asset, assetList)
    boolD("UnExpect", asset, assetList)
}

const roomGameActionAssetU = (asset: Ref<Asset>, assetList: Ref) => {
    numberU("ActionType", asset, assetList)
    numberU("ActionValue", asset, assetList)
    numberArrU("Params", asset, assetList)
    boolU("UnExpect", asset, assetList)
}

const updatePropertyD = (asset: Ref<Asset>, assetList: Ref) => {
    strD("IdType", asset, assetList, 'None')
    strD("PropID", asset, assetList, 'None')
    numberD("PropValue", asset, assetList, 0)
    boolD("UnExpect", asset, assetList)
}

const updatePropertyU = (asset: Ref<Asset>, assetList: Ref) => {
    // 更新属性类断言
    strU("IdType", asset, assetList)
    strU("PropID", asset, assetList)
    numberU("PropValue", asset, assetList)
    boolU("UnExpect", asset, assetList)
}

const normalCardD = (asset: Ref<Asset>, assetList: Ref) => {
    if (asset.value.msgName == AssetEnum.DrawCard) {
        numberD("ActionType", asset, assetList, 0)
        numberD("ActionValue", asset, assetList, 0)
    }
    numberD("Count", asset, assetList, 0)
    numberArrD("Cards", asset, assetList)
    numberArrD("UnexpectCards", asset, assetList)
    boolD("Random", asset, assetList)
    boolD("UnExpect", asset, assetList)
}

const normalCardU = (asset: Ref<Asset>, assetList: Ref) => {
    // 一般卡牌类
    if (asset.value.msgName == AssetEnum.DrawCard) {
        numberU("ActionType", asset, assetList)
        numberU("ActionValue", asset, assetList)
    }
    numberU("Count", asset, assetList)
    numberArrU("Cards", asset, assetList)
    numberArrU("UnexpectCards", asset, assetList)
    boolU("Random", asset, assetList)
    boolU("UnExpect", asset, assetList)
}

const replaceDeserialize = (asset: Ref<Asset>, assetList: Ref, key: string) => {
    if (asset.value.attr[key])
        // 123:123,245|true 345:123456,123|false 转为 Map<number, [number[], bool]>
        assetList.value[key] = Object.fromEntries(
            asset.value.attr[key].split(" ").filter(cs => {
                const sp = cs.split(":")
                return sp.length > 0 && Number.isFinite(Number(sp[0]))
            }).map(cs => {
                const kv = cs.split(":")
                return [kv[0], kv[1] ? kv[1].split(",").filter(c => Number.isFinite(Number(c))).map(c => Number(c)) : []]
            })
        ) || {}
    else assetList.value[key] = {}
}

const replaceSerialize = (asset: Ref<Asset>, assetList: Ref, key: string) => {
    if (assetList.value[key] && Object.keys(assetList.value[key] as { [key: number]: number[] }).length > 0) {
        // 过滤数值
        Object.values(assetList.value[key] as { [key: number]: number[] })
            .forEach((arr) => {
                if (!Array.isArray(arr)) {
                    return
                }
                // 直接修改原数组
                for (let i = arr.length - 1; i >= 0; i--) {
                    if (!Number.isFinite(arr[i])) {
                        arr.splice(i, 1);
                    }
                }
            })

        asset.value.attr[key] = Object.entries(assetList.value[key] as { [key: number]: number[] })
            .filter(([k, v]) => Number.isFinite(Number(k)) && Array.isArray(v))
            .map(([k, v]) => {
                return `${k}:${(v.join(","))}`
            }).join(" ")
        // console.log(asset.value.attr['ReplaceCards_ChangeCardsToCards'])
    }
}

const replaceCardD = (asset: Ref<Asset>, assetList: Ref) => {
    numberArrD("ReplaceCards_ChangeCards", asset, assetList)
    numberArrD("ReplaceCards_DrawCards", asset, assetList)
    replaceDeserialize(asset, assetList, "ReplaceCards_ChangeCardsToCards")
    replaceDeserialize(asset, assetList, "ReplaceCards_ChangeCardsToPoints")
    replaceDeserialize(asset, assetList, "ReplaceCards_ChangeCardsToAttrTypes")
    replaceDeserialize(asset, assetList, "ReplaceCards_DrawCardsToCards")
    replaceDeserialize(asset, assetList, "ReplaceCards_DrawCardsToPoints")
    replaceDeserialize(asset, assetList, "ReplaceCards_DrawCardsToAttrTypes")
    boolD("ReplaceCards_ChangeCardsRandom", asset, assetList)
    boolD("ReplaceCards_DrawCardsRandom", asset, assetList)
}

const replaceCardU = (asset: Ref<Asset>, assetList: Ref) => {
    // 替换卡牌类
    numberArrU("ReplaceCards_ChangeCards", asset, assetList)
    numberArrU("ReplaceCards_DrawCards", asset, assetList)
    replaceSerialize(asset, assetList, "ReplaceCards_ChangeCardsToCards")
    replaceSerialize(asset, assetList, "ReplaceCards_ChangeCardsToPoints")
    replaceSerialize(asset, assetList, "ReplaceCards_ChangeCardsToAttrTypes")
    replaceSerialize(asset, assetList, "ReplaceCards_DrawCardsToCards")
    replaceSerialize(asset, assetList, "ReplaceCards_DrawCardsToPoints")
    replaceSerialize(asset, assetList, "ReplaceCards_DrawCardsToAttrTypes")
    boolU("ReplaceCards_ChangeCardsRandom", asset, assetList)
    boolU("ReplaceCards_DrawCardsRandom", asset, assetList)
}


const updateHeroSkillDeserialize = (asset: Ref<Asset>, assetList: Ref) => {
    // 确保 CardParamMap 存在
    if (!assetList.value.CardParamMap) {
        assetList.value.CardParamMap = {};
    }

    const skillUUids = assetList.value['SkillUUid'] || [];

    if (skillUUids.length > 0) {
        // 解析 CardParamList (卡牌参数列表)
        const paramListStr = asset.value.attr['CardParamList'] || "";
        const paramListArr = paramListStr === "" ? [] : paramListStr.split(",");

        // 解析 IsInvalid (是否无效)
        const IsInvalidStr = asset.value.attr['IsInvalid'] || "";
        const IsInvalidArr = IsInvalidStr === "" ? [] : IsInvalidStr.split(",");

        // 确保数组长度一致
        const maxLength = Math.max(paramListArr.length, IsInvalidArr.length, skillUUids.length);

        skillUUids.forEach((uuid, i) => {
            if (i < maxLength) {
                // 解析卡牌参数
                const paramStr = paramListArr[i] || "";
                const paramArray = paramStr === "" ? [] :
                    paramStr.split(/\s+/)
                        .map(n => Number(n))
                        .filter(n => !isNaN(n));  // 更安全的过滤

                // 解析是否有效
                const IsInvalidStr = IsInvalidArr[i] || "false";
                const IsInvalid = IsInvalidStr.toLowerCase() === "true";

                assetList.value.CardParamMap[uuid] = {
                    CardParamList: paramArray,
                    IsInvalid: IsInvalid
                };
            } else {
                // 如果数据不足，使用默认值
                assetList.value.CardParamMap[uuid] = {
                    CardParamList: [],
                    IsInvalid: false
                };
            }
        });
    } else {
        assetList.value.CardParamMap = {};
    }
}

const updateHeroSkillSerialize = (asset: Ref<Asset>, assetList: Ref) => {
    // 序列化 CardParamMap
    const data = assetList.value.CardParamMap || {};
    const skillUUids = assetList.value['SkillUUid'] || [];

    const CardParamList: string[] = [];
    const isValidList: string[] = [];

    skillUUids.forEach(uuid => {
        const item = data[uuid];
        if (item) {
            // 卡牌参数：用空格分隔数字
            const paramStr = Array.isArray(item.CardParamList)
                ? item.CardParamList.join(" ")
                : "";
            CardParamList.push(paramStr);

            // 是否无效：转换为字符串
            isValidList.push(item.IsInvalid === true ? "true" : "false");
        } else {
            // 如果没有对应数据，使用默认值
            CardParamList.push("");
            isValidList.push("false");
        }
    });

    // 更新 asset 属性
    asset.value.attr['CardParamList'] = CardParamList.join(",");
    asset.value.attr['IsInvalid'] = isValidList.join(",");
}

const updateHeroSkillD = (asset: Ref<Asset>, assetList: Ref) => {
    numberArrD("SkillUUid", asset, assetList)
    boolD("UnExpect", asset, assetList)
    updateHeroSkillDeserialize(asset, assetList)
}

const updateHeroSkillU = (asset: Ref<Asset>, assetList: Ref) => {
    // 替换卡牌类
    numberArrU("SkillUUid", asset, assetList)
    boolD("UnExpect", asset, assetList);
    updateHeroSkillSerialize(asset, assetList)
}

const commonHpChangeD = (asset: Ref<Asset>, assetList: Ref) => {
    strD("HpSrc", asset, assetList)
    numberD("ChangeHp", asset, assetList)
    numberD("CurHp", asset, assetList)
    numberD("MaxHp", asset, assetList)
    boolD("UnExpect", asset, assetList)
}

const commonHpChangeU = (asset: Ref<Asset>, assetList: Ref) => {
    // 生命值变化
    strU("HpSrc", asset, assetList)
    numberU("ChangeHp", asset, assetList)
    numberU("CurHp", asset, assetList)
    numberU("MaxHp", asset, assetList)
    boolU("UnExpect", asset, assetList)
}

const attrChangeD = (asset: Ref<Asset>, assetList: Ref) => {
    numberD("ShaCount", asset, assetList)
    numberD("HandLimit", asset, assetList)
    numberD("EquipLimit", asset, assetList)
    numberD("AttackRange", asset, assetList)
    numberD("DistIncr", asset, assetList)
    numberD("DistDecr", asset, assetList)
    numberD("MaxHp", asset, assetList)
    numberD("CurHp", asset, assetList)
    boolD("UnExpect", asset, assetList)
}

const attrChangeU = (asset: Ref<Asset>, assetList: Ref) => {
    // 属性变化
    numberU("ShaCount", asset, assetList)
    numberU("HandLimit", asset, assetList)
    numberU("EquipLimit", asset, assetList)
    numberU("AttackRange", asset, assetList)
    numberU("DistIncr", asset, assetList)
    numberU("DistDecr", asset, assetList)
    numberU("MaxHp", asset, assetList)
    numberU("CurHp", asset, assetList)
    boolU("UnExpect", asset, assetList)
}

const changeCountryD = (asset: Ref<Asset>, assetList: Ref) => {
    numberArrD("MainCountry", asset, assetList)
    numberArrD("ExtraCountry", asset, assetList)
}

const changeCountryU = (asset: Ref<Asset>, assetList: Ref) => {
    // 势力变化
    numberArrU("MainCountry", asset, assetList)
    numberArrU("ExtraCountry", asset, assetList)
}

const skillTriggerD = (asset: Ref<Asset>, assetList: Ref) => {
    numberD("ActionValue", asset, assetList)
    numberArrD("DestSeatIds", asset, assetList)
    numberD("Param", asset, assetList)
    boolD("Random", asset, assetList)
    boolD("UnExpect", asset, assetList)
}

const skillTriggerU = (asset: Ref<Asset>, assetList: Ref) => {
    // 势力变化
    numberU("ActionValue", asset, assetList)
    numberArrU("DestSeatIds", asset, assetList)
    numberU("Param", asset, assetList)
    boolU("Random", asset, assetList)
    boolU("UnExpect", asset, assetList)
}

const equipChangeD = (asset: Ref<Asset>, assetList: Ref) => {
    numberD("AddEquip", asset, assetList)
    numberD("Count", asset, assetList)
    numberArrD("RemoveEquip", asset, assetList)
    boolD("Random", asset, assetList)
    boolD("UnExpect", asset, assetList)
}

const equipChangeU = (asset: Ref<Asset>, assetList: Ref) => {
    // 装备变化
    numberU("AddEquip", asset, assetList)
    numberU("Count", asset, assetList)
    numberArrU("RemoveEquip", asset, assetList)
    boolU("Random", asset, assetList)
    boolU("UnExpect", asset, assetList)
}

const normalCardAssetType = (asset: Ref<Asset>) => {
    return normalCardAssetValueType(asset.value)
}

export const normalCardAssetValueType = (asset: Asset) => {
    return asset.msgName == AssetEnum.DrawCard
        || asset.msgName == AssetEnum.PlayCard
        || asset.msgName == AssetEnum.DisCard
        || asset.msgName == AssetEnum.GiveCard
        || asset.msgName == AssetEnum.XianBaTouChou
        || asset.msgName == AssetEnum.CardEnhance
}

export const assetMap2AssetList = (asset: Ref<Asset>, assetList: Ref) => {
    if (asset.value.msgName.endsWith("Ack")) {
        ackD(asset, assetList)
    } else if (asset.value.msgName == AssetEnum.RoomGameActionAsset) {
        roomGameActionAssetD(asset, assetList)
    } else if (asset.value.msgName == AssetEnum.UpdateProperty) {
        updatePropertyD(asset, assetList)
    } else if (asset.value.msgName == AssetEnum.UpdateHeroSkill) {
        updateHeroSkillD(asset, assetList)
    } else if (!asset.value.msgName.endsWith("Ack") && asset.value.msgName != AssetEnum.UpdateProperty) {
        // 33+分支

        if (normalCardAssetType(asset)) {
            // 一般卡牌
            normalCardD(asset, assetList)
        } else if (asset.value.msgName == AssetEnum.ReplaceCard) {
            // 替换
            replaceCardD(asset, assetList)
        } else if (asset.value.msgName == AssetEnum.CommonHpChange) {
            // 生命
            commonHpChangeD(asset, assetList)
        } else if (asset.value.msgName == AssetEnum.AttrChange) {
            // 属性
            attrChangeD(asset, assetList)
        } else if (asset.value.msgName == AssetEnum.ChangeCountry) {
            // 势力
            changeCountryD(asset, assetList)
        } else if (asset.value.msgName == AssetEnum.SkillTrigger) {
            // 技能触发
            skillTriggerD(asset, assetList)
        } else if (asset.value.msgName == AssetEnum.EquipChange) {
            // 装备改变
            equipChangeD(asset, assetList)
        } else {

        }
        if (asset.value.attr['UnExpect']) {
            assetList.value['UnExpect'] = asset.value.attr['UnExpect'] == "true"
        }
    }
}

export const assetList2AssetMap = (asset: Ref<Asset>, assetList: Ref, ackBool: Ref<boolean>) => {
    // 首先清空AssetMap
    if (asset.value)
        asset.value.attr = {}

    // 同步到Map
    if (asset.value.msgName.endsWith("Ack")) {
        ackU(asset, assetList, ackBool)
    } else if (asset.value.msgName == AssetEnum.RoomGameActionAsset) {
        roomGameActionAssetU(asset, assetList)
    } else if (asset.value.msgName == AssetEnum.UpdateProperty) {
        updatePropertyU(asset, assetList)
    } else if (asset.value.msgName == AssetEnum.UpdateHeroSkill) {
        updateHeroSkillU(asset, assetList)
    } else if (normalCardAssetType(asset)) {
        normalCardU(asset, assetList)
    } else if (asset.value.msgName == AssetEnum.ReplaceCard) {
        replaceCardU(asset, assetList)
    } else if (asset.value.msgName == AssetEnum.CommonHpChange) {
        commonHpChangeU(asset, assetList)
    } else if (asset.value.msgName == AssetEnum.AttrChange) {
        attrChangeU(asset, assetList)
    } else if (asset.value.msgName == AssetEnum.ChangeCountry) {
        changeCountryU(asset, assetList)
    } else if (asset.value.msgName == AssetEnum.SkillTrigger) {
        skillTriggerU(asset, assetList)
    } else if (asset.value.msgName == AssetEnum.EquipChange) {
        equipChangeU(asset, assetList)
    } else {

    }
    if (assetList.value.UnExpect)
        asset.value!.attr['UnExpect'] = assetList.value.UnExpect.toString()
}