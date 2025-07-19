# 日志 XML 结构自动生成 Go 代码最佳实践提示词

## 1. 需求与输入输出
- 输入：XML 日志结构文件、类型映射配置（XML）
- 输出：自动生成的 Go 代码，包括日志函数、参数结构体、formatter、单元测试、性能测试等，所有内容需健壮、模板化、可扩展
- 结构和命名需全局唯一，支持多 struct、边界用例、自动测试和 benchmark

## 2. 项目结构与主流程
- 自动创建 glog-gen 目录、go.mod、internal、templates、config、output 等子目录
- config 目录存放 demo.xml、type_mapping.xml
- main.go 负责命令行参数解析，主流程调用 internal/generator.go
- output 目录为所有自动生成的 go、test、bench 文件输出地

## 3. 核心模块设计
- internal/model.go：定义 Struct、Entry、TypeMapping、TemplateData 等模板数据结构
- internal/mapping.go：实现类型映射 XML 解析
- internal/parser.go：实现 XML 解析，支持字段驼峰化、边界处理
- internal/generator.go：主流程，包括类型映射校验、模板数据填充、模板渲染、文件输出，支持所有 struct、base_param.go、writer.go 的自动生成

## 4. 模板与自动化
- templates/struct.go.tmpl、func.go.tmpl、test.go.tmpl、bench.go.tmpl、base_param.go.tmpl：实现结构体、日志函数、测试、性能测试、基础参数的模板
- 自动生成所有 struct 的 go、test、bench 文件，支持 gofakeit 自动填充结构体 mock 数据
- writer.go 单独定义 WriteFunc 及相关方法，供所有日志函数文件引用
- 所有模板渲染均通过 text/template，所有字段顺序、类型、注释、唯一命名等均在模板中实现

## 5. 调试与常见问题修复
- 类型未映射、字段名驼峰化不一致、结构体字段遗漏、模板未正确 range、非法字符、未使用 import、字段名与模板引用不一致等问题，均需优先修正模板和主流程
- 断言内容为空问题：需确保 generator.go 渲染 test.go.tmpl 时传递 AllFieldNames，否则模板断言 expected 为空
- fmt 未使用导致的编译错误，需将 _ = fmt.Sprint 放到测试函数体内
- 所有依赖、import、mock、断言、临时文件等细节均在模板中实现

## 6. 自动测试与性能测试
- 所有 *_test.go 文件自动集成 gofakeit.Struct(&param) 填充 mock 数据
- 所有 *_bench_test.go 文件自动集成 gofakeit.Struct(&params[i])
- test.go.tmpl 自动生成期望日志内容断言（拼接 param 字段，和实际日志内容比对）
- benchmark 测试写入临时文件或 mock write，避免 I/O 干扰
- 测试和 benchmark 文件命名规范：<struct>_test.go、<struct>_bench_test.go
- 运行 go test ./output，所有自动生成的测试和 benchmark 均可通过，日志内容断言自动化

## 7. 健壮性与最佳实践
- 生成代码需健壮，处理异常和空值，类型未映射、order缺失、属性缺失等需有报错或默认处理
- 注释需完整，便于维护
- 类型映射需灵活可扩展
- 日志内容顺序严格按 order
- formatter函数、单元测试需覆盖所有生成内容
- 所有生成的类型、方法、formatter函数名需全局唯一，避免命名冲突
- 推荐每个日志类型单独生成一个包/目录，或在命名中加入唯一标识
- mock数据生成推荐使用 gofakeit/faker 等库，自动填充结构体
- write函数采用依赖注入，便于mock和性能测试
- 单元测试和性能测试写入临时文件，避免污染主日志

## 8. 模板优先与维护
- 所有生成内容均应优先通过模板维护和修正，避免直接修改单个生成文件
- 只需维护模板和主流程，后续可批量更新所有生成代码

## 9. 常见问题与修复
- 断言内容为空：需传递 AllFieldNames 给 test.go.tmpl
- fmt 未使用：_ = fmt.Sprint 放到测试函数体内
- 字段顺序、类型、注释、唯一命名等问题，均应在模板和主流程统一修正

---

本提示词可直接用于后续日志 XML 结构自动生成 Go 代码、测试和性能测试工具的最佳实践与自动化实现。
