package testdb

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadMySQLConfig 加载MySQL配置
func LoadMySQLConfig(configPath string) (MySQLConfig, error) {
	var config MySQLConfig

	// 查找配置文件
	absPath, err := findConfigFile(configPath)
	if err != nil {
		return config, fmt.Errorf("查找配置文件失败: %v", err)
	}

	// 读取文件
	data, err := ioutil.ReadFile(absPath)
	if err != nil {
		return config, fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 解析YAML
	if err := yaml.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("解析YAML配置失败: %v", err)
	}

	// 设置默认值
	if config.Server.Host == "" {
		config.Server.Host = "127.0.0.1"
	}

	if config.Server.User == "" {
		config.Server.User = "root"
	}

	if config.Server.Port == 0 {
		config.Server.Port = 3306
	}

	return config, nil
}


// findConfigFile 查找配置文件
func findConfigFile(configPath string) (string, error) {
	// 如果是绝对路径，直接使用
	if filepath.IsAbs(configPath) {
		if fileExists(configPath) {
			return configPath, nil
		}
		return "", fmt.Errorf("配置文件不存在: %s", configPath)
	}

	// 尝试相对路径查找
	searchPaths := []string{
		configPath,
		filepath.Join("configs", configPath),
		filepath.Join("pkg", "testdb", configPath),
		filepath.Join("pkg", "testdb", "configs", configPath),
		filepath.Join("lvan", "pkg", "testdb", "configs", configPath),
	}

	for _, path := range searchPaths {
		if fileExists(path) {
			return path, nil
		}
	}

	return "", fmt.Errorf("配置文件未找到: %s", configPath)
}

// fileExists 检查文件是否存在
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// SaveMySQLConfig 保存MySQL配置
func SaveMySQLConfig(config MySQLConfig, configPath string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	if err := ioutil.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	return nil
}

