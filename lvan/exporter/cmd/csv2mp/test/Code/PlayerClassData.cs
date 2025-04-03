using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class PlayerClassData : IDataBase
{
    /// <summary>
    /// id
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 职业名称
    /// </summary>
    [Key(1)]
    public string className;

    /// <summary>
    /// 默认玩法
    /// </summary>
    [Key(2)]
    public UInt32 world;

    /// <summary>
    /// 是否默认角色
    /// </summary>
    [Key(3)]
    public UInt32 isDefault;

    /// <summary>
    /// 技能栏
    /// </summary>
    [Key(4)]
    public List<UInt32> skillBar;

    /// <summary>
    /// 初始线性养成等级
    /// </summary>
    [Key(5)]
    public List<KeyValuePair<UInt32, UInt32>> growthLinear;

    /// <summary>
    /// 初始装备养成线
    /// </summary>
    [Key(6)]
    public UInt32[][] growthEquip;

    /// <summary>
    /// 初始属性
    /// </summary>
    [Key(7)]
    public int[][] attr;

    /// <summary>
    /// 初始技能
    /// </summary>
    [Key(8)]
    public List<KeyValuePair<UInt32, UInt32>> skill;

    /// <summary>
    /// 初始道具
    /// </summary>
    [Key(9)]
    public List<KeyValuePair<UInt32, UInt32>> item;

    /// <summary>
    /// 角色模型
    /// </summary>
    [Key(10)]
    public string model;

    /// <summary>
    /// 角色预制体
    /// </summary>
    [Key(11)]
    public UInt32 prefab;

    /// <summary>
    /// 模型参数
    /// </summary>
    [Key(12)]
    public string modelConfig;

    /// <summary>
    /// AI
    /// </summary>
    [Key(13)]
    public UInt32 ai;

    /// <summary>
    /// 反应时间(毫秒)
    /// </summary>
    [Key(14)]
    public UInt32 reactionTime;

    /// <summary>
    /// 头像
    /// </summary>
    [Key(15)]
    public string icon;

    /// <summary>
    /// 立绘
    /// </summary>
    [Key(16)]
    public string portrait;

    /// <summary>
    /// 立绘配置
    /// </summary>
    [Key(17)]
    public string portraitConfig;

    /// <summary>
    /// 阴影
    /// </summary>
    [Key(18)]
    public string shadow;

    /// <summary>
    /// 阴影配置
    /// </summary>
    [Key(19)]
    public string shadowConfig;

    /// <summary>
    /// 碰撞盒
    /// </summary>
    [Key(20)]
    public List<int> collisionBox;

    /// <summary>
    /// 受击表现
    /// </summary>
    [Key(21)]
    public List<int> hit;

    /// <summary>
    /// 受击动作
    /// </summary>
    [Key(22)]
    public string hitAction;

    /// <summary>
    /// 受击表现冷却
    /// </summary>
    [Key(23)]
    public UInt32 hitCoolDown;

    /// <summary>
    /// 死亡动画
    /// </summary>
    [Key(24)]
    public UInt32 deadAnimation;

    /// <summary>
    /// 死亡墓碑
    /// </summary>
    [Key(25)]
    public string tombstone;

    /// <summary>
    /// 标签
    /// </summary>
    [Key(26)]
    public List<UInt32> label;

    /// <summary>
    /// 标签类
    /// </summary>
    [Key(27)]
    public List<UInt32> labelClass;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
