using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class VariableData : IDataBase
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
    /// 变量类型
    /// </summary>
    [Key(2)]
    public byte valueType;

    /// <summary>
    /// 宿主类型
    /// </summary>
    [Key(3)]
    public byte ownerType;

    /// <summary>
    /// 宿主id
    /// </summary>
    [Key(4)]
    public UInt32 ownerId;

    /// <summary>
    /// 重置规则
    /// </summary>
    [Key(5)]
    public byte resetRule;

    /// <summary>
    /// 重置规则内容
    /// </summary>
    [Key(6)]
    public string resetRuleContent;

    /// <summary>
    /// 算子
    /// </summary>
    [Key(7)]
    public byte operatorAlt;

    /// <summary>
    /// 变量默认值
    /// </summary>
    [Key(8)]
    public string defaultValue;

    /// <summary>
    /// 附加参数
    /// </summary>
    [Key(9)]
    public List<UInt32> paramsAlt;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
