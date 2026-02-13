package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Mode 配置加载模式
type Mode string

const (
	// ModeAuto 自动检测（开发用 Excel，生产用 CSV）
	ModeAuto Mode = "auto"
	// ModeExcel 强制读取 Excel
	ModeExcel Mode = "excel"
	// ModeCSV 强制读取 CSV
	ModeCSV Mode = "csv"
	// ModeMemory 从内存数据加载（用于 Mock 测试）
	ModeMemory Mode = "memory"
)

// LoadOptions 加载选项
type LoadOptions struct {
	// Mode 加载模式
	Mode Mode
	// HeaderRow 表头行索引（默认 0）
	HeaderRow int
	// TypeRow 类型行索引（可选，默认不使用）
	TypeRow int
	// DataStart 数据开始行索引（默认为 HeaderRow + 1）
	DataStart int
	// TagName struct tag 名称（默认 "excel"）
	TagName string
	// MockData Mock 数据源（用于 Memory 模式）
	MockData [][]string
}

// Loader 配置加载器（泛型）
type Loader[T any] struct {
	basePath  string
	sheetName string
	options   LoadOptions
	mapper    *StructMapper[T]
	data      [][]string // 内存数据源（用于 Memory 模式）
	dataMu    sync.RWMutex // 保护 data 字段的读写锁
}

// NewLoader 创建配置加载器
// basePath: 配置文件路径（Excel 文件或 CSV 文件）
// sheetName: Sheet 名称（用于 Excel）或 CSV 文件名（用于 CSV）
func NewLoader[T any](basePath string, sheetName string, options LoadOptions) *Loader[T] {
	// 默认选项
	if options.HeaderRow == 0 && options.Mode == "" {
		options.HeaderRow = 0
	}
	if options.Mode == "" {
		options.Mode = ModeAuto
	}

	return &Loader[T]{
		basePath:  basePath,
		sheetName: sheetName,
		options:   options,
		mapper:    NewStructMapper[T](),
	}
}

// Load 加载配置到结构体切片
func (l *Loader[T]) Load() ([]T, error) {
	// 确定加载模式
	mode := l.options.Mode
	if mode == ModeAuto {
		mode = l.detectMode()
	}

	switch mode {
	case ModeExcel:
		return l.loadFromExcel()
	case ModeCSV:
		return l.loadFromCSV()
	case ModeMemory:
		return l.loadFromMemory()
	default:
		return nil, fmt.Errorf("不支持的加载模式: %s", mode)
	}
}

// Reload 重新加载
func (l *Loader[T]) Reload() ([]T, error) {
	return l.Load()
}

// loadFromMemory 从内存数据加载
// 并发安全：使用读锁保护数据访问
func (l *Loader[T]) loadFromMemory() ([]T, error) {
	// 如果 MockData 不为空，使用 MockData
	if len(l.options.MockData) > 0 {
		return l.parseRows(l.options.MockData)
	}

	// 使用读锁保护内部数据访问
	l.dataMu.RLock()
	defer l.dataMu.RUnlock()
	return l.parseRows(l.data)
}

// detectMode 自动检测加载模式
func (l *Loader[T]) detectMode() Mode {
	// 检查 basePath 是否是 Excel 文件
	if strings.HasSuffix(strings.ToLower(l.basePath), ".xlsx") {
		return ModeExcel
	}

	// 检查是否存在对应的 CSV 文件
	csvPath := l.getCSVPath()
	if _, err := os.Stat(csvPath); err == nil {
		return ModeCSV
	}

	// 默认使用 CSV 模式（如果不是 .xlsx 文件）
	return ModeCSV
}

// getCSVPath 获取 CSV 文件路径
// 格式: {basePath}/{sheetName}.csv
// 如果 basePath 是 .xlsx 文件，则使用同名目录
// 如果 basePath 是 .csv 文件，则直接使用它
func (l *Loader[T]) getCSVPath() string {
	// 如果 basePath 是 Excel 文件
	if strings.HasSuffix(strings.ToLower(l.basePath), ".xlsx") {
		baseDir := filepath.Dir(l.basePath)
		excelName := strings.TrimSuffix(filepath.Base(l.basePath), ".xlsx")
		return filepath.Join(baseDir, excelName, l.sheetName+".csv")
	}

	// 如果 basePath 已经是 CSV 文件，直接返回
	if strings.HasSuffix(strings.ToLower(l.basePath), ".csv") {
		return l.basePath
	}

	// 否则假设 basePath 是目录
	return filepath.Join(l.basePath, l.sheetName+".csv")
}

// loadFromExcel 从 Excel 加载
func (l *Loader[T]) loadFromExcel() ([]T, error) {
	reader, err := NewExcelReader(l.basePath)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	// 读取 Sheet 数据
	rows, err := reader.ReadSheet(l.sheetName)
	if err != nil {
		return nil, err
	}

	return l.parseRows(rows)
}

// loadFromCSV 从 CSV 加载
func (l *Loader[T]) loadFromCSV() ([]T, error) {
	csvPath := l.getCSVPath()

	reader := NewCSVReader(csvPath)
	defer reader.Close()

	// 读取 CSV 数据
	rows, err := reader.Read()
	if err != nil {
		return nil, err
	}

	return l.parseRows(rows)
}

// parseRows 解析行数据
func (l *Loader[T]) parseRows(rows [][]string) ([]T, error) {
	if len(rows) == 0 {
		return nil, nil
	}

	// 确定表头行和数据开始行
	headerRow := l.options.HeaderRow
	dataStart := l.options.DataStart
	if dataStart == 0 {
		dataStart = headerRow + 1
	}

	// 检查是否有版本行
	if len(rows) > 0 && len(rows[0]) > 0 && rows[0][0] == "__version__" {
		// 跳过版本行
		headerRow = 1
		dataStart = 2
		// 检查是否有变更说明行
		if len(rows) > 1 && len(rows[1]) > 0 && rows[1][0] == "__changes__" {
			headerRow = 2
			dataStart = 3
		}
	}

	// 验证行数
	if len(rows) <= headerRow {
		return nil, fmt.Errorf("数据行数不足")
	}

	// 获取表头
	headers := rows[headerRow]

	// 获取数据行
	dataRows := rows[dataStart:]

	// 使用映射器映射数据
	return l.mapper.MapRows(headers, dataRows)
}

// GetVersion 获取配置版本号
func (l *Loader[T]) GetVersion() (int, error) {
	mode := l.options.Mode
	if mode == ModeAuto {
		mode = l.detectMode()
	}

	if mode == ModeExcel {
		reader, err := NewExcelReader(l.basePath)
		if err != nil {
			return 0, err
		}
		defer reader.Close()
		return reader.GetVersion(l.sheetName)
	}

	// CSV 模式下不支持版本号
	return 0, nil
}

// SetSchemaManager 设置 Schema 管理器（暂未实现）
func (l *Loader[T]) SetSchemaManager(sm interface{}) {
	// TODO: 实现 Schema 迁移
	_ = sm
}

// ConfigWithComments 带批注的配置数据
type ConfigWithComments[T any] struct {
	Data     []T
	Comments map[string]string // "字段名" -> "批注内容"
}

// LoadWithComments 加载配置并读取批注
func LoadWithComments[T any](basePath string, sheetName string, options LoadOptions) (*ConfigWithComments[T], error) {
	mode := options.Mode
	if mode == ModeAuto {
		// 检测模式
		if strings.HasSuffix(strings.ToLower(basePath), ".xlsx") {
			mode = ModeExcel
		} else {
			mode = ModeExcel // 默认使用 Excel
		}
	}

	if mode != ModeExcel {
		return nil, fmt.Errorf("批注功能仅支持 Excel 模式")
	}

	reader, err := NewExcelReader(basePath)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	// 读取批注
	comments := readComments(reader, sheetName)

	// 加载数据
	loader := NewLoader[T](basePath, sheetName, options)
	data, err := loader.Load()
	if err != nil {
		return nil, err
	}

	return &ConfigWithComments[T]{
		Data:     data,
		Comments: comments,
	}, nil
}

// readComments 读取 Excel 批注
func readComments(reader *ExcelReader, sheetName string) map[string]string {
	// TODO: 实现批注读取
	// 目前返回空 map，后续在 comment.go 中实现
	return make(map[string]string)
}

// SetMockData 设置 Mock 数据（用于测试）
// 将数据存储在内部，当 Mode 为 ModeMemory 时使用
// 并发安全：使用写锁保护
func (l *Loader[T]) SetMockData(data [][]string) {
	l.dataMu.Lock()
	defer l.dataMu.Unlock()
	l.data = data
}

// GetMockData 获取当前 Mock 数据（用于测试）
// 并发安全：使用读锁保护
func (l *Loader[T]) GetMockData() [][]string {
	l.dataMu.RLock()
	defer l.dataMu.RUnlock()
	return l.data
}
