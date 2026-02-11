# reflectparam

Go语言CSV解析库对比与反射参数处理学习项目。

## 项目目的

- 对比三个主流Go CSV库的使用方式和性能特点
- 学习使用Go反射机制动态处理结构体
- 提供实用的CSV解析示例代码

## 支持的CSV库

| 库 | 特点 | 适用场景 |
|---|---|---|
| [why2go/csv_parser](https://github.com/why2go/csv_parser) | 支持复杂嵌套、channel流式处理、详细错误信息 | 需要解析复杂嵌套结构 |
| [jszwec/csvutil](https://github.com/jszwec/csvutil) | 简洁API、高性能、泛型支持 | 追求性能和简洁性 |
| [foolin/gocsv](https://github.com/foolin/gocsv) | 功能丰富、流式处理 | 批量数据处理 |

## 安装

```bash
go get github.com/wangtengda0310/gobee/reflectparam
```

## 反射参数处理

使用反射动态解析字符串到结构体：

```go
import "github.com/wangtengda0310/gobee/reflectparam"

// 基本解析
args := reflectparam.Parse("hello 123")
// 结果: Args{S: "hello", V: 123}

// 泛型解析
result := reflectparam.ParseGeneric("world 456", func() Args { return Args{} })
// 结果: Args{S: "world", V: 456}
```

## CSV解析示例

### csv_parser (支持复杂嵌套)

```go
r := csv.NewReader(bytes.NewBufferString(data))
parser, _ := csv_parser.NewCsvParser[Demo](r)
defer parser.Close()

for row := range parser.DataChan(ctx) {
    fmt.Println(row.Data)
}
```

支持语法：
- `{{attri:age}}` - 映射到 map
- `[[msg]]` - 映射到数组

### csvutil (高性能)

```go
reader := csv.NewReader(file)
decoder, _ := csvutil.NewDecoder(reader)

var person Person
for decoder.Decode(&person) == nil {
    fmt.Println(person)
}
```

### gocsv (文件操作)

```go
var people []Person
gocsv.UnmarshalFile("data.csv", &people)
```

## 运行示例

```bash
# 运行CSV库对比
go run ./example/compare.go

# 运行主程序示例
go run ./main/main.go

# 运行测试
go test -v ./...
```

## 依赖

- Go 1.24.2+
- github.com/stretchr/testify
- github.com/why2go/csv_parser
- github.com/jszwec/csvutil
- github.com/foolin/gocsv

## License

MIT
