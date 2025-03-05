package main

import (
	"bufio"
	"io"
	"strconv"
)

type DataType string

var dataTypeMap = map[uint8]DataType{
	'*': Arrays,
	':': Integers,
	'$': BulkStrings,
	'-': SimpleErrors,
	'+': SimpleStrings,
}

const (
	Arrays        DataType = "Arrays"
	Integers      DataType = "Integers"
	BulkStrings   DataType = "BulkStrings"
	SimpleErrors  DataType = "SimpleErrors"
	SimpleStrings DataType = "SimpleStrings"
)

const (
	delimString = "\r\n"
	delimByte   = '\n'
)

type Command struct {
	Type  DataType
	Token string
}

func GetSize(reader *bufio.Reader) (int, error) {
	part, err := reader.ReadString(delimByte)
	if err != nil {
		return 0, err
	}
	// -2 for removing the delimiter string
	return strconv.Atoi(part[1 : len(part)-2])
}

func StringHandler(reader *bufio.Reader) (Command, error) {
	size, err := GetSize(reader)
	if err != nil {
		return Command{}, err
	}

	buf := make([]byte, size+2)
	io.ReadFull(reader, buf)

	return Command{SimpleStrings, string(buf[:size])}, nil
}

func DecodeMessage(reader *bufio.Reader) (commands []Command, err error) {
	prefix, err := reader.Peek(1)
	if err != nil {
		return commands, err
	}

	switch prefix[0] {
	case '$':
		cmd, err := StringHandler(reader)
		if err != nil {
			return commands, err
		}
		commands = append(commands, cmd)
	case '*':
		size, err := GetSize(reader)
		if err != nil {
			return commands, err
		}
		commands = append(commands, Command{Arrays, ""})
		for i := 0; i < size; i++ {
			cmd, err := StringHandler(reader)
			if err != nil {
				return commands, err
			}
			commands = append(commands, cmd)
		}
	default:
		part, err := reader.ReadString(delimByte)
		if err != nil {
			return commands, err
		}
		commands = append(commands, Command{SimpleStrings, part[1 : len(part)-2]})
	}

	return
}
