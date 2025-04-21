using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class GrowthData : IDataBase
{
    /// <summary>
    /// 养成线id
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 养成线名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 养成线类型
    /// </summary>
    [Key(2)]
    public UInt16 type;

    /// <summary>
    /// 类别id
    /// </summary>
    [Key(3)]
    public UInt32 categoryId;

    /// <summary>
    /// 所属对象
    /// </summary>
    [Key(4)]
    public UInt32 owner;

    /// <summary>
    /// 道具消耗参数(itemId|trendTypeId|trendParam)
    /// </summary>
    [Key(5)]
    public UInt32[][] costParam;

    /// <summary>
    /// 属性加成参数(attrId|attrDomainId|trendTypeId|trendParam)
    /// </summary>
    [Key(6)]
    public string attrParam;

    /// <summary>
    /// 开启条件
    /// </summary>
    [Key(7)]
    public string openCondition;

    /// <summary>
    /// 开启条件描述
    /// </summary>
    [Key(8)]
    public string openConditionDesc;

    /// <summary>
    /// 生效条件
    /// </summary>
    [Key(9)]
    public string effectiveCondition;

    /// <summary>
    /// 生效条件描述
    /// </summary>
    [Key(10)]
    public string effectiveConditionDesc;

    /// <summary>
    /// 超时自动升级
    /// </summary>
    [Key(11)]
    public byte timeoutAutoUpgrade;

    /// <summary>
    /// 最小等级
    /// </summary>
    [Key(12)]
    public UInt32 minLevel;

    /// <summary>
    /// 最大等级
    /// </summary>
    [Key(13)]
    public UInt32 maxLevel;

    /// <summary>
    /// 单次升级上限
    /// </summary>
    [Key(14)]
    public UInt32 perUpgradeMax;

    /// <summary>
    /// 基础经验属性
    /// </summary>
    [Key(15)]
    public UInt32 basicExpAttr;

    /// <summary>
    /// 吞噬经验属性
    /// </summary>
    [Key(16)]
    public UInt32 devourExpAttr;

    /// <summary>
    /// 局内升级消耗物品
    /// </summary>
    [Key(17)]
    public UInt32 costItem;

    /// <summary>
    /// 标签
    /// </summary>
    [Key(18)]
    public List<UInt32> funcLabel;

    /// <summary>
    /// 标签类
    /// </summary>
    [Key(19)]
    public List<UInt32> funcLabelClass;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
