using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class SelectorData : IDataBase
{
    /// <summary>
    /// 选择器ID
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
    /// 数据源
    /// </summary>
    [Key(3)]
    public UInt32 source;

    /// <summary>
    /// 目标对象类型
    /// </summary>
    [Key(4)]
    public UInt32 objectType;

    /// <summary>
    /// 筛选规则
    /// </summary>
    [Key(5)]
    public string filterRule;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
