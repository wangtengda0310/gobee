# 常见陷阱与调试参考

> 本文件包含活动Wiki开发中的常见陷阱与解决方案、数据过滤与关联范围控制、以及现有活动类型和关联表的参考信息。
>
> 内容来源：SKILL.md 的"常见陷阱与调试"、"数据过滤与关联范围控制"、"现有活动类型和关联表"章节。

## 目录

- [现有活动类型和关联表](#现有活动类型和关联表)
- [数据过滤与关联范围控制](#数据过滤与关联范围控制)
  - [问题场景](#问题场景)
  - [过滤原则](#过滤原则)
  - [过滤实现示例](#过滤实现示例)
  - [新增配表时的过滤设计](#新增配表时的过滤设计)
  - [反模式警告](#反模式警告)
- [常见陷阱与解决方案](#常见陷阱与解决方案)
  - [陷阱1：数据格式误判](#陷阱1数据格式误判)
  - [陷阱2：嵌套条件分支](#陷阱2嵌套条件分支)
  - [陷阱3：反向关联遗漏](#陷阱3反向关联遗漏)
  - [陷阱4：调试信息不足](#陷阱4调试信息不足)

---

## 现有活动类型和关联表

| 活动类型 | 标识 | 关联模式 | 关联表 |
|----------|------|----------|--------|
| 丹青阁（皮肤抽奖） | `ActTypeSkinRaffle` | 时间匹配 | Activity, DrawSkin, DropRule, DropGroup, DropItem, LimitSkinTimesReward, Shop, ShopGoods, HeroSkinCollition, ItemHeroSkin, HeroSkinItem, HeroSkinSpine |
| 结缘亭 | `ActTypeDrawPet` | 时间匹配+反向关联 | Activity, DrawPet, Pet, PetAudio, DropRule, DropGroup, DropItem |
| 赛季战令 | 独立子系统（无ActivityType） | 时间匹配 | SeasonPass, SeasonPassBag, SeasonPassReward, SeasonPassTask |

## 数据过滤与关联范围控制

### 问题场景

某些配表（如Item、Task等）可能包含**全量数据**，数据量庞大（数千至上万条），且被多个活动共享。如果直接展示全量数据：
- 页面加载缓慢，内存占用高
- 无关数据干扰阅读，降低信息密度
- 前端渲染性能下降

### 过滤原则

**核心原则：只展示与当前活动直接关联的数据。**

目前代码中的过滤模式：

| 过滤方式 | 说明 | 示例 |
|----------|------|------|
| **ID精确匹配** | 通过唯一ID关联单条记录 | DropRule(Id)、ItemHeroSkin(SkinItemId) |
| **枚举字符串过滤** | 通过ActIdStr/EActivityId匹配子集 | LimitSkinTimesReward(ActIdStr) |
| **类型枚举过滤** | 通过Type字段匹配子集 | Shop(ShopType)、ShopGoods(ShopType) |
| **链式推导过滤** | 通过关联链推导相关数据 | DropItem(DropGroupId) |
| **时间匹配过滤** | 按StartTime/EndTime匹配当前期，无则取最新 | DrawSkin、SeasonPass、DrawPet |

### 过滤实现示例

**示例1：通过ActIdStr过滤子集数据（LimitSkinTimesReward）**

```go
// 构建索引时：按ActIdStr分组，每个活动只取自己的分组
func buildTimesRewardsByStrIndex(timesRewardDiff *[]limit_skin_times_reward.LimitSkinTimesRewardDiff) map[string][]*limit_skin_times_reward.LimitSkinTimesRewardDiff {
    index := make(map[string][]*limit_skin_times_reward.LimitSkinTimesRewardDiff)
    if timesRewardDiff == nil {
        return index
    }
    for i := range *timesRewardDiff {
        tr := &(*timesRewardDiff)[i]
        key := tr.ActIdStr  // 如 "Activity_LimitTimeSkin"
        if key == "" {
            key = "invalid"
        }
        index[key] = append(index[key], tr)  // 同一活动的奖励归为一组
    }
    return index
}

// 使用时：只取当前活动EActivityId对应的分组
if rewards, ok := timesRewardsByActIdStr[act.EActivityId]; ok {
    data.TimesRewards = rewards  // 只包含该活动的奖励，而非全表
}
```

**示例2：通过链式关联过滤（DropItem）**

```go
// DropItem全表可能有数千条，但只取当前DropRule关联的DropGroup对应的项
for _, dgId := range dr.DropGroup {  // 遍历当前规则的所有掉落组
    if dg, ok := dropGroupById[dgId]; ok {
        data.DropGroups = append(data.DropGroups, dg)
        if items, ok := dropItemsByGroupId[dgId]; ok {
            data.DropItems = append(data.DropItems, items...)  // 只取这些组的Item
        }
    }
}
```

**示例3：通过ShopType过滤商店商品（ShopGoods）**

```go
// ShopGoods全表可能有数百条，但只取当前活动类型的商品
if goods, ok := shopGoodsByShopType["ShopTypeSkinRaffle"]; ok {
    data.ShopGoods = goods  // 只包含皮肤抽奖商店的商品
}
```

### 新增配表时的过滤设计

当为活动新增关联配表时，必须考虑：

1. **该表是否可能被多个活动共享？**
   - 是 -> 需要设计过滤机制
   - 否 -> 可直接关联单条记录

2. **如何确定一条记录属于当前活动？**
   - 检查表中是否有 `ActivityId`、`ActId`、`ActIdStr`、`ShopType` 等关联字段
   - 检查是否可通过链式关联推导（如 A->B->C，通过A过滤B，通过B过滤C）

3. **过滤实现位置**
   - **推荐**：在 `BuildActivityWikiDiff` 中关联时过滤（后端过滤，减少数据传输）
   - 备选：在前端用 `computed` 过滤（仅适用于数据量极小的情况）

### 反模式警告

**不要这样做**：

```go
// 错误：将全量数据直接关联到活动
data.AllItems = container.ItemDiff  // 全表数千条，无过滤
```

```vue
<!-- 错误：前端展示全量列表 -->
<n-table :data="props.activityData.AllItems">  <!-- 数据量过大，渲染缓慢 -->
```

**正确做法**：

```go
// 正确：只关联当前活动相关的子集
func buildItemsByActivityId(itemDiff *[]item.ItemDiff) map[int][]*item.ItemDiff {
    index := make(map[int][]*item.ItemDiff)
    for i := range *itemDiff {
        it := &(*itemDiff)[i]
        if it.ActivityId > 0 {
            index[it.ActivityId] = append(index[it.ActivityId], it)
        }
    }
    return index
}

// 使用时只取当前活动的
if items, ok := itemsByActivityId[act.Id]; ok {
    data.Items = items  // 只包含该活动的道具
}
```

## 常见陷阱与解决方案

### 陷阱1：数据格式误判

**现象**：代码逻辑正确，但解析出的字段值为空或默认值。

**根因**：没有验证Excel实际数据格式，凭经验假设。

**案例**：
- 假设ActivityId是`{id;weight}`格式，实际是简单整数
- 假设CustomParma是逗号分隔数组，实际是空值

**解决方案**：
1. 用Python脚本读取Excel打印原始值
2. 根据实际格式编写解析代码
3. 添加解析失败的日志输出

### 陷阱2：嵌套条件分支

**现象**：新活动类型的逻辑永远不会执行。

**根因**：代码被嵌套在另一个活动类型的if块内部。

**解决方案**：
1. 使用`if/else if`平行结构
2. 代码审查时检查缩进层级
3. 添加单元测试验证各分支

### 陷阱3：反向关联遗漏

**现象**：CustomParma为空时，数据无法关联。

**根因**：只考虑了正向关联（Activity -> 目标表），未考虑反向关联。

**解决方案**：
1. 检查Activity表的CustomParma是否可能为空
2. 如果可能为空，检查目标表是否有ActivityId字段
3. 实现反向关联作为备选方案

### 陷阱4：调试信息不足

**现象**：页签不显示，无法快速定位问题。

**解决方案**：
1. 在ActivityCompleteData中添加调试字段
2. 前端显示关联数据状态（如CustomParma值、关联结果等）
3. 后端添加日志输出关键索引的大小和内容

### 陷阱5：时间匹配展示过期数据

**现象**：活动展示的配表数据不是当前进行中的那一期，而是CustomParma[0]指向的过期数据。

**根因**：直接使用 `CustomParma[0] -> DrawSkin.Id` 关联，但 CustomParma 指向的可能已过期。或者遍历map取第一个元素，顺序不确定。

**解决方案**：
1. 不要直接用 CustomParma[0] 关联多期数据表，改用时间匹配 `findCurrentXxx(sorted)`
2. 全表按 StartTime 排序后，优先匹配 StartTime <= now <= EndTime 的记录
3. 无进行中的数据时返回最新一期（sorted 最后一个）
4. 参见 code-templates.md 中的"时间匹配模式"章节

### 陷阱6：三期数据用 map 遍历顺序不确定

**现象**：战令/丹青阁展示的三期数据顺序每次刷新可能不同。

**根因**：Go map 遍历顺序是随机的，直接遍历 map 构建数据会导致顺序不确定。

**解决方案**：
1. 使用 `buildAllSortedXxx()` 按 StartTime 排序后再处理
2. 在排序切片中用索引位置确定上一期/下一期，而非依赖 map 顺序

---

> 内容来源：SKILL.md 的"常见陷阱与调试"、"数据过滤与关联范围控制"、"现有活动类型和关联表"章节。拆分时间：2026-04-30。
