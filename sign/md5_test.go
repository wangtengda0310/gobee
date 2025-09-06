package sign

import (
	"crypto/md5"
	"fmt"
	"reflect"
	"sort"
	"testing"
	"testing/quick"
)

// 单元测试
func TestSign(t *testing.T) {
	tests := []struct {
		name      string
		model     interface{}
		key       string
		expected  string
		expectErr bool
	}{
		{
			name: "normal case",
			model: &struct {
				OrderID   string `sign:"yes"`
				UserID    int    `sign:"yes"`
				Secret    string `sign:"no"`
				Signature string `sign:"access"`
			}{
				OrderID: "ORDER123",
				UserID:  98765,
				Secret:  "private",
			},
			key:       "API_SECRET",
			expected:  "d41d8cd98f00b204e9800998ecf8427e", // 空值测试需要修改实际值
			expectErr: false,
		},
		{
			name: "with special characters",
			model: &struct {
				Data string `sign:"yes"`
				Flag bool   `sign:"yes"`
				Sig  string `sign:"access"`
			}{
				Data: "hello 世界!",
				Flag: true,
			},
			key:       "test",
			expected:  "", // 需要计算实际值
			expectErr: false,
		},
		{
			name: "missing access field",
			model: &struct {
				Name string `sign:"yes"`
			}{},
			key:       "test",
			expected:  "",
			expectErr: true,
		},
		{
			name: "empty values",
			model: &struct {
				A string `sign:"yes"`
				B int    `sign:"yes"`
				S string `sign:"access"`
			}{},
			key:       "empty",
			expected:  "", // 需要计算实际值
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil && !tt.expectErr {
					t.Errorf("unexpected panic: %v", r)
				}
			}()

			Md5Apply(tt.key, tt.model)

			// 获取签名字段
			sigValue := reflect.ValueOf(tt.model).Elem().FieldByName("Signature")
			if tt.expectErr {
				if sigValue.IsValid() {
					t.Error("expected error but got valid signature")
				}
				return
			}

			if !sigValue.IsValid() {
				t.Fatal("signature field not found")
			}

			// 这里需要根据实际情况计算期望值
			// 示例需要替换为真实计算逻辑
			actual := sigValue.String()
			if actual != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, actual)
			}
		})
	}
}

// 模糊测试（属性测试）
func TestSignFuzz(t *testing.T) {
	f := func(model interface{}) bool {
		// 确保模型是结构体指针
		v := reflect.ValueOf(model)
		if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
			return true
		}

		// 检查是否存在access字段
		hasAccess := false
		tags := make(map[string]string)
		for i := 0; i < v.Elem().NumField(); i++ {
			field := v.Elem().Type().Field(i)
			tag := field.Tag.Get("sign")
			tags[field.Name] = tag
			if tag == "access" {
				hasAccess = true
			}
		}

		if !hasAccess {
			return true // 跳过没有access字段的测试用例
		}

		// 生成随机key
		key := "test_key_123"

		// 复制原始模型
		modelCopy := reflect.New(v.Elem().Type()).Interface()

		// 执行签名
		Md5Apply(key, modelCopy)

		// 验证签名字段存在
		sigField := reflect.ValueOf(modelCopy).Elem().FieldByName("Signature")
		if !sigField.IsValid() {
			return false
		}

		// 重新计算预期结果
		params := make([]struct {
			Name  string
			Value string
		}, 0)

		originalModel := v.Elem()
		for i := 0; i < originalModel.NumField(); i++ {
			field := originalModel.Type().Field(i)
			tag := field.Tag.Get("sign")
			if tag == "yes" || tag == "access" {
				val := originalModel.Field(i)
				params = append(params, struct {
					Name  string
					Value string
				}{
					Name:  field.Name,
					Value: fmt.Sprint(val.Interface()),
				})
			}
		}

		// 排序参数
		sort.Slice(params, func(i, j int) bool {
			return params[i].Name < params[j].Name
		})

		// 构建待签字符串
		paramStr := ""
		for _, p := range params {
			if paramStr != "" {
				paramStr += "&"
			}
			paramStr += p.Name + "=" + p.Value
		}

		fullStr := paramStr + "&key=" + key
		expected := fmt.Sprintf("%x", md5.Sum([]byte(fullStr)))

		// 获取实际结果
		actual := sigField.String()

		// 验证长度和格式
		if len(actual) != 32 {
			return false
		}

		// 验证内容
		return actual == expected
	}

	// 运行模糊测试
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}
