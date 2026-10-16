package resp

import (
	"bufio"
	"errors"
	"io"
	"strconv"
	"strings"
)

const (
	maxArrayLength      = 1_000_000
	maxBulkStringLength = 64 * 1024 * 1024
)

func readLine(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	if len(line) < 2 || !strings.HasSuffix(line, "\r\n") {
		return "", errors.New("invalid RESP line")
	}
	return line[:len(line)-2], nil
}

// ReadCommand decodes one RESP2 array of non-null bulk strings.
func ReadCommand(r *bufio.Reader) ([]string, error) {
	prefix, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	if prefix != '*' {
		return nil, errors.New("expected RESP array")
	}
	line, err := readLine(r)
	if err != nil {
		return nil, err
	}
	count, err := strconv.Atoi(line)
	if err != nil || count < 1 || count > maxArrayLength {
		return nil, errors.New("invalid array length")
	}
	args := make([]string, 0, count)
	for i := 0; i < count; i++ {
		prefix, err = r.ReadByte()
		if err != nil {
			return nil, err
		}
		if prefix != '$' {
			return nil, errors.New("expected bulk string")
		}
		line, err = readLine(r)
		if err != nil {
			return nil, err
		}
		length, parseErr := strconv.Atoi(line)
		if parseErr != nil || length < 0 || length > maxBulkStringLength {
			return nil, errors.New("invalid bulk string length")
		}
		value := make([]byte, length+2)
		if _, err := io.ReadFull(r, value); err != nil {
			return nil, err
		}
		if value[length] != '\r' || value[length+1] != '\n' {
			return nil, errors.New("invalid bulk string terminator")
		}
		args = append(args, string(value[:length]))
	}
	return args, nil
}
