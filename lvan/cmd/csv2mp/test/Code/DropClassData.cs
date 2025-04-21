using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class DropClassData : IDataBase
{
    /// <summary>
    /// 掉落系统id
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 掉落系统名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 描述
    /// </summary>
    [Key(2)]
    public string desc;

    /// <summary>
    /// 掉落类（1.物品2.属性）
    /// </summary>
    [Key(3)]
    public UInt32 objType;

    /// <summary>
    /// 计数类型（1.不计数；2.玩家计数；3.场景计数；4.玩法计数；5.全服计数；6.跨服计数）
    /// </summary>
    [Key(4)]
    public UInt32 countType;

    /// <summary>
    /// 奖池类型（1.无奖池；2.玩家奖池；3.场景奖池；4.玩法奖池；5.本服奖池；6.跨服奖池）
    /// </summary>
    [Key(5)]
    public UInt32 poolType;

    /// <summary>
    /// 掉落模型
    /// </summary>
    [Key(6)]
    public UInt32 mode;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
