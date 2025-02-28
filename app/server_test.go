package main

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMessageHandler(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer clientConn.Close()

	go MessageHandler(serverConn)

	testCommand := "*2\r\n$4\r\nECHO\r\n$3\r\nhey\r\n"
	clientConn.Write([]byte(testCommand))

	time.Sleep(100 * time.Millisecond)

	buf := make([]byte, 1024)
	n, err := clientConn.Read(buf)
	if err != nil {
		t.Fatalf("error reading response: %v", err)
	}

	expected := "$3\r\nhey\r\n"
	response := string(buf[:n])

	assert.Equal(t, expected, response, "Response should match expected value")
}
