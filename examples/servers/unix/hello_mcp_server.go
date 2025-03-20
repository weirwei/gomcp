package main

import (
	"fmt"
	"os"

	"github.com/weirwei/gomcp"
)

func main() {
	args := os.Args
	if len(args) < 2 {
		fmt.Println("Usage: mcp [server|client]")
		os.Exit(1)
	} else if args[1] == "server" {
		helloServer()
	} else if args[1] == "client" {
		helloClient(args[2])
	} else {
		fmt.Println("Unknown command:", args[1])
	}
}

func helloServer() {
	server := gomcp.NewUnixServer("/tmp/mcp.sock")
	server.RegisterHandler("hello", func(params map[string]interface{}) (interface{}, error) {
		fmt.Printf("Received request: %+v\n", params)
		return "Hello World!", nil
	})
	server.RegisterHandler("close", func(params map[string]interface{}) (interface{}, error) {
		server.Stop()
		return "close", nil
	})
	server.Start()
	server.(*gomcp.UnixServer).Wait()
}

func helloClient(method string) {
	client, err := gomcp.NewUnixClient("/tmp/mcp.sock")
	if err != nil {
		panic(err)
	}
	defer client.Close()
	err = client.SendRequest(method, nil)
	if err != nil {
		panic(err)
	}
	result, err := client.ReceiveResponse()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Received response: %+v\n", result)
}
