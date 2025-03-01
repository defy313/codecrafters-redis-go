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
	"time"
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
	SET  command = "SET"
	GET  command = "GET"
)

type argument string

const (
	PX argument = "PX"
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

		switch strings.ToUpper(params[0]) {
		case string(ECHO):
			conn.Write([]byte("$" + strconv.Itoa(len(params[1])) + delimString + params[1] + delimString))
		case string(PING):
			conn.Write([]byte("+PONG\r\n"))
		case string(SET):
			err = setValue(params[1:])
			if err != nil {
				conn.Write([]byte("-ERR not enough arguments" + delimString))
				continue
			}
			conn.Write([]byte("+OK\r\n"))
		case string(GET):
			val, ok := getValue(params[1])
			expired := checkExpired(params[1])
			if !ok || expired {
				conn.Write([]byte("$-1\r\n"))
				return
			}
			conn.Write([]byte("$" + strconv.Itoa(len(val)) + delimString + val + delimString))
		}
	}
	conn.Close()
}

func checkExpired(key string) bool {
	expiryInMilli, exists := os.LookupEnv(key + "expiry")
	if !exists {
		return false
	}
	val, err := strconv.ParseInt(expiryInMilli, 10, 0)
	if err != nil {
		fmt.Printf("unable to read expiry for key: %s, err: %v", key, err)
		return false
	}

	return time.Now().UnixMilli() > val
}

// GetValue check if the key is present and returns if so
func getValue(key string) (string, bool) {
	return os.LookupEnv(key)
}

// SetValue sets the value
func setValue(params []string) (err error) {
	if len(params) < 2 {
		return errors.New("must provide at least key and value pair for set command")
	}

	os.Setenv(params[0], params[1])

	// let's set the expiry using unix timestamp
	idx := 2
	for idx < len(params) {
		switch strings.ToUpper(params[idx]) {
		case string(PX):
			expiryInMilli, err := strconv.Atoi(params[idx+1])
			if err != nil {
				return errors.New("expiry must be an integer value")
			}
			expiryTime := time.Now().Add(time.Millisecond * time.Duration(expiryInMilli)).UnixMilli()
			os.Setenv(params[0]+"expiry", strconv.FormatInt(expiryTime, 10))
			idx += 2
		default:
			idx++
		}
	}

	return
}
