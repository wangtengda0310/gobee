using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class ShopPlayerGroupData : IDataBase
{
    /// <summary>
    /// 玩家分群ID
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 类型名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 玩家分群条件
    /// </summary>
    [Key(2)]
    public string playerGroupCondition;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
