package functiontest

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	internal "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/appconfig"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/feishu"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/game"
	"git.devcloud.ztgame.com/v-tangfangda/rain-robot/log_service"
	"git.devcloud.ztgame.com/v-tangfangda/rain-robot/project/xcard/xcard_client"
	"git.devcloud.ztgame.com/v-tangfangda/rain-robot/vars"
)

// JsonCaseService JSON 用例服务，提供用例管理、执行等功能
// @frontend @mcp
type JsonCaseService struct {
	emitter    RobotLogEmitter
	isRunning  atomic.Bool // 使用原子操作
	cancelFunc context.CancelFunc
}

// NewJsonCaseService 创建 JSON 用例服务实例。
// emitter 用于向 GUI（Wails）前端推送 robotLog 事件；CLI/MCP 场景传 nil 即可。
func NewJsonCaseService(emitter RobotLogEmitter) *JsonCaseService {
	return &JsonCaseService{
		emitter: emitter,
	}
}

// emitRobotLog 向 GUI 推送机器人日志事件。emitter 为 nil（CLI/MCP 场景）时安全跳过。
func (a *JsonCaseService) emitRobotLog(log *log_service.Wails3Log, seq uint64) {
	if a.emitter == nil || log == nil {
		return
	}
	a.emitter.Emit("robotLog", log, seq)
}

// QAFuncCase JSON数据结构体
// ⚠️ 以下字段名与 cases/fight_cases_schema.json 及前端 use-case-data.ts 强耦合，修改需同步更新
// ⚠️ 历史问题参考: commit 2ee0e9c (字段名不一致导致功能失效)
type QAFuncCase struct {
	Case      string    `json:"case"`      // @json-critical 用例唯一标识，前端 label 映射
	Desc      string    `json:"desc"`      // @json-critical 用例描述，前端 caseDesc 映射
	Manager   string    `json:"manager"`   // @json-critical 负责人，前端 caseManager 映射
	InitYanWu InitYanWu `json:"initYanWu"` // @json-critical 初始演武配置
	Steps     []Step    `json:"steps"`     // @json-critical 执行步骤列表，前端 caseSteps 映射
}

type InitYanWu struct {
	PresentId     int32        `json:"presentId"`     // @json-critical 场景ID
	CardPile      int32        `json:"cardPile"`      // @json-critical 牌堆类型
	Cards         []int32      `json:"cards"`         // @json-critical 初始牌堆卡牌ID列表
	OperateTimeMs int64        `json:"operateTimeMs"` // @json-critical 操作超时时间(毫秒)
	CustomHeroes  []CustomHero `json:"customHeroes"`  // @json-critical 自定义英雄列表
}

type CustomHero struct {
	Identity      int32               `json:"identity"`      // @json-critical 身份标识
	Color         int32               `json:"color"`         // @json-critical 阵营颜色（身份阵营颜色编号，取值见前端 Identity.ts excelIdentityColorMap，权威定义为游戏配表 IdentityEncodeRule；⚠️ reverse_translate.go 误用 countryMap 国家表翻译，已知 bug）
	HeroId        uint32              `json:"heroId"`        // @json-critical 英雄配置ID
	InitCards     []uint32            `json:"initCards"`     // @json-critical 初始手牌ID列表
	InitEquips    []uint32            `json:"initEquips"`    // @json-critical 初始装备ID列表
	AugurCards    []uint32            `json:"augurCards"`    // @json-critical 卜卦牌ID列表
	AddSkills     []uint32            `json:"addSkills"`     // @json-critical 添加技能ID列表
	DelSkills     []uint32            `json:"delSkills"`     // @json-critical 移除技能ID列表
	ExEquips      []uint32            `json:"exEquips"`      // @json-critical 扩展装备ID列表
	SkillCardsMap map[uint32][]uint32 `json:"skillCardsMap"` // @json-critical 技能卡牌映射
}

type Step struct {
	ID             int      `json:"id,omitempty"`             // @json-critical 步骤编号
	Desc           string   `json:"desc,omitempty"`           // @json-critical 步骤描述
	RobotIdx       int      `json:"robotIdx"`                 // @json-critical 机器人索引，对应局内seatId
	Action         string   `json:"action"`                   // @json-critical 动作类型，枚举值见 fight_cases_schema.json
	TargetsId      []int    `json:"targetsId,omitempty"`      // @json-critical 目标玩家ID列表
	Cards          []uint32 `json:"cards,omitempty"`          // @json-critical 使用的卡牌ID列表
	CardsConcat    []uint32 `json:"cardsConcat,omitempty"`    // @json-critical 卡牌拼接，用于多张牌的操作
	Confirm        bool     `json:"confirm"`                  // @json-critical 是否确认，用于OptRoomAction
	HeroSkillUUID  uint32   `json:"heroSkillUuid,omitempty"`  // @json-critical 英雄技能UUID
	TransCardSkill uint32   `json:"transCardSkill,omitempty"` // @json-critical 转化技能ID
	Timeout        float64  `json:"timeout,omitempty"`        // @json-critical 超时时间(秒)
	SleepTime      float64  `json:"sleepTime,omitempty"`      // @json-critical 等待时间(秒)
	Assets         []Asset  `json:"assets,omitempty"`         // @json-critical 断言列表
}

type Asset struct {
	MsgName string            `json:"msgName"` // @json-critical 消息名称，如 AttrChange, PlayCardAck, DrawCard 等
	Desc    string            `json:"desc"`    // @json-critical 断言描述
	Attr    map[string]string `json:"attr"`    // @json-critical 属性键值对断言，所有值都是字符串类型
}

// CaseInfo 用例基本信息，用于列表展示
type CaseInfo struct {
	Case       string     `json:"case"`
	Desc       string     `json:"desc"`
	StepCount  int        `json:"stepCount"`
	HeroCount  int        `json:"heroCount"`
	FileName   string     `json:"fileName"`   // 所属的JSON文件名
	Category   string     `json:"category"`   // 分类（目录名）
	FullPath   string     `json:"fullPath"`   // 完整文件路径
	QAFuncCase QAFuncCase `json:"qaFuncCase"` // 直接全部返回吧
}

// CategoryInfo 分类信息
type CategoryInfo struct {
	Name      string     `json:"name"`
	Path      string     `json:"path"`
	CaseCount int        `json:"caseCount"`
	Cases     []CaseInfo `json:"cases"`
}

// GenerateAssetDesc 生成单个 asset 的智能描述（前端"应用智能描述"按钮用）。
//
// 流程：构造 GameExcelService（单例感知，查不到名字返回空），调用 reverse_translate.go
// 中的 generateAiDesc（与原 TS use-asset-ai-desc.ts 行为一致，3662 case 验证通过）。
//
// @frontend
func (a *JsonCaseService) GenerateAssetDesc(asset Asset, step Step, initYanWu InitYanWu) string {
	// GameExcelService 新建实例即可（IsInitialized 单例感知，查不到名字时返回空）
	gameExcel := game.NewGameExcelService()
	return generateAiDesc(asset, step, &initYanWu, newNameResolver(gameExcel))
}

// RunRobotTest 执行机器人测试
// @frontend @mcp
//
// maxTimeoutPerCase 为单批用例的整体执行超时；传 0 表示用默认 10 分钟。
// CLI（cobra.go）把 --timeout 透传至此，避免卡死用例长时间阻塞。
func (a *JsonCaseService) RunRobotTest(
	ip, port, prefix, desc, feishuGuid string,
	loginTime float64,
	filterCasesOption, filterCaseNameOption []string,
	dir *string,
	filesData *map[string][]byte,
	opTimeMs int,
	feishuNtf, debugLevel, debugLog bool,
	concurrency uint,
	maxTimeoutPerCase time.Duration,
) error {
	seq := atomic.Uint64{}

	// 检查是否已经在运行
	if !a.isRunning.CompareAndSwap(false, true) {
		return fmt.Errorf("测试正在运行中，请等待完成")
	}
	defer a.isRunning.Store(false)

	// 创建可取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelFunc = cancel
	defer func() {
		cancel()
		a.cancelFunc = nil
	}()

	fmt.Println("ip:", ip)
	fmt.Println("port:", port)
	fmt.Println("prefix:", prefix)
	fmt.Println("desc:", desc)
	fmt.Println("feishuGuid:", feishuGuid)
	if dir != nil {
		fmt.Println("dir:", *dir)
	}
	fmt.Println("loginTime:", loginTime)
	fmt.Println("filterCasesOption:", filterCasesOption)
	fmt.Println("filterCaseNameOption:", filterCaseNameOption)
	fmt.Println("opTimeMs:", opTimeMs)
	fmt.Println("feishuNtf:", feishuNtf)
	fmt.Println("debugLevel:", debugLevel)
	fmt.Println("debugLog:", debugLog)
	fmt.Println("concurrency:", concurrency)

	// 检查飞书消息劫持开关
	interceptService := feishu.GetInterceptService()
	actualFeiShuNtf := feishuNtf
	if interceptService.IsEnabled() {
		actualFeiShuNtf = false // 劫持时不真正发送飞书消息
		fmt.Println("飞书消息劫持已开启，消息将被拦截到本地显示")
	}

	// 启用日志缓存，以便 MCP 可以获取日志
	vars.SaveQAFuncLogCache = true

	// 设置 Excel 配表资源目录：读取 function_test 配置段的 excel_resources_dir，
	// 传递给 robot 库（xcard_client.ExcelResourcesPath），使其加载与当前 pb.go 同步的 bytes，
	// 避免因工作目录不同导致配表加载失败（wrong wireType）及连锁的 nil 断言 panic。
	// 为空时 robot 库回退到默认硬编码路径 "../rain-robot/project/xcard/xcard_excel/resources"。
	if cfg, cfgErr := NewFuncCaseConfigService().GetConfig(); cfgErr == nil && cfg.ExcelResourcesDir != "" {
		xcard_client.ExcelResourcesPath = cfg.ExcelResourcesDir
	}

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	var logChan chan *log_service.Wails3Log
	if maxTimeoutPerCase <= 0 {
		maxTimeoutPerCase = 10 * time.Minute
	}
	runErr := xcard_client.RobotsYanWuQAFunctionCase(ip, port, prefix, desc, feishuGuid, loginTime, filterCasesOption, filterCaseNameOption,
		dir, filesData, concurrency, opTimeMs, actualFeiShuNtf, debugLevel, maxTimeoutPerCase,
		func(wails3LogChan chan *log_service.Wails3Log, funcCtx context.Context, funcCancel context.CancelFunc) {
			logChan = wails3LogChan
			go func() {
				for {
					select {
					case <-funcCtx.Done():
						return
					case <-ticker.C:
						if !a.isRunning.Load() {
							return
						}
						select {
						case log := <-wails3LogChan:
							a.emitRobotLog(log, seq.Load())
							seq.Add(1)
							for range len(wails3LogChan) {
								log := <-wails3LogChan
								a.emitRobotLog(log, seq.Load())
								seq.Add(1)
							}
						default:
						}
					}
				}
			}()
		}, debugLog, ctx,
	)

WAIT:
	for {
		select {
			case log, ok := <-logChan:
				if !ok {
					break WAIT
				}
				// 处理日志
				a.emitRobotLog(log, seq.Load())
				seq.Add(1)
		case <-time.After(3 * time.Second): // 等待100ms没有新内容就退出
			fmt.Println("通道没有内容，退出循环")
			break WAIT
		}
	}

	// L6 报告落盘：实现见 fight_test_report.go；须与 fight_test_report_test.go 同次 git 提交（见本包 CLAUDE.md「L6 结构化报告落盘」）。
	persistFightTestReportAfterRun(a.GetTestLogs(""), desc, prefix, dir)

	if runErr != nil {
		return fmt.Errorf("运行战斗测试失败: %w", runErr)
	}
	return nil
}

// StopRobotTest 提供停止方法
// @frontend @mcp
func (a *JsonCaseService) StopRobotTest() error {
	if a.cancelFunc != nil {
		a.cancelFunc()
		a.cancelFunc = nil
	}
	a.isRunning.Store(false)
	return nil
}

// IsRunningRobotTest 检查是否在运行
// @frontend @mcp
func (a *JsonCaseService) IsRunningRobotTest() bool {
	return a.isRunning.Load()
}

// LogEntry 日志条目
type LogEntry struct {
	Case         string `json:"case"`
	ID           int    `json:"id"`
	Level        string `json:"level"`
	Type         string `json:"type"`
	RobotName    string `json:"robotName"`
	Msg          string `json:"msg"`
	Time         string `json:"time"`
	CodeLocation string `json:"codeLocation"`
}

// GetTestLogs 获取测试日志
// @frontend @mcp
func (a *JsonCaseService) GetTestLogs(caseName string) map[string][]LogEntry {
	allLogs := log_service.GetAllLogs()
	result := make(map[string][]LogEntry)

	for cn, logs := range allLogs {
		// 如果指定了用例名，只返回该用例的日志
		if caseName != "" && cn != caseName {
			continue
		}

		var entries []LogEntry
		for _, log := range logs {
			entries = append(entries, LogEntry{
				Case:         log.Case,
				ID:           log.ID,
				Level:        log.Level.String(),
				Type:         fmt.Sprintf("%v", log.Type),
				RobotName:    log.RobotName,
				Msg:          log.Msg,
				Time:         log.Time.Format("15:04:05.000"),
				CodeLocation: log.CodeLocation,
			})
		}
		result[cn] = entries
	}

	return result
}

// ClearTestLogs 清除测试日志缓存
// @frontend @mcp
func (a *JsonCaseService) ClearTestLogs() {
	log_service.ClearLogs()
}

func (a *JsonCaseService) readJSONFile(filePath string) ([]QAFuncCase, string, error) {
	// 获取文件信息
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, "", err
	}

	// 如果是目录，返回错误（应该在调用层处理目录）
	if fileInfo.IsDir() {
		return nil, "", os.ErrNotExist
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, "", err
	}

	var testCases []QAFuncCase
	err = json.Unmarshal(data, &testCases)
	if err != nil {
		return nil, "", err
	}

	return testCases, filepath.Base(filePath), nil
}

// ReadJSONFileOrDir 读取 JSON 文件或目录，返回所有用例
func (a *JsonCaseService) ReadJSONFileOrDir(filePath string) ([]QAFuncCase, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	var allCases []QAFuncCase

	if fileInfo.IsDir() {
		// 处理目录
		categories, err := a.scanDirectory(filePath)
		if err != nil {
			return nil, err
		}

		for _, category := range categories {
			for _, caseInfo := range category.Cases {
				cases, _, err := a.readJSONFile(caseInfo.FullPath)
				if err != nil {
					continue // 跳过读取失败的文件
				}
				allCases = append(allCases, cases...)
			}
		}
	} else {
		// 处理单个文件
		cases, _, err := a.readJSONFile(filePath)
		if err != nil {
			return nil, err
		}
		allCases = cases
	}

	return allCases, nil
}

// SaveJSONFile 保存 QAFuncCase 数组到 JSON 文件
// @frontend
// filePath: 原始文件路径
// newFileName: 新文件名引用，如果为空则使用原文件名，如果不为空则替换原文件
// data: 要保存的数据
func (a *JsonCaseService) SaveJSONFile(filePath string, newFileName *string, data []QAFuncCase) error {
	// 确定最终的文件路径
	finalFilePath := filePath

	// 如果传入了新文件名，则先删除原文件，然后使用新文件名
	if newFileName != nil && *newFileName != "" {
		// 先删除原文件
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("删除原文件失败: %w", err)
		}

		// 使用新文件名
		dir := filepath.Dir(filePath)
		finalFilePath = filepath.Join(dir, *newFileName)
	}

	// 确保目录存在
	dir := filepath.Dir(finalFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	jsonData, err := json.MarshalIndent(data, " ", " ")
	if err != nil {
		return err
	}

	err = os.WriteFile(finalFilePath, jsonData, 0644)
	if err != nil {
		return err
	}

	return nil
}

// ReNameJsonFile 重命名 JSON 文件
func (a *JsonCaseService) ReNameJsonFile(oldPath, newPath string) error {
	// 检查原文件是否存在
	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		return fmt.Errorf("源文件不存在: %s", oldPath)
	}

	// 检查新路径是否已存在
	if _, err := os.Stat(newPath); err == nil {
		return fmt.Errorf("目标文件已存在: %s", newPath)
	}

	// 确保新路径的目录存在
	dir := filepath.Dir(newPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 执行重命名操作
	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("重命名文件失败: %w", err)
	}

	return nil
}

// DeleteJSONFile 删除 JSON 文件
// @frontend
func (a *JsonCaseService) DeleteJSONFile(filePath string) error {
	// 检查文件是否存在
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("文件不存在: %s", filePath)
		}
		return fmt.Errorf("检查文件状态失败: %w", err)
	}

	// 确保是文件而不是目录
	if fileInfo.IsDir() {
		return fmt.Errorf("路径是目录而不是文件: %s", filePath)
	}

	// 验证文件扩展名（可选，根据需求决定是否严格检查）
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext != ".json" {
		return fmt.Errorf("文件不是JSON格式: %s", filePath)
	}

	// 执行删除操作
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("删除文件失败: %w", err)
	}

	return nil
}

// SaveCase 保存 QAFuncCase 到 JSON 文件中，支持指定插入位置和去重
// filePath: JSON 文件路径
// newCase: 要保存的新用例
// position: 插入位置，0=第一个，1=第二个，-1=末尾，如果超出数组长度则放在末尾
// 如果存在相同 Case 名称的用例，则会覆盖原有用例并移动到新位置
func (a *JsonCaseService) SaveCase(filePath string, newCase QAFuncCase, position int) error {
	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 读取现有数据
	existingCases, _, err := a.readJSONFile(filePath)
	if err != nil {
		return err
	}

	// 检查是否存在相同 Case 名称的用例
	var updatedCases []QAFuncCase
	existingIndex := -1

	// 查找是否已存在相同 Case 名称的用例
	for i, existingCase := range existingCases {
		if existingCase.Case == newCase.Case {
			existingIndex = i
			break
		}
	}

	if existingIndex != -1 {
		// 存在重复用例，先移除原有的
		updatedCases = make([]QAFuncCase, 0, len(existingCases))
		for i, existingCase := range existingCases {
			if i != existingIndex {
				updatedCases = append(updatedCases, existingCase)
			}
		}
	} else {
		// 没有重复用例，直接使用原数据
		updatedCases = make([]QAFuncCase, len(existingCases))
		copy(updatedCases, existingCases)
	}

	// 处理插入位置
	var finalCases []QAFuncCase

	if position < 0 || position >= len(updatedCases) {
		// 位置为负数或超出范围，添加到末尾
		finalCases = append(updatedCases, newCase)
	} else if position == 0 {
		// 插入到第一个位置
		finalCases = append([]QAFuncCase{newCase}, updatedCases...)
	} else {
		// 插入到指定位置
		finalCases = make([]QAFuncCase, 0, len(updatedCases)+1)
		finalCases = append(finalCases, updatedCases[:position]...)
		finalCases = append(finalCases, newCase)
		finalCases = append(finalCases, updatedCases[position:]...)
	}

	// 序列化并保存
	err = a.SaveJSONFile(filePath, nil, finalCases)
	if err != nil {
		return err
	}

	return nil
}

// UpdateCase 按 case 名原地更新一条用例（保持数组顺序不变）
// filePath: JSON 文件路径
// caseName: 要更新的用例名（用原 Case 字段定位）
// newCase: 新的用例完整内容
// 与 SaveCase 区别：保持用例在数组中的原位置不移动；用例不存在则报错。
// 适用 read-modify-write 流程：先 get_case_by_name 读出 → 修改 → UpdateCase 写回。
// @frontend @mcp
func (a *JsonCaseService) UpdateCase(filePath string, caseName string, newCase QAFuncCase) error {
	// 读取现有数据
	existingCases, _, err := a.readJSONFile(filePath)
	if err != nil {
		return err
	}

	// 查找并原地替换（保持顺序）
	for i, c := range existingCases {
		if c.Case == caseName {
			existingCases[i] = newCase
			return a.SaveJSONFile(filePath, nil, existingCases)
		}
	}

	return fmt.Errorf("用例 %q 不存在", caseName)
}

// UpdateJSONData 更新 JSON 数据中的特定字段
func (a *JsonCaseService) UpdateJSONData(filePath string, caseName string, updates map[string]interface{}) error {
	// 读取现有数据
	data, err := a.ReadJSONFileOrDir(filePath)
	if err != nil {
		return err
	}

	// 查找并更新指定的测试用例
	for _, testCase := range data {
		if testCase.Case == caseName {
			// 这里可以根据需要添加具体的更新逻辑
			// 例如：data[i].Desc = updates["desc"].(string)
			break
		}
	}

	// 保存更新后的数据
	err = a.SaveJSONFile(filePath, nil, data)
	if err != nil {
		return err
	}

	return nil
}

// AddTestCase 添加新的测试用例到指定文件
func (a *JsonCaseService) AddTestCase(filePath string, newTestCase QAFuncCase) error {
	// 读取现有数据
	data, err := a.ReadJSONFileOrDir(filePath)
	if err != nil {
		// 如果文件不存在，创建新的数据
		data = []QAFuncCase{}
	}

	// 添加新的测试用例
	data = append(data, newTestCase)

	// 保存更新后的数据
	err = a.SaveJSONFile(filePath, nil, data)
	if err != nil {
		return err
	}

	return nil
}

// GetTestCase 获取特定测试用例（从整个目录中搜索）
// @frontend @mcp
func (a *JsonCaseService) GetTestCase(filePath string, caseName string) (*QAFuncCase, error) {
	data, err := a.ReadJSONFileOrDir(filePath)
	if err != nil {
		return nil, err
	}

	for _, testCase := range data {
		if testCase.Case == caseName {
			return &testCase, nil
		}
	}

	return nil, nil // 未找到
}

// scanDirectory 扫描目录结构
func (a *JsonCaseService) scanDirectory(dirPath string) ([]CategoryInfo, error) {
	var categories []CategoryInfo

	// 读取目录
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// 处理子目录
			categoryPath := filepath.Join(dirPath, entry.Name())
			categoryCases, err := a.scanCategory(categoryPath, entry.Name())
			if err != nil {
				continue // 跳过处理失败的目录
			}
			categories = append(categories, categoryCases)
		} else if strings.HasSuffix(strings.ToLower(entry.Name()), ".json") {
			// 处理根目录下的JSON文件（作为默认分类）
			filePath := filepath.Join(dirPath, entry.Name())
			cases, fileName, err := a.readJSONFile(filePath)
			if err != nil {
				continue
			}

			var category = "noFileName"
			if cutName, ok := strings.CutSuffix(entry.Name(), ".json"); ok {
				category = cutName
			}

			var caseInfos []CaseInfo
			for _, testCase := range cases {
				heroCount := 0
				if testCase.InitYanWu.CustomHeroes != nil {
					heroCount = len(testCase.InitYanWu.CustomHeroes)
				}

				caseInfos = append(caseInfos, CaseInfo{
					Case:       testCase.Case,
					Desc:       testCase.Desc,
					StepCount:  len(testCase.Steps),
					HeroCount:  heroCount,
					FileName:   fileName,
					Category:   category,
					FullPath:   filePath,
					QAFuncCase: testCase,
				})
			}

			categories = append(categories, CategoryInfo{
				Name:      category,
				Path:      filePath,
				CaseCount: len(caseInfos),
				Cases:     caseInfos,
			})
		}
	}

	return categories, nil
}

// scanCategory 扫描分类目录
func (a *JsonCaseService) scanCategory(categoryPath, categoryName string) (CategoryInfo, error) {
	category := CategoryInfo{
		Name: categoryName,
		Path: categoryPath,
	}

	entries, err := os.ReadDir(categoryPath)
	if err != nil {
		return category, err
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".json") {
			filePath := filepath.Join(categoryPath, entry.Name())
			cases, fileName, err := a.readJSONFile(filePath)
			if err != nil {
				continue
			}

			for _, testCase := range cases {
				heroCount := 0
				if testCase.InitYanWu.CustomHeroes != nil {
					heroCount = len(testCase.InitYanWu.CustomHeroes)
				}

				category.Cases = append(category.Cases, CaseInfo{
					Case:      testCase.Case,
					Desc:      testCase.Desc,
					StepCount: len(testCase.Steps),
					HeroCount: heroCount,
					FileName:  fileName,
					Category:  categoryName,
					FullPath:  filePath,
				})
			}
		}
	}

	category.CaseCount = len(category.Cases)
	return category, nil
}

// GetCaseList 获取用例列表（支持文件和目录）
// @mcp
func (a *JsonCaseService) GetCaseList(filePath string) ([]CaseInfo, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	var allCases []CaseInfo

	if fileInfo.IsDir() {
		// 处理目录
		categories, err := a.scanDirectory(filePath)
		if err != nil {
			return nil, err
		}

		for _, category := range categories {
			allCases = append(allCases, category.Cases...)
		}
	} else {
		// 处理单个文件
		cases, fileName, err := a.readJSONFile(filePath)
		if err != nil {
			return nil, err
		}

		for _, testCase := range cases {
			heroCount := 0
			if testCase.InitYanWu.CustomHeroes != nil {
				heroCount = len(testCase.InitYanWu.CustomHeroes)
			}

			allCases = append(allCases, CaseInfo{
				Case:      testCase.Case,
				Desc:      testCase.Desc,
				StepCount: len(testCase.Steps),
				HeroCount: heroCount,
				FileName:  fileName,
				Category:  "single_file",
				FullPath:  filePath,
			})
		}
	}

	return allCases, nil
}

// GetCategories 获取分类信息
// @frontend @mcp
func (a *JsonCaseService) GetCategories(dirPath string) ([]CategoryInfo, error) {
	fileInfo, err := os.Stat(dirPath)
	if err != nil {
		return nil, err
	}

	if !fileInfo.IsDir() {
		// 如果是文件，返回单分类
		cases, fileName, err := a.readJSONFile(dirPath)
		if err != nil {
			return nil, err
		}

		var caseInfos []CaseInfo
		for _, testCase := range cases {
			heroCount := 0
			if testCase.InitYanWu.CustomHeroes != nil {
				heroCount = len(testCase.InitYanWu.CustomHeroes)
			}

			caseInfos = append(caseInfos, CaseInfo{
				Case:       testCase.Case,
				Desc:       testCase.Desc,
				StepCount:  len(testCase.Steps),
				HeroCount:  heroCount,
				FileName:   fileName,
				Category:   fileName,
				FullPath:   dirPath,
				QAFuncCase: testCase,
			})
		}

		return []CategoryInfo{
			{
				Name:      fileName,
				Path:      filepath.Dir(dirPath),
				CaseCount: len(caseInfos),
				Cases:     caseInfos,
			},
		}, nil
	}

	return a.scanDirectory(dirPath)
}

// DeleteTestCase 删除测试用例
func (a *JsonCaseService) DeleteTestCase(filePath string, caseName string) error {
	// 读取现有数据
	data, err := a.ReadJSONFileOrDir(filePath)
	if err != nil {
		return err
	}

	// 查找并删除指定的测试用例
	var newData []QAFuncCase
	for _, testCase := range data {
		if testCase.Case != caseName {
			newData = append(newData, testCase)
		}
	}

	// 如果长度没变，说明没找到
	if len(newData) == len(data) {
		return nil // 或者返回一个错误，表示用例不存在
	}

	// 保存更新后的数据
	return a.SaveJSONFile(filePath, nil, newData)
}

// GetCaseCountResponse 获取用例数量统计返回对象
type GetCaseCountResponse struct {
	TotalCases  int
	TotalSteps  int
	TotalHeroes int
}

// GetCaseCount 获取用例数量统计
// @frontend
func (a *JsonCaseService) GetCaseCount(filePath string) (GetCaseCountResponse, error) {
	testCases, err := a.ReadJSONFileOrDir(filePath)
	if err != nil {
		return GetCaseCountResponse{}, err
	}

	totalSteps := 0
	totalHeroes := 0
	for _, testCase := range testCases {
		totalSteps += len(testCase.Steps)
		if testCase.InitYanWu.CustomHeroes != nil {
			totalHeroes += len(testCase.InitYanWu.CustomHeroes)
		}
	}

	return GetCaseCountResponse{
		TotalCases:  len(testCases),
		TotalSteps:  totalSteps,
		TotalHeroes: totalHeroes,
	}, nil
}

// SearchCases 搜索用例（支持目录搜索）
// @frontend @mcp
func (a *JsonCaseService) SearchCases(filePath string, keyword string) ([]CaseInfo, error) {
	allCases, err := a.GetCaseList(filePath)
	if err != nil {
		return nil, err
	}

	var results []CaseInfo
	for _, caseInfo := range allCases {
		// 在用例名和描述中搜索关键词
		if strings.Contains(strings.ToLower(caseInfo.Case), strings.ToLower(keyword)) ||
			strings.Contains(strings.ToLower(caseInfo.Desc), strings.ToLower(keyword)) {
			results = append(results, caseInfo)
		}
	}

	return results, nil
}

// FuncCaseConfigService 配置管理服务
// @frontend @mcp
type FuncCaseConfigService struct {
	section *appconfig.Section
}

// NewFuncCaseConfigService 创建配置服务实例
func NewFuncCaseConfigService() *FuncCaseConfigService {
	return &FuncCaseConfigService{
		section: appconfig.New("function_test"),
	}
}

// getDefaultConfig 获取默认配置
func (cs *FuncCaseConfigService) getDefaultConfig() *internal.FuncCaseConfig {
	return &internal.FuncCaseConfig{
		JsonsDir:           "cases/fight_cases",
		ServerAddr:         "10.254.114.241",
		ServerPort:         20144,
		Desc:               "本地测试",
		RobotPrefix:        "pf_qa",
		SingleCaseRunCount: 1,
		LoginTime:          1,
		RoomOpTime:         300000,
		FeiShuNtf:          false,
		FeiShuGUID:         "36732a0b-9b65-4456-8294-17044223114f",
		DebugLevel:         false,
		DebugLog:           false,
		Concurrency:        1,
		AutoSave:           true,
		InterceptEnabled:   false,
		ExcelResourcesDir:  "../rain-robot/project/xcard/xcard_excel/resources",
	}
}

// GetConfig 获取当前配置
// @frontend @mcp
func (cs *FuncCaseConfigService) GetConfig() (*internal.FuncCaseConfig, error) {
	var config internal.FuncCaseConfig
	if err := cs.section.Load(&config); err != nil {
		return nil, fmt.Errorf("读取配置失败: %v", err)
	}

	// section 不存在时返回默认值（不写盘，避免读操作产生副作用）
	if !cs.section.Exists() {
		return cs.getDefaultConfig(), nil
	}

	// 旧配置可能缺少新字段，回退到默认值
	if config.ExcelResourcesDir == "" {
		config.ExcelResourcesDir = cs.getDefaultConfig().ExcelResourcesDir
	}

	return &config, nil
}

// SaveConfig 保存配置
// @frontend @mcp
func (cs *FuncCaseConfigService) SaveConfig(config *internal.FuncCaseConfig) error {
	if config == nil {
		return fmt.Errorf("配置不能为空")
	}
	return cs.section.Save(config)
}

// UpdateConfig 更新配置（部分更新）
// @frontend @mcp
func (cs *FuncCaseConfigService) UpdateConfig(updates map[string]interface{}) (*internal.FuncCaseConfig, error) {
	// 获取当前配置
	currentConfig, err := cs.GetConfig()
	if err != nil {
		return nil, err
	}

	// 临时转换为 map 以便更新
	configMap := make(map[string]interface{})
	data, err := json.Marshal(currentConfig)
	if err != nil {
		return nil, fmt.Errorf("序列化配置失败: %w", err)
	}
	if err := json.Unmarshal(data, &configMap); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	// 应用更新
	for key, value := range updates {
		if _, exists := configMap[key]; exists {
			configMap[key] = value
		}
	}

	// 转换回 FuncCaseConfig 结构体
	updatedData, _ := json.Marshal(configMap)
	var updatedConfig internal.FuncCaseConfig
	if err := json.Unmarshal(updatedData, &updatedConfig); err != nil {
		return nil, fmt.Errorf("更新配置失败: %v", err)
	}

	// 保存更新后的配置
	if err := cs.SaveConfig(&updatedConfig); err != nil {
		return nil, err
	}

	return &updatedConfig, nil
}

// ResetToDefault 重置为默认配置
func (cs *FuncCaseConfigService) ResetToDefault() (*internal.FuncCaseConfig, error) {
	defaultConfig := cs.getDefaultConfig()
	if err := cs.SaveConfig(defaultConfig); err != nil {
		return nil, err
	}
	return defaultConfig, nil
}

// GetConfigFilePath 获取配置文件路径
func (cs *FuncCaseConfigService) GetConfigFilePath() string {
	return appconfig.FilePath()
}

// ConfigExists 检查配置是否存在
func (cs *FuncCaseConfigService) ConfigExists() bool {
	return cs.section.Exists()
}

// DeleteConfig 删除配置
func (cs *FuncCaseConfigService) DeleteConfig() error {
	return cs.section.Delete()
}

// GetConfigSummary 获取配置摘要（用于前端显示）
func (cs *FuncCaseConfigService) GetConfigSummary() (map[string]interface{}, error) {
	config, err := cs.GetConfig()
	if err != nil {
		return nil, err
	}

	summary := map[string]interface{}{
		"config_file":     appconfig.FilePath(),
		"file_exists":     cs.ConfigExists(),
		"server_address":  fmt.Sprintf("%s:%d", config.ServerAddr, config.ServerPort),
		"concurrency":     config.Concurrency,
		"test_case_count": config.SingleCaseRunCount,
		"notification":    config.FeiShuNtf,
	}

	return summary, nil
}
