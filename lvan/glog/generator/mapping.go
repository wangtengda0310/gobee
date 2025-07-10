package generator

import (
	"encoding/xml"
	"os"
)

type TypeMapping struct {
	XMLName xml.Name          `xml:"mapping"`
	Types   []TypeMappingItem `xml:"type"`
}

type TypeMappingItem struct {
	XML string `xml:"xml,attr"`
	Go  string `xml:"go,attr"`
}

func LoadTypeMapping(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var mapping TypeMapping
	decoder := xml.NewDecoder(f)
	if err := decoder.Decode(&mapping); err != nil {
		return nil, err
	}
	m := map[string]string{
		"bool":    "bool",
		"byte":    "byte",
		"int":     "int",
		"int8":    "int8",
		"int16":   "int16",
		"int32":   "int32",
		"int64":   "int64",
		"uint":    "uint",
		"uint8":   "uint8",
		"uint16":  "uint16",
		"uint32":  "uint32",
		"uint64":  "uint64",
		"float32": "float32",
		"float64": "float64",
		"string":  "string",
	}
	for _, t := range mapping.Types {
		m[t.XML] = t.Go
	}
	return m, nil
}
