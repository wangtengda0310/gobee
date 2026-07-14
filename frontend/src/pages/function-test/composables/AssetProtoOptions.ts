import {errCodeMap} from "../config/ErrorCode";
import {countryList} from "../config/ECountry";
import {GameExcelService} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/game";

export const updatePropertyOptions = [
    {
        label: "None(0)",
        value: "None",
        chs: "None",
        int: 0
    },
    {
        label: "Hero(1)",
        value: "Hero",
        chs: "Hero",
        int: 1
    },
    {
        label: "Card(2)",
        value: "Card",
        chs: "Card",
        int: 2
    },
    {
        label: "User(3)",
        value: "User",
        chs: "User",
        int: 3
    },
    {
        label: "Room(4)",
        value: "Room",
        chs: "Room",
        int: 4
    },
    {
        label: "Max(5)",
        value: "Max",
        chs: "Max",
        int: 5
    },
]

// 从pb的 PropertyType_name 拿
let ids1: { [p: `${number}`]: string } = {}
let ids2: { [p: `${number}`]: string } = {}

await Promise.all([
    GameExcelService.GetPropertyTypeMap().then(res => {
        ids1 = res as { [p: `${number}`]: string }
    }).catch(err => {
        console.log("获取proto失败")
    }),
    GameExcelService.GetOptActionTypeMap().then(res => {
        ids2 = res as { [p: `${number}`]: string }
    }).catch(err => {
        console.log("获取proto失败")
    })
])

export const propertyValueOptions = Object.entries(ids1).map(([k,v])=>{
    return {label: `${v}(${k})`, value: v, chs: v, int: Number(k)}
})

export const optActionTypeOptions = Object.entries(ids2).map(([k,v])=>{
    return {label: `${v}(${k})`, value: Number(k),}
})

export const hpSrcOptions = [
    {label: "空", value: "HpNone",},
    {label: "普通伤害", value: "HarmNormal",},
    {label: "火伤", value: "HarmHuo",},
    {label: "雷伤", value: "HarmLei",},
    {label: "治疗Hp", value: "CureHp",},
    {label: "失去HP", value: "LostHp",},
    {label: "断角獬豸", value: "DuanJiaoXieZhi",},
]

export const errCodeOptions = Object.keys(errCodeMap).map(k => {
    return {
        label: errCodeMap[Number(k)] + `(${k})`,
        value: Number(k),
        chs: errCodeMap[Number(k)]
    }
})

export const countryOptions = countryList.map(([id, chs]) => {
    return {
        label: chs + `(${id})`,
        value: id,
        chs
    }
})