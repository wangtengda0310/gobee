using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class CombatTypeData : IDataBase
{
    /// <summary>
    /// 战斗行为ID
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 类型名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 伤害显示
    /// </summary>
    [Key(2)]
    public UInt32[][] damageDisplay;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
