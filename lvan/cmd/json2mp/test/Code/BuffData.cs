using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class BuffData : IDataBase
{
    /// <summary>
    /// BuffId
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// Buff类
    /// </summary>
    [Key(2)]
    public UInt32 classAlt;

    /// <summary>
    /// 描述
    /// </summary>
    [Key(3)]
    public string desc;

    /// <summary>
    /// 命中效果
    /// </summary>
    [Key(4)]
    public List<UInt32> hitEffect;

    /// <summary>
    /// 持续效果
    /// </summary>
    [Key(5)]
    public List<UInt32> durationEffect;

    /// <summary>
    /// 结束效果
    /// </summary>
    [Key(6)]
    public List<UInt32> endEffect;

    /// <summary>
    /// 中断执行效果
    /// </summary>
    [Key(7)]
    public List<UInt32> discontinueEffect;

    /// <summary>
    /// 持续时间
    /// </summary>
    [Key(8)]
    public UInt32 duration;

    /// <summary>
    /// 免疫净化时间
    /// </summary>
    [Key(9)]
    public UInt32 protectedTime;

    /// <summary>
    /// 消失规则
    /// </summary>
    [Key(10)]
    public List<UInt32> clearRule;

    /// <summary>
    /// 替换规则
    /// </summary>
    [Key(11)]
    public List<UInt32> replaceRule;

    /// <summary>
    /// 叠加规则
    /// </summary>
    [Key(12)]
    public List<UInt32> stackRule;

    /// <summary>
    /// 优先级
    /// </summary>
    [Key(13)]
    public UInt32 priority;

    /// <summary>
    /// 显示排序
    /// </summary>
    [Key(14)]
    public UInt32 displayOrder;

    /// <summary>
    /// 客户端是否显示
    /// </summary>
    [Key(15)]
    public byte clientShow;

    /// <summary>
    /// 图标
    /// </summary>
    [Key(16)]
    public string icon;

    /// <summary>
    /// 命中效果表现
    /// </summary>
    [Key(17)]
    public List<UInt32> hitArtConfig;

    /// <summary>
    /// 持续效果表现
    /// </summary>
    [Key(18)]
    public List<UInt32> continueArtConfig;

    /// <summary>
    /// 结束效果表现
    /// </summary>
    [Key(19)]
    public List<UInt32> endArtConfig;

    /// <summary>
    /// 标签
    /// </summary>
    [Key(20)]
    public List<UInt32> tag;

    /// <summary>
    /// 标签类
    /// </summary>
    [Key(21)]
    public List<UInt32> tagClass;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
