using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class PlotDialogueUnit_10004Data : IDataBase
{
    /// <summary>
    /// 对话单元id
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 描述
    /// </summary>
    [Key(2)]
    public string desc;

    /// <summary>
    /// 所属对话
    /// </summary>
    [Key(3)]
    public UInt32 plotDialogue;

    /// <summary>
    /// 所属对话集
    /// </summary>
    [Key(4)]
    public UInt32 plotDialogueSet;

    /// <summary>
    /// 序号
    /// </summary>
    [Key(5)]
    public UInt32 order;

    /// <summary>
    /// 文本
    /// </summary>
    [Key(6)]
    public string text;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
