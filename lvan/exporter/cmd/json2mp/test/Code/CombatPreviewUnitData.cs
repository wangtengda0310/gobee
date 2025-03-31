using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class CombatPreviewUnitData : IDataBase
{
    /// <summary>
    /// 预览单位id
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 所属战斗预览
    /// </summary>
    [Key(2)]
    public UInt32 combatPreview;

    /// <summary>
    /// 战斗站位配置
    /// </summary>
    [Key(3)]
    public UInt32 combatUnitSlot;

    /// <summary>
    /// 位置
    /// </summary>
    [Key(4)]
    public List<UInt32> pos;

    /// <summary>
    /// 对象类型
    /// </summary>
    [Key(5)]
    public byte objType;

    /// <summary>
    /// 对象id
    /// </summary>
    [Key(6)]
    public UInt32 objId;

    /// <summary>
    /// 技能配置
    /// </summary>
    [Key(7)]
    public List<UInt32> skills;

    /// <summary>
    /// 属性配置
    /// </summary>
    [Key(8)]
    public long[][] attributes;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
