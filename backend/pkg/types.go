package internal

// FuncCaseConfig 功能测试配置
type FuncCaseConfig struct {
	JsonsDir           string `json:"jsons_dir"`
	ServerAddr         string `json:"server_addr"`
	ServerPort         int    `json:"server_port"`
	Desc               string `json:"desc"`
	RobotPrefix        string `json:"robot_prefix"`
	SingleCaseRunCount int    `json:"single_case_run_count"`
	LoginTime          int    `json:"login_time"`
	RoomOpTime         int    `json:"room_op_time"`
	FeiShuNtf          bool   `json:"fei_shu_ntf"`
	FeiShuGUID         string `json:"fei_shu_guid"`
	DebugLevel         bool   `json:"debug_level"`
	DebugLog           bool   `json:"debug_log"`
	Concurrency        int    `json:"concurrency"`
	AutoSave           bool   `json:"auto_save"`
	InterceptEnabled   bool   `json:"intercept_enabled"`   // 消息劫持开关
	ExcelResourcesDir  string `json:"excel_resources_dir"` // Excel 资源目录（.bytes 文件路径）
}

// ChatConfig 聊天配置
type ChatConfig struct {
	Provider        string          `json:"provider"` // "anthropic" | "openai" | "deepseek"
	AnthropicConfig AnthropicConfig `json:"anthropicConfig"`
	OpenAIConfig    OpenAIConfig    `json:"openaiConfig"`
	DeepSeekConfig  DeepSeekConfig  `json:"deepseekConfig"`
	SystemPrompt    string          `json:"systemPrompt"`
}

// AnthropicConfig Anthropic 配置
type AnthropicConfig struct {
	APIKey  string `json:"apiKey"`
	BaseURL string `json:"baseUrl"`
	Model   string `json:"model"`
}

// OpenAIConfig OpenAI 配置
type OpenAIConfig struct {
	APIKey  string `json:"apiKey"`
	BaseURL string `json:"baseUrl"`
	Model   string `json:"model"`
}

// DeepSeekConfig DeepSeek 配置
type DeepSeekConfig struct {
	APIKey  string `json:"apiKey"`
	BaseURL string `json:"baseUrl"`
	Model   string `json:"model"`
}
