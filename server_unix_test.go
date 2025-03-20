package gomcp

import (
	"errors"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	jsoniter "github.com/json-iterator/go"
)

// 测试创建新的UnixServer
func TestNewUnixServer(t *testing.T) {
	socketPath := "/tmp/test_unix_server.sock"
	server := &UnixServer{
		socketPath: socketPath,
		handlers:   make(map[string]RequestHandler),
		done:       make(chan struct{}),
	}

	if server.socketPath != socketPath {
		t.Errorf("socketPath不正确: 期望 %s, 得到 %s", socketPath, server.socketPath)
	}

	if server.handlers == nil {
		t.Error("handlers不应该为nil")
	}

	if server.done == nil {
		t.Error("done通道不应该为nil")
	}
}

// 测试注册处理器
func TestUnixServer_RegisterHandler(t *testing.T) {
	server := &UnixServer{
		socketPath: "/tmp/test_unix_server.sock",
		handlers:   make(map[string]RequestHandler),
		done:       make(chan struct{}),
	}

	// 注册一个处理器
	testHandler := func(params map[string]interface{}) (interface{}, error) {
		return "test_result", nil
	}

	server.RegisterHandler("test_method", testHandler)
	// 验证处理器是否已注册
	server.mu.RLock()
	handler, exists := server.handlers["test_method"]
	server.mu.RUnlock()

	if !exists {
		t.Error("处理器应该已被注册")
	}

	// 测试处理器是否能正常工作
	result, err := handler(map[string]interface{}{})
	if err != nil {
		t.Errorf("处理器执行出错: %v", err)
	}

	if result != "test_result" {
		t.Errorf("处理器返回了错误的结果: 期望 'test_result', 得到 %v", result)
	}
}

// 测试处理请求 - 方法不存在的情况
func TestUnixServer_HandleRequest_MethodNotFound(t *testing.T) {
	server := &UnixServer{
		socketPath: "/tmp/test_unix_server.sock",
		handlers:   make(map[string]RequestHandler),
		done:       make(chan struct{}),
	}

	// 创建一个请求，使用不存在的方法
	request := Request{
		JsonRPC: "2.0",
		Method:  "non_existent_method",
		Params:  map[string]interface{}{},
		ID:      1,
	}

	// 处理请求
	response := server.handleRequest(request)

	// 验证响应
	if response.JsonRPC != "2.0" {
		t.Errorf("响应中的jsonrpc字段错误: 期望 '2.0', 得到 %v", response.JsonRPC)
	}

	if response.ID != 1 {
		t.Errorf("响应中的id字段错误: 期望 1, 得到 %v", response.ID)
	}

	if response.Error == nil {
		t.Fatal("响应中应该包含error字段")
	}

	if response.Error.Code != MethodNotFound {
		t.Errorf("错误代码错误: 期望 %d, 得到 %d", MethodNotFound, response.Error.Code)
	}

	if response.Error.Message != "Method not found: non_existent_method" {
		t.Errorf("错误消息错误: 期望 'Method not found: non_existent_method', 得到 %s", response.Error.Message)
	}
}

// 测试处理请求 - 成功的情况
func TestUnixServer_HandleRequest_Success(t *testing.T) {
	server := &UnixServer{
		socketPath: "/tmp/test_unix_server.sock",
		handlers:   make(map[string]RequestHandler),
		done:       make(chan struct{}),
	}

	// 注册一个测试处理器
	server.RegisterHandler("test_method", func(params map[string]interface{}) (interface{}, error) {
		return "success_result", nil
	})

	// 创建一个请求
	request := Request{
		JsonRPC: "2.0",
		Method:  "test_method",
		Params:  map[string]interface{}{"key": "value"},
		ID:      2,
	}

	// 处理请求
	response := server.handleRequest(request)

	// 验证响应
	if response.JsonRPC != "2.0" {
		t.Errorf("响应中的jsonrpc字段错误: 期望 '2.0', 得到 %v", response.JsonRPC)
	}

	if response.ID != 2 {
		t.Errorf("响应中的id字段错误: 期望 2, 得到 %v", response.ID)
	}

	if response.Result != "success_result" {
		t.Errorf("响应中的result字段错误: 期望 'success_result', 得到 %v", response.Result)
	}

	if response.Error != nil {
		t.Error("响应中不应该包含error字段")
	}
}

// 测试处理请求 - 处理器返回错误的情况
func TestUnixServer_HandleRequest_HandlerError(t *testing.T) {
	server := &UnixServer{
		socketPath: "/tmp/test_unix_server.sock",
		handlers:   make(map[string]RequestHandler),
		done:       make(chan struct{}),
	}

	// 注册一个返回错误的处理器
	server.RegisterHandler("error_method", func(params map[string]interface{}) (interface{}, error) {
		return nil, errors.New("handler error")
	})

	// 创建一个请求
	request := Request{
		JsonRPC: "2.0",
		Method:  "error_method",
		Params:  map[string]interface{}{},
		ID:      3,
	}

	// 处理请求
	response := server.handleRequest(request)

	// 验证响应
	if response.JsonRPC != "2.0" {
		t.Errorf("响应中的jsonrpc字段错误: 期望 '2.0', 得到 %v", response.JsonRPC)
	}

	if response.ID != 3 {
		t.Errorf("响应中的id字段错误: 期望 3, 得到 %v", response.ID)
	}

	if response.Error == nil {
		t.Fatal("响应中应该包含error字段")
	}

	if response.Error.Code != ParseError {
		t.Errorf("错误代码错误: 期望 %d, 得到 %d", ParseError, response.Error.Code)
	}

	if response.Error.Message != "handler error" {
		t.Errorf("错误消息错误: 期望 'handler error', 得到 %v", response.Error.Message)
	}
}

// 测试处理请求 - 无效方法类型的情况
func TestUnixServer_HandleRequest_InvalidMethodType(t *testing.T) {
	server := &UnixServer{
		socketPath: "/tmp/test_unix_server.sock",
		handlers:   make(map[string]RequestHandler),
		done:       make(chan struct{}),
	}

	// 创建一个请求，使用非字符串类型的方法
	request := Request{
		JsonRPC: "2.0",
		Method:  "invalid_method", // 使用字符串类型，因为 Request 结构体已经限制了 Method 为 string
		Params:  map[string]interface{}{},
		ID:      4,
	}

	// 处理请求
	response := server.handleRequest(request)

	// 验证响应
	if response.JsonRPC != "2.0" {
		t.Errorf("响应中的jsonrpc字段错误: 期望 '2.0', 得到 %v", response.JsonRPC)
	}

	if response.ID != 4 {
		t.Errorf("响应中的id字段错误: 期望 4, 得到 %v", response.ID)
	}

	if response.Error == nil {
		t.Fatal("响应中应该包含error字段")
	}

	if response.Error.Code != MethodNotFound {
		t.Errorf("错误代码错误: 期望 %d, 得到 %d", MethodNotFound, response.Error.Code)
	}

	if response.Error.Message != "Method not found: invalid_method" {
		t.Errorf("错误消息错误: 期望 'Method not found: invalid_method', 得到 %v", response.Error.Message)
	}
}

// 测试启动和停止服务器
func TestUnixServer_StartStop(t *testing.T) {
	// 使用临时目录创建socket路径
	tempDir, err := os.MkdirTemp("", "unix_server_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	socketPath := filepath.Join(tempDir, "test.sock")
	server := &UnixServer{
		socketPath: socketPath,
		handlers:   make(map[string]RequestHandler),
		done:       make(chan struct{}),
	}

	// 启动服务器
	err = server.Start()
	if err != nil {
		t.Fatalf("启动服务器失败: %v", err)
	}

	// 验证socket文件是否存在
	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		t.Error("socket文件应该存在")
	}

	// 验证socket文件权限
	info, err := os.Stat(socketPath)
	if err != nil {
		t.Errorf("获取socket文件信息失败: %v", err)
	} else {
		perm := info.Mode().Perm()
		if perm != 0666 {
			t.Errorf("socket文件权限错误: 期望 0666, 得到 %o", perm)
		}
	}

	// 停止服务器
	err = server.Stop()
	if err != nil {
		t.Errorf("停止服务器失败: %v", err)
	}

	// 验证done通道是否已关闭
	select {
	case _, ok := <-server.done:
		if ok {
			t.Error("done通道应该已关闭")
		}
	default:
		t.Error("done通道应该已关闭并可读取")
	}
}

// 测试客户端和服务器之间的通信
func TestUnixServer_ClientServerCommunication(t *testing.T) {
	// 使用临时目录创建socket路径
	tempDir, err := os.MkdirTemp("", "unix_server_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	socketPath := filepath.Join(tempDir, "test.sock")
	server := NewUnixServer(socketPath)

	// 注册一个回显处理器
	server.RegisterHandler("echo", func(params map[string]interface{}) (interface{}, error) {
		return params["message"], nil
	})

	// 启动服务器
	err = server.Start()
	if err != nil {
		t.Fatalf("启动服务器失败: %v", err)
	}
	defer server.Stop()

	// 给服务器一些时间启动
	time.Sleep(100 * time.Millisecond)

	// 创建客户端连接
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("连接到服务器失败: %v", err)
	}
	defer conn.Close()

	// 创建请求
	request := Request{
		JsonRPC: "2.0",
		Method:  "echo",
		Params:  map[string]interface{}{"message": "hello"},
		ID:      1,
	}

	// 使用jsoniter发送请求
	requestBytes, err := jsoniter.Marshal(request)
	if err != nil {
		t.Fatalf("序列化请求失败: %v", err)
	}

	// 直接写入字节并添加换行符
	_, err = conn.Write(requestBytes)
	if err != nil {
		t.Fatalf("发送请求失败: %v", err)
	}

	// 使用jsoniter接收响应
	var response Response
	decoder := jsoniter.NewDecoder(conn)
	err = decoder.Decode(&response)
	if err != nil {
		t.Fatalf("接收响应失败: %v", err)
	}

	// 验证响应
	if response.JsonRPC != "2.0" {
		t.Errorf("响应中的jsonrpc字段错误: 期望 '2.0', 得到 %v", response.JsonRPC)
	}

	if response.ID != 1 {
		t.Errorf("响应中的id字段错误: 期望 1, 得到 %v", response.ID)
	}

	if response.Result != "hello" {
		t.Errorf("响应中的result字段错误: 期望 'hello', 得到 %v", response.Result)
	}
}
