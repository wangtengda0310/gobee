package mp

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/vmihailenco/msgpack/v5"
)

// https://sugendran.github.io/msgpack-visualizer/

type Module struct {
	XMLName xml.Name        `xml:"module"`
	Name    string          `xml:"name,attr"`
	Beans   map[string]Bean `xml:"bean"` // 使用map存储，key为bean name
	Tables  []Table         `xml:"table"`
}

type Bean struct {
	Name string `xml:"name,attr"`
	Vars []Var  `xml:"var"`
}

type Var struct {
	Name string `xml:"name,attr"`
	Type string `xml:"type,attr"`
}

type Table struct {
	Name  string `xml:"name,attr"`
	Mode  string `xml:"mode,attr"`
	Value string `xml:"value,attr"`
	Input string `xml:"input,attr"`
}

// 自定义UnmarshalXML方法处理map
func (mod *Module) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type Alias Module // 创建别名以避免递归调用
	aux := &struct {
		*Alias
		BeanList []Bean `xml:"bean"` // 临时存储bean列表
	}{
		Alias: (*Alias)(mod),
	}

	if err := d.DecodeElement(&aux, &start); err != nil {
		return err
	}

	// 初始化map
	mod.Beans = make(map[string]Bean)

	// 将bean列表转换为map
	for _, bean := range aux.BeanList {
		mod.Beans[bean.Name] = bean
	}

	return nil
}

type Type interface {
	pack(msgpacker *msgpack.Encoder, fval interface{}, mod *Module) error
}
type MonoType struct {
	Type string
}

func (b *MonoType) pack(msgpacker *msgpack.Encoder, fval interface{}, mod *Module) error {
	bean, ok := mod.Beans[b.Type]
	if ok {
		return pack(mod, msgpacker, fval, bean)
	}

	switch b.Type {
	case "byte", "Byte", "BYTE":
		switch fvalTyped := fval.(type) {
		case float64:
			if err := msgpacker.EncodeUint8(byte(fvalTyped)); err != nil {
				return err
			}
		case string:
			v := fvalTyped
			atoi, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			if err := msgpacker.EncodeUint8(byte(atoi)); err != nil {
				return err
			}
		default:
			return fmt.Errorf("不支持的数据类型 byte:%v", fval)
		}
	case "sbyte", "SByte", "sByte", "Sbyte":
		switch fvalTyped := fval.(type) {
		case float64:
			if err := msgpacker.EncodeInt8(int8(fvalTyped)); err != nil {
				return err
			}
		case string:
			v := fval.(string)
			atoi, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			if err := msgpacker.EncodeInt8(int8(atoi)); err != nil {
				return err
			}
		default:
			return fmt.Errorf("不支持的数据类型 sbyte:%v", fval)
		}
	case "UInt16", "uint16":
		switch fvalTyped := fval.(type) {
		case float64:
			if err := msgpacker.EncodeUint16(uint16(fvalTyped)); err != nil {
				return err
			}
		case string:
			v := fvalTyped
			atoi, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			if err := msgpacker.EncodeUint16(uint16(atoi)); err != nil {
				return err
			}
		}
	case "short", "Short", "SHORT":
		switch fvalTyped := fval.(type) {
		case float64:
			if err := msgpacker.EncodeInt16(int16(fvalTyped)); err != nil {
				return err
			}
		case string:
			v := fvalTyped
			atoi, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			if err := msgpacker.EncodeInt16(int16(atoi)); err != nil {
				return err
			}
		default:
			return fmt.Errorf("不支持的数据类型 short:%v", fval)
		}
	case "UInt32", "uint32":
		switch fvalTyped := fval.(type) {
		case float64:
			if err := msgpacker.EncodeUint32(uint32(fvalTyped)); err != nil {
				return err
			}
		case string:
			v := fvalTyped
			atoi, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			if err := msgpacker.EncodeUint32(uint32(atoi)); err != nil {
				return err
			}
		default:
			return fmt.Errorf("不支持的数据类型 uint32:%v", fval)
		}
	case "int", "Int", "INT":
		switch fvalTyped := fval.(type) {
		case float64:
			if err := msgpacker.EncodeInt32(int32(fvalTyped)); err != nil {
				return err
			}
		case string:
			v := fvalTyped
			atoi, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			if err := msgpacker.EncodeInt32(int32(atoi)); err != nil {
				return err
			}
		default:
			return fmt.Errorf("不支持的数据类型 int:%v", fval)
		}
	case "UInt64", "uint64":
		switch fvalTyped := fval.(type) {
		case float64:
			if err := msgpacker.EncodeUint64(uint64(fvalTyped)); err != nil {
				return err
			}
		case string:
			v := fvalTyped
			atoi, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			if err := msgpacker.EncodeUint64(uint64(atoi)); err != nil {
				return err
			}
		default:
			return fmt.Errorf("不支持的数据类型 uint64:%v", fval)
		}
	case "long", "Long", "LONG":
		switch fvalTyped := fval.(type) {
		case float64:
			if err := msgpacker.EncodeInt64(int64(fvalTyped)); err != nil {
				return err
			}
		case string:
			v := fvalTyped
			atoi, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			if err := msgpacker.EncodeInt64(int64(atoi)); err != nil {
				return err
			}
		default:
			return fmt.Errorf("不支持的数据类型 long:%v", fval)
		}
	case "bool", "boolean", "Bool", "Boolean", "BOOL", "BOOLEAN":
		switch fvalTyped := fval.(type) {
		case float64:
			if err := msgpacker.EncodeBool(fvalTyped != 0); err != nil {
				return err
			}
		case bool:
			if err := msgpacker.EncodeBool(fvalTyped); err != nil {
				return err
			}
		case string:
			if err := msgpacker.EncodeBool(fval == "true" || fval == "1"); err != nil {
				return err
			}
		default:
			return fmt.Errorf("不支持的数据类型 bool:%v", fval)
		}
	case "string":
		if strVal, ok := fval.(string); ok {
			if err := msgpacker.EncodeString(strVal); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("不支持的数据类型 string:%v", fval)
		}
	default:
		return fmt.Errorf("不支持的数据类型:%s", b.Type)
	}
	return nil
}

type KvType struct {
	Separator string
	KType     MonoType
	VType     MonoType
}

func (k KvType) pack(msgpacker *msgpack.Encoder, fval interface{}, mod *Module) error {

	switch fvalTyped := fval.(type) {
	case string:
		m := fvalTyped
		if err := msgpacker.EncodeArrayLen(len(m)); err != nil {
			return err
		}
		split := strings.Split(m, k.Separator)
		for i := 0; i < len(split); i += 2 {
			err := k.KType.pack(msgpacker, split[i], mod)
			if err != nil {
				return err
			}
			err = k.VType.pack(msgpacker, split[i+1], mod)
			if err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("不支持的数据类型:%v", fval)
	}
	return nil
}

type ListType struct {
	Separator string
	BaseType  Type
}

func (t *ListType) pack(msgpacker *msgpack.Encoder, fval interface{}, mod *Module) error {

	switch fvalTyped := fval.(type) {
	case string:
		l := fvalTyped
		split := strings.Split(l, t.Separator)
		if err := msgpacker.EncodeArrayLen(len(split)); err != nil {
			return err
		}
		for _, v := range split {
			err := t.BaseType.pack(msgpacker, v, mod)
			if err != nil {
				return err
			}
		}
	case map[any]any:
		err := t.BaseType.pack(msgpacker, fval, mod)
		if err != nil {
			return err
		}
	case []any:
		err := t.BaseType.pack(msgpacker, fval, mod)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("不支持的数据类型:%v", fval)
	}
	return nil
}

func parseComplexType(typeStr string) Type {

	//(map#sep=,),type1,type2                    对应C#字典Dictionary<type1, type2>   这里使用＂,＂作为分隔符,","可以被替换
	//(listKVP#sep=,),type1,type2                对应C#列表List<KeyValuePair<type1, type2>>
	//	(list#sep=,),type                         对应C#列表　List<type>
	//(array#sep=,),((array#sep=|),type)             对应C#数组type[][]          这里使用＂,＂作为分隔符,可以替换

	separator := ""
	var basetype = typeStr
	if strings.HasPrefix(typeStr, "(map") {
		parts := strings.SplitN(typeStr, "),", 2)
		if len(parts) < 2 {
			panic("不支持的数据类型:" + typeStr)
		}
		sepPart := strings.SplitN(parts[0], "sep=", 2)
		separator = ","
		if len(sepPart) > 1 {
			separator = strings.Trim(sepPart[1], "#=)")
		}
		tp := strings.Split(parts[1], ",")
		return KvType{separator, MonoType{Type: tp[0]}, MonoType{tp[1]}}
	} else if strings.HasPrefix(typeStr, "(listKVP") {
		parts := strings.SplitN(typeStr, "),", 2)
		if len(parts) < 2 {
			panic("不支持的数据类型:" + typeStr)
		}
		sepPart := strings.SplitN(parts[0], "sep=", 2)
		separator = ","
		if len(sepPart) > 1 {
			separator = strings.Trim(sepPart[1], "#=)")
		}
		tp := strings.Split(parts[1], ",")
		return KvType{separator, MonoType{Type: tp[0]}, MonoType{tp[1]}}
	} else if strings.HasPrefix(typeStr, "(list") {
		parts := strings.SplitN(typeStr, "),", 2)
		if len(parts) < 2 {
			panic("不支持的数据类型:" + typeStr)
		}

		listPart := strings.TrimPrefix(parts[0], "(list")
		sepPart := strings.SplitN(listPart, "sep=", 2)
		separator = ","
		if len(sepPart) > 1 {
			separator = strings.Trim(sepPart[1], "#=)")
		}

		return &ListType{Separator: separator, BaseType: &MonoType{Type: parts[1]}}
	} else if strings.HasPrefix(typeStr, "(array") {
		parts := strings.SplitN(typeStr, "),", 2)
		if len(parts) < 2 {
			panic("不支持的数据类型:" + typeStr)
		}
		sepPart := strings.SplitN(parts[0], "sep=", 2)
		separator = ","
		if len(sepPart) > 1 {
			separator = strings.Trim(sepPart[1], "#=)")
		}
		if strings.HasPrefix(parts[1], "(") {
			complexType := parseComplexType(parts[1][1 : len(parts[1])-1])
			return &ListType{Separator: separator, BaseType: complexType}
		}
		return &ListType{Separator: separator, BaseType: &MonoType{Type: parts[1]}}
	}

	return &MonoType{Type: basetype}
}

func printVar(v Var) {
	//complexType := parseComplexType(v.Type)
	//if complexType.IsList {
	//	fmt.Printf("    Var: %s (list of %s, sep=%s)\n",
	//		v.Name, complexType.BaseType, complexType.Separator)
	//} else {
	//	fmt.Printf("    Var: %s (%s)\n", v.Name, v.Type)
	//}
}

func tryParseXML(filepathstr, outputdir string) {

	file, err := os.Open(filepathstr)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	all, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	var module Module
	err = xml.Unmarshal(all, &module)
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}

	fmt.Printf("Module Name: %s\n", module.Name)

	// 获取绝对路径
	absPath, err := filepath.Abs(file.Name())
	if err != nil {
		fmt.Println("Error getting absolute path:", err)
		return
	}

	// 获取目录
	dir := filepath.Dir(absPath)
	fmt.Println("Absolute file directory:", dir)

	fmt.Println("\nBeans:")
	for _, bean := range module.Beans {
		fmt.Printf("  Bean Name: %s\n", bean.Name)
		for _, v := range bean.Vars {
			printVar(v)
		}
	}

	fmt.Println("\nTables:")
	for _, table := range module.Tables {
		manifest, packed := module.parsejosnfile(path.Join(dir, table.Input[2:]), module.Beans[table.Value])
		var name = strings.ToLower(table.Name)
		fmt.Printf("  Table Name: %s\n", table.Name)
		// 5. 保存文件
		err := os.MkdirAll(outputdir, 0755)
		if err != nil {
			panic(fmt.Errorf("创建文件夹失败:%w", err))
		}
		var msgpackfilename string
		if module.Name == "" {
			msgpackfilename = fmt.Sprintf("%s/%smanifest.bytes", outputdir, name)
		} else {
			msgpackfilename = fmt.Sprintf("%s/%s_%smanifest.bytes", outputdir, strings.ToLower(module.Name), name)
		}
		if err := os.WriteFile(msgpackfilename, manifest, 0644); err != nil {
			panic(fmt.Errorf("写入文件失败:%w", err))
		}
		var manifestfilename string
		if module.Name == "" {
			manifestfilename = fmt.Sprintf("%s/%s.bytes", outputdir, name)
		} else {
			manifestfilename = fmt.Sprintf("%s/%s_%s.bytes", outputdir, strings.ToLower(module.Name), name)
		}
		if err := os.WriteFile(manifestfilename, packed, 0644); err != nil {
			panic(fmt.Errorf("写入文件失败:%w", err))
		}

	}
}

func Mainjson(jsondir, output string) {

	// 1. 读取 CSV 文件
	err := filepath.WalkDir(jsondir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		fmt.Println(path)

		if strings.HasSuffix(path, ".xml") {
			tryParseXML(path, output)
		}
		return nil
	})
	if err != nil {
		fmt.Println("Error:", err)
	}

	log.Println("转换成功！")
}

func (mod *Module) parsejosnfile(filepath string, bean Bean) (m, d []byte) {
	fmt.Println(filepath)
	file, err := os.Open(filepath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	all, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}
	var v []interface{}

	err = json.Unmarshal(all, &v)
	if err != nil {
		panic(err)
	}

	// 按idindex排序rows（字典序）
	//slices.SortFunc(rows, func(a, b []string) int {
	//	aID := a[idindex]
	//	bID := b[idindex]
	//	return cmp.Compare(aID, bID) // 直接比较字符串
	//})

	buffer := &bytes.Buffer{}
	msgpacker := msgpack.NewEncoder(buffer)

	err = msgpacker.EncodeArrayLen(len(v))
	if err != nil {
		panic(err)
	}

	// 解析数据
	var manifests = make([]interface{}, 0)
	for _, row := range v {
		manifest := bean.packRecord(mod, msgpacker, buffer, row)
		manifests = append(manifests, manifest)
	}

	// 序列化为 MessagePack
	packedmanifest, err := msgpack.Marshal(manifests)
	if err != nil {
		panic(fmt.Errorf("序列化manifest失败:%w", err))
	}
	fmt.Println(hex.Dump(buffer.Bytes()))
	return packedmanifest, buffer.Bytes()
}

func (bean Bean) packRecord(mod *Module, msgpacker *msgpack.Encoder, buffer *bytes.Buffer, v interface{}) []interface{} {
	l := buffer.Len()

	if err := pack(mod, msgpacker, v, bean); err != nil {
		panic(fmt.Errorf("序列化失败:%w", err))
	}

	id := uint32(v.(map[string]interface{})["id"].(float64))
	return []interface{}{id, l, buffer.Len() - l}
}
func pack(mod *Module, msgpacker *msgpack.Encoder, v interface{}, bean Bean) error {
	switch vTyped := v.(type) {
	case []interface{}:

		if err := msgpacker.EncodeArrayLen(len(vTyped)); err != nil {
			return err
		}
		for _, vinner := range v.([]interface{}) {

			if err := pack(mod, msgpacker, vinner, bean); err != nil {
				return err
			}
		}
	case map[string]interface{}:
		m := v.(map[string]interface{})
		if err := msgpacker.EncodeArrayLen(len(bean.Vars)); err != nil {
			return err
		}
		for _, field := range bean.Vars {
			fval, ok := m[field.Name]
			if !ok {
				continue
			}
			complexType := parseComplexType(field.Type)
			err := complexType.pack(msgpacker, fval, mod)
			if err != nil {
				return err
			}
		}

	}
	return nil
}
