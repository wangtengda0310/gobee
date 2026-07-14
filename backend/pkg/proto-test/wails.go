package prototest

import (
	"encoding/json"
	"strings"

	protocol "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/msg"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/params"
	_ "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/variables"
)

// RangeValue 范围输入值（对应前端 range-input 组件）
// 类型别名，定义在 params 包中
type RangeValue = params.RangeValue

// FieldValues 字段4态值（对应前端 FieldFourState 接口）
// 类型别名，定义在 params 包中
type FieldValues = params.FieldValues

// RecordFileData 录制文件数据（传给前端） 修改结构时需要更新version
type RecordFileData struct {
	Version      int               `json:"version"`
	RecordedAt   string            `json:"recorded_at"`
	ServerAddr   string            `json:"server_addr"`
	MessageCount int               `json:"message_count"`
	Messages     []RecordEntryView `json:"messages"`
}

// RecordEntryView 录制条目视图（Payload 直接为 map[string]any，前端无需 JSON.parse） 修改时需要同步proto_cases/目录下记载的schema
type RecordEntryView struct {
	Index       int                    `json:"index"`
	OffsetMs    int                    `json:"offset_ms"`
	MsgID       uint16                 `json:"msg_id"`
	MsgName     string                 `json:"msg_name"`
	SeqID       uint32                 `json:"seq_id"`
	Payload     map[string]any         `json:"payload"`
	Direction   string                 `json:"direction"`
	AccountID   string                 `json:"account_id"`             // 重放时的账号标识（如 test1、test2），录制时为空
	Descript    string                 `json:"descript"`               // 消息描述信息（用户可编辑）
	FieldValues map[string]FieldValues `json:"field_values,omitempty"` // 字段路径→4态值映射
}

// singleEntryToView 将单条 RecordEntry 转换为 RecordEntryView（唯一转换点）
// 所有 RecordEntry → RecordEntryView 的转换都必须调用此函数，禁止重复实现
func singleEntryToView(e *protocol.RecordEntry, index int) RecordEntryView {
	dir := e.Direction
	if dir == "" {
		dir = "→"
	}
	// 将 json.RawMessage 解析为 map[string]any
	var payload map[string]any
	if len(e.PayloadJSON) > 0 {
		_ = json.Unmarshal(e.PayloadJSON, &payload)
	}
	// 确保 payload 不为 nil，避免 Wails $Create.Map 将 null 反序列化为 undefined
	if payload == nil {
		payload = map[string]any{}
	}
	return RecordEntryView{
		Index:       index,
		OffsetMs:    e.OffsetMs,
		MsgID:       e.MsgID,
		MsgName:     e.MsgName,
		SeqID:       e.SeqID,
		Payload:     payload,
		Direction:   dir,
		AccountID:   "",         // 录制时不携带账号信息，重放时通过 entryViewFromJSON 设置
		Descript:    e.Descript, // 消息描述
		FieldValues: convertFieldValues(e.FieldValues),
	}
}

// entryViewFromJSON 从散列参数构建 RecordEntryView（供 EmitReplayMessage 使用）
// payloadJSON 为 JSON 字符串，在此统一解析为 map[string]any
func entryViewFromJSON(msgName string, msgID uint16, seqID uint32, payloadJSON string, offsetMs int, direction string, accountID string) RecordEntryView {
	var payload map[string]any
	if payloadJSON != "" {
		_ = json.Unmarshal([]byte(payloadJSON), &payload)
	}
	if payload == nil {
		payload = map[string]any{}
	}
	dir := direction
	if dir == "" {
		dir = "→"
	}
	return RecordEntryView{
		Index:     0, // 由前端追加时设置
		OffsetMs:  offsetMs,
		MsgID:     msgID,
		MsgName:   msgName,
		SeqID:     seqID,
		Payload:   payload,
		Direction: dir,
		AccountID: accountID, // 重放时的账号标识
		Descript:  "",        // 重放时不设置描述
	}
}

// recordingToViews 将 Recording 的消息转换为 RecordEntryView 切片
func recordingToViews(rec *protocol.Recording) []RecordEntryView {
	views := make([]RecordEntryView, len(rec.Messages))
	for i, e := range rec.Messages {
		views[i] = singleEntryToView(e, i)
	}
	return views
}

// convertFieldValues 转换 FieldValues: map[string]any → map[string]FieldValues
func convertFieldValues(fvMap map[string]any) map[string]FieldValues {
	if fvMap == nil {
		return nil
	}

	result := make(map[string]FieldValues)
	for k, v := range fvMap {
		// 由于 JSON 序列化/反序列化的特性，需要处理类型转换
		// 这里简化处理：假设 v 就是 FieldValues 类型
		if fv, ok := v.(map[string]any); ok {
			// 从 map[string]any 重建 FieldValues
			result[k] = FieldValues{
				RangeValue:   convertRangeValue(fv["range_value"]),
				EnumValue:    convertAnySlice(fv["enum_value"]),
				ComboValue:   convertAnySlice(fv["combo_value"]),
				InputType:    strValue(fv["input_type"]),
				VariableName: strValue(fv["variable_name"]),
			}
		}
	}
	return result
}

// convertRangeValue 转换 RangeValue: map[string]any → RangeValue
func convertRangeValue(rv any) RangeValue {
	if rv == nil {
		return RangeValue{}
	}
	if rvMap, ok := rv.(map[string]any); ok {
		return RangeValue{
			Start: intVal(rvMap["start"]),
			Step:  intVal(rvMap["step"]),
			End:   intVal(rvMap["end"]),
		}
	}
	return RangeValue{}
}

// intValue 安全获取整数值
func intVal(v any) int {
	switch val := v.(type) {
	case int:
		return val
	case float64:
		return int(val)
	case float32:
		return int(val)
	default:
		return 0
	}
}

// strValue 安全获取字符串值
func strValue(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// convertAnySlice 将 interface 转为 []any，支持 nil、[]any、旧版逗号分隔字符串
func convertAnySlice(v any) []any {
	if v == nil {
		return nil
	}
	// 已经是数组
	if arr, ok := v.([]any); ok {
		return arr
	}
	// 兼容旧格式：逗号分隔字符串
	if s, ok := v.(string); ok && s != "" {
		parts := strings.Split(s, ",")
		result := make([]any, len(parts))
		for i, p := range parts {
			result[i] = params.ParseValueString(p)
		}
		return result
	}
	return nil
}
