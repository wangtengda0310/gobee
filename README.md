# CSV Unmarshaler

一个使用TDD（测试驱动开发）方法实现的CSV解析库，演示了李氏替换原则。

## 项目结构

```
csv/
├── go.mod                 # Go模块定义
├── interface.go           # Unmarshaler接口定义
├── test_common.go         # 通用测试函数
├── liskov_test.go         # 李氏替换原则综合测试
├── gocsv/                 # github.com/gocarina/gocsv实现
│   ├── gocsv.go
│   └── gocsv_test.go
├── csvutil/               # github.com/jszwec/csvutil实现
│   ├── csvutil.go
│   └── csvutil_test.go
└── encoding/              # encoding/csv标准库实现
    ├── encoding.go
    └── encoding_test.go
```

## 接口定义

```go
type Unmarshaler interface {
    Unmarshal(data []byte, v interface{}) error
}
```

## 实现方式

本项目提供了三种不同的CSV解析实现：

1. **gocsv**: 使用 `github.com/gocarina/gocsv` 库
2. **csvutil**: 使用 `github.com/jszwec/csvutil` 库
3. **encoding**: 使用标准库 `encoding/csv`

所有实现都遵循相同的接口，可以根据李氏替换原则互相替换。

## 使用示例

```go
package main

import (
    "csv-unmarshaler"
    "csv-unmarshaler/gocsv"
    "csv-unmarshaler/csvutil"
    "csv-unmarshaler/encoding"
    "fmt"
)

type Person struct {
    Name string `csv:"name"`
    Age  int    `csv:"age"`
}

func main() {
    csvData := `name,age
John,30
Alice,25`

    // 使用不同的实现，它们都可以互相替换
    implementations := []csvunmarshaler.Unmarshaler{
        csvunmarshaler.UnmarshalerFunc(gocsv.NewGoCSVUnmarshaler().Unmarshal),
        csvunmarshaler.UnmarshalerFunc(csvutil.NewCSVUtilUnmarshaler().Unmarshal),
        csvunmarshaler.UnmarshalerFunc(encoding.NewEncodingCSVUnmarshaler().Unmarshal),
    }

    for i, unmarshaler := range implementations {
        var people []Person
        err := unmarshaler.Unmarshal([]byte(csvData), &people)
        if err != nil {
            fmt.Printf("Implementation %d failed: %v\n", i+1, err)
            continue
        }
        fmt.Printf("Implementation %d result: %+v\n", i+1, people)
    }
}
```

## 测试

运行所有测试：
```bash
go test ./...
```

运行特定实现的测试：
```bash
go test ./gocsv
go test ./csvutil
go test ./encoding
```

运行李氏替换原则测试：
```bash
go test -run Liskov
```

## TDD开发流程

本项目严格遵循红-绿-重构的TDD开发流程：

1. **红色阶段**: 先编写测试用例，确保测试失败
2. **绿色阶段**: 编写最少代码使测试通过
3. **重构阶段**: 优化代码结构，保持测试通过

每个实现都按照这个流程开发，确保代码质量和测试覆盖率。

## 李氏替换原则

所有CSV解析实现都遵循李氏替换原则：
- 实现相同的 `Unmarshaler` 接口
- 可以互相替换而不影响程序正确性
- 具有一致的行为契约

## 性能特点

- **gocsv**: 功能丰富，支持复杂映射
- **csvutil**: 性能优秀，类型安全
- **encoding**: 标准库，简单可靠

选择适合你项目需求的实现。