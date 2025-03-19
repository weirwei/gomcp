package gomcp

// Client 定义了 MCP 客户端的接口
type Client interface {
	// SendRequest 发送 MCP 请求
	SendRequest(method string, params map[string]interface{}) error
	// ReceiveResponse 接收 MCP 响应
	ReceiveResponse() (map[string]interface{}, error)
	// Close 关闭客户端连接
	Close() error
}

// Request 定义了 MCP 请求结构体
type Request struct {
	JsonRPC string                 `json:"jsonrpc"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params,omitempty"`
	ID      int                    `json:"id"`
}
