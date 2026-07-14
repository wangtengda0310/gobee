package robotext

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-robot/project/xcard/xcard_case"
	"git.devcloud.ztgame.com/v-tangfangda/rain-robot/project/xcard/xcard_case_def"
	"git.devcloud.ztgame.com/v-tangfangda/rain-robot/project/xcard/xcard_client"
	"git.devcloud.ztgame.com/v-tangfangda/rain-robot/project/xcard/xcard_excel/excel"
	"git.devcloud.ztgame.com/v-tangfangda/rain-robot/project/xcard/xcard_excel/excel_manager_impl"
	"git.devcloud.ztgame.com/v-tangfangda/rain-robot/project/xcard/xcard_msg_def"
	protoMsg "git.devcloud.ztgame.com/v-tangfangda/rain-robot/project/xcard/xcard_pb"
	"git.devcloud.ztgame.com/v-tangfangda/rain-robot/vars"

	mcpgo "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ========== 初始化 ==========

// 初始化标志和锁
var (
	initOnce      sync.Once
	isInitialized bool
)

// ensureInitialized 确保 Cases 和 proto 消息已注册（Excel 初始化由外部控制）
func ensureInitialized() {
	initOnce.Do(func() {
		// 禁用 Mark 日志，避免阻塞（Mark 函数会发送到无接收者的通道）
		vars.MarkInLog = false
		// 禁用日志通道，避免阻塞（MCP 调用时没有日志消费者）
		vars.UseSendToWails3QAFuncLogChan = false
		vars.UseSendDebugLogToWails3QAFuncLogChan = false
		// 禁用 OLAP 日志服务，避免 channel 阻塞（SinkService 未启动时发送日志会卡住）
		vars.UseOLAP = false
		// 禁用控制台日志打印，减少输出干扰
		vars.LogNoPrint = true
		// 注册所有 Cases
		xcard_case.RegisterCases(xcard_case_def.CaseDefInstance)
		// 注册所有 proto 消息
		protoMsg.RegisterMsg(xcard_msg_def.MsgDefInstance)
		isInitialized = true
	})
}

// InitExcel 使用指定路径初始化 Excel 配表数据
func InitExcel(path string) error {
	return excel_manager_impl.Init(&path)
}

// ========== 数据结构 ==========

// GuildLeaderInfo 工会会长信息
type GuildLeaderInfo struct {
	GuildID     uint64 `json:"guildId"`     // 工会ID
	GuildName   string `json:"guildName"`   // 工会名称
	LeaderID    uint64 `json:"leaderId"`    // 会长玩家ID
	LeaderName  string `json:"leaderName"`  // 会长角色名
	AccountID   string `json:"accountId"`   // 会长账号，用于登录（FakeSDK模式）
	Country     int32  `json:"country"`     // 势力
	MemberCount int32  `json:"memberCount"` // 成员数量
}

// GuildCreationResult 工会创建结果
type GuildCreationResult struct {
	Success   bool     `json:"success"`             // 是否成功
	GuildID   uint64   `json:"guildId"`             // 工会ID
	GuildName string   `json:"guildName"`           // 工会名称
	AccountID string   `json:"accountId,omitempty"` // 机器人账号
	CityIDs   []uint64 `json:"cityIds,omitempty"`   // 创建的城池ID列表
	Message   string   `json:"message"`             // 结果消息
}

// CityUpgradeResult 城池升级结果
type CityUpgradeResult struct {
	Success     bool   `json:"success"`     // 是否成功
	CityID      uint64 `json:"cityId"`      // 城池ID
	BeforeLevel int32  `json:"beforeLevel"` // 升级前等级
	AfterLevel  int32  `json:"afterLevel"`  // 升级后等级
	Message     string `json:"message"`     // 结果消息
}

// ========== RobotExtService - Wails 服务 ==========

// RobotExtService robot 扩展服务
// 提供工会操作相关的 MCP 工具服务
type RobotExtService struct{}

// NewRobotExtService 创建 robot 扩展服务
func NewRobotExtService() *RobotExtService {
	return &RobotExtService{}
}

// GetGuildLeaders 查询服务器工会列表及会长账号信息
// @mcp
func (s *RobotExtService) GetGuildLeaders(serverIp, serverPort string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	client := NewGuildClient(serverIp, serverPort)
	leaders, err := client.GetGuildLeaders(ctx)
	if err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(leaders, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// CreateGuildWithCity 创建工会并设置野战城池
// @mcp
func (s *RobotExtService) CreateGuildWithCity(prefix string, process, idx int, country int32, cityCount int, serverIp, serverPort string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	client := NewGuildClient(serverIp, serverPort)
	result, err := client.CreateGuildWithCity(ctx, prefix, process, idx, country, cityCount)
	if err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// UpgradeGuildCity 升级工会城池等级
// @mcp
func (s *RobotExtService) UpgradeGuildCity(accountId string, cityId uint64, targetLevel int32, serverIp, serverPort string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	client := NewGuildClient(serverIp, serverPort)
	result, err := client.UpgradeGuildCity(ctx, accountId, cityId, targetLevel)
	if err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ========== GuildClient - 工会操作客户端 ==========

// GuildClient 工会操作客户端
// 封装了机器人登录和工会操作的逻辑
type GuildClient struct {
	serverIp   string
	serverPort string
}

// NewGuildClient 创建工会客户端
func NewGuildClient(serverIp, serverPort string) *GuildClient {
	// 确保消息注册和配表初始化
	ensureInitialized()
	return &GuildClient{
		serverIp:   serverIp,
		serverPort: serverPort,
	}
}

// GetGuildLeaders 获取工会列表及会长账号信息
func (c *GuildClient) GetGuildLeaders(ctx context.Context) ([]GuildLeaderInfo, error) {
	childCtx, childCancel := context.WithCancel(ctx)
	defer childCancel()

	client, cancel, err := c.createAndLoginClient("query", 0, 1, childCtx)
	if err != nil {
		return nil, fmt.Errorf("登录失败: %w", err)
	}
	defer cancel()

	call, _ := client.Call(&protoMsg.GetGuildListReq{Index: 1}, protoMsg.EGameMsgID_GetGuildListReq_id)
	client.Mark("GetGuildListReq sent, waiting for Ack")

	wait, err := client.Wait(protoMsg.EGameMsgID_GetGuildListAck_id, 60*time.Second, childCtx)
	if err != nil {
		return nil, fmt.Errorf("获取工会列表超时: %w", err)
	}

	ack := wait.Msg.(*protoMsg.GetGuildListAck)
	if ack.ErrCode != 0 {
		return nil, fmt.Errorf("获取工会列表失败: errCode=%d", ack.ErrCode)
	}

	client.Record("GetGuildList", call, wait.Time)

	var leaders []GuildLeaderInfo
	for _, guild := range ack.GuildList {
		if guild == nil || guild.Leader == nil || guild.Leader.SimpleInfo == nil {
			continue
		}

		info := GuildLeaderInfo{
			GuildID:     guild.SimpleInfo.Id,
			GuildName:   guild.SimpleInfo.Name,
			LeaderID:    guild.Leader.SimpleInfo.PlayerID,
			LeaderName:  guild.Leader.SimpleInfo.RoleName,
			AccountID:   guild.Leader.SimpleInfo.AccountID,
			Country:     int32(guild.Leader.SimpleInfo.Country),
			MemberCount: guild.CurMemNum,
		}
		leaders = append(leaders, info)
	}

	return leaders, nil
}

// CreateGuildWithCity 创建工会并设置野战城池
func (c *GuildClient) CreateGuildWithCity(ctx context.Context, prefix string, process, idx int, country int32, cityCount int) (*GuildCreationResult, error) {
	childCtx, childCancel := context.WithCancel(ctx)
	defer childCancel()

	client, clientCancel, err := c.createAndLoginClient(prefix, process, idx, childCtx)
	if err != nil {
		childCancel()
		return nil, fmt.Errorf("登录失败: %w", err)
	}
	defer clientCancel()

	result := &GuildCreationResult{
		Success: false,
		CityIDs: make([]uint64, 0),
	}

	// GM 添加所有英雄和等级
	gm := &xcard_case.GMModule{}
	gm.GMAddAllHeroModule(client, childCtx, childCancel)
	gm.GMAddLevelModule(client, 20, childCtx, childCancel)

	// 修改势力
	uim := xcard_case.UserInfoModule{}
	countryEnum := excel.ECountry(country)
	if countryEnum < excel.ECountry_Qin || countryEnum > excel.ECountry_Zhao {
		countryEnum = excel.ECountry_Qin
	}
	uim.ModifyCountry(client, countryEnum, ctx)

	// 创建工会
	gum := xcard_case.GuildModule{}
	guildName := fmt.Sprintf("测%s", client.GetFullName())
	guildId := gum.TryGuildCreate(client, guildName, "自动化测试工会", 1, 20, countryEnum, true, ctx)
	if guildId == 0 {
		result.Message = "创建工会失败"
		return result, nil
	}

	result.GuildID = guildId
	result.GuildName = guildName
	result.AccountID = client.GetFullName()

	// 添加工会资源（赤玉）用于创建城池
	costItemId := int32(400000)
	xcard_case.CheckItemAndRestockModule(client, ctx, "赤玉", 10001, costItemId)

	// 创建城池
	gwm := xcard_case.GuildWarModule{}
	for i := 0; i < cityCount; i++ {
		cityList := gwm.GuildCityList(client, 1, ctx)
		if len(cityList) >= 8 {
			break
		}

		gwm.CreateGuildCity(client, 1, ctx)

		newCityList := gwm.GuildCityList(client, 1, ctx)
		if len(newCityList) > len(cityList) {
			newCity := newCityList[len(newCityList)-1]
			result.CityIDs = append(result.CityIDs, newCity.SimpleInfo.Id)
		}

		time.Sleep(500 * time.Millisecond)
	}

	result.Success = true
	result.Message = fmt.Sprintf("创建工会成功，共创建 %d 个城池", len(result.CityIDs))
	return result, nil
}

// UpgradeGuildCity 升级工会城池
func (c *GuildClient) UpgradeGuildCity(ctx context.Context, accountId string, cityId uint64, targetLevel int32) (*CityUpgradeResult, error) {
	prefix, process, idx := parseAccountId(accountId)

	client, cancel, err := c.createAndLoginClient(prefix, process, idx, ctx)
	if err != nil {
		return nil, fmt.Errorf("登录失败: %w", err)
	}
	defer cancel()

	result := &CityUpgradeResult{
		Success: false,
		CityID:  cityId,
	}

	gwm := xcard_case.GuildWarModule{}
	cityList := gwm.GuildCityList(client, 1, ctx)
	var currentLevel int32
	for _, city := range cityList {
		if city.SimpleInfo.Id == cityId {
			currentLevel = int32(city.SimpleInfo.Level)
			break
		}
	}

	if currentLevel == 0 {
		result.Message = "未找到指定城池"
		return result, nil
	}

	result.BeforeLevel = currentLevel

	// 循环升级直到目标等级
	for currentLevel < targetLevel {
		call, _ := client.Call(&protoMsg.GuildCityUpgradeReq{
			GuildCityId: cityId,
		}, protoMsg.EGameMsgID_GuildCityUpgradeReq_id)

		wait, err := client.Wait(protoMsg.EGameMsgID_GuildCityUpgradeAck_id, 30*time.Second, ctx)
		if err != nil {
			result.Message = fmt.Sprintf("升级请求超时，当前等级: %d", currentLevel)
			result.AfterLevel = currentLevel
			return result, nil
		}

		ack := wait.Msg.(*protoMsg.GuildCityUpgradeAck)
		if ack.ErrCode != 0 {
			result.Message = fmt.Sprintf("升级失败(errCode=%d)，当前等级: %d", ack.ErrCode, currentLevel)
			result.AfterLevel = currentLevel
			return result, nil
		}

		client.Record("GuildCityUpgrade", call, wait.Time)
		currentLevel = int32(ack.NewLevel)

		time.Sleep(300 * time.Millisecond)
	}

	result.Success = true
	result.AfterLevel = currentLevel
	result.Message = fmt.Sprintf("升级成功: %d -> %d", result.BeforeLevel, result.AfterLevel)
	return result, nil
}

// createAndLoginClient 创建客户端并登录
func (c *GuildClient) createAndLoginClient(prefix string, process, idx int, ctx context.Context) (*xcard_client.ClientXCard, context.CancelFunc, error) {
	vars.ServerIp = c.serverIp
	vars.ServerPort = c.serverPort
	vars.TLSEnabled = false

	childCtx, cancel := context.WithCancel(ctx)

	client := xcard_client.NewClientXCard(prefix, process, idx)

	err := xcard_case.LoginModule(client, false, 60*time.Second, "0.8.3001.1.1", false, childCtx, cancel)
	if err != nil {
		cancel()
		return nil, cancel, fmt.Errorf("登录失败: %w", err)
	}

	go func() {
		<-childCtx.Done()
		client.Logout()
	}()

	originalCancel := cancel
	cancel = func() {
		originalCancel()
		time.Sleep(100 * time.Millisecond)
	}

	return client, cancel, nil
}

// ========== 辅助函数 ==========

// parseAccountId 解析账号ID
// 账号格式: {prefix}x{process}x{idx}
func parseAccountId(accountId string) (prefix string, process, idx int) {
	prefix = accountId
	process = 1
	idx = 1

	for i := len(accountId) - 1; i >= 0; i-- {
		if accountId[i] == 'x' {
			for j := i - 1; j >= 0; j-- {
				if accountId[j] == 'x' {
					prefix = accountId[:j]
					process = parseInt(accountId[j+1 : i])
					idx = parseInt(accountId[i+1:])
					return
				}
			}
		}
	}
	return
}

// parseInt 解析整数
func parseInt(s string) int {
	var result int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + int(c-'0')
		} else {
			break
		}
	}
	return result
}

// ========== MCP 工具注册 ==========

// RegisterMCPTools 注册 robot 工会操作相关的 MCP Tools
func RegisterMCPTools(s *mcpgo.Server, svc *RobotExtService) {
	// get_guild_leaders - 查询服务器工会列表及会长账号信息
	s.AddTool(&mcpgo.Tool{
		Name:        "get_guild_leaders",
		Description: "查询服务器工会列表及会长账号信息，返回可用于登录的会长账号列表",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"serverIp": map[string]any{
					"type":        "string",
					"description": "服务器 IP 地址",
				},
				"serverPort": map[string]any{
					"type":        "string",
					"description": "服务器端口",
				},
			},
			"required": []string{"serverIp", "serverPort"},
		},
	}, func(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		var args struct {
			ServerIp   string `json:"serverIp"`
			ServerPort string `json:"serverPort"`
		}
		if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
			return ErrorResult("解析参数失败: " + err.Error()), nil
		}

		result, err := svc.GetGuildLeaders(args.ServerIp, args.ServerPort)
		if err != nil {
			return ErrorResultFromError(err), nil
		}

		return TextResult(result), nil
	})

	// create_guild_with_city - 创建工会并设置野战城池
	s.AddTool(&mcpgo.Tool{
		Name:        "create_guild_with_city",
		Description: "创建工会并设置野战城池。使用机器人账号创建新工会，并创建指定数量的城池。机器人账号格式为 {prefix}x{process}x{idx}",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"prefix": map[string]any{
					"type":        "string",
					"description": "机器人账号前缀",
				},
				"process": map[string]any{
					"type":        "integer",
					"description": "机器人进程编号（默认为 1）",
				},
				"idx": map[string]any{
					"type":        "integer",
					"description": "机器人序号（默认为 1）",
				},
				"country": map[string]any{
					"type":        "integer",
					"description": "势力ID（1-15）：1=秦 2=西楚 3=西汉 4=东汉 5=黄 6=曹魏 7=蜀 8=孙吴 9=张楚 10=魏 11=楚 12=汉 13=燕 14=齐 15=赵",
				},
				"cityCount": map[string]any{
					"type":        "integer",
					"description": "要创建的城池数量（最多 8 个）",
				},
				"serverIp": map[string]any{
					"type":        "string",
					"description": "服务器 IP 地址",
				},
				"serverPort": map[string]any{
					"type":        "string",
					"description": "服务器端口",
				},
			},
			"required": []string{"prefix", "country", "cityCount", "serverIp", "serverPort"},
		},
	}, func(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		var args struct {
			Prefix     string `json:"prefix"`
			Process    int    `json:"process"`
			Idx        int    `json:"idx"`
			Country    int32  `json:"country"`
			CityCount  int    `json:"cityCount"`
			ServerIp   string `json:"serverIp"`
			ServerPort string `json:"serverPort"`
		}
		if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
			return ErrorResult("解析参数失败: " + err.Error()), nil
		}

		// 设置默认值
		if args.Process == 0 {
			args.Process = 1
		}
		if args.Idx == 0 {
			args.Idx = 1
		}

		result, err := svc.CreateGuildWithCity(args.Prefix, args.Process, args.Idx, args.Country, args.CityCount, args.ServerIp, args.ServerPort)
		if err != nil {
			return ErrorResultFromError(err), nil
		}

		return TextResult(result), nil
	})

	// upgrade_guild_city - 升级工会城池等级
	s.AddTool(&mcpgo.Tool{
		Name:        "upgrade_guild_city",
		Description: "升级工会城池等级。使用会长账号登录，循环升级城池直到目标等级。",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"accountId": map[string]any{
					"type":        "string",
					"description": "会长账号（从 get_guild_leaders 获取）",
				},
				"cityId": map[string]any{
					"type":        "integer",
					"description": "城池ID",
				},
				"targetLevel": map[string]any{
					"type":        "integer",
					"description": "目标等级",
				},
				"serverIp": map[string]any{
					"type":        "string",
					"description": "服务器 IP 地址",
				},
				"serverPort": map[string]any{
					"type":        "string",
					"description": "服务器端口",
				},
			},
			"required": []string{"accountId", "cityId", "targetLevel", "serverIp", "serverPort"},
		},
	}, func(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		var args struct {
			AccountId   string `json:"accountId"`
			CityId      uint64 `json:"cityId"`
			TargetLevel int32  `json:"targetLevel"`
			ServerIp    string `json:"serverIp"`
			ServerPort  string `json:"serverPort"`
		}
		if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
			return ErrorResult("解析参数失败: " + err.Error()), nil
		}

		result, err := svc.UpgradeGuildCity(args.AccountId, args.CityId, args.TargetLevel, args.ServerIp, args.ServerPort)
		if err != nil {
			return ErrorResultFromError(err), nil
		}

		return TextResult(result), nil
	})
}

// ========== MCP 工具函数 ==========

// TextResult 创建文本结果
func TextResult(text string) *mcpgo.CallToolResult {
	return &mcpgo.CallToolResult{
		Content: []mcpgo.Content{
			&mcpgo.TextContent{Text: text},
		},
	}
}

// ErrorResult 创建错误结果（从字符串）
func ErrorResult(errMsg string) *mcpgo.CallToolResult {
	return &mcpgo.CallToolResult{
		IsError: true,
		Content: []mcpgo.Content{
			&mcpgo.TextContent{Text: "错误: " + errMsg},
		},
	}
}

// ErrorResultFromError 创建错误结果（从 error 类型）
func ErrorResultFromError(err error) *mcpgo.CallToolResult {
	return &mcpgo.CallToolResult{
		IsError: true,
		Content: []mcpgo.Content{
			&mcpgo.TextContent{Text: fmt.Sprintf("错误: %v", err)},
		},
	}
}
