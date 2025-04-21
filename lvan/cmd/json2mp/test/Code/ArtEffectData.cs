using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class ArtEffectData : IDataBase
{
    /// <summary>
    /// 美术预制体id
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 描述
    /// </summary>
    [Key(2)]
    public string desc;

    /// <summary>
    /// 特效资源ID
    /// </summary>
    [Key(3)]
    public UInt32 resId;

    /// <summary>
    /// 层级
    /// </summary>
    [Key(4)]
    public byte frame;

    /// <summary>
    /// 位置
    /// </summary>
    [Key(5)]
    public byte pos;

    /// <summary>
    /// 循环模式
    /// </summary>
    [Key(6)]
    public byte loop;

    /// <summary>
    /// 偏移量
    /// </summary>
    [Key(7)]
    public List<int> offset;

    /// <summary>
    /// 缩放值
    /// </summary>
    [Key(8)]
    public List<long> size;

    /// <summary>
    /// 角度
    /// </summary>
    [Key(9)]
    public UInt16 angle;

    /// <summary>
    /// 播放模式
    /// </summary>
    [Key(10)]
    public string mode;

    /// <summary>
    /// 播放速率
    /// </summary>
    [Key(11)]
    public UInt32 interval;

    /// <summary>
    /// 音效
    /// </summary>
    [Key(12)]
    public UInt32 sound;

    /// <summary>
    /// 动作名
    /// </summary>
    [Key(13)]
    public string action;

    /// <summary>
    /// 延迟出现时间
    /// </summary>
    [Key(14)]
    public UInt16 delayTime;

    /// <summary>
    /// 是否适配全屏进行缩放
    /// </summary>
    [Key(15)]
    public UInt16 isZoom;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
