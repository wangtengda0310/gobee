package sign

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"reflect"
	"sort"
)

func Md5Apply(key string, params any) {
	md5Sum, accessValue := md5Field(key, params)
	if !accessValue.CanSet() {
		panic("无法设置签名字段，可能字段不可导出")
	}
	accessValue.SetString(md5Sum)
}
func Md5(key string, params any) string {
	md5Sum, _ := md5Field(key, params)
	return md5Sum
}
func md5Field(key string, params any) (string, reflect.Value) {
	v := reflect.ValueOf(params)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()

	var accessField *reflect.StructField
	paramsList := make([]struct {
		Name  string
		Value string
	}, 0)

	// 遍历结构体字段，收集需要参与计算的字段和查找access字段
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("sign")

		// 查找签名存储字段
		if tag == "access" {
			accessField = &field
		}

		// 处理参与签名的字段
		if tag == "yes" {
			fieldValue := v.Field(i)
			paramsList = append(paramsList, struct {
				Name  string
				Value string
			}{
				Name:  field.Name,
				Value: fmt.Sprint(fieldValue.Interface()),
			})
		}
	}

	if accessField == nil {
		panic("结构体中未找到 sign=\"access\" 的字段")
	}

	// 按字段名字典序排序
	sort.Slice(paramsList, func(i, j int) bool {
		return paramsList[i].Name < paramsList[j].Name
	})

	// 拼接参数字符串
	paramStr := ""
	for _, param := range paramsList {
		if paramStr != "" {
			paramStr += "&"
		}
		paramStr += fmt.Sprintf("%s=%s", param.Name, param.Value)
	}

	// 拼接密钥并计算MD5
	fullStr := paramStr + "&key=" + key
	hasher := md5.New()
	hasher.Write([]byte(fullStr))
	md5Sum := hex.EncodeToString(hasher.Sum(nil))
	// 设置签名结果到目标字段
	accessValue := v.FieldByIndex(accessField.Index)
	return md5Sum, accessValue
}
