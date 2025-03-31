using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class UIModuleData : IDataBase
{
    /// <summary>
    /// UI模块ID
    /// </summary>
    [Key(0)]
    public UInt16 id;

    /// <summary>
    /// UI模块名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 模块类型
    /// </summary>
    [Key(2)]
    public UInt32 type;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
