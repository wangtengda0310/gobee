using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class SkillDisplayData : IDataBase
{
    /// <summary>
    /// 技能动作表现id
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 技能动作表现名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 完整动作时间
    /// </summary>
    [Key(2)]
    public List<UInt16> actWholeTime;

    /// <summary>
    /// 技能前摇动作
    /// </summary>
    [Key(3)]
    public string readyAction;

    /// <summary>
    /// 技能施法动作
    /// </summary>
    [Key(4)]
    public string castAction;

    /// <summary>
    /// 技能结束动作
    /// </summary>
    [Key(5)]
    public string endAction;

    /// <summary>
    /// 技能被击动作
    /// </summary>
    [Key(6)]
    public string beAttackAction;

    /// <summary>
    /// 技能被击特效
    /// </summary>
    [Key(7)]
    public UInt32 beAtkEffect;

    /// <summary>
    /// 技能起手音效
    /// </summary>
    [Key(8)]
    public UInt32 readySound;

    /// <summary>
    /// 技能释放音效
    /// </summary>
    [Key(9)]
    public UInt32 castSound;

    /// <summary>
    /// 死亡是否击飞
    /// </summary>
    [Key(10)]
    public byte deadFly;

    /// <summary>
    /// 引导类型
    /// </summary>
    [Key(11)]
    public byte pressesType;

    /// <summary>
    /// 引导参数
    /// </summary>
    [Key(12)]
    public List<string> pressesParam;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
