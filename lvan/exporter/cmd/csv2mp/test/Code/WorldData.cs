using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class WorldData : IDataBase
{
    /// <summary>
    /// 世界id
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 世界名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 世界类型
    /// </summary>
    [Key(2)]
    public byte type;

    /// <summary>
    /// 类型参数
    /// </summary>
    [Key(3)]
    public List<UInt32> typeParam;

    /// <summary>
    /// 开启条件
    /// </summary>
    [Key(4)]
    public string openCondition;

    /// <summary>
    /// 开启条件描述
    /// </summary>
    [Key(5)]
    public string openConditionDesc;

    /// <summary>
    /// 开启消耗
    /// </summary>
    [Key(6)]
    public List<KeyValuePair<UInt32, UInt32>> openCost;

    /// <summary>
    /// 开启消耗描述
    /// </summary>
    [Key(7)]
    public string openCostDesc;

    /// <summary>
    /// 进入条件
    /// </summary>
    [Key(8)]
    public string entryCondition;

    /// <summary>
    /// 进入条件描述
    /// </summary>
    [Key(9)]
    public string entryConditionDesc;

    /// <summary>
    /// 进入消耗
    /// </summary>
    [Key(10)]
    public List<KeyValuePair<UInt32, UInt32>> entryCost;

    /// <summary>
    /// 进入消耗描述
    /// </summary>
    [Key(11)]
    public string entryCostDesc;

    /// <summary>
    /// 关联场景
    /// </summary>
    [Key(12)]
    public List<UInt32> scene;

    /// <summary>
    /// 自动进入场景
    /// </summary>
    [Key(13)]
    public byte autoEnterScene;

    /// <summary>
    /// 是否掉线重连
    /// </summary>
    [Key(14)]
    public byte isReEnter;

    /// <summary>
    /// 组队规则
    /// </summary>
    [Key(15)]
    public UInt32 teamRule;

    /// <summary>
    /// 玩家AI
    /// </summary>
    [Key(16)]
    public UInt32 playerAI;

    /// <summary>
    /// 跨服类别
    /// </summary>
    [Key(17)]
    public string crossServerCategory;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
