using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class ComposeClassData : IDataBase
{
    /// <summary>
    /// 合成系统id
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 合成系统类型
    /// </summary>
    [Key(2)]
    public UInt16 type;

    /// <summary>
    /// 开启条件
    /// </summary>
    [Key(3)]
    public string openCondition;

    /// <summary>
    /// 开启条件描述
    /// </summary>
    [Key(4)]
    public string openConditionDesc;

    /// <summary>
    /// 合成方式(1:指定;2:批量)
    /// </summary>
    [Key(5)]
    public UInt32 composeMode;

    /// <summary>
    /// 累积消耗记录
    /// </summary>
    [Key(6)]
    public byte cumCostRecord;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
