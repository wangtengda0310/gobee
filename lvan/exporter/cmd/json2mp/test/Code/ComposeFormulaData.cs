using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class ComposeFormulaData : IDataBase
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
    /// 描述
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
    public UInt32 reactionType;

    /// <summary>
    /// 品质
    /// </summary>
    [Key(6)]
    public UInt32 quality;

    /// <summary>
    /// 产出显示
    /// </summary>
    [Key(7)]
    public UInt32[][] gainDisplayItem;

    /// <summary>
    /// 是否显示数量
    /// </summary>
    [Key(8)]
    public byte isNumDisplay;

    /// <summary>
    /// 固定产出
    /// </summary>
    [Key(9)]
    public UInt32[][] fixedGain;

    /// <summary>
    /// 开启条件描述
    /// </summary>
    [Key(10)]
    public string openConditionDesc;

    /// <summary>
    /// 材料过滤选择器
    /// </summary>
    [Key(11)]
    public UInt32[][] materialSelector;

    /// <summary>
    /// 消耗组
    /// </summary>
    [Key(12)]
    public UInt32[][] costGroup;

    /// <summary>
    /// 消耗
    /// </summary>
    [Key(13)]
    public UInt32[][] cost;

    /// <summary>
    /// 消耗描述
    /// </summary>
    [Key(14)]
    public string costDesc;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
