package internal

import (
	"encoding/xml"
	"fmt"
	"os"
)

type mappingXML struct {
	Types []struct {
		XML string `xml:"xml,attr"`
		Go  string `xml:"go,attr"`
	} `xml:"type"`
}

func LoadTypeMapping(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var m mappingXML
	if err := xml.NewDecoder(f).Decode(&m); err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for _, t := range m.Types {
		if t.XML == "" || t.Go == "" {
			return nil, fmt.Errorf("类型映射配置有空值: %+v", t)
		}
		result[t.XML] = t.Go
	}
	return result, nil
}
