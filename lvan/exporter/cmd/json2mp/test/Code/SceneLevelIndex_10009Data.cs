using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class SceneLevelIndex_10009Data : IDataBase
{
    /// <summary>
    /// 关卡波次id
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 描述
    /// </summary>
    [Key(2)]
    public string desc;

    /// <summary>
    /// 所属关卡
    /// </summary>
    [Key(3)]
    public UInt32 sceneLevel;

    /// <summary>
    /// 波次
    /// </summary>
    [Key(4)]
    public UInt32 index;

    /// <summary>
    /// 事件流
    /// </summary>
    [Key(5)]
    public UInt32 eventFlow;

    /// <summary>
    /// 关卡初始化配置
    /// </summary>
    [Key(6)]
    public UInt32[][] initConfig;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
