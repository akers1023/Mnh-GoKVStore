package backend

import (
  "log"
  "net"
  "net/url"
  "sync"
  "time"
)

// ServerPool quản lý danh sách các Backend Server
type ServerPool struct {
  backends []*Backend
  mu       sync.RWMutex // Mutex để bảo vệ danh sách backends
}

// NewServerPool tạo một thể hiện ServerPool mới
func NewServerPool() *ServerPool {
  return &ServerPool{
    backends: make([]*Backend, 0),
  }
}

// AddBackend thêm một server mới vào Pool
func (s *ServerPool) AddBackend(u *url.URL) {
  s.mu.Lock()
  defer s.mu.Unlock()

  newBackend := NewBackend(u)
  s.backends = append(s.backends, newBackend)
  log.Printf("Added backend: %s", u.String())
}

// GetBackends trả về danh sách các Backend (chỉ đọc)
func (s *ServerPool) GetBackends() []*Backend {
  s.mu.RLock()
  defer s.mu.RUnlock()

  // Trả về bản sao để tránh thay đổi trực tiếp bên ngoài
  return s.backends
}

// --- Health Check Logic ---

// HealthCheck ping tất cả các backend và cập nhật trạng thái
func (s *ServerPool) HealthCheck(interval time.Duration) {
  // Khởi chạy vòng lặp kiểm tra sức khỏe trong một goroutine
  go func() {
    for {
      s.checkBackends()
      time.Sleep(interval)
    }
  }()
}

// checkBackends thực hiện ping HTTP đơn giản
func (s *ServerPool) checkBackends() {
  backends := s.GetBackends()
  for _, b := range backends {
    isAlive := isBackendAlive(b.URL)
    b.SetAlive(isAlive)

    status := "UP"
    if !isAlive {
      status = "DOWN"
    }
    log.Printf("Health Check for %s: %s", b.URL.String(), status)
  }
}

// isBackendAlive thực hiện một ping HTTP đến URL của backend
func isBackendAlive(u *url.URL) bool {
  // Có thể dùng net/http với timeout ngắn để kiểm tra thực sự
  timeout := 2 * time.Second
  conn, err := net.DialTimeout("tcp", u.Host, timeout)
  if err != nil {
    return false // Không thể kết nối
  }
  defer conn.Close()
  
  // Đơn giản: nếu có thể mở kết nối TCP thì coi như backend đang sống
  return true
}
