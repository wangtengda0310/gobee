# 回归测试用例索引

本文档提供回归测试的快速索引，详细信息请查看测试代码中的注释。

## 快速查看

```bash
# 查看所有回归测试
grep "^func TestRegression" tests/regression_test.go

# 查看某个回归测试的详细信息
grep -A 30 "TestRegression_Issue1" tests/regression_test.go
```

## 回归测试列表

| Issue | 描述 | 测试函数 | 修复日期 |
|-------|------|----------|----------|
| #1 | 并发加载竞争条件 | `TestRegression_Issue1_ConcurrentLoad` | 2025-02-15 |
| #2 | 空 Sheet 名称 panic | `TestRegression_Issue2_EmptySheetName` | 2025-02-15 |
| #3 | SetMockData 并发竞争 | `TestRegression_Issue3_SetMockData_Concurrent` | 2025-02-15 |
| #4 | 条件字段依赖顺序 | `TestRegression_Issue4_ConditionalField_DependencyOrder` | 2025-02-15 |
| #5 | 空 Mock 数据处理 | `TestRegression_Issue5_MockData_EmptyData` | 2025-02-15 |

## 文档规范

**重要**：回归测试的详细文档直接在测试代码的注释中维护（`tests/regression_test.go`）。

这样做的好处：
- ✅ 文档与代码永远同步
- ✅ 修改测试时自然更新文档
- ✅ 无需额外的 .md 文件维护

## 测试注释模板

添加新的回归测试时，请使用以下模板：

```go
// TestRegression_IssueX_<描述> 测试<简要描述>
//
// 问题描述：<详细描述 bug 的表现>
//
// 复现步骤：
//   <代码示例>
//
// 解决方案：<如何修复的>
//
// 修复日期：YYYY-MM-DD
//
// 相关提交：<commit hash> <commit message>
//
// 测试覆盖：<测试验证了什么>
func TestRegression_IssueX_...(t *testing.T) {
    // 测试代码
}
```

## 运行回归测试

```bash
# 运行所有回归测试
go test -v ./tests/... -run TestRegression

# 运行特定回归测试
go test -v ./tests/... -run TestRegression_Issue1

# 带竞态检测
go test -v ./tests/... -run TestRegression -race

# 查看测试覆盖率
go test ./tests/... -coverprofile=coverage.out -run TestRegression
go tool cover -html=coverage.out
```

---

## 为什么将文档放在测试代码中？

传统做法的问题：
- ❌ 文档（.md）和代码（.go）分离
- ❌ 修改代码时忘记更新文档
- ❌ 两者逐渐不一致

更好的做法：
- ✅ 文档即代码（Documentation as Code）
- ✅ Go 注释就是文档
- ✅ 永远同步，无法分离
