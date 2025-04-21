using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class ArtPrefabData : IDataBase
{
    /// <summary>
    /// 预制体id
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
    /// 类型
    /// </summary>
    [Key(3)]
    public byte type;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
