package reflectparam

import (
	"reflect"
	"strconv"
	"strings"
)

type Args struct {
	S string `csv:"name"`
	V int    `csv:"value"`
}

// Parse 使用反射将字符串解析到结构体
// 例如: "asdf 123" -> Args{S: "asdf", V: 123}
func Parse(s string) Args {
	parts := strings.Split(s, " ")
	result := Args{}

	// 获取结构体的反射值
	rv := reflect.ValueOf(&result).Elem()

	// 设置第一个字段 S (string)
	if len(parts) > 0 {
		rv.Field(0).SetString(parts[0])
	}

	// 设置第二个字段 V (int)
	if len(parts) > 1 {
		val, err := strconv.Atoi(parts[1])
		if err == nil {
			rv.Field(1).SetInt(int64(val))
		}
	}

	return result
}

// ParseGeneric 使用泛型解析字符串到任意结构体
func ParseGeneric[T any](s string, factory func() T) T {
	instance := factory()
	rv := reflect.ValueOf(&instance).Elem()

	parts := strings.Split(s, " ")

	for i := 0; i < rv.NumField() && i < len(parts); i++ {
		field := rv.Field(i)
		switch field.Kind() {
		case reflect.String:
			field.SetString(parts[i])
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if val, err := strconv.Atoi(parts[i]); err == nil {
				field.SetInt(int64(val))
			}
		}
	}

	return instance
}
