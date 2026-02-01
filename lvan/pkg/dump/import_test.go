package dump

import (
	"testing"
)

// TestProcessFieldValue_EmptyValue 测试空值处理
//
// 修复场景：Error 1366 (HY000): error: '' is not a valid value for 'float'
// 原因：空字节数组 [] 被转换为空字符串 ""，数值类型不接受空字符串
//
// 修复方案：在 processFieldValue 函数开头检查 len(value) == 0，返回 nil
func TestProcessFieldValue_EmptyValue(t *testing.T) {
	tests := []struct {
		name     string
		value    []byte
		fieldType string
		wantNil  bool
	}{
		// ===== 空值测试 =====
		{
			name:     "empty int value",
			value:    []byte{},
			fieldType: "int",
			wantNil:  true,
		},
		{
			name:     "empty float value",
			value:    []byte{},
			fieldType: "float",
			wantNil:  true,
		},
		{
			name:     "empty double value",
			value:    []byte{},
			fieldType: "double",
			wantNil:  true,
		},
		{
			name:     "empty decimal value",
			value:    []byte{},
			fieldType: "decimal",
			wantNil:  true,
		},
		{
			name:     "empty varchar value",
			value:    []byte{},
			fieldType: "varchar",
			wantNil:  true,
		},
		{
			name:     "empty timestamp value",
			value:    []byte{},
			fieldType: "timestamp",
			wantNil:  true,
		},
		{
			name:     "empty date value",
			value:    []byte{},
			fieldType: "date",
			wantNil:  true,
		},
		{
			name:     "empty blob value",
			value:    []byte{},
			fieldType: "blob",
			wantNil:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := processFieldValue(tt.value, tt.fieldType)
			if (got == nil) != tt.wantNil {
				t.Errorf("processFieldValue() = %v, wantNil %v", got, tt.wantNil)
			}
		})
	}
}

// TestProcessFieldValue_BlobType 测试 BLOB 类型处理
func TestProcessFieldValue_BlobType(t *testing.T) {
	tests := []struct {
		name      string
		value     []byte
		fieldType string
		want      interface{}
	}{
		{
			name:      "binary data",
			value:     []byte{0x00, 0x01, 0x02, 0x03},
			fieldType: "blob",
			want:      []byte{0x00, 0x01, 0x02, 0x03},
		},
		{
			name:      "protobuf data",
			value:     []byte{0x08, 0x96, 0x01, 0x01, 0x0A, 0x07, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6E, 0x67},
			fieldType: "longblob",
			want:      []byte{0x08, 0x96, 0x01, 0x01, 0x0A, 0x07, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6E, 0x67},
		},
		{
			name:      "empty blob",
			value:     []byte{},
			fieldType: "blob",
			want:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := processFieldValue(tt.value, tt.fieldType)
			if tt.want == nil {
				if got != nil {
					t.Errorf("processFieldValue() = %v, want %v", got, tt.want)
				}
			} else {
				if got == nil {
					t.Errorf("processFieldValue() = nil, want %v", tt.want)
					return
				}
				gotBytes, ok := got.([]byte)
				if !ok {
					t.Errorf("processFieldValue() = %T, want []byte", got)
					return
				}
				if string(gotBytes) != string(tt.want.([]byte)) {
					t.Errorf("processFieldValue() = %v, want %v", gotBytes, tt.want)
				}
			}
		})
	}
}

// TestProcessFieldValue_NumericType 测试数值类型处理
func TestProcessFieldValue_NumericType(t *testing.T) {
	tests := []struct {
		name      string
		value     []byte
		fieldType string
		wantNil   bool
	}{
		{
			name:      "valid int value",
			value:     []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			fieldType: "int",
			wantNil:   false,
		},
		{
			name:      "empty int value",
			value:     []byte{},
			fieldType: "int",
			wantNil:   true,
		},
		{
			name:      "empty bigint value",
			value:     []byte{},
			fieldType: "bigint",
			wantNil:   true,
		},
		{
			name:      "empty smallint value",
			value:     []byte{},
			fieldType: "smallint",
			wantNil:   true,
		},
		{
			name:      "empty tinyint value",
			value:     []byte{},
			fieldType: "tinyint",
			wantNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := processFieldValue(tt.value, tt.fieldType)
			if (got == nil) != tt.wantNil {
				t.Errorf("processFieldValue() = %v, wantNil %v", got, tt.wantNil)
			}
		})
	}
}

// TestProcessFieldValue_StringType 测试字符串类型处理
func TestProcessFieldValue_StringType(t *testing.T) {
	tests := []struct {
		name      string
		value     []byte
		fieldType string
		want      string
		wantNil   bool
	}{
		{
			name:      "normal varchar",
			value:     []byte("hello"),
			fieldType: "varchar",
			want:      "hello",
			wantNil:   false,
		},
		{
			name:      "chinese text",
			value:     []byte("测试文本"),
			fieldType: "text",
			want:      "测试文本",
			wantNil:   false,
		},
		{
			name:      "special characters",
			value:     []byte("包含\"引号\"和'单引号'"),
			fieldType: "varchar",
			want:      "包含\"引号\"和'单引号'",
			wantNil:   false,
		},
		{
			name:      "empty varchar",
			value:     []byte{},
			fieldType: "varchar",
			wantNil:   true,
		},
		{
			name:      "empty text",
			value:     []byte{},
			fieldType: "text",
			wantNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := processFieldValue(tt.value, tt.fieldType)
			if tt.wantNil {
				if got != nil {
					t.Errorf("processFieldValue() = %v, want nil", got)
				}
			} else {
				if got == nil {
					t.Errorf("processFieldValue() = nil, want %s", tt.want)
					return
				}
				gotStr, ok := got.(string)
				if !ok {
					t.Errorf("processFieldValue() = %T, want string", got)
					return
				}
				if gotStr != tt.want {
					t.Errorf("processFieldValue() = %q, want %q", gotStr, tt.want)
				}
			}
		})
	}
}

// TestProcessFieldValue_TimeType 测试时间类型处理
func TestProcessFieldValue_TimeType(t *testing.T) {
	tests := []struct {
		name      string
		value     []byte
		fieldType string
		want      string
		wantNil   bool
	}{
		{
			name:      "normal timestamp",
			value:     []byte("2026-01-31 10:50:38"),
			fieldType: "timestamp",
			want:      "2026-01-31 10:50:38",
			wantNil:   false,
		},
		{
			name:      "datetime",
			value:     []byte("2026-01-31 10:50:38"),
			fieldType: "datetime",
			want:      "2026-01-31 10:50:38",
			wantNil:   false,
		},
		{
			name:      "ISO 8601 format",
			value:     []byte("2025-08-19T17:00:52Z"),
			fieldType: "datetime",
			want:      "2025-08-19 17:00:52",
			wantNil:   false,
		},
		{
			name:      "empty timestamp",
			value:     []byte{},
			fieldType: "timestamp",
			wantNil:   true,
		},
		{
			name:      "empty datetime",
			value:     []byte{},
			fieldType: "datetime",
			wantNil:   true,
		},
		{
			name:      "empty date",
			value:     []byte{},
			fieldType: "date",
			wantNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := processFieldValue(tt.value, tt.fieldType)
			if tt.wantNil {
				if got != nil {
					t.Errorf("processFieldValue() = %v, want nil", got)
				}
			} else {
				if got == nil {
					t.Errorf("processFieldValue() = nil, want %s", tt.want)
					return
				}
				gotStr, ok := got.(string)
				if !ok {
					t.Errorf("processFieldValue() = %T, want string", got)
					return
				}
				if gotStr != tt.want {
					t.Errorf("processFieldValue() = %q, want %q", gotStr, tt.want)
				}
			}
		})
	}
}

// TestDecodeTimestampValue 测试时间戳解码
func TestDecodeTimestampValue(t *testing.T) {
	tests := []struct {
		name  string
		value []byte
		want  string
	}{
		{
			name:  "standard datetime format",
			value: []byte("2026-01-31 10:50:38"),
			want:  "2026-01-31 10:50:38",
		},
		{
			name:  "ISO 8601 format with Z",
			value: []byte("2025-08-19T17:00:52Z"),
			want:  "2025-08-19 17:00:52",
		},
		{
			name:  "empty value",
			value: []byte{},
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := decodeTimestampValue(tt.value)
			if tt.want == "" {
				if got != nil {
					t.Errorf("decodeTimestampValue() = %v, want nil", got)
				}
			} else {
				if got == nil {
					t.Errorf("decodeTimestampValue() = nil, want %s", tt.want)
					return
				}
				gotStr, ok := got.(string)
				if !ok {
					t.Errorf("decodeTimestampValue() = %T, want string", got)
					return
				}
				if gotStr != tt.want {
					t.Errorf("decodeTimestampValue() = %q, want %q", gotStr, tt.want)
				}
			}
		})
	}
}

// TestDecodeNumericValue 测试数值类型解码
func TestDecodeNumericValue(t *testing.T) {
	tests := []struct {
		name      string
		value     []byte
		fieldType string
		wantNil   bool
	}{
		{
			name:      "uint64 value",
			value:     []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			fieldType: "bigint",
			wantNil:   false,
		},
		{
			name:      "uint32 value",
			value:     []byte{0x01, 0x00, 0x00, 0x00},
			fieldType: "int",
			wantNil:   false,
		},
		{
			name:      "uint16 value",
			value:     []byte{0x01, 0x00},
			fieldType: "smallint",
			wantNil:   false,
		},
		{
			name:      "uint8 value",
			value:     []byte{0x01},
			fieldType: "tinyint",
			wantNil:   false,
		},
		{
			name:      "empty value",
			value:     []byte{},
			fieldType: "int",
			wantNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := decodeNumericValue(tt.value, tt.fieldType)
			if (got == nil) != tt.wantNil {
				t.Errorf("decodeNumericValue() = %v, wantNil %v", got, tt.wantNil)
			}
		})
	}
}
