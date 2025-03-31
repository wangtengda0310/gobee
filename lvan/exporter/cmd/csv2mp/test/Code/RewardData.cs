using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class RewardData : IDataBase
{
    /// <summary>
    /// 奖励id
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 奖励名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 奖励类型
    /// </summary>
    [Key(2)]
    public UInt32 type;

    /// <summary>
    /// 最高阶梯
    /// </summary>
    [Key(3)]
    public UInt32 maxStep;

    /// <summary>
    /// 阶梯类型
    /// </summary>
    [Key(4)]
    public List<UInt32> stepType;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
