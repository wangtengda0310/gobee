using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class ItemData : IDataBase
{
    /// <summary>
    /// 物品id
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 道具名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 显示名称
    /// </summary>
    [Key(2)]
    public string displayName;

    /// <summary>
    /// 道具类型
    /// </summary>
    [Key(3)]
    public byte type;

    /// <summary>
    /// 子类型（暂时没用到）
    /// </summary>
    [Key(4)]
    public UInt32 subType;

    /// <summary>
    /// 道具描述
    /// </summary>
    [Key(5)]
    public string description;

    /// <summary>
    /// 道具等级
    /// </summary>
    [Key(6)]
    public string level;

    /// <summary>
    /// 道具品质
    /// </summary>
    [Key(7)]
    public byte quality;

    /// <summary>
    /// 道具图标
    /// </summary>
    [Key(8)]
    public string icon;

    /// <summary>
    /// 模型（美术预制体）
    /// </summary>
    [Key(9)]
    public UInt32 model;

    /// <summary>
    /// 模型参数
    /// </summary>
    [Key(10)]
    public string modelConfig;

    /// <summary>
    /// 立绘
    /// </summary>
    [Key(11)]
    public string portrait;

    /// <summary>
    /// 立绘参数
    /// </summary>
    [Key(12)]
    public string portraitConfig;

    /// <summary>
    /// 道具特效
    /// </summary>
    [Key(13)]
    public string effect;

    /// <summary>
    /// 使用条件
    /// </summary>
    [Key(14)]
    public string useCondition;

    /// <summary>
    /// 使用效果
    /// </summary>
    [Key(15)]
    public int[][] useEffect;

    /// <summary>
    /// 获取途径
    /// </summary>
    [Key(16)]
    public List<UInt32> access;

    /// <summary>
    /// 道具期限
    /// </summary>
    [Key(17)]
    public List<byte> timeLimit;

    /// <summary>
    /// 是否堆叠
    /// </summary>
    [Key(18)]
    public byte isStack;

    /// <summary>
    /// 最大堆叠数
    /// </summary>
    [Key(19)]
    public UInt32 stackNum;

    /// <summary>
    /// 所属背包
    /// </summary>
    [Key(20)]
    public UInt32 bag;

    /// <summary>
    /// 镶嵌类型
    /// </summary>
    [Key(21)]
    public List<byte> jewelType;

    /// <summary>
    /// 是否绑定
    /// </summary>
    [Key(22)]
    public byte isBinding;

    /// <summary>
    /// 自动使用
    /// </summary>
    [Key(23)]
    public byte autoUse;

    /// <summary>
    /// 过期时间
    /// </summary>
    [Key(24)]
    public UInt32[][] expireTime;

    /// <summary>
    /// 过期返还道具
    /// </summary>
    [Key(25)]
    public UInt32[][] expireItemReturn;

    /// <summary>
    /// 功能枚举
    /// </summary>
    [Key(26)]
    public string funcLabel;

    /// <summary>
    /// 使用效果类型
    /// </summary>
    [Key(27)]
    public UInt32 useEffectType;

    /// <summary>
    /// 使用效果参数
    /// </summary>
    [Key(28)]
    public string useEffectParam;

    /// <summary>
    /// 单次最大使用个数
    /// </summary>
    [Key(29)]
    public UInt32 maxUseCount;

    /// <summary>
    /// 可用合成配方
    /// </summary>
    [Key(30)]
    public List<UInt32> composeFormula;

    /// <summary>
    /// 标签
    /// </summary>
    [Key(31)]
    public List<UInt32> label;

    /// <summary>
    /// 标签类
    /// </summary>
    [Key(32)]
    public List<UInt32> labelClass;

    /// <summary>
    /// GM限制
    /// </summary>
    [Key(33)]
    public byte gMLimit;

    /// <summary>
    /// 物品品级
    /// </summary>
    [Key(34)]
    public List<UInt32> itemLevel;

    /// <summary>
    /// 物品品级类
    /// </summary>
    [Key(35)]
    public List<UInt32> itemLevelClass;

    /// <summary>
    /// 消耗跳转导航项
    /// </summary>
    [Key(36)]
    public List<UInt32> costUI;

    /// <summary>
    /// 渲染图层
    /// </summary>
    [Key(37)]
    public UInt64 sceneLayer;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
