using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class MailData : IDataBase
{
    /// <summary>
    /// 邮件ID
    /// </summary>
    [Key(0)]
    public UInt32 id;

    /// <summary>
    /// 名称
    /// </summary>
    [Key(1)]
    public string name;

    /// <summary>
    /// 标题
    /// </summary>
    [Key(2)]
    public string title;

    /// <summary>
    /// 正文提示语id
    /// </summary>
    [Key(3)]
    public List<UInt32> body;

    /// <summary>
    /// 发送模式
    /// </summary>
    [Key(4)]
    public byte sendMode;

    /// <summary>
    /// 发送模式参数
    /// </summary>
    [Key(5)]
    public string sendModeParam;

    /// <summary>
    /// 过滤条件
    /// </summary>
    [Key(6)]
    public string filterCondition;

    /// <summary>
    /// 邮件显示奖励
    /// </summary>
    [Key(7)]
    public UInt32[][] rewardShow;

    /// <summary>
    /// 奖励
    /// </summary>
    [Key(8)]
    public UInt32[][] reward;

    /// <summary>
    /// 激活天数
    /// </summary>
    [Key(9)]
    public UInt32 activationDays;

    /// <summary>
    /// 邮箱类型id
    /// </summary>
    [Key(10)]
    public UInt32 mailBoxTypeId;

    /// <summary>
    /// 邮件领取数量上限
    /// </summary>
    [Key(11)]
    public UInt32 mailLimit;

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}
