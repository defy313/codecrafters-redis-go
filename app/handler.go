package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type Data struct {
	Value  string `json:"value"`
	Expiry int64  `json:"expiry,omitempty"`
}

type command string

const (
	CONFIG command = "CONFIG"
	ECHO   command = "ECHO"
	PING   command = "PING"
	SET    command = "SET"
	GET    command = "GET"
)

type argument string

const (
	PX argument = "PX"
)

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
			conn.Write(encodeSimpleString(commands[2].Token))
		case string(PING):
			conn.Write([]byte("+PONG\r\n"))
		case string(SET):
			err = setValue(commands[1:])
			if err != nil {
				conn.Write([]byte("-ERR " + err.Error() + delimString))
				continue
			}
			conn.Write([]byte("+OK\r\n"))
		case string(CONFIG):
			val, err := ConfigHandler(commands[1:])
			if err != nil {
				conn.Write([]byte("-ERR " + err.Error() + delimString))
				continue
			}
			conn.Write(encodeArrayBulkStrings([]string{commands[2].Token, val}))
		case string(GET):
			val, ok := getValue(commands[2].Token)
			if !ok {
				conn.Write([]byte("$-1\r\n"))
				continue
			}
			conn.Write(encodeSimpleString(val))
		}
	}
	conn.Close()
}

func encodeArrayBulkStrings(values []string) []byte {
	encodedVal := "*"
	encodedVal += strconv.Itoa(len(values)) + delimString

	for _, val := range values {
		encodedVal += "$" + strconv.Itoa(len(val)) + delimString + val + delimString
	}

	return []byte(encodedVal)
}
func ConfigHandler(commands []Command) (string, error) {
	switch strings.ToUpper(commands[1].Token) {
	case "GET":
		if commands[2].Token == "dir" {
			return *dir, nil
		} else if commands[2].Token == "dbfilename" {
			return *dbfilename, nil
		} else {
			return "", errors.New("unknown config")
		}
	default:
		return "", errors.New("invalid config command")
	}

}
func encodeSimpleString(val string) []byte {
	return []byte("$" + strconv.Itoa(len(val)) + delimString + val + delimString)
}

// GetValue check if the key is present and returns if so
func getValue(key string) (string, bool) {
	val, ok := os.LookupEnv(key)
	if !ok {
		return "", false
	}

	var data Data
	json.Unmarshal([]byte(val), &data)
	if data.Expiry != 0 && time.Now().UnixMilli() > data.Expiry {
		os.Unsetenv(key)
		return "", false
	}

	return data.Value, true
}

// SetValue sets the value
func setValue(params []Command) (err error) {
	if len(params) < 2 {
		return errors.New("must provide at least key and value pair for set command")
	}

	data := Data{Value: params[2].Token}

	// let's set the expiry using unix timestamp
	idx := 3
	for idx < len(params) {
		switch strings.ToUpper(params[idx].Token) {
		case string(PX):
			expiryInMilli, err := strconv.Atoi(params[idx+1].Token)
			if err != nil {
				return errors.New("expiry must be an integer value")
			}
			data.Expiry = time.Now().Add(time.Millisecond * time.Duration(expiryInMilli)).UnixMilli()
			idx += 2
		default:
			idx++
		}
	}

	jsonVal, _ := json.Marshal(data)
	os.Setenv(params[1].Token, string(jsonVal))
	return
}
