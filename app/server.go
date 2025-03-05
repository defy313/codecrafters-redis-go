package main

import (
	"flag"
	"fmt"
	"net"
	"os"
)

var dir = flag.String("dir", "", "specify the rdb file directory")
var dbfilename = flag.String("dbfilename", "", "specify the name of rdb file")

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here! ")

	flag.Parse()

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
