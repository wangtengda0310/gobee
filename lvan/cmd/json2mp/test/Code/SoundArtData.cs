using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class SoundArtData : IDataBase
{
    /// <summary>
    /// id
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 是否循环
    /// </summary>
    [Key(2)]
    public byte loop;

    /// <summary>
    /// 音效资产id
    /// </summary>
    [Key(3)]
    public string resId;

    /// <summary>
    /// 延时播放时间（毫秒）
    /// </summary>
    [Key(4)]
    public UInt16 delayTime;

    /// <summary>
    /// 音量
    /// </summary>
    [Key(5)]
    public UInt32 volume;

    /// <summary>
    /// 标签
    /// </summary>
    [Key(6)]
    public List<UInt32> label;

    /// <summary>
    /// 标签类别
    /// </summary>
    [Key(7)]
    public List<UInt32> labelClass;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
