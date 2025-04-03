using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class SkillBarData : IDataBase
{
    /// <summary>
    /// 技能栏id
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 技能栏名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 技能栏类型
    /// </summary>
    [Key(2)]
    public UInt32 type;

    /// <summary>
    /// 释放次数
    /// </summary>
    [Key(3)]
    public int castNum;

    /// <summary>
    /// 消耗次数概率
    /// </summary>
    [Key(4)]
    public UInt32 castNumProb;

    /// <summary>
    /// 技能栏CD
    /// </summary>
    [Key(5)]
    public UInt32 barCD;

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
    /// 是否隐藏
    /// </summary>
    [Key(8)]
    public byte isHide;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
