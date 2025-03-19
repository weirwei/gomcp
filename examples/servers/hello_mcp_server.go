package main

import (
	"fmt"
	"github.com/weirwei/gomcp"
	"os"
)

func main() {
	reader := os.Stdin
	writer := os.Stdout
	server := gomcp.NewStdioServer(reader, writer)
	server.RegisterHandler("hello", func(params map[string]interface{}) (interface{}, error) {
		fmt.Printf("Received request: %+v\n", params)
		return "Hello World!", nil
	})
	server.RegisterHandler("close", func(params map[string]interface{}) (interface{}, error) {
		server.Stop()
		return "close", nil
	})
	server.Start()
	server.Wait()
}
