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
				found := false
				for _, be := range baseEntries {
					if be.Name == e.Name {
						found = true
						break
					}
				}
				if !found {
					baseEntries = append(baseEntries, e)
				}
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

	// 4. 渲染模板并输出到 outDir
	var tmplContent string

	// 4.1 生成 base_param.go
	fmt.Println("开始生成 base_param.go ...")
	tmplContent, err = GetTemplate("base_param.go.tmpl")
	if err != nil {
		fmt.Println("加载base_param模板失败:", err)
		return err
	}
	baseTmpl, err := template.New("base_param.go.tmpl").Funcs(template.FuncMap{
		"lower": func(s string) string { return strings.ToLower(s) },
	}).Parse(tmplContent)
	if err != nil {
		fmt.Println("加载base_param模板失败:", err)
		return err
	}
	baseFile := filepath.Join(outDir, "base_param.go")
	fmt.Println("准备创建文件:", baseFile)
	bf, err := os.Create(baseFile)
	if err != nil {
		fmt.Println("创建base_param.go失败:", err)
		return err
	}
	defer bf.Close()
	fmt.Println("准备渲染模板到文件:", baseFile)
	if err := baseTmpl.Execute(bf, map[string]interface{}{
		"BaseEntries": baseEntries,
	}); err != nil {
		fmt.Println("渲染base_param.go失败:", err)
		return err
	}
	fmt.Println("base_param.go 生成完成")

	fmt.Println("即将进入struct循环，生成各类文件...")
	for _, st := range structs {
		fmt.Printf("生成 %s 相关文件...\n", st.Name)
		pkg := "output"
		// 结构体文件
		fmt.Printf("生成 %s.go ...\n", st.Name)
		tmplContent, err = GetTemplate("struct.go.tmpl")
		if err != nil {
			fmt.Println("加载struct模板失败:", err)
			return err
		}
		structTmpl, err := template.New("struct.go.tmpl").Funcs(template.FuncMap{
			"lower": func(s string) string { return strings.ToLower(s) },
		}).Parse(tmplContent)
		if err != nil {
			fmt.Println("加载struct模板失败:", err)
			return err
		}
		structFile := filepath.Join(outDir, st.Name+".go")
		f, err := os.Create(structFile)
		if err != nil {
			fmt.Println("创建struct文件失败:", err)
			return err
		}
		extEntries := []Entry{}
		for _, e := range st.Entries {
			if e.Catalog == "ext" {
				extEntries = append(extEntries, e)
			}
		}
		if err := structTmpl.Execute(f, map[string]interface{}{
			"Package":     pkg,
			"Struct":      st,
			"ExtEntries":  extEntries,
			"IsBase":      false,
			"BaseEntries": baseEntries,
		}); err != nil {
			fmt.Println("渲染struct文件失败:", err)
			return err
		}
		f.Close()

		// 日志函数文件
		fmt.Printf("生成 %s_func.go ...\n", st.Name)
		tmplContent, err = GetTemplate("func.go.tmpl")
		if err != nil {
			fmt.Println("加载func模板失败:", err)
			return err
		}
		funcTmpl, err := template.New("func.go.tmpl").Funcs(template.FuncMap{
			"lower": func(s string) string { return strings.ToLower(s) },
		}).Parse(tmplContent)
		if err != nil {
			fmt.Println("加载func模板失败:", err)
			return err
		}
		funcFile := filepath.Join(outDir, st.Name+"_func.go")
		ff, err := os.Create(funcFile)
		if err != nil {
			fmt.Println("创建func文件失败:", err)
			return err
		}
		allFields := st.Entries
		allFieldNames := []string{}
		for _, e := range allFields {
			allFieldNames = append(allFieldNames, e.Name)
		}
		if err := funcTmpl.Execute(ff, map[string]interface{}{
			"Package":       pkg,
			"Struct":        st,
			"AllFields":     allFields,
			"AllFieldNames": allFieldNames,
		}); err != nil {
			fmt.Println("渲染func文件失败:", err)
			return err
		}
		ff.Close()

		// 单元测试文件
		fmt.Printf("生成 %s_test.go ...\n", st.Name)
		tmplContent, err = GetTemplate("test.go.tmpl")
		if err != nil {
			fmt.Println("加载test模板失败:", err)
			return err
		}
		testTmpl, err := template.New("test.go.tmpl").Funcs(template.FuncMap{
			"lower": func(s string) string { return strings.ToLower(s) },
		}).Parse(tmplContent)
		if err != nil {
			fmt.Println("加载test模板失败:", err)
			return err
		}
		testFile := filepath.Join(outDir, st.Name+"_test.go")
		tf, err := os.Create(testFile)
		if err != nil {
			fmt.Println("创建test文件失败:", err)
			return err
		}
		allFields = st.Entries
		allFieldNames = []string{}
		for _, e := range allFields {
			allFieldNames = append(allFieldNames, e.Name)
		}
		if err := testTmpl.Execute(tf, map[string]interface{}{
			"Package":       pkg,
			"Struct":        st,
			"AllFieldNames": allFieldNames,
		}); err != nil {
			fmt.Println("渲染test文件失败:", err)
			return err
		}
		tf.Close()

		// benchmark 文件
		fmt.Printf("生成 %s_bench_test.go ...\n", st.Name)
		tmplContent, err = GetTemplate("bench.go.tmpl")
		if err != nil {
			fmt.Println("加载bench模板失败:", err)
			return err
		}
		benchTmpl, err := template.New("bench.go.tmpl").Funcs(template.FuncMap{
			"lower": func(s string) string { return strings.ToLower(s) },
		}).Parse(tmplContent)
		if err != nil {
			fmt.Println("加载bench模板失败:", err)
			return err
		}
		benchFile := filepath.Join(outDir, st.Name+"_bench_test.go")
		bf, err := os.Create(benchFile)
		if err != nil {
			fmt.Println("创建bench文件失败:", err)
			return err
		}
		allFields = st.Entries
		allFieldNames = []string{}
		for _, e := range allFields {
			allFieldNames = append(allFieldNames, e.Name)
		}
		if err := benchTmpl.Execute(bf, map[string]interface{}{
			"Package":       pkg,
			"Struct":        st,
			"AllFields":     allFields,
			"AllFieldNames": allFieldNames,
		}); err != nil {
			fmt.Println("渲染bench文件失败:", err)
			return err
		}
		bf.Close()
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
