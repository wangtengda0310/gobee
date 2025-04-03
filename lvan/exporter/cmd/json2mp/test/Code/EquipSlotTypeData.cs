using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class EquipSlotTypeData : IDataBase
{
    /// <summary>
    /// 装备类别id
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 类别名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 描述
    /// </summary>
    [Key(2)]
    public string desc;

    /// <summary>
    /// 图标
    /// </summary>
    [Key(3)]
    public string icon;

    /// <summary>
    /// 排序
    /// </summary>
    [Key(4)]
    public UInt32 order;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
