package internal

import (
	"encoding/xml"
	"os"
	"strconv"

	"github.com/xuri/excelize/v2"
)

type metalibXML struct {
	Structs []structXML `xml:"struct"`
}

type structXML struct {
	Name    string     `xml:"name,attr"`
	Version string     `xml:"version,attr"`
	Desc    string     `xml:"desc,attr"`
	Obj     string     `xml:"obj,attr"`
	Source  string     `xml:"source,attr"`
	Code    string     `xml:"code,attr"`
	Level   string     `xml:"level,attr"`
	IsGlog  string     `xml:"isglog,attr"`
	Type    string     `xml:"type,attr"`
	Trigger string     `xml:"trigger,attr"`
	Use     string     `xml:"use,attr"`
	Entries []entryXML `xml:"entry"`
}

type entryXML struct {
	Name      string `xml:"name,attr"`
	Type      string `xml:"type,attr"`
	Order     string `xml:"order,attr"`
	Title     string `xml:"title,attr"`
	Desc      string `xml:"desc,attr"`
	Catalog   string `xml:"catalog,attr"`
	Required  string `xml:"required,attr"`
	ExtType   string `xml:"ext_type,attr"`
	ExtID     string `xml:"ext_id,attr"`
	ExtIDDict string `xml:"ext_id_dict,attr"`
}

func ParseXML(path string) ([]Struct, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var m metalibXML
	if err := xml.NewDecoder(f).Decode(&m); err != nil {
		return nil, err
	}
	var structs []Struct
	for _, s := range m.Structs {
		st := Struct{
			Name:     s.Name,
			FuncName: funcName(toCamel(s.Name)),
			Version:  s.Version,
			Desc:     s.Desc,
			Obj:      s.Obj,
			Source:   s.Source,
			Code:     s.Code,
			Level:    s.Level,
			IsGlog:   s.IsGlog == "true",
			Type:     s.Type,
			Trigger:  s.Trigger,
			Use:      s.Use,
		}
		for _, e := range s.Entries {
			order, _ := strconv.Atoi(e.Order)
			st.Entries = append(st.Entries, Entry{
				Name:      toCamel(e.Name),
				EntryName: e.Name,
				XMLType:   e.Type,
				Order:     order,
				Title:     e.Title,
				Desc:      e.Desc,
				Catalog:   e.Catalog,
				Required:  e.Required == "true",
				ExtType:   e.ExtType,
				ExtID:     e.ExtID,
				ExtIDDict: e.ExtIDDict,
			})
		}
		structs = append(structs, st)
	}
	return structs, nil
}

// ParseExcel 解析 Excel 文件，返回 []Struct，结构与 ParseXML 完全一致
func ParseExcel(path string) ([]Struct, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, err
	}
	if len(rows) < 2 {
		return nil, nil // 没有数据
	}

	var structs []Struct
	rowIdx := 1 // 跳过表头
	for rowIdx < len(rows) {
		row := rows[rowIdx]
		// 检查A~L列是否有内容，判定为struct起始
		if len(row) < 13 {
			rowIdx++
			continue
		}
		// 解析struct属性
		st := Struct{
			Name:     row[3], // D 事件名称
			FuncName: funcName(toCamel(row[3])),
			Version:  row[4],        // E 版本号
			Desc:     row[0],        // A 事件
			Obj:      row[7],        // H 上报端
			Source:   row[10],       // K 通用/定制
			Code:     row[1],        // B 事件编码
			Level:    row[2],        // C 事件等级
			IsGlog:   row[9] == "是", // J 是否落G库
			Type:     row[5],        // F 事件类型
			Trigger:  row[6],        // G 埋点触发时机
			Use:      row[8],        // I 应用场景
		}
		// 事件备注（L）可作为注释desc补充
		if len(row) > 11 && row[11] != "" {
			st.Desc += " " + row[11]
		}
		// 统计struct的entry范围（A~L列合并单元格，直到下一个struct）
		start := rowIdx
		end := rowIdx
		for end+1 < len(rows) && (len(rows[end+1]) < 13 || allEmpty(rows[end+1][:12])) {
			end++
		}
		// 解析entry
		for i := start; i <= end; i++ {
			entryRow := rows[i]
			if len(entryRow) < 20 {
				continue
			}
			order, _ := strconv.Atoi(entryRow[12]) // M 序号
			st.Entries = append(st.Entries, Entry{
				Name:      toCamel(entryRow[13]), // N 字段
				EntryName: entryRow[13],
				XMLType:   entryRow[18], // S 字段类型
				Order:     order,
				Title:     entryRow[15],        // P 字段名称
				Desc:      entryRow[19],        // T 字段说明
				Catalog:   entryRow[14],        // O 字段分类
				Required:  entryRow[16] == "是", // Q 是否必传
				ExtType:   "",                  // 可扩展
				ExtID:     "",                  // 可扩展
				ExtIDDict: "",                  // 可扩展
			})
		}
		structs = append(structs, st)
		rowIdx = end + 1
	}
	return structs, nil
}

func funcName(name string) string {
	return name + "Log"
}

// allEmpty 判断一组单元格是否全为空
func allEmpty(cols []string) bool {
	for _, c := range cols {
		if c != "" {
			return false
		}
	}
	return true
}

func toCamel(s string) string {
	if s == "" {
		return s
	}
	b := []byte(s)
	res := make([]byte, 0, len(b))
	capNext := true
	for i := 0; i < len(b); i++ {
		if b[i] == '_' {
			capNext = true
			continue
		}
		if capNext && b[i] >= 'a' && b[i] <= 'z' {
			res = append(res, b[i]-'a'+'A')
			capNext = false
		} else {
			res = append(res, b[i])
			capNext = false
		}
	}
	return string(res)
}
