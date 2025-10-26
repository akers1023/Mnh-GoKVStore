package balancer

import (
  "log"
  "sync/atomic"

  "mnhgo/mnh-go-load-balancer/internal/backend"
)

// RoundRobin là struct triển khai thuật toán Round Robin
type RoundRobin struct {
  pool *backend.ServerPool // Tham chiếu đến ServerPool để lấy danh sách backends
  // current là bộ đếm nguyên tử (atomic) để đảm bảo an toàn truy cập đồng thời
  current uint64
}

// NewRoundRobin tạo một thể hiện RoundRobin mới
func NewRoundRobin(pool *backend.ServerPool) *RoundRobin {
  return &RoundRobin{
    pool: pool,
  }
}

// NextBackend triển khai phương thức Balancer.NextBackend
func (r *RoundRobin) NextBackend() *backend.Backend {
  backends := r.pool.GetBackends()
  if len(backends) == 0 {
    return nil
  }

  // Lặp tối đa số lượng server để tìm server đang UP
  for i := 0; i < len(backends); i++ {
    // Tăng bộ đếm và lấy chỉ mục server tiếp theo
    index := atomic.AddUint64(&r.current, 1) % uint64(len(backends))

    backend := backends[index]

    // Kiểm tra trạng thái sống
    if backend.IsAlive() {
      return backend
    }

    log.Printf("Backend %s is DOWN, skipping...", backend.URL.String())
  }

  // Nếu lặp qua tất cả mà không tìm thấy server nào đang UP
  log.Println("No healthy backend found.")
  return nil
}
