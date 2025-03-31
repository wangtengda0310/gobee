package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/vmihailenco/msgpack/v5"
	"io"
	"log"
	"os"
	"path"
	"strings"
)

func main() {
	// 1. 读取 CSV 文件
	csvdir := os.Args[1]
	// 遍历目录
	entries, err := os.ReadDir(csvdir) // 替换为目标目录
	if err != nil {
		panic(err)
	}
	for _, entry := range entries {
		file := entry.Name()
		fmt.Println(file) // 仅文件名，不包含路径
		if strings.HasSuffix(file, ".json") {
			manifest, packed := parsejosnfile(path.Join(csvdir, file))
			var name = strings.ToLower(file[:len(file)-5])
			// 5. 保存文件
			os.MkdirAll("output", 0755)
			if err := os.WriteFile(fmt.Sprintf("%s/%smanifest.bytes", os.Args[2], name), manifest, 0644); err != nil {
				log.Fatal("写入文件失败:", err)
			}

			if err := os.WriteFile(fmt.Sprintf("%s/%s.bytes", os.Args[2], name), packed, 0644); err != nil {
				log.Fatal("写入文件失败:", err)
			}

		}
	}

	log.Println("转换成功！")
}

func parsejosnfile(filepath string) (m, d []byte) {
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
	var manifests []interface{}
	for _, row := range v {
		manifest := packRecord(msgpacker, buffer, row)
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

func packRecord(msgpacker *msgpack.Encoder, buffer *bytes.Buffer, v interface{}) []interface{} {
	var id uint32
	id = uint32(v.(map[string]interface{})["id"].(float64))

	l := buffer.Len()

	if err := pack(msgpacker, v); err != nil {
		log.Fatal("序列化失败:", err)
	}
	return []interface{}{id, l, buffer.Len() - l}
}
func pack(msgpacker *msgpack.Encoder, v interface{}) error {
	switch v.(type) {
	case string:

		if err := msgpacker.EncodeString(v.(string)); err != nil {
			return err
		}
	case []interface{}:

		if err := msgpacker.EncodeArrayLen(len(v.([]interface{}))); err != nil {
			return err
		}
		for _, vinner := range v.([]interface{}) {

			if err := pack(msgpacker, vinner); err != nil {
				return err
			}
		}
	case map[string]interface{}:

		if err := msgpacker.EncodeMapLen(len(v.(map[string]interface{}))); err != nil {
			return err
		}
		for k, vinner := range v.(map[string]interface{}) {

			if err := msgpacker.EncodeString(k); err != nil {
				return err
			}

			if err := pack(msgpacker, vinner); err != nil {
				return err
			}
		}
	}
	return nil
}
