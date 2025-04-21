using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class GrowthCategoryData : IDataBase
{
    /// <summary>
    /// 类别id
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 类别名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 图标
    /// </summary>
    [Key(2)]
    public string icon;

    /// <summary>
    /// 描述
    /// </summary>
    [Key(3)]
    public string desc;

    /// <summary>
    /// 父类别id
    /// </summary>
    [Key(4)]
    public UInt32 parentCategoryId;

    /// <summary>
    /// 类型
    /// </summary>
    [Key(5)]
    public UInt32 type;

    /// <summary>
    /// 标签
    /// </summary>
    [Key(6)]
    public string funcLabel;

    /// <summary>
    /// 标签类
    /// </summary>
    [Key(7)]
    public string funcLabelClass;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
