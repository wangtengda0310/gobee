using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class AttributeData : IDataBase
{
    /// <summary>
    /// 属性id
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 属性名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 显示名
    /// </summary>
    [Key(2)]
    public string displayName;

    /// <summary>
    /// 属性类型
    /// </summary>
    [Key(3)]
    public UInt32 type;

    /// <summary>
    /// 复合属性类型
    /// </summary>
    [Key(4)]
    public byte hierarchy;

    /// <summary>
    /// 数值类型（整数、百分比、万分比）
    /// </summary>
    [Key(5)]
    public UInt32 valueType;

    /// <summary>
    /// 显示类型
    /// </summary>
    [Key(6)]
    public UInt32 displayType;

    /// <summary>
    /// 属性源
    /// </summary>
    [Key(7)]
    public List<UInt32> source;

    /// <summary>
    /// 战力价值单价
    /// </summary>
    [Key(8)]
    public UInt32 score;

    /// <summary>
    /// 对象类型
    /// </summary>
    [Key(9)]
    public List<UInt32> objType;

    /// <summary>
    /// 标签
    /// </summary>
    [Key(10)]
    public UInt32 tag;

    /// <summary>
    /// 最大值
    /// </summary>
    [Key(11)]
    public int max;

    /// <summary>
    /// 最小值
    /// </summary>
    [Key(12)]
    public int min;

    /// <summary>
    /// 是否客户端面板显示
    /// </summary>
    [Key(13)]
    public UInt32 isClientProp;

    /// <summary>
    /// 功能枚举
    /// </summary>
    [Key(14)]
    public string funcLabel;

    /// <summary>
    /// 自定义标签
    /// </summary>
    [Key(15)]
    public List<UInt32> label;

    /// <summary>
    /// 自定义标签类
    /// </summary>
    [Key(16)]
    public List<UInt32> labelClass;

    /// <summary>
    /// 词缀类型
    /// </summary>
    [Key(17)]
    public UInt32 affixType;

    /// <summary>
    /// 词缀参数
    /// </summary>
    [Key(18)]
    public string affixParam;

    /// <summary>
    /// 属性枚举类
    /// </summary>
    [Key(19)]
    public UInt32 attrEnumClassId;

    /// <summary>
    /// 属性描述
    /// </summary>
    [Key(20)]
    public string desc;

    /// <summary>
    /// icon
    /// </summary>
    [Key(21)]
    public string icon;

    /// <summary>
    /// 初始是否生效
    /// </summary>
    [Key(22)]
    public UInt32 initialActive;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
