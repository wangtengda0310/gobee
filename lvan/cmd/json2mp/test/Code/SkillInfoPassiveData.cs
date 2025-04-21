using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class SkillInfoPassiveData : IDataBase
{
    /// <summary>
    /// 编号
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 技能id
    /// </summary>
    [Key(1)]
    public UInt32 skillId;

    /// <summary>
    /// 技能名称
    /// </summary>
    [Key(2)]
    public string name;

    /// <summary>
    /// 技能说明
    /// </summary>
    [Key(3)]
    public string skillInfo;

    /// <summary>
    /// 技能图标
    /// </summary>
    [Key(4)]
    public string icon;

    /// <summary>
    /// 技能品质
    /// </summary>
    [Key(5)]
    public byte quality;

    /// <summary>
    /// 等级
    /// </summary>
    [Key(6)]
    public UInt32 level;

    /// <summary>
    /// 升级条件
    /// </summary>
    [Key(7)]
    public string upgradeCondition;

    /// <summary>
    /// 技能属性
    /// </summary>
    [Key(8)]
    public int[][] skillAttribute;

    /// <summary>
    /// 技能战力
    /// </summary>
    [Key(9)]
    public UInt32 skillCE;

    /// <summary>
    /// 技能效果
    /// </summary>
    [Key(10)]
    public List<UInt32> passiveEffect;

    /// <summary>
    /// 技能效果说明
    /// </summary>
    [Key(11)]
    public string passiveEffectDes;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
