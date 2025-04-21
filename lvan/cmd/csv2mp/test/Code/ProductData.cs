using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class ProductData : IDataBase
{
    /// <summary>
    /// 商品id
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 商品名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 商品图标
    /// </summary>
    [Key(2)]
    public string productIcon;

    /// <summary>
    /// 商品标签
    /// </summary>
    [Key(3)]
    public List<UInt32> label;

    /// <summary>
    /// 商品标签类
    /// </summary>
    [Key(4)]
    public List<UInt32> labelClass;

    /// <summary>
    /// 内含物品
    /// </summary>
    [Key(5)]
    public UInt32[][] itemConfiguration;

    /// <summary>
    /// 内含物品掉落包
    /// </summary>
    [Key(6)]
    public UInt32[][] itemDropBox;

    /// <summary>
    /// 展示物品配置
    /// </summary>
    [Key(7)]
    public UInt32[][] displayItemConfiguration;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
