using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class EquipItemData : IDataBase
{
    /// <summary>
    /// 装备id
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 装备名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 装备类别
    /// </summary>
    [Key(2)]
    public UInt32 slotType;

    /// <summary>
    /// 装备类别名称
    /// </summary>
    [Key(3)]
    public string slotTypeName;

    /// <summary>
    /// 层次等级
    /// </summary>
    [Key(4)]
    public List<UInt32> displayLevel;

    /// <summary>
    /// 等级名称
    /// </summary>
    [Key(5)]
    public string levelName;

    /// <summary>
    /// 穿戴条件
    /// </summary>
    [Key(6)]
    public string useCondition;

    /// <summary>
    /// 固定属性加成
    /// </summary>
    [Key(7)]
    public int[][] fixedAttr;

    /// <summary>
    /// 随机属性
    /// </summary>
    [Key(8)]
    public UInt32[][] randomAttr;

    /// <summary>
    /// 装备升级养成线
    /// </summary>
    [Key(9)]
    public UInt32[][] equipUpgradeGrowth;

    /// <summary>
    /// 洗炼养成线
    /// </summary>
    [Key(10)]
    public UInt32 refineGrowth;

    /// <summary>
    /// 技能
    /// </summary>
    [Key(11)]
    public UInt32[][] skill;

    /// <summary>
    /// 品级
    /// </summary>
    [Key(12)]
    public List<UInt32> level;

    /// <summary>
    /// 装备品级类
    /// </summary>
    [Key(13)]
    public string equipLevelClass;

    /// <summary>
    /// 吞噬养成线
    /// </summary>
    [Key(14)]
    public UInt32[][] devourGrowth;

    /// <summary>
    /// 吞噬养成线
    /// </summary>
    [Key(15)]
    public UInt32[][] devourGrowthClient;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
