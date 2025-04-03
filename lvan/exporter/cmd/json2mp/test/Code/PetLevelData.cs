using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class PetLevelData : IDataBase
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
    /// 等级类
    /// </summary>
    [Key(2)]
    public UInt32 classAlt;

    /// <summary>
    /// 等级
    /// </summary>
    [Key(3)]
    public UInt32 level;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
