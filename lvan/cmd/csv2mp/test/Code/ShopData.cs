using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class ShopData : IDataBase
{
    /// <summary>
    /// 商城ID
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 商城名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 商城类型
    /// </summary>
    [Key(2)]
    public UInt32 shopType;

    /// <summary>
    /// 商城开启条件
    /// </summary>
    [Key(3)]
    public string shopActivationCondition;

    /// <summary>
    /// 商品列表
    /// </summary>
    [Key(4)]
    public List<UInt32> productList;

    /// <summary>
    /// 重置规则
    /// </summary>
    [Key(5)]
    public UInt32 resetRule;

    /// <summary>
    /// 重置规则参数
    /// </summary>
    [Key(6)]
    public string resetRuleParam;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
