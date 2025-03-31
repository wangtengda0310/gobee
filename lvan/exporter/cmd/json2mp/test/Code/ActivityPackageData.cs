using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class ActivityPackageData : IDataBase
{
    /// <summary>
    /// 活动包ID
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 活动包名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 所属活动包组ID
    /// </summary>
    [Key(2)]
    public UInt32 activityPackageGroupId;

    /// <summary>
    /// 包含活动
    /// </summary>
    [Key(3)]
    public List<UInt32> chooseActivityId;

    /// <summary>
    /// 显示名称
    /// </summary>
    [Key(4)]
    public string displayName;

    /// <summary>
    /// 活动包UI
    /// </summary>
    [Key(5)]
    public UInt32 activityPackageUi;

    /// <summary>
    /// 排序
    /// </summary>
    [Key(6)]
    public UInt32 order;

    /// <summary>
    /// 占位比例
    /// </summary>
    [Key(7)]
    public UInt32 gridRatio;

    /// <summary>
    /// 展示模式
    /// </summary>
    [Key(8)]
    public UInt32[][] displayMode;

    /// <summary>
    /// 组件样式
    /// </summary>
    [Key(9)]
    public string componentStyle;

    /// <summary>
    /// UI动效表现
    /// </summary>
    [Key(10)]
    public UInt32 uiArtEffect;

    /// <summary>
    /// ICON样式
    /// </summary>
    [Key(11)]
    public UInt32 iconStyle;

    /// <summary>
    /// UI功能ID
    /// </summary>
    [Key(12)]
    public List<UInt32> funcNode;

    /// <summary>
    /// 红点节点ID
    /// </summary>
    [Key(13)]
    public List<UInt32> redDotNode;

    /// <summary>
    /// 默认ICON
    /// </summary>
    [Key(14)]
    public List<string> defaultIcon;

    /// <summary>
    /// 选中ICON
    /// </summary>
    [Key(15)]
    public List<string> selectedIcon;

    /// <summary>
    /// 禁用ICON
    /// </summary>
    [Key(16)]
    public List<string> disabledIcon;

    /// <summary>
    /// 组件样式参数
    /// </summary>
    [Key(17)]
    public string componentStyleParam;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
