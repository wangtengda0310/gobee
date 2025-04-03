using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class PetLevelClassData : IDataBase
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
    /// 最高等级
    /// </summary>
    [Key(2)]
    public UInt32 maxLevel;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
