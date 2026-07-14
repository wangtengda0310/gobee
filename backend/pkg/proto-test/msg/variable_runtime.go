package protocol

import (
	"fmt"
	"log"
	"strings"
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/params"
)

// ScanFieldValuesForVariables 根据消息的 FieldValues 元数据扫描变量依赖。
//
// 返回值语义（F1 修复 2026-06-15）:
//   - hasVariable: 只要存在任何 InputType=="variable" 且 VariableName 非空的字段就为 true，
//     不论该变量是否已注册。这样未注册变量（拼写错误）也会进入惰性提取路径，
//     由 ExtractVariablesForMessage 返回 error 并经 onMessage 上报跳过，
//     而非静默用写死值发送。
//   - watchedIDs: 仅收集已注册变量的 WatchMsgIDs（未注册变量无 WatchMsgIDs）。
//     即使为空（全部变量未注册），FrameMux 仍可正常创建——只是不缓存任何帧。
func ScanFieldValuesForVariables(messages []RecordMessage) (hasVariable bool, watchedIDs []uint16) {
	idSet := make(map[uint16]bool)
	for _, msg := range messages {
		for _, fv := range msg.FieldValues {
			if fv.InputType != "variable" || fv.VariableName == "" {
				continue
			}
			// 有变量字段即置 hasVariable，无论是否注册
			hasVariable = true
			def := params.FindVariableByShortName(fv.VariableName)
			if def == nil {
				// 未注册变量：不贡献 watchedIDs，但仍计入 hasVariable，
				// 以便后续 ExtractVariablesForMessage 能报错上报（而非静默兜底）
				continue
			}
			for _, id := range def.WatchMsgIDs {
				idSet[id] = true
			}
		}
	}
	if hasVariable {
		watchedIDs = make([]uint16, 0, len(idSet))
		for id := range idSet {
			watchedIDs = append(watchedIDs, id)
		}
	}
	return
}

// msgNeedsVariable 判断单条消息是否依赖变量(有 InputType=="variable" 且 VariableName 非空的字段)。
// 用于发送循环中决定是否对该消息触发惰性变量提取,替代全局 hasVariable 开关。
func msgNeedsVariable(msg RecordMessage) bool {
	for _, fv := range msg.FieldValues {
		if fv.InputType == "variable" && fv.VariableName != "" {
			return true
		}
	}
	return false
}

// ExtractVariablesForMessage 按需提取单条消息所需的变量。
//
// 与 ExtractVariableValues 的区别(惰性提取的核心):
//   - 只处理该消息 FieldValues 中声明的变量,而非全局 watchedIDs
//   - 先检查 variableStore,跳过已有值的变量(避免重复 WaitMsg 5s 阻塞)
//   - 仅对 variableStore 中缺失的变量调用 WaitMsg
//
// 返回提取过程中发生的错误(如某个变量 WaitMsg 超时)。即使返回错误,
// variableStore 中可能已填充了部分成功提取的变量。调用方据错误决定是否跳过该消息。
func ExtractVariablesForMessage(msg RecordMessage, mux *FrameMux, variableStore map[string]any) error {
	if mux == nil {
		return nil
	}
	if !msgNeedsVariable(msg) {
		return nil
	}

	// 收集该消息需要的、且尚未提取的变量名 → VariableDef
	missingDefs := make([]*params.VariableDef, 0)
	seen := make(map[string]bool)
	var unregisteredVars []string
	for _, fv := range msg.FieldValues {
		if fv.InputType != "variable" || fv.VariableName == "" {
			continue
		}
		if seen[fv.VariableName] {
			continue
		}
		seen[fv.VariableName] = true
		// 已有值:跳过(复用之前消息提取的结果)
		if _, exists := variableStore[fv.VariableName]; exists {
			continue
		}
		def := params.FindVariableByShortName(fv.VariableName)
		if def == nil {
			// 未注册变量是配置错误,必须报错而非静默用写死值兜底
			// (否则 QA 工具会"测试了错误数据却以为成功")
			unregisteredVars = append(unregisteredVars, fv.VariableName)
			continue
		}
		missingDefs = append(missingDefs, def)
	}

	// 未注册变量直接返回错误,阻止消息用写死值发送
	if len(unregisteredVars) > 0 {
		return fmt.Errorf("变量未注册(检查 field_values.variable_name 拼写): %s", strings.Join(unregisteredVars, ", "))
	}

	if len(missingDefs) == 0 {
		return nil
	}

	// 对每个缺失变量,逐个 WaitMsg 其关注的 MsgID
	var errs []string
	for _, def := range missingDefs {
		extracted := false
		for _, watchID := range def.WatchMsgIDs {
			frame, err := mux.WaitMsg(watchID, 5*time.Second)
			if err != nil {
				continue // 此 MsgID 超时,尝试下一个 watchID
			}
			val, extractErr := def.ExtractFunc(frame)
			if extractErr != nil {
				log.Printf("[变量提取] %s 提取失败(MsgID=%d): %v", def.ShortName, watchID, extractErr)
				continue
			}
			if val == nil {
				continue
			}
			variableStore[def.ShortName] = val
			log.Printf("[变量提取] %s = %v (从 MsgID=%d, 惰性按需)", def.ShortName, val, watchID)
			extracted = true
			break
		}
		if !extracted {
			errs = append(errs, fmt.Sprintf("%s(关注 %v)", def.ShortName, def.WatchMsgIDs))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("变量提取失败: %s", strings.Join(errs, ", "))
	}
	return nil
}

// ResolveMessageVariables 根据 FieldValues 元数据将变量值注入到 payload 中
func ResolveMessageVariables(payload map[string]any, fieldValues map[string]FieldMetaValue, variableStore map[string]any) (map[string]any, error) {
	if len(fieldValues) == 0 || len(variableStore) == 0 {
		return payload, nil
	}
	var missingVars []string
	for field, fv := range fieldValues {
		if fv.InputType != "variable" || fv.VariableName == "" {
			continue
		}
		val, ok := variableStore[fv.VariableName]
		if !ok {
			missingVars = append(missingVars, fv.VariableName)
			continue
		}
		payload[field] = val
	}
	if len(missingVars) > 0 {
		return payload, fmt.Errorf("变量未解析: %v", missingVars)
	}
	return payload, nil
}
