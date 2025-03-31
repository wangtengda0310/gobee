using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class PetSlotData : IDataBase
{
    /// <summary>
    /// id
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 所属宠物系统
    /// </summary>
    [Key(2)]
    public UInt32 classAlt;

    /// <summary>
    /// 所属宠物槽位组
    /// </summary>
    [Key(3)]
    public UInt32 slotGroup;

    /// <summary>
    /// 可装配宠物类别
    /// </summary>
    [Key(4)]
    public List<UInt32> slotType;

    /// <summary>
    /// 开启消耗
    /// </summary>
    [Key(5)]
    public UInt32[][] openCost;

    /// <summary>
    /// 开启消耗描述
    /// </summary>
    [Key(6)]
    public string openCostDesc;

    /// <summary>
    /// 开启条件
    /// </summary>
    [Key(7)]
    public string openCondition;

    /// <summary>
    /// 开启条件描述
    /// </summary>
    [Key(8)]
    public string openConditionDesc;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
