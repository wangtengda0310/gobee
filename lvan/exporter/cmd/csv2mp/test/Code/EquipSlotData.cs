using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class EquipSlotData : IDataBase
{
    /// <summary>
    /// 槽位id
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 所属装备穿戴养成线id
    /// </summary>
    [Key(1)]
    public UInt32 equipGrowthId;

    /// <summary>
    /// 所属装备槽位组
    /// </summary>
    [Key(2)]
    public UInt32 slotGroup;

    /// <summary>
    /// 槽位名称
    /// </summary>
    [Key(3)]
    public string name;

    /// <summary>
    /// 开启消耗
    /// </summary>
    [Key(4)]
    public UInt32[][] openCost;

    /// <summary>
    /// 开启消耗描述
    /// </summary>
    [Key(5)]
    public string openCostDesc;

    /// <summary>
    /// 开启条件
    /// </summary>
    [Key(6)]
    public string openCondition;

    /// <summary>
    /// 开启条件描述
    /// </summary>
    [Key(7)]
    public string openConditionDesc;

    /// <summary>
    /// 可穿戴装备类别
    /// </summary>
    [Key(8)]
    public List<UInt32> itemSlotType;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
