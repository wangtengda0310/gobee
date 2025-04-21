using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class PlotEventFlowData : IDataBase
{
    /// <summary>
    /// 事件流id
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
    /// 触发类型
    /// </summary>
    [Key(3)]
    public byte triggerType;

    /// <summary>
    /// 触发参数
    /// </summary>
    [Key(4)]
    public long[][] triggerParam;

    /// <summary>
    /// 事件
    /// </summary>
    [Key(5)]
    public UInt32 eventAlt;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
