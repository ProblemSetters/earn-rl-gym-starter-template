package command

import (
	"fmt"
	"sync"
	"testing"

	"goredis/internal/store"
)

func TestCommands(t *testing.T) {
	c := New(store.New())
	cases := []struct {
		args []string
		want string
	}{
		{nil, "-ERR empty command\r\n"},
		{[]string{"pInG"}, "+PONG\r\n"},
		{[]string{"PING", "extra"}, "-ERR wrong number of arguments for 'ping' command\r\n"},
		{[]string{"SET", "key"}, "-ERR wrong number of arguments for 'set' command\r\n"},
		{[]string{"GET"}, "-ERR wrong number of arguments for 'get' command\r\n"},
		{[]string{"DEL"}, "-ERR wrong number of arguments for 'del' command\r\n"},
		{[]string{"UNKNOWN"}, "-ERR unknown command\r\n"},
		{[]string{"GET", "missing"}, "$-1\r\n"},
		{[]string{"set", "key", "hello"}, "+OK\r\n"},
		{[]string{"GET", "key"}, "$5\r\nhello\r\n"},
		{[]string{"SET", "key", "a\x00\r\nβ"}, "+OK\r\n"},
		{[]string{"GET", "key"}, "$6\r\na\x00\r\nβ\r\n"},
		{[]string{"SET", "empty", ""}, "+OK\r\n"},
		{[]string{"GET", "empty"}, "$0\r\n\r\n"},
		{[]string{"DEL", "key", "missing", "key", "empty"}, ":2\r\n"},
		{[]string{"GET", "key"}, "$-1\r\n"},
		{[]string{"GET", "empty"}, "$-1\r\n"},
		{[]string{"DEL", "key"}, ":0\r\n"},
	}
	for _, tc := range cases {
		if got := string(c.Execute(tc.args)); got != tc.want {
			t.Errorf("execute(%q) = %q, want %q", tc.args, got, tc.want)
		}
	}
}

func TestConcurrentClientsShareStore(t *testing.T) {
	c := New(store.New())
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", i)
			c.Execute([]string{"SET", key, "value"})
			if got := string(c.Execute([]string{"GET", key})); got != "$5\r\nvalue\r\n" {
				t.Errorf("GET %s = %q", key, got)
			}
			if got := string(c.Execute([]string{"DEL", key, key})); got != ":1\r\n" {
				t.Errorf("DEL %s = %q", key, got)
			}
		}(i)
	}
	wg.Wait()
}
