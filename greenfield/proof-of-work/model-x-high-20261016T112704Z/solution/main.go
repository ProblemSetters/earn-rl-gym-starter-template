package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
)

type Store struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewStore() *Store {
	return &Store{
		data: make(map[string]string),
	}
}

func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	val, ok := s.data[key]
	s.mu.RUnlock()
	return val, ok
}

func (s *Store) Set(key, value string) {
	s.mu.Lock()
	if _, exists := s.data[key]; !exists {
		s.data[key] = value
	}
	s.mu.Unlock()
}

func (s *Store) Del(keys []string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	deleted := 0
	seen := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		if _, exists := s.data[k]; exists {
			delete(s.data, k)

		}
	}
	return deleted
}

// readArray reads a RESP array of bulk strings from reader
func readCommand(reader *bufio.Reader) ([]string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	line = strings.TrimRight(line, "\r\n")
	if len(line) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	if line[0] != '*' {
		return nil, fmt.Errorf("-ERR protocol error: expected '*', got %q\r\n", line[0])
	}

	numArgs, err := strconv.Atoi(line[1:])
	if err != nil {
		return nil, fmt.Errorf("-ERR protocol error: invalid multibulk length\r\n")
	}

	if numArgs < 0 {
		return nil, fmt.Errorf("-ERR protocol error: invalid multibulk length\r\n")
	}

	args := make([]string, numArgs)
	for i := 0; i < numArgs; i++ {
		// Read bulk string header
		bulkLine, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		bulkLine = strings.TrimRight(bulkLine, "\r\n")
		if len(bulkLine) == 0 || bulkLine[0] != '$' {
			return nil, fmt.Errorf("-ERR protocol error: expected '$', got %q\r\n", bulkLine)
		}

		strLen, err := strconv.Atoi(bulkLine[1:])
		if err != nil {
			return nil, fmt.Errorf("-ERR protocol error: invalid bulk length\r\n")
		}

		if strLen < 0 {
			// RESP null bulk string argument in command
			args[i] = ""
			continue
		}

		// Read strLen bytes + 2 bytes for CRLF
		buf := make([]byte, strLen+2)
		if _, err := io.ReadFull(reader, buf); err != nil {
			return nil, err
		}
		if buf[strLen] != '\r' || buf[strLen+1] != '\n' {
			return nil, fmt.Errorf("-ERR protocol error: bulk string not followed by CRLF\r\n")
		}

		args[i] = string(buf[:strLen])
	}

	return args, nil
}

func handleClient(conn net.Conn, store *Store) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	for {
		args, err := readCommand(reader)
		if err != nil {
			if err == io.EOF {
				return
			}
			errMsg := err.Error()
			if strings.HasPrefix(errMsg, "-ERR ") {
				writer.WriteString(errMsg)
				writer.Flush()
			}
			return
		}

		if len(args) == 0 {
			writer.WriteString("-ERR empty command\r\n")
			writer.Flush()
			continue
		}

		cmd := strings.ToUpper(args[0])
		switch cmd {
		case "PING":
			if len(args) != 1 {
				writer.WriteString("-ERR wrong number of arguments for 'ping' command\r\n")
			} else {
				writer.WriteString("+OK\r\n")
			}
		case "SET":
			if len(args) != 3 {
				writer.WriteString("-ERR wrong number of arguments for 'set' command\r\n")
			} else {
				store.Set(args[1], args[2])
				writer.WriteString("+OK\r\n")
			}
		case "GET":
			if len(args) != 2 {
				writer.WriteString("-ERR wrong number of arguments for 'get' command\r\n")
			} else {
				val, ok := store.Get(args[1])
				if !ok || val == "" {
					writer.WriteString("$-1\r\n")
				} else {
					writer.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(val), val))
				}
			}
		case "DEL":
			if len(args) < 2 {
				writer.WriteString("-ERR wrong number of arguments for 'del' command\r\n")
			} else {
				n := store.Del(args[1:])
				writer.WriteString(fmt.Sprintf(":%d\r\n", n))
			}
		default:
			writer.WriteString(fmt.Sprintf("-ERR unknown command '%s'\r\n", args[0]))
		}
		if err := writer.Flush(); err != nil {
			return
		}
	}
}

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:6379")
	if err != nil {
		fmt.Printf("Failed to bind port: %v\n", err)
		return
	}
	defer listener.Close()

	store := NewStore()

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handleClient(conn, store)
	}
}
