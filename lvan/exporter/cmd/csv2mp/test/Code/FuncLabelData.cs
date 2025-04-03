using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class FuncLabelData : IDataBase
{
    /// <summary>
    /// 功能标签id
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 功能标签类
    /// </summary>
    [Key(1)]
    public UInt32 classAlt;

    /// <summary>
    /// 功能标签类名称
    /// </summary>
    [Key(2)]
    public string name;

    /// <summary>
    /// 所属模块
    /// </summary>
    [Key(3)]
    public UInt32 module;

    /// <summary>
    /// 标签
    /// </summary>
    [Key(4)]
    public string label;

    /// <summary>
    /// 描述
    /// </summary>
    [Key(5)]
    public string desc;

    /// <summary>
    /// 图标
    /// </summary>
    [Key(6)]
    public string icon;

    /// <summary>
    /// 排序
    /// </summary>
    [Key(7)]
    public UInt32 sort;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
