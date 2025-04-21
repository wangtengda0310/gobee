using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class NavigateItemData : IDataBase
{
    /// <summary>
    /// 导航项ID
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 导航项名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 所属导航组
    /// </summary>
    [Key(2)]
    public UInt32 navigateGroup;

    /// <summary>
    /// 显示名
    /// </summary>
    [Key(3)]
    public string displayName;

    /// <summary>
    /// 名称图标
    /// </summary>
    [Key(4)]
    public string nameIcon;

    /// <summary>
    /// 排序
    /// </summary>
    [Key(5)]
    public UInt32 order;

    /// <summary>
    /// 目标类型
    /// </summary>
    [Key(6)]
    public UInt32 targetType;

    /// <summary>
    /// 目标
    /// </summary>
    [Key(7)]
    public UInt32 target;

    /// <summary>
    /// 占位比例（万分比）
    /// </summary>
    [Key(8)]
    public UInt32 gridRatio;

    /// <summary>
    /// UI功能
    /// </summary>
    [Key(9)]
    public List<UInt32> funcNode;

    /// <summary>
    /// 红点节点
    /// </summary>
    [Key(10)]
    public List<UInt32> redDotNode;

    /// <summary>
    /// 展示模式
    /// </summary>
    [Key(11)]
    public UInt32[][] displayMode;

    /// <summary>
    /// 组件样式
    /// </summary>
    [Key(12)]
    public string componentStyle;

    /// <summary>
    /// UI动效表现
    /// </summary>
    [Key(13)]
    public UInt32 uiArtEffect;

    /// <summary>
    /// Icon样式
    /// </summary>
    [Key(14)]
    public UInt32 iconStyle;

    /// <summary>
    /// 默认图标
    /// </summary>
    [Key(15)]
    public List<UInt32> defaultIcon;

    /// <summary>
    /// 选中图标
    /// </summary>
    [Key(16)]
    public List<UInt32> selectedIcon;

    /// <summary>
    /// 禁用Icon
    /// </summary>
    [Key(17)]
    public List<UInt32> disabledIcon;

    /// <summary>
    /// 描述
    /// </summary>
    [Key(18)]
    public string desc;

    /// <summary>
    /// 位置偏移
    /// </summary>
    [Key(19)]
    public string posOffset;

    /// <summary>
    /// 扩展参数
    /// </summary>
    [Key(20)]
    public string paramsAlt;

    /// <summary>
    /// 解锁模式
    /// </summary>
    [Key(21)]
    public UInt32 unlockMode;

    /// <summary>
    /// 解锁显示名提示语
    /// </summary>
    [Key(22)]
    public UInt32 unlockName;

    /// <summary>
    /// 解锁图标
    /// </summary>
    [Key(23)]
    public string unlockIcon;

    /// <summary>
    /// 是否有动效
    /// </summary>
    [Key(24)]
    public UInt32 isEffect;

    /// <summary>
    /// 解锁描述提示语
    /// </summary>
    [Key(25)]
    public UInt32 unlockDesc;

    /// <summary>
    /// 未开启提示语
    /// </summary>
    [Key(26)]
    public UInt32 lockMessage;

    /// <summary>
    /// 未开启点击提示语
    /// </summary>
    [Key(27)]
    public UInt32 lockClickMessage;

    /// <summary>
    /// 组件样式参数
    /// </summary>
    [Key(28)]
    public string componentStyleParam;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
