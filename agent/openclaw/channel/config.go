package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Gateway GatewayConfig `json:"gateway" yaml:"gateway"`
	API     APIConfig     `json:"api" yaml:"api"`
	Log     LogConfig     `json:"log" yaml:"log"`
}

// GatewayConfig represents gateway connection configuration
type GatewayConfig struct {
	URL             string        `json:"url" yaml:"url"`
	Token           string        `json:"token,omitempty" yaml:"token,omitempty"`
	ReconnectDelay  time.Duration `json:"reconnectDelay" yaml:"reconnectDelay"`
	MaxReconnect    int           `json:"maxReconnect" yaml:"maxReconnect"`
	PingInterval    time.Duration `json:"pingInterval" yaml:"pingInterval"`
	Timeout         time.Duration `json:"timeout" yaml:"timeout"`
	EnableHeartbeat bool          `json:"enableHeartbeat" yaml:"enableHeartbeat"`
}

// APIConfig represents API server configuration
type APIConfig struct {
	Host            string        `json:"host" yaml:"host"`
	Port            int           `json:"port" yaml:"port"`
	ExternalAddress string        `json:"externalAddress" yaml:"externalAddress"` // For reverse mode: the address connectors use to reach this server
	ReadTimeout     time.Duration `json:"readTimeout" yaml:"readTimeout"`
	WriteTimeout    time.Duration `json:"writeTimeout" yaml:"writeTimeout"`
	ShutdownTimeout time.Duration `json:"shutdownTimeout" yaml:"shutdownTimeout"`
}

// LogConfig represents logging configuration
type LogConfig struct {
	Level  string `json:"level" yaml:"level"`
	Format string `json:"format" yaml:"format"`
	Output string `json:"output" yaml:"output"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Gateway: GatewayConfig{
			URL:             "ws://127.0.0.1:18789",
			Token:           "",
			ReconnectDelay:  DefaultReconnectDelay,
			MaxReconnect:    10,
			PingInterval:    DefaultPingInterval,
			Timeout:         DefaultTimeout,
			EnableHeartbeat: true,
		},
		API: APIConfig{
			Host:            "0.0.0.0",
			Port:            8080,
			ReadTimeout:     DefaultReadTimeout,
			WriteTimeout:    DefaultWriteTimeout,
			ShutdownTimeout: DefaultShutdownTimeout,
		},
		Log: LogConfig{
			Level:  "info",
			Format: "text",
			Output: "stdout",
		},
	}
}

// ConfigManager manages configuration loading and caching
type ConfigManager struct {
	mu     sync.RWMutex
	config *Config
	path   string
}

var (
	configManager     *ConfigManager
	configManagerOnce sync.Once
)

// GetConfigManager returns the singleton config manager
func GetConfigManager() *ConfigManager {
	configManagerOnce.Do(func() {
		configManager = &ConfigManager{
			config: DefaultConfig(),
		}
	})
	return configManager
}

// LoadFromFile loads configuration from a file
func (cm *ConfigManager) LoadFromFile(path string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse based on extension
	config := DefaultConfig()
	ext := filepath.Ext(path)
	switch ext {
	case ".json":
		if err := json.Unmarshal(data, config); err != nil {
			return fmt.Errorf("failed to parse JSON config: %w", err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, config); err != nil {
			return fmt.Errorf("failed to parse YAML config: %w", err)
		}
	default:
		return fmt.Errorf("unsupported config file format: %s", ext)
	}

	cm.config = config
	cm.path = path
	return nil
}

// LoadFromEnv loads configuration from environment variables
func (cm *ConfigManager) LoadFromEnv() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if url := os.Getenv("OPENCLAW_GATEWAY_URL"); url != "" {
		cm.config.Gateway.URL = url
	}
	if token := os.Getenv("OPENCLAW_GATEWAY_TOKEN"); token != "" {
		cm.config.Gateway.Token = token
	}
	if host := os.Getenv("OPENCLAW_API_HOST"); host != "" {
		cm.config.API.Host = host
	}
	if port := os.Getenv("OPENCLAW_API_PORT"); port != "" {
		var portInt int
		if _, err := fmt.Sscanf(port, "%d", &portInt); err == nil {
			cm.config.API.Port = portInt
		}
	}
	if remoteURL := os.Getenv("OPENCLAW_REMOTE_URL"); remoteURL != "" {
		cm.config.API.ExternalAddress = remoteURL
	}
	if level := os.Getenv("OPENCLAW_LOG_LEVEL"); level != "" {
		cm.config.Log.Level = level
	}

	return nil
}

// Get returns the current configuration
func (cm *ConfigManager) Get() *Config {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.config
}

// Set updates the configuration
func (cm *ConfigManager) Set(config *Config) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.config = config
}

// Save saves the configuration to file
func (cm *ConfigManager) Save() error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if cm.path == "" {
		return fmt.Errorf("no config file path set")
	}

	var data []byte
	var err error

	ext := filepath.Ext(cm.path)
	switch ext {
	case ".json":
		data, err = json.MarshalIndent(cm.config, "", "  ")
	case ".yaml", ".yml":
		data, err = yaml.Marshal(cm.config)
	default:
		return fmt.Errorf("unsupported config file format: %s", ext)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(cm.path, data, 0644)
}
