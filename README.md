# Mnh-GoKVStore

A simple Redis-like key-value store implementation in Go with RESP protocol support.

## Features

- **RESP Protocol**: Full Redis Serialization Protocol implementation
- **In-Memory Storage**: Fast in-memory data structure with concurrent access (RWMutex)
- **AOF Persistence**: Append-Only File for durability
- **TTL Support**: Time-to-live expiration for keys
- **Hash Operations**: HSET, HGET, HGETALL
- **Basic Commands**: SET, GET, DEL, PING, EXISTS, TTL

## Supported Commands

### String Operations
- `SET key value [EX seconds]` - Set a key-value pair with optional expiration
- `GET key` - Get the value of a key
- `DEL key [key ...]` - Delete one or more keys
- `EXISTS key [key ...]` - Check if one or more keys exist
- `TTL key` - Get the remaining time to live of a key in seconds

### Hash Operations
- `HSET key field value [field value ...]` - Set hash field(s)
- `HGET key field` - Get hash field value
- `HGETALL key` - Get all fields and values of a hash

### Connection
- `PING` - Returns PONG (keepalive check)

## Installation & Usage

### Run the Server

```bash
cd mnh-go-kv-store/cmd/server
go run main.go
```

The server will listen on `:6379` by default.

### Use the Client

The project includes a client in `pkg/client/`:

```go
client, err := client.NewClient("localhost:6379")
defer client.Close()

// Set a value
client.SET("mykey", "myvalue", 10*time.Second)

// Get a value
value, err := client.GET("mykey")

// Hash operations
client.HSET("user:1", "name", "John")
name, err := client.HGET("user:1", "name")
```

## Architecture

```
mnh-go-kv-store/
├── cmd/
│   └── server/
│       └── main.go          # Server entry point
├── internal/
│   ├── protocol/
│   │   └── resp.go          # RESP protocol implementation
│   └── store/
│       ├── store.go         # In-memory store
│       └── aof.go           # AOF persistence
├── pkg/
│   └── client/
│       └── client.go        # Redis client
└── service/
    ├── server.go            # TCP server
    └── commands_handler.go  # Command handlers
```

## Data Persistence

The server uses AOF (Append-Only File) for persistence. All write commands are logged to `database.aof` and replayed on startup to restore state.

## License

MIT 
