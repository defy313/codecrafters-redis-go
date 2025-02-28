package main

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type TestMessageHandler struct {
	Input          string
	ExpectedOutput string
}

func TestMessageHandler_MessageHandler(t *testing.T) {
	tests := []TestMessageHandler{
		{
			"*2\r\n$4\r\nECHO\r\n$3\r\nhey\r\n",
			"$3\r\nhey\r\n",
		},
		{
			"*1\r\n$4\r\nPING\r\n",
			"+PONG\r\n",
		},
	}

	serverConn, clientConn := net.Pipe()
	defer clientConn.Close()

	go MessageHandler(serverConn)

	for _, test := range tests {
		clientConn.Write([]byte(test.Input))

		time.Sleep(100 * time.Millisecond)

		buf := make([]byte, 1024)
		n, err := clientConn.Read(buf)
		if err != nil {
			t.Fatalf("error reading response: %v", err)
		}

		response := string(buf[:n])
		assert.Equal(t, test.ExpectedOutput, response, "Response should match expected value")
	}
}
