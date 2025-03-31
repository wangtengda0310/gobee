using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class PlotEventData : IDataBase
{
    /// <summary>
    /// 事件id
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 类型
    /// </summary>
    [Key(2)]
    public byte type;

    /// <summary>
    /// 描述
    /// </summary>
    [Key(3)]
    public string desc;

    /// <summary>
    /// 触发条件
    /// </summary>
    [Key(4)]
    public long[][] triggerCondition;

    /// <summary>
    /// 事件盒节点类型
    /// </summary>
    [Key(5)]
    public byte eventBoxType;

    /// <summary>
    /// 事件盒节点参数
    /// </summary>
    [Key(6)]
    public UInt32[][] eventBox;

    /// <summary>
    /// 事件类型
    /// </summary>
    [Key(7)]
    public UInt32 eventType;

    /// <summary>
    /// 事件
    /// </summary>
    [Key(8)]
    public string eventParam;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
