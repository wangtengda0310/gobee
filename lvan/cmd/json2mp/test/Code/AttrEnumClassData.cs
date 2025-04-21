using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class AttrEnumClassData : IDataBase
{
    /// <summary>
    /// 属性枚举类ID
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 属性枚举类名称
    /// </summary>
    [Key(1)]
    public string name;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
