import {Asset, Step} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/function-test";

export enum ActionEnum {
    Sleep = "Sleep",
    PlayCard = "PlayCard",
    DisCard = "DisCard",
    OptRoomAction = "OptRoomAction",
    UseHeroSkill = "UseHeroSkill",
    WaitAction = "WaitAction",
    PlayCardOver = "PlayCardOver",
    OnlyAsset = "OnlyAsset",
}

export const isStepsHasAck = (step: Step) => {
    return step.action == ActionEnum.OptRoomAction
        || step.action == ActionEnum.PlayCard
        || step.action == ActionEnum.PlayCardOver
        || step.action == ActionEnum.DisCard
        || step.action == ActionEnum.UseHeroSkill
}

export const isAssetIsAck = (asset: Asset) => {
    return asset.msgName.endsWith("Ack")
}

export const chsAndEngSelectFilter = (pattern: string, option: any) => {
    if (/^[a-zA-Z]+.*/.test(pattern)) {
        return (option.value as string).toString().startsWith(pattern)
    } else {
        return (option.chs as string).includes(pattern)
    }
}

export const engAndSelectFilterSpecial = (pattern: string, option: any) => {
    if (/^[a-zA-Z]+.*/.test(pattern)) {
        return (option.value as string).includes(pattern)
    } else {
        return (option.int as string).toString().startsWith(pattern)
    }
}

export const actionsSelectOption = [
    {
        label: "等待",
        value: ActionEnum.Sleep,
        chs: "等待"
    },
    {
        label: "出牌",
        value: ActionEnum.PlayCard,
        chs: "出牌"
    },
    {
        label: "弃牌",
        value: ActionEnum.DisCard,
        chs: "弃牌"
    },
    {
        label: "响应",
        value: ActionEnum.OptRoomAction,
        chs: "响应"
    },
    {
        label: "使用技能",
        value: ActionEnum.UseHeroSkill,
        chs: "使用技能"
    },
    {
        label: "结束打牌",
        value: ActionEnum.PlayCardOver,
        chs: "结束打牌"
    },
    {
        label: "仅断言",
        value: ActionEnum.OnlyAsset,
        chs: "仅断言"
    },
    {
        label: "等待行动",
        value: ActionEnum.WaitAction,
        chs: "等待行动"
    },
]

export enum AssetEnum {
    RoomGameActionAsset = "RoomGameActionAsset", // 新增断言前一个自己的Action消息, 来断言某技能是否发动
    UpdateProperty = "UpdateProperty",
    UpdateHeroSkill = "UpdateHeroSkill",
    DrawCard = "DrawCard",
    PlayCard = "PlayCard",
    DisCard = "DisCard",
    TurnCard = "TurnCard",
    GiveCard = "GiveCard",
    CommonHpChange = "CommonHpChange",
    ReplaceCard = "ReplaceCard",
    SkillChange = "SkillChange",
    EquipChange = "EquipChange",
    AugurChange = "AugurChange",
    SeatChange = "SeatChange",
    AttrChange = "AttrChange",
    BuffChange = "BuffChange",
    CureDeadNotify = "CureDeadNotify",
    ChangeCardInfo = "ChangeCardInfo",
    ChangeIdentity = "ChangeIdentity",
    ChangeCountry = "ChangeCountry",
    PublicIdentity = "PublicIdentity",
    Ally = "Ally",
    ConfirmSkillTrigger = "ConfirmSkillTrigger",
    SkillTrigger = "SkillTrigger",
    HarmEffectTrigger = "HarmEffectTrigger",
    CardEnhance = "CardEnhance",
    TaiPingYaoShu = "TaiPingYaoShu",
    XianBaTouChou = "XianBaTouChou",
    QiMenDunJia = "QiMenDunJia",
    CancelSkill = "CancelSkill",
    TransferSkill = "TransferSkill",
    UserRelive = "UserRelive",
    ZiJianZhenXi = "ZiJianZhenXi",
    ResetUserSeat = "ResetUserSeat",
    ChangeWeather = "ChangeWeather",
    RefreshDisposeCards = "RefreshDisposeCards",
}

const assetAllSelectOption = [
    // 前一个动作的断言
    {
        label: '动作通知断言(RoomGameActionAsset)',
        value: AssetEnum.RoomGameActionAsset,
        chs: "动作通知断言",
    },
    // 两个例外UpdateProperty和xxAck, Ack动态生成
    {
        label: "属性更新(UpdateProperty)",
        value: AssetEnum.UpdateProperty,
        chs: "属性更新",
    },
    // 两个例外UpdateProperty和xxAck, Ack动态生成
    {
        label: "技能更新(UpdateHeroSkill)",
        value: AssetEnum.UpdateHeroSkill,
        chs: "技能更新",
    },
    // Ntf的Element
    {
        label: "抽卡(DrawCard)",
        value: AssetEnum.DrawCard,
        chs: "抽卡",
    },
    {
        label: "出牌(PlayCard)",
        value: AssetEnum.PlayCard,
        chs: "出牌",
    },
    {
        label: "弃牌(DisCard)",
        value: AssetEnum.DisCard,
        chs: "弃牌",
    },
    // {
    //     label: "转移牌(TurnCard)",
    //     value: AssetEnum.TurnCard,
    //     chs: "转移牌",
    // },
    {
        label: "给牌(GiveCard)",
        value: AssetEnum.GiveCard,
        chs: "给牌",
    },
    {
        label: "体力变化(CommonHpChange)",
        value: AssetEnum.CommonHpChange,
        chs: "体力变化",
    },
    {
        label: "替换牌(ReplaceCard)",
        value: AssetEnum.ReplaceCard,
        chs: "替换牌",
    },
    // {
    //     label: "技能变化(SkillChange)",
    //     value: AssetEnum.SkillChange,
    //     chs: "技能变化",
    // },
    {
        label: "装备变化(EquipChange)",
        value: AssetEnum.EquipChange,
        chs: "装备变化",
    },
    // {
    //     label: "卜卦变化(AugurChange)",
    //     value: AssetEnum.AugurChange,
    //     chs: "卜卦变化",
    // },
    // {
    //     label: "位置变化(SeatChange)",
    //     value: AssetEnum.SeatChange,
    //     chs: "位置变化",
    // },
    {
        label: "属性变化(AttrChange)",
        value: AssetEnum.AttrChange,
        chs: "属性变化",
    },
    // {
    //     label: "BUFF变化(BuffChange)",
    //     value: AssetEnum.BuffChange,
    //     chs: "BUFF变化",
    // },
    // {
    //     label: "重伤救援通知(CureDeadNotify)",
    //     value: AssetEnum.CureDeadNotify,
    //     chs: "重伤救援通知",
    // },
    // {
    //     label: "花色点数变化(ChangeCardInfo)",
    //     value: AssetEnum.ChangeCardInfo,
    //     chs: "花色点数变化",
    // },
    // {
    //     label: "身份变化(ChangeIdentity)",
    //     value: AssetEnum.ChangeIdentity,
    //     chs: "身份变化",
    // },
    {
        label: "势力变化(ChangeCountry)",
        value: AssetEnum.ChangeCountry,
        chs: "势力变化",
    },
    // {
    //     label: "公开身份通知(PublicIdentity)",
    //     value: AssetEnum.PublicIdentity,
    //     chs: "公开身份通知",
    // },
    // {
    //     label: "结盟(Ally)",
    //     value: AssetEnum.Ally,
    //     chs: "结盟",
    // },
    // {
    //     label: "确认发动技能(ConfirmSkillTrigger)",
    //     value: AssetEnum.ConfirmSkillTrigger,
    //     chs: "确认发动技能",
    // },
    {
        label: "技能触发(SkillTrigger)",
        value: AssetEnum.SkillTrigger,
        chs: "技能触发",
    },
    // {
    //     label: "伤害效果触发(HarmEffectTrigger)",
    //     value: AssetEnum.HarmEffectTrigger,
    //     chs: "伤害效果触发",
    // },
    {
        label: "卡牌强化(CardEnhance)",
        value: AssetEnum.CardEnhance,
        chs: "卡牌强化",
    },
    // {
    //     label: "太平要术(TaiPingYaoShu)",
    //     value: AssetEnum.TaiPingYaoShu,
    //     chs: "太平要术",
    // },
    {
        label: "先拔头筹(XianBaTouChou)",
        value: AssetEnum.XianBaTouChou,
        chs: "先拔头筹",
    },
    // {
    //     label: "奇门遁甲(QiMenDunJia)",
    //     value: AssetEnum.QiMenDunJia,
    //     chs: "奇门遁甲",
    // },
    // {
    //     label: "取消技能通知(CancelSkill)",
    //     value: AssetEnum.CancelSkill,
    //     chs: "取消技能通知",
    // },
    // {
    //     label: "转移技能通知(TransferSkill)",
    //     value: AssetEnum.TransferSkill,
    //     chs: "转移技能通知",
    // },
    // {
    //     label: "复活通知(UserRelive)",
    //     value: AssetEnum.UserRelive,
    //     chs: "复活通知",
    // },
    // {
    //     label: "自荐枕席(ZiJianZhenXi)",
    //     value: AssetEnum.ZiJianZhenXi,
    //     chs: "自荐枕席",
    // },
    // {
    //     label: "重置座位(ResetUserSeat)",
    //     value: AssetEnum.ResetUserSeat,
    //     chs: "重置座位",
    // },
    // {
    //     label: "天气变化(ChangeWeather)",
    //     value: AssetEnum.ChangeWeather,
    //     chs: "天气变化",
    // },
    // {
    //     label: "刷新处理区(RefreshDisposeCards)",
    //     value: AssetEnum.RefreshDisposeCards,
    //     chs: "刷新处理区",
    // },
]

export const assetSelectOption = (step: Step) => {
    // 把一些行动带有Ack的添加到所有SelectOption前面
    return (isStepsHasAck(step)) ? [{
        label: `结果(${step.action}Ack)`,
        value: `${step.action}Ack`,
        chs: '结果',
    }, ...assetAllSelectOption,
    ] : assetAllSelectOption
}