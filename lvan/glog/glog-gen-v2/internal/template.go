package internal

import (
	_ "embed"
	"fmt"
	"os"
)

var GlobalTemplateDir = "./templates"

//go:embed templates/struct.go.tmpl
var structTmpl string

//go:embed templates/func.go.tmpl
var funcTmpl string

//go:embed templates/test.go.tmpl
var testTmpl string

//go:embed templates/bench.go.tmpl
var benchTmpl string

//go:embed templates/base_param.go.tmpl
var baseParamTmpl string

func GetTemplate(name string) (string, error) {
	path := fmt.Sprintf("%s/%s", GlobalTemplateDir, name)
	if b, err := os.ReadFile(path); err == nil {
		return string(b), nil
	}
	// fallback 到 embed 时增加命令行提示
	fmt.Printf("[glog-gen] 未找到本地模板 %s，自动 fallback 到内嵌模板\n", path)
	switch name {
	case "struct.go.tmpl":
		return structTmpl, nil
	case "func.go.tmpl":
		return funcTmpl, nil
	case "test.go.tmpl":
		return testTmpl, nil
	case "bench.go.tmpl":
		return benchTmpl, nil
	case "base_param.go.tmpl":
		return baseParamTmpl, nil
	default:
		return "", fmt.Errorf("模板 %s 未内嵌", name)
	}
}
