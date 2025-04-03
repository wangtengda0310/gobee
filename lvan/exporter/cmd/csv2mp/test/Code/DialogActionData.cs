using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class DialogActionData : IDataBase
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
    /// 对话ID
    /// </summary>
    [Key(2)]
    public UInt32 dialogId;

    /// <summary>
    /// 行为序号
    /// </summary>
    [Key(3)]
    public UInt32 order;

    /// <summary>
    /// 行为类型
    /// </summary>
    [Key(4)]
    public UInt32 type;

    /// <summary>
    /// 行为参数
    /// </summary>
    [Key(5)]
    public string typeParam;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
