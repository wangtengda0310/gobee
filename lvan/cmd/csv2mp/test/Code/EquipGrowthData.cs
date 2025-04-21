using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class EquipGrowthData : IDataBase
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
    /// 属性同步模式
    /// </summary>
    [Key(5)]
    public byte propSyncMode;

    /// <summary>
    /// 品级类
    /// </summary>
    [Key(6)]
    public UInt32[][] levelClass;

    /// <summary>
    /// 装备唯一规则
    /// </summary>
    [Key(7)]
    public UInt32 uniqueRule;

    /// <summary>
    /// 开启条件
    /// </summary>
    [Key(8)]
    public string openCondition;

    /// <summary>
    /// 开启条件描述
    /// </summary>
    [Key(9)]
    public string openConditionDesc;

    /// <summary>
    /// 标签
    /// </summary>
    [Key(10)]
    public List<UInt32> funcLabel;

    /// <summary>
    /// 标签类
    /// </summary>
    [Key(11)]
    public List<UInt32> funcLabelClass;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
