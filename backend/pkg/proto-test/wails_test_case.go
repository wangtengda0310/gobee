package prototest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	protocol "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/msg"
)

// TestCaseService 测试用例管理服务（对应前端 packet-tab.vue 和 testcase-tab.vue 用例管理）
//
// 时序图：
// ┌─────────────────┐   LoadCaseList    ┌───────────────────┐
// │ testcase-tab.vue│ ─────────────────> │ TestCaseService   │
// └─────────────────┘                    │   (后端)          │
//
//	     │                                  └───────────────────┘
//	     │                                        │
//	     │                                        │ 遍历目录
//	     │                                        ▼
//	     │                                  cases/proto_cases/
//	     │                                        │
//	     │          返回用例列表                   │
//	     │ <──────────────────────────────────────┘
//	     │                                        │
//	     ▼                                        ▼
//	显示用例下拉框                          解析 .json 元信息
//
// ┌─────────────────┐   LoadTestCase    ┌───────────────────┐
// │ testcase-tab.vue│ ─────────────────> │ TestCaseService   │
// └─────────────────┘                    │   (后端)          │
//
//	     │                                  └───────────────────┘
//	     │                                        │
//	     │                                        │ LoadRecordFile()
//	     │                                        ▼
//	     │                                  ┌──────────────────┐
//	     │                                  │RecordFileService │
//	     │                                  └──────────────────┘
//	     │                                        │
//	     │          返回录制数据                   │
//	     │ <──────────────────────────────────────┘
//	     │                                        │
//	     ▼                                        ▼
//	设置到组件状态                          加载录制文件
//
// ┌─────────────────┐   SaveTestCase    ┌───────────────────┐
// │ testcase-tab.vue│ ─────────────────> │ TestCaseService   │
// └─────────────────┘                    │   (后端)          │
//
//	     │                                  └───────────────────┘
//	     │                                        │
//	     │                                        │ SaveRecordFile()
//	     │                                        ▼
//	     │                                  cases/proto_cases/<name>.json
//	     │                                        │
//	     │          保存成功                      │
//	     │ <──────────────────────────────────────┘
//	     │
//	     ▼
//	刷新用例列表
type TestCaseService struct {
	recordFileService *RecordFileService
}

// NewTestCaseService 创建测试用例服务实例
func NewTestCaseService(recordFileService *RecordFileService) *TestCaseService {
	return &TestCaseService{
		recordFileService: recordFileService,
	}
}

// ProtoCaseMeta 测试用例元信息
type ProtoCaseMeta struct {
	Name         string `json:"name"`
	MessageCount int    `json:"message_count"`
	ServerAddr   string `json:"server_addr"`
	CreatedAt    string `json:"created_at"`
}

const protoCaseDir = "cases/proto_cases"

var (
	resolvedCaseDir string
	resolveOnce     sync.Once
)

// getProtoCaseDir 获取用例存储目录。
//
// 解析顺序：
//  1. 若 cwd 下存在 cases/proto_cases/（即当前在 rain-qa-func 项目目录中），返回该相对路径
//  2. 否则 fallback 到 exe 同级 cases 目录（支持 skill 自带用例的独立运行场景）
//
// GUI（wails3 dev）总是从项目根目录启动，命中分支 1，行为不变。
// CLI 从任意目录运行时，若不在项目目录中则自动使用 exe 旁边的自带用例。
func getProtoCaseDir() string {
	resolveOnce.Do(func() {
		// 分支 1：cwd 下有 cases/proto_cases/
		if fi, err := os.Stat(protoCaseDir); err == nil && fi.IsDir() {
			resolvedCaseDir = protoCaseDir
			return
		}
		// 分支 2：exe 同级 cases 目录（exe 在 bin/ 下，cases 在 bin/ 的上级）
		if exePath, err := os.Executable(); err == nil {
			fallback := filepath.Join(filepath.Dir(exePath), "..", "cases")
			if abs, err := filepath.Abs(fallback); err == nil {
				fallback = abs
			}
			if fi, err := os.Stat(fallback); err == nil && fi.IsDir() {
				resolvedCaseDir = fallback
				return
			}
		}
		// 兜底：返回相对路径，让后续报错信息有意义
		resolvedCaseDir = protoCaseDir
	})
	return resolvedCaseDir
}

// ensureCaseDir 确保用例目录存在（仅用于写入场景）
func ensureCaseDir() error {
	return os.MkdirAll(getProtoCaseDir(), 0755)
}

// caseFilePath 用例文件路径
func caseFilePath(name string) string {
	safeName := strings.ReplaceAll(name, " ", "_")
	safeName = strings.ReplaceAll(safeName, "/", "_")
	return filepath.Join(getProtoCaseDir(), safeName+".json")
}

// SaveTestCase 保存测试用例（仅 Req，省略 direction/offset_ms；兼容 LoadRecordFile 和 StartReplay）
// name: 用例名称
// data: 录制文件数据（从中提取消息）
//
// 注意：SaveTestCase 语义为覆盖写入，用于 testcase-tab.vue 的"新增模块"/"删除消息"/"保存顺序"等场景。
// 从 packet-tab.vue 多选 Req "追加"到已有用例时，请使用 AppendTestCase，避免覆盖已有数据。
func (s *TestCaseService) SaveTestCase(name string, data *RecordFileData) error {
	if name == "" {
		return fmt.Errorf("用例名称不能为空")
	}

	path := caseFilePath(name)
	if err := ensureCaseDir(); err != nil {
		return fmt.Errorf("创建用例目录失败: %v", err)
	}

	// 转换为 Recording 后按用例精简格式写入（仅 Req，省略 direction/offset_ms）
	entries := viewsToEntries(data.Messages)
	recordedAt := data.RecordedAt
	if recordedAt == "" {
		recordedAt = time.Now().Format(time.RFC3339)
	}
	rec := &protocol.Recording{
		Version:    protocol.RecordingVersion,
		RecordedAt: recordedAt,
		ServerAddr: data.ServerAddr,
		Messages:   entries,
	}
	return SaveTestCaseToFile(path, rec)
}

// AppendTestCase 向已存在用例追加 Req 消息。
// 若用例不存在则创建新用例；存在则保留原用例元信息，将新 Req 追加到末尾。
// 供 packet-tab.vue "保存到用例" 的追加场景使用，避免覆盖 testcase-tab.vue 已维护的用例数据。
func (s *TestCaseService) AppendTestCase(name string, data *RecordFileData) error {
	if name == "" {
		return fmt.Errorf("用例名称不能为空")
	}

	path := caseFilePath(name)
	if err := ensureCaseDir(); err != nil {
		return fmt.Errorf("创建用例目录失败: %v", err)
	}

	entries := viewsToEntries(data.Messages)
	recordedAt := data.RecordedAt
	if recordedAt == "" {
		recordedAt = time.Now().Format(time.RFC3339)
	}
	rec := &protocol.Recording{
		Version:    protocol.RecordingVersion,
		RecordedAt: recordedAt,
		ServerAddr: data.ServerAddr,
		Messages:   entries,
	}
	return AppendTestCaseToFile(path, rec)
}

// LoadTestCaseList 获取所有测试用例列表
func (s *TestCaseService) LoadTestCaseList() ([]*ProtoCaseMeta, error) {
	entries, err := os.ReadDir(getProtoCaseDir())
	if err != nil {
		return nil, fmt.Errorf("读取用例目录失败: %v", err)
	}

	var cases []*ProtoCaseMeta
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".json")
		path := filepath.Join(getProtoCaseDir(), entry.Name())

		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var meta struct {
			RecordedAt string `json:"recorded_at"`
			ServerAddr string `json:"server_addr"`
			Messages   []any  `json:"messages"`
		}
		if err := json.Unmarshal(data, &meta); err != nil {
			cases = append(cases, &ProtoCaseMeta{
				Name:         name,
				MessageCount: 0,
				CreatedAt:    "",
			})
			continue
		}
		cases = append(cases, &ProtoCaseMeta{
			Name:         name,
			MessageCount: len(meta.Messages),
			ServerAddr:   meta.ServerAddr,
			CreatedAt:    meta.RecordedAt,
		})
	}

	sort.Slice(cases, func(i, j int) bool {
		return cases[i].Name < cases[j].Name
	})

	return cases, nil
}

// LoadTestCase 加载指定测试用例（自动过滤 Ack/Ntf，补全 direction 供重放）
func (s *TestCaseService) LoadTestCase(name string) (*RecordFileData, error) {
	path := caseFilePath(name)
	return s.recordFileService.LoadRecordFile(path)
}

// DeleteTestCase 删除指定测试用例
func (s *TestCaseService) DeleteTestCase(name string) error {
	path := caseFilePath(name)
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("用例不存在: %s", name)
		}
		return fmt.Errorf("删除用例文件失败: %v", err)
	}
	return nil
}
