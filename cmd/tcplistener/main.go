package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
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
		linesChannel := getLinesChannel(connection)

		for line := range linesChannel {
			fmt.Println(line)
		}

		fmt.Println("Connection/Channel Closed")
	}
}

func getLinesChannel(connection net.Conn) <-chan string {
	lineChannel := make(chan string)
	go func() {
		currentLine := ""
		for {
			bytes := make([]byte, 8)
			n, err := connection.Read(bytes)
			if err != nil {
				if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) {
					if currentLine != "" {
						lineChannel <- currentLine
					}

					close(lineChannel)
					return
				}

				panic(err)
			}

			currentLine += string(bytes[:n])
			parts := strings.Split(currentLine, "\n")
			for i := 0; i < len(parts)-1; i++ {
				lineChannel <- parts[i]
			}

			currentLine = parts[len(parts)-1]
		}
	}()

	return lineChannel
}
