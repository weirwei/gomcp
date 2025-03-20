package gomcp

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"
	"time"
)

// 测试注册处理器
func TestStdioServer_RegisterHandler(t *testing.T) {
	server := NewStdioServer(nil, nil)

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

// 测试启动和停止服务器
func TestStdioServer_StartStop(t *testing.T) {
	reader, writer, _ := os.Pipe()
	server := NewStdioServer(reader, writer)

	// 启动服务器
	err := server.Start()
	if err != nil {
		t.Errorf("启动服务器出错: %v", err)
	}

	// 给一些时间让goroutine启动
	time.Sleep(10 * time.Millisecond)

	// 停止服务器
	err = server.Stop()
	if err != nil {
		t.Errorf("停止服务器出错: %v", err)
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

// 测试处理请求 - 方法不存在的情况
func TestStdioServer_HandleRequest_MethodNotFound(t *testing.T) {
	server := NewStdioServer(nil, nil)

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

	err := response.Error
	if err == nil {
		t.Fatal("响应中应该包含error字段")
	}

	if err.Code != -32601 {
		t.Errorf("错误代码错误: 期望 -32601, 得到 %v", err.Code)
	}

	if !strings.Contains(err.Message, "non_existent_method") {
		t.Errorf("错误消息应该包含方法名: %s", err.Message)
	}
}

// 测试处理请求 - 成功的情况
func TestStdioServer_HandleRequest_Success(t *testing.T) {
	server := NewStdioServer(nil, nil)

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
func TestStdioServer_HandleRequest_HandlerError(t *testing.T) {
	server := NewStdioServer(nil, nil)

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

	err := response.Error
	if err == nil {
		t.Fatal("响应中应该包含error字段")
	}

	if err.Code != -32603 {
		t.Errorf("错误代码错误: 期望 -32603, 得到 %v", err.Code)
	}
}

// 测试处理消息 - 模拟输入输出
func TestStdioServer_HandleMessages(t *testing.T) {
	// 创建输入和输出缓冲区
	inputBuffer := bytes.NewBufferString(`{"jsonrpc":"2.0","method":"echo","params":{"message":"hello"},"id":1}`)
	outputBuffer := &bytes.Buffer{}

	// 创建服务器并设置输入输出
	server := &StdioServer{
		reader:   inputBuffer,
		writer:   outputBuffer,
		done:     make(chan struct{}),
		handlers: make(map[string]RequestHandler),
	}

	// 注册一个回显处理器
	server.RegisterHandler("echo", func(params map[string]interface{}) (interface{}, error) {
		return params["message"], nil
	})

	// 启动消息处理
	go Safe(server.handleMessages)()

	// 等待处理完成
	time.Sleep(100 * time.Millisecond)

	// 停止服务器
	server.Stop()

	// 验证输出
	output := outputBuffer.String()
	if !strings.Contains(output, `"result":"hello"`) {
		t.Errorf("输出应该包含结果 'hello', 实际输出: %s", output)
	}

	if !strings.Contains(output, `"id":1`) {
		t.Errorf("输出应该包含id 1, 实际输出: %s", output)
	}
}
