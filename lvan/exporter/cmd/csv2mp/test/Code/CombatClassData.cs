using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class CombatClassData : IDataBase
{
    /// <summary>
    /// 战斗系统id
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

    /// <summary>
    /// 胜利条件
    /// </summary>
    [Key(4)]
    public long[][] victoryRule;

    /// <summary>
    /// 胜利结算
    /// </summary>
    [Key(5)]
    public long[][] victorySettle;

    /// <summary>
    /// 失败条件
    /// </summary>
    [Key(6)]
    public long[][] defeatedRule;

    /// <summary>
    /// 失败结算
    /// </summary>
    [Key(7)]
    public long[][] defeatedSettle;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
