using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class EffectData : IDataBase
{
    /// <summary>
    /// 效果ID
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 效果名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 效果模板类型
    /// </summary>
    [Key(2)]
    public UInt32 effectType;

    /// <summary>
    /// 效果模板参数
    /// </summary>
    [Key(3)]
    public string effectParam;

    /// <summary>
    /// 触发类型
    /// </summary>
    [Key(4)]
    public UInt16 triggerType;

    /// <summary>
    /// 触发类型参数
    /// </summary>
    [Key(5)]
    public UInt32[][] triggerParam;

    /// <summary>
    /// 触发概率
    /// </summary>
    [Key(6)]
    public UInt16 triggerRate;

    /// <summary>
    /// 触发条件
    /// </summary>
    [Key(7)]
    public int[][] triggerCondition;

    /// <summary>
    /// 冷却信息
    /// </summary>
    [Key(8)]
    public List<UInt32> coolDown;

    /// <summary>
    /// 消耗组
    /// </summary>
    [Key(9)]
    public UInt32[][] costGroup;

    /// <summary>
    /// 消耗
    /// </summary>
    [Key(10)]
    public UInt32[][] cost;

    /// <summary>
    /// 目标拾取类型
    /// </summary>
    [Key(11)]
    public byte pickType;

    /// <summary>
    /// 范围拾取锚点
    /// </summary>
    [Key(12)]
    public byte pickAnchor;

    /// <summary>
    /// 范围拾取形状
    /// </summary>
    [Key(13)]
    public List<int> pickArea;

    /// <summary>
    /// 范围拾取形状缩放
    /// </summary>
    [Key(14)]
    public UInt32 pickAreaScale;

    /// <summary>
    /// 范围拾取偏移
    /// </summary>
    [Key(15)]
    public List<int> pickOffset;

    /// <summary>
    /// 范围拾取过滤规则
    /// </summary>
    [Key(16)]
    public int[][] pickFilter;

    /// <summary>
    /// 范围拾取后排序规则
    /// </summary>
    [Key(17)]
    public int[][] pickSort;

    /// <summary>
    /// 范围拾取目标剔除类型
    /// </summary>
    [Key(18)]
    public List<byte> pickExclude;

    /// <summary>
    /// 范围拾取数量上限
    /// </summary>
    [Key(19)]
    public UInt16 pickCount;

    /// <summary>
    /// 范围拾取数量下限
    /// </summary>
    [Key(20)]
    public UInt16 pickMinCount;

    /// <summary>
    /// 目标拾取参数
    /// </summary>
    [Key(21)]
    public List<int> pickParam;

    /// <summary>
    /// 表现
    /// </summary>
    [Key(22)]
    public string art;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
