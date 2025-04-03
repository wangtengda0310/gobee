using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class EquipLevelClassData : IDataBase
{
    /// <summary>
    /// 品级类id
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

    /// <summary>
    /// UI样式类型
    /// </summary>
    [Key(3)]
    public UInt32 style;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
