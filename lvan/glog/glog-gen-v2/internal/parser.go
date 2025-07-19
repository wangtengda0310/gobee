package internal

import (
	"encoding/xml"
	"os"
	"strconv"
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
			FuncName: toCamel(s.Name) + "Log",
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
