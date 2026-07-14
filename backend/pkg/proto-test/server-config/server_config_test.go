package serverconfig

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

// createServerConfigXLSX 创建符合名将杀配表规范的临时服务器配置表
// 表头结构：第1行中文名、第2行类型、第3行字段名、第4行导出标识
func createServerConfigXLSX(t *testing.T, filePath string) {
	t.Helper()
	f := excelize.NewFile()
	defer f.Close()

	require.NoError(t, f.SetSheetName("Sheet1", "服务器配置表|Server"))

	_ = f.SetSheetRow("服务器配置表|Server", "A1", &[]any{"服务器ID", "服务器名称", "是否保存", "IP端口", "IP端口HeroPoint", "保持连接", "服务器区ID"})
	_ = f.SetSheetRow("服务器配置表|Server", "A2", &[]any{"int", "string", "int", "string", "string", "bool", "int"})
	_ = f.SetSheetRow("服务器配置表|Server", "A3", &[]any{"Id", "ServelName", "IsSave", "IpPort", "IpPortHeroPoint", "KeepAlive", "ServerZoneId"})
	_ = f.SetSheetRow("服务器配置表|Server", "A4", &[]any{"client", "client", "client", "client", "client", "client", "client"})

	// 一条已有数据（在第5行）
	_ = f.SetSheetRow("服务器配置表|Server", "A5", &[]any{1, "现有服务器", 1, "http://10.0.0.1:20144/authlogin", "http://10.0.0.1:30244", 1, 1})

	require.NoError(t, f.SaveAs(filePath))
}

func createDummyFile(t *testing.T, path string) {
	t.Helper()
	require.NoError(t, os.WriteFile(path, []byte(""), 0644))
}

// TestInjectUnityServer_CreatesNewRowInRow5 验证新数据插入到第5行（数据区首行），已有数据下移
func TestInjectUnityServer_CreatesNewRowInRow5(t *testing.T) {
	tmpDir := t.TempDir()
	excelDir := filepath.Join(tmpDir, "excel")
	require.NoError(t, os.MkdirAll(excelDir, 0755))

	filePath := filepath.Join(excelDir, "服务器配置表.xlsx")
	createServerConfigXLSX(t, filePath)

	origGetter := localIPGetter
	localIPGetter = func() (string, error) { return "192.168.1.100", nil }
	defer func() { localIPGetter = origGetter }()

	cfg := ServerXlsxConfig{
		Id:              999,
		ServerName:      "rain-qa-func",
		IsSave:          1,
		IpPortHeroPoint: "http://localhost:30244",
		KeepAlive:       1,
		ServerZoneId:    42,
		ExcelDir:        tmpDir,
		HTTPListenPort:  20144,
	}

	err := InjectUnityServer(cfg)
	require.NoError(t, err)

	f, err := excelize.OpenFile(filePath)
	require.NoError(t, err)
	defer f.Close()

	rows, err := f.GetRows("服务器配置表|Server")
	require.NoError(t, err)
	require.Len(t, rows, 6, "表头4行+新增数据+原有数据=6行")

	// 新增数据应排在第5行（数据区首行，0-based index 4）
	row := rows[4]
	assert.Equal(t, "999", getCell(row, 0))
	assert.Equal(t, "rain-qa-func", getCell(row, 1))
	assert.Equal(t, "1", getCell(row, 2))
	assert.Equal(t, "http://192.168.1.100:20144/authlogin", getCell(row, 3))
	assert.Equal(t, "http://localhost:30244", getCell(row, 4))
	assert.Equal(t, "1", getCell(row, 5))
	assert.Equal(t, "42", getCell(row, 6))

	// 原有数据被下移到第6行
	oldRow := rows[5]
	assert.Equal(t, "1", getCell(oldRow, 0))
	assert.Equal(t, "现有服务器", getCell(oldRow, 1))
}

// TestInjectUnityServer_UpdatesExistingRow 验证已存在的 Id 行原地更新，不插入新行
func TestInjectUnityServer_UpdatesExistingRow(t *testing.T) {
	tmpDir := t.TempDir()
	excelDir := filepath.Join(tmpDir, "excel")
	require.NoError(t, os.MkdirAll(excelDir, 0755))

	filePath := filepath.Join(excelDir, "服务器配置表.xlsx")
	createServerConfigXLSX(t, filePath)

	origGetter := localIPGetter
	localIPGetter = func() (string, error) { return "10.0.0.99", nil }
	defer func() { localIPGetter = origGetter }()

	// 先注入 Id=999 → 插入到第5行
	cfg1 := ServerXlsxConfig{
		Id:              999,
		ServerName:      "old-name",
		IsSave:          0,
		IpPort:          "http://1.1.1.1:1111/authlogin",
		IpPortHeroPoint: "http://localhost:30244",
		KeepAlive:       0,
		ServerZoneId:    0,
		ExcelDir:        tmpDir,
		HTTPListenPort:  20144,
	}
	require.NoError(t, InjectUnityServer(cfg1))

	// 再次注入 Id=999 → 应更新第5行，不新增
	cfg2 := ServerXlsxConfig{
		Id:              999,
		ServerName:      "rain-qa-func",
		IsSave:          1,
		IpPortHeroPoint: "http://localhost:30244",
		KeepAlive:       1,
		ServerZoneId:    42,
		ExcelDir:        tmpDir,
		HTTPListenPort:  20144,
	}
	require.NoError(t, InjectUnityServer(cfg2))

	f, err := excelize.OpenFile(filePath)
	require.NoError(t, err)
	defer f.Close()

	rows, err := f.GetRows("服务器配置表|Server")
	require.NoError(t, err)
	require.Len(t, rows, 6, "第一次插入第5行+原有第6行=6行，第二次应更新不新增")

	// Id=999 在第5行，已被更新
	row := rows[4]
	assert.Equal(t, "999", getCell(row, 0))
	assert.Equal(t, "rain-qa-func", getCell(row, 1))
	assert.Equal(t, "1", getCell(row, 2))
	assert.Equal(t, "http://10.0.0.99:20144/authlogin", getCell(row, 3))
	assert.Equal(t, "http://localhost:30244", getCell(row, 4))
	assert.Equal(t, "1", getCell(row, 5))
	assert.Equal(t, "42", getCell(row, 6))
}

func TestInjectUnityServer_UsesProvidedIpPort(t *testing.T) {
	tmpDir := t.TempDir()
	excelDir := filepath.Join(tmpDir, "excel")
	require.NoError(t, os.MkdirAll(excelDir, 0755))

	filePath := filepath.Join(excelDir, "服务器配置表.xlsx")
	createServerConfigXLSX(t, filePath)

	cfg := ServerXlsxConfig{
		Id:              999,
		ServerName:      "rain-qa-func",
		IsSave:          1,
		IpPort:          "http://10.20.30.40:8080/authlogin",
		IpPortHeroPoint: "http://localhost:30244",
		KeepAlive:       1,
		ServerZoneId:    42,
		ExcelDir:        tmpDir,
		HTTPListenPort:  20144,
	}

	require.NoError(t, InjectUnityServer(cfg))

	f, err := excelize.OpenFile(filePath)
	require.NoError(t, err)
	defer f.Close()

	rows, err := f.GetRows("服务器配置表|Server")
	require.NoError(t, err)
	// 新数据插入第5行
	row := rows[4]
	assert.Equal(t, "http://10.20.30.40:8080/authlogin", getCell(row, 3))
}

func TestInjectUnityServer_MissingExcelDir(t *testing.T) {
	cfg := ServerXlsxConfig{Id: 999}
	err := InjectUnityServer(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "excel_dir")
}

func TestInjectUnityServer_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := ServerXlsxConfig{
		Id:             999,
		ExcelDir:       tmpDir,
		HTTPListenPort: 20144,
	}
	err := InjectUnityServer(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "服务器配置表.xlsx")
}

func TestInjectUnityServer_LocalIPError(t *testing.T) {
	tmpDir := t.TempDir()
	excelDir := filepath.Join(tmpDir, "excel")
	require.NoError(t, os.MkdirAll(excelDir, 0755))
	createServerConfigXLSX(t, filepath.Join(excelDir, "服务器配置表.xlsx"))

	origGetter := localIPGetter
	localIPGetter = func() (string, error) { return "", assert.AnError }
	defer func() { localIPGetter = origGetter }()

	cfg := ServerXlsxConfig{
		Id:              999,
		ExcelDir:        tmpDir,
		HTTPListenPort:  20144,
		IpPortHeroPoint: "http://localhost:30244",
	}
	err := InjectUnityServer(cfg)
	assert.Error(t, err)
}

// TestExportClientConfig_RunsBatch 验证始终调用 export_client.bat。
// 即使目录下同时存在 export.py，也不应绕过 bat 自行拼接 python 命令。
func TestExportClientConfig_RunsBatch(t *testing.T) {
	tmpDir := t.TempDir()
	createDummyFile(t, filepath.Join(tmpDir, "export_client.bat"))
	// 同时存在 export.py，验证不会被优先走 python 分支绕过 bat
	createDummyFile(t, filepath.Join(tmpDir, "export.py"))

	var gotName string
	var gotArgs []string
	origRunner := commandRunner
	commandRunner = func(name string, arg ...string) *exec.Cmd {
		gotName = name
		gotArgs = arg
		return exec.Command("echo", "mock")
	}
	defer func() { commandRunner = origRunner }()

	err := ExportClientConfig(tmpDir)
	require.NoError(t, err)

	assert.Equal(t, "cmd", gotName)
	assert.Equal(t, []string{"/c", "export_client.bat"}, gotArgs)
	assert.Equal(t, tmpDir, lastCmdDir)
}

// TestExportClientConfig_NoBatch 验证缺少 export_client.bat 时返回明确错误
func TestExportClientConfig_NoBatch(t *testing.T) {
	tmpDir := t.TempDir()

	origRunner := commandRunner
	commandRunner = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("false")
	}
	defer func() { commandRunner = origRunner }()

	err := ExportClientConfig(tmpDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "export_client.bat")
}

// getCell 安全获取切片元素，越界返回空字符串
func getCell(row []string, idx int) string {
	if idx < 0 || idx >= len(row) {
		return ""
	}
	return row[idx]
}
