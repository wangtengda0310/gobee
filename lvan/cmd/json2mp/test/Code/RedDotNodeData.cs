using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class RedDotNodeData : IDataBase
{
    /// <summary>
    /// 红点节点id
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 红点节点名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 类型
    /// </summary>
    [Key(2)]
    public UInt32 type;

    /// <summary>
    /// 父节点
    /// </summary>
    [Key(3)]
    public UInt32 fatherId;

    /// <summary>
    /// 关联功能
    /// </summary>
    [Key(4)]
    public List<UInt32> functions;

    /// <summary>
    /// 展示规则
    /// </summary>
    [Key(5)]
    public string showRule;

    /// <summary>
    /// 展示规则参数
    /// </summary>
    [Key(6)]
    public UInt32[][] showRuleParam;

    /// <summary>
    /// 消失规则
    /// </summary>
    [Key(7)]
    public UInt32 disappearRule;

    /// <summary>
    /// 消失规则参数
    /// </summary>
    [Key(8)]
    public UInt32[][] disappearRuleParam;

    /// <summary>
    /// 样式类型
    /// </summary>
    [Key(9)]
    public UInt32 style;

    /// <summary>
    /// 样式参数
    /// </summary>
    [Key(10)]
    public string styleParams;

    /// <summary>
    /// 样式美术资产
    /// </summary>
    [Key(11)]
    public string resource;

    /// <summary>
    /// 位置
    /// </summary>
    [Key(12)]
    public UInt32 location;

    /// <summary>
    /// 父节点跟随优先级
    /// </summary>
    [Key(13)]
    public UInt32 followPriority;

    /// <summary>
    /// 样式响应规则
    /// </summary>
    [Key(14)]
    public UInt32 styleResposeRule;

    /// <summary>
    /// 消失响应规则
    /// </summary>
    [Key(15)]
    public UInt32 disappearResponseRule;

    /// <summary>
    /// 自身与子节点优先级
    /// </summary>
    [Key(16)]
    public UInt32 selfChildPriority;

    /// <summary>
    /// 描述提示语
    /// </summary>
    [Key(17)]
    public UInt32 desc;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
