using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class AttrEnumData : IDataBase
{
    /// <summary>
    /// 属性枚举ID
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 属性枚举类ID
    /// </summary>
    [Key(1)]
    public UInt32 attrEnumClassId;

    /// <summary>
    /// 属性枚举名称
    /// </summary>
    [Key(2)]
    public string name;

    /// <summary>
    /// 属性枚举值
    /// </summary>
    [Key(3)]
    public UInt32 value;

    /// <summary>
    /// icon
    /// </summary>
    [Key(4)]
    public string icon;

    /// <summary>
    /// 属性功能标签
    /// </summary>
    [Key(5)]
    public string funcTag;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
