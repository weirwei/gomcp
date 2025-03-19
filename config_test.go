package gomcp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// 创建临时配置文件
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.json")

	// 创建测试配置
	testConfig := Config{
		MCPServers: map[string]ServerConfig{
			"test-server": {
				Command: "echo",
				Args:    []string{"test"},
				Env: map[string]string{
					"TEST_VAR": "test_value",
				},
			},
		},
	}

	// 将配置写入文件
	data, err := json.MarshalIndent(testConfig, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// 加载配置
	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// 验证配置
	serverConfig, exists := config.MCPServers["test-server"]
	if !exists {
		t.Fatal("Server config not found")
	}
	if serverConfig.Command != "echo" {
		t.Errorf("Unexpected command: %v", serverConfig.Command)
	}
	if len(serverConfig.Args) != 1 || serverConfig.Args[0] != "test" {
		t.Errorf("Unexpected args: %v", serverConfig.Args)
	}
	if serverConfig.Env["TEST_VAR"] != "test_value" {
		t.Errorf("Unexpected env: %v", serverConfig.Env)
	}
}

func TestGetServerConfig(t *testing.T) {
	config := &Config{
		MCPServers: map[string]ServerConfig{
			"test-server": {
				Command: "echo",
				Args:    []string{"test"},
			},
			"disabled-server": {
				Command:  "echo",
				Args:     []string{"test"},
				Disabled: true,
			},
		},
	}

	// 测试正常服务器
	serverConfig, err := config.GetServerConfig("test-server")
	if err != nil {
		t.Fatalf("Failed to get server config: %v", err)
	}
	if serverConfig.Command != "echo" {
		t.Errorf("Unexpected command: %v", serverConfig.Command)
	}

	// 测试不存在的服务器
	_, err = config.GetServerConfig("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent server")
	}

	// 测试禁用的服务器
	_, err = config.GetServerConfig("disabled-server")
	if err == nil {
		t.Error("Expected error for disabled server")
	}
}

func TestStartServer(t *testing.T) {
	config := &Config{
		MCPServers: map[string]ServerConfig{
			"test-server": {
				Command: "echo",
				Args:    []string{"test"},
				Env: map[string]string{
					"TEST_VAR": "test_value",
				},
			},
		},
	}

	// 启动服务器
	cmd, err := config.BuildServer("test-server")
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// 等待服务器进程结束
	if err := cmd.Wait(); err != nil {
		t.Fatalf("Server process failed: %v", err)
	}
}
