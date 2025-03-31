using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class PlotDialogueSetData : IDataBase
{
    /// <summary>
    /// 对话集id
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
    /// 抽取策略
    /// </summary>
    [Key(3)]
    public byte dropType;

    /// <summary>
    /// 生命周期
    /// </summary>
    [Key(4)]
    public byte lifeTime;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
