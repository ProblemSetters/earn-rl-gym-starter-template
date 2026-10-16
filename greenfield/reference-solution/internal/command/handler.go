// Package command validates commands and applies them to the shared store.
package command

import (
	"strings"

	"goredis/internal/resp"
	"goredis/internal/store"
)

type Handler struct{ store *store.Store }

func New(values *store.Store) *Handler { return &Handler{store: values} }

func (h *Handler) Execute(args []string) []byte {
	if len(args) == 0 {
		return resp.Error("empty command")
	}
	switch strings.ToUpper(args[0]) {
	case "PING":
		if len(args) != 1 {
			return arityError("ping")
		}
		return resp.SimpleString("PONG")
	case "SET":
		if len(args) != 3 {
			return arityError("set")
		}
		h.store.Set(args[1], args[2])
		return resp.SimpleString("OK")
	case "GET":
		if len(args) != 2 {
			return arityError("get")
		}
		value, ok := h.store.Get(args[1])
		if !ok {
			return resp.NullBulkString()
		}
		return resp.BulkString(value)
	case "DEL":
		if len(args) < 2 {
			return arityError("del")
		}
		return resp.Integer(h.store.Delete(args[1:]...))
	default:
		return resp.Error("unknown command")
	}
}

func arityError(name string) []byte {
	return resp.Error("wrong number of arguments for '" + name + "' command")
}
