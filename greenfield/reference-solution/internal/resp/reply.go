package resp

import "fmt"

func SimpleString(value string) []byte { return []byte("+" + value + "\r\n") }
func Error(message string) []byte      { return []byte("-ERR " + message + "\r\n") }
func BulkString(value string) []byte {
	return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(value), value))
}
func NullBulkString() []byte   { return []byte("$-1\r\n") }
func Integer(value int) []byte { return []byte(fmt.Sprintf(":%d\r\n", value)) }
