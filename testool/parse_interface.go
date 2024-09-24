package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
	"text/template"
)

func gen(filename string, sourceCode any) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, sourceCode, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	//ast.Print(fset, node)

	// 遍历AST以查找接口
	ast.Inspect(node, func(n ast.Node) bool {
		// 检查节点是否为类型声明
		genDecl, ok := n.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			return true
		}

		// 遍历类型声明中的所有规格
		for _, spec := range genDecl.Specs {
			// 检查规格是否为类型规格
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			collectInterfaceMethods(typeSpec, false)
		}
		return true
	})
	//format.Node(os.Stderr, fset, node)
	//printer.Fprint(os.Stdout, fset, node)
}

type interfaceSignature struct {
	Name    string
	Methods []*methodSignature
}

func (i *interfaceSignature) parseTemplate(templateName string) {
	parseTemplate(templates[templateName], i)
}

type methodSignature struct {
	Name    string
	Params  [][]string
	Returns []string
}

func (m *methodSignature) ParamsWithType() string {
	var s []string
	for _, param := range m.Params {
		s = append(s, strings.Join(param, " "))
	}
	return strings.Join(s, ",")
}
func (m *methodSignature) parseTemplate(templateName string) {
	parseTemplate(templates[templateName], m)
}
func (m *methodSignature) ReturnsWithVarName() []string {
	var s []string
	for i, r := range m.Returns {
		i++
		s = append(s, fmt.Sprint("p", i, " ", r))
	}
	return s
}
func (m *methodSignature) ReturnsWithoutVarName() []string {
	var s []string
	for i, _ := range m.Returns {
		i++
		s = append(s, fmt.Sprint("p", i))
	}
	return s
}
func collectInterfaceMethods(typeSpec *ast.TypeSpec, isEmbeded bool) (methods []*methodSignature) {
	// 检查类型是否为接口类型
	if ifaceType, ok := typeSpec.Type.(*ast.InterfaceType); ok {
		// 如果需要，可以进一步遍历接口的方法
		for _, field := range ifaceType.Methods.List {
			// 检查是否为方法
			if field.Names != nil {
				var method []*methodSignature
				method = methodDescriber(field.Type.(*ast.FuncType), field)
				methods = append(methods, method...)
			} else {
				m := collectInterfaceMethods(field.Type.(*ast.Ident).Obj.Decl.(*ast.TypeSpec), true)
				methods = append(methods, m...)
			}
		}
		if !isEmbeded {

			// type Mock{{.Interface}}Impl struct {
			//    {{range .SliceItems}}
			//        {{.Name}}Func {{.Name}}Func
			//    {{end}}
			// }

			(&interfaceSignature{Name: typeSpec.Name.Name, Methods: methods}).
				parseTemplate("createInterface")
			for _, ms := range methods {
				ms.parseTemplate("withMockerMethod")
			}

		}
	}
	return
}

var templates = map[string]string{
	"withMockerMethod": `
	type {{.Name}}Func func({{.ParamsWithType}}) {{if .Returns}}({{join .Returns ","}}){{end}}
	func (x *MockI){{.Name}}({{.ParamsWithType}}) {{if .Returns}}({{join .Returns ","}}){{end}} {
		{{if .Returns}}return {{end}}x.impl.{{.Name}}Func()
	}
	func with{{.Name}}(t *testing.T, {{if .ReturnsWithVarName}}{{join .ReturnsWithVarName ","}}{{end}}) func(m *Mock{{.Name}}) { // 这里多个参数的处理
		t.Helper()
		return  func(m *Mock{{.Name}}) {
			m.impl.{{.Name}}Func = func({{.ParamsWithType}}) {{if .Returns}}({{join .Returns ","}}){{end}} {
				{{if .Returns}}return {{end}}{{join .ReturnsWithoutVarName ","}}
			}
		}
	}
`,

	"createInterface": `
	type Mock{{.Name}}Impl struct {
		{{range .Methods}} {{.Name}}Func {{.Name}}Func
		{{end}}
	}

	type Mock{{.Name}} struct {
		impl Mock{{.Name}}Impl
	}
	type mocker func(*testing.T, *Mock{{.Name}})
	func Create{{.Name}}(t *testing.T, mockers ...mocker) *Mock{{.Name}} {
		t.Helper()
		attributeOwner := &Mock{{.Name}}{}
		for _, m := range mockers {
			m(t, attributeOwner)
		}
		return attributeOwner
	}
`,
}

func parseTemplate(tn string, signature any) {

	// 创建一个新的template.Template实例。
	toolMethodTmpl, err := template.New(tn).
		Funcs(template.FuncMap{"join": strings.Join}).
		Parse(tn)
	if err != nil {
		log.Fatal(err)
	}

	// 使用工具方法模板和数据生成Go代码。
	if err := toolMethodTmpl.Execute(os.Stdout, signature); err != nil {
		log.Fatal(err)
	}
}

func methodDescriber(funcType *ast.FuncType, field *ast.Field) (methods []*methodSignature) {

	// 接口定义的方法
	for _, name := range field.Names {
		// 遍历函数类型的参数，构建参数和返回值的字符串表示。
		p := accessParams(funcType, field)
		// 遍历函数类型的返回值，构建参数和返回值的字符串表示。
		returns := accessResutl(funcType, field)
		methods = append(methods, &methodSignature{
			Name:    name.Name,
			Params:  p,
			Returns: returns,
		})
	}

	return
}

func accessParams(sigFunc *ast.FuncType, field *ast.Field) (params [][]string) {
	// 遍历函数类型的参数，构建参数和返回值的字符串表示。
	for _, f := range sigFunc.Params.List {
		for _, name := range f.Names {
			paramType := f.Type.(*ast.Ident).Name
			params = append(params, []string{name.Name, paramType})
		}
	}
	return
}

func accessResutl(sigFunc *ast.FuncType, field *ast.Field) (returns []string) {
	returnType := ""
	// 遍历函数类型的返回值，构建参数和返回值的字符串表示。
	if sigFunc.Results != nil {
		for _, f := range sigFunc.Results.List {
			returnType = f.Type.(*ast.Ident).Name
			returns = append(returns, returnType)
		}
	}

	return
}

type methodAcceptor func(*ast.FuncType, *ast.Field)
