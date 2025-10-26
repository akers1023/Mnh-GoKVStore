package service

import (
  "fmt"
  "log"
  "net/http"

  "mnhgo/mnh-go-load-balancer/internal/backend"
  "mnhgo/mnh-go-load-balancer/internal/balancer"
)

// LoadBalancerService là dịch vụ chính xử lý yêu cầu HTTP
type LoadBalancerService struct {
  pool     *backend.ServerPool // Danh sách và trạng thái các Backend
  balancer balancer.Balancer   // Thuật toán cân bằng tải (ví dụ: Round Robin)
}

// NewLoadBalancerService tạo một instance của LoadBalancerService
func NewLoadBalancerService(p *backend.ServerPool, b balancer.Balancer) *LoadBalancerService {
  return &LoadBalancerService{
    pool:     p,
    balancer: b,
  }
}

// ServeHTTP là phương thức cốt lõi, implement http.Handler interface.
// Nó được gọi cho mỗi yêu cầu HTTP đến Load Balancer.
func (lbs *LoadBalancerService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  // Kiểm tra nếu là request đến /status endpoint
  if r.URL.Path == "/status" && r.Method == "GET" {
    lbs.handleStatus(w, r)
    return
  }

  // 1. CHỌN BACKEND: Sử dụng thuật toán Balancer để chọn server tiếp theo
  selectedBackend := lbs.balancer.NextBackend()

  if selectedBackend == nil {
    // 2. XỬ LÝ LỖI: Nếu không tìm thấy backend nào đang UP
    http.Error(w, "Service Unavailable: No healthy backend found.", http.StatusServiceUnavailable)
    log.Println("Request rejected: No healthy backend available.")
    return
  }

  // 3. TĂNG KẾT NỐI: Tăng bộ đếm kết nối đang hoạt động của backend được chọn
  selectedBackend.IncActiveConnections()
  defer selectedBackend.DecActiveConnections() // Đảm bảo giảm khi request kết thúc

  log.Printf(
    "Request forwarded to: %s (Active connections: %d)",
    selectedBackend.URL.String(),
    selectedBackend.GetActiveConnections(),
  )

  // 4. THIẾT LẬP CONTEXT VÀ CHUYỂN TIẾP (Reverse Proxy)
  // Proxy chính thức chuyển tiếp yêu cầu và chuyển phản hồi ngược lại
  selectedBackend.Proxy.ServeHTTP(w, r)
}

// handleStatus hiển thị trạng thái của Load Balancer
func (lbs *LoadBalancerService) handleStatus(w http.ResponseWriter, r *http.Request) {
  backends := lbs.pool.GetBackends()
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)

  json := "{\n  \"backends\": [\n"
  for i, backend := range backends {
    status := "UP"
    if !backend.IsAlive() {
      status = "DOWN"
    }
    json += "    {\n"
    json += fmt.Sprintf("      \"url\": \"%s\",\n", backend.URL.String())
    json += fmt.Sprintf("      \"status\": \"%s\",\n", status)
    json += fmt.Sprintf("      \"active_connections\": %d\n", backend.GetActiveConnections())
    json += "    }"
    if i < len(backends)-1 {
      json += ","
    }
    json += "\n"
  }
  json += "  ]\n}"

  w.Write([]byte(json))
}
