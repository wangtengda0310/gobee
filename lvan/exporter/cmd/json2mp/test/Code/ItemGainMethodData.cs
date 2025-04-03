using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class ItemGainMethodData : IDataBase
{
    /// <summary>
    /// id
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 显示名提示语
    /// </summary>
    [Key(2)]
    public string displayName;

    /// <summary>
    /// 标签
    /// </summary>
    [Key(3)]
    public List<UInt32> label;

    /// <summary>
    /// 标签类
    /// </summary>
    [Key(4)]
    public List<UInt32> labelClass;

    /// <summary>
    /// 图标
    /// </summary>
    [Key(5)]
    public string icon;

    /// <summary>
    /// 导航项
    /// </summary>
    [Key(6)]
    public UInt32[][] navItem;

    /// <summary>
    /// 扩展参数
    /// </summary>
    [Key(7)]
    public UInt32[][] paramsAlt;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
