using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class GrowthTimeLinearData_10010Data : IDataBase
{
    /// <summary>
    /// 编号
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 养成线ID
    /// </summary>
    [Key(1)]
    public UInt32 growthId;

    /// <summary>
    /// 等级
    /// </summary>
    [Key(2)]
    public int level;

    /// <summary>
    /// 下一等级
    /// </summary>
    [Key(3)]
    public int nextLevel;

    /// <summary>
    /// 下一等级ID
    /// </summary>
    [Key(4)]
    public int nextLevelId;

    /// <summary>
    /// 上一等级ID
    /// </summary>
    [Key(5)]
    public int previousLevelId;

    /// <summary>
    /// 显示等级
    /// </summary>
    [Key(6)]
    public string displayLevel;

    /// <summary>
    /// 等级名称
    /// </summary>
    [Key(7)]
    public string levelName;

    /// <summary>
    /// 图标
    /// </summary>
    [Key(8)]
    public string icon;

    /// <summary>
    /// 开启条件描述
    /// </summary>
    [Key(9)]
    public string openConditionDesc;

    /// <summary>
    /// 升级条件描述
    /// </summary>
    [Key(10)]
    public string upgradeConditionDesc;

    /// <summary>
    /// 升级提示
    /// </summary>
    [Key(11)]
    public string upgradeMessage;

    /// <summary>
    /// 升级时间
    /// </summary>
    [Key(12)]
    public UInt32 upgradeTime;

    /// <summary>
    /// 升级时间描述
    /// </summary>
    [Key(13)]
    public string upgradeTimeDesc;

    /// <summary>
    /// 跳过时间消耗
    /// </summary>
    [Key(14)]
    public UInt32[][] skipTimeCost;

    /// <summary>
    /// 跳过时间消耗描述
    /// </summary>
    [Key(15)]
    public string skipTimeCostDesc;

    /// <summary>
    /// 消耗组
    /// </summary>
    [Key(16)]
    public UInt32[][] costGroup;

    /// <summary>
    /// 消耗
    /// </summary>
    [Key(17)]
    public UInt32[][] cost;

    /// <summary>
    /// 属性加成
    /// </summary>
    [Key(18)]
    public string attr;

    /// <summary>
    /// 累计属性加成
    /// </summary>
    [Key(19)]
    public string totalAttr;

    /// <summary>
    /// 道具消耗描述
    /// </summary>
    [Key(20)]
    public string costDesc;

    /// <summary>
    /// 属性加成描述
    /// </summary>
    [Key(21)]
    public string attrDesc;

    /// <summary>
    /// 一键升级停顿
    /// </summary>
    [Key(22)]
    public byte instantUpgradePause;

    /// <summary>
    /// 标签
    /// </summary>
    [Key(23)]
    public List<UInt32> label;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
