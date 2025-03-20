package gomcp

import (
	"fmt"
	"net"
	"os"
	"sync"

	jsoniter "github.com/json-iterator/go"
)

// UnixServer 实现了基于 Unix Domain Socket 的 MCP 服务器
type UnixServer struct {
	socketPath string
	listener   net.Listener
	handlers   map[string]RequestHandler
	done       chan struct{}
	mu         sync.RWMutex // 保护 handlers 的并发访问
	err        error
}

// NewUnixServer 创建一个新的 Unix Domain Socket MCP 服务器
func NewUnixServer(socketPath string) Server {
	return &UnixServer{
		socketPath: socketPath,
		handlers:   make(map[string]RequestHandler),
		done:       make(chan struct{}),
	}
}

// RegisterHandler 注册一个请求处理程序
func (s *UnixServer) RegisterHandler(method string, handler RequestHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[method] = handler
}

// Start 启动服务器
func (s *UnixServer) Start() error {
	// 确保 socket 文件不存在
	_ = os.Remove(s.socketPath)

	// 创建 Unix Domain Socket
	listener, err := net.Listen("unix", s.socketPath)
	if err != nil {
		return fmt.Errorf("failed to create socket: %w", err)
	}
	s.listener = listener

	// 设置 socket 文件权限
	if err := os.Chmod(s.socketPath, 0666); err != nil {
		return fmt.Errorf("failed to set socket permissions: %w", err)
	}

	go Safe(s.acceptConnections)()
	return nil
}

// Stop 停止服务器
func (s *UnixServer) Stop() error {
	close(s.done)
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

// Wait 阻塞直到服务器停止或发生错误
func (s *UnixServer) Wait() {
	<-s.done
}

func (s *UnixServer) acceptConnections() {
	for {
		select {
		case <-s.done:
			return
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				if isClosedError(err) {
					return // 服务器已关闭，退出
				}
				continue
			}

			// 启动新的 goroutine 处理连接
			go Safe(func() {
				s.handleConnection(conn)
			})()
		}
	}
}

func (s *UnixServer) handleConnection(conn net.Conn) {
	defer func() {
		if err := conn.Close(); err != nil && !isClosedError(err) {
			s.err = err
		}
	}()

	decoder := jsoniter.NewDecoder(conn)
	encoder := jsoniter.NewEncoder(conn)

	for {
		// 检查是否已关闭
		select {
		case <-s.done:
			return
		default:
			// 继续处理
		}

		var request Request
		if err := decoder.Decode(&request); err != nil {
			if !isClosedError(err) {
				s.err = err
			}
			return
		}

		// 处理请求
		response := s.handleRequest(request)

		// 使用互斥锁保护写入操作
		if err := encoder.Encode(response); err != nil {
			if !isClosedError(err) {
				s.err = err
			}
			return
		}
	}
}

func (s *UnixServer) handleRequest(request Request) Response {
	// 构建基本响应
	response := Response{
		JsonRPC: "2.0",
		ID:      request.ID,
	}

	// 查找并执行处理程序
	s.mu.RLock()
	handler, exists := s.handlers[request.Method]
	s.mu.RUnlock()

	if exists {
		result, err := handler(request.Params)
		if err != nil {
			response.Error = &Error{
				Code:    ParseError,
				Message: err.Error(),
			}
		} else {
			response.Result = result
		}
	} else {
		response.Error = &Error{
			Code:    MethodNotFound,
			Message: fmt.Sprintf("Method not found: %s", request.Method),
		}
	}

	return response
}
