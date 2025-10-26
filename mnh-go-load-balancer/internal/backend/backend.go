package backend

import (
  "net/http/httputil"
  "net/url"
  "sync"
  "sync/atomic"
)

// Backend đại diện cho một server đích (target server)
type Backend struct {
  URL     *url.URL               // Địa chỉ URL của server (ví dụ: http://localhost:8081)
  Proxy   *httputil.ReverseProxy // Proxy để chuyển tiếp yêu cầu đến server này
  active  uint64                 // Số lượng kết nối đang hoạt động (cho các thuật toán nâng cao)
  mu      sync.RWMutex           // Mutex để bảo vệ trạng thái Alive
  isAlive bool                   // Trạng thái sức khỏe của server
}

// NewBackend tạo một thể hiện Backend mới
func NewBackend(u *url.URL) *Backend {
  return &Backend{
    URL:     u,
    Proxy:   httputil.NewSingleHostReverseProxy(u), // Tạo ReverseProxy sẵn
    isAlive: true,                                  // Mặc định là sống khi khởi tạo
  }
}

// SetAlive cập nhật trạng thái sức khỏe của Backend
func (b *Backend) SetAlive(alive bool) {
  b.mu.Lock()
  b.isAlive = alive
  b.mu.Unlock()
}

// IsAlive kiểm tra trạng thái sức khỏe hiện tại (sử dụng RLock để đọc an toàn)
func (b *Backend) IsAlive() (alive bool) {
  b.mu.RLock()
  alive = b.isAlive
  b.mu.RUnlock()
  return
}

// GetActiveConnections trả về số lượng kết nối đang hoạt động
func (b *Backend) GetActiveConnections() uint64 {
  return atomic.LoadUint64(&b.active)
}

// IncActiveConnections tăng số lượng kết nối đang hoạt động
func (b *Backend) IncActiveConnections() {
  atomic.AddUint64(&b.active, 1)
}

// DecActiveConnections giảm số lượng kết nối đang hoạt động
func (b *Backend) DecActiveConnections() {
  atomic.AddUint64(&b.active, ^uint64(0)) // Giảm đi 1 cách an toàn
}
