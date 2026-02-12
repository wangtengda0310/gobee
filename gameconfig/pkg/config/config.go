// Package config 提供游戏配置管理的对外 API
//
// 基本使用:
//
//	type Equipment struct {
//	    ID     int    `excel:"id"`
//	    Name   string `excel:"name,required"`
//	    Attack int    `excel:"attack,default:0"`
//	}
//
//	loader := config.NewLoader[Equipment](
//	    "config/装备表.xlsx", "武器",
//	    config.LoadOptions{Mode: config.ModeAuto},
//	)
//	equipments, err := loader.Load()
package config

import (
	"context"
	"time"

	"github.com/wangtengda0310/gobee/gameconfig/internal/config"
)

// Mode 配置加载模式
type Mode = config.Mode

const (
	ModeAuto  = config.ModeAuto
	ModeExcel = config.ModeExcel
	ModeCSV   = config.ModeCSV
)

// LoadOptions 加载选项
type LoadOptions = config.LoadOptions

// Loader 配置加载器（泛型）
type Loader[T any] struct {
	inner *config.Loader[T]
}

// NewLoader 创建配置加载器
func NewLoader[T any](basePath string, sheetName string, options LoadOptions) *Loader[T] {
	return &Loader[T]{
		inner: config.NewLoader[T](basePath, sheetName, config.LoadOptions(options)),
	}
}

// Load 加载配置到结构体切片
func (l *Loader[T]) Load() ([]T, error) {
	return l.inner.Load()
}

// Reload 重新加载
func (l *Loader[T]) Reload() ([]T, error) {
	return l.inner.Reload()
}

// GetVersion 获取配置版本号
func (l *Loader[T]) GetVersion() (int, error) {
	return l.inner.GetVersion()
}

// ConfigWithComments 带批注的配置数据
type ConfigWithComments[T any] struct {
	Data     []T
	Comments map[string]string
}

// LoadWithComments 加载配置并读取批注
func LoadWithComments[T any](basePath string, sheetName string, options LoadOptions) (*ConfigWithComments[T], error) {
	result, err := config.LoadWithComments[T](basePath, sheetName, config.LoadOptions(options))
	if err != nil {
		return nil, err
	}

	return &ConfigWithComments[T]{
		Data:     result.Data,
		Comments: result.Comments,
	}, nil
}

// SchemaManager Schema 管理器（对外）
type SchemaManager = config.SchemaManager

// NewSchemaManager 创建 Schema 管理器
func NewSchemaManager() *SchemaManager {
	return config.NewSchemaManager()
}

// Watcher 文件监听器（对外）
type Watcher[T any] struct {
	inner *config.Watcher[T]
}

// NewWatcher 创建文件监听器
func NewWatcher[T any](loader *Loader[T]) *Watcher[T] {
	return &Watcher[T]{
		inner: config.NewWatcher[T](loader.inner),
	}
}

// OnChange 设置变化回调
func (w *Watcher[T]) OnChange(fn func([]T)) {
	w.inner.OnChange(fn)
}

// SetDebounce 设置防抖时间
func (w *Watcher[T]) SetDebounce(duration time.Duration) {
	w.inner.SetDebounce(duration)
}

// Watch 开始监听文件变化
func (w *Watcher[T]) Watch(ctx context.Context) error {
	return w.inner.Watch(ctx)
}

// Stop 停止监听
func (w *Watcher[T]) Stop() {
	w.inner.Stop()
}

// ExcelExporter Excel 导出器（对外）
type ExcelExporter = config.ExcelExporter

// NewExcelExporter 创建 Excel 导出器
func NewExcelExporter(excelPath, outputDir string) *ExcelExporter {
	return config.NewExcelExporter(excelPath, outputDir)
}

// Export 执行导出
func ExportCSV(e *ExcelExporter) error {
	return (*config.ExcelExporter)(e).Export()
}

// SetSheets 设置要导出的 Sheet 列表
func SetSheets(e *ExcelExporter, sheets []string) {
	(*config.ExcelExporter)(e).SetSheets(sheets)
}

// ConvertToType 类型转换（对外）
func ConvertToType(value string, targetType string) (interface{}, error) {
	return config.ConvertToType(value, targetType)
}
