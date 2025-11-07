package template

import (
	"bytes"
	"fmt"
	"os"
	"text/template"
)

// RenderTemplate 读取模板文件并替换变量
// 参数:
//
//	filePath: 模板文件路径
//	data: 变量映射表 (key: 模板变量名, value: 替换值)
//
// 返回值:
//
//	string: 渲染后的内容
//	error: 错误信息
func RenderTemplate(filePath string, data map[string]string) (string, error) {
	// 读取模板文件
	tmplBytes, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %w", err)
	}

	// 创建模板并解析
	tmpl, err := template.New("").Parse(string(tmplBytes))
	if err != nil {
		return "", fmt.Errorf("解析模板失败: %w", err)
	}

	// 执行模板渲染
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("渲染模板失败: %w", err)
	}

	return buf.String(), nil
}
