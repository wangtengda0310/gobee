using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class DamageData : IDataBase
{
    /// <summary>
    /// 积木块节点ID
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 参数
    /// </summary>
    [Key(1)]
    public string[][] param;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
