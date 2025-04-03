using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class MailBoxTypeData : IDataBase
{
    /// <summary>
    /// 邮箱类型id
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 邮箱类型描述
    /// </summary>
    [Key(2)]
    public string desc;

    /// <summary>
    /// 邮箱领取数量上限
    /// </summary>
    [Key(3)]
    public UInt32 mailBoxLimit;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
