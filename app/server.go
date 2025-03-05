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

var dataTypeMap = map[uint8]DataType{
	'*': Arrays,
	':': Integers,
	'$': BulkStrings,
	'-': SimpleErrors,
	'+': SimpleStrings,
}

// MessageHandler returns the decoded message as a slice of strings
func MessageHandler(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		commands, err := DecodeMessage(reader)
		if errors.Is(err, io.EOF) {
			fmt.Println("Client closed connection, exiting")
			break
		}
		if err != nil {
			break
		}

		if commands[0].Type != Arrays {
			// command is not an array of bulk strings
			// check if it's just a PING
			if commands[0].Type == SimpleStrings && commands[0].Token == string(PING) {
				conn.Write([]byte("+PONG\r\n"))
				continue
			} else {
				conn.Write([]byte("-ERR unknown command " + commands[0].Token + delimString))
				continue
			}
		}

		switch strings.ToUpper(commands[1].Token) {
		case string(ECHO):
			conn.Write([]byte("$" + strconv.Itoa(len(commands[2].Token)) + delimString + commands[2].Token + delimString))
		case string(PING):
			conn.Write([]byte("+PONG\r\n"))
		case string(SET):
			err = setValue(commands[1:])
			if err != nil {
				conn.Write([]byte("-ERR not enough arguments" + delimString))
				continue
			}
			conn.Write([]byte("+OK\r\n"))
		case string(GET):
			val, ok := getValue(commands[2].Token)
			expired := checkExpired(commands[2].Token)
			if !ok || expired {
				conn.Write([]byte("$-1\r\n"))
				continue
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
func setValue(params []Command) (err error) {
	if len(params) < 2 {
		return errors.New("must provide at least key and value pair for set command")
	}

	os.Setenv(params[1].Token, params[2].Token)

	// let's set the expiry using unix timestamp
	idx := 3
	for idx < len(params) {
		switch strings.ToUpper(params[idx].Token) {
		case string(PX):
			expiryInMilli, err := strconv.Atoi(params[idx+1].Token)
			if err != nil {
				return errors.New("expiry must be an integer value")
			}
			expiryTime := time.Now().Add(time.Millisecond * time.Duration(expiryInMilli)).UnixMilli()
			os.Setenv(params[1].Token+"expiry", strconv.FormatInt(expiryTime, 10))
			idx += 2
		default:
			idx++
		}
	}

	return
}
