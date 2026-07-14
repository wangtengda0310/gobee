package mcp

import (
	"context"
	"encoding/json"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/game"
	excel "git.devcloud.ztgame.com/v-tangfangda/rain-robot/project/xcard/xcard_excel/excel"
	excel_config "git.devcloud.ztgame.com/v-tangfangda/rain-robot/project/xcard/xcard_excel/excel_config"
	mcpgo "github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterGameExcelTools 注册游戏数据相关的 MCP Tools
func RegisterGameExcelTools(s *mcpgo.Server, svc *game.GameExcelService) {
	// get_all_hero_cfg - 获取所有英雄配置（全量）
	s.AddTool(&mcpgo.Tool{
		Name:        "get_all_hero_cfg",
		Description: "获取所有英雄配置(全量, 大量消耗token); 按名查询单个/批量英雄请优先用 get_hero_cfg_by_name 省 token",
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

	// get_hero_cfg_by_name - 按英雄名称精准查询（省 token，优先于 get_all_hero_cfg）
	s.AddTool(&mcpgo.Tool{
		Name:        "get_hero_cfg_by_name",
		Description: "按英雄名称查询英雄配置，支持单个或批量；仅返回命中的配置，远比 get_all_hero_cfg 省 token",
		InputSchema: nameArgSchema("hero_name", "英雄名称（单个或批量）"),
	}, func(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		var args struct {
			HeroName any `json:"hero_name"`
		}
		if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
			return ErrorResult("解析参数失败: " + err.Error()), nil
		}
		names := toStringSlice(args.HeroName)
		if len(names) == 0 {
			return ErrorResult("hero_name 不能为空"), nil
		}
		// 命名武将名来自嵌入的 HerosTemplate.Name；nil 配置返回 "" 跳过
		found, notFound := filterByName(svc.GetAllHeroCfg(), names, func(cfg *excel_config.HeroConfig) string {
			if cfg == nil || cfg.HerosTemplate == nil {
				return ""
			}
			return cfg.Name
		})
		return marshalNameQueryResult("heroes", found, notFound)
	})

	// get_all_card_cfg - 获取所有卡牌配置（全量）
	s.AddTool(&mcpgo.Tool{
		Name:        "get_all_card_cfg",
		Description: "获取所有卡牌配置(全量, 大量消耗token); 按名查询单个/批量卡牌请优先用 get_card_cfg_by_name 省 token",
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

	// get_card_cfg_by_name - 按卡牌名称精准查询（省 token，优先于 get_all_card_cfg）
	s.AddTool(&mcpgo.Tool{
		Name:        "get_card_cfg_by_name",
		Description: "按卡牌名称查询卡牌配置，支持单个或批量；仅返回命中的配置，远比 get_all_card_cfg 省 token。注意同名卡可能有多个 id（如普杀/火杀/雷杀）",
		InputSchema: nameArgSchema("card_name", "卡牌名称（单个或批量，如 杀/闪/桃/丈八蛇矛）"),
	}, func(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		var args struct {
			CardName any `json:"card_name"`
		}
		if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
			return ErrorResult("解析参数失败: " + err.Error()), nil
		}
		names := toStringSlice(args.CardName)
		if len(names) == 0 {
			return ErrorResult("card_name 不能为空"), nil
		}
		found, notFound := filterByName(svc.GetAllCardCfg(), names, func(cfg *excel.CardsTemplate) string {
			if cfg == nil {
				return ""
			}
			return cfg.Name
		})
		return marshalNameQueryResult("cards", found, notFound)
	})

	// get_all_skill_cfg - 获取所有技能配置（全量）
	s.AddTool(&mcpgo.Tool{
		Name:        "get_all_skill_cfg",
		Description: "获取所有技能配置(全量, 大量消耗token); 按名查询单个/批量技能请优先用 get_skill_cfg_by_name 省 token",
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

	// get_skill_cfg_by_name - 按技能名称精准查询（省 token，优先于 get_all_skill_cfg）
	s.AddTool(&mcpgo.Tool{
		Name:        "get_skill_cfg_by_name",
		Description: "按技能名称查询技能配置，支持单个或批量；仅返回命中的配置，远比 get_all_skill_cfg 省 token",
		InputSchema: nameArgSchema("skill_name", "技能名称（单个或批量，如 龙胆/七进七出）"),
	}, func(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		var args struct {
			SkillName any `json:"skill_name"`
		}
		if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
			return ErrorResult("解析参数失败: " + err.Error()), nil
		}
		names := toStringSlice(args.SkillName)
		if len(names) == 0 {
			return ErrorResult("skill_name 不能为空"), nil
		}
		// 技能名是 SkillName 字段（注意不是 Name）
		found, notFound := filterByName(svc.GetAllSkillCfg(), names, func(cfg *excel.SkillsTemplate) string {
			if cfg == nil {
				return ""
			}
			return cfg.SkillName
		})
		return marshalNameQueryResult("skills", found, notFound)
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

// ========== by_name 系列工具的公共辅助 ==========

// nameArgSchema 生成「单个 string 或 string 数组」的 JSON Schema（by_name 系列工具复用）
// argName 为参数名（如 hero_name），desc 为参数说明
func nameArgSchema(argName, desc string) map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			argName: map[string]any{
				"oneOf": []any{
					map[string]any{"type": "string"},
					map[string]any{
						"type":  "array",
						"items": map[string]any{"type": "string"},
					},
				},
				"description": desc,
			},
		},
		"required": []string{argName},
	}
}

// toStringSlice 将 any（string 或 []any）统一转为非空 string 列表
// 用于 by_name 系列工具解析「单个或批量名称」参数
func toStringSlice(v any) []string {
	names := make([]string, 0)
	switch t := v.(type) {
	case string:
		if t != "" {
			names = append(names, t)
		}
	case []any:
		for _, item := range t {
			if s, ok := item.(string); ok && s != "" {
				names = append(names, s)
			}
		}
	}
	return names
}

// filterByName 在全量配置 map 中按名字列表精确过滤
// getName 返回配置的匹配名，返回 "" 表示该项应跳过（如 nil 配置）
// 同名多条会全部收集（如"杀"对应普杀/火杀/雷杀多个 id）
// 返回：命中的 name→配置数组 映射（供 JSON 序列化）、未命中名称列表
func filterByName[K comparable, V any](all map[K]V, names []string, getName func(V) string) (map[string][]any, []string) {
	found := make(map[string][]any)
	notFound := make([]string, 0, len(names))
	for _, name := range names {
		matches := make([]any, 0)
		for _, cfg := range all {
			if getName(cfg) == name {
				matches = append(matches, cfg)
			}
		}
		if len(matches) > 0 {
			found[name] = matches
		} else {
			notFound = append(notFound, name)
		}
	}
	return found, notFound
}

// marshalNameQueryResult 将 by_name 查询结果序列化为 {resultKey: {name:[...], ...}, notFound: [...]} 文本
// resultKey 为命中映射的键名（如 heroes/cards/skills），每个 name 对应命中的配置数组（同名多条全保留）
func marshalNameQueryResult(resultKey string, found map[string][]any, notFound []string) (*mcpgo.CallToolResult, error) {
	data, err := json.Marshal(map[string]any{
		resultKey:  found,
		"notFound": notFound,
	})
	if err != nil {
		return ErrorResult("序列化查询结果失败: " + err.Error()), nil
	}
	return TextResult(string(data)), nil
}
