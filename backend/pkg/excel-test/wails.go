package exceltest

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	internal "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/appconfig"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/feishu"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/feishu/notification"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/feishu/notification/handlers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/game"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/engine"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/ruleconfig"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/workflow"
	"github.com/xuri/excelize/v2"
)

// ========== Wails 服务 ==========

// FeishuSender 飞书发送接口（便于测试 mock）
// Deprecated: 使用 notification.CheckResultHandler 替代
type FeishuSender interface {
	SendText(robotGUID string, msg string)
}

// FeishuConfigProvider 飞书配置提供者接口
type FeishuConfigProvider interface {
	GetConfig() (*internal.FuncCaseConfig, error)
}

// defaultFeishuSender 默认飞书发送器实现
// Deprecated: 仅用于向后兼容
type defaultFeishuSender struct{}

func (s *defaultFeishuSender) SendText(robotGUID string, msg string) {
	// 不再使用文本消息，保留接口兼容
	fmt.Printf("[ExcelCheck] 发送文本消息（已废弃）: %s\n", robotGUID)
}

// SheetParseError Sheet解析错误
type SheetParseError struct {
	FileName  string `json:"fileName"`  // Excel文件名
	SheetName string `json:"sheetName"` // Sheet名称
	Error     string `json:"error"`     // 错误信息
}

// interceptServiceAdapter 劫持服务适配器
// 将 feishu.InterceptService 适配为 notification.InterceptService
type interceptServiceAdapter struct {
	service *feishu.InterceptService
	emit    notification.EventsEmit
}

// AddMessage 实现 notification.InterceptService 接口
func (a *interceptServiceAdapter) AddMessage(robotGUID, msgType, content string) notification.InterceptedMessage {
	// 调用实际的劫持服务
	msg := a.service.AddMessage(robotGUID, msgType, content)

	// 转换为 notification.InterceptedMessage 格式（Timestamp 为字符串）
	result := notification.InterceptedMessage{
		ID:        msg.ID,
		RobotGUID: msg.RobotGUID,
		MsgType:   msg.MsgType,
		Content:   msg.Content,
		Timestamp: msg.Timestamp.Format("2006-01-02 15:04:05"),
	}

	// 通过事件推送到前端
	if a.emit != nil {
		a.emit("feishu:intercepted", result)
	}

	return result
}

// ExcelCheckService Excel 检查服务
type ExcelCheckService struct {
	configProvider   FeishuConfigProvider
	feishuSender     FeishuSender // Deprecated: 保留向后兼容
	gameService      *game.GameExcelService
	dispatcher       *notification.CheckResultDispatcher // 新的事件分发器
	interceptService *feishu.InterceptService            // 劫持服务（用于测试模式）
	adapter          *interceptServiceAdapter            // 劫持适配器（用于设置事件回调）
	eventsEmit       notification.EventsEmit             // 事件发送函数（用于推送劫持消息）
}

// SetEventsEmit 设置事件发送函数（由 Wails app 初始化后调用）
// @frontend
func (e *ExcelCheckService) SetEventsEmit(emit notification.EventsEmit) {
	e.eventsEmit = emit
	if e.adapter != nil {
		e.adapter.emit = emit
	}
}

// NewExcelCheckService 创建服务（向后兼容，不支持飞书通知）
func NewExcelCheckService() *ExcelCheckService {
	dispatcher := notification.NewDispatcher()
	dispatcher.Register(handlers.NewConsoleHandler())

	return &ExcelCheckService{
		configProvider: nil,
		feishuSender:   nil,
		gameService:    game.NewGameExcelService(),
		dispatcher:     dispatcher,
	}
}

// NewExcelCheckServiceWithConfig 创建服务（支持飞书通知）
func NewExcelCheckServiceWithConfig(configProvider FeishuConfigProvider) *ExcelCheckService {
	dispatcher := notification.NewDispatcher()
	dispatcher.Register(handlers.NewConsoleHandler())

	// 获取全局劫持服务实例
	interceptSvc := feishu.GetInterceptService()

	// 创建适配器
	adapter := &interceptServiceAdapter{
		service: interceptSvc,
		emit:    nil,
	}

	// 注册通知处理器
	registerNotificationHandlers(dispatcher, configProvider, interceptSvc, adapter)

	return &ExcelCheckService{
		configProvider:   configProvider,
		feishuSender:     &defaultFeishuSender{},
		gameService:      game.NewGameExcelService(),
		dispatcher:       dispatcher,
		interceptService: interceptSvc,
		adapter:          adapter,
		eventsEmit:       nil,
	}
}

// registerNotificationHandlers 注册通知处理器（飞书卡片 + 劫持）
// 劫持处理器始终注册，通过 IsEnabled() 动态检查开关
// 飞书卡片处理器始终注册（GUID 有效时），通过 WithEnabledFunc 动态读取 FeiShuNtf 配置
func registerNotificationHandlers(
	dispatcher *notification.CheckResultDispatcher,
	configProvider FeishuConfigProvider,
	interceptSvc *feishu.InterceptService,
	adapter *interceptServiceAdapter,
) {
	// 劫持处理器始终注册（不依赖飞书配置）
	dispatcher.Register(handlers.NewInterceptHandler(
		interceptSvc.IsEnabled,
		adapter,
		func(name string, data any) {}, // 事件回调在 SetEventsEmit 中设置
	))

	// 飞书卡片处理器：始终注册（GUID 有效时），通过 WithEnabledFunc 动态读取配置
	if configProvider == nil {
		return
	}
	config, err := configProvider.GetConfig()
	if err != nil || config == nil || config.FeiShuGUID == "" || config.FeiShuGUID == "none" {
		return
	}

	dispatcher.Register(handlers.NewFeishuCardHandler(
		config.FeiShuGUID,
		handlers.WithEnabledFunc(func() bool {
			cfg, err := configProvider.GetConfig()
			if err != nil || cfg == nil {
				return false
			}
			return cfg.FeiShuNtf
		}),
	))
}

// NewExcelCheckServiceWithSender 创建服务（测试用，可注入 mock）
// Deprecated: 使用 NewExcelCheckServiceWithConfig 替代
func NewExcelCheckServiceWithSender(configProvider FeishuConfigProvider, sender FeishuSender) *ExcelCheckService {
	dispatcher := notification.NewDispatcher()
	dispatcher.Register(handlers.NewConsoleHandler())

	// 获取全局劫持服务实例
	interceptSvc := feishu.GetInterceptService()

	// 创建适配器
	adapter := &interceptServiceAdapter{
		service: interceptSvc,
		emit:    nil, // 不支持事件发送（测试模式）
	}

	// 注册通知处理器
	registerNotificationHandlers(dispatcher, configProvider, interceptSvc, adapter)

	return &ExcelCheckService{
		configProvider:   configProvider,
		feishuSender:     sender,
		gameService:      game.NewGameExcelService(),
		dispatcher:       dispatcher,
		interceptService: interceptSvc,
	}
}

func (e *ExcelCheckService) Test() string {
	return "ok1"
}

func (e *ExcelCheckService) GetERuleParam() json_rule.ERuleParam {
	return json_rule.NONE
}

// GetAllTableRuleMetas 获取所有表级规则元数据
// @frontend
func (e *ExcelCheckService) GetAllTableRuleMetas() []*json_rule.TableRuleMeta {
	return json_rule.GetAllTableRuleMetas()
}

// GetTableRuleMetasForSheet 获取指定表适用的表级规则元数据
// 根据 TargetSheets 字段过滤，只返回适用于指定 Sheet 的规则
// @frontend
func (e *ExcelCheckService) GetTableRuleMetasForSheet(sheetName string) []*json_rule.TableRuleMeta {
	return json_rule.GetTableRuleMetasForSheet(sheetName)
}

// GetAllExcelRules 获取所有Excel规则
// 从指定目录加载所有 JSON 规则配置文件
// @frontend @mcp
func (e *ExcelCheckService) GetAllExcelRules(dir string) (rules []*json_rule.SheetRule, err error) {
	return ruleconfig.LoadJsonRules(dir)
}

// SaveAllExcelRules 保存所有Excel规则
// @frontend @mcp
func (e *ExcelCheckService) SaveAllExcelRules(dir string, sheetRules []*json_rule.SheetRule) error {
	return ruleconfig.SaveCheck(dir, sheetRules)
}

// ExcelCheckResult Excel检查结果（包含列级、表级和解析错误）
type ExcelCheckResult struct {
	ColResults   []*json_rule.ColCheckResult   `json:"colResults"`   // 列级检查结果
	TableResults []*json_rule.TableCheckResult `json:"tableResults"` // 表级检查结果
	ParseErrors  []*SheetParseError            `json:"parseErrors"`  // 文件解析错误
}

// CheckAllExcelRules 执行所有Excel规则检查（包含列级和表级）
// @frontend @mcp
func (e *ExcelCheckService) CheckAllExcelRules(dir string, sheetRules []*json_rule.SheetRule) (*ExcelCheckResult, error) {
	// 使用统一工作流执行全量检查
	wfResult, err := workflow.RunCheckWorkflow(&workflow.CheckWorkflowConfig{
		ExcelPath: dir,
		Mode:      workflow.CheckModeFull,
		Rules:     sheetRules,
	})
	if err != nil {
		return nil, err
	}

	// 转换为前端需要的结构（保持 JSON 序列化兼容）
	result := &ExcelCheckResult{
		ColResults:   wfResult.ColResults,
		TableResults: wfResult.TableResults,
		ParseErrors:  convertWorkflowParseErrors(wfResult.ParseErrors),
	}

	// [DEBUG] 输出最终的 TableResults 数量
	fmt.Printf("[DEBUG] CheckAllExcelRules 完成: ColResults=%d, TableResults=%d, ParseErrors=%d\n",
		len(result.ColResults), len(result.TableResults), len(result.ParseErrors))
	for i, tr := range result.TableResults {
		fmt.Printf("[DEBUG]   TableResults[%d]: SheetName=%s, RuleType=%s, Ok=%v, Reason=%s\n",
			i, *tr.SheetName, tr.RuleType, tr.Ok, tr.Reason)
	}

	// 使用事件分发器统一处理结果（控制台输出 + 飞书通知）
	e.dispatchCheckResult(result)
	fmt.Printf("[excel-test] StoreCheckResults: ColResults=%d, TableResults=%d\n", len(result.ColResults), len(result.TableResults))
	engine.StoreCheckResults(result.ColResults, result.TableResults)
	// 通知活动 Wiki 页面刷新角标数据
	if e.eventsEmit != nil {
		e.eventsEmit("excelCheckCompleted", nil)
	}

	return result, nil
}

// CheckSingleColumn 检查单个列的规则
// 只执行指定 Sheet 中指定列的列级规则检查
// @frontend
func (e *ExcelCheckService) CheckSingleColumn(dir string, sheetName string, attrName string, colRule *json_rule.SheetColRule) (*json_rule.ColCheckResult, error) {
	// 按需加载与当前列规则相关的 Excel（跨表引用、关系链等），避免全量加载目录下所有表
	sheetMap, err := engine.GetSheetMapForColumnCheck(dir, sheetName, colRule)
	if err != nil {
		return nil, fmt.Errorf("加载 Excel 文件失败: %w", err)
	}

	// 获取目标 Sheet 的 Excel 文件
	xlsx, exist := sheetMap[sheetName]
	if !exist {
		return nil, fmt.Errorf("未找到 Sheet: %s", sheetName)
	}

	// 执行单列检查
	result, err := engine.CheckSingleColumn(xlsx, sheetName, colRule, sheetMap)
	if err != nil {
		return nil, fmt.Errorf("单列检查失败: %w", err)
	}

	return result, nil
}

// CheckIncremental 增量检查（基于本地 git diff 过滤变更文件）
// @frontend @mcp
func (e *ExcelCheckService) CheckIncremental(dir string, sheetRules []*json_rule.SheetRule) (*ExcelCheckResult, error) {
	// 使用统一工作流执行增量检查
	wfResult, err := workflow.RunCheckWorkflow(&workflow.CheckWorkflowConfig{
		ExcelPath: dir,
		Mode:      workflow.CheckModeIncremental,
		Rules:     sheetRules,
	})
	if err != nil {
		return nil, err
	}

	// 转换为前端需要的结构
	result := &ExcelCheckResult{
		ColResults:   wfResult.ColResults,
		TableResults: wfResult.TableResults,
		ParseErrors:  convertWorkflowParseErrors(wfResult.ParseErrors),
	}

	// 使用事件分发器统一处理结果
	e.dispatchCheckResult(result)

	// 缓存检查结果，供活动 Wiki 等页面读取错误计数
	engine.StoreCheckResults(result.ColResults, result.TableResults)
	// 通知活动 Wiki 页面刷新角标数据
	if e.eventsEmit != nil {
		e.eventsEmit("excelCheckCompleted", nil)
	}
	return result, nil
}

// dispatchCheckResult 使用事件分发器处理检查结果
func (e *ExcelCheckService) dispatchCheckResult(result *ExcelCheckResult) {
	// 构建检查结果事件
	event := &notification.CheckResultEvent{
		ColResults:   result.ColResults,
		TableResults: result.TableResults,
		ParseErrors:  convertToNotificationParseErrors(result.ParseErrors),
		CheckTime:    time.Now(),
	}

	// 分发事件到所有注册的通道
	e.dispatcher.Dispatch(event)
}

// convertToNotificationParseErrors 转换解析错误类型
func convertToNotificationParseErrors(errors []*SheetParseError) []*notification.SheetParseError {
	if errors == nil {
		return nil
	}
	result := make([]*notification.SheetParseError, len(errors))
	for i, e := range errors {
		result[i] = &notification.SheetParseError{
			FileName:  e.FileName,
			SheetName: e.SheetName,
			Error:     e.Error,
		}
	}
	return result
}

// convertWorkflowParseErrors 将 workflow.SheetParseError 转换为本地 SheetParseError
func convertWorkflowParseErrors(errors []*workflow.SheetParseError) []*SheetParseError {
	if errors == nil {
		return nil
	}
	result := make([]*SheetParseError, len(errors))
	for i, e := range errors {
		result[i] = &SheetParseError{
			FileName:  e.FileName,
			SheetName: e.SheetName,
			Error:     e.Error,
		}
	}
	return result
}

// sendLoadFailedNotification 加载配表失败时发送飞书通知
// Deprecated: 使用 dispatcher 机制替代
func (e *ExcelCheckService) sendLoadFailedNotification(parseErrors []*SheetParseError) {
	// 使用新的 dispatcher 机制
	if len(parseErrors) == 0 {
		return
	}

	// 构建事件并分发
	event := &notification.CheckResultEvent{
		ParseErrors: convertToNotificationParseErrors(parseErrors),
		CheckTime:   time.Now(),
	}
	e.dispatcher.Dispatch(event)
}

// filterFailedColResults 过滤出失败的列级检查结果
// Deprecated: 内部使用，保留向后兼容
func filterFailedColResults(results []*json_rule.ColCheckResult) []*json_rule.ColCheckResult {
	var failed []*json_rule.ColCheckResult
	for _, r := range results {
		if r != nil && !r.Ok {
			failed = append(failed, r)
		}
	}
	return failed
}

// filterFailedTableResults 过滤出失败的表级检查结果
// Deprecated: 内部使用，保留向后兼容
func filterFailedTableResults(results []*json_rule.TableCheckResult) []*json_rule.TableCheckResult {
	var failed []*json_rule.TableCheckResult
	for _, r := range results {
		if r != nil && !r.Ok {
			failed = append(failed, r)
		}
	}
	return failed
}

type ExcelInfo struct {
	Path   string
	Sheets []*excelio.Sheet
}

type ExcelFailInfo struct {
	Path   string
	Sheets string
	Reason string
}

// GetAllExcels 获取所有Excel文件信息
// @mcp
func (e *ExcelCheckService) GetAllExcels(dirPath string) (excelsInfo []ExcelInfo, err error) {
	excels, err := excelio.ReadFileOrDir(dirPath)
	if err != nil {
		return nil, err
	}
	defer closeExcels(excels)

	result, err := filterToExcelInfoList(excels, false)
	if err != nil {
		return nil, err
	}
	return result.([]ExcelInfo), nil
}

// GetAllExcelsConcurrent 并发获取所有Excel文件信息
// @frontend
func (e *ExcelCheckService) GetAllExcelsConcurrent(dirPath string) (excelsInfo []*ExcelInfo, err error) {
	excels, err := excelio.ReadFileOrDirConcurrent(dirPath)
	if err != nil {
		return nil, err
	}
	defer closeExcels(excels)

	result, err := filterToExcelInfoList(excels, true)
	if err != nil {
		return nil, err
	}
	return result.([]*ExcelInfo), nil
}

// closeExcels 关闭所有 Excel 文件
func closeExcels(excels []*excelize.File) {
	for _, excel := range excels {
		excel.Close()
	}
}

// filterToExcelInfoList 过滤并转换为 ExcelInfo 列表
// 返回指针切片还是值切片由 usePointer 参数控制
func filterToExcelInfoList(excels []*excelize.File, usePointer bool) (interface{}, error) {
	if len(excels) == 0 {
		if usePointer {
			return []*ExcelInfo{}, nil
		}
		return []ExcelInfo{}, nil
	}

	filter, err := excelio.ExcelFilter(excels)
	if err != nil {
		return nil, err
	}

	// 注意：不在此处发送解析错误通知，避免与 CheckAllExcelRules 重复发送
	// 解析错误通知由 CheckAllExcelRules 统一发送

	// 转为列表
	// 注意：ExcelFilter 已经过滤了非"中文|英文"格式的 sheet
	if usePointer {
		xlsxList := make([]*ExcelInfo, 0, len(filter))
		for f, s := range filter {
			xlsxList = append(xlsxList, &ExcelInfo{Path: f.Path, Sheets: s})
		}
		return xlsxList, nil
	}

	xlsxList := make([]ExcelInfo, 0, len(filter))
	for f, s := range filter {
		xlsxList = append(xlsxList, ExcelInfo{Path: f.Path, Sheets: s})
	}
	return xlsxList, nil
}

// collectParseErrors 从 filter 中收集解析错误
// 注意：ExcelFilter 已经过滤了非"中文|英文"格式的 sheet，这里直接收集所有错误
func collectParseErrors(filter map[*excelize.File][]*excelio.Sheet) []*SheetParseError {
	var errors []*SheetParseError
	for file, sheets := range filter {
		for _, sheet := range sheets {
			if sheet.Error != "" {
				errors = append(errors, &SheetParseError{
					FileName:  file.Path,
					SheetName: sheet.Name,
					Error:     sheet.Error,
				})
			}
		}
	}
	return errors
}

// SheetPreviewResult Sheet 预览结果
type SheetPreviewResult struct {
	FilePath   string           `json:"filePath"`   // 文件路径
	SheetName  string           `json:"sheetName"`  // Sheet 名称
	Headers    []string         `json:"headers"`    // 表头（中文名）
	Types      []string         `json:"types"`      // 字段类型
	FieldNames []string         `json:"fieldNames"` // 字段名
	ExportTags []string         `json:"exportTags"` // 服务端/客户端标识
	Columns    []*ColumnPreview `json:"columns"`    // 列详细信息
	DataRows   [][]string       `json:"dataRows"`   // 数据行（前 N 行）
}

// ColumnPreview 列预览信息
type ColumnPreview struct {
	CHSName   string `json:"chsName"`   // 中文名称
	FieldName string `json:"fieldName"` // 字段名
	Type      string `json:"type"`      // 字段类型
	ExportTag string `json:"exportTag"` // 导出标识
	ColStatus string `json:"colStatus"` // 列状态
}

// PreviewExcelSheet 预览 Excel Sheet 数据
// 参数:
//   - filePath: Excel 文件路径
//   - sheetName: Sheet 名称
//   - rows: 预览行数（默认 10，最小 1）
//
// @mcp
//
// 返回:
//   - Sheet 预览结果，包含表头信息和数据行
func (e *ExcelCheckService) PreviewExcelSheet(filePath string, sheetName string, rows int) (*SheetPreviewResult, error) {
	// 参数验证
	if rows < 1 {
		rows = 10
	}

	// 打开 Excel 文件
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer f.Close()

	// 获取 Sheet 数据（支持长名称：先尝试直接使用，失败则通过后缀查找）
	allRows, err := f.GetRows(sheetName)
	if err != nil {
		// 尝试通过后缀查找 Sheet（处理名称超过31字符限制的情况）
		suffix := sheetName
		if idx := strings.LastIndex(sheetName, "|"); idx >= 0 {
			suffix = sheetName[idx+1:]
		}
		// 遍历所有 Sheet 找到匹配后缀的
		allSheets := f.GetSheetList()
		for _, s := range allSheets {
			if strings.HasSuffix(s, "|"+suffix) || s == suffix {
				allRows, err = f.GetRows(s)
				sheetName = s // 更新为实际找到的 Sheet 名称
				break
			}
		}
		if err != nil {
			return nil, fmt.Errorf("读取 Sheet 失败: %w", err)
		}
	}

	// 检查行数是否足够（名将杀配表前4行是表头）
	const fixedRows = 4
	if len(allRows) < fixedRows {
		return nil, fmt.Errorf("表头行数不足，至少需要 %d 行", fixedRows)
	}

	// 解析表头
	result := &SheetPreviewResult{
		FilePath:   filePath,
		SheetName:  sheetName,
		Headers:    make([]string, 0),
		Types:      make([]string, 0),
		FieldNames: make([]string, 0),
		ExportTags: make([]string, 0),
		Columns:    make([]*ColumnPreview, 0),
		DataRows:   make([][]string, 0),
	}

	// 第1行：中文名称
	// 第2行：字段类型
	// 第3行：字段名
	// 第4行：服务端/客户端标识
	rowCHS := allRows[0]
	rowTypes := allRows[1]
	rowNames := allRows[2]
	rowExport := allRows[3]

	maxCols := len(rowTypes)
	for i := 0; i < maxCols; i++ {
		// 跳过空列（类型和字段名都为空）
		if i < len(rowTypes) && i < len(rowNames) {
			if rowTypes[i] == "" && rowNames[i] == "" {
				continue
			}
		}

		// 收集表头信息
		if i < len(rowCHS) {
			result.Headers = append(result.Headers, rowCHS[i])
		} else {
			result.Headers = append(result.Headers, "")
		}
		if i < len(rowTypes) {
			result.Types = append(result.Types, rowTypes[i])
		} else {
			result.Types = append(result.Types, "")
		}
		if i < len(rowNames) {
			result.FieldNames = append(result.FieldNames, rowNames[i])
		} else {
			result.FieldNames = append(result.FieldNames, "")
		}
		if i < len(rowExport) {
			result.ExportTags = append(result.ExportTags, rowExport[i])
		} else {
			result.ExportTags = append(result.ExportTags, "")
		}

		// 构建列信息
		col := &ColumnPreview{}
		if i < len(rowCHS) {
			col.CHSName = rowCHS[i]
		}
		if i < len(rowNames) {
			col.FieldName = rowNames[i]
		}
		if i < len(rowTypes) {
			col.Type = rowTypes[i]
		}
		if i < len(rowExport) {
			col.ExportTag = rowExport[i]
		}

		// 确定列状态
		if i < len(rowTypes) && i < len(rowNames) {
			if rowTypes[i] == "" && rowNames[i] == "" {
				col.ColStatus = "EMPTY"
			} else if rowTypes[i] == "#" {
				col.ColStatus = "COMMENT"
			} else if strings.HasPrefix(rowTypes[i], "E#") {
				col.ColStatus = "ENUM"
			} else if rowTypes[i] == "" || rowNames[i] == "" {
				col.ColStatus = "ERROR"
			} else {
				col.ColStatus = "NORMAL"
			}
		}
		result.Columns = append(result.Columns, col)
	}

	// 收集数据行（从第5行开始）
	dataStartRow := fixedRows
	dataEndRow := dataStartRow + rows
	if dataEndRow > len(allRows) {
		dataEndRow = len(allRows)
	}

	for i := dataStartRow; i < dataEndRow; i++ {
		row := allRows[i]
		// 确保行长度与列数一致
		rowData := make([]string, len(result.FieldNames))
		for j := 0; j < len(rowData) && j < len(row); j++ {
			rowData[j] = row[j]
		}
		result.DataRows = append(result.DataRows, rowData)
	}

	return result, nil
}

// ColumnInfo 列详细信息
type ColumnInfo struct {
	CHSName   string `json:"chsName"`   // 中文名称
	AttrName  string `json:"attrName"`  // 字段名
	AttrType  string `json:"attrType"`  // 字段类型
	ColStatus string `json:"colStatus"` // 列状态
	Error     string `json:"error"`     // 错误信息（如果有）
}

// SheetInfo 单个 Sheet 的信息
type SheetInfo struct {
	Name  string `json:"name"`  // Sheet 名称
	Index int    `json:"index"` // Sheet 索引（从 0 开始）
}

// GetExcelSheets 获取单个 Excel 文件中的所有 Sheet 名称
// 参数:
//   - filePath: Excel 文件路径
//
// @mcp
// 返回:
//   - Sheet 信息列表，包含名称和索引
func (e *ExcelCheckService) GetExcelSheets(filePath string) ([]*SheetInfo, error) {
	// 打开 Excel 文件
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer f.Close()

	// 获取所有 Sheet 名称
	sheetNames := f.GetSheetList()
	sheets := make([]*SheetInfo, 0, len(sheetNames))
	for i, name := range sheetNames {
		sheets = append(sheets, &SheetInfo{
			Name:  name,
			Index: i,
		})
	}

	return sheets, nil
}

// GetExcelColumnInfo 获取 Excel Sheet 的列详细信息
// 参数:
//   - filePath: Excel 文件路径
//   - sheetName: Sheet 名称
//
// @mcp
// 返回:
//   - 列属性列表，包含每列的中文名、类型、字段名和状态
func (e *ExcelCheckService) GetExcelColumnInfo(filePath string, sheetName string) ([]*ColumnInfo, error) {
	// 打开 Excel 文件
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer f.Close()

	// 获取 Sheet 数据
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("读取 Sheet 失败: %w", err)
	}

	// 检查行数是否足够
	const fixedRows = 4
	if len(rows) < fixedRows {
		return nil, fmt.Errorf("表头行数不足，至少需要 %d 行", fixedRows)
	}

	// 解析表头行
	rowCHS := rows[0]
	rowTypes := rows[1]
	rowNames := rows[2]
	rowExport := rows[3]

	// 构建列信息列表
	columns := make([]*ColumnInfo, 0, len(rowTypes))

	for i := 0; i < len(rowTypes); i++ {
		col := &ColumnInfo{}

		// 中文名称
		if i < len(rowCHS) {
			col.CHSName = rowCHS[i]
		}

		// 字段名
		if i < len(rowNames) {
			col.AttrName = rowNames[i]
		}

		// 字段类型
		if i < len(rowTypes) {
			col.AttrType = rowTypes[i]
		}

		// 确定列状态和错误信息
		if i < len(rowTypes) && i < len(rowNames) {
			if rowTypes[i] == "" && rowNames[i] == "" {
				col.ColStatus = "EMPTY"
			} else if rowTypes[i] == "#" {
				col.ColStatus = "COMMENT"
			} else if strings.HasPrefix(rowTypes[i], "E#") {
				col.ColStatus = "ENUM"
			} else if rowTypes[i] == "" || rowNames[i] == "" {
				col.ColStatus = "ERROR"
				col.Error = fmt.Sprintf("类型:[%s]和名称:[%s]不对应", rowTypes[i], rowNames[i])
			} else {
				col.ColStatus = "NORMAL"
			}
		}

		// 检查导出标识
		if i < len(rowExport) {
			exportTag := rowExport[i]
			if exportTag != "" && exportTag != "server" && exportTag != "client" &&
				exportTag != "server/client" && exportTag != "client/server" {
				col.ColStatus = "ERROR"
				col.Error = fmt.Sprintf("[%s]不为空、server、client或server/client、client/server", exportTag)
			}
		}

		columns = append(columns, col)
	}

	return columns, nil
}

// checkDefaultTableRules 执行默认表级规则检查
// 为 sheetMap 中所有有默认规则的表执行检查，无需 JSON 配置
func (e *ExcelCheckService) checkDefaultTableRules(sheetMap map[string]*excelize.File) []*json_rule.TableCheckResult {
	results := make([]*json_rule.TableCheckResult, 0)

	// 获取所有可用的表级规则元数据
	allMetas := json_rule.GetAllTableRuleMetas()

	// 遍历 sheetMap 中的每个表
	for sheetName, xlsxFile := range sheetMap {
		// 获取该表的默认规则
		defaultRuleTypes := json_rule.GetDefaultTableRulesForSheet(sheetName)
		if len(defaultRuleTypes) == 0 {
			continue
		}

		// 获取表的列数据
		cols, err := xlsxFile.GetCols(sheetName)
		if err != nil {
			continue
		}

		// 对每个默认规则执行检查
		for _, ruleType := range defaultRuleTypes {
			// 查找规则的元数据
			var meta *json_rule.TableRuleMeta
			for _, m := range allMetas {
				if m.Type == ruleType {
					meta = m
					break
				}
			}
			if meta == nil {
				continue
			}

			// 检查规则是否适用于当前表
			if !isRuleApplicableToSheet(sheetName, meta.TargetSheets) {
				continue
			}

			// 获取检查器
			checker := engine.TableManager.GetChecker(meta.Type)
			if checker == nil {
				continue
			}

			// 使用默认参数执行检查
			params := meta.ResolveParams(nil)

			// 使用 ID 列确定数据结束位置（与 excel_func.go CheckTableRules 保持一致）
			endIndex := helpers.GetColEndIndex(cols, 0, excelio.MJS_FIXED_ROWS_NUM,
				helpers.ParseBreakLine(nil), nil)

			result := checker.Check(json_rule.CheckParam{
				SheetName:   sheetName,
				Cols:        cols,
				StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
				EndIndex:    endIndex,
				Params:      params,
				SheetMap:    sheetMap,
			})
			if result != nil {
				result.SheetName = &sheetName
				result.RuleType = meta.Type
				result.DisplayName = meta.DisplayName
				result.TableName = &xlsxFile.Path
				results = append(results, result)
			}
		}
	}

	return results
}

// isRuleApplicableToSheet 检查规则是否适用于指定的表
// 支持后缀匹配：targetSheet "ArenaSeason" 可以匹配 "竞技场赛季|ArenaSeason"
func isRuleApplicableToSheet(sheetName string, targetSheets []string) bool {
	if len(targetSheets) == 0 {
		// 如果没有指定目标表，则适用于所有表
		return true
	}

	for _, target := range targetSheets {
		// 精确匹配
		if sheetName == target {
			return true
		}
		// 后缀匹配：支持 "中文|英文" 格式
		if strings.HasSuffix(sheetName, "|"+target) {
			return true
		}
	}

	return false
}

// filterSheetMapByGitChanges 根据 git 最新提交变更的文件过滤 sheetMap
// 只保留 HEAD~1 到 HEAD 之间变更的 Excel 文件对应的表
// 如果获取 git 变更失败（非 git 仓库等），回退到返回原始 sheetMap
func (e *ExcelCheckService) filterSheetMapByGitChanges(sheetMap map[string]*excelize.File, excelDir string) map[string]*excelize.File {
	//// 获取 git 仓库根目录
	//repoRoot, err := helpers.GetGitRepoRoot(excelDir)
	//if err != nil {
	//	fmt.Printf("[INFO] 获取 git 仓库失败，检查所有表: %v\n", err)
	//	return sheetMap
	//}
	//
	//// 获取 HEAD~1 到 HEAD 之间变更的 Excel 文件
	//changedFiles, err := helpers.GetGitChangedExcelFiles(repoRoot, "HEAD~1", ".xlsx")
	//if err != nil {
	//	fmt.Printf("[INFO] 获取 git 变更文件失败，检查所有表: %v\n", err)
	//	return sheetMap
	//}
	//
	//if len(changedFiles) == 0 {
	//	fmt.Printf("[INFO] 没有发现 git 变更的 Excel 文件，跳过表级变更检查\n")
	//	return make(map[string]*excelize.File)
	//}
	//
	//// 使用 PathNormalizer 统一路径格式
	//normalizer := helpers.NewPathNormalizer(excelDir)
	//
	//// 构建变更文件的规范化路径集合
	//changedSet := make(map[string]bool, len(changedFiles))
	//for _, f := range changedFiles {
	//	normalized := normalizer.Normalize(f.AbsPath)
	//	changedSet[normalized] = true
	//	fmt.Printf("[DEBUG] 变更文件: %s\n", normalized)
	//}
	//
	//// 过滤 sheetMap，只保留变更文件对应的表
	//filtered := make(map[string]*excelize.File)
	//for sheetName, xlsxFile := range sheetMap {
	//	normalizedPath := normalizer.Normalize(xlsxFile.Path)
	//	fmt.Printf("[DEBUG] 检查表: %s, 文件路径(规范化后): %s\n", sheetName, normalizedPath)
	//	if changedSet[normalizedPath] {
	//		filtered[sheetName] = xlsxFile
	//		fmt.Printf("[DEBUG]   -> 匹配成功 ✓\n")
	//	} else {
	//		fmt.Printf("[DEBUG]   -> 未匹配 ✗\n")
	//	}
	//}
	//
	//fmt.Printf("[INFO] git 变更 Excel 文件 %d 个，过滤后待检查 %d 个表（原始 %d 个）\n",
	//	len(changedFiles), len(filtered), len(sheetMap))
	//
	//return filtered
	return nil
}

// checkGeneralTableRules 执行通用表级规则检查
// 通用规则是指 TargetSheets 为空的规则，它们自动应用于所有表
// 包括：NEW_ROW_NOTIFY（新增行/列通知）、ROW_CHANGE_NOTIFY（行变更字段通知）
func (e *ExcelCheckService) checkGeneralTableRules(sheetMap map[string]*excelize.File) []*json_rule.TableCheckResult {
	//results := make([]*json_rule.TableCheckResult, 0)
	//
	//// 获取所有表级规则元数据
	//allMetas := json_rule.GetAllTableRuleMetas()
	//fmt.Printf("[DEBUG] GetAllTableRuleMetas() 返回 %d 条规则\n", len(allMetas))
	//for _, meta := range allMetas {
	//	fmt.Printf("[DEBUG]   - %s: TargetSheets=%v\n", meta.Type, meta.TargetSheets)
	//}
	//
	//// 筛选通用规则（TargetSheets 为空）
	//var generalMetas []*json_rule.TableRuleMeta
	//for _, meta := range allMetas {
	//	if len(meta.TargetSheets) == 0 {
	//		generalMetas = append(generalMetas, meta)
	//	}
	//}
	//fmt.Printf("[DEBUG] 筛选后通用规则: %d 条\n", len(generalMetas))
	//for _, meta := range generalMetas {
	//	fmt.Printf("[DEBUG]   - %s (%s)\n", meta.Type, meta.DisplayName)
	//}
	//
	//// 如果没有通用规则，直接返回
	//if len(generalMetas) == 0 {
	//	return results
	//}
	//
	//// 对每个 sheet 执行通用规则
	//sheetCount := 0
	//for sheetName, xlsxFile := range sheetMap {
	//	// 限制调试输出，只打印前10个sheet
	//	if sheetCount < 10 {
	//		fmt.Printf("[DEBUG] 处理 sheet: %s (文件: %s)\n", sheetName, xlsxFile.Path)
	//	}
	//	sheetCount++
	//
	//	// 获取表的列数据
	//	cols, err := xlsxFile.GetCols(sheetName)
	//	if err != nil {
	//		fmt.Printf("[DEBUG]   GetCols 失败: %v\n", err)
	//		continue
	//	}
	//
	//	// 尝试获取 git 仓库根目录（而非文件所在目录）
	//	gitRepoPath := filepath.Dir(xlsxFile.Path)
	//	// 使用文件所在目录来查找 git 仓库根目录
	//	if repoRoot, err := helpers.GetGitRepoRoot(gitRepoPath); err == nil {
	//		gitRepoPath = repoRoot
	//	} else {
	//		// 不在 git 仓库中，跳过通用规则检查
	//		fmt.Printf("[DEBUG]   不在 git 仓库中，跳过 (path=%s): %v\n", gitRepoPath, err)
	//		continue
	//	}
	//
	//	// 对每个通用规则执行检查
	//	for _, meta := range generalMetas {
	//		// 获取检查器
	//		checker := engine.TableManager.GetChecker(meta.Type)
	//		if checker == nil {
	//			fmt.Printf("[DEBUG]   未找到检查器: %s\n", meta.Type)
	//			continue
	//		}
	//
	//		// 构建默认参数
	//		params := meta.ResolveParams(nil)
	//
	//		// 自动设置 git 相关参数
	//		// 使用 git diff 模式，仓库路径优先使用 git 根目录
	//		params["useGitDiff"] = "true"
	//		params["gitRepoPath"] = gitRepoPath
	//
	//		// 执行检查
	//		result := checker.Check(json_rule.CheckParam{
	//			SheetName:   sheetName,
	//			Cols:        cols,
	//			StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
	//			Params:      params,
	//			SheetMap:    sheetMap,
	//		})
	//		if result == nil {
	//			continue
	//		}
	//
	//		// 只添加有实际内容的结果（过滤掉"首次运行"和"无变更"的噪音）
	//		// 对于通知规则（Ok=true）：Reason 非空表示有变更汇总
	//		// 对于错误规则（Ok=false）：ErrCells 非空表示有错误详情
	//		hasContent := false
	//		if result.Ok {
	//			// 通知规则：检查 Reason 是否有内容
	//			hasContent = result.Reason != ""
	//		} else {
	//			// 错误规则：检查 ErrCells 是否有内容
	//			hasContent = len(result.ErrCells) > 0
	//		}
	//
	//		// 限制调试输出，只打印有内容的结果
	//		if sheetCount <= 10 || hasContent {
	//			fmt.Printf("[DEBUG]   检查结果: %s, Ok=%v, Reason长度=%d, ErrCells数=%d, hasContent=%v\n",
	//				meta.Type, result.Ok, len(result.Reason), len(result.ErrCells), hasContent)
	//		}
	//
	//		if hasContent {
	//			result.SheetName = &sheetName
	//			result.RuleType = meta.Type
	//			result.DisplayName = meta.DisplayName
	//			result.TableName = &xlsxFile.Path
	//			results = append(results, result)
	//		}
	//	}
	//}
	//
	//fmt.Printf("[DEBUG] checkGeneralTableRules 完成，处理了 %d 个 sheet，返回 %d 条结果\n", sheetCount, len(results))
	//return results
	return nil
}

// ========== 创建 Excel 文件相关功能 ==========

// ExcelColumnDef Excel 列定义
type ExcelColumnDef struct {
	CHSName   string `json:"chsName"`   // 中文名称
	FieldType string `json:"fieldType"` // 字段类型
	FieldName string `json:"fieldName"` // 字段名
	ExportTag string `json:"exportTag"` // 导出标识: server/client/server/client
}

// ExcelSheetDef Excel Sheet 定义
type ExcelSheetDef struct {
	SheetName string            `json:"sheetName"` // Sheet 名称
	Columns   []*ExcelColumnDef `json:"columns"`   // 列定义
	DataRows  [][]string        `json:"dataRows"`  // 数据行
}

// CreateExcelFile 创建 Excel 文件
// 参数:
//   - filePath: 输出文件路径
//   - sheets: Sheet 定义数组
//
// @mcp
// 返回:
//   - 创建的文件路径，如果有错误则返回错误信息
func (e *ExcelCheckService) CreateExcelFile(filePath string, sheets []*ExcelSheetDef) (string, error) {
	if filePath == "" {
		return "", fmt.Errorf("文件路径不能为空")
	}
	if len(sheets) == 0 {
		return "", fmt.Errorf("至少需要一个 Sheet")
	}

	// 创建新文件
	f := excelize.NewFile()

	// 删除默认 Sheet
	defaultSheet := "Sheet1"
	sheetList := f.GetSheetList()
	if len(sheetList) > 0 {
		f.DeleteSheet(defaultSheet)
	}

	// 创建每个 Sheet
	for _, sheetDef := range sheets {
		if err := createSheet(f, sheetDef); err != nil {
			return "", fmt.Errorf("创建 Sheet %s 失败: %w", sheetDef.SheetName, err)
		}
	}

	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("创建目录失败: %w", err)
	}

	// 保存文件
	if err := f.SaveAs(filePath); err != nil {
		return "", fmt.Errorf("保存文件失败: %w", err)
	}
	f.Close()

	return filePath, nil
}

// createSheet 创建单个 Sheet
func createSheet(f *excelize.File, sheetDef *ExcelSheetDef) error {
	// 创建新 Sheet
	index, err := f.NewSheet(sheetDef.SheetName)
	if err != nil {
		return fmt.Errorf("创建 Sheet 失败: %w", err)
	}

	// 设置为活动 Sheet
	f.SetActiveSheet(index)

	// 将列按行组织（名将杀配表格式：前4行是表头）
	// 第1行：中文名称
	// 第2行：字段类型
	// 第3行：字段名
	// 第4行：导出标识
	rowCHS := make([]string, len(sheetDef.Columns))
	rowTypes := make([]string, len(sheetDef.Columns))
	rowNames := make([]string, len(sheetDef.Columns))
	rowExport := make([]string, len(sheetDef.Columns))

	for i, col := range sheetDef.Columns {
		rowCHS[i] = col.CHSName
		rowTypes[i] = col.FieldType
		rowNames[i] = col.FieldName
		rowExport[i] = col.ExportTag
	}

	// 写入表头（4行）
	for colIdx, value := range rowCHS {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 1)
		f.SetCellValue(sheetDef.SheetName, cell, value)
	}
	for colIdx, value := range rowTypes {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 2)
		f.SetCellValue(sheetDef.SheetName, cell, value)
	}
	for colIdx, value := range rowNames {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 3)
		f.SetCellValue(sheetDef.SheetName, cell, value)
	}
	for colIdx, value := range rowExport {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 4)
		f.SetCellValue(sheetDef.SheetName, cell, value)
	}

	// 写入数据行（从第5行开始）
	dataStartRow := 5
	for rowIdx, rowData := range sheetDef.DataRows {
		actualRow := dataStartRow + rowIdx
		for colIdx, value := range rowData {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, actualRow)
			f.SetCellValue(sheetDef.SheetName, cell, value)
		}
	}

	return nil
}

// ========== 过滤查询相关功能 ==========

// FilterCondition 过滤条件
type FilterCondition struct {
	ColumnName string `json:"columnName"` // 列名（字段名或中文名）
	Value      string `json:"value"`      // 匹配值
	Operator   string `json:"operator"`   // 操作符: eq(等于), neq(不等于), contains(包含), startsWith(开头), endsWith(结尾)
}

// FilterResult 过滤查询结果
type FilterResult struct {
	FilePath    string     `json:"filePath"`    // 文件路径
	SheetName   string     `json:"sheetName"`   // Sheet 名称
	TotalRows   int        `json:"totalRows"`   // 总行数
	MatchedRows int        `json:"matchedRows"` // 匹配行数
	Headers     []string   `json:"headers"`     // 表头（如果有）
	DataRows    [][]string `json:"dataRows"`    // 匹配的数据行
}

// FilterExcelData 过滤 Excel 数据
// 参数:
//   - filePath: Excel 文件路径
//   - sheetName: Sheet 名称
//   - conditions: 过滤条件数组（多个条件为 AND 关系）
//   - includeHeader: 是否包含表头行
//
// 返回:
// @mcp
//   - 过滤结果，包含匹配的行数和数据
func (e *ExcelCheckService) FilterExcelData(filePath string, sheetName string, conditions []*FilterCondition, includeHeader bool) (*FilterResult, error) {
	// 打开 Excel 文件
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer f.Close()

	// 获取 Sheet 数据
	allRows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("读取 Sheet 失败: %w", err)
	}

	// 检查行数是否足够
	const fixedRows = 4
	if len(allRows) < fixedRows {
		return nil, fmt.Errorf("表头行数不足，至少需要 %d 行", fixedRows)
	}

	// 解析表头，获取列索引映射
	colIndexMap := make(map[string]int) // 字段名/中文名 -> 列索引
	rowNames := allRows[2]              // 第3行：字段名
	rowCHS := allRows[0]                // 第1行：中文名

	for i := 0; i < len(rowNames); i++ {
		if rowNames[i] != "" {
			colIndexMap[rowNames[i]] = i
		}
		if i < len(rowCHS) && rowCHS[i] != "" {
			colIndexMap[rowCHS[i]] = i
		}
	}

	// 构建结果
	result := &FilterResult{
		FilePath:  filePath,
		SheetName: sheetName,
		TotalRows: len(allRows) - fixedRows,
	}

	// 如果需要包含表头
	if includeHeader {
		result.Headers = make([]string, 0)
		for i := 0; i < len(rowNames); i++ {
			if i < len(rowCHS) {
				result.Headers = append(result.Headers, rowCHS[i])
			} else {
				result.Headers = append(result.Headers, "")
			}
		}
	}

	// 过滤数据行（从第5行开始，索引为4）
	dataStartRow := fixedRows
	for rowIdx := dataStartRow; rowIdx < len(allRows); rowIdx++ {
		row := allRows[rowIdx]

		// 检查是否满足所有条件
		matched := true
		for _, cond := range conditions {
			colIdx, exists := colIndexMap[cond.ColumnName]
			if !exists {
				matched = false
				break
			}

			// 获取单元格值
			cellValue := ""
			if colIdx < len(row) {
				cellValue = row[colIdx]
			}

			// 根据操作符判断
			if !matchCondition(cellValue, cond.Value, cond.Operator) {
				matched = false
				break
			}
		}

		if matched {
			result.DataRows = append(result.DataRows, row)
		}
	}

	result.MatchedRows = len(result.DataRows)
	return result, nil
}

// matchCondition 判断单元格值是否满足条件
func matchCondition(cellValue, targetValue, operator string) bool {
	switch operator {
	case "", "eq": // 默认等于
		return cellValue == targetValue
	case "neq": // 不等于
		return cellValue != targetValue
	case "contains": // 包含
		return strings.Contains(cellValue, targetValue)
	case "startsWith": // 开头
		return strings.HasPrefix(cellValue, targetValue)
	case "endsWith": // 结尾
		return strings.HasSuffix(cellValue, targetValue)
	default:
		return cellValue == targetValue
	}
}

// ========== 范围查询相关功能 ==========

// RangeQueryResult 范围查询结果
type RangeQueryResult struct {
	FilePath   string     `json:"filePath"`   // 文件路径
	SheetName  string     `json:"sheetName"`  // Sheet 名称
	StartRow   int        `json:"startRow"`   // 起始行号（数据行，从1开始）
	EndRow     int        `json:"endRow"`     // 结束行号
	TotalRows  int        `json:"totalRows"`  // 总行数
	ReturnRows int        `json:"returnRows"` // 返回行数
	Headers    []string   `json:"headers"`    // 表头（如果有）
	DataRows   [][]string `json:"dataRows"`   // 数据行
}

// QueryExcelRange 查询指定范围的 Excel 数据
// 参数:
//   - filePath: Excel 文件路径
//   - sheetName: Sheet 名称
//   - startRow: 起始行号（数据行从1开始，1表示第5行实际数据）
//   - endRow: 结束行号（包含此行），0 或 -1 表示到末尾
//   - includeHeader: 是否包含表头行
//
// 返回:
//   - 范围查询结果
//
// @mcp
func (e *ExcelCheckService) QueryExcelRange(filePath string, sheetName string, startRow int, endRow int, includeHeader bool) (*RangeQueryResult, error) {
	// 打开 Excel 文件
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer f.Close()

	// 获取 Sheet 数据（支持长名称：先尝试直接使用，失败则通过后缀查找）
	allRows, err := f.GetRows(sheetName)
	if err != nil {
		// 尝试通过后缀查找 Sheet（处理名称超过31字符限制的情况）
		suffix := sheetName
		if idx := strings.LastIndex(sheetName, "|"); idx >= 0 {
			suffix = sheetName[idx+1:]
		}
		// 遍历所有 Sheet 找到匹配后缀的
		allSheets := f.GetSheetList()
		for _, s := range allSheets {
			if strings.HasSuffix(s, "|"+suffix) || s == suffix {
				allRows, err = f.GetRows(s)
				sheetName = s // 更新为实际找到的 Sheet 名称
				break
			}
		}
		if err != nil {
			return nil, fmt.Errorf("读取 Sheet 失败: %w", err)
		}
	}

	// 检查行数是否足够
	const fixedRows = 4
	if len(allRows) < fixedRows {
		return nil, fmt.Errorf("表头行数不足，至少需要 %d 行", fixedRows)
	}

	// 计算实际数据行
	dataStartRow := fixedRows // 数据从第5行开始（索引4）
	totalDataRows := len(allRows) - fixedRows

	// 参数校验和调整
	if startRow < 1 {
		startRow = 1
	}
	if startRow > totalDataRows {
		return &RangeQueryResult{
			FilePath:   filePath,
			SheetName:  sheetName,
			StartRow:   startRow,
			EndRow:     startRow - 1,
			TotalRows:  totalDataRows,
			ReturnRows: 0,
			DataRows:   [][]string{},
		}, nil
	}

	// 结束行处理：0 或 -1 表示到末尾
	if endRow <= 0 || endRow > totalDataRows {
		endRow = totalDataRows
	}

	// 确保 startRow <= endRow
	if startRow > endRow {
		startRow, endRow = endRow, startRow
	}

	// 构建结果
	result := &RangeQueryResult{
		FilePath:  filePath,
		SheetName: sheetName,
		StartRow:  startRow,
		EndRow:    endRow,
		TotalRows: totalDataRows,
	}

	// 如果需要包含表头
	if includeHeader {
		rowCHS := allRows[0]
		result.Headers = make([]string, len(rowCHS))
		copy(result.Headers, rowCHS)
	}

	// 提取指定范围的数据行
	for i := startRow - 1; i < endRow; i++ {
		actualRowIndex := dataStartRow + i
		if actualRowIndex < len(allRows) {
			result.DataRows = append(result.DataRows, allRows[actualRowIndex])
		}
	}

	result.ReturnRows = len(result.DataRows)
	return result, nil
}

// ========== ExcelConfigService 配置管理服务 ==========

// ExcelConfigService 配置管理服务
type ExcelConfigService struct {
	section *appconfig.Section
}

// ExcelConfig 应用配置结构
type ExcelConfig struct {
	ExcelResourceDir string `json:"excel_resources_dir"`
	ExcelCaseDir     string `json:"excel_case_dir"`
	ClientPath       string `json:"client_path"` // 客户端项目路径，用于资源校验
}

// NewExcelConfigService 创建配置服务实例
func NewExcelConfigService() *ExcelConfigService {
	return &ExcelConfigService{
		section: appconfig.New("excel_test"),
	}
}

// getDefaultConfig 获取默认配置
func (cs *ExcelConfigService) getDefaultConfig() *ExcelConfig {
	return &ExcelConfig{
		ExcelResourceDir: "excel_resources",
		ExcelCaseDir:     "cases/excel_cases",
	}
}

// GetConfig 获取当前配置
// @frontend @mcp
func (cs *ExcelConfigService) GetConfig() (*ExcelConfig, error) {
	var config ExcelConfig
	if err := cs.section.Load(&config); err != nil {
		return nil, fmt.Errorf("读取配置失败: %v", err)
	}

	if !cs.section.Exists() {
		return cs.getDefaultConfig(), nil
	}

	return &config, nil
}

// SaveConfig 保存配置
// @frontend @mcp
func (cs *ExcelConfigService) SaveConfig(config *ExcelConfig) error {
	if config == nil {
		return fmt.Errorf("配置不能为空")
	}
	return cs.section.Save(config)
}

// UpdateConfig 更新配置（部分更新）
func (cs *ExcelConfigService) UpdateConfig(updates map[string]interface{}) (*ExcelConfig, error) {
	// 获取当前配置
	currentConfig, err := cs.GetConfig()
	if err != nil {
		return nil, err
	}

	// 临时转换为 map 以便更新
	configMap := make(map[string]interface{})
	data, _ := json.Marshal(currentConfig)
	json.Unmarshal(data, &configMap)

	// 应用更新
	for key, value := range updates {
		if _, exists := configMap[key]; exists {
			configMap[key] = value
		}
	}

	// 转换回 ExcelConfig 结构体
	updatedData, _ := json.Marshal(configMap)
	var updatedConfig ExcelConfig
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
func (cs *ExcelConfigService) ResetToDefault() (*ExcelConfig, error) {
	defaultConfig := cs.getDefaultConfig()
	if err := cs.SaveConfig(defaultConfig); err != nil {
		return nil, err
	}
	return defaultConfig, nil
}

// GetConfigFilePath 获取配置文件路径
func (cs *ExcelConfigService) GetConfigFilePath() string {
	return appconfig.FilePath()
}

// ConfigExists 检查配置是否存在
func (cs *ExcelConfigService) ConfigExists() bool {
	return cs.section.Exists()
}

// DeleteConfig 删除配置
func (cs *ExcelConfigService) DeleteConfig() error {
	return cs.section.Delete()
}

// ========== ExcelTestGameService Excel 测试页面专用的 Game 服务包装器 ==========

// ExcelTestGameService Excel 测试页面专用的 Game 服务包装器
// 只暴露 Excel 测试页面需要的方法，遵循接口隔离原则
type ExcelTestGameService struct {
	core *game.GameExcelService
}

// NewExcelTestGameService 创建 Excel 测试页面的 Game 服务实例
func NewExcelTestGameService() *ExcelTestGameService {
	return &ExcelTestGameService{
		core: game.NewGameExcelService(),
	}
}

// GetAllHeroCfg 获取所有英雄配置
// @frontend @mcp
func (s *ExcelTestGameService) GetAllHeroCfg() interface{} {
	return s.core.GetAllHeroCfg()
}

// ========== 表变更检测相关功能 ==========

// TableChangeResult 表变更检测结果
type TableChangeResult struct {
	HasChanges    bool             `json:"hasChanges"`    // 是否有变更
	IsNewFile     bool             `json:"isNewFile"`     // 是否是新文件（上一个 commit 不存在）
	CommitHash    string           `json:"commitHash"`    // 基准 commit hash
	CommitMessage string           `json:"commitMessage"` // commit 消息
	AddedRows     []*RowChangeInfo `json:"addedRows"`     // 新增行
	RemovedRows   []*RowChangeInfo `json:"removedRows"`   // 删除行
	ModifiedRows  []*RowModifyInfo `json:"modifiedRows"`  // 修改行
	AddedCols     []string         `json:"addedCols"`     // 新增列
	RemovedCols   []string         `json:"removedCols"`   // 删除列
}

// RowChangeInfo 行变更信息
type RowChangeInfo struct {
	RowId   string `json:"rowId"`   // 行 ID
	RowName string `json:"rowName"` // 行名称
}

// RowModifyInfo 行修改信息
type RowModifyInfo struct {
	RowId   string         `json:"rowId"`   // 行 ID
	RowName string         `json:"rowName"` // 行名称
	Changes []*FieldChange `json:"changes"` // 字段变更列表
}

// FieldChange 字段变更信息
type FieldChange struct {
	ColName  string `json:"colName"`  // 列名
	OldValue string `json:"oldValue"` // 旧值
	NewValue string `json:"newValue"` // 新值
}

// GetTableChanges 获取表变更信息
// 使用 git diff 对比当前 Excel 文件与指定 commit 的差异
// @mcp
func (e *ExcelCheckService) GetTableChanges(excelPath, sheetName, baseCommit, idColName, nameColName string) (*TableChangeResult, error) {
	//result := &TableChangeResult{
	//	HasChanges:   false,
	//	IsNewFile:    false,
	//	AddedRows:    make([]*RowChangeInfo, 0),
	//	RemovedRows:  make([]*RowChangeInfo, 0),
	//	ModifiedRows: make([]*RowModifyInfo, 0),
	//	AddedCols:    make([]string, 0),
	//	RemovedCols:  make([]string, 0),
	//}
	//
	//// 获取 git 仓库根目录
	//repoPath := filepath.Dir(excelPath)
	//gitRoot, err := helpers.GetGitRepoRoot(repoPath)
	//if err != nil {
	//	return nil, fmt.Errorf("路径不在 git 仓库中: %s", repoPath)
	//}
	//repoPath = gitRoot
	//
	//// 获取当前 commit 信息
	//hash, message, _, err := helpers.GetCurrentCommitInfo(repoPath)
	//if err != nil {
	//	return nil, fmt.Errorf("获取 commit 信息失败: %w", err)
	//}
	//result.CommitHash = hash
	//result.CommitMessage = message
	//
	//// 获取旧版本文件内容
	//oldContent, err := helpers.GetFileAtCommit(repoPath, baseCommit, excelPath)
	//if err != nil {
	//	return nil, fmt.Errorf("获取旧版本文件失败: %w", err)
	//}
	//
	//// 旧文件不存在，表示这是新增文件
	//if oldContent == nil {
	//	result.IsNewFile = true
	//	return result, nil
	//}
	//
	//// 解析当前 Excel 文件
	//currentFile, err := excelize.OpenFile(excelPath)
	//if err != nil {
	//	return nil, fmt.Errorf("打开 Excel 文件失败: %w", err)
	//}
	//defer currentFile.Close()
	//
	//// 获取当前 sheet 的列数据
	//currentCols, err := currentFile.GetCols(sheetName)
	//if err != nil {
	//	return nil, fmt.Errorf("获取 sheet 列数据失败: %w", err)
	//}
	//
	//// 解析旧版本 Excel
	//oldCols, err := helpers.ParseExcelFromBytes(oldContent, sheetName)
	//if err != nil {
	//	// 旧版本解析失败（可能是 sheet 不存在），返回空结果
	//	return result, nil
	//}
	//
	//// 构建快照并检测差异
	//oldSnapshot := diff.BuildSnapshot(sheetName, oldCols, excelio.MJS_FIXED_ROWS_NUM, idColName, nameColName)
	//diffResult := diff.DetectDiff(oldSnapshot, currentCols, excelio.MJS_FIXED_ROWS_NUM, idColName, nameColName)
	//
	//// 转换差异结果
	//result.HasChanges = diffResult.HasChanges()
	//
	//for _, row := range diffResult.AddedRows {
	//	result.AddedRows = append(result.AddedRows, &RowChangeInfo{
	//		RowId:   row.RowId,
	//		RowName: row.RowName,
	//	})
	//}
	//
	//for _, row := range diffResult.RemovedRows {
	//	result.RemovedRows = append(result.RemovedRows, &RowChangeInfo{
	//		RowId:   row.RowId,
	//		RowName: row.RowName,
	//	})
	//}
	//
	//for _, row := range diffResult.ModifiedRows {
	//	modifyInfo := &RowModifyInfo{
	//		RowId:   row.RowId,
	//		RowName: row.RowName,
	//		Changes: make([]*FieldChange, 0),
	//	}
	//	for _, change := range row.Changes {
	//		modifyInfo.Changes = append(modifyInfo.Changes, &FieldChange{
	//			ColName:  change.ColName,
	//			OldValue: change.OldValue,
	//			NewValue: change.NewValue,
	//		})
	//	}
	//	result.ModifiedRows = append(result.ModifiedRows, modifyInfo)
	//}
	//
	//result.AddedCols = diffResult.AddedCols
	//result.RemovedCols = diffResult.RemovedCols
	//
	//return result, nil
	return nil, errors.New("没实现")
}

// ========== Git 变更文件相关功能 ==========

// GitChangedFileInfo Git 变更文件信息
type GitChangedFileInfo struct {
	RelPath    string `json:"relPath"`    // 相对于仓库根目录的路径
	AbsPath    string `json:"absPath"`    // 绝对路径
	ChangeType string `json:"changeType"` // 变更类型: A(新增), M(修改), D(删除)
	FileName   string `json:"fileName"`   // 文件名
}
