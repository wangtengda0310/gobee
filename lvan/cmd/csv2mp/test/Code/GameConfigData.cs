using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class GameConfigData : IDataBase
{
    /// <summary>
    /// ID
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 类型
    /// </summary>
    [Key(2)]
    public UInt16 type;

    /// <summary>
    /// 全局配置功能
    /// </summary>
    [Key(3)]
    public string func;

    /// <summary>
    /// 全局配置值
    /// </summary>
    [Key(4)]
    public UInt32 value;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
