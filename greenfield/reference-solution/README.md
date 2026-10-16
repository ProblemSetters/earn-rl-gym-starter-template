# GoRedis reference solution

This Go program stores named values while it runs and accepts requests at `127.0.0.1:6379`. It implements the behavior described in `../instruction.md`.

## Build and run

Requires Go 1.22 or newer. From this directory:

```sh
go build -o goredis ./cmd/goredis
./goredis
```

The included `config.toml` contains the setup and run commands. Use `./goredis -port 6380` to choose another port.

## Check the implementation

```sh
go test ./...
```

The task's separate behavior checks are in `../verifier/tests/`.
