using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class MissionData : IDataBase
{
    /// <summary>
    /// 任务ID
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 任务名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 所属任务系统ID
    /// </summary>
    [Key(2)]
    public UInt32 classAlt;

    /// <summary>
    /// 所属任务链
    /// </summary>
    [Key(3)]
    public UInt32 chain;

    /// <summary>
    /// 前置任务
    /// </summary>
    [Key(4)]
    public List<UInt32> preMission;

    /// <summary>
    /// 后置任务
    /// </summary>
    [Key(5)]
    public List<UInt32> postMission;

    /// <summary>
    /// 激活条件
    /// </summary>
    [Key(6)]
    public string activeCondition;

    /// <summary>
    /// 接受条件
    /// </summary>
    [Key(7)]
    public string acceptCondition;

    /// <summary>
    /// 已接受状态点击跳转导航项
    /// </summary>
    [Key(8)]
    public UInt32[][] acceptClick;

    /// <summary>
    /// 完成条件
    /// </summary>
    [Key(9)]
    public string completeCondition;

    /// <summary>
    /// 提交条件
    /// </summary>
    [Key(10)]
    public string submitCondition;

    /// <summary>
    /// 完成目标进度值
    /// </summary>
    [Key(11)]
    public UInt32 targetProgress;

    /// <summary>
    /// 有效期类型
    /// </summary>
    [Key(12)]
    public UInt32 expireType;

    /// <summary>
    /// 有效期参数
    /// </summary>
    [Key(13)]
    public string expireParam;

    /// <summary>
    /// 完成目标进度值显示类型
    /// </summary>
    [Key(14)]
    public UInt32 targetProgressShow;

    /// <summary>
    /// 触发动作
    /// </summary>
    [Key(15)]
    public List<UInt32> trigger;

    /// <summary>
    /// 任务标题
    /// </summary>
    [Key(16)]
    public string title;

    /// <summary>
    /// 任务奖励
    /// </summary>
    [Key(17)]
    public UInt32[][] reward;

    /// <summary>
    /// 奖励发放模式
    /// </summary>
    [Key(18)]
    public byte rewardDeliveryMode;

    /// <summary>
    /// 奖励发放模式参数
    /// </summary>
    [Key(19)]
    public List<long> rewardDeliveryModeParam;

    /// <summary>
    /// 显示奖励
    /// </summary>
    [Key(20)]
    public UInt32[][] rewardShow;

    /// <summary>
    /// 显示奖励模式
    /// </summary>
    [Key(21)]
    public UInt32 rewardShowType;

    /// <summary>
    /// 可接受状态任务描述
    /// </summary>
    [Key(22)]
    public string activeDesc;

    /// <summary>
    /// 已接受状态任务描述
    /// </summary>
    [Key(23)]
    public string acceptDesc;

    /// <summary>
    /// 完成状态任务描述
    /// </summary>
    [Key(24)]
    public string completeDesc;

    /// <summary>
    /// 任务图标
    /// </summary>
    [Key(25)]
    public string icon;

    /// <summary>
    /// 已完成图标
    /// </summary>
    [Key(26)]
    public string completeIcon;

    /// <summary>
    /// 已提交图标
    /// </summary>
    [Key(27)]
    public string submitIcon;

    /// <summary>
    /// 任务排序
    /// </summary>
    [Key(28)]
    public UInt32 order;

    /// <summary>
    /// 标签
    /// </summary>
    [Key(29)]
    public List<UInt32> label;

    /// <summary>
    /// 标签类
    /// </summary>
    [Key(30)]
    public List<UInt32> labelClass;

    /// <summary>
    /// 展示类型
    /// </summary>
    [Key(31)]
    public UInt32 displayType;

    /// <summary>
    /// 展示类型参数
    /// </summary>
    [Key(32)]
    public string displayTypeParam;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
