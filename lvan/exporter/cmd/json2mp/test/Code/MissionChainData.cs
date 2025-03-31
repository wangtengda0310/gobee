using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class MissionChainData : IDataBase
{
    /// <summary>
    /// 任务链ID
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 任务链名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 所属任务系统ID
    /// </summary>
    [Key(2)]
    public UInt32 classAlt;

    /// <summary>
    /// 任务链顺序
    /// </summary>
    [Key(3)]
    public List<UInt32> missionLine;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
