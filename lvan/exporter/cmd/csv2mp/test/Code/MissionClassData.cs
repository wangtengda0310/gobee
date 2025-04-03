using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class MissionClassData : IDataBase
{
    /// <summary>
    /// 任务系统ID
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 任务系统名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 任务系统类型
    /// </summary>
    [Key(2)]
    public UInt32 type;

    /// <summary>
    /// 归属
    /// </summary>
    [Key(3)]
    public List<UInt32> owner;

    /// <summary>
    /// 开启条件
    /// </summary>
    [Key(4)]
    public string openCondition;

    /// <summary>
    /// 自动接受点击
    /// </summary>
    [Key(5)]
    public byte autoAcceptClick;

    /// <summary>
    /// 自动完成点击
    /// </summary>
    [Key(6)]
    public byte autoCompleteClick;

    /// <summary>
    /// 自动提交点击
    /// </summary>
    [Key(7)]
    public byte autoSubmitClick;

    /// <summary>
    /// 重置规则
    /// </summary>
    [Key(8)]
    public string resetRule;

    /// <summary>
    /// 自定义标签
    /// </summary>
    [Key(9)]
    public List<UInt32> label;

    /// <summary>
    /// 自定义标签类
    /// </summary>
    [Key(10)]
    public List<UInt32> labelClass;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
