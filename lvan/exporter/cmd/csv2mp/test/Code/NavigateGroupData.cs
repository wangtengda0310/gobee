using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class NavigateGroupData : IDataBase
{
    /// <summary>
    /// 导航组ID
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 导航组名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 显示区域
    /// </summary>
    [Key(2)]
    public UInt32 area;

    /// <summary>
    /// 父级导航项
    /// </summary>
    [Key(3)]
    public List<UInt32> parent;

    /// <summary>
    /// 打开模式
    /// </summary>
    [Key(4)]
    public UInt32 openMode;

    /// <summary>
    /// 默认导航项
    /// </summary>
    [Key(5)]
    public UInt32 defaultItem;

    /// <summary>
    /// 打开模式参数
    /// </summary>
    [Key(6)]
    public string openModeParam;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
