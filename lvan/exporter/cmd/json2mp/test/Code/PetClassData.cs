using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class PetClassData : IDataBase
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
    /// 养成线
    /// </summary>
    [Key(2)]
    public List<UInt32> growth;

    /// <summary>
    /// 宠物的背包
    /// </summary>
    [Key(3)]
    public List<UInt32> bag;

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

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
