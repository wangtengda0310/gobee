package internal

import (
	_ "embed"
	"fmt"
	"os"
)

var GlobalTemplateDir = "./templates"

//go:embed templates/test.go.tmpl
var testTmpl string

//go:embed templates/base_param.go.tmpl
var baseParamTmpl string

//go:embed templates/func_struct.go.tmpl
var funcStructTmpl string

func GetTemplate(name string) (string, error) {
	path := fmt.Sprintf("%s/%s", GlobalTemplateDir, name)
	if b, err := os.ReadFile(path); err == nil {
		return string(b), nil
	}
	// fallback 到 embed 时增加命令行提示
	fmt.Printf("[glog-gen] 未找到本地模板 %s，自动 fallback 到内嵌模板\n", path)
	switch name {
	case "test.go.tmpl":
		return testTmpl, nil
	case "base_param.go.tmpl":
		return baseParamTmpl, nil
	case "func_struct.go.tmpl":
		return funcStructTmpl, nil
	default:
		return "", fmt.Errorf("模板 %s 未内嵌", name)
	}
}
