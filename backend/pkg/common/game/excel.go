package game

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-robot/project/xcard/xcard_excel/excel"
	"git.devcloud.ztgame.com/v-tangfangda/rain-robot/project/xcard/xcard_excel/excel_config"
	"git.devcloud.ztgame.com/v-tangfangda/rain-robot/project/xcard/xcard_excel/excel_data"
	"git.devcloud.ztgame.com/v-tangfangda/rain-robot/project/xcard/xcard_excel/excel_manager_impl"
	protoMsg "git.devcloud.ztgame.com/v-tangfangda/rain-robot/project/xcard/xcard_pb"

	mcpgo "github.com/modelcontextprotocol/go-sdk/mcp"
)

// GameExcelService 游戏Excel配置服务
type GameExcelService struct {
	initCalled bool // 标记 excel_manager_impl.Init 是否已被调用过（once 不可逆）
}

// NewGameExcelService 创建游戏Excel服务实例
func NewGameExcelService() *GameExcelService {
	return &GameExcelService{}
}

// resolveExcelPath 尝试解析Excel资源路径
// 如果传入的相对路径不存在，尝试从可执行文件所在目录解析
func resolveExcelPath(inputPath string) (string, error) {
	// 如果是绝对路径，直接使用
	if filepath.IsAbs(inputPath) {
		if info, err := os.Stat(inputPath); err == nil && info.IsDir() {
			return inputPath, nil
		}
		return "", fmt.Errorf("绝对路径不存在或不是目录: %s", inputPath)
	}

	// 尝试从当前工作目录解析
	if info, err := os.Stat(inputPath); err == nil && info.IsDir() {
		return inputPath, nil
	}

	// 尝试从可执行文件所在目录解析
	execPath, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(execPath)
		candidate := filepath.Join(execDir, inputPath)
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}

		// 如果路径包含 "../"，尝试从可执行文件目录向上追溯
		if strings.Contains(inputPath, "..") {
			// 尝试从项目根目录（rain-qa-func的父目录）解析
			// 可执行文件通常在 rain-qa-func/ 或 rain-qa-func/build/ 下，
			// 也可能在 git worktree 中（rain-qa-func/.claude/worktrees/<name>/），
			// 此时需要向上追溯 4 级才能到达包含 rain-robot 的 xcard-qa-tools 根目录
			projectRoot := execDir
			for i := 0; i < 5; i++ {
				parent := filepath.Dir(projectRoot)
				if parent == projectRoot {
					break
				}
				projectRoot = parent
				candidate := filepath.Join(projectRoot, "rain-robot", "project", "xcard", "xcard_excel", "resources")
				if info, err := os.Stat(candidate); err == nil && info.IsDir() {
					return candidate, nil
				}
			}
		}
	}

	// 最后尝试从当前工作目录的父目录查找（针对wails3 dev模式）
	cwd, err := os.Getwd()
	if err == nil {
		candidate := filepath.Join(cwd, inputPath)
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("无法找到Excel资源路径: %s", inputPath)
}

// InitExcel 使用指定路径初始化 Excel 数据
// 首次调用使用 Init（同步），后续调用使用 Reload（异步）
// Reload 使用 rain-robot 硬编码的 resPath，不支持自定义路径
func (e *GameExcelService) InitExcel(path string) error {
	log.Printf("[GameExcelService.InitExcel] 被调用, path=%s, initCalled=%v", path, e.initCalled)

	// 解析路径（处理相对路径）
	resolvedPath, err := resolveExcelPath(path)
	if err != nil {
		return err
	}
	log.Printf("[GameExcelService.InitExcel] 解析后的路径: %s", resolvedPath)

	if e.initCalled {
		log.Println("[GameExcelService.InitExcel] initCalled=true, 调用 Reload()")
		excel_manager_impl.Reload()
		if !e.IsInitialized() {
			return fmt.Errorf("Reload 后 Excel 数据未成功初始化")
		}
		log.Printf("[GameExcelService.InitExcel] Reload() 完成, IsInitialized=%v", e.IsInitialized())
		return nil
	}
	e.initCalled = true
	log.Printf("[GameExcelService.InitExcel] 调用 excel_manager_impl.Init, path=%s", resolvedPath)
	excel_manager_impl.Init(&resolvedPath)
	if !e.IsInitialized() {
		return fmt.Errorf("Init 后 Excel 数据未成功初始化, path=%s", resolvedPath)
	}
	log.Printf("[GameExcelService.InitExcel] Init 完成, IsInitialized=%v", e.IsInitialized())
	return nil
}

// IsInitialized 直接检查底层数据是否已加载（不依赖实例级标记）
// 这样所有 GameExcelService 实例都能正确感知初始化状态
func (e *GameExcelService) IsInitialized() bool {
	// 使用 defer recover 防止 excel_manager_impl 内部 nil 指针 panic
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[GameExcelService.IsInitialized] recover from panic: %v", r)
		}
	}()
	result := excel_manager_impl.GetExcel("SkillsTemplate") != nil
	log.Printf("[GameExcelService.IsInitialized] 检查结果=%v", result)
	return result
}

// GetAllHeroCfg 获取所有英雄配置
// @frontend @mcp
func (e *GameExcelService) GetAllHeroCfg() map[excel.EHeroId]*excel_config.HeroConfig {
	log.Printf("[GameExcelService.GetAllHeroCfg] 被调用, IsInitialized=%v", e.IsInitialized())
	if !e.IsInitialized() {
		log.Println("[GameExcelService.GetAllHeroCfg] Excel未初始化, 返回空map")
		return map[excel.EHeroId]*excel_config.HeroConfig{}
	}
	result := excel_data.GetAllHeroCfg()
	log.Printf("[GameExcelService.GetAllHeroCfg] 返回英雄数量=%d", len(result))
	return result
}

// GetAllCardCfg 获取所有卡牌配置
// @frontend @mcp
func (e *GameExcelService) GetAllCardCfg() map[int32]*excel.CardsTemplate {
	log.Printf("[GameExcelService.GetAllCardCfg] 被调用, IsInitialized=%v", e.IsInitialized())
	if !e.IsInitialized() {
		log.Println("[GameExcelService.GetAllCardCfg] Excel未初始化, 返回空map")
		return map[int32]*excel.CardsTemplate{}
	}
	result := excel_data.GetAllCardCfg()
	log.Printf("[GameExcelService.GetAllCardCfg] 返回卡牌数量=%d", len(result))
	return result
}

// GetAllSkillCfg 获取所有技能配置
// @frontend @mcp
func (e *GameExcelService) GetAllSkillCfg() map[excel.ESkillId]*excel.SkillsTemplate {
	log.Printf("[GameExcelService.GetAllSkillCfg] 被调用, IsInitialized=%v", e.IsInitialized())
	if !e.IsInitialized() {
		log.Println("[GameExcelService.GetAllSkillCfg] Excel未初始化, 返回空map")
		return map[excel.ESkillId]*excel.SkillsTemplate{}
	}
	result := excel_data.GetAllSkillCfg()
	log.Printf("[GameExcelService.GetAllSkillCfg] 返回技能数量=%d", len(result))
	return result
}

// GetEGameMsgIdNameMap 获取游戏消息ID名称映射
// @mcp
func (e *GameExcelService) GetEGameMsgIdNameMap() map[int32]string {
	return protoMsg.EGameMsgID_name
}

// GetErrorCodeMap 获取错误代码映射
// @frontend @mcp
func (e *GameExcelService) GetErrorCodeMap() map[int32]string {
	return excel.EErrorCode_name
}

// GetPropertyTypeMap 获取属性类型映射
// @frontend @mcp
func (e *GameExcelService) GetPropertyTypeMap() (res map[int32]string) {
	res = make(map[int32]string)
	for i, s := range protoMsg.PropertyType_name {
		if ns, ok := strings.CutPrefix(s, "Pro_"); ok {
			res[i] = ns
		} else {
			res[i] = s
		}
	}
	return res
}

// GetOptActionTypeMap 获取操作动作类型映射
// @frontend
func (e *GameExcelService) GetOptActionTypeMap() map[int32]string {
	return protoMsg.OptActionType_name
}

// ========== MCP 工具注册 ==========

// RegisterMCPTools 注册游戏数据相关的 MCP Tools
func RegisterMCPTools(s *mcpgo.Server, svc *GameExcelService) {

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
			&mcpgo.TextContent{Text: errMsg},
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
