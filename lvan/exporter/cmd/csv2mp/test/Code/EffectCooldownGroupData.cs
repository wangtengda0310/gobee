using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class EffectCooldownGroupData : IDataBase
{
    /// <summary>
    /// 冷却组id
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 冷却时间
    /// </summary>
    [Key(1)]
    public UInt32 time;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
