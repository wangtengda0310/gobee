using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class SkillBodyData : IDataBase
{
    /// <summary>
    /// 技能体ID
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 技能体名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 技能体说明
    /// </summary>
    [Key(2)]
    public string desc;

    /// <summary>
    /// 标签
    /// </summary>
    [Key(3)]
    public List<UInt32> tag;

    /// <summary>
    /// 标签类
    /// </summary>
    [Key(4)]
    public List<UInt32> tagClass;

    /// <summary>
    /// 初始位置
    /// </summary>
    [Key(5)]
    public UInt32 initPos;

    /// <summary>
    /// 初始位置偏移参数
    /// </summary>
    [Key(6)]
    public List<int> initOffset;

    /// <summary>
    /// 从屏幕外射出
    /// </summary>
    [Key(7)]
    public byte offScreen;

    /// <summary>
    /// 巡航类型
    /// </summary>
    [Key(8)]
    public byte cruiseType;

    /// <summary>
    /// 巡航参数
    /// </summary>
    [Key(9)]
    public List<long> cruiseParams;

    /// <summary>
    /// 施放个数
    /// </summary>
    [Key(10)]
    public UInt16 castNum;

    /// <summary>
    /// 多子弹时间间隔
    /// </summary>
    [Key(11)]
    public UInt16 castInterval;

    /// <summary>
    /// 多子弹释放角度类型
    /// </summary>
    [Key(12)]
    public byte castAngleType;

    /// <summary>
    /// 多子弹角度间隔
    /// </summary>
    [Key(13)]
    public short castAngle;

    /// <summary>
    /// 技能体旋转规则
    /// </summary>
    [Key(14)]
    public byte rotationType;

    /// <summary>
    /// 技能体旋转规则参数
    /// </summary>
    [Key(15)]
    public List<int> rotationTypeParam;

    /// <summary>
    /// 初始角度
    /// </summary>
    [Key(16)]
    public List<int> initAngle;

    /// <summary>
    /// 速度
    /// </summary>
    [Key(17)]
    public List<int> moveSpeed;

    /// <summary>
    /// 穿透次数
    /// </summary>
    [Key(18)]
    public UInt32 penetrate;

    /// <summary>
    /// 空间间距
    /// </summary>
    [Key(19)]
    public List<long> spaceGap;

    /// <summary>
    /// 无视阻挡
    /// </summary>
    [Key(20)]
    public UInt32 ignoreBlock;

    /// <summary>
    /// 体积缩放
    /// </summary>
    [Key(21)]
    public UInt16 scale;

    /// <summary>
    /// 碰撞盒
    /// </summary>
    [Key(22)]
    public List<int> collisionBox;

    /// <summary>
    /// 实体碰撞规则
    /// </summary>
    [Key(23)]
    public List<int> collision;

    /// <summary>
    /// 是否允许被碰撞
    /// </summary>
    [Key(24)]
    public byte allowCollided;

    /// <summary>
    /// 边缘反弹规则
    /// </summary>
    [Key(25)]
    public byte edgeBounceRule;

    /// <summary>
    /// 边缘反弹参数
    /// </summary>
    [Key(26)]
    public List<int> edgeBounceParams;

    /// <summary>
    /// 边缘反弹消耗穿透次数
    /// </summary>
    [Key(27)]
    public byte bounceDeduce;

    /// <summary>
    /// 消失类型和参数
    /// </summary>
    [Key(28)]
    public UInt32[][] disapperParams;

    /// <summary>
    /// 碰撞效果
    /// </summary>
    [Key(29)]
    public UInt32[][] collisionEffect;

    /// <summary>
    /// 结束效果
    /// </summary>
    [Key(30)]
    public List<UInt32> endEffect;

    /// <summary>
    /// 挂载主动技能
    /// </summary>
    [Key(31)]
    public List<UInt32> skills;

    /// <summary>
    /// 是否碰撞父级目标
    /// </summary>
    [Key(32)]
    public byte collideParentTarget;

    /// <summary>
    /// 结束碰撞执行效果（持续碰撞体）
    /// </summary>
    [Key(33)]
    public List<UInt32> endColEff;

    /// <summary>
    /// 美术表现
    /// </summary>
    [Key(34)]
    public List<UInt32> artConfig;

    /// <summary>
    /// 碰撞表现
    /// </summary>
    [Key(35)]
    public List<UInt32> hitArtConfig;

    /// <summary>
    /// 结束表现
    /// </summary>
    [Key(36)]
    public List<UInt32> endArtConfig;

    /// <summary>
    /// 初始锚点替换规则
    /// </summary>
    [Key(37)]
    public UInt32[][] initAnchorReplace;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
