using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class ComposeCategoryData : IDataBase
{
    /// <summary>
    /// 配方类别id
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 父类别id
    /// </summary>
    [Key(2)]
    public UInt32 parentCategoryId;

    /// <summary>
    /// 合成系统id
    /// </summary>
    [Key(3)]
    public UInt32 classId;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
