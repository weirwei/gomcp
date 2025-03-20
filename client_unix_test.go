package gomcp

import (
	"errors"
	jsoniter "github.com/json-iterator/go"
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// 测试UnixClient与UnixServer的通信
func TestUnixClient_Communication(t *testing.T) {
	// 创建临时目录用于socket文件
	tempDir, err := os.MkdirTemp("", "unix_client_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	socketPath := filepath.Join(tempDir, "test.sock")

	// 创建并启动服务器
	server := NewUnixServer(socketPath)

	// 注册一个回显处理器
	server.RegisterHandler("echo", func(params map[string]interface{}) (interface{}, error) {
		return params["message"], nil
	})

	// 注册一个返回错误的处理器
	server.RegisterHandler("error", func(params map[string]interface{}) (interface{}, error) {
		return nil, errors.New("test error")
	})

	// 启动服务器
	err = server.Start()
	if err != nil {
		t.Fatalf("启动服务器失败: %v", err)
	}
	defer server.Stop()

	// 给服务器一些时间启动
	time.Sleep(100 * time.Millisecond)

	// 创建客户端
	client, err := NewUnixClient(socketPath)
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}
	defer client.Close()

	// 测试发送请求和接收响应 - 成功情况
	t.Run("成功请求", func(t *testing.T) {
		// 发送请求
		err = client.SendRequest("echo", map[string]interface{}{"message": "hello"})
		if err != nil {
			t.Fatalf("发送请求失败: %v", err)
		}

		// 接收响应
		response, err := client.ReceiveResponse()
		if err != nil {
			t.Fatalf("接收响应失败: %v", err)
		}

		// 验证响应
		if response["jsonrpc"] != "2.0" {
			t.Errorf("响应中的jsonrpc字段错误: 期望 '2.0', 得到 %v", response["jsonrpc"])
		}

		if response["result"] != "hello" {
			t.Errorf("响应中的result字段错误: 期望 'hello', 得到 %v", response["result"])
		}
	})

	// 测试发送请求和接收响应 - 错误情况
	t.Run("错误请求", func(t *testing.T) {
		// 发送请求
		err = client.SendRequest("error", map[string]interface{}{})
		if err != nil {
			t.Fatalf("发送请求失败: %v", err)
		}

		// 接收响应
		response, err := client.ReceiveResponse()
		if err != nil {
			t.Fatalf("接收响应失败: %v", err)
		}

		// 验证响应
		if response["jsonrpc"] != "2.0" {
			t.Errorf("响应中的jsonrpc字段错误: 期望 '2.0', 得到 %v", response["jsonrpc"])
		}

		errorObj, ok := response["error"].(map[string]interface{})
		if !ok {
			t.Fatal("响应中应该包含error字段")
		}

		if errorObj["message"] != "test error" {
			t.Errorf("错误消息错误: 期望 'test error', 得到 %v", errorObj["message"])
		}
	})

	// 测试方法不存在的情况
	t.Run("方法不存在", func(t *testing.T) {
		// 发送请求
		err = client.SendRequest("non_existent_method", map[string]interface{}{})
		if err != nil {
			t.Fatalf("发送请求失败: %v", err)
		}

		// 接收响应
		response, err := client.ReceiveResponse()
		if err != nil {
			t.Fatalf("接收响应失败: %v", err)
		}

		// 验证响应
		if response["jsonrpc"] != "2.0" {
			t.Errorf("响应中的jsonrpc字段错误: 期望 '2.0', 得到 %v", response["jsonrpc"])
		}

		errorObj, ok := response["error"].(map[string]interface{})
		if !ok {
			t.Fatal("响应中应该包含error字段")
		}

		if errorObj["code"] != float64(-32601) {
			t.Errorf("错误代码错误: 期望 -32601, 得到 %v", errorObj["code"])
		}
	})
}

// 测试连接到不存在的socket
func TestUnixClient_ConnectionError(t *testing.T) {
	// 使用一个不存在的socket路径
	socketPath := "/tmp/non_existent_socket.sock"

	// 尝试创建客户端
	_, err := NewUnixClient(socketPath)

	// 应该返回错误
	if err == nil {
		t.Error("应该返回连接错误，但没有")
	}
}

// 测试关闭连接
func TestUnixClient_Close(t *testing.T) {
	// 创建一个mock连接
	mockConn := &mockConn{}

	// 创建客户端
	client := &UnixClient{conn: mockConn}

	// 关闭连接
	err := client.Close()
	if err != nil {
		t.Errorf("关闭连接失败: %v", err)
	}

	// 验证连接已关闭
	if !mockConn.closed {
		t.Error("连接应该已关闭")
	}
}

// mockConn 实现了net.Conn接口，用于测试
type mockConn struct {
	closed    bool
	readData  []byte
	writeData []byte
}

func (m *mockConn) Read(b []byte) (n int, err error) {
	if m.closed {
		return 0, net.ErrClosed
	}
	if len(m.readData) == 0 {
		return 0, io.EOF
	}
	n = copy(b, m.readData)
	m.readData = m.readData[n:]
	return n, nil
}

func (m *mockConn) Write(b []byte) (n int, err error) {
	if m.closed {
		return 0, net.ErrClosed
	}
	m.writeData = append(m.writeData, b...)
	return len(b), nil
}

func (m *mockConn) Close() error {
	m.closed = true
	return nil
}

func (m *mockConn) LocalAddr() net.Addr {
	return &net.UnixAddr{Name: "mock", Net: "unix"}
}

func (m *mockConn) RemoteAddr() net.Addr {
	return &net.UnixAddr{Name: "mock", Net: "unix"}
}

func (m *mockConn) SetDeadline(t time.Time) error {
	return nil
}

func (m *mockConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (m *mockConn) SetWriteDeadline(t time.Time) error {
	return nil
}

// 测试使用mock连接的SendRequest和ReceiveResponse
func TestUnixClient_MockConnection(t *testing.T) {
	// 创建一个响应JSON
	responseJSON := `{"jsonrpc":"2.0","id":1,"result":"mock result"}`

	// 创建mock连接
	mockConn := &mockConn{
		readData: []byte(responseJSON),
	}

	// 创建客户端
	client := &UnixClient{conn: mockConn}

	// 发送请求
	err := client.SendRequest("test_method", map[string]interface{}{"key": "value"})
	if err != nil {
		t.Fatalf("发送请求失败: %v", err)
	}

	// 验证写入的数据
	var request map[string]interface{}
	err = jsoniter.Unmarshal(mockConn.writeData, &request)
	if err != nil {
		t.Fatalf("解析请求数据失败: %v", err)
	}

	if request["method"] != "test_method" {
		t.Errorf("请求中的method字段错误: 期望 'test_method', 得到 %v", request["method"])
	}

	// 接收响应
	response, err := client.ReceiveResponse()
	if err != nil {
		t.Fatalf("接收响应失败: %v", err)
	}

	// 验证响应
	if response["result"] != "mock result" {
		t.Errorf("响应中的result字段错误: 期望 'mock result', 得到 %v", response["result"])
	}
}
