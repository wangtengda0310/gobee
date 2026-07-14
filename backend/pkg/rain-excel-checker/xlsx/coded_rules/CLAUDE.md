# coded_rules 目录文档

Excel 配表校验规则实现模块。

## 三层结构组织

| 目录 | 说明 | 示例规则 |
|------|------|----------|
| [cross_table/](./cross_table/) | 跨表级别规则（按业务域分子目录） | hero/activity/draw/drop/* |
| [table/](./table/) | 表级别规则 | 竞技场赛季检查、战令武将开放时间检查 |
| [general/](./general/) | 通用规则 | 新增行/列通知、行变更通知、列级通用检查 |
| [general/column_check/](./general/column_check/) | 列级通用检查（按业务域分子目录） | base/datatype/business/* |

## 规则类型定义位置

所有规则类型常量定义在 `../json_rule/` 目录中。

## 新增规则流程

### 1. 确定规则类型

根据规则影响范围选择目录：
- **跨表关联** → `cross_table/`
- **单表业务逻辑** → `table/`
- **通用列检查** → `general/column_check/` 的对应子目录（base/datatype/business/*）
- **通用表检查** → `general/`

### 2. 创建规则文件

规则文件命名规范：`excel_check_{rule_type}.go`

实现模板参见各子目录下现有规则文件。核心要点：
- 实现对应接口（`Checker` 或 `TableChecker`）
- `Meta()` 方法设置正确的 `TargetSheets`
- 依赖表加载失败时返回错误而非 panic

### 3. 注册规则

- **列级规则**：在 `engine/column_registry.go` 的 `init()` 中注册（import 对应子包）
- **表级规则**：在 `check_manager/table_check_manager.go` 注册

### 规则注释规范

每个校验规则文件顶部必须包含完善的注释，说明以下内容：

1. **规则作用**：当前校验规则检查什么、解决什么问题
2. **前端展示**：对应前端 `excel-rules-template.ts` 中的 `EColRule` 枚举值及前端组件文件
3. **校验成功示例**：什么样的数据应该通过检查
4. **校验失败示例**：什么样的数据应该报错
5. **参数说明**：每个参数的含义和默认值

示例结构：

```go
// YourCheckRule 规则名称
//
// 作用：说明规则检查的业务目的
// 前端对应：EColRule.XXX — 前端显示名（参数组件：XxxParams.vue）
//
// 校验成功示例：
//   - 示例1
//   - 示例2
//
// 校验失败示例：
//   - 示例1
//   - 示例2
//
// 参数说明：
//   - allowEmpty: 是否允许空值
//   - allowCommit: 是否允许提交时存在错误
```

### 4. 编写单元测试

测试文件命名：`{rule_type}_check_test.go`，与源文件放在**同一目录**。

**测试必须包含：**
- 正常通过场景(ShouldPass)
- 应该警告的场景(ShouldWarn)
- 应该失败的场景(ShouldFail)
- 边界条件(缺少列、空数据等)
- **测试用例中需要以注释形式说明被测校验规则的业务含义**，包括：规则检查的业务目的、正常通过示例、应该失败的示例

**测试数据结构**：
- `cols` 是按列存储的二维数组，每个元素是一列的所有行数据
- `MJS_FIXED_ROWS_NAME = 2`，列名在第 3 行(索引 2)
- `MJS_FIXED_ROWS_NUM = 4`，数据从第 5 行(索引 4)开始

```go
// 示例：构造测试数据
cols = [][]string{
    {"", "", "Id", "", "1", "2"},           // Id 列
    {"", "", "Name", "", "数据1", "数据2"},  // Name 列
}
```

**sheetMap 构造要点**：
- 跨表校验规则必须构造完整的 `sheetMap`(`map[string]*excelize.File`，键为 Sheet 名称)
- 使用 `excelize.NewFile()` 创建真实的 Excel 文件对象
- `FindSheetBySuffix` 支持后缀匹配查找(如 "Hero" 匹配 "武将|Hero")

**严禁虚假测试：**
- 测试数据必须能触发被测代码的核心逻辑
- 禁止写"假装测试"的代码(只返回空数据然后标记测试通过)
- 每个规则必须有至少一个校验不通过的测试用例和至少一个校验通过的测试用例

**运行测试**：
```bash
go test ./xlsx/coded_rules_test/... -v -run TestXxxRule
```

## 注意事项

- 规则文件必须放在正确的子目录中
- 测试文件与规则文件放在同一目录
- 导入路径使用 `git.devcloud.ztgame.com/v-tangfangda/rain-excel-checker/xlsx/`
- 飞书消息格式见 formatFieldChangeReason方法
- **ArenaScoreRewards.xlsx 的 sheet 英文后缀是 `ArenaScoreReward`(单数)，FindSheetBySuffix 必须使用单数**

### ⚠️ TargetSheets 挂载原则

规则的 `TargetSheets` 决定增量模式下哪些表变更会触发该规则。**必须挂载在核心检查字段所在的表上**，否则增量检查不会触发。

**判断方法**：规则检查的核心字段在哪个表，就挂载到哪个表。跨表读取的辅助数据放入 `RequiredSheets`。

| 场景 | 正确挂载 | 错误挂载 | 原因 |
|------|---------|---------|------|
| 检查 Item.IsSynthetic | **Item** | ~~Hero~~ | 核心字段在 Item，改 Item 应触发 |
| 检查 Hero.CanMelt | **Hero** | Item | 核心字段在 Hero |
| 检查 Hero.OpenDate vs 赛季时间 | **Hero** | ArenaSeason | 核心字段在 Hero |
| 检查 ArenaScoreReward.Reward 引用 | **ArenaScoreReward** | Hero | 核心数据在 ArenaScoreReward |

**历史教训**(2026-04-14 Bug 1)：
`HERO_SYNTHESIS_CHECK` 最初挂在 Hero 表，但核心字段 `IsSynthetic` 在 Item 表。导致策划修改 Item 表合成字段后增量检查不触发，线上遗漏配置错误。已迁移到 Item 表。

**历史教训**(2026-05-13 Bug 2)：
`SEASON_PASS_HERO_OPEN_CHECK` 和 `ARENA_GENERAL_HERO_OPEN_CHECK` 的 `Meta()` 中缺少 `RequiredSheets` 声明，导致以下场景执行检查时依赖表被过滤：

| 执行路径 | 是否受影响 | 原因 |
|---------|-----------|------|
| 右键"执行分类" | **是** | `filterSheetMapByRules` 按 `RequiredSheets` 过滤 sheetMap，依赖表被排除 |
| 悬浮按钮执行检查 | 否 | 单列检查 `GetSheetMapForColumnCheck` 按列规则关联表按需加载 |
| 菜单"执行检查" | 否 | `preloadedRules == nil`，不走 `filterSheetMapByRules` |
| `-mode=full` CLI | 否 | 同上，全量加载 |
| git 提交模式 | **是** | `collectRequiredSheets` 按 `RequiredSheets` 补充关联表，缺失时无法补充 |

**修复**：在 `Meta()` 中显式声明 `RequiredSheets`：
- `SEASON_PASS_HERO_OPEN_CHECK`: `RequiredSheets: []string{"Hero", "SeasonPass"}`
- `ARENA_GENERAL_HERO_OPEN_CHECK`: `RequiredSheets: []string{"Hero", "ArenaSeason"}`
