package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
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

const delimString = "\r\n"
const delimByte = '\n'

type command string

const (
	ECHO command = "ECHO"
	PING command = "PING"
)

type DataType string

const (
	Arrays        DataType = "Arrays"
	Integers      DataType = "Integers"
	BulkStrings   DataType = "BulkStrings"
	SimpleErrors  DataType = "SimpleErrors"
	SimpleStrings DataType = "SimpleStrings"
)

var dataTypeMap = map[uint8]DataType{
	'*': Arrays,
	':': Integers,
	'$': BulkStrings,
	'-': SimpleErrors,
	'+': SimpleStrings,
}

//var handlerByDataType = map[DataType]func(string, net.Conn) []string{
//	BulkStrings: BulkStringsHandler,
//	Arrays:      ArraysHandler,
//}

/*
*2\r\n$4\r\nECHO\r\n$3\r\nhey\r\n
 */

// ArraysHandler handler input of type Arrays
func ArraysHandler(prevData string, reader *bufio.Reader) (params []string) {
	size, err := strconv.Atoi(prevData[1:])
	if err != nil {
		fmt.Println("unable to figure out size of the array, err: ", err.Error())
		return
	}
	// We have size number of elements
	// each of which have to be handled here
	for i := 0; i < size; i++ {
		childParams, err := DecodeMessage(reader)
		if err != nil {
			return
		}
		params = append(params, childParams...)
	}
	return
}

// BulkStringsHandler handles input of type BulkString
func BulkStringsHandler(prevData string, reader *bufio.Reader) (params []string) {
	size, err := strconv.Atoi(prevData[1:])
	if err != nil {
		fmt.Println("unable to read the size of the string, err: ", err.Error())
		return
	}

	// Adding 2 for reading the delimiter string
	buf := make([]byte, size+2)
	_, err = io.ReadFull(reader, buf)
	if err != nil {
		fmt.Println("unable to read bulk string, err: ", err.Error())
		return
	}

	arg := strings.TrimSuffix(string(buf), delimString)
	params = append(params, arg)
	return
}

func DecodeMessage(reader *bufio.Reader) (params []string, err error) {
	data, err := reader.ReadString(delimByte)
	if err != nil {
		return
	}

	// strip the delimiter from the data
	data = strings.TrimSuffix(data, delimString)

	// figure out the dataType
	dataType, ok := dataTypeMap[data[0]]
	if !ok {
		fmt.Println("Unrecognized dataType marker: ", data[0])
		return
	}

	if dataType == Arrays {
		params = ArraysHandler(data, reader)
	} else if dataType == BulkStrings {
		params = BulkStringsHandler(data, reader)
	}

	return
}

// MessageHandler returns the decoded message as a slice of strings
func MessageHandler(conn net.Conn) {
	for {
		reader := bufio.NewReader(conn)
		params, err := DecodeMessage(reader)
		if errors.Is(err, io.EOF) {
			fmt.Println("Client closed connection, exiting")
			break
		}
		if err != nil {
			break
		}

		switch params[0] {
		case string(ECHO):
			conn.Write([]byte("$" + strconv.Itoa(len(params[1])) + "\r\n" + params[1] + "\r\n"))
		case string(PING):
			conn.Write([]byte("+PONG\r\n"))
		}
	}
	conn.Close()
	return
}
