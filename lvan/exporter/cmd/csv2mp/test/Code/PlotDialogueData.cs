using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class PlotDialogueData : IDataBase
{
    /// <summary>
    /// 对话id
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 描述
    /// </summary>
    [Key(2)]
    public string desc;

    /// <summary>
    /// 所属对话集
    /// </summary>
    [Key(3)]
    public UInt32 plotDialogueSet;

    /// <summary>
    /// 标签
    /// </summary>
    [Key(4)]
    public List<UInt32> label;

    /// <summary>
    /// 匹配规则
    /// </summary>
    [Key(5)]
    public byte matchType;

    /// <summary>
    /// 匹配参数
    /// </summary>
    [Key(6)]
    public List<long> matchParam;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
