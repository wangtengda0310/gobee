# activity/ 目录文档

活动相关的跨表校验规则（package: `activity`）。

## 规则文件

| 文件 | 规则类型 | 说明 | 源表 | 依赖表 |
|------|---------|------|------|--------|
| danqingge_customparam.go | DANQINGGE_CUSTOM_PARAM_IS_ITEMID_CHECK | 丹青阁自定义参数检查：CustomParma 非空且 DrawSkin 表 ID 存在 | Activity | DrawSkin |
| drawskin_cross_reference.go | ACTIVITY_DRAWSKIN_CROSS_REFERENCE_CHECK | Activity 与 DrawSkin 交叉引用检查：双向一致性、活动类型校验 | DrawSkin, Activity | Activity, DrawSkin |
| drawskin_time_overlap.go | ACTIVITY_DRAWSKIN_TIME_OVERLAP_CHECK | Activity 与 DrawSkin 时间交集检查：关联活动与抽奖池时间范围有交集 | DrawSkin | Activity |
| task_reward.go | ACTIVITY_TASK_REWARD_CHECK | 活动任务奖励检查：奖励道具在 Item 表存在且数量 > 0 | ActivityTask | Item |

## 开发注意事项

- 丹青阁活动识别：Name 列包含"丹青阁"关键字，类型标识为 `ActTypeSkinRaffle`
- 交叉引用检查包含三个子检查：DSK-05 引用检查、XC-01 双向一致性、XC-04 活动类型
- 时间交集公式：`DrawSkin.StartTime <= Activity.EndTime && Activity.StartTime <= DrawSkin.EndTime`
- 时间检查为 Warning 性质，不影响 Ok 状态
- CustomParma 未配置为 Warning(ACT-02)，格式错误或引用不存在为 Error
- Activity 表变更时只检查丹青阁活动的反向引用一致性
- DrawSkin 表变更时检查所有三个子检查
