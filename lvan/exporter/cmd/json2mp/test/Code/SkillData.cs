using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class SkillData : IDataBase
{
    /// <summary>
    /// 技能id
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 技能名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 技能类型
    /// </summary>
    [Key(2)]
    public byte type;

    /// <summary>
    /// 所属对象
    /// </summary>
    [Key(3)]
    public byte owner;

    /// <summary>
    /// 技能初始等级
    /// </summary>
    [Key(4)]
    public UInt32 initLevel;

    /// <summary>
    /// 技能最大等级
    /// </summary>
    [Key(5)]
    public UInt32 maxLevel;

    /// <summary>
    /// 开启条件
    /// </summary>
    [Key(6)]
    public string openCondition;

    /// <summary>
    /// 开启条件描述
    /// </summary>
    [Key(7)]
    public string openConditionDesc;

    /// <summary>
    /// 生效条件
    /// </summary>
    [Key(8)]
    public string effectiveCondition;

    /// <summary>
    /// 生效条件描述
    /// </summary>
    [Key(9)]
    public string effectiveConditionDesc;

    /// <summary>
    /// 技能类别
    /// </summary>
    [Key(10)]
    public List<UInt32> category;

    /// <summary>
    /// 标签
    /// </summary>
    [Key(11)]
    public List<UInt32> label;

    /// <summary>
    /// 标签类
    /// </summary>
    [Key(12)]
    public List<UInt32> labelClass;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
