using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class ActivityData : IDataBase
{
    /// <summary>
    /// 活动ID
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 活动名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 所属活动包ID
    /// </summary>
    [Key(2)]
    public UInt32 activityPackage;

    /// <summary>
    /// 是否生效
    /// </summary>
    [Key(3)]
    public UInt32 takeEffect;

    /// <summary>
    /// 活动类型
    /// </summary>
    [Key(4)]
    public UInt32 activityType;

    /// <summary>
    /// 是否由Gm平台控制
    /// </summary>
    [Key(5)]
    public UInt32 gmControl;

    /// <summary>
    /// 能否重复创建活动
    /// </summary>
    [Key(6)]
    public UInt32 repeatedActivityCreation;

    /// <summary>
    /// 是否匹配通道
    /// </summary>
    [Key(7)]
    public UInt32 matchingChannel;

    /// <summary>
    /// 活动起止类型
    /// </summary>
    [Key(8)]
    public UInt32 configureActivityStartAndEnd;

    /// <summary>
    /// 活动开始时间
    /// </summary>
    [Key(9)]
    public string activityStartTime;

    /// <summary>
    /// 活动开始条件
    /// </summary>
    [Key(10)]
    public string activityStartCondition;

    /// <summary>
    /// 活动结束时间
    /// </summary>
    [Key(11)]
    public string activityEndTime;

    /// <summary>
    /// 活动时长（秒数）
    /// </summary>
    [Key(12)]
    public UInt32 activityDuration;

    /// <summary>
    /// 是否循环
    /// </summary>
    [Key(13)]
    public UInt32 isLoop;

    /// <summary>
    /// 循环间隔
    /// </summary>
    [Key(14)]
    public string cycleInterval;

    /// <summary>
    /// 活动循环次数
    /// </summary>
    [Key(15)]
    public UInt32 activityLoopCount;

    /// <summary>
    /// 活动描述
    /// </summary>
    [Key(16)]
    public string activityDescription;

    /// <summary>
    /// 显示名称
    /// </summary>
    [Key(17)]
    public string displayName;

    /// <summary>
    /// 活动页
    /// </summary>
    [Key(18)]
    public UInt32 activityPageId;

    /// <summary>
    /// 排序
    /// </summary>
    [Key(19)]
    public UInt32 order;

    /// <summary>
    /// 占位比例
    /// </summary>
    [Key(20)]
    public UInt32 gridRatio;

    /// <summary>
    /// 展示模式
    /// </summary>
    [Key(21)]
    public UInt32[][] displayMode;

    /// <summary>
    /// 组件样式
    /// </summary>
    [Key(22)]
    public string componentStyle;

    /// <summary>
    /// UI动效表现
    /// </summary>
    [Key(23)]
    public UInt32 uiArtEffect;

    /// <summary>
    /// ICON样式
    /// </summary>
    [Key(24)]
    public UInt32 iconStyle;

    /// <summary>
    /// UI功能ID
    /// </summary>
    [Key(25)]
    public List<UInt32> funcNode;

    /// <summary>
    /// 红点节点ID
    /// </summary>
    [Key(26)]
    public List<UInt32> redDotNode;

    /// <summary>
    /// 默认ICON
    /// </summary>
    [Key(27)]
    public List<string> defaultIcon;

    /// <summary>
    /// 选中ICON
    /// </summary>
    [Key(28)]
    public List<string> selectedIcon;

    /// <summary>
    /// 禁用ICON
    /// </summary>
    [Key(29)]
    public List<string> disabledIcon;

    /// <summary>
    /// 组件样式参数
    /// </summary>
    [Key(30)]
    public string componentStyleParam;

    /// <summary>
    /// 开启显示名
    /// </summary>
    [Key(31)]
    public string openName;

    /// <summary>
    /// 关闭显示名
    /// </summary>
    [Key(32)]
    public string closeName;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
