package resp

import (
	"bufio"
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestReadBinaryAndPipelinedCommands(t *testing.T) {
	r := bufio.NewReaderSize(strings.NewReader("*3\r\n$3\r\nSET\r\n$0\r\n\r\n$6\r\na\x00\r\nβ\r\n*1\r\n$4\r\nPING\r\n"), 16)
	for _, want := range [][]string{{"SET", "", "a\x00\r\nβ"}, {"PING"}} {
		got, err := ReadCommand(r)
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Fatalf("readCommand = %q, %v; want %q", got, err, want)
		}
	}
	if _, err := ReadCommand(r); err != io.EOF {
		t.Fatalf("expected EOF, got %v", err)
	}
}

func TestRejectMalformedFrames(t *testing.T) {
	for _, input := range []string{
		"PING\r\n", "*1\n", "*-1\r\n", "*0\r\n", "*1000001\r\n", "*x\r\n",
		"*1\r\n+PING\r\n", "*1\r\n$-1\r\n", "*1\r\n$67108865\r\n", "*1\r\n$x\r\n",
		"*1\r\n$4\r\nPI", "*1\r\n$4\r\nPINGxx",
	} {
		if _, err := ReadCommand(bufio.NewReader(strings.NewReader(input))); err == nil {
			t.Errorf("accepted malformed frame %q", input)
		}
	}
}
