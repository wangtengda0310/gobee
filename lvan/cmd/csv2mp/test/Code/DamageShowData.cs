using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class DamageShowData : IDataBase
{
    /// <summary>
    /// 伤害字体表现id
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 取值区间
    /// </summary>
    [Key(2)]
    public List<int> rectParams;

    /// <summary>
    /// 优先级
    /// </summary>
    [Key(3)]
    public UInt16 priority;

    /// <summary>
    /// 延迟出现时间
    /// </summary>
    [Key(4)]
    public UInt16 delayTime;

    /// <summary>
    /// 符号前缀
    /// </summary>
    [Key(5)]
    public string simbolPrev;

    /// <summary>
    /// 符号后缀
    /// </summary>
    [Key(6)]
    public string simbolSuffix;

    /// <summary>
    /// 资产id
    /// </summary>
    [Key(7)]
    public string resId;

    /// <summary>
    /// 伤害/治疗分段
    /// </summary>
    [Key(8)]
    public byte valueSplite;

    /// <summary>
    /// 分段时间间隔
    /// </summary>
    [Key(9)]
    public UInt32 valueSpliteDelay;

    /// <summary>
    /// 冒字高度调整
    /// </summary>
    [Key(10)]
    public int heightAdjust;

    /// <summary>
    /// 整体透明值
    /// </summary>
    [Key(11)]
    public byte wholeAlpha;

    /// <summary>
    /// 整体缩放值
    /// </summary>
    [Key(12)]
    public UInt16 wholeScale;

    /// <summary>
    /// 飘字动画类型
    /// </summary>
    [Key(13)]
    public byte flyType;

    /// <summary>
    /// 文字形式描述
    /// </summary>
    [Key(14)]
    public string preText;

    /// <summary>
    /// 初始点类型
    /// </summary>
    [Key(15)]
    public byte pivotType;

    /// <summary>
    /// 隐藏模式
    /// </summary>
    [Key(16)]
    public byte hideMode;

    /// <summary>
    /// 单位换算
    /// </summary>
    [Key(17)]
    public string[][] conversion;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
