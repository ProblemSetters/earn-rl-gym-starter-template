# Build a small data store in Go

Write a Go program that stores named text values while it runs. Use only packages included with Go. Accept connections at `127.0.0.1:6379`, and allow several people or programs to use it at once.

## Requests and replies

Each request starts with `*<number of words>\r\n`. Each word then appears as `$<number of bytes>\r\n<word>\r\n`. Here, `\r\n` means a carriage return followed by a newline. A connection may carry several requests. Command names ignore letter case.

For example, `PING` is sent as `*1\r\n$4\r\nPING\r\n`. Send these replies, including the ending `\r\n`:

| Request | Reply |
| --- | --- |
| `PING` | `+PONG\r\n` |
| `SET key value` | Save or replace the value; `+OK\r\n` |
| `GET key` | If found, `$<number of bytes>\r\n<value>\r\n`; otherwise, `$-1\r\n` |
| `DEL key [key ...]` | Remove the named values; `:<number removed>\r\n` |

An empty saved value must return `$0\r\n\r\n`. For `DEL`, count each existing name only once, even if it appears more than once. If a command has the wrong number of words, reply with `-ERR ` followed by an explanation and `\r\n`. Keep accepting requests on that connection afterward. Saved values last until the program stops.

## Running the program

The project root contains `config.toml` with setup and run commands. Update those commands if your implementation needs different ones.
