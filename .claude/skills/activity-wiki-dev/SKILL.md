---
name: activity-wiki-dev
description: |
  活动Wiki开发技能。当用户需要为活动Wiki页面增加新的活动类型、添加新的关联配表展示面板、
  扩展ActivityCompleteData数据结构、或修改活动Wiki前端展示时使用此技能。
  触发关键词：活动Wiki、activity wiki、ActivityCompleteData、关联表、页签、tab、
  活动类型、ActType、展示面板、丹青阁、商店配置、皮肤抽奖、掉落规则。
  只要用户提到为活动Wiki增加功能、修改活动展示、或添加新的配表关联，就应该使用此技能。
---

<!--
技能标识: activity-wiki-dev
用途: 指导活动Wiki页面的开发，包括新增活动类型、关联配表展示面板等
版本追踪: 本技能生成的代码应包含技能标识注释，便于后续升级时识别和迁移
-->

# 活动Wiki开发技能

## 概述

活动Wiki页面用于展示游戏中各类活动的完整配置信息。每个活动类型（如丹青阁`ActTypeSkinRaffle`）
关联多个配表（如DrawSkin、DropRule、Shop等），在前端以页签(Tab)形式展示。

## 核心架构

### 项目结构

活动Wiki功能涉及多个代码库，路径因开发环境而异。使用前应通过 `project-path-resolver` 技能确认路径：

| 代码库 | 环境变量 | 技术栈 | 职责 |
|--------|----------|--------|------|
| **QA工具（当前项目）** | `PROJECT_QA_TOOLS_PATH` | Go + Wails v3 + Vue 3 | 配表解析、数据关联、前端展示 |
| **游戏客户端** | `PROJECT_CLIENT_PATH` | Unity/C# | 活动逻辑实现、预制体定义 |
| **游戏服务端** | `PROJECT_SERVER_PATH` | 服务端语言 | 活动数据验证、逻辑处理 |

**路径确认**：
```bash
# 检查路径是否已设置
if [ -z "$PROJECT_CLIENT_PATH" ] || [ -z "$PROJECT_SERVER_PATH" ]; then
    # 需要运行 project-path-resolver 技能确认路径
fi
```

### 数据流

```
Excel配表 -> diff_map.go解析 -> DataContainer聚合 -> BuildActivityWikiDiff关联 -> 前端Vue展示
```

### 关键文件位置（QA工具项目内）

| 层级 | 文件路径 | 职责 |
|------|----------|------|
| 表定义 | `rain-resources-checker/mjs_excel/<表名>/def.go` | 定义Excel列索引常量 |
| 表解析 | `rain-resources-checker/mjs_excel/<表名>/diff_map.go` | 解析Excel为Go结构体 |
| 数据聚合 | `rain-resources-checker/activitywiki_def/def.go` | 定义ActivityCompleteData |
| 关联逻辑 | `rain-resources-checker/activitywiki/format.go` | BuildActivityWikiDiff关联各表 |
| 数据容器 | `rain-resources-checker/diff/interface.go` | DataContainer定义 |
| 初始化 | `rain-resources-checker/mjs_excel/diff_excel_init.go` | 初始化所有表解析 |
| 前端展示 | `frontend/src/pages/activity-wiki-check/components/activity-panel.vue` | 页签展示 |

## 如何确定活动关联的配表

在为新活动开发Wiki前，**先查记忆**：

```
1. 调用 mempalace_search 查询：
   - 关键词：ActivityType、CustomParma 使用模式、已确认的关联链路
   - 检查是否有历史"关联模式总结"或"实现详解"文档
2. 调用 claude-mem 查询（通过 /mem search 或查看历史摘要）：
   - 搜索：该活动类型的关联分析记录、配表推导过程
   - 检查近期会话中是否有已确认的 CustomParma 含义或关联模式
3. 调用 forgetful query_memory 查询：
   - 搜索：resources-checker 架构、活动 Wiki 数据模型、关联模式
   - 检查是否有已记录的"关联模式总结"或"数据格式陷阱"
- 如历史记录中已验证过该活动类型的关联逻辑，直接复用
```

然后再确定该活动涉及哪些配表。这是最关键的一步，决定了后续所有开发工作。

### 信息来源（按优先级排序）

**1. 活动表（Activity）的字段分析**

活动表本身提供了最重要的关联线索：

| 字段 | 作用 | 关联方式 |
|------|------|----------|
| `ActivityType` | 活动类型标识（如 `ActTypeSkinRaffle`） | 根据类型决定走哪条关联分支 |
| `CustomParma` | 自定义参数数组（int[]） | 通常 `CustomParma[0]` 是核心关联ID |
| `CustomParma2` | 自定义参数字符串 | 可能包含额外关联信息 |
| `EActivityId` | 枚举标识（如 `Activity_LimitTimeSkin`） | 用于匹配其他表中的枚举字段 |
| `ActivityPrefabType` | 活动预制体类型 | 可提示活动涉及的系统 |

**示例分析（丹青阁）：**
```
ActivityType = "ActTypeSkinRaffle"  -> 这是皮肤抽奖活动
时间匹配 -> findCurrentDrawSkin(sortedDrawSkins) -> 当前进行中的 DrawSkin
EActivityId = "Activity_LimitTimeSkin" -> 关联 LimitSkinTimesReward.ActIdStr
```

**2. 从已有关联推导**

现有活动类型的关联模式可以作为参考：

```go
// 丹青阁的关联链（使用时间匹配确定当前期 DrawSkin）：
Activity -> findCurrentDrawSkin(sortedDrawSkins)
  -> DrawSkin.OnceDropRule -> DropRule(Id)
    -> DropRule.DropGroup[] -> DropGroup(Id)
      -> DropItem.DropGroup -> DropItem
  -> DrawSkin.BigAwardItemId -> ItemHeroSkin(SkinItemId)
    -> ItemHeroSkin.SkinItemId -> HeroSkinItem(SkinItemId)
    -> ItemHeroSkin.SkinItemId -> HeroSkinSpine(SkinItemId)
    -> HeroSkinItem.CollitionType -> HeroSkinCollition(Type)
  -> Activity.EActivityId -> LimitSkinTimesReward.ActIdStr
  -> 固定 ShopType "ShopTypeSkinRaffle" -> Shop + ShopGoods
```

**3. 询问策划/程序同事**

如果无法从表中直接推导，需要向了解该活动的人确认：
- 这个活动涉及哪些系统？（抽奖？商店？任务？）
- CustomParma 字段分别代表什么？
- 活动数据存储在哪些表中？

**4. 查看客户端和服务端代码**

`ActivityPrefabType` 字段对应客户端的活动预制体，通过搜索客户端代码可以了解活动涉及的功能模块。

**前置条件**：确保路径已通过 `project-path-resolver` 技能确认。

搜索方式：
```bash
CLIENT_PATH=${PROJECT_CLIENT_PATH:-""}
SERVER_PATH=${PROJECT_SERVER_PATH:-""}

# 如果路径未设置，提示用户
if [ -z "$CLIENT_PATH" ] || [ -z "$SERVER_PATH" ]; then
    echo "路径未设置，请先运行 project-path-resolver 技能确认项目路径"
    exit 1
fi

# 在客户端代码中搜索ActivityPrefabType对应的预制体
cd "$CLIENT_PATH"
grep -r "ActivityPrefabType" --include="*.cs" .
# 或搜索具体的活动类型
grep -r "ActTypeSkinRaffle" --include="*.cs" .

# 在服务端代码中搜索活动类型
cd "$SERVER_PATH"
grep -r "ActTypeSkinRaffle" --include="*.go" --include="*.cs" --include="*.java" .
```

### 关联模式总结

从代码分析，目前存在以下几种关联模式：

| 模式 | 说明 | 示例 |
|------|------|------|
| **ID直接关联** | Activity.CustomParma[0] -> 目标表.Id | DropRule |
| **枚举字符串匹配** | Activity.EActivityId -> 目标表.ActIdStr | LimitSkinTimesReward |
| **类型枚举匹配** | 固定枚举值 -> 目标表.Type字段 | Shop(ShopType), HeroSkinCollition(Type) |
| **链式关联** | A表 -> B表 -> C表 | DrawSkin -> DropRule -> DropGroup -> DropItem |
| **反向关联** | 目标表字段指向Activity | DrawPet.ActivityId -> Activity.Id |
| **时间匹配** | 按StartTime/EndTime匹配当前期，无进行中取最新一期 | DrawSkin, DrawPet, SeasonPass |
| **独立子系统** | 无ActivityType，独立索引+时间匹配 | SeasonPass（战令系统） |

### 独立子系统关联模式详解

战令系统（SeasonPass）是一个典型的**独立子系统**，与Activity表是平行关系：

**特征**：
- 没有对应ActivityType（Activity表中无战令类型）
- 数据通过独立字段在ActivityWikiDiff中存储（`SeasonPasses` map）
- 关联逻辑独立于Activity遍历循环

**实现方式**：
```go
// 1. 在ActivityWikiDiff中新增独立字段
type ActivityWikiDiff struct {
    Activities   map[int]*ActivityCompleteData   // 按Activity.Id索引
    SeasonPasses map[int]*SeasonPassCompleteData // 按SeasonPass.Id索引（新增）
}

// 2. 在BuildActivityWikiDiff中，Activity遍历结束后独立构建战令数据
// 使用时间匹配确定当前期，再获取上一期/下一期
sortedSeasonPasses := buildAllSortedSeasonPasses(container.SeasonPassDiff)
currentSeasonPass := findCurrentSeasonPass(sortedSeasonPasses)
if currentSeasonPass != nil {
    prev, next := findPrevNextSeasonPass(sortedSeasonPasses, currentSeasonPass.Id)
    seasonPassData := &SeasonPassCompleteData{Basic: currentSeasonPass}
    if prev != nil { seasonPassData.PrevSeasonPass = prev }
    if next != nil { seasonPassData.NextSeasonPass = next }
    // 关联礼包、奖励、任务...
}
```

**时间匹配模式详解**：

对于有多期数据的活动（丹青阁、战令、结缘亭），使用以下模式确定当前期和上下期：

1. **排序**：`buildAllSortedXxx(diff)` — 按StartTime升序排列全表
2. **匹配当前期**：`findCurrentXxx(sorted)` — 优先匹配 StartTime <= now <= EndTime 的记录，无则取最新一期
3. **获取上下期**：`findPrevNextXxx(sorted, currentId)` — 在排序切片中取当前的前一个和后一个
4. **构建三期数据**：将 Prev/Current/Next 放入 CompleteData，前端使用 PeriodCard 子组件展示

**前端展示**：
- 在index.vue中独立遍历`activityWikiDiff.SeasonPasses`
- 使用独立组件（SeasonPassPanel）展示，基础信息页签内用 PeriodCard 子组件展示三期数据
- 不混入ActivityPanel的页签体系

**独立子系统完整适配步骤**：

开发新的独立子系统（如竞技场赛季、公会系统等）时，按以下步骤执行：

**后端**：
1. 新增配表解析模块 `mjs_excel/<表名>/def.go` + `diff_map.go`（主表和所有子表）
2. 注册到 DataContainer：`diff/interface.go` 新增字段 + `diff_excel_init.go` 初始化解析
3. 定义 XxxCompleteData：`activitywiki_def/def.go`，包含主表指针 + 子表切片 + 三期字段（PrevXxx/Basic/NextXxx）
4. 在 ActivityWikiDiff 中新增独立 map 字段（如 `XxxSystems map[int]*XxxCompleteData`）
5. 在 BuildActivityWikiDiff 的 Activity 遍历循环**之后**独立构建：
   - 排序全表 → 时间匹配找当前期 → findPrevNext 获取上下期 → 关联子表 → 写入 wiki

**前端**：
6. 新增 XxxPanel.vue（参考 SeasonPassPanel），包含页签：基础信息 + 各子表页签
7. 新增 XxxPeriodCard.vue（参考 SeasonPassPeriodCard），支持 highlight/periodLabel props
8. index.vue 中独立遍历 + 添加锚点导航节点
9. 基础信息页签添加关联关系链可视化（n-steps，参考 activity-panel.vue:293-316）

---

## 使用Subagent分析活动关联配表

当需要为新活动确定关联配表时，可以创建专门的分析任务分配给subagent执行。

### 分析任务流程

```
主Agent创建任务 -> Subagent执行分析 -> 返回关联配表报告 -> 主Agent根据报告开发
```

### Subagent分析任务模板

**任务名称**：分析 `{活动类型}` 关联配表

**任务描述**：
1. 读取 `rain-resources-checker/mjs_excel/activity/def.go` 和 `diff_map.go`，了解Activity表结构
2. 读取 `rain-resources-checker/activitywiki/format.go`，了解现有关联模式
3. 列出 `rain-resources-checker/mjs_excel/` 下所有已实现的配表模块
4. 分析目标活动类型的关联逻辑：
   - 如果Activity表中有该活动类型的数据，分析CustomParma字段值
   - 根据ActivityType和CustomParma推导可能关联的表
   - 检查其他表中是否有字段可能关联到该活动
   - 如需进一步了解活动功能，可搜索客户端代码（`$PROJECT_CLIENT_PATH`）和服务端代码（`$PROJECT_SERVER_PATH`）中该ActivityType的使用情况。路径未设置时需先运行 `project-path-resolver` 技能
5. 输出报告：
   - 该活动类型应关联哪些配表
   - 每张表的关联条件（通过什么字段关联）
   - 关联模式（ID直接关联/枚举匹配/链式关联等）
   - 哪些表已存在解析模块，哪些需要新建

**Subagent返回格式**：

```markdown
## 活动类型：{ActTypeXxx} 关联配表分析报告

### 核心关联
| 配表 | Sheet名 | 关联条件 | 关联模式 | 是否已存在 |
|------|---------|----------|----------|------------|
| Xxx | 表名|Xxx | CustomParma[0] -> Id | ID直接关联 | 是/否 |

### 链式关联
```
Activity -> Xxx -> Yyy -> Zzz
```

### 需要新建模块
- [ ] xxx/def.go + diff_map.go

### 建议前端页签
1. 基础信息（Activity）
2. Xxx配置（表名|Xxx）
```

### 主Agent使用分析结果的流程

1. **接收报告**：读取subagent返回的关联配表报告
2. **确认关联**：检查报告中的关联逻辑是否合理，必要时询问用户确认
3. **分配开发任务**：根据报告中的"需要新建模块"列表，为每个新表创建subagent开发任务
4. **并行开发**：
   - Subagent A：开发新表的def.go + diff_map.go
   - Subagent B：注册到DataContainer和初始化
   - Subagent C：添加到ActivityCompleteData和关联逻辑
   - Subagent D：开发前端页签
5. **集成验证**：所有subagent完成后，主Agent集成并验证

---

> 代码模板（Go/Vue注释格式、def.go模板、diff_map.go模板、DataContainer注册模板、ActivityCompleteData模板、前端页签模板、打开Excel模板、步骤5支持新活动类型）详见 [references/code-templates.md](references/code-templates.md)

> 前端页面设计规范（页面布局结构、组件使用规范、页签内容设计模式、关联关系展示规范、样式规范）详见 [references/design-patterns.md](references/design-patterns.md)

> 常见陷阱与调试、数据过滤与关联范围控制、现有活动类型和关联表 详见 [references/troubleshooting.md](references/troubleshooting.md)

---

## 开发前检查清单（必须逐项确认）

在开始编码前，必须完成以下检查：

### 1. 数据验证

**目标**：确认Excel中关键关联字段有实际数据，避免代码正确但数据为空导致页签不显示。

**验证脚本模板**（Python）：
```python
import openpyxl

# 验证Activity表中目标活动的CustomParma
wb = openpyxl.load_workbook('Activity_活动表.xlsx')
sheet = wb['活动表|Activity']
for row in sheet.iter_rows(min_row=5, values_only=True):
    if row[2] == 'ActTypeXxx':  # ActivityType列
        print(f"Activity Id={row[0]}, CustomParma={row[13]}, CustomParma2={row[14]}")

# 验证目标表的关联字段
wb2 = openpyxl.load_workbook('Xxx_表名.xlsx')
sheet2 = wb2['表名|Xxx']
for row in sheet2.iter_rows(min_row=5, values_only=True):
    print(f"Xxx Id={row[0]}, ActivityId={row[10]}")  # 根据实际列索引调整
```

**必须验证的字段**：
- [ ] Activity表中该活动类型的CustomParma是否有值
- [ ] 目标关联表的Id列是否有数据
- [ ] 目标表的关联字段（如ActivityId、ActIdStr等）是否有值
- [ ] 链式关联的下游表（如DropGroup、DropItem）是否有数据

### 2. 列索引与数据格式验证

**目标**：确认def.go中的列索引与实际Excel列匹配，数据格式符合预期。

**验证方法**：
```python
import openpyxl

wb = openpyxl.load_workbook('Xxx_表名.xlsx')
sheet = wb['表名|Xxx']

# 打印第3行（字段名）验证列索引
row3 = list(sheet.iter_rows(min_row=3, max_row=3, values_only=True))[0]
for i, name in enumerate(row3):
    if name:
        print(f"Col {i}: {name}")

# 打印第5行（第一条数据）验证数据格式
row5 = list(sheet.iter_rows(min_row=5, max_row=5, values_only=True))[0]
print(f"First data row: {row5}")
```

**常见数据格式陷阱**：
| 字段类型 | Excel实际格式 | 常见误判 | 正确处理方式 |
|----------|--------------|----------|-------------|
| 整数 | `31` | 无 | `strconv.Atoi` |
| 整数数组 | `1,2,3` | 无 | `strings.Split` + `strconv.Atoi` |
| 花括号格式 | `{1039016;1}` | 简单整数 | 正则提取 `{(\d+);\d+}` |
| 枚举字符串 | `Activity_LimitTimeSkin` | 整数 | 保留字符串匹配 |
| 布尔值 | `0`/`1` 或 `true`/`false` | 整数 | `strconv.ParseBool` |

**必须确认**：
- [ ] def.go中的iota常量顺序与Excel实际列顺序一致
- [ ] 数据格式与解析代码匹配（不要假设，要验证）
- [ ] 空值处理逻辑正确（空字符串、None、0的区别）

### 3. 关联模式确认

**必须确认该活动使用哪种关联模式**：

| 模式 | 特征 | 是否需要备选方案 |
|------|------|-----------------|
| ID直接关联 | Activity.CustomParma[0] -> 目标表.Id | 是（CustomParma可能为空） |
| 枚举字符串匹配 | Activity.EActivityId -> 目标表.ActIdStr | 否 |
| 类型枚举匹配 | 固定枚举值 -> 目标表.Type字段 | 否 |
| 链式关联 | A表 -> B表 -> C表 | 是（中间环节可能断裂） |
| **反向关联** | 目标表.ActivityId -> Activity.Id | **必须考虑** |
| **时间匹配** | 按StartTime/EndTime匹配当前期，无则取最新 | 是（全表为空时返回nil） |

**反向关联代码模板**：
```go
// 当CustomParma为空时，通过目标表的ActivityId反向查找
var dp *draw_pet.DrawPetDiff
if len(act.CustomParma) > 0 {
    if found, ok := drawPetById[act.CustomParma[0]]; ok {
        dp = found
    }
} else {
    // 反向关联：通过ActivityId查找
    if found, ok := drawPetByActivityId[act.Id]; ok {
        dp = found
    }
}
```

### 4. 条件分支结构检查

**目标**：避免嵌套if导致逻辑被跳过。

**正确结构**：
```go
if act.ActivityType == "ActTypeA" {
    // A类型逻辑
} else if act.ActivityType == "ActTypeB" {
    // B类型逻辑
} else if act.ActivityType == "ActTypeC" {
    // C类型逻辑
}
```

**错误结构（嵌套陷阱）**：
```go
if act.ActivityType == "ActTypeA" {
    // A类型逻辑
    if act.ActivityType == "ActTypeB" {  // 永远不会执行！
        // B类型逻辑
    }
}
```

---


## 规则覆盖角标

活动Wiki页面在Tab标题和字段旁显示规则覆盖角标，让用户直观看到哪些表/字段被规则覆盖。

### 角标规则

| 位置 | 角标内容 | 颜色 |
|------|----------|------|
| Tab标题右侧 | 表级规则数量 | 默认绿色，校验失败变红 |
| 字段label旁 | 列级规则+行级规则数量 | 默认绿色，校验失败变红 |
| 无规则覆盖 | 不显示角标 | — |

### 颜色逻辑
- 默认绿色（有规则覆盖，尚未执行检查或全部通过）
- 收到校验失败结果 → 变红色
- 行级规则在每个涉及字段上各计1

### 后端数据结构
- `RuleCoverageData` — 规则覆盖数据（按Sheet名索引）
- `SheetRuleStats` — 单个Sheet的表级规则数和字段规则统计
- `FieldRuleStat` — 字段级规则统计（列级+行级，行级预留当前为0）

### 后端方法
- `ActivityWikiCheckService.GetRuleCoverage(caseDir string)` — 获取规则覆盖统计

### 前端渲染函数
- `renderTabWithBadge(tabName, sheetName, label)` — Tab标题+表级规则角标
- `renderFieldWithBadge(fieldName, sheetName, label)` — 字段label+字段规则角标

---

## 测试规范

新增活动类型开发完成后，必须执行以下三层测试。测试不通过不应提交代码。

### 第1层：单元测试

为后端关联逻辑编写单元测试，验证数据解析和关联正确性。

**测试范围**：
1. 新增配表解析模块的 `GetXxxDiffMap` 函数
2. `BuildActivityWikiDiff` 中新活动类型的关联分支
3. 反向关联逻辑（CustomParma 为空时的备选路径）

**测试规范**：
- 使用 `testify/assert` 库，不使用 `t.Log` 系列
- 测试文件与代码文件放在同一目录下
- 测试函数必须包含业务语言注释

**示例**：

```go
// TestBuildActivityWikiDiff_DrawPet 测试结缘庭活动的关联逻辑
// 业务场景：Activity.CustomParma[0] 指向 DrawPet.Id，应正确关联抽奖配置
func TestBuildActivityWikiDiff_DrawPet(t *testing.T) {
    // 构建测试数据...
    result := activitywiki.BuildActivityWikiDiff(container)
    act := result.Activities[31]
    assert.NotNil(t, act.DrawPet)
    assert.Equal(t, 1, act.DrawPet.Id)
}
```

### 第2层：代码审核（subagent）

功能开发完成后，使用 code-reviewer subagent 对变更进行审核。

**审核范围**：
1. 后端：关联逻辑是否正确、空值保护是否完整、Sheet 名是否与 Excel 一致
2. 前端：组件是否按规范使用、条件渲染是否正确、TypeScript 类型是否匹配
3. 公共影响：是否误改了其他活动类型的逻辑、新增字段是否注册完整

**触发方式**：开发完成后主动启动 code-reviewer subagent。

### 第3层：Playwright E2E 测试（subagent）

使用 Playwright 自动化测试验证前端功能。

**测试范围**：
1. 页面加载：执行检查后页面是否正常渲染
2. 活动卡片：新增活动类型的卡片是否显示、页签是否出现
3. 数据展示：关联数据是否正确渲染（标签、表格、描述）
4. 交互功能：页签切换、打开Excel按钮、筛选条件

**触发方式**：代码审核通过后，使用 playwright subagent 编写并运行 E2E 测试。

**测试文件位置**：`frontend/e2e/activity-wiki/activity-wiki.spec.ts`

---

## 文档同步规范

开发完成后必须同步更新以下文档，不遗漏任何一项。

### 1. 代码注释

- 对外暴露的方法添加中文注释（用途、参数说明）
- 超过20行的方法补充流程视角的注释
- 修改代码时同步修改已有注释

### 2. CLAUDE.md 层级文档

根据变更范围更新对应的 CLAUDE.md：

| 变更范围 | 需更新的文档 |
|----------|-------------|
| 新增配表解析模块 | `mjs_excel/CLAUDE.md`（如存在）、`rain-resources-checker/CLAUDE.md` |
| 修改 ActivityCompleteData | `activitywiki_def/CLAUDE.md`（如存在）、`activity-wiki-check/CLAUDE.md` |
| 修改 BuildActivityWikiDiff | `activitywiki/CLAUDE.md`（如存在） |
| 新增 MCP 工具 | `docs/MCP-USAGE.md` |
| 新增公共方法 | 对应包的 CLAUDE.md |

### 3. 前端布局文档

使用 `frontend-layout-docs` 技能更新布局文档：
- 新增页面：生成完整的布局文档（`frontend/docs/layout/pages/activity-wiki-check/`）
- 修改页面结构：更新对应的布局文档

### 4. 功能实现详解

如果新增活动类型的关联逻辑较复杂（涉及3张以上配表或链式关联），在 `docs/` 下编写实现详解文档，内容包括：
- 关联链路图
- 关键字段说明
- 反向关联等特殊逻辑

---

## 注意事项

1. **Sheet名格式**：必须使用 `中文名|英文名` 格式，与Excel实际sheet名一致
2. **关联条件**：通常通过 `Activity.CustomParma` 数组或 `Activity.EActivityId` 枚举字符串关联
3. **索引构建**：为每个关联表构建合适的索引（按Id、按Type、按ActIdStr等）
4. **空值处理**：索引构建函数中必须检查 `nil`，避免panic
5. **前端条件渲染**：使用 `v-if` 确保只在数据存在时显示页签
6. **类型转换**：Excel读取的数据都是字符串，需要按需转换为int、bool、数组等
7. **数据过滤**：全量共享表必须通过ActivityId/ActIdStr/链式关联等方式过滤，只展示当前活动相关数据。详见 [references/troubleshooting.md](references/troubleshooting.md)
8. **测试先行**：开发完成后必须通过三层测试（单元测试 -> 代码审核 -> E2E测试）才能提交
9. **文档同步**：代码变更必须同步更新注释、CLAUDE.md、布局文档
10. **独立子系统适配**：如战令系统（SeasonPass）没有对应ActivityType，需要在ActivityWikiDiff中新增独立字段（如`SeasonPasses`），在前端独立展示（不混入ActivityPanel）
