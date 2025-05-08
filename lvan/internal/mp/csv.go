package mp

import (
	"bytes"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/vmihailenco/msgpack/v5"
)

// 客户端原来的特殊处理逻辑 抄过来
func adaptContent(csv string) string {
	csv = strings.TrimRight(csv, "|;")
	csv = strings.Trim(csv, "\"")
	csv = strings.TrimPrefix(csv, "<p>")
	csv = strings.TrimSuffix(csv, "</p>")
	csv = strings.Replace(csv, "\"\"\"\"", "\"", -1)
	return csv
}

func parseBool(s string) (bool,error) {
	if s == "true" || s == "1" {
		return true, nil
	} else if s == "false" || s == "0" || s == "" {
		return false, nil
	} else {
		return false, errors.New("invalid bool value")
	}
}

// 自动推断字符串类型，转换为 interface{}
func inferType(vtype string, s string) (interface{}, bool) {
	firstSep := ";"
	secondSep := "|"
	sep := "|"
	switch vtype {
	case "string", "String":
		if s == "0" {
			return "", false // 是""不是"0"
		} else {
			return s, false
		}
	case "bool", "Bool", "boolean", "Boolean":
		if b,e := parseBool(s); e == nil {
			return b, false
		} else {
			return false, true
		}
	case "byte", "Byte":
		if s == "" {
			return byte(0), false
		}
		// 尝试转换为整数
		if intVal, err := strconv.ParseUint(s, 10, 8); err == nil {
			return byte(intVal), false
		} else {
			panic(err)
		}
	case "int8", "Int8":
		if s == "" {
			return int8(0), false
		}
		// 尝试转换为整数
		if intVal, err := strconv.ParseInt(s, 10, 8); err == nil {
			return int8(intVal), false
		} else {
			panic(err)
		}
	case "uint8", "Uint8":
		if s == "" {
			return uint8(0), false
		}
		// 尝试转换为整数
		if intVal, err := strconv.ParseUint(s, 10, 8); err == nil {
			return uint8(intVal), false
		} else {
			panic(err)
		}
	case "int16", "Int16", "short", "Short":
		if s == "" {
			return int16(0), false
		}
		// 尝试转换为整数
		if intVal, err := strconv.ParseInt(s, 10, 16); err == nil {
			return int16(intVal), false
		} else {
			panic(err)
		}
	case "uint16", "Uint16":
		if s == "" {
			return uint16(0), false
		}
		if intVal, err := strconv.ParseUint(s, 10, 16); err == nil {
			return uint16(intVal), false
		} else {
			panic(err)
		}
	case "int", "Int":
		if s == "" {
			return 0, false
		}
		if intVal, err := strconv.Atoi(s); err == nil {
			return intVal, false
		} else {
			panic(err)
		}
	case "int32", "Int32":
		if s == "" {
			return int32(0), false
		}
		if intVal, err := strconv.Atoi(s); err == nil {
			return int32(intVal), false
		} else {
			panic(err)
		}
	case "uint32", "Uint32":
		if s == "" {
			return uint32(0), false
		}
		if intVal, err := strconv.ParseUint(s, 10, 32); err == nil {
			return uint32(intVal), false
		} else {
			panic(err)
		}
	case "int64", "Int64", "long", "Long":
		if s == "" {
			return int64(0), false
		}
		if intVal, err := strconv.ParseInt(s, 10, 64); err == nil {
			return int64(intVal), false
		} else {
			panic(err)
		}
	case "uint64", "Uint64":
		if s == "" {
			return uint64(0), false
		}
		if intVal, err := strconv.ParseUint(s, 10, 64); err == nil {
			return uint64(intVal), false
		} else {
			panic(err)
		}
	case "string[]", "String[]":
		var r = make([]string, 0)
		if s == "" {
			return r, false
		}
		split := strings.Split(s, sep)
		r = append(r, split...)
		return r, false
	case "byte[]", "Byte[]":
		var r = make([]byte, 0)
		if s == "" {
			return r, false
		}
		split := strings.Split(s, sep)
		for _, v := range split {
			if intVal, err := strconv.ParseUint(v, 10, 8); err == nil {
				r = append(r, byte(intVal))
			} else {
				panic(err)
			}
		}
		return r, false
	case "int8[]", "Int8[]":
		var r = make([]int8, 0)
		if s == "" {
			return r, false
		}
		split := strings.Split(s, sep)
		for _, v := range split {
			if v == "" {
				r = append(r, 0)
				continue
			}
			if intVal, err := strconv.ParseInt(v, 10, 8); err == nil {
				r = append(r, int8(intVal))
			} else {
				panic(err)
			}
		}
		return r, false
	case "uint8[]", "Uint8[]": // 使用 github.com/shamaton/msgpack/v2 会序列化成字符串
		var r = make([]uint8, 0)
		if s == "" {
			return r, false
		}
		split := strings.Split(s, secondSep)
		for _, v := range split {
			if v == "" {
				r = append(r, 0)
				continue
			}
			if intVal, err := strconv.ParseUint(v,10,8); err == nil {
				r = append(r, uint8(intVal))
			} else {
				panic(err)
			}
		}
		return r,false
	case "int16[]", "Int16[]":
		var r = make([]int16, 0)
		if s == "" {
			return r, false
		}
		split := strings.Split(s, sep)
		for _, v := range split {
			if intVal, err := strconv.ParseInt(v, 10, 16); err == nil {
				r = append(r, int16(intVal))
			} else {
				panic(err)
			}
		}
		return r, false
	case "uint16[]", "Uint16[]":
		var r = make([]uint16, 0)
		if s == "" {
			return r, false
		}
		split := strings.Split(s, sep)
		for _, v := range split {
			if intVal, err := strconv.ParseUint(v, 10, 16); err == nil {
				r = append(r, uint16(intVal))
			} else {
				panic(err)
			}
		}
		return r, false
	case "int32[]", "Int32[]":
		var r = make([]int32, 0)
		if s == "" {
			return r, false
		}
		split := strings.Split(s, sep)
		for _, v := range split {
			if intVal, err := strconv.ParseInt(v, 10, 32); err == nil {
				r = append(r, int32(intVal))
			} else {
				panic(err)
			}
		}
		return r, false
	case "uint32[]", "Uint32[]":
		var r = make([]uint32, 0)
		if s == "" {
			return r, false
		}
		split := strings.Split(s, sep)
		for _, v := range split {
			if intVal, err := strconv.ParseUint(v, 10, 32); err == nil {
				r = append(r, uint32(intVal))
			} else {
				panic(err)
			}
		}
		return r, false
	case "int64[]", "Int64[]":
		var r = make([]int64, 0)
		if s == "" {
			return r, false
		}
		split := strings.Split(s, sep)
		for _, v := range split {
			if intVal, err := strconv.ParseInt(v, 10, 64); err == nil {
				r = append(r, int64(intVal))
			} else {
				panic(err)
			}
		}
		return r, false
	case "uint64[]", "Uint64[]":
		var r = make([]uint64, 0)
		if s == "" {
			return r, false
		}
		split := strings.Split(s, sep)
		for _, v := range split {
			if intVal, err := strconv.ParseUint(v, 10, 64); err == nil {
				r = append(r, uint64(intVal))
			} else {
				panic(err)
			}
		}
		return r, false
	case "int32[][]", "Int32[][]", "int32:int32", "Int32:Int32":
		var r = make([][]int32, 0)
		if s == "" {
			return r, false
		}
		splitOut := strings.Split(s, firstSep)
		for _, sinner := range splitOut {
			if sinner == "" {
				r = append(r, []int32{})
				continue
			}
			split := strings.Split(sinner, secondSep)
			var rinner []int32
			for _, v := range split {
				if intVal, err := strconv.ParseInt(v, 10, 32); err == nil {
					rinner = append(rinner, int32(intVal))
				} else {
					panic(err)
				}
			}
			r = append(r, rinner)
		}
		return r, false
	case "uint32[][]", "Uint32[][]", "uint32:uint32", "Uint32:Uint32":
		var r = make([][]uint32, 0)
		if s == "" {
			return r, false
		}
		splitOut := strings.Split(s, firstSep)
		for _, sinner := range splitOut {
			var rinner = make([]uint32, 0)
			if sinner == "" {
				r = append(r, rinner)
				continue
			}
			split := strings.Split(sinner, secondSep)
			for _, v := range split {
				if intVal, err := strconv.ParseUint(v, 10, 32); err == nil {
					rinner = append(rinner, uint32(intVal))
				} else {
					panic(err)
				}
			}
			r = append(r, rinner)
		}
		return r, false
	case "kv:string", "kv:String":
		var r = make(map[uint32]string)
		if s == "" {
			return r, false
		}
		splitOut := strings.Split(s, firstSep)
		for _, sinner := range splitOut {
			if sinner == "" {
				continue
			}
			split := strings.Split(sinner, secondSep)
			if k, err := strconv.ParseUint(split[0], 10, 32); err == nil {
				if _,ok := r[uint32(k)]; ok {
					panic("key already exists")
				}
				r[uint32(k)] = split[1]
			} else {
				panic(err)
			}
		}
		return r, false
	case "kv:Bool", "kv:bool","kv:Boolean", "kv:boolean":
		var r = make(map[uint32]bool)
		if s == "" {
			return r, false
		}
		splitOut := strings.Split(s, firstSep)
		for _, sinner := range splitOut {
			if sinner == "" {
				continue
			}
			split := strings.Split(sinner, secondSep)
			var key uint32
			var valb bool
			if k, err := strconv.ParseUint(split[0], 10, 32); err == nil {
				key = uint32(k)
				if _,ok := r[key]; ok {
					panic("key already exists")
				}
			} else {
				panic(err)
			}
			if b,e := parseBool(split[1]); e != nil {
				panic(e)
			} else {
				valb = b
			}
			r[uint32(key)] = valb
		}
		return r, false
	case "kv:int", "kv:Int","kv:int32", "kv:Int32":
		var r = make(map[uint32]int32)
		if s == "" {
			return r, false
		}
		splitOut := strings.Split(s, firstSep)
		for _, sinner := range splitOut {
			if sinner == "" {
				continue
			}
			split := strings.Split(sinner, secondSep)
			var key uint32
			var val int32
			if k, err := strconv.ParseUint(split[0], 10, 32); err == nil {
				key = uint32(k)
				if _,ok := r[key]; ok {
					panic("key already exists")
				}
			} else {
				panic(err)
			}
			if v, err := strconv.ParseInt(split[1], 10, 32); err == nil {
				val = int32(v)
			} else {
				panic(err)
			}
			r[key] = val
		}
		return r, false
	case "kv:uint32", "kv:Uint32","kv:uint", "kv:Uint":
		var r = make(map[uint32]uint32)
		if s == "" {
			return r, false
		}
		splitOut := strings.Split(s, firstSep)
		for _, sinner := range splitOut {
			if sinner == "" {
				continue
			}
			split := strings.Split(sinner, secondSep)
			var key uint32
			var val uint32
			if k, err := strconv.ParseUint(split[0], 10, 32); err == nil {
				key = uint32(k)
				if _,ok := r[key]; ok {
					panic("key already exists")
				}
			} else {
				panic(err)
			}
			if v, err := strconv.ParseInt(split[1], 10, 32); err == nil {
				val = uint32(v)
			} else {
				panic(err)
			}
			r[key] = val
		}
		return r, false
	case "uint32:int64", "Uint32:Int64", "Uint32:int64", "uint32:Int64":
		var r = make([][]interface{}, 0)
		if s == "" {
			return r, false
		}
		splitOut := strings.Split(s, firstSep)
		for _, sinner := range splitOut {
			var rinner = make([]interface{}, 0)
			if sinner == "" {
				r = append(r, rinner)
				continue
			}
			// key
			split := strings.Split(sinner, secondSep)
			if intVal, err := strconv.ParseUint(split[0], 10, 32); err == nil {
				rinner = append(rinner, uint32(intVal))
			} else {
				panic(err)
			}
			// value
			if intVal, err := strconv.ParseInt(split[1], 10, 64); err == nil {
				rinner = append(rinner, int64(intVal))
			} else {
				panic(err)
			}
			r = append(r, rinner)
		}
		return r, false
	case "int64[][]", "Int64[][]", "int64:int64", "Int64:Int64":
		var r = make([][]int64, 0)
		if s == "" {
			return r, false
		}
		splitOut := strings.Split(s, firstSep)
		for _, sinner := range splitOut {
			if sinner == "" {
				r = append(r, []int64{})
				continue
			}
			split := strings.Split(sinner, secondSep)
			var rinner []int64
			for _, v := range split {
				if intVal, err := strconv.ParseInt(v, 10, 64); err == nil {
					rinner = append(rinner, int64(intVal))
				} else {
					panic(err)
				}
			}
			r = append(r, rinner)
		}
		return r, false
	case "uint64[][]", "Uint64[][]", "uint64:uint64", "Uint64:Uint64":
		var r = make([][]uint64, 0)
		if s == "" {
			return r, false
		}
		splitOut := strings.Split(s, firstSep)
		for _, sinner := range splitOut {
			if sinner == "" {
				r = append(r, []uint64{})
				continue
			}
			split := strings.Split(sinner, secondSep)
			var rinner []uint64
			for _, v := range split {
				if intVal, err := strconv.ParseUint(v, 10, 64); err == nil {
					rinner = append(rinner, uint64(intVal))
				} else {
					panic(err)
				}
			}
			r = append(r, rinner)
		}
		return r, false
	case "string[][]", "String[][]", "string:string", "String:String":
		var r = make([][]string, 0)
		if s == "" {
			return r, false
		}
		splitOut := strings.Split(s, firstSep)
		for _, sinner := range splitOut {
			split := strings.Split(sinner, secondSep)
			r = append(r, split)
		}
		return r, false
	}

	// 默认返回字符串
	return s, true
}

func Maincsv(csvdir, outputdir string) {

	// 1. 读取 CSV 文件
	// 遍历目录
	entries, err := os.ReadDir(csvdir) // 替换为目标目录
	if err != nil {
		panic(err)
	}
	for _, entry := range entries {
		file := entry.Name()
		fmt.Println(file) // 仅文件名，不包含路径
		if strings.HasSuffix(file, ".csv") {
			manifest, packed := parsecsvfile(path.Join(csvdir, file))
			var name = strings.ToLower(file[:len(file)-4])
			// 5. 保存文件
			err := os.MkdirAll(outputdir, 0755)
			if err != nil {
				log.Fatal("创建文件夹失败", err)
			}
			if err := os.WriteFile(fmt.Sprintf("%s/%smanifest.bytes", outputdir, name), manifest, 0644); err != nil {
				log.Fatal("写入文件失败:", err)
			}

			if err := os.WriteFile(fmt.Sprintf("%s/%s.bytes", outputdir, name), packed, 0644); err != nil {
				log.Fatal("写入文件失败:", err)
			}

		}
	}

	log.Println("csv2mspack 转换成功！")
}

func parsecsvfile(filepath string) (m, d []byte) {
	file, err := os.Open(filepath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}

	if len(records) < 1 {
		panic(fmt.Errorf("CSV 文件无数据"))
	}

	// 获取表头和类型
	headers := records[0]
	types := records[1]

	// 确定id列的索引
	idindex := -1
	for i, t := range headers {
		if t == "id" {
			idindex = i
			break
		}
	}
	if idindex == -1 {
		panic(fmt.Errorf("CSV文件中未找到id列"))
	}

	var rows [][]string
	rows = append(rows, records[4:]...)

	// 按idindex排序rows（字典序）
	//slices.SortFunc(rows, func(a, b []string) int {
	//	aID := a[idindex]
	//	bID := b[idindex]
	//	return cmp.Compare(aID, bID) // 直接比较字符串
	//})

	buffer := bytes.Buffer{}
	msgpacker := msgpack.NewEncoder(&buffer)

	err = msgpacker.EncodeArrayLen(len(rows))
	if err != nil {
		panic(err)
	}

	// 解析数据
	var manifest = make([]interface{}, 0)
	for _, row := range rows {
		var item []interface{}
		var id uint32
		for i, val := range row {
			defer func(i int, v string) {
				if err := recover(); err != nil {
					panic(fmt.Sprintf("解析失败:%d列 %s %v\n", i, v, err))
				}
			}(i, val)
			if i >= len(headers) {
				continue // 忽略多余列
			}

			content := adaptContent(val) // 适配原来客户端的逻辑

			if types[i] == "jsonArray" || types[i] == "json" { // 这个还不知道具体的内容格式
				continue
			}

			v, unsupported := inferType(types[i], content)
			if unsupported {
				panic(fmt.Errorf("不支持的类型: %s %s", types[i], content))
			}
			item = append(item, v)
			if i == idindex {
				idu, ok := v.(int32)
				if !ok {
					idstr, err := strconv.Atoi(content)
					if err != nil {
						panic("无法解析id")
					}
					id = uint32(idstr)
				} else {
					id = uint32(idu)
				}
			}
		}

		l := buffer.Len()
		// 序列化为 MessagePack
		err := msgpacker.Encode(item)
		if err != nil {
			panic(fmt.Errorf("序列化失败: %v", err))
		}

		manifest = append(manifest, []interface{}{id, l, buffer.Len() - l})
	}

	// 序列化为 MessagePack
	packedmanifest, err := msgpack.Marshal(manifest)
	if err != nil {
		panic(fmt.Errorf("序列化失败: %v", err))
	}
	fmt.Println(hex.Dump(buffer.Bytes()))
	return packedmanifest, buffer.Bytes()
}
