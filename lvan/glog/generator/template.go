package generator

import (
	_ "embed"
	"fmt"
	"os"
	"text/template"
)

//go:embed templates/code.tmpl
var codeTmpl string

//go:embed templates/test.tmpl
var testTmpl string

//go:embed templates/bench.tmpl
var benchTmpl string

func RenderTemplate(tmplName string, data interface{}, outputPath string) error {
	var content string
	switch tmplName {
	case "code.tmpl":
		content = codeTmpl
	case "test.tmpl":
		content = testTmpl
	case "bench.tmpl":
		content = benchTmpl
	default:
		return fmt.Errorf("未知模板: %s", tmplName)
	}

	tmpl, err := template.New("").Funcs(FuncMap()).Parse(content)
	if err != nil {
		return err
	}
	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()
	return tmpl.Execute(f, data)
}

func getFileName(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[i+1:]
		}
	}
	return path
}

// FuncMap 提供模板辅助函数
func FuncMap() template.FuncMap {
	return template.FuncMap{
		"lower": func(s string) string {
			if len(s) == 0 {
				return s
			}
			return string(s[0]+32) + s[1:]
		},
		"sub1": func(i int) int {
			return i - 1
		},
	}
}
