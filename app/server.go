package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here! ")

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go MessageHandler(conn)
	}
}

func MessageHandler(conn net.Conn) {
	for {
		data := make([]byte, 1024)
		_, err := conn.Read(data)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			fmt.Println("Error reading: ", err.Error())
			os.Exit(1)
		}
		conn.Write([]byte("+PONG\r\n"))
	}
	conn.Close()
}
