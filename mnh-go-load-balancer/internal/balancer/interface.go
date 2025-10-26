package balancer

import "mnhgo/mnh-go-load-balancer/internal/backend"

// Balancer định nghĩa giao diện cho bất kỳ thuật toán cân bằng tải nào
type Balancer interface {
  // NextBackend chọn và trả về Backend Server tiếp theo dựa trên thuật toán
  NextBackend() *backend.Backend
}
