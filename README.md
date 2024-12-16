# OTLP Log Parser (Go)

Perfroms calculations based on log records attributes

## Usage

Build the application:
```shell
go build ./...
```

Run the application:
```shell
go run ./...
```

Run tests
```shell
go test ./...
```

## CLI attributes

- <b>attrubute</b> - string. The name of attribute we're looking for
- <b>duration</b> - duration. The duration of window in seconds (default 30s)
- <b>listenAddr</b> - string The listen address (default "localhost:4317")
- <b>maxReceiveMessageSize</b> - int The max message size in bytes the server can receive (default 16777216)

## Example
Build app:
```shell
go build .
```
Then run:
```Shell
./otlp-log-processor-backend -duration 10s
```
