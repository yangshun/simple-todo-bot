# Simple Todo Bot

A telegram bot that manages your todos:

<img src="screenshots/example.png" alt="Screenshot" width="75%"/>

### Set up:

1. Create a Postgres database and run `todos.sql`.
2. Copy `config.go.example` as `config.go` and replace the respective values in `config.go`.
```
$ cp config/config.go.example config/config.go
```

3. Run the go program:

```
$ go build index.go && ./index
```
