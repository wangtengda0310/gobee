using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class PlotEventVarData : IDataBase
{
    /// <summary>
    /// 事件变量id
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
    /// 类型
    /// </summary>
    [Key(3)]
    public byte type;

    /// <summary>
    /// 算子类型
    /// </summary>
    [Key(4)]
    public byte operatorAlt;

    /// <summary>
    /// 生命周期
    /// </summary>
    [Key(5)]
    public byte lifeTime;

    /// <summary>
    /// 默认值
    /// </summary>
    [Key(6)]
    public long defaultValue;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
