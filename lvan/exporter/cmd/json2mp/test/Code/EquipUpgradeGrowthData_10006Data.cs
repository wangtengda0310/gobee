using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class EquipUpgradeGrowthData_10006Data : IDataBase
{
    /// <summary>
    /// 编号
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 所属装备升级养成线
    /// </summary>
    [Key(2)]
    public UInt32 growthId;

    /// <summary>
    /// 等级
    /// </summary>
    [Key(3)]
    public int level;

    /// <summary>
    /// 下一等级
    /// </summary>
    [Key(4)]
    public int nextLevel;

    /// <summary>
    /// 等级名称
    /// </summary>
    [Key(5)]
    public string levelName;

    /// <summary>
    /// 消耗
    /// </summary>
    [Key(6)]
    public UInt32[][] cost;

    /// <summary>
    /// 属性加成
    /// </summary>
    [Key(7)]
    public string attr;

    /// <summary>
    /// 累计属性加成
    /// </summary>
    [Key(8)]
    public string totalAttr;

    /// <summary>
    /// 道具消耗描述
    /// </summary>
    [Key(9)]
    public string costDesc;

    /// <summary>
    /// 属性加成描述
    /// </summary>
    [Key(10)]
    public string attrDesc;

    /// <summary>
    /// 消耗返还
    /// </summary>
    [Key(11)]
    public UInt32[][] costReturn;

    /// <summary>
    /// 升级条件
    /// </summary>
    [Key(12)]
    public string upgradeCondition;

    /// <summary>
    /// 升级条件描述
    /// </summary>
    [Key(13)]
    public string upgradeConditionDesc;

    /// <summary>
    /// 一键升级是否停顿
    /// </summary>
    [Key(14)]
    public byte instantUpgradePause;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
