# 活动 Wiki 检查模块

展示游戏活动的配表关联数据，仿照武将 Wiki 检查页面。

## 目录结构

```
activity-wiki-check/         # Wails 服务入口（本目录）
├── CLAUDE.md                # 本文档
├── wails.go                 # Check() 入口：加载 Excel → InitDiffRefExcel → BuildActivityWikiDiff
└── mcp.go                   # MCP 工具接口

activitywiki/                # 核心关联逻辑（pkg/rain-resources-checker/ 下）
└── format.go                # BuildActivityWikiDiff：按活动类型 switch 关联配表

activitywiki_def/            # 数据结构定义（pkg/rain-resources-checker/ 下）
└── def.go                   # ActivityWikiDiff、ActivityCompleteData

activity/                    # Activity 表解析（mjs_excel/ 下）
draw_skin/                   # DrawSkin 表解析
limit_skin_times_reward/     # LimitSkinTimesReward 表解析
drop_rule/                   # DropRule 表解析
drop_group/                  # DropGroup 表解析
drop_item/                   # DropItem 表解析

frontend/src/pages/activity-wiki-check/  # 前端页面
├── index.vue                # 主页面：筛选、列表、调用 Check()
└── components/
    └── activity-panel.vue   # 活动详情面板：基础信息 + 条件渲染 tab
```

## 已支持的活动类型

### 丹青阁 (ActTypeSkinRaffle)

关联链路：

```
Activity (Id=14, EActivityId="Activity_LimitTimeSkin")
  │
  ├─ CustomParma[0] → DrawSkin.Id
  │     ├─ OnceDropRule / TenDropRule → DropRule.Id
  │     │     └─ DropRule.DropGroup[] → DropGroup.Id
  │     │           └─ DropItem.DropGroup → DropGroup.Id
  │     └─ BigAwardCount, BigAwardItemId (保底配置)
  │
  └─ EActivityId(枚举字符串) → LimitSkinTimesReward.ActIdStr
        └─ DrawTimes=[20,40,60,80] 对应奖励记录
```

## 扩展新活动类型步骤

### 1. 确认关联关系

向策划/服务端确认：
- `CustomParma` 指向哪张表的哪个字段？
- 关联字段是数字 ID 还是枚举字符串（`E#xxx` 类型）？
- Excel 实际列顺序（不能假设与代码 iota 常量一致）

### 2. 确认 DataContainer 是否已加载关联表

查看 `pkg/rain-resources-checker/mjs_excel/diff_excel_init.go` 中的 `InitDiffRefExcel` 函数。

- **已加载**：直接使用 `container.XxxDiff` 字段
- **未加载**：
  1. 在 `mjs_excel/` 下新建解析包（参考 `draw_skin/` 目录结构：`def.go` + `diff_map.go`）
  2. 在 `diff/interface.go` 的 `DataContainer` 中新增字段
  3. 在 `diff_excel_init.go` 中调用解析函数并赋值

### 3. 修改后端数据结构和关联逻辑

```
activitywiki_def/def.go       → ActivityCompleteData 新增字段
activitywiki/format.go        → BuildActivityWikiDiff 新增 switch 分支
```

关联逻辑模板：

```go
case "ActTypeXxx":
    if len(act.CustomParma) > 0 {
        refId := act.CustomParma[0]
        if ref, ok := xxxIndex[refId]; ok {
            data.XxxRef = ref
            // 继续关联下游表...
        }
    }
```

### 4. 修改前端渲染

`activity-panel.vue` 中按 `props.activityData.XxxRef` 是否存在条件渲染新 tab。

### 5. 可复用的掉落链路

`DropRule → DropGroup → DropItem` 被多个活动共用（丹青阁、结缘亭、开年有兽等）。
`format.go` 中已构建 `dropRuleById`、`dropGroupById`、`dropItemsByGroupId` 三个索引，
任何活动类型只要拿到 DropRule.Id 即可关联到完整掉落数据。

## E2E 测试

| 测试文件 | 覆盖范围 |
|----------|----------|
| [`e2e/activity-wiki/activity-wiki.spec.ts`](../../../frontend/e2e/activity-wiki/activity-wiki.spec.ts) | 页面加载、活动列表、详情面板、Tab切换、数据关联展示 |

## 已知陷阱

| 问题 | 说明 |
|------|------|
| 列索引偏移 | Excel 表实际列顺序可能与 iota 连续递增不匹配，必须对照 Excel 第3行字段名确认。用显式数字常量替代 iota。 |
| 枚举字段 | 类型标注为 `EActivityId` 等枚举的列，实际值是字符串（如 `"Activity_LimitTimeSkin"`），不能当数字解析。保留原始字符串用于匹配。 |
| 默认值 -1 | `strconv.Atoi` 失败时默认赋 -1，需在关联条件中处理（如 `OnceDropRule` 和 `TenDropRule` 互为备选）。 |
