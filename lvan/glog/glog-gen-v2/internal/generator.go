package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

// Parse 类型，便于方法值和闭包适配
// 返回 Struct 列表和错误
type Parse func() ([]Struct, error)

// 修改 Generate 签名，支持传入解析函数
func Generate(parse Parse, mappingPath, outDir string) error {
	fmt.Println("进入Generate主流程...")
	fmt.Printf("mappingPath: %s, outDir: %s\n", mappingPath, outDir)
	// 1. 解析类型映射
	typeMap, err := LoadTypeMapping(mappingPath)
	if err != nil {
		fmt.Printf("类型映射加载失败: %v\n", err)
		return fmt.Errorf("类型映射加载失败: %w", err)
	}
	fmt.Println("类型映射加载成功")
	// 2. 解析结构体（通过 parse 适配 XML/Excel）
	fmt.Println("即将解析结构体...")
	structs, err := parse()
	if err != nil {
		fmt.Printf("结构体解析失败: %v\n", err)
		return fmt.Errorf("结构体解析失败: %w", err)
	}
	fmt.Printf("解析得到 struct 数量: %d\n", len(structs))
	if len(structs) == 0 {
		fmt.Println("警告: 解析结果为空，未生成任何 struct")
	}
	fmt.Println("结构体解析成功")

	// 3. 填充模板数据结构并校验类型映射
	var baseEntries []Entry
	var newStructs []Struct
	for _, st := range structs {
		entries := make([]Entry, len(st.Entries))
		for i, e := range st.Entries {
			// 清洗字段名和类型
			e.Name = cleanStr(e.Name)
			e.Type = cleanStr(e.Type)
			e.XMLType = cleanStr(e.XMLType)
			goType, ok := typeMap[e.XMLType]
			if !ok {
				fmt.Printf("警告: struct %s 字段 %s 类型 %s 未映射，自动映射为 string\n", st.Name, e.Name, e.XMLType)
				goType = "string"
			}
			e.Type = cleanStr(goType)
			entries[i] = e
			if e.Catalog == "base" {
				baseEntries = addBaseEntryIfNotExists(baseEntries, e)
			}
		}
		st.Entries = entries
		newStructs = append(newStructs, st)
	}
	structs = newStructs
	// data := TemplateData{
	// 	Structs:     structs,
	// 	BaseEntries: baseEntries,
	// 	TypeMap:     typeMap,
	// }

	// 通用模板渲染写文件方法
	funcMap := template.FuncMap{"lower": func(s string) string { return strings.ToLower(s) }}

	// 4.1 生成 base_param.go
	fmt.Println("开始生成 base_param.go ...")
	baseFile := filepath.Join(outDir, "base_param.go")
	pkg := filepath.Base(outDir)
	if err := RenderTemplateToFile("base_param.go.tmpl", baseFile, map[string]interface{}{
		"BaseEntries": baseEntries,
		"Package":     pkg,
	}, funcMap); err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("base_param.go 生成完成")

	fmt.Println("即将进入struct循环，生成各类文件...")
	for _, st := range structs {
		fmt.Printf("生成 %s 相关文件...\n", st.Name)
		pkg := filepath.Base(outDir)
		allFields := st.Entries
		allFieldNames := getAllFieldNames(allFields)
		extEntries := getExtEntries(st.Entries)
		// 合并结构体和函数到一个文件
		funcStructFile := filepath.Join(outDir, st.Name+".go")
		if err := RenderTemplateToFile("func_struct.go.tmpl", funcStructFile, map[string]interface{}{
			"Package":       pkg,
			"Struct":        st,
			"AllFields":     allFields,
			"AllFieldNames": allFieldNames,
			"ExtEntries":    extEntries,
			"IsBase":        false,
			"BaseEntries":   baseEntries,
		}, funcMap); err != nil {
			fmt.Println(err)
			return err
		}
		// 单元测试文件
		testFile := filepath.Join(outDir, st.Name+"_test.go")
		if err := RenderTemplateToFile("test.go.tmpl", testFile, map[string]interface{}{
			"Package":       pkg,
			"Struct":        st,
			"AllFieldNames": allFieldNames,
		}, funcMap); err != nil {
			fmt.Println(err)
			return err
		}
		fmt.Printf("%s 相关文件生成完成\n", st.Name)
	}
	fmt.Println("所有struct文件生成完毕")
	return nil
}

func cleanStr(s string) string {
	// 移除所有不可见字符（控制字符）
	reg := regexp.MustCompile(`[\x00-\x1F\x7F]`)
	return reg.ReplaceAllString(s, "")
}

// 工具函数：如 baseEntries 去重、字段名和扩展字段提取
func addBaseEntryIfNotExists(baseEntries []Entry, e Entry) []Entry {
	for _, be := range baseEntries {
		if be.Name == e.Name {
			return baseEntries
		}
	}
	return append(baseEntries, e)
}

func getAllFieldNames(entries []Entry) []string {
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name)
	}
	return names
}

func getExtEntries(entries []Entry) []Entry {
	extEntries := []Entry{}
	for _, e := range entries {
		if e.Catalog == "ext" {
			extEntries = append(extEntries, e)
		}
	}
	return extEntries
}

// 通用模板渲染写文件方法
func RenderTemplateToFile(tmplName, fileName string, data map[string]interface{}, funcMap template.FuncMap) error {
	tmplContent, err := GetTemplate(tmplName)
	if err != nil {
		return fmt.Errorf("加载模板 %s 失败: %w", tmplName, err)
	}
	tmpl, err := template.New(tmplName).Funcs(funcMap).Parse(tmplContent)
	if err != nil {
		return fmt.Errorf("解析模板 %s 失败: %w", tmplName, err)
	}
	f, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("创建文件 %s 失败: %w", fileName, err)
	}
	defer f.Close()
	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("渲染模板 %s 到文件 %s 失败: %w", tmplName, fileName, err)
	}
	return nil
}
