package server

import (
	"context"
	"io"
	"net"
	"testing"
	"time"

	"goredis/internal/command"
	"goredis/internal/store"
)

func TestConnectionKeepsWorkingAfterCommandError(t *testing.T) {
	server, client := net.Pipe()
	defer client.Close()
	client.SetDeadline(time.Now().Add(2 * time.Second))
	done := make(chan struct{})
	go func() { defer close(done); serve(context.Background(), server, command.New(store.New()).Execute) }()
	defer func() { client.Close(); <-done }()
	go io.WriteString(client, "*1\r\n$3\r\nGET\r\n*1\r\n$4\r\nPING\r\n")
	want := "-ERR wrong number of arguments for 'get' command\r\n+PONG\r\n"
	got := make([]byte, len(want))
	if _, err := io.ReadFull(client, got); err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestMalformedRequestClosesOnlyItsConnection(t *testing.T) {
	server, client := net.Pipe()
	defer client.Close()
	client.SetDeadline(time.Now().Add(2 * time.Second))
	go serve(context.Background(), server, command.New(store.New()).Execute)
	go io.WriteString(client, "bad request\r\n")
	got, err := io.ReadAll(client)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "-ERR invalid request\r\n" {
		t.Fatalf("got %q", got)
	}
}

func TestServeSharesStoreAndClosesIdleClientsOnCancel(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- Serve(ctx, listener, command.New(store.New()).Execute) }()
	defer func() {
		cancel()
		select {
		case err := <-done:
			if err != nil {
				t.Errorf("Serve: %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Error("Serve did not stop after cancellation")
		}
	}()
	first, err := net.DialTimeout("tcp", listener.Addr().String(), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	second, err := net.DialTimeout("tcp", listener.Addr().String(), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	for _, conn := range []net.Conn{first, second} {
		conn.SetDeadline(time.Now().Add(2 * time.Second))
	}
	exchange := func(conn net.Conn, request, want string) {
		t.Helper()
		if _, err := io.WriteString(conn, request); err != nil {
			t.Fatal(err)
		}
		got := make([]byte, len(want))
		if _, err := io.ReadFull(conn, got); err != nil {
			t.Fatal(err)
		}
		if string(got) != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	}
	exchange(first, "*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n", "+OK\r\n")
	exchange(second, "*2\r\n$3\r\nGET\r\n$3\r\nkey\r\n", "$5\r\nvalue\r\n")
	cancel()
	for _, conn := range []net.Conn{first, second} {
		if _, err := io.ReadAll(conn); err != nil {
			t.Errorf("idle connection did not close cleanly: %v", err)
		}
	}
}
