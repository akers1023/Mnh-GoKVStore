# Mnh-Go-Load-Balancer

A simple HTTP load balancer implementation in Go with health checking and multiple load balancing algorithms.

## Features

- **Multiple Load Balancing Algorithms**: 
  - Round Robin (default)
  - Least Connections (ready to use)
- **Health Checking**: Automatic health monitoring with configurable intervals
- **Failover**: Automatically skips unhealthy backends
- **Connection Tracking**: Monitors active connections per backend
- **Reverse Proxy**: Uses Go's standard reverse proxy for request forwarding
- **Status Endpoint**: JSON API endpoint to monitor backend health
- **Easy Backend Deployment**: Sample backend server with port configuration

## Architecture

```
mnh-go-load-balancer/
├── cmd/
│   └── lb/
│       ├── main.go              # Load balancer entry point
│       └── temp_backend.go      # Sample backend server
├── internal/
│   ├── backend/
│   │   ├── backend.go           # Backend server representation
│   │   └── pool.go              # Server pool management & health checks
│   ├── balancer/
│   │   ├── interface.go         # Load balancing algorithm interface
│   │   └── roundrobin.go       # Round Robin implementation
│   └── service/
│       └── proxy.go            # Load balancer service (HTTP handler)
└── README.md
```

## Installation & Usage

### Prerequisites

- Go 1.21 or higher

### Run Backend Servers

Start multiple backend servers using the provided sample:

```bash
# Terminal 1: Start backend server on port 8081
cd mnh-go-load-balancer/cmd/lb
export PORT=8081
go run temp_backend.go

# Terminal 2: Start backend server on port 8082
export PORT=8082
go run temp_backend.go

# Terminal 3: Start backend server on port 8083
export PORT=8083
go run temp_backend.go
```

**Windows (PowerShell/Cmd):**
```powershell
# Terminal 1
cd mnh-go-load-balancer/cmd/lb
$env:PORT="8081"
go run temp_backend.go

# Terminal 2
$env:PORT="8082"
go run temp_backend.go

# Terminal 3
$env:PORT="8083"
go run temp_backend.go
```

### Run Load Balancer

```bash
cd mnh-go-load-balancer/cmd/lb
go run main.go
```

The load balancer will start on `:8000` by default.

### Test the Load Balancer

```bash
# Send multiple requests and see how they're distributed
curl http://localhost:8000
curl http://localhost:8000
curl http://localhost:8000
```

You should see requests being distributed across the backend servers.

### Check Backend Status

```bash
# View status of all backends
curl http://localhost:8000/status
```

This returns JSON showing the health and active connections for each backend:

```json
{
  "backends": [
    {
      "url": "http://localhost:8081",
      "status": "UP",
      "active_connections": 2
    },
    {
      "url": "http://localhost:8082",
      "status": "UP",
      "active_connections": 1
    },
    {
      "url": "http://localhost:8083",
      "status": "DOWN",
      "active_connections": 0
    }
  ]
}
```

## Configuration

You can modify the backend URLs and settings in `cmd/lb/main.go`:

```go
// Backend servers
backendURLs := []string{
    "http://localhost:8081",
    "http://localhost:8082",
    "http://localhost:8083",
}

// Load balancer port
LoadBalancerPort = ":8000"

// Health check interval
HealthCheckInterval = 10 * time.Second
```

## How It Works

### Load Balancing Algorithm

Currently implements **Round Robin**:
- Distributes requests sequentially across all healthy backends
- Automatically skips servers that are down
- Thread-safe with atomic counters for concurrent requests

### Health Checking

- Runs in a background goroutine with configurable intervals (default: 10 seconds)
- Uses TCP connection to check if backend is alive
- Logs the health status of each backend
- Automatically routes around dead servers

### Request Flow

1. Client sends request to Load Balancer (port 8000)
2. Load Balancer uses Round Robin to select next healthy backend
3. Request is forwarded via Reverse Proxy
4. Backend processes request and responds
5. Response is forwarded back to client
6. Connection counter is updated

## Switching Load Balancing Algorithms

### Use Least Connections Algorithm

The `LeastConnections` balancer is already implemented. To use it, modify `cmd/lb/main.go`:

```go
// Change from Round Robin:
roundRobinBalancer := balancer.NewRoundRobin(pool)

// To Least Connections:
lcBalancer := balancer.NewLeastConnections(pool)
lbService := service.NewLoadBalancerService(pool, lcBalancer)
```

### Add Your Own Algorithm

Create a new file in `internal/balancer/` implementing the `Balancer` interface:

```go
package balancer

type CustomBalancer struct {
    pool *backend.ServerPool
}

func NewCustomBalancer(pool *backend.ServerPool) *CustomBalancer {
    return &CustomBalancer{pool: pool}
}

func (cb *CustomBalancer) NextBackend() *backend.Backend {
    // Your custom algorithm here
    backends := cb.pool.GetBackends()
    // ... select logic
    return selectedBackend
}
```

Then use it in `main.go`:

```go
customBalancer := balancer.NewCustomBalancer(pool)
lbService := service.NewLoadBalancerService(pool, customBalancer)
```

## Project Structure

- **backend**: Manages individual backend servers and their health status
- **balancer**: Load balancing algorithm implementations
- **service**: HTTP handler that coordinates load balancing logic
- **cmd/lb**: Main entry point and sample backend

## Load Balancing Algorithms

### Round Robin
- Distributes requests sequentially
- Best for servers with similar capacity
- Ensures even distribution over time

### Least Connections
- Selects backend with fewest active connections
- Best for servers with varying processing times
- Automatically balances load based on server load

## Future Enhancements

Potential features to add:
- Weighted Round Robin (configure weights per backend)
- IP Hash algorithm for session affinity
- HTTPS support
- Request/Response modification (headers, etc.)
- Metrics and monitoring endpoints (Prometheus integration)
- Configuration file support (YAML/JSON)
- Dynamic backend registration (add/remove at runtime)
- Circuit breaker pattern
- Rate limiting

## License

MIT

