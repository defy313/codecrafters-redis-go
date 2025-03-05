package main

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type TestMessageHandler struct {
	Input           string
	ExpectedOutput  string
	WaitTimeInMilli int
}

func runRedisTest(t testing.TB) {
	tests := []TestMessageHandler{
		{
			"*2\r\n$4\r\nECHO\r\n$3\r\nhey\r\n",
			"$3\r\nhey\r\n",
			0,
		},
		{
			"*1\r\n$4\r\nPING\r\n",
			"+PONG\r\n",
			0,
		},
		{
			"*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nbar\r\n",
			"+OK\r\n",
			0,
		},
		{
			"*5\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nbar\r\n$2\r\nPX\r\n$3\r\n500\r\n",
			"+OK\r\n",
			0,
		},
		{
			"*2\r\n$3\r\nGET\r\n$3\r\nfoo\r\n",
			"$3\r\nbar\r\n",
			1000,
		},
		{
			"*2\r\n$3\r\nGET\r\n$3\r\nfoo\r\n",
			"$-1\r\n",
			200,
		},
	}

	serverConn, clientConn := net.Pipe()
	defer clientConn.Close()

	go MessageHandler(serverConn)

	for _, test := range tests {
		clientConn.Write([]byte(test.Input))

		time.Sleep(time.Millisecond * time.Duration(test.WaitTimeInMilli))

		buf := make([]byte, 1024)
		n, err := clientConn.Read(buf)
		if err != nil {
			t.Fatalf("error reading response: %v", err)
		}

		response := string(buf[:n])
		assert.Equal(t, test.ExpectedOutput, response, "Response should match expected value")
	}
}

func TestMessageHandler_MessageHandler(t *testing.T) {
	runRedisTest(t)
}

func BenchmarkRedisServer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runRedisTest(b)
	}
}
