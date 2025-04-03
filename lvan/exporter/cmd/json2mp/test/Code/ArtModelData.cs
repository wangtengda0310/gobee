using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class ArtModelData : IDataBase
{
    /// <summary>
    /// 模型ID
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
    public string des;

    /// <summary>
    /// 资源ID
    /// </summary>
    [Key(3)]
    public UInt32 resId;

    /// <summary>
    /// 模型参数配置
    /// </summary>
    [Key(4)]
    public string modelConfig;

    /// <summary>
    /// 缩放值
    /// </summary>
    [Key(5)]
    public List<int> size;

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
    /// 角度
    /// </summary>
    [Key(8)]
    public UInt16 angle;

    /// <summary>
    /// 播放模式
    /// </summary>
    [Key(9)]
    public string mode;

    /// <summary>
    /// 播放速率
    /// </summary>
    [Key(10)]
    public UInt32 interval;

    /// <summary>
    /// 动作名
    /// </summary>
    [Key(11)]
    public string action;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
