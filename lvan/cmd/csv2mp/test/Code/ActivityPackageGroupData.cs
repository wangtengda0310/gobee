using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class ActivityPackageGroupData : IDataBase
{
    /// <summary>
    /// 活动包组ID
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 活动包组名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 包含活动包
    /// </summary>
    [Key(2)]
    public List<UInt32> activityPackage;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
