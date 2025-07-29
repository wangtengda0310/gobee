# 日志 XML/Excel 结构自动生成 Go 代码最佳实践（修正版）

## 1. 需求与输入输出
- **输入**：XML 或 Excel 日志结构文件、类型映射配置（XML）
- **输出**：自动生成的 Go 代码，包括日志函数、参数结构体、基础参数、单元测试、性能测试等，所有内容需健壮、模板化、可扩展
- **要求**：结构和命名需全局唯一，支持多 struct、边界用例、自动测试和 benchmark

## 2. 项目结构与主流程
- 自动创建 `glog-gen-v2` 目录，包含 `go.mod`、`internal`、`templates`、`config`、`output`、`cmd`、`testdata` 等子目录
- `config` 目录存放 `type_mapping.xml`
- `testdata` 目录存放 `demo.xml`（或 Excel）
- `cmd/glog-gen/main.go` 负责命令行参数解析，主流程调用 `internal/generator.go`
- `output` 目录为所有自动生成的 go、test、bench 文件输出地，生成子目录（如 `output/glog`）包名与目录名一致

## 3. 核心模块设计
- `internal/model.go`：定义 Struct、Entry、TypeMapping、TemplateData 等模板数据结构
- `internal/mapping.go`：实现类型映射 XML 解析
- `internal/parser.go`：实现 XML/Excel 解析，支持字段驼峰化、边界处理
- `internal/generator.go`：主流程，包括类型映射校验、模板数据填充、模板渲染、文件输出，支持所有 struct、base_param.go、writer.go 的自动生成
- `internal/template.go`：模板加载，优先本地文件，找不到时 fallback 到 go:embed 内嵌模板

## 4. 模板与自动化
- `templates/func_struct.go.tmpl`：合并结构体和日志函数的模板
- `templates/test.go.tmpl`：单元测试和 fuzz 测试模板
- `templates/base_param.go.tmpl`：基础参数模板，包名动态
- 所有模板渲染均通过 `text/template`，所有字段顺序、类型、注释、唯一命名等均在模板中实现

## 5. 日志写入与依赖注入
- `output/writer.go`：定义 `WriteFunc`、`DefaultWriter`、`SetDefaultWriter`、`NewFileWriter`，每条日志内容单独一行写入
- 生成目录（如 `output/glog`）会自动复制一份 `writer.go`，保证 `DefaultWriter` 可用
- 日志函数通过 `DefaultWriter([]string{s})` 写入日志

## 6. 自动测试与性能测试
- 所有 `*_test.go` 文件自动集成 `gofakeit.Struct(&param)` 填充 mock 数据
- 性能测试统一放在 `output/writer_bench_test.go`，包含单线程和并发写入 benchmark
- 测试和 benchmark 文件命名规范：`<struct>_test.go`、`writer_bench_test.go`
- 运行 `go test ./output`，所有自动生成的测试和 benchmark 均可通过，日志内容断言自动化

## 7. 健壮性与最佳实践
- 生成代码需健壮，处理异常和空值，类型未映射、order缺失、属性缺失等需有报错或默认处理
- 注释需完整，便于维护
- 类型映射需灵活可扩展
- 日志内容顺序严格按 order
- formatter函数、单元测试需覆盖所有生成内容
- 所有生成的类型、方法、formatter函数名需全局唯一，避免命名冲突
- mock数据生成推荐使用 gofakeit
- write函数采用依赖注入，便于mock和性能测试
- 单元测试和性能测试写入临时文件，避免污染主日志

## 8. 模板优先与维护
- 所有生成内容均应优先通过模板维护和修正，避免直接修改单个生成文件
- 只需维护模板和主流程，后续可批量更新所有生成代码

## 9. 常见问题与修复
- 断言内容为空：需传递 AllFieldNames 给 test.go.tmpl
- fmt 未使用：_ = fmt.Sprint 放到测试函数体内
- 字段顺序、类型、注释、唯一命名等问题，均应在模板和主流程统一修正
- 若模板文件缺失，自动 fallback 到 go:embed 内嵌模板

---

如需进一步细化某一部分或补充示例，请告知！
