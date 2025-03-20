package gomcp

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
)

// UnixClient 实现了基于 Unix Domain Socket 的 MCP 客户端
type UnixClient struct {
	conn net.Conn
}

// NewUnixClient 创建一个新的 Unix Domain Socket MCP 客户端
func NewUnixClient(socketPath string) (Client, error) {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to socket: %w", err)
	}
	return &UnixClient{conn: conn}, nil
}

// Close 关闭客户端连接
func (c *UnixClient) Close() error {
	return c.conn.Close()
}

// SendRequest 发送 MCP 请求
func (c *UnixClient) SendRequest(method string, params map[string]interface{}) error {
	request := Request{
		JsonRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      1,
	}
	return json.NewEncoder(c.conn).Encode(request)
}

// ReceiveResponse 接收 MCP 响应
func (c *UnixClient) ReceiveResponse() (map[string]interface{}, error) {
	var response map[string]interface{}
	err := json.NewDecoder(c.conn).Decode(&response)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return response, nil
}
