# reflectparam 项目知识库

## 项目概述
Go语言CSV解析库对比与反射参数处理学习项目。

## 项目架构

```
reflectparam/
├── reflect_param.go              # 核心：反射参数解析功能
├── reflect_param_test.go         # 单元测试（100%覆盖）
├── reflect_param_bench_test.go   # 性能基准测试
├── main/main.go                 # csv_parser 复杂嵌套示例
├── example/compare.go           # CSV库对比示例
└── examples/gocsv/main.go        # gocsv 库说明（含局限性）
```

## CSV库对比

| 库 | 优点 | 缺点 | 适用场景 |
|------|--------|--------|----------|
| **why2go/csv_parser** | 支持复杂嵌套、channel流式、错误详细 | 性能较低 | 复杂数据结构 |
| **jszwec/csvutil** | 高性能、简洁API、泛型支持 | 不支持嵌套 | 大型文件、简单结构 |
| **foolin/gocsv** | 功能丰富 | ⚠️ 2017年停更、有bug | 不推荐使用 |

### 性能对比（基准测试）

```
直接赋值      : 0.27 ns/op  (理论基准)
反射解析      : 70.15 ns/op  (本项目实现)
泛型解析      : 82.51 ns/op  (本项目实现)
标准库CSV    : 142 μs/op
csvutil      : 207 μs/op
csv_parser    : 1301 μs/op
```

## 核心功能

### 1. 反射参数解析

```go
// 基础解析
args := reflectparam.Parse("hello 123")
// → Args{S: "hello", V: 123}

// 泛型解析
result := reflectparam.ParseGeneric("test 456", func() Args { return Args{} })
// → Args{S: "test", V: 456}
```

**实现原理**：
- 使用 `reflect.ValueOf(&result).Elem()` 获取结构体反射值
- 通过 `Field(i)` 访问字段
- 根据字段 `Kind()` 设置对应类型的值

### 2. CSV 解析

#### csv_parser - 复杂语法支持

```csv
name,{{attri:age}},{{attri:height}},[[msg]],[[msg]]
Alice,20,,"Hi, I'm Alice.",Nice to meet you!
```

- `{{attri:age}}` → 映射到 `map[string]int16`
- `[[msg]]` → 映射到 `[]string`

#### csvutil - 高性能简洁

```go
decoder, _ := csvutil.NewDecoder(csvReader)
decoder.Decode(&struct{})
```

#### gocsv - 已过时（2017年最后更新）

**问题**：
- 不支持 `io.Reader`，仅文件路径
- 标签解析存在bug
- 建议使用 csvutil 或 csv_parser 替代

## 测试策略

### 测试覆盖
- **覆盖率**: 100%
- **测试用例数**: 36+
- **测试分类**:
  - 单元测试
  - 边界测试
  - 并发测试
  - 性能测试

### 运行测试
```bash
# 单元测试
go test -v ./...

# 覆盖率
go test -cover ./...

# 基准测试
go test -bench=. -benchmem

# 运行示例
go run ./example/compare.go
go run ./main/main.go
```

## 开发规范

### Go 版本
- 项目使用 Go 1.24.2

### 依赖管理
```bash
go mod tidy  # 清理未使用依赖
go mod xxx@version  # 添加特定版本
```

### 命名规范
- 包名：`reflectparam`
- 测试文件：`xxx_test.go`
- 基准测试：`xxx_bench_test.go`
- 示例文件：使用描述性名称

## 常见问题

### gocsv 库问题
如 `examples/gocsv/main.go` 所示，gocsv 库存在已知问题，已在该文件中添加说明。

### 反射使用
- 获取指针类型的元素：`reflect.ValueOf(&ptr).Elem()`
- 检查字段类型：`field.Kind()` 返回 `reflect.Kind` 枚举

## 待办事项
- [ ] 考虑添加更多CSV格式的支持（如TSV）
- [ ] 优化 csv_parser 性能以接近 csvutil
- [ ] 添加流式处理的更多示例

## 参考链接
- [why2go/csv_parser](https://github.com/why2go/csv_parser)
- [jszwec/csvutil](https://github.com/jszwec/csvutil)
- [foolin/gocsv](https://github.com/foolin/gocsv) - ⚠️ 已停更
