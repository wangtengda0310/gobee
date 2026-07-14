package mcp

import (
	"context"
	"encoding/json"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/game"
	mcpgo "github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterGameExcelTools 注册游戏数据相关的 MCP Tools
func RegisterGameExcelTools(s *mcpgo.Server, svc *game.GameExcelService) {
	// get_all_hero_cfg - 获取所有英雄配置
	s.AddTool(&mcpgo.Tool{
		Name:        "get_all_hero_cfg",
		Description: "获取所有英雄配置",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}, func(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		heroCfg := svc.GetAllHeroCfg()
		data, err := json.Marshal(heroCfg)
		if err != nil {
			return ErrorResult("序列化英雄配置失败: " + err.Error()), nil
		}
		return TextResult(string(data)), nil
	})

	// get_all_card_cfg - 获取所有卡牌配置
	s.AddTool(&mcpgo.Tool{
		Name:        "get_all_card_cfg",
		Description: "获取所有卡牌配置",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}, func(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		cardCfg := svc.GetAllCardCfg()
		data, err := json.Marshal(cardCfg)
		if err != nil {
			return ErrorResult("序列化卡牌配置失败: " + err.Error()), nil
		}
		return TextResult(string(data)), nil
	})

	// get_all_skill_cfg - 获取所有技能配置
	s.AddTool(&mcpgo.Tool{
		Name:        "get_all_skill_cfg",
		Description: "获取所有技能配置",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}, func(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		skillCfg := svc.GetAllSkillCfg()
		data, err := json.Marshal(skillCfg)
		if err != nil {
			return ErrorResult("序列化技能配置失败: " + err.Error()), nil
		}
		return TextResult(string(data)), nil
	})

	// get_msg_id_map - 获取消息 ID 映射
	s.AddTool(&mcpgo.Tool{
		Name:        "get_msg_id_map",
		Description: "获取消息 ID 映射 (EGameMsgId 名称映射)",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}, func(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		msgIdMap := svc.GetEGameMsgIdNameMap()
		data, err := json.Marshal(msgIdMap)
		if err != nil {
			return ErrorResult("序列化消息ID映射失败: " + err.Error()), nil
		}
		return TextResult(string(data)), nil
	})

	// get_error_code_map - 获取错误码映射
	s.AddTool(&mcpgo.Tool{
		Name:        "get_error_code_map",
		Description: "获取错误码映射 (ErrorCode 名称映射)",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}, func(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		errorCodeMap := svc.GetErrorCodeMap()
		data, err := json.Marshal(errorCodeMap)
		if err != nil {
			return ErrorResult("序列化错误码映射失败: " + err.Error()), nil
		}
		return TextResult(string(data)), nil
	})

	// get_property_type_map - 获取属性类型映射
	s.AddTool(&mcpgo.Tool{
		Name:        "get_property_type_map",
		Description: "获取属性类型映射 (PropertyType 名称映射，已去除 Pro_ 前缀)",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}, func(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		propertyTypeMap := svc.GetPropertyTypeMap()
		data, err := json.Marshal(propertyTypeMap)
		if err != nil {
			return ErrorResult("序列化属性类型映射失败: " + err.Error()), nil
		}
		return TextResult(string(data)), nil
	})
}
