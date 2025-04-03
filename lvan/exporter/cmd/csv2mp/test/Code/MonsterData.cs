using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class MonsterData : IDataBase
{
    /// <summary>
    /// 怪物ID
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 编辑器标签
    /// </summary>
    [Key(1)]
    public string tag;

    /// <summary>
    /// 怪物名称
    /// </summary>
    [Key(2)]
    public string name;

    /// <summary>
    /// 显示名称
    /// </summary>
    [Key(3)]
    public string displayName;

    /// <summary>
    /// 描述
    /// </summary>
    [Key(4)]
    public string desc;

    /// <summary>
    /// 怪物类型
    /// </summary>
    [Key(5)]
    public byte type;

    /// <summary>
    /// 怪物等级
    /// </summary>
    [Key(6)]
    public UInt16 level;

    /// <summary>
    /// 怪物头像
    /// </summary>
    [Key(7)]
    public string icon;

    /// <summary>
    /// 怪物模型
    /// </summary>
    [Key(8)]
    public string model;

    /// <summary>
    /// 怪物预制体
    /// </summary>
    [Key(9)]
    public UInt32 prefab;

    /// <summary>
    /// 模型配置
    /// </summary>
    [Key(10)]
    public string modelConfig;

    /// <summary>
    /// 立绘
    /// </summary>
    [Key(11)]
    public string portrait;

    /// <summary>
    /// 立绘参数配置
    /// </summary>
    [Key(12)]
    public string portraitConfig;

    /// <summary>
    /// 阴影
    /// </summary>
    [Key(13)]
    public string shadow;

    /// <summary>
    /// 阴影参数配置
    /// </summary>
    [Key(14)]
    public string shadowConfig;

    /// <summary>
    /// 待机动作
    /// </summary>
    [Key(15)]
    public string idleAnimation;

    /// <summary>
    /// 水平镜像移动
    /// </summary>
    [Key(16)]
    public byte horizontalFlip;

    /// <summary>
    /// 碰撞盒
    /// </summary>
    [Key(17)]
    public List<int> collisionBox;

    /// <summary>
    /// 怪物属性
    /// </summary>
    [Key(18)]
    public int[][] attr;

    /// <summary>
    /// 技能栏
    /// </summary>
    [Key(19)]
    public List<UInt32> skillBar;

    /// <summary>
    /// 怪物技能组
    /// </summary>
    [Key(20)]
    public List<KeyValuePair<UInt32, UInt32>> skill;

    /// <summary>
    /// 怪物巡逻半径
    /// </summary>
    [Key(21)]
    public UInt32 patrolRadius;

    /// <summary>
    /// 怪物视野半径
    /// </summary>
    [Key(22)]
    public UInt32 viewRadius;

    /// <summary>
    /// 怪物追击半径
    /// </summary>
    [Key(23)]
    public UInt32 chaseRadius;

    /// <summary>
    /// 固定掉落
    /// </summary>
    [Key(24)]
    public List<KeyValuePair<UInt32, UInt32>> fixedDrop;

    /// <summary>
    /// 概率掉落
    /// </summary>
    [Key(25)]
    public List<UInt32> personalProbDrop;

    /// <summary>
    /// 掉落触发时机
    /// </summary>
    [Key(26)]
    public List<UInt32> dropTrigger;

    /// <summary>
    /// 血量比例掉落
    /// </summary>
    [Key(27)]
    public UInt32[][] proportionDrop;

    /// <summary>
    /// 掉落校验规则
    /// </summary>
    [Key(28)]
    public UInt64 dropSettleRule;

    /// <summary>
    /// 掉落预览
    /// </summary>
    [Key(29)]
    public UInt32[][] dropPreview;

    /// <summary>
    /// 受击表现
    /// </summary>
    [Key(30)]
    public List<int> hit;

    /// <summary>
    /// 受击动作
    /// </summary>
    [Key(31)]
    public string hitAction;

    /// <summary>
    /// 受击表现冷却
    /// </summary>
    [Key(32)]
    public UInt32 hitCoolDown;

    /// <summary>
    /// 死亡动画
    /// </summary>
    [Key(33)]
    public UInt32 deadAnimation;

    /// <summary>
    /// 死亡墓碑
    /// </summary>
    [Key(34)]
    public string tombstone;

    /// <summary>
    /// 是否播放死亡动作
    /// </summary>
    [Key(35)]
    public byte isDeadAction;

    /// <summary>
    /// AI
    /// </summary>
    [Key(36)]
    public UInt32 ai;

    /// <summary>
    /// 反应时间(毫秒)
    /// </summary>
    [Key(37)]
    public UInt32 reactionTime;

    /// <summary>
    /// 标签
    /// </summary>
    [Key(38)]
    public List<UInt32> label;

    /// <summary>
    /// 标签类
    /// </summary>
    [Key(39)]
    public List<UInt32> labelClass;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
