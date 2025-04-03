using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class EquipUpgradeGrowthData : IDataBase
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
    /// 最大等级
    /// </summary>
    [Key(4)]
    public UInt32 maxLevel;

    /// <summary>
    /// 标签
    /// </summary>
    [Key(5)]
    public List<UInt32> label;

    /// <summary>
    /// 标签类
    /// </summary>
    [Key(6)]
    public List<UInt32> labelClass;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
