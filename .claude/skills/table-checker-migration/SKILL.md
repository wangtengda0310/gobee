---
name: table-checker-migration
description: 将 TableChecker 实现从 Check 方法迁移到 CheckWithParam 方法。注意：此 skill 仅适用于 rain-excel-checker 项目，完成当前迁移工作后应迁移到项目级 .claude/skills/ 目录下
---

# TableChecker 方法迁移

## 目标

将 TableChecker 接口的实现从旧的 `Check` 方法迁移到新的 `CheckWithParam` 方法。

## 背景

`TableChecker` 接口已更新，新增了 `CheckWithParam(param CheckParam)` 方法来替代原来的多参数 `Check` 方法。

## 执行步骤

### 1. 定位目标文件

根据传入的参数 `struct_name` 和 `file_path`，找到需要修改的 Go 文件。

### 2. 添加 CheckWithParam 方法

在现有的 `Check` 方法后添加新方法 `CheckWithParam`：

```go
func (c *YourStructName) CheckWithParam(param check_manager.CheckParam) *check_rule_def.TableCheckResult {
    return c.Check(param.SheetName, param.Cols, param.StartRowIdx, param.Params, param.SheetMap)
}
```

**重要**：
- 新方法应该调用原有的 `Check` 方法
- 参数从 `param` 结构体中提取
- 保持返回值不变

### 3. 验证修改

- 确保代码编译通过
- 确保新方法的签名与接口定义完全匹配

### 4. 报告完成

- 确认已添加 `CheckWithParam` 方法
- 确认代码编译无误

## 参数

调用此 skill 时需要提供：
- `struct_name`: 要迁移的结构体名称
- `file_path`: 包含该结构体的文件路径

## 示例

**原始代码**：
```go
func (c *HeroDropCheckRule) Check(sheetName string, cols [][]string, startRowIdx int, params map[string]string, sheetMap map[string]*excelize.File) *check_rule_def.TableCheckResult {
    // ... 原有逻辑
}
```

**添加后**：
```go
func (c *HeroDropCheckRule) Check(sheetName string, cols [][]string, startRowIdx int, params map[string]string, sheetMap map[string]*excelize.File) *check_rule_def.TableCheckResult {
    // ... 原有逻辑
}

func (c *HeroDropCheckRule) CheckWithParam(param check_manager.CheckParam) *check_rule_def.TableCheckResult {
    return c.Check(param.SheetName, param.Cols, param.StartRowIdx, param.Params, param.SheetMap)
}
```

## 注意事项

- 不要修改原有的 `Check` 方法逻辑
- `CheckWithParam` 方法应该简单委托调用 `Check` 方法
- 确保导入 `check_manager` 包（如果还没有导入）
