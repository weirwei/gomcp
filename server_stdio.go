package gomcp

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"io"
	"sync"
)

// StdioServer 实现了基于标准输入输出的 MCP 服务器
type StdioServer struct {
	reader   io.Reader
	writer   io.Writer
	done     chan struct{}
	handlers map[string]RequestHandler
	mu       sync.RWMutex // 保护 handlers 的并发访问
}

// NewStdioServer 创建一个新的标准输入输出 MCP 服务器
func NewStdioServer(reader io.Reader, writer io.Writer) *StdioServer {
	return &StdioServer{
		reader:   reader,
		writer:   writer,
		done:     make(chan struct{}),
		handlers: make(map[string]RequestHandler),
	}
}

// Start 启动服务器
func (s *StdioServer) Start() error {
	go Safe(s.handleMessages)()
	return nil
}

// Stop 停止服务器
func (s *StdioServer) Stop() error {
	close(s.done)
	return nil
}

// Wait 等待服务器停止
func (s *StdioServer) Wait() {
	<-s.done
}

// RegisterHandler 注册一个方法处理器
func (s *StdioServer) RegisterHandler(method string, handler RequestHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[method] = handler
}

func (s *StdioServer) handleMessages() {
	decoder := jsoniter.NewDecoder(s.reader)
	encoder := jsoniter.NewEncoder(s.writer)

	for {
		select {
		case <-s.done:
			return
		default:
			var (
				request  Request
				response Response
			)
			if err := decoder.Decode(&request); err != nil {
				if !isClosedError(err) {
					response.Error = &Error{
						Code:    ParseError,
						Message: fmt.Sprintf("Parse Error: %v", err),
					}
				}
				continue
			}

			// 处理请求
			response = s.handleRequest(request)
			if err := encoder.Encode(response); err != nil {
				response.Error = &Error{
					Code:    ParseError,
					Message: fmt.Sprintf("Encode Error: %v", err)}
				continue
			}
		}
	}
}

func (s *StdioServer) handleRequest(request Request) Response {
	// 构建基本响应
	response := Response{
		JsonRPC: "2.0",
		ID:      request.ID,
	}

	var (
		method = request.Method
		params = request.Params
	)

	// 查找处理器
	s.mu.RLock()
	handler, exists := s.handlers[method]
	s.mu.RUnlock()

	if !exists {
		response.Error = &Error{
			Code:    MethodNotFound,
			Message: fmt.Sprintf("Method not found: %s", method),
		}
		return response
	}

	// 调用处理器
	result, err := handler(params)
	if err != nil {
		response.Error = &Error{
			Code:    InternalError,
			Message: fmt.Sprintf("Method deal failed: %s", err.Error()),
		}
	} else {
		response.Result = result
	}

	return response
}
