using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class ClientVariableData : IDataBase
{
    /// <summary>
    /// 变量id
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 变量名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 宿主类型
    /// </summary>
    [Key(2)]
    public UInt32 ownerType;

    /// <summary>
    /// 宿主id
    /// </summary>
    [Key(3)]
    public UInt32 ownerId;

    /// <summary>
    /// 变量类型
    /// </summary>
    [Key(4)]
    public UInt32 valueType;

    /// <summary>
    /// 算子
    /// </summary>
    [Key(5)]
    public UInt32 operatorAlt;

    /// <summary>
    /// 默认值
    /// </summary>
    [Key(6)]
    public string defaultValue;

    /// <summary>
    /// 重置规则
    /// </summary>
    [Key(7)]
    public UInt32 resetRule;

    /// <summary>
    /// 定时重置规则
    /// </summary>
    [Key(8)]
    public string resetRuleContent;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
