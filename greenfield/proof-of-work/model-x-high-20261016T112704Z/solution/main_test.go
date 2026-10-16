package main

import (
	"bufio"
	"fmt"
	"net"
	"sync"
	"testing"
)

func startTestServer(t *testing.T) string {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	addr := listener.Addr().String()

	store := NewStore()
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go handleClient(conn, store)
		}
	}()

	t.Cleanup(func() {
		listener.Close()
	})

	return addr
}

func sendCommand(reader *bufio.Reader, writer *bufio.Writer, lines ...string) (string, error) {
	for _, l := range lines {
		if _, err := writer.WriteString(l + "\r\n"); err != nil {
			return "", err
		}
	}
	if err := writer.Flush(); err != nil {
		return "", err
	}
	return reader.ReadString('\n')
}

func TestRedisCommands(t *testing.T) {
	addr := startTestServer(t)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	// Test PING
	resp, err := sendCommand(reader, writer, "*1", "$4", "PING")
	if err != nil || resp != "+PONG\r\n" {
		t.Fatalf("PING failed: got %q, err %v", resp, err)
	}

	// Test ping lower-case
	resp, err = sendCommand(reader, writer, "*1", "$4", "ping")
	if err != nil || resp != "+PONG\r\n" {
		t.Fatalf("ping failed: got %q, err %v", resp, err)
	}

	// Test ping with argument (should error)
	resp, err = sendCommand(reader, writer, "*2", "$4", "PING", "$4", "pong")
	if err != nil || resp != "-ERR wrong number of arguments for 'ping' command\r\n" {
		t.Fatalf("ping with args failed: got %q, err %v", resp, err)
	}

	// Test GET non-existent
	resp, err = sendCommand(reader, writer, "*2", "$3", "GET", "$3", "foo")
	if err != nil || resp != "$-1\r\n" {
		t.Fatalf("GET missing key failed: got %q, err %v", resp, err)
	}

	// Test SET
	resp, err = sendCommand(reader, writer, "*3", "$3", "SET", "$3", "foo", "$5", "hello")
	if err != nil || resp != "+OK\r\n" {
		t.Fatalf("SET failed: got %q, err %v", resp, err)
	}

	// Test GET existing
	resp, err = sendCommand(reader, writer, "*2", "$3", "GET", "$3", "foo")
	if err != nil || resp != "$5\r\n" {
		t.Fatalf("GET length failed: got %q, err %v", resp, err)
	}
	val, err := reader.ReadString('\n')
	if err != nil || val != "hello\r\n" {
		t.Fatalf("GET value failed: got %q, err %v", val, err)
	}

	// Test SET overwrite
	resp, err = sendCommand(reader, writer, "*3", "$3", "set", "$3", "foo", "$5", "world")
	if err != nil || resp != "+OK\r\n" {
		t.Fatalf("SET overwrite failed: got %q, err %v", resp, err)
	}
	resp, err = sendCommand(reader, writer, "*2", "$3", "get", "$3", "foo")
	if err != nil || resp != "$5\r\n" {
		t.Fatalf("GET length failed: got %q, err %v", resp, err)
	}
	val, err = reader.ReadString('\n')
	if err != nil || val != "world\r\n" {
		t.Fatalf("GET value failed: got %q, err %v", val, err)
	}

	// Test DEL
	resp, err = sendCommand(reader, writer, "*3", "$3", "SET", "$1", "a", "$1", "1")
	if err != nil || resp != "+OK\r\n" {
		t.Fatalf("SET a failed: got %q, err %v", resp, err)
	}
	resp, err = sendCommand(reader, writer, "*3", "$3", "SET", "$1", "b", "$1", "2")
	if err != nil || resp != "+OK\r\n" {
		t.Fatalf("SET b failed: got %q, err %v", resp, err)
	}

	// DEL without args
	resp, err = sendCommand(reader, writer, "*1", "$3", "DEL")
	if err != nil || resp != "-ERR wrong number of arguments for 'del' command\r\n" {
		t.Fatalf("DEL without args failed: got %q, err %v", resp, err)
	}

	// DEL distinct keys + duplicates + non-existent keys
	// keys: a, b, a, c (a and b exist, c doesn't exist, a repeated)
	resp, err = sendCommand(reader, writer, "*5", "$3", "DEL", "$1", "a", "$1", "b", "$1", "a", "$1", "c")
	if err != nil || resp != ":2\r\n" {
		t.Fatalf("DEL failed: got %q, err %v", resp, err)
	}

	// DEL again (keys already deleted)
	resp, err = sendCommand(reader, writer, "*3", "$3", "DEL", "$1", "a", "$1", "b")
	if err != nil || resp != ":0\r\n" {
		t.Fatalf("DEL again failed: got %q, err %v", resp, err)
	}

	// Test unknown command
	resp, err = sendCommand(reader, writer, "*1", "$4", "INFO")
	if err != nil || resp != "-ERR unknown command 'INFO'\r\n" {
		t.Fatalf("unknown command failed: got %q, err %v", resp, err)
	}

	// Verify connection is still usable after error
	resp, err = sendCommand(reader, writer, "*1", "$4", "PING")
	if err != nil || resp != "+PONG\r\n" {
		t.Fatalf("PING after error failed: got %q, err %v", resp, err)
	}
}

func TestConcurrentClients(t *testing.T) {
	addr := startTestServer(t)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(clientNum int) {
			defer wg.Done()
			conn, err := net.Dial("tcp", addr)
			if err != nil {
				t.Errorf("client %d dial failed: %v", clientNum, err)
				return
			}
			defer conn.Close()

			reader := bufio.NewReader(conn)
			writer := bufio.NewWriter(conn)

			key := fmt.Sprintf("key_%d", clientNum)
			val := fmt.Sprintf("val_%d", clientNum)

			// SET
			resp, err := sendCommand(reader, writer, "*3", "$3", "SET", fmt.Sprintf("$%d", len(key)), key, fmt.Sprintf("$%d", len(val)), val)
			if err != nil || resp != "+OK\r\n" {
				t.Errorf("client %d SET failed: got %q, err %v", clientNum, resp, err)
				return
			}

			// GET
			resp, err = sendCommand(reader, writer, "*2", "$3", "GET", fmt.Sprintf("$%d", len(key)), key)
			expectedHeader := fmt.Sprintf("$%d\r\n", len(val))
			if err != nil || resp != expectedHeader {
				t.Errorf("client %d GET header failed: got %q, err %v", clientNum, resp, err)
				return
			}
			body, err := reader.ReadString('\n')
			if err != nil || body != val+"\r\n" {
				t.Errorf("client %d GET body failed: got %q, err %v", clientNum, body, err)
				return
			}
		}(i)
	}
	wg.Wait()
}

func TestEmptyCommand(t *testing.T) {
	addr := startTestServer(t)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	// Send array of length 0
	resp, err := sendCommand(reader, writer, "*0")
	if err != nil || resp != "-ERR empty command\r\n" {
		t.Fatalf("empty command failed: got %q, err %v", resp, err)
	}

	// Connection should still work
	resp, err = sendCommand(reader, writer, "*1", "$4", "PING")
	if err != nil || resp != "+PONG\r\n" {
		t.Fatalf("PING after empty array failed: got %q, err %v", resp, err)
	}
}

