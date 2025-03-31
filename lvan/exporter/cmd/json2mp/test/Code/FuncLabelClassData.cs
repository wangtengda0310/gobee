using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class FuncLabelClassData : IDataBase
{
    /// <summary>
    /// 功能标签类id
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 所属模块
    /// </summary>
    [Key(1)]
    public UInt32 module;

    /// <summary>
    /// 功能标签类名称
    /// </summary>
    [Key(2)]
    public string name;

    /// <summary>
    /// 描述
    /// </summary>
    [Key(3)]
    public string desc;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
