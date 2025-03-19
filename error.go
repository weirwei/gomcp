package gomcp

type ErrorCode int

// 参考了Python 的 sdk https://github.com/modelcontextprotocol/python-sdk/blob/08f4e01b8f9ab77417f08738bb5cec26a5ebc94f/src/mcp/types.py#L144
var (
	ParseError     ErrorCode = -32700
	InvalidRequest ErrorCode = -32600
	MethodNotFound ErrorCode = -32601
	InvalidParams  ErrorCode = -32602
	InternalError  ErrorCode = -32603
)

// Error 错误信息
type Error struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}
