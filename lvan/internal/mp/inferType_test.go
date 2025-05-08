package mp

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_inferType(t *testing.T) {
	tests := []struct {
		name       string
		vtype      string
		s          string
		wantResult interface{}
		wantErr    bool
	}{
		// 字符串类型测试
		{"string类型-普通字符串", "string", "test", "test", false},
		{"string类型-空字符串", "string", "", "", false},
		{"string类型-0值", "string", "0", "", false},
		{"String类型-大写", "String", "test", "test", false},
		{"string类型-特殊字符", "string", "test|with|pipes", "test|with|pipes", false},
		{"string类型-数字字符串", "string", "123", "123", false},
		{"长字符串输入", "string", strings.Repeat("a", 100), strings.Repeat("a", 100), false},
		{"混合类型-字符串中带分隔符", "string", "a|b;c", "a|b;c", false},

		// 布尔类型测试
		{"bool类型-true", "bool", "true", true, false},
		{"bool类型-1", "bool", "1", true, false},
		{"bool类型-false", "bool", "false", false, false},
		{"bool类型-0", "bool", "0", false, false},
		{"bool类型-空字符串", "bool", "", false, false},
		{"bool类型-无效值", "bool", "invalid", false, true},
		{"Bool类型-大写", "Bool", "true", true, false},
		{"boolean类型", "boolean", "true", true, false},
		{"Boolean类型", "Boolean", "true", true, false},
		{"bool类型-大写TRUE", "bool", "TRUE", false, true},
		{"bool类型-大写FALSE", "bool", "FALSE", false, true},

		// byte类型测试
		{"byte类型-有效值", "byte", "123", byte(123), false},
		{"byte类型-空字符串", "byte", "", byte(0), false},
		{"Byte类型-大写", "Byte", "123", byte(123), false},
		{"byte类型-最大值", "byte", "255", byte(255), false},

		// int8类型测试
		{"int8类型-有效值", "int8", "123", int8(123), false},
		{"int8类型-空字符串", "int8", "", int8(0), false},
		{"Int8类型-大写", "Int8", "123", int8(123), false},
		{"int8类型-最大值", "int8", "127", int8(127), false},
		{"int8类型-最小值", "int8", "-128", int8(-128), false},

		// uint8类型测试
		{"uint8类型-有效值", "uint8", "123", uint8(123), false},
		{"uint8类型-空字符串", "uint8", "", uint8(0), false},
		{"Uint8类型-大写", "Uint8", "123", uint8(123), false},
		{"uint8类型-最大值", "uint8", "255", uint8(255), false},

		// int16类型测试
		{"int16类型-有效值", "int16", "123", int16(123), false},
		{"int16类型-空字符串", "int16", "", int16(0), false},
		{"Int16类型-大写", "Int16", "123", int16(123), false},
		{"short类型", "short", "123", int16(123), false},
		{"Short类型", "Short", "123", int16(123), false},
		{"int16类型-最大值", "int16", "32767", int16(32767), false},
		{"int16类型-最小值", "int16", "-32768", int16(-32768), false},

		// uint16类型测试
		{"uint16类型-有效值", "uint16", "123", uint16(123), false}, // 注意：代码中这里返回的是int16而不是uint16
		{"uint16类型-空字符串", "uint16", "", uint16(0), false},
		{"Uint16类型-大写", "Uint16", "123", uint16(123), false},
		{"uint16类型-最大值", "uint16", "32767", uint16(32767), false}, // 注意：代码中uint16返回的实际是int16

		// int类型测试
		{"int类型-有效值", "int", "123", 123, false},
		{"int类型-空字符串", "int", "", 0, false},
		{"Int类型-大写", "Int", "123", 123, false},
		{"int类型-大值", "int", "2147483647", 2147483647, false},
		{"int类型-负值", "int", "-2147483648", -2147483648, false},

		// int32类型测试
		{"int32类型-有效值", "int32", "123", int32(123), false},
		{"int32类型-空字符串", "int32", "", int32(0), false},
		{"Int32类型-大写", "Int32", "123", int32(123), false},
		{"int32类型-最大值", "int32", "2147483647", int32(2147483647), false},
		{"int32类型-最小值", "int32", "-2147483648", int32(-2147483648), false},

		// uint32类型测试
		{"uint32类型-有效值", "uint32", "123", uint32(123), false},
		{"uint32类型-空字符串", "uint32", "", uint32(0), false},
		{"Uint32类型-大写", "Uint32", "123", uint32(123), false},
		{"uint32类型-最大值", "uint32", "4294967295", uint32(4294967295), false},

		// int64类型测试
		{"int64类型-有效值", "int64", "123", int64(123), false},
		{"int64类型-空字符串", "int64", "", int64(0), false},
		{"Int64类型-大写", "Int64", "123", int64(123), false},
		{"long类型", "long", "123", int64(123), false},
		{"Long类型", "Long", "123", int64(123), false},
		{"int64类型-大值", "int64", "2147483647", int64(2147483647), false},
		{"int64类型-负值", "int64", "-2147483648", int64(-2147483648), false},

		// uint64类型测试
		{"uint64类型-有效值", "uint64", "123", uint64(123), false},
		{"uint64类型-空字符串", "uint64", "", uint64(0), false},
		{"Uint64类型-大写", "Uint64", "123", uint64(123), false},
		{"uint64类型-最大值", "uint64", "4294967295", uint64(4294967295), false},

		// 字符串数组测试
		{"string[]类型-有效值", "string[]", "a|b|c", []string{"a", "b", "c"}, false},
		{"string[]类型-空字符串", "string[]", "", []string{}, false},
		{"String[]类型-大写", "String[]", "a|b|c", []string{"a", "b", "c"}, false},
		{"string[]类型-单元素", "string[]", "single", []string{"single"}, false},
		{"string[]类型-空元素", "string[]", "a||c", []string{"a", "", "c"}, false},
		{"特殊分隔符-多个竖线", "string[]", "a|||d", []string{"a", "", "", "d"}, false},

		// byte数组测试
		{"byte[]类型-有效值", "byte[]", "1|2|3", []byte{1, 2, 3}, false},
		{"byte[]类型-空字符串", "byte[]", "", []byte{}, false},
		{"Byte[]类型-大写", "Byte[]", "1|2|3", []byte{1, 2, 3}, false},
		{"byte[]类型-单元素", "byte[]", "42", []byte{42}, false},
		{"byte[]类型-边界值", "byte[]", "0|255", []byte{0, 255}, false},

		// int8数组测试
		{"int8[]类型-有效值", "int8[]", "1|2|3", []int8{1, 2, 3}, false},
		{"int8[]类型-空字符串", "int8[]", "", []int8{}, false},
		{"Int8[]类型-大写", "Int8[]", "1|2|3", []int8{1, 2, 3}, false},
		{"int8[]类型-负值", "int8[]", "-1|-128|127", []int8{-1, -128, 127}, false},

		// uint8数组测试
		{"uint8[]类型-有效值", "uint8[]", "1|2|3", []uint8{1, 2, 3}, false}, // 注意：代码中返回的是[]int8
		{"uint8[]类型-空字符串", "uint8[]", "", []uint8{}, false},
		{"Uint8[]类型-大写", "Uint8[]", "1|2|3", []uint8{1, 2, 3}, false},
		{"uint8[]类型-边界值", "uint8[]", "0|255", []uint8{0, 255}, false}, // 255会被截断为-1

		// int16数组测试
		{"int16[]类型-有效值", "int16[]", "1|2|3", []int16{1, 2, 3}, false},
		{"int16[]类型-空字符串", "int16[]", "", []int16{}, false},
		{"Int16[]类型-大写", "Int16[]", "1|2|3", []int16{1, 2, 3}, false},
		{"int16[]类型-负值", "int16[]", "-1|-32768|32767", []int16{-1, -32768, 32767}, false},

		// uint16数组测试
		{"uint16[]类型-有效值", "uint16[]", "1|2|3", []uint16{1, 2, 3}, false},
		{"uint16[]类型-空字符串", "uint16[]", "", []uint16{}, false},
		{"Uint16[]类型-大写", "Uint16[]", "1|2|3", []uint16{1, 2, 3}, false},
		{"uint16[]类型-边界值", "uint16[]", "0|65535", []uint16{0, 65535}, false},

		// int32数组测试
		{"int32[]类型-有效值", "int32[]", "1|2|3", []int32{1, 2, 3}, false},
		{"int32[]类型-空字符串", "int32[]", "", []int32{}, false},
		{"Int32[]类型-大写", "Int32[]", "1|2|3", []int32{1, 2, 3}, false},
		{"int32[]类型-负值", "int32[]", "-1|-2147483648|2147483647", []int32{-1, -2147483648, 2147483647}, false},

		// uint32数组测试
		{"uint32[]类型-有效值", "uint32[]", "1|2|3", []uint32{1, 2, 3}, false},
		{"uint32[]类型-空字符串", "uint32[]", "", []uint32{}, false},
		{"Uint32[]类型-大写", "Uint32[]", "1|2|3", []uint32{1, 2, 3}, false},
		{"uint32[]类型-边界值", "uint32[]", "0|4294967295", []uint32{0, 4294967295}, false},

		// int64数组测试
		{"int64[]类型-有效值", "int64[]", "1|2|3", []int64{1, 2, 3}, false},
		{"int64[]类型-空字符串", "int64[]", "", []int64{}, false},
		{"Int64[]类型-大写", "Int64[]", "1|2|3", []int64{1, 2, 3}, false},
		{"int64[]类型-负值", "int64[]", "-1|-9223372036854775808|9223372036854775807", []int64{-1, -9223372036854775808, 9223372036854775807}, false},

		// uint64数组测试
		{"uint64[]类型-有效值", "uint64[]", "1|2|3", []uint64{1, 2, 3}, false},
		{"uint64[]类型-空字符串", "uint64[]", "", []uint64{}, false},
		{"Uint64[]类型-大写", "Uint64[]", "1|2|3", []uint64{1, 2, 3}, false},
		{"uint64[]类型-边界值", "uint64[]", "0|18446744073709551615", []uint64{0, 18446744073709551615}, false},

		// int32二维数组测试
		{"int32[][]类型-有效值", "int32[][]", "1|2;3|4", [][]int32{{1, 2}, {3, 4}}, false},
		{"int32[][]类型-空字符串", "int32[][]", "", [][]int32{}, false},
		{"Int32[][]类型-大写", "Int32[][]", "1|2;3|4", [][]int32{{1, 2}, {3, 4}}, false},
		{"int32:int32类型", "int32:int32", "1|2;3|4", [][]int32{{1, 2}, {3, 4}}, false},
		{"Int32:Int32类型", "Int32:Int32", "1|2;3|4", [][]int32{{1, 2}, {3, 4}}, false},
		{"int32[][]类型-单元素", "int32[][]", "42", [][]int32{{42}}, false},
		{"int32[][]类型-空行", "int32[][]", "1|2;;", [][]int32{{1, 2}, {}, {}}, false},
		{"int32[][]类型-不规则", "int32[][]", "1|2|3;4|5;", [][]int32{{1, 2, 3}, {4, 5}, {}}, false},
		{"特殊分隔符-多个分号", "int32[][]", "1|2;;;3|4", [][]int32{{1, 2}, {}, {}, {3, 4}}, false},

		// uint32二维数组测试
		{"uint32[][]类型-有效值", "uint32[][]", "1|2;3|4", [][]uint32{{1, 2}, {3, 4}}, false},
		{"uint32[][]类型-空字符串", "uint32[][]", "", [][]uint32{}, false},
		{"uint32[][]类型-空内部数组", "uint32[][]", "1|2;;", [][]uint32{{1, 2}, {}, {}}, false},
		{"Uint32[][]类型-大写", "Uint32[][]", "1|2;3|4", [][]uint32{{1, 2}, {3, 4}}, false},
		{"uint32:uint32类型", "uint32:uint32", "1|2;3|4", [][]uint32{{1, 2}, {3, 4}}, false},
		{"Uint32:Uint32类型", "Uint32:Uint32", "1|2;3|4", [][]uint32{{1, 2}, {3, 4}}, false},
		{"uint32[][]类型-单元素", "uint32[][]", "42", [][]uint32{{42}}, false},

		// kv:string类型测试
		{"kv:string类型-有效值", "kv:string", "1|a;2|b", map[uint32]string{1: "a", 2: "b"}, false},
		{"kv:string类型-空字符串", "kv:string", "", map[uint32]string{}, false},
		{"kv:string类型-空内部元素", "kv:string", "1|a;;", map[uint32]string{1: "a"}, false},
		{"kv:String类型-大写", "kv:String", "1|a;2|b", map[uint32]string{1: "a", 2: "b"}, false},
		{"kv:string类型-单对", "kv:string", "1|a", map[uint32]string{1: "a"}, false},
		{"kv:string类型-单对-空值", "kv:string", "1|", map[uint32]string{1: ""}, false},
		{name: "kv:string类型-单对-空键", vtype: "kv:string", s: "|a", wantResult: map[uint32]string{}, wantErr: true},
		// kv:bool类型测试
		{"kv:bool类型-有效值", "kv:bool", "1|true;2|false", map[uint32]bool{1: true, 2: false}, false},
		{"kv:bool类型-空字符串", "kv:bool", "", map[uint32]bool{}, false},
		{"kv:Bool类型-大写", "kv:Bool", "1|true;2|false", map[uint32]bool{1: true, 2: false}, false},
		{"kv:bool类型-单对", "kv:bool", "1|true", map[uint32]bool{1: true}, false},
		{"kv:bool类型-单对-空值", "kv:bool", "1|", map[uint32]bool{1: false}, false},
		{"kv:bool类型-单对-空键", "kv:bool", "|true", map[uint32]bool{}, true},

		// kv:int类型测试
		{"kv:int类型-有效值", "kv:int", "1|2;3|4", map[uint32]int32{1: 2, 3: 4}, false},
		{"kv:int类型-空字符串", "kv:int", "", map[uint32]int32{}, false},
		{"kv:int类型-空内部元素", "kv:int", "1|2;;", map[uint32]int32{1: 2}, false},
		{"kv:Int类型-大写", "kv:Int", "1|2;3|4", map[uint32]int32{1: 2, 3: 4}, false},
		{"kv:int类型-单对", "kv:int", "1|2", map[uint32]int32{1: 2}, false},
		{"kv:int类型-单对-空值", "kv:int", "1|", map[uint32]int32{}, true},
		{"kv:int类型-单对-空键", "kv:int", "|2", map[uint32]int32{}, true},

		// kv:uint类型测试
		{"kv:uint类型-有效值", "kv:uint", "1|2;3|4", map[uint32]uint32{1: 2, 3: 4}, false},
		{"kv:uint类型-空字符串", "kv:uint", "", map[uint32]uint32{}, false},
		{"kv:uint类型-空内部元素", "kv:uint", "1|2;;", map[uint32]uint32{1: 2}, false},
		{"kv:Uint类型-大写", "kv:Uint", "1|2;3|4", map[uint32]uint32{1: 2, 3: 4}, false},
		{"kv:uint类型-单对", "kv:uint", "1|2", map[uint32]uint32{1: 2}, false},
		{"kv:uint类型-单对-空值", "kv:uint", "1|", map[uint32]uint32{}, true},
		{"kv:uint类型-单对-空键", "kv:uint", "|2", map[uint32]uint32{}, true},

		// kv:int32类型测试
		{"kv:int32类型-有效值", "kv:int32", "1|2;3|4", map[uint32]int32{1: 2, 3: 4}, false},
		{"kv:int32类型-空字符串", "kv:int32", "", map[uint32]int32{}, false},
		{"kv:int32类型-空内部元素", "kv:int32", "1|2;;", map[uint32]int32{1: 2}, false},
		{"kv:Int32类型-大写", "kv:Int32", "1|2;3|4", map[uint32]int32{1: 2, 3: 4}, false},
		{"kv:int32类型-单对", "kv:int32", "1|2", map[uint32]int32{1: 2}, false},
		{"kv:int32类型-单对-空值", "kv:int32", "1|", map[uint32]int32{}, true},
		{"kv:int32类型-单对-空键", "kv:int32", "|2", map[uint32]int32{}, true},

		// kv:uint32类型测试
		{"kv:uint32类型-有效值", "kv:uint32", "1|2;3|4", map[uint32]uint32{1: 2, 3: 4}, false},
		{"kv:uint32类型-空字符串", "kv:uint32", "", map[uint32]uint32{}, false},
		{"kv:uint32类型-空内部元素", "kv:uint32", "1|2;;", map[uint32]uint32{1: 2}, false},
		{"kv:Uint32类型-大写", "kv:Uint32", "1|2;3|4", map[uint32]uint32{1: 2, 3: 4}, false},
		{"kv:uint32类型-单对", "kv:uint32", "1|2", map[uint32]uint32{1: 2}, false},
		{"kv:uint32类型-多对", "kv:uint32", "1|2;3|4;5|6", map[uint32]uint32{1: 2, 3: 4, 5: 6}, false},

		// uint32:int64类型测试
		{"uint32:int64类型-有效值", "uint32:int64", "1|2;3|4", [][]interface{}{{uint32(1), int64(2)}, {uint32(3), int64(4)}}, false},
		{"uint32:int64类型-空字符串", "uint32:int64", "", [][]interface{}{}, false},
		{"uint32:int64类型-空内部元素", "uint32:int64", "1|2;;", [][]interface{}{{uint32(1), int64(2)}, {}, {}}, false},
		{"Uint32:Int64类型-大写", "Uint32:Int64", "1|2;3|4", [][]interface{}{{uint32(1), int64(2)}, {uint32(3), int64(4)}}, false},
		{"uint32:int64类型-单对", "uint32:int64", "1|2", [][]interface{}{{uint32(1), int64(2)}}, false},
		{"uint32:int64类型-多对", "uint32:int64", "1|2;3|4;5|6", [][]interface{}{{uint32(1), int64(2)}, {uint32(3), int64(4)}, {uint32(5), int64(6)}}, false},
		{"uint32:int64类型-负值", "uint32:int64", "1|-2", [][]interface{}{{uint32(1), int64(-2)}}, false},
		{
			name:  "uint32:int64混合类型",
			vtype: "uint32:int64",
			s:     "100|200;300|400",
			wantResult: [][]interface{}{
				{uint32(100), int64(200)},
				{uint32(300), int64(400)},
			},
			wantErr: false,
		},

		// int64[][]类型测试
		{"int64[][]类型-有效值", "int64[][]", "1|2;3|4", [][]int64{{1, 2}, {3, 4}}, false},
		{"int64[][]类型-空字符串", "int64[][]", "", [][]int64{}, false},
		{"Int64[][]类型-大写", "Int64[][]", "1|2;3|4", [][]int64{{1, 2}, {3, 4}}, false},
		{"int64:int64类型", "int64:int64", "1|2;3|4", [][]int64{{1, 2}, {3, 4}}, false},
		{"Int64:Int64类型", "Int64:Int64", "1|2;3|4", [][]int64{{1, 2}, {3, 4}}, false},
		{"int64[][]类型-单元素", "int64[][]", "42", [][]int64{{42}}, false},
		{"int64[][]类型-空行", "int64[][]", "1|2;;", [][]int64{{1, 2}, {}, {}}, false},
		{
			name:       "int64[][]正常解析",
			vtype:      "int64[][]",
			s:          "1|2;3|4;5|6",
			wantResult: [][]int64{{1, 2}, {3, 4}, {5, 6}},
			wantErr:    false,
		},
		{
			name:       "空值处理",
			vtype:      "Int64:Int64",
			s:          "",
			wantResult: [][]int64{},
			wantErr:    false,
		},

		// uint64[][]类型测试
		{"uint64[][]类型-有效值", "uint64[][]", "1|2;3|4", [][]uint64{{1, 2}, {3, 4}}, false},
		{"uint64[][]类型-空字符串", "uint64[][]", "", [][]uint64{}, false},
		{"Uint64[][]类型-大写", "Uint64[][]", "1|2;3|4", [][]uint64{{1, 2}, {3, 4}}, false},
		{"uint64:uint64类型", "uint64:uint64", "1|2;3|4", [][]uint64{{1, 2}, {3, 4}}, false},
		{"Uint64:Uint64类型", "Uint64:Uint64", "1|2;3|4", [][]uint64{{1, 2}, {3, 4}}, false},
		{"uint64[][]类型-单元素", "uint64[][]", "42", [][]uint64{{42}}, false},
		{"uint64[][]类型-空行", "uint64[][]", "1|2;;", [][]uint64{{1, 2}, {}, {}}, false},

		// string[][]类型测试
		{"string[][]类型-有效值", "string[][]", "a|b;c|d", [][]string{{"a", "b"}, {"c", "d"}}, false},
		{"string[][]类型-空字符串", "string[][]", "", [][]string{}, false},
		{"String[][]类型-大写", "String[][]", "a|b;c|d", [][]string{{"a", "b"}, {"c", "d"}}, false},
		{"string:string类型", "string:string", "a|b;c|d", [][]string{{"a", "b"}, {"c", "d"}}, false},
		{"String:String类型", "String:String", "a|b;c|d", [][]string{{"a", "b"}, {"c", "d"}}, false},
		{"string[][]类型-单元素", "string[][]", "text", [][]string{{"text"}}, false},
		{"string[][]类型-空元素", "string[][]", "a||c;d|e", [][]string{{"a", "", "c"}, {"d", "e"}}, false},
		{"string[][]类型-空行", "string[][]", "a|b|c;;", [][]string{{"a", "b", "c"}, {""}, {""}}, false},

		{"kv:string类型-有效值", "kv:string", "1|a;2|b", map[uint32]string{1: "a", 2: "b"}, false},
		{"kv:string类型-空字符串", "kv:string", "", map[uint32]string{}, false},
		{"kv:string类型-空内部元素", "kv:string", "1|a;;", map[uint32]string{1: "a"}, false},
		{"kv:String类型-大写", "kv:String", "1|a;2|b", map[uint32]string{1: "a", 2: "b"}, false},
		{"kv:string类型-单对", "kv:string", "1|a", map[uint32]string{1: "a"}, false},

		// 未知类型测试
		{"未知类型", "unknown", "test", "test", true},
		{"空类型名称", "", "test", "test", true},
		{"非常规类型名称-带空格", "int 32", "123", "123", true},
		{"非常规类型名称-带特殊字符", "int#32", "123", "123", true},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s %s %s", tt.name, tt.vtype, tt.s), func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.wantErr {
						t.Errorf("inferType() panic = %v, wantErr %v", r, tt.wantErr)
					}
				}
			}()

			gotResult, gotErr := inferType(tt.vtype, tt.s)
			if gotErr != tt.wantErr {
				t.Errorf("inferType() error = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("inferType() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

// 测试panic情况
func Test_inferType_Panic(t *testing.T) {
	// 测试各种类型的无效输入导致的panic
	panicTests := []struct {
		name  string
		vtype string
		s     string
	}{
		{"byte类型-无效值", "byte", "invalid"},
		{"int8类型-无效值", "int8", "invalid"},
		{"uint8类型-无效值", "uint8", "invalid"},
		{"int16类型-无效值", "int16", "invalid"},
		{"uint16类型-无效值", "uint16", "invalid"},
		{"int类型-无效值", "int", "invalid"},
		{"int32类型-无效值", "int32", "invalid"},
		{"uint32类型-无效值", "uint32", "invalid"},
		{"int64类型-无效值", "int64", "invalid"},
		{"uint64类型-无效值", "uint64", "invalid"},
		{"byte[]类型-无效值", "byte[]", "1|invalid"},
		{"int8[]类型-无效值", "int8[]", "1|invalid"},
		{"int16[]类型-无效值", "int16[]", "1|invalid"},
		{"uint16[]类型-无效值", "uint16[]", "1|invalid"},
		{"int32[]类型-无效值", "int32[]", "1|invalid"},
		{"uint32[]类型-无效值", "uint32[]", "1|invalid"},
		{"int64[]类型-无效值", "int64[]", "1|invalid"},
		{"uint64[]类型-无效值", "uint64[]", "1|invalid"},
		{"int32[][]类型-无效值", "int32[][]", "1|2;invalid"},
		{"uint32[][]类型-无效值", "uint32[][]", "1|2;invalid"},
		{"kv:uint32类型-无效键", "kv:uint32", "invalid|2"},
		{"kv:uint32类型-无效值", "kv:uint32", "1|invalid"},

		// From inferType_additional_test.go
		{"int64[][]类型-无效值", "int64[][]", "1|2;invalid"},
		{"uint64[][]类型-无效值", "uint64[][]", "1|2;invalid"},
		{"uint32:int64类型-无效键", "uint32:int64", "invalid|2"},
		{"uint32:int64类型-无效值", "uint32:int64", "1|invalid"},

		// From inferType_complete_test.go
		{"byte[]类型-无效值", "byte[]", "invalid"}, // 重复，但保持完整性
		{"int8[]类型-无效值", "int8[]", "invalid"}, // 重复
		{"uint8[]类型-无效值", "uint8[]", "invalid"},
		{"int16[]类型-无效值", "int16[]", "invalid"},   // 重复
		{"uint16[]类型-无效值", "uint16[]", "invalid"}, // 重复
		{"int32[]类型-无效值", "int32[]", "invalid"},   // 重复
		{"uint32[]类型-无效值", "uint32[]", "invalid"}, // 重复
		{"int64[]类型-无效值", "int64[]", "invalid"},   // 重复
		{"uint64[]类型-无效值", "uint64[]", "invalid"}, // 重复

		{"int32[][]类型-无效值", "int32[][]", "a|b;c|invalid"},
		{"uint32[][]类型-无效值", "uint32[][]", "a|b;c|invalid"},
		{"int64[][]类型-无效值", "int64[][]", "a|b;c|invalid"},   // 重复
		{"uint64[][]类型-无效值", "uint64[][]", "a|b;c|invalid"}, // 重复

		{"kv:uint32类型-无效键", "kv:uint32", "invalid|2"},
		{"kv:uint32类型-无效值", "kv:uint32", "1|invalid"},

		{"uint32:int64类型-无效键", "uint32:int64", "invalid|2"},
		{"uint32:int64类型-无效值", "uint32:int64", "1|invalid"},

		{"kv:uint32类型-无效键", "kv:uint32", "invalid|2"},
		{"kv:uint32类型-无效值", "kv:uint32", "1|invalid"},
		{"kv:uint32类型-重复键", "kv:uint32", "1|1;1|2"},
	}

	for _, tt := range panicTests {
		t.Run(fmt.Sprintf("%s %s %s", tt.name, tt.vtype, tt.s), func(t *testing.T) {
			assert.Panics(t, func() {
				inferType(tt.vtype, tt.s)
			}, "期望发生panic但没有发生")
		})
	}
}
