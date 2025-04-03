using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class SkillInfoActiveData : IDataBase
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
    public UInt32 quality;

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
    /// 消耗组
    /// </summary>
    [Key(8)]
    public UInt32[][] costGroup;

    /// <summary>
    /// 消耗
    /// </summary>
    [Key(9)]
    public UInt32[][] cost;

    /// <summary>
    /// 技能战力
    /// </summary>
    [Key(10)]
    public UInt32 skillCE;

    /// <summary>
    /// 技能属性
    /// </summary>
    [Key(11)]
    public int[][] skillAttribute;

    /// <summary>
    /// 最小施放距离
    /// </summary>
    [Key(12)]
    public UInt32 minDistance;

    /// <summary>
    /// 最大施放距离
    /// </summary>
    [Key(13)]
    public UInt32 maxDistance;

    /// <summary>
    /// 强制锁敌
    /// </summary>
    [Key(14)]
    public List<UInt32> lockTarget;

    /// <summary>
    /// 技能目标是否允许被嘲讽改变
    /// </summary>
    [Key(15)]
    public byte beTaunted;

    /// <summary>
    /// 施法方式
    /// </summary>
    [Key(16)]
    public UInt32 castType;

    /// <summary>
    /// 引导次数
    /// </summary>
    [Key(17)]
    public UInt32 repeatCount;

    /// <summary>
    /// 引导间隔时间（毫秒）
    /// </summary>
    [Key(18)]
    public UInt32 repeatInterval;

    /// <summary>
    /// 是否允许移动施法
    /// </summary>
    [Key(19)]
    public byte moveCast;

    /// <summary>
    /// 最大充能层数
    /// </summary>
    [Key(20)]
    public UInt32 chargeMaxCount;

    /// <summary>
    /// 充能CD
    /// </summary>
    [Key(21)]
    public UInt32 chargeCD;

    /// <summary>
    /// 初始充能层数
    /// </summary>
    [Key(22)]
    public UInt32 chargeCount;

    /// <summary>
    /// 引导时长
    /// </summary>
    [Key(23)]
    public UInt32 channelingTime;

    /// <summary>
    /// 引导CD时机
    /// </summary>
    [Key(24)]
    public UInt32 channelingCdTiming;

    /// <summary>
    /// 引导绑定单位
    /// </summary>
    [Key(25)]
    public UInt32[][] channelingBind;

    /// <summary>
    /// 释放效果
    /// </summary>
    [Key(26)]
    public List<UInt32> castEffect;

    /// <summary>
    /// 施放阶段时间（毫秒）
    /// </summary>
    [Key(27)]
    public UInt32 castTime;

    /// <summary>
    /// 技能施放效果说明
    /// </summary>
    [Key(28)]
    public string castEffectDes;

    /// <summary>
    /// 施放阶段打断规则
    /// </summary>
    [Key(29)]
    public int[][] castBreakRule;

    /// <summary>
    /// 技能激活效果
    /// </summary>
    [Key(30)]
    public List<UInt32> activeEffect;

    /// <summary>
    /// 激活阶段时间（毫秒）
    /// </summary>
    [Key(31)]
    public UInt32 activeTime;

    /// <summary>
    /// 激活阶段打断规则
    /// </summary>
    [Key(32)]
    public int[][] activeBreakRule;

    /// <summary>
    /// 结束效果
    /// </summary>
    [Key(33)]
    public List<UInt32> endEffect;

    /// <summary>
    /// 后摇时长（毫秒）
    /// </summary>
    [Key(34)]
    public UInt32 endTime;

    /// <summary>
    /// 结束阶段打断规则
    /// </summary>
    [Key(35)]
    public int[][] endBreakRule;

    /// <summary>
    /// 表现
    /// </summary>
    [Key(36)]
    public string art;

    /// <summary>
    /// 释放顺序
    /// </summary>
    [Key(37)]
    public UInt32 castOrder;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
