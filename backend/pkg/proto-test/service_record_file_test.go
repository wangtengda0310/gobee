package prototest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUpdateMessageFieldValues 测试字段4态值更新功能
func TestUpdateMessageFieldValues(t *testing.T) {
	// 1. 创建临时测试文件
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_recording.json")

	// 2. 准备测试数据
	testData := &RecordFileData{
		Version:    1,
		RecordedAt: "2026-06-04T12:20:00Z",
		ServerAddr: "127.0.0.1:18000",
		Messages: []RecordEntryView{
			{
				Index:     0,
				OffsetMs:  0,
				MsgID:     1001,
				MsgName:   "LoginReq",
				SeqID:     1,
				Payload:   map[string]any{"open_id": "test_user_001", "device_id": "device_123"},
				Direction: "→",
			},
		},
	}

	// 3. 先保存初始数据
	service := NewRecordFileService()
	err := service.SaveRecordFile(testFile, testData)
	require.NoError(t, err, "保存初始数据失败")

	// 4.准备4态值数据
	fieldValues := map[string]FieldValues{
		"user_id": {
			RangeValue: RangeValue{
				Start: 10000,
				Step:  100,
				End:   20000,
			},
			EnumValue:  []any{"12345", "67890", "11111"},
			ComboValue: []any{"vip", "premium"},
			InputType:  "range",
		},
		"level": {
			RangeValue: RangeValue{
				Start: 1,
				Step:  1,
				End:   100,
			},
			EnumValue:  []any{1, 10, 50, 100},
			ComboValue: nil,
			InputType:  "enum",
		},
	}

	// 5. 调用 UpdateMessageFieldValues
	updatedData, err := service.UpdateMessageFieldValues(testFile, 0, fieldValues)
	require.NoError(t, err, "更新字段4态值失败")

	// 6. 验证返回数据（不比较完整结构，只验证关键字段）
	assert.NotNil(t, updatedData)
	assert.Equal(t, 1, updatedData.MessageCount)
	assert.Equal(t, 0, updatedData.Messages[0].Index)
	assert.Equal(t, 0, updatedData.Messages[0].OffsetMs)
	assert.Equal(t, uint16(1001), updatedData.Messages[0].MsgID)
	assert.Equal(t, "LoginReq", updatedData.Messages[0].MsgName)
	assert.Equal(t, uint32(1), updatedData.Messages[0].SeqID)
	assert.Equal(t, "→", updatedData.Messages[0].Direction)
	assert.Contains(t, updatedData.Messages[0].Payload["open_id"], "test_user_001")
	assert.Contains(t, updatedData.Messages[0].Payload["device_id"], "device_123")

	// 7. 重新加载文件验证持久化
	reloadedData, err := service.LoadRecordFile(testFile)
	require.NoError(t, err, "重新加载文件失败")

	// 8. 验证 field_values 正确保存
	assert.Equal(t, 10000, reloadedData.Messages[0].FieldValues["user_id"].RangeValue.Start)
	assert.Equal(t, 100, reloadedData.Messages[0].FieldValues["user_id"].RangeValue.Step)
	assert.Equal(t, 20000, reloadedData.Messages[0].FieldValues["user_id"].RangeValue.End)
	assert.Equal(t, []any{"12345", "67890", "11111"}, reloadedData.Messages[0].FieldValues["user_id"].EnumValue)
	assert.Equal(t, []any{"vip", "premium"}, reloadedData.Messages[0].FieldValues["user_id"].ComboValue)
	assert.Equal(t, "range", reloadedData.Messages[0].FieldValues["user_id"].InputType)

	// 9. 验证 level 字段
	assert.Equal(t, []any{float64(1), float64(10), float64(50), float64(100)}, reloadedData.Messages[0].FieldValues["level"].EnumValue)
	assert.Equal(t, "enum", reloadedData.Messages[0].FieldValues["level"].InputType)
}

// TestFieldValuesJSONSerialization 测试 FieldValues JSON 序列化
func TestFieldValuesJSONSerialization(t *testing.T) {
	// 准备测试数据
	fieldValues := map[string]FieldValues{
		"test_field": {
			RangeValue: RangeValue{
				Start: 0,
				Step:  1,
				End:   10,
			},
			EnumValue:  []any{"a", "b", "c"},
			ComboValue: []any{"x", "y", "z"},
			InputType:  "original",
		},
	}

	// 创建包含 field_values 的 RecordEntryView
	entry := RecordEntryView{
		Index:       0,
		OffsetMs:    100,
		MsgID:       3001,
		MsgName:     "TestMsg",
		SeqID:       1,
		Payload:     map[string]any{"test_field": "test_value"},
		Direction:   "→",
		FieldValues: fieldValues,
	}

	// 创建测试录制文件
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_serialize.json")

	service := NewRecordFileService()
	testData := &RecordFileData{
		Version:    1,
		RecordedAt: "2026-06-04T12:20:00Z",
		ServerAddr: "127.0.0.1:18000",
		Messages:   []RecordEntryView{entry},
	}

	// 保存文件
	err := service.SaveRecordFile(testFile, testData)
	require.NoError(t, err, "保存文件失败")

	// 读取文件内容验证 JSON 格式
	content, err := os.ReadFile(testFile)
	require.NoError(t, err, "读取文件失败")

	// 验证 JSON 包含 field_values 字段
	jsonContent := string(content)
	assert.Contains(t, jsonContent, "field_values")
	assert.Contains(t, jsonContent, "range_value")
	assert.Contains(t, jsonContent, "enum_value")
	assert.Contains(t, jsonContent, "combo_value")
	assert.Contains(t, jsonContent, "input_type")
}

// TestLoadRecordFileWithFieldValues 测试加载包含 field_values 的录制文件
func TestLoadRecordFileWithFieldValues(t *testing.T) {
	// 创建包含 field_values 的测试文件
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_load.json")

	service := NewRecordFileService()
	testData := &RecordFileData{
		Version:    1,
		RecordedAt: "2026-06-04T12:20:00Z",
		ServerAddr: "127.0.0.1:18000",
		Messages: []RecordEntryView{
			{
				Index:     0,
				OffsetMs:  0,
				MsgID:     1001,
				MsgName:   "LoginReq",
				SeqID:     1,
				Payload:   map[string]any{"open_id": "test"},
				Direction: "→",
				FieldValues: map[string]FieldValues{
					"open_id": {
						RangeValue: RangeValue{
							Start: 1,
							Step:  1,
							End:   100,
						},
						EnumValue:  nil,
						ComboValue: nil,
						InputType:  "original",
					},
				},
			},
		},
	}

	// 保存文件
	err := service.SaveRecordFile(testFile, testData)
	require.NoError(t, err, "保存文件失败")

	// 重新加载
	loadedData, err := service.LoadRecordFile(testFile)
	require.NoError(t, err, "加载文件失败")

	// 验证 field_values 正确恢复
	assert.NotNil(t, loadedData.Messages[0].FieldValues)
	assert.Equal(t, "original", loadedData.Messages[0].FieldValues["open_id"].InputType)
}
