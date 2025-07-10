package generator

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Metalib struct {
	XMLName xml.Name     `xml:"metalib"`
	Structs []StructItem `xml:"struct"`
}

type StructItem struct {
	Name    string      `xml:"name,attr"`
	Version string      `xml:"version,attr"`
	Desc    string      `xml:"desc,attr"`
	Obj     string      `xml:"obj,attr"`
	Source  string      `xml:"source,attr"`
	Code    string      `xml:"code,attr"`
	Level   string      `xml:"level,attr"`
	Isglog  string      `xml:"isglog,attr"`
	Type    string      `xml:"type,attr"`
	Trigger string      `xml:"trigger,attr"`
	Use     string      `xml:"use,attr"`
	Entries []EntryItem `xml:"entry"`
}

type EntryItem struct {
	Name     string `xml:"name,attr"`
	Catelogd string `xml:"catelogd,attr"`
	Type     string `xml:"type,attr"`
	Order    string `xml:"order,attr"`
	Title    string `xml:"title,attr"`
}

func Generate(xmlPath, mappingPath, outDir string) error {
	// 1. 解析类型映射
	typeMap, err := LoadTypeMapping(mappingPath)
	if err != nil {
		return fmt.Errorf("类型映射加载失败: %v", err)
	}

	// 2. 解析XML
	f, err := os.Open(xmlPath)
	if err != nil {
		return fmt.Errorf("XML文件打开失败: %v", err)
	}
	defer f.Close()
	var meta Metalib
	decoder := xml.NewDecoder(f)
	if err := decoder.Decode(&meta); err != nil {
		return fmt.Errorf("XML解析失败: %v", err)
	}

	// 3. 处理BaseParam
	baseMap := map[string]TemplateEntry{}
	for _, s := range meta.Structs {
		for _, e := range s.Entries {
			if e.Catelogd == "base" {
				order, _ := strconv.Atoi(e.Order)
				baseMap[e.Name] = TemplateEntry{
					Name:     ToCamel(e.Name),
					RawName:  e.Name,
					Type:     typeMap[e.Type],
					XmlType:  e.Type,
					Order:    order,
					Catelogd: e.Catelogd,
					Title:    e.Title,
				}
			}
		}
	}
	baseEntries := make([]TemplateEntry, 0, len(baseMap))
	for _, v := range baseMap {
		baseEntries = append(baseEntries, v)
	}
	sort.Slice(baseEntries, func(i, j int) bool { return baseEntries[i].Order < baseEntries[j].Order })

	// 4. 处理每个struct
	structs := []TemplateStruct{}
	for _, s := range meta.Structs {
		entries := []TemplateEntry{}
		extEntries := []TemplateEntry{}
		for _, e := range s.Entries {
			if e.Type == "" || e.Order == "" {
				return fmt.Errorf("struct %s entry %s 缺少type或order", s.Name, e.Name)
			}
			goType, ok := typeMap[e.Type]
			if !ok {
				goType = "string" // 未映射类型默认string
			}
			order, _ := strconv.Atoi(e.Order)
			te := TemplateEntry{
				Name:     ToCamel(e.Name),
				RawName:  e.Name,
				Type:     goType,
				XmlType:  e.Type,
				Order:    order,
				Catelogd: e.Catelogd,
				Title:    e.Title,
			}
			entries = append(entries, te)
			if e.Catelogd == "ext" {
				extEntries = append(extEntries, te)
			}
		}
		sort.Slice(entries, func(i, j int) bool { return entries[i].Order < entries[j].Order })
		sort.Slice(extEntries, func(i, j int) bool { return extEntries[i].Order < extEntries[j].Order })
		comment := []string{
			fmt.Sprintf("name: %s", s.Name),
			fmt.Sprintf("version: %s", s.Version),
			fmt.Sprintf("desc: %s", s.Desc),
			fmt.Sprintf("obj: %s", s.Obj),
			fmt.Sprintf("source: %s", s.Source),
			fmt.Sprintf("code: %s", s.Code),
			fmt.Sprintf("level: %s", s.Level),
			fmt.Sprintf("isglog: %s", s.Isglog),
			fmt.Sprintf("type: %s", s.Type),
			fmt.Sprintf("trigger: %s", s.Trigger),
			fmt.Sprintf("use: %s", s.Use),
		}
		structs = append(structs, TemplateStruct{
			Name:        ToCamel(s.Name),
			ParamName:   ToCamel(s.Name) + "Param",
			Comment:     comment,
			Entries:     entries,
			BaseEntries: baseEntries,
			ExtEntries:  extEntries,
		})
	}

	packageName := filepath.Base(outDir)
	data := TemplateData{
		Structs:     structs,
		BaseEntries: baseEntries,
		PackageName: packageName,
	}

	// 5. 渲染模板
	os.MkdirAll(outDir, 0755)
	if err := RenderTemplate("code.tmpl", data, outDir+"/log_gen.go"); err != nil {
		return fmt.Errorf("代码模板渲染失败: %v", err)
	}
	if err := RenderTemplate("test.tmpl", data, outDir+"/log_gen_test.go"); err != nil {
		return fmt.Errorf("测试模板渲染失败: %v", err)
	}
	if err := RenderTemplate("bench.tmpl", data, outDir+"/log_gen_bench_test.go"); err != nil {
		return fmt.Errorf("性能测试模板渲染失败: %v", err)
	}
	return nil
}

// ToCamel 字符串转驼峰，自动过滤非法字符
func ToCamel(s string) string {
	// 先替换非法字符为下划线
	runes := []rune(s)
	var badStr bool
	for i, r := range runes {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_') {
			runes[i] = '_'
			badStr = true
		}
	}
	if badStr {
		fmt.Fprintf(os.Stderr, "struct %s entry %s 包含非法字符，已替换为下划线", s, s)
	}
	safe := string(runes)
	parts := strings.Split(safe, "_")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, "")
}
