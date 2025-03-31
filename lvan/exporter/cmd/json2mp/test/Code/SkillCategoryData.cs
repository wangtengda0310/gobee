using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class SkillCategoryData : IDataBase
{
    /// <summary>
    /// 技能类别ID
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 类别名称
    /// </summary>
    [Key(1)]
    public string name;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
