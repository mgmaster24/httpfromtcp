package main

import (
	"fmt"
	"net"

	"github.com/mgmaster24/httpfromtcp/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", "localhost:42069")
	if err != nil {
		panic(err)
	}

	defer listener.Close()
	for {
		connection, err := listener.Accept()
		if err != nil {
			panic(err)
		}

		fmt.Println("Connection established")
		request, err := request.RequestFromReader(connection)
		if err != nil {
			panic(err)
		}

		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", request.RequestLine.Method)
		fmt.Printf("- Target: %s\n", request.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", request.RequestLine.HttpVersion)
		fmt.Println("Headers:")
		for k, v := range request.Headers {
			fmt.Printf("- %s: %s\n", k, v)
		}
		fmt.Println("Body:")
		fmt.Println(string(request.Body))

		fmt.Println("Connection/Channel Closed")
	}
}
