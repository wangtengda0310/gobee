using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class PetData : IDataBase
{
    /// <summary>
    /// id
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 显示名称
    /// </summary>
    [Key(2)]
    public string displayName;

    /// <summary>
    /// 品质
    /// </summary>
    [Key(3)]
    public UInt32 quality;

    /// <summary>
    /// 宠物系统
    /// </summary>
    [Key(4)]
    public UInt32 classAlt;

    /// <summary>
    /// 宠物类别
    /// </summary>
    [Key(5)]
    public UInt32 slotType;

    /// <summary>
    /// 初始养成线
    /// </summary>
    [Key(6)]
    public UInt32[][] growth;

    /// <summary>
    /// 初始技能
    /// </summary>
    [Key(7)]
    public UInt32[][] skill;

    /// <summary>
    /// 技能栏
    /// </summary>
    [Key(8)]
    public List<UInt32> skillBar;

    /// <summary>
    /// 固定属性
    /// </summary>
    [Key(9)]
    public int[][] fixedAttr;

    /// <summary>
    /// 随机属性
    /// </summary>
    [Key(10)]
    public List<UInt32> randomAttr;

    /// <summary>
    /// 固定道具
    /// </summary>
    [Key(11)]
    public UInt32[][] fixedItem;

    /// <summary>
    /// 随机道具
    /// </summary>
    [Key(12)]
    public UInt32[][] randomItem;

    /// <summary>
    /// AI
    /// </summary>
    [Key(13)]
    public UInt32 ai;

    /// <summary>
    /// 模型
    /// </summary>
    [Key(14)]
    public string model;

    /// <summary>
    /// 宠物预制体
    /// </summary>
    [Key(15)]
    public UInt32 prefab;

    /// <summary>
    /// 模型配置
    /// </summary>
    [Key(16)]
    public string modelConfig;

    /// <summary>
    /// 头像
    /// </summary>
    [Key(17)]
    public string avatar;

    /// <summary>
    /// 立绘
    /// </summary>
    [Key(18)]
    public string portrait;

    /// <summary>
    /// 立绘配置
    /// </summary>
    [Key(19)]
    public string portraitConfig;

    /// <summary>
    /// 阴影
    /// </summary>
    [Key(20)]
    public string shadow;

    /// <summary>
    /// 阴影配置
    /// </summary>
    [Key(21)]
    public string shadowConfig;

    /// <summary>
    /// 水平镜像移动
    /// </summary>
    [Key(22)]
    public byte horizontalFlip;

    /// <summary>
    /// 碰撞盒
    /// </summary>
    [Key(23)]
    public List<int> collisionBox;

    /// <summary>
    /// 受击表现
    /// </summary>
    [Key(24)]
    public List<int> hit;

    /// <summary>
    /// 受击动作
    /// </summary>
    [Key(25)]
    public string hitAction;

    /// <summary>
    /// 受击表现冷却
    /// </summary>
    [Key(26)]
    public UInt32 hitCoolDown;

    /// <summary>
    /// 死亡动画
    /// </summary>
    [Key(27)]
    public UInt32 deadAnimation;

    /// <summary>
    /// 死亡墓碑
    /// </summary>
    [Key(28)]
    public string tombstone;

    /// <summary>
    /// 宠物品级类
    /// </summary>
    [Key(29)]
    public List<UInt32> petLevelClass;

    /// <summary>
    /// 宠物品级
    /// </summary>
    [Key(30)]
    public List<UInt32> petLevel;

    /// <summary>
    /// 标签
    /// </summary>
    [Key(31)]
    public List<UInt32> label;

    /// <summary>
    /// 标签类
    /// </summary>
    [Key(32)]
    public List<UInt32> labelClass;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
