using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class HaloData : IDataBase
{
    /// <summary>
    /// 光环ID
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 光环名
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 半径
    /// </summary>
    [Key(2)]
    public UInt16 radius;

    /// <summary>
    /// 范围拾取过滤规则
    /// </summary>
    [Key(3)]
    public int[][] pickFilter;

    /// <summary>
    /// 命中效果id列表
    /// </summary>
    [Key(4)]
    public List<UInt32> intEffect;

    /// <summary>
    /// 结束效果id列表
    /// </summary>
    [Key(5)]
    public List<UInt32> outEffect;

    /// <summary>
    /// 特殊效果
    /// </summary>
    [Key(6)]
    public UInt32[][] specialEffect;

    /// <summary>
    /// 进入范围后添加的持续效果id列表
    /// </summary>
    [Key(7)]
    public List<UInt32> durationEffect;

    /// <summary>
    /// 持续效果距离系数类型
    /// </summary>
    [Key(8)]
    public byte attenuationType;

    /// <summary>
    /// 持续效果距离系数
    /// </summary>
    [Key(9)]
    public List<UInt32> attenuationParams;

    /// <summary>
    /// 死亡后继续生效
    /// </summary>
    [Key(10)]
    public byte deathActive;

    /// <summary>
    /// 过滤施法者
    /// </summary>
    [Key(11)]
    public byte filterCaster;

    /// <summary>
    /// 表现
    /// </summary>
    [Key(12)]
    public UInt32 art;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
