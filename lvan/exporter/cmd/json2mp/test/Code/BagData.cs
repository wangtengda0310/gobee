using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class BagData : IDataBase
{
    /// <summary>
    /// 背包id
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 类型名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 最大格子数
    /// </summary>
    [Key(2)]
    public int maxCellNum;

    /// <summary>
    /// 是否带格子
    /// </summary>
    [Key(3)]
    public UInt32 isGrid;

    /// <summary>
    /// 默认格子数
    /// </summary>
    [Key(4)]
    public UInt32 bagOpenSize;

    /// <summary>
    /// 功能标签
    /// </summary>
    [Key(5)]
    public string funcLabel;

    /// <summary>
    /// 装备穿戴后是否展示
    /// </summary>
    [Key(6)]
    public UInt32 isEquipProp;

    /// <summary>
    /// 所属对象
    /// </summary>
    [Key(7)]
    public UInt32 ownerType;

    /// <summary>
    /// 排序规则
    /// </summary>
    [Key(8)]
    public UInt32[][] sortRule;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
