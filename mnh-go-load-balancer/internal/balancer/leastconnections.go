package balancer

import (
  "log"
  "sync"

  "mnhgo/mnh-go-load-balancer/internal/backend"
)

// LeastConnections triển khai thuật toán Load Balancing dựa trên số lượng kết nối
// Server có ít kết nối nhất sẽ được chọn để xử lý request tiếp theo
type LeastConnections struct {
  pool *backend.ServerPool
  mu   sync.Mutex // Để đảm bảo thread-safe khi truy cập
}

// NewLeastConnections tạo một instance mới của LeastConnections balancer
func NewLeastConnections(pool *backend.ServerPool) *LeastConnections {
  return &LeastConnections{
    pool: pool,
  }
}

// NextBackend trả về server có ít kết nối đang hoạt động nhất
func (lc *LeastConnections) NextBackend() *backend.Backend {
  backends := lc.pool.GetBackends()
  if len(backends) == 0 {
    return nil
  }

  var selectedBackend *backend.Backend
  var minConnections uint64 = ^uint64(0) // Max value

  lc.mu.Lock()
  defer lc.mu.Unlock()

  // Tìm backend đang sống và có ít kết nối nhất
  for _, backend := range backends {
    if backend.IsAlive() {
      connections := backend.GetActiveConnections()
      if connections < minConnections {
        minConnections = connections
        selectedBackend = backend
      }
    }
  }

  if selectedBackend == nil {
    log.Println("No healthy backend found.")
    return nil
  }

  log.Printf(
    "Selected %s (Active connections: %d)",
    selectedBackend.URL.String(),
    selectedBackend.GetActiveConnections(),
  )

  return selectedBackend
}

