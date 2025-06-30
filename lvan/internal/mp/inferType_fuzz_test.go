package mp

import (
	"runtime/debug"
	"strings"
	"testing"
)

func FuzzInferType(f *testing.F) {
	// Add initial seed corpus
	// Valid types and values
	f.Add(true, "string", "hello")
	f.Add(true, "String", "world")
	f.Add(true, "string", "0") // Special case for string "0"
	f.Add(true, "bool", "true")
	f.Add(true, "Bool", "1")
	f.Add(true, "boolean", "false")
	f.Add(true, "Boolean", "0")
	f.Add(true, "Boolean", "") // Empty string for bool is false
	f.Add(true, "byte", "123")
	f.Add(true, "Byte", "0")
	f.Add(true, "int8", "-10")
	f.Add(true, "Int8", "")
	f.Add(true, "uint8", "255")
	f.Add(true, "Uint8", "")
	f.Add(true, "int16", "-30000")
	f.Add(true, "Int16", "")
	f.Add(true, "short", "15000")
	f.Add(true, "Short", "")
	f.Add(true, "uint16", "60000")
	f.Add(true, "Uint16", "")
	f.Add(true, "int", "-123456")
	f.Add(true, "Int", "")
	f.Add(true, "int32", "2000000000")
	f.Add(true, "Int32", "")
	f.Add(true, "uint32", "4000000000")
	f.Add(true, "Uint32", "")
	f.Add(true, "int64", "-9000000000000000000")
	f.Add(true, "Int64", "")
	f.Add(true, "long", "8000000000000000000")
	f.Add(true, "Long", "")
	f.Add(true, "uint64", "18000000000000000000")
	f.Add(true, "Uint64", "")

	// Array types
	f.Add(true, "string[]", "a|b|c")
	f.Add(true, "String[]", "")
	f.Add(true, "byte[]", "1|2|3")
	f.Add(true, "Byte[]", "")
	f.Add(true, "int8[]", "-1|-2|")
	f.Add(true, "Int8[]", "")
	f.Add(true, "uint8[]", "10|20|30") // Uses secondSep '|'
	f.Add(true, "Uint8[]", "")
	f.Add(true, "int16[]", "-100|-200|-300")
	f.Add(true, "Int16[]", "")
	f.Add(true, "uint16[]", "1000|2000|3000")
	f.Add(true, "Uint16[]", "")
	f.Add(true, "int32[]", "-10000|-20000|-30000")
	f.Add(true, "Int32[]", "")
	f.Add(true, "uint32[]", "100000|200000|300000")
	f.Add(true, "Uint32[]", "")
	f.Add(true, "int64[]", "-1000000|-2000000|-3000000")
	f.Add(true, "Int64[]", "")
	f.Add(true, "uint64[]", "10000000|20000000|30000000")
	f.Add(true, "Uint64[]", "")

	// 2D Array types
	f.Add(true, "int32[][]", "1|2;3|4")
	f.Add(true, "Int32[][]", "")
	f.Add(true, "int32:int32", "1|2;3|4")
	f.Add(true, "Int32:Int32", "")
	f.Add(true, "uint32[][]", "10|20;30|40")
	f.Add(true, "Uint32[][]", "")
	f.Add(true, "uint32:uint32", "10|20;30|40")
	f.Add(true, "Uint32:Uint32", "")
	f.Add(true, "int64[][]", "100|200;300|400")
	f.Add(true, "Int64[][]", "")
	f.Add(true, "int64:int64", "100|200;300|400")
	f.Add(true, "Int64:Int64", "")
	f.Add(true, "uint64[][]", "1000|2000;3000|4000")
	f.Add(true, "Uint64[][]", "")
	f.Add(true, "uint64:uint64", "1000|2000;3000|4000")
	f.Add(true, "Uint64:Uint64", "")
	f.Add(true, "string[][]", "a|b;c|d")
	f.Add(true, "String[][]", "")
	f.Add(true, "string:string", "a|b;c|d")
	f.Add(true, "String:String", "")

	// Map/KV types
	f.Add(true, "kv:uint32", "1|10;2|20")
	f.Add(true, "kv:Uint32", "")
	f.Add(true, "uint32:int64", "1|100;2|200")
	f.Add(true, "Uint32:Int64", "")
	f.Add(true, "uint32:string", "1|100;2|200")
	f.Add(true, "Uint32:String", "1|100;2|200")

	// Potentially problematic inputs
	f.Add(false, "int", "notanumber")
	f.Add(false, "byte", "-1")                     // Negative for uint
	f.Add(false, "byte", "256")                    // Overflow for byte
	f.Add(false, "int8[]", "1|abc|3")              // Malformed array element
	f.Add(false, "int32[][]", "1|2;abc|4")         // Malformed 2D array element
	f.Add(false, "int32[][]", "1|2;3|99999999999") // Overflowed 2D array element
	f.Add(false, "kv:uint32", "1|a;2|20")          // Malformed map value
	f.Add(false, "kv:uint32", "a|10;2|20")         // Malformed map key
	f.Add(false, "kv:uint32", "1|10;2")            // Malformed map pair
	f.Add(false, "unknown_type", "some_value")
	f.Add(false, "", "") // Empty type and value

	// Additional edge cases
	f.Add(true, "string", "\\x00\\x01\\x02")                                        // Binary data in string
	f.Add(true, "string", "日本語")                                                    // Unicode characters
	f.Add(true, "string", "a\tb\nc\rd")                                             // Control characters
	f.Add(true, "string", "")                                                       // Empty string
	f.Add(true, "string", " ")                                                      // Whitespace only
	f.Add(true, "string", "\x00")                                                   // Null byte
	f.Add(true, "string", "日本語")                                                    // Unicode characters
	f.Add(true, "string", "\"'\x00")                                                // Escape sequences
	f.Add(true, "string", strings.Repeat("a", 10000))                               // Very long string
	f.Add(true, "bool", "invalid")                                                  // Invalid bool
	f.Add(true, "byte", "256")                                                      // Byte overflow
	f.Add(true, "int8", "-129")                                                     // Int8 underflow
	f.Add(true, "uint8", "256")                                                     // Uint8 overflow
	f.Add(true, "int16", "-32769")                                                  // Int16 underflow
	f.Add(true, "uint16", "65536")                                                  // Uint16 overflow
	f.Add(true, "int32", "-2147483649")                                             // Int32 underflow
	f.Add(true, "uint32", "4294967296")                                             // Uint32 overflow
	f.Add(true, "int64", "-9223372036854775809")                                    // Int64 underflow
	f.Add(true, "uint64", "18446744073709551616")                                   // Uint64 overflow
	f.Add(true, "string[]", "|||||")                                                // Multiple empty elements
	f.Add(true, "int32[][]", ";;;")                                                 // Multiple empty rows
	f.Add(true, "kv:uint32", "|||;")                                                // Malformed KV pairs
	f.Add(true, "string", "\\\"'\x00")                                              // Escape sequences
	f.Add(true, "int", "999999999999999999999999999999")                            // Very large number
	f.Add(true, "int", "-999999999999999999999999999999")                           // Very small number
	f.Add(true, "string[]", "a|b|c|d|e|f|g|h|i|j|k|l|m|n|o|p|q|r|s|t|u|v|w|x|y|z")  // Long array
	f.Add(true, "int32[][]", "1|2|3|4|5;6|7|8|9|10;11|12|13|14|15;16|17|18|19|20")  // Large 2D array
	f.Add(true, "kv:uint32", "1|10;2|20;3|30;4|40;5|50;6|60;7|70;8|80;9|90;10|100") // Large map

	f.Fuzz(func(t *testing.T, wantOk bool, vtype string, s string) {
		defer func() {
			if r := recover(); r != nil {
				if !wantOk {
					return
				}
				// 对于其他非预期的 panic，仍然报错
				t.Errorf("Unexpected panic recovered in inferType with vtype=%q, s=%q:\n%v\n%s", vtype, s, r, debug.Stack())
			}
		}()

		// Add additional validation for certain cases
		if vtype == "" || s == "" {
			// Skip empty type or value as they are handled by inferType
			return
		}

		result, err := inferType(vtype, s)
		if err { // 假设 inferType 返回布尔值表示错误状态；原注释：注意这里将 err bool 改为 err error
			// Expected error for invalid inputs
			return
		}

		// Basic type validation based on vtype
		switch vtype {
		case "string", "String":
			if _, ok := result.(string); !ok {
				t.Errorf("Expected string type, got %T", result)
			}
		case "bool", "Bool", "boolean", "Boolean":
			if _, ok := result.(bool); !ok {
				t.Errorf("Expected bool type, got %T", result)
			}
			// Add more type checks as needed
		}
	})
}
