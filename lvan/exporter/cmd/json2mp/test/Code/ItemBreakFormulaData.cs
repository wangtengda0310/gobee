using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class ItemBreakFormulaData : IDataBase
{
    /// <summary>
    /// 合成配方id
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 合成配方名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 配方描述
    /// </summary>
    [Key(2)]
    public string desc;

    /// <summary>
    /// 配方类别id
    /// </summary>
    [Key(3)]
    public UInt32 categoryId;

    /// <summary>
    /// 合成系统id
    /// </summary>
    [Key(4)]
    public UInt32 classId;

    /// <summary>
    /// 反应类型
    /// </summary>
    [Key(5)]
    public byte reactionType;

    /// <summary>
    /// 品质
    /// </summary>
    [Key(6)]
    public UInt32 quality;

    /// <summary>
    /// 开启条件描述
    /// </summary>
    [Key(7)]
    public string openConditionDesc;

    /// <summary>
    /// 源物品
    /// </summary>
    [Key(8)]
    public List<UInt32> source;

    /// <summary>
    /// 源物品选择器
    /// </summary>
    [Key(9)]
    public UInt32[][] sourceSelector;

    /// <summary>
    /// 材料物品数量
    /// </summary>
    [Key(10)]
    public UInt32 materialNum;

    /// <summary>
    /// 材料物品配置ID
    /// </summary>
    [Key(11)]
    public List<UInt32> material;

    /// <summary>
    /// 材料物品选择器数量
    /// </summary>
    [Key(12)]
    public List<UInt32> materialSelectorNum;

    /// <summary>
    /// 材料物品选择器
    /// </summary>
    [Key(13)]
    public UInt32[][] materialSelector;

    /// <summary>
    /// 消耗组
    /// </summary>
    [Key(14)]
    public UInt32[][] costGroup;

    /// <summary>
    /// 消耗
    /// </summary>
    [Key(15)]
    public UInt32[][] cost;

    /// <summary>
    /// 消耗描述
    /// </summary>
    [Key(16)]
    public string costDesc;

    /// <summary>
    /// 目标物品
    /// </summary>
    [Key(17)]
    public UInt32 target;

    /// <summary>
    /// 固定产出
    /// </summary>
    [Key(18)]
    public UInt32[][] fixedItemGain;

    /// <summary>
    /// 掉落产出
    /// </summary>
    [Key(19)]
    public List<UInt32> dropItemGain;

    /// <summary>
    /// 产出描述
    /// </summary>
    [Key(20)]
    public string gainDesc;

    /// <summary>
    /// 道具产出显示
    /// </summary>
    [Key(21)]
    public UInt32[][] gainItemDisplay;

    /// <summary>
    /// 物品消耗显示
    /// </summary>
    [Key(22)]
    public UInt32[][] costItemDisplay;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
