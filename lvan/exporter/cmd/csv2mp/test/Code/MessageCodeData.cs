using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class MessageCodeData : IDataBase
{
    /// <summary>
    /// 提示语ID
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 提示语名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 提示语类型
    /// </summary>
    [Key(2)]
    public byte type;

    /// <summary>
    /// 内容
    /// </summary>
    [Key(3)]
    public string message;

    /// <summary>
    /// 动态参数
    /// </summary>
    [Key(4)]
    public string[][] paramMeta;

    /// <summary>
    /// UI模块
    /// </summary>
    [Key(5)]
    public UInt32 uiModuleId;

    /// <summary>
    /// 提示语背景
    /// </summary>
    [Key(6)]
    public string bg;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
