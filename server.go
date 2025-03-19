package gomcp

import (
	"errors"
	"io"
)

// Server 定义了 MCP 服务器的接口
type Server interface {
	// Start 启动服务器
	Start() error
	// Stop 停止服务器
	Stop() error
	// RegisterHandler 注册一个请求处理器，用于处理指定的方法名
	RegisterHandler(method string, handler RequestHandler)
}

// RequestHandler 是处理特定请求方法的函数类型
type RequestHandler func(params map[string]interface{}) (interface{}, error)

// Response mcp server 的响应结构体
type Response struct {
	JsonRPC string `json:"jsonrpc"`
	ID      int    `json:"id,omitempty"`
	Result  any    `json:"result,omitempty"`
	Error   *Error `json:"error,omitempty"`
}

// isClosedError 检查错误是否是由于连接关闭导致的
func isClosedError(err error) bool {
	return errors.Is(err, io.EOF)
}
