package gomcp

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Config 表示 MCP 配置文件的结构
type Config struct {
	MCPServers map[string]ServerConfig `json:"mcpServers"`
}

// ServerConfig 表示单个 MCP 服务器的配置
type ServerConfig struct {
	Command     string            `json:"command"`
	Args        []string          `json:"args"`
	Env         map[string]string `json:"env,omitempty"`
	Disabled    bool              `json:"disabled,omitempty"`
	AutoApprove []string          `json:"autoApprove,omitempty"`
}

// LoadConfig 从文件加载 MCP 配置
func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// GetServerConfig 获取指定名称的服务器配置
func (c *Config) GetServerConfig(name string) (*ServerConfig, error) {
	config, exists := c.MCPServers[name]
	if !exists {
		return nil, fmt.Errorf("server config not found: %s", name)
	}
	if config.Disabled {
		return nil, fmt.Errorf("server is disabled: %s", name)
	}
	return &config, nil
}

// BuildServer 构建指定的 MCP 服务器
func (c *Config) BuildServer(name string) (*exec.Cmd, error) {
	config, err := c.GetServerConfig(name)
	if err != nil {
		return nil, err
	}

	// 使用更直接的方式连接标准输入输出
	cmd := exec.Command(config.Command, config.Args...)

	// 设置环境变量
	if len(config.Env) > 0 {
		cmd.Env = os.Environ()
		for k, v := range config.Env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	return cmd, nil
}

// GetDefaultConfigPath 获取默认配置文件路径
func GetDefaultConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ".mcp-config.json"
	}
	return filepath.Join(homeDir, ".mcp-config.json")
}
