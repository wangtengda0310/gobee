# tests 目录文档

Excel 检查模块的集成测试。

## 目录结构

```
tests/
├── resources/           # 测试资源(Excel 文件)
│   ├── ArenaScore.xlsx      # 竞技场积分测试表
│   └── Activity_活动表.xlsx  # 活动表测试文件
└── xlsx_test.go         # 集成测试
```

## 测试用例概览

| 测试函数 | 行号 | 说明 | 依赖资源 |
|----------|------|------|----------|
| `TestReadXlsx` | :18 | Excel 读取和格式解析 | resources/ |
| `TestCheckXlsx` | :45 | 列级检查规则 | resources/ |
| `TestSaveCheck` | :121 | 规则配置保存 | - |
| `TestLoadJsonRules` | :185 | 规则配置加载 | - |
| `TestSimpleLoadJsonRules` | :194 | 从 rain-qa-func 加载实际配置 | rain-qa-func/cases/ |
| `TestCleanup` | :227 | 清理测试数据 | - |
| `TestCheckAll` | :235 | 完整检查流程 | resources/ |
| `TestFindEnum` | :285 | 枚举类型查找 | resources/ |

每个测试的详细逻辑请直接查看 `xlsx_test.go` 中对应行号的代码注释。

## 运行测试

```bash
# 运行所有测试
go test ./xlsx/tests/...

# 运行单个测试
go test ./xlsx/tests/ -run TestReadXlsx -v
```

## 添加新测试

1. 在 `resources/` 添加测试 Excel 文件(确保格式符合配表规范)
2. 在 `xlsx_test.go` 中添加测试函数，参考 `TestReadXlsx`(:18)的模板：
   - 使用 `excel_internal.ReadFileOrDir()` 读取资源
   - 使用 `defer` 确保 Excel 文件关闭
   - 使用 `testify/assert` 进行断言

## 依赖

- `check_internal`, `check_manager`, `json_rule`, `excel_internal`
- `github.com/xuri/excelize/v2`, `github.com/google/uuid`

## 注意事项

- `TestSimpleLoadJsonRules` 依赖 rain-qa-func 配置文件，无该文件时会跳过
- 部分测试依赖固定文件路径，移动项目需更新
- 使用 `t.Fatal()` 处理严重错误，`assert` 处理断言
