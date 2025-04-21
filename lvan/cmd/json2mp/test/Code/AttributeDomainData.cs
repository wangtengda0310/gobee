using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class AttributeDomainData : IDataBase
{
    /// <summary>
    /// 域id
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 父域
    /// </summary>
    [Key(2)]
    public UInt32 parent;

    /// <summary>
    /// 描述
    /// </summary>
    [Key(3)]
    public string desc;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
