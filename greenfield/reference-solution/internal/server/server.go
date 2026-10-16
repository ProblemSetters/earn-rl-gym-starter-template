// Package server owns TCP connections and dispatches RESP requests.
package server

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"sync"

	"goredis/internal/resp"
)

// Serve accepts concurrent clients until cancellation or a listener error.
// It owns the listener and closes all active connections before returning.
func Serve(ctx context.Context, listener net.Listener, execute func([]string) []byte) error {
	ctx, cancel := context.WithCancel(ctx)
	var clients sync.WaitGroup
	defer func() {
		cancel()
		listener.Close()
		clients.Wait()
	}()
	stop := context.AfterFunc(ctx, func() { listener.Close() })
	defer stop()
	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		clients.Add(1)
		go func() {
			defer clients.Done()
			serve(ctx, conn, execute)
		}()
	}
}

func serve(ctx context.Context, conn net.Conn, execute func([]string) []byte) {
	defer conn.Close()
	stop := context.AfterFunc(ctx, func() { conn.Close() })
	defer stop()
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	for {
		args, err := resp.ReadCommand(reader)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				writer.Write(resp.Error("invalid request"))
				writer.Flush()
			}
			return
		}
		if _, err := writer.Write(execute(args)); err != nil {
			return
		}
		if err := writer.Flush(); err != nil {
			return
		}
	}
}
