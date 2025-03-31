using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class RewardStepDetailData : IDataBase
{
    /// <summary>
    /// 奖励明细id
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 奖励明细名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 奖励id
    /// </summary>
    [Key(2)]
    public UInt32 rewardId;

    /// <summary>
    /// 起始阶梯
    /// </summary>
    [Key(3)]
    public UInt32 fromStep;

    /// <summary>
    /// 结束阶梯
    /// </summary>
    [Key(4)]
    public UInt32 toStep;

    /// <summary>
    /// 固定奖励
    /// </summary>
    [Key(5)]
    public UInt32[][] fixedReward;

    /// <summary>
    /// 随机奖励
    /// </summary>
    [Key(6)]
    public UInt32[][] randomReward;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
