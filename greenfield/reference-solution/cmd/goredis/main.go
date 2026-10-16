package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"goredis/internal/command"
	"goredis/internal/server"
	"goredis/internal/store"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	port := flag.Int("port", 6379, "TCP port to listen on")
	flag.Parse()
	if flag.NArg() != 0 || *port < 1 || *port > 65535 {
		return fmt.Errorf("usage: goredis [-port 1..65535]")
	}
	address := fmt.Sprintf("127.0.0.1:%d", *port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", address, err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	log.Printf("GoRedis listening on %s", address)
	return server.Serve(ctx, listener, command.New(store.New()).Execute)
}
