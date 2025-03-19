package gomcp

import (
	"bufio"
	"fmt"
	"io"
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"
)

// StdioClient 实现了基于标准输入输出的 MCP 客户端
type StdioClient struct {
	reader     io.ReadCloser
	writer     io.WriteCloser
	mutex      sync.Mutex
	responseCh chan map[string]interface{}
	errorCh    chan error
}

// NewStdioClient 创建一个新的标准输入输出 MCP 客户端
func NewStdioClient(reader io.ReadCloser, writer io.WriteCloser) *StdioClient {
	// 否则使用系统的标准输入和标准输出
	client := &StdioClient{
		reader:     reader,
		writer:     writer,
		responseCh: make(chan map[string]interface{}, 1),
		errorCh:    make(chan error, 1),
	}

	// 启动一个 goroutine 来读取响应
	go Safe(client.readResponses)()

	return client
}

// Close 关闭客户端连接（对于 StdioClient 来说是一个空操作）
func (c *StdioClient) Close() error {
	c.reader.Close()
	c.writer.Close()
	return nil
}

// SendRequest 发送 MCP 请求（通过标准输出）
func (c *StdioClient) SendRequest(method string, params map[string]interface{}) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	request := Request{
		JsonRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      1,
	}

	data, err := jsoniter.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}
	if _, err := c.writer.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write request: %w", err)
	}
	return nil
}

// ReceiveResponse 接收 MCP 响应（从标准输入）
func (c *StdioClient) ReceiveResponse() (map[string]interface{}, error) {
	select {
	case response := <-c.responseCh:
		return response, nil
	case err := <-c.errorCh:
		return nil, err
	case <-time.After(10 * time.Second):
		return nil, fmt.Errorf("timeout waiting for response")
	}
}

// readResponses 持续读取响应
func (c *StdioClient) readResponses() {
	reader := bufio.NewReader(c.reader)
	for {
		response := make(map[string]interface{})
		line, _, err := reader.ReadLine()
		if err != nil {
			if err != io.EOF {
				c.errorCh <- fmt.Errorf("failed to unmarshal response: %v", err)
				return
			}
		}
		if len(line) == 0 {
			continue
		}
		if err = jsoniter.Unmarshal(line, &response); err != nil {
			c.errorCh <- fmt.Errorf("failed to unmarshal response: %v", err)
			return
		}

		c.responseCh <- response
	}
}
