package mai

import (
	"encoding/csv"
	"os"
	"strings"
	"text/template"
)

// 假设CSV解析后的数据结构为TemplateData，包含类名和字段信息
type Field struct {
	Index   int
	Name    string
	Type    string
	Comment string
}

type TemplateData struct {
	ClassName string
	Fields    []Field
}

// text/template 示例
const codeTemplate = `using System;
using System.Collections;
using System.Collections.Generic;
using MessagePack;
using UnityEngine;

namespace DevSample1.GamePlay.MsgDataParse
{
[Serializable]
[MessagePackObject]
public class {{.ClassName}} : IDataBase
{
    {{range .Fields}}
    /// <summary>
    /// {{.Comment}}
    /// </summary>
    [Key({{.Index}})]
    public {{.Type}} {{.Name}};
    {{end}}

    public UInt32 GetKey()
    {
        return (uint)id;
    }
}
}`

func mainTemplate() {
	// 示例CSV数据（实际应从文件读取）
	csvData := `id,name
UInt32,string
角色ID,名称`

	// 解析CSV
	r := csv.NewReader(strings.NewReader(csvData))
	rows, _ := r.ReadAll()

	// 创建字段列表
	var fields []Field
	for col := 0; col < len(rows[0]); col++ {
		fields = append(fields, Field{
			Index:   col,
			Name:    rows[0][col], // 第一行：字段名
			Type:    rows[1][col], // 第二行：类型
			Comment: rows[2][col], // 第三行：注释
		})
	}

	// 生成类名（示例文件名）
	className := filenameToClassName("attr_enum.csv")

	// 准备模板数据
	data := TemplateData{
		ClassName: className,
		Fields:    fields,
	}

	// 执行模板
	tmpl := template.Must(template.New("code").Parse(codeTemplate))
	tmpl.Execute(os.Stdout, data)
}

// 文件名转类名逻辑
func filenameToClassName(filename string) string {
	// 去除扩展名
	name := strings.Split(filename, ".")[0]

	// 处理命名格式：下划线转驼峰
	parts := strings.Split(name, "_")
	for i := range parts {
		parts[i] = strings.Title(parts[i])
	}

	return strings.Join(parts, "") + "ClassData"
}
