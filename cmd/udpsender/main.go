package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	updAddr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		panic(err)
	}

	conn, err := net.DialUDP("udp", nil, updAddr)
	if err != nil {
		panic(err)
	}

	defer conn.Close()
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf(">")
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Error reading string from Stdin")
			os.Exit(1)
		}

		_, err = conn.Write([]byte(line))
		if err != nil {
			log.Println("Error writing to UDP connection")
			os.Exit(1)
		}

		fmt.Printf("Message sent: %s", line)
	}
}
