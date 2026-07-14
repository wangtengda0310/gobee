// Package serverconfig 提供 Server Config 页面所需的后端能力，包括向策划配表注入
// Unity 服务器列表条目并触发客户端配置导出。
package serverconfig

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

// ErrExcelDirRequired excel_dir 未提供且无法从统一配置读取时返回
var ErrExcelDirRequired = errors.New("excel_dir 不能为空")

const (
	serverConfigFileName = "服务器配置表.xlsx"
	serverConfigSheet    = "服务器配置表|Server"
	serverConfigSubDir   = "excel"
)

// ServerXlsxConfig Unity 服务器列表行配置
// 字段顺序与服务器配置表.xlsx 第3行字段名保持一致
// @frontend @mcp
type ServerXlsxConfig struct {
	Id              int    `json:"id"`
	ServerName      string `json:"server_name"`
	IsSave          int    `json:"is_save"`
	IpPort          string `json:"ip_port"`
	IpPortHeroPoint string `json:"ip_port_hero_point"`
	KeepAlive       int    `json:"keep_alive"`
	ServerZoneId    int    `json:"server_zone_id"`
	ExcelDir        string `json:"excel_dir"`        // 策划配表目录（其下包含 excel/ 子目录）
	HTTPListenPort  int    `json:"http_listen_port"` // 抽屉面板配置的 HTTP 监听端口号
}

// 可替换的依赖，便于单元测试
var (
	localIPGetter = defaultLocalIPGetter
	commandRunner = defaultCommandRunner
	lastCmdDir    string // 仅用于测试验证最后一次命令的工作目录
)

// defaultLocalIPGetter 获取本机"对外通信"的 IPv4 地址。
// 通过 UDP Dial 到公网地址确定出口网卡，解决多网卡选 IP 的问题。
func defaultLocalIPGetter() (string, error) {
	// UDP Dial 不会真正发送数据，只是确定路由
	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		// 回退到 InterfaceAddrs
		return fallbackLocalIP()
	}
	defer conn.Close()
	addr := conn.LocalAddr().(*net.UDPAddr)
	return addr.IP.String(), nil
}

// fallbackLocalIP 遍历所有网卡获取第一个非回环 IPv4 地址
func fallbackLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", fmt.Errorf("获取本机IP失败: %w", err)
	}
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok || ipNet.IP.IsLoopback() || ipNet.IP.To4() == nil {
			continue
		}
		return ipNet.IP.String(), nil
	}
	return "", fmt.Errorf("未找到可用的本机IPv4地址")
}

// defaultCommandRunner 创建外部命令
func defaultCommandRunner(name string, arg ...string) *exec.Cmd {
	return exec.Command(name, arg...)
}

// InjectUnityServer 向策划配表的服务器配置表.xlsx 中写入或更新一行 Unity 服务器配置。
// 新增时插入到第5行（数据区首行），排在已有数据最前面。
// 当 cfg.IpPort 为空时，自动使用本机IP和 cfg.HTTPListenPort 构造：
// http://{本机ip}:{http_listen_port}/authlogin
// @frontend @mcp
func InjectUnityServer(cfg ServerXlsxConfig) error {
	if cfg.ExcelDir == "" {
		return ErrExcelDirRequired
	}

	filePath := filepath.Join(cfg.ExcelDir, serverConfigSubDir, serverConfigFileName)
	if _, err := os.Stat(filePath); err != nil {
		return fmt.Errorf("服务器配置表不存在: %s, err: %w", filePath, err)
	}

	if cfg.IpPort == "" {
		localIP, err := localIPGetter()
		if err != nil {
			return fmt.Errorf("构造 IpPort 失败: %w", err)
		}
		cfg.IpPort = fmt.Sprintf("http://%s:%d/authlogin", localIP, cfg.HTTPListenPort)
	}

	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return fmt.Errorf("打开服务器配置表失败: %w", err)
	}
	defer f.Close()

	rows, sheetName, err := getServerConfigRows(f)
	if err != nil {
		return err
	}

	if len(rows) < 4 {
		return fmt.Errorf("服务器配置表表头行数不足，期望至少4行")
	}

	colMap := buildColumnMap(rows[2]) // 第3行为字段名
	required := []string{"Id", "ServelName", "IsSave", "IpPort", "IpPortHeroPoint", "KeepAlive", "ServerZoneId"}
	for _, name := range required {
		if _, ok := colMap[name]; !ok {
			return fmt.Errorf("服务器配置表缺少必要字段列: %s", name)
		}
	}

	values := map[string]any{
		"Id":              cfg.Id,
		"ServelName":      cfg.ServerName,
		"IsSave":          cfg.IsSave,
		"IpPort":          cfg.IpPort,
		"IpPortHeroPoint": cfg.IpPortHeroPoint,
		"KeepAlive":       cfg.KeepAlive,
		"ServerZoneId":    cfg.ServerZoneId,
	}

	existingRow := findRowByID(rows, colMap["Id"], cfg.Id)
	if existingRow > 0 {
		// 已存在：原地更新
		for name, col := range colMap {
			cell, err := excelize.CoordinatesToCellName(col+1, existingRow)
			if err != nil {
				return fmt.Errorf("计算单元格坐标失败: %w", err)
			}
			if err := f.SetCellValue(sheetName, cell, values[name]); err != nil {
				return fmt.Errorf("写入单元格 %s 失败: %w", cell, err)
			}
		}
	} else {
		// 不存在：在第5行（数据区首行）插入空行，然后写入
		if err := f.InsertRows(sheetName, 5, 1); err != nil {
			return fmt.Errorf("在第5行插入空行失败: %w", err)
		}
		for name, col := range colMap {
			cell, err := excelize.CoordinatesToCellName(col+1, 5)
			if err != nil {
				return fmt.Errorf("计算单元格坐标失败: %w", err)
			}
			if err := f.SetCellValue(sheetName, cell, values[name]); err != nil {
				return fmt.Errorf("写入单元格 %s 失败: %w", cell, err)
			}
		}
	}

	if err := f.Save(); err != nil {
		return fmt.Errorf("保存服务器配置表失败: %w", err)
	}
	return nil
}

// getServerConfigRows 读取服务器配置表的所有行，支持通过后缀匹配 Sheet 名
func getServerConfigRows(f *excelize.File) ([][]string, string, error) {
	rows, err := f.GetRows(serverConfigSheet)
	if err == nil {
		return rows, serverConfigSheet, nil
	}

	sheets := f.GetSheetList()
	suffix := serverConfigSheet
	if idx := strings.LastIndex(serverConfigSheet, "|"); idx >= 0 {
		suffix = serverConfigSheet[idx+1:]
	}
	for _, sheet := range sheets {
		if strings.HasSuffix(sheet, "|"+suffix) || sheet == suffix {
			rows, err = f.GetRows(sheet)
			if err == nil {
				return rows, sheet, nil
			}
		}
	}
	return nil, "", fmt.Errorf("读取服务器配置表 Sheet 失败: %w", err)
}

// buildColumnMap 根据字段名行建立字段名->列索引的映射
func buildColumnMap(fieldRow []string) map[string]int {
	m := make(map[string]int)
	for i, name := range fieldRow {
		if name == "" {
			continue
		}
		m[name] = i
	}
	return m
}

// findRowByID 在现有数据行中查找指定 Id 所在的行号（excelize 1-based）
func findRowByID(rows [][]string, idCol, targetID int) int {
	for i := 4; i < len(rows); i++ { // 数据从第5行开始（0-based index 4）
		row := rows[i]
		if idCol >= len(row) {
			continue
		}
		val := strings.TrimSpace(row[idCol])
		if val == "" {
			continue
		}
		id, err := strconv.Atoi(val)
		if err == nil && id == targetID {
			return i + 1 // 转换为 1-based 行号
		}
	}
	return -1
}

// ExportClientConfig 在策划配表目录下执行客户端配置导出批处理 export_client.bat。
// bat 内部通过 start 在独立控制台窗口运行 export.py，导出日志显示在该窗口。
// 因 export.py 末尾存在 os.system("pause")，导出完成后窗口需用户手动按键关闭：
// start 创建的新控制台不继承父进程 stdin 重定向，无法从外部喂入按键绕过 pause，
// 此为 Windows 控制台机制限制，故接受手动关闭而非尝试自动关闭。
// @frontend @mcp
func ExportClientConfig(excelDir string) error {
	if excelDir == "" {
		return fmt.Errorf("excel_dir 不能为空")
	}

	batPath := filepath.Join(excelDir, "export_client.bat")
	if _, err := os.Stat(batPath); err != nil {
		return fmt.Errorf("未找到客户端导出脚本: 在 %s 下未找到 export_client.bat", excelDir)
	}

	lastCmdDir = excelDir
	cmd := commandRunner("cmd", "/c", "export_client.bat")
	cmd.Dir = excelDir
	return cmd.Run()
}
