using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class CombatPreviewData : IDataBase
{
    /// <summary>
    /// 战斗预览id
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 战斗系统
    /// </summary>
    [Key(2)]
    public UInt32 combatClassId;

    /// <summary>
    /// 场景资源
    /// </summary>
    [Key(3)]
    public string sceneRes;

    /// <summary>
    /// 血条显示
    /// </summary>
    [Key(4)]
    public byte hpBarShow;

    /// <summary>
    /// buff显示
    /// </summary>
    [Key(5)]
    public byte buffIconShow;

    /// <summary>
    /// 回合数显示
    /// </summary>
    [Key(6)]
    public byte roundShow;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
