package main

import (
	"bufio"
	"io"
	"strconv"
)

type DataType string

const (
	Arrays        DataType = "Arrays"
	Integers      DataType = "Integers"
	BulkStrings   DataType = "BulkStrings"
	SimpleErrors  DataType = "SimpleErrors"
	SimpleStrings DataType = "SimpleStrings"
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

//// ReadPart "*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nbar\r\n",
//// ReadPart return the next part
//func ReadPart(reader *bufio.Reader) (command Command, err error) {
//	part, err := reader.ReadString(delimByte)
//	if err != nil {
//		return
//	}
//
//	// remove the trailing \r\n
//	part = strings.TrimSuffix(part, delimString)
//
//	switch part[0] {
//	case '$':
//		// this is a bulk string, first part only represents size
//		size, err := strconv.Atoi(part[1:])
//		if err != nil {
//			return command, err
//		}
//		// now I will just read size part from the reader
//		buf := make([]byte, size+2)
//		_, err = io.ReadFull(reader, buf)
//		if err != nil {
//			return command, err
//		}
//		return Command{BulkStrings, string(buf[:size])}, nil
//	case '*':
//		return Command{Arrays, ""}, nil
//	default:
//		// We can directly read the content excluding the first
//		return Command{SimpleStrings, strings.TrimSuffix(part[1:], delimString)}, nil
//	}
//}
