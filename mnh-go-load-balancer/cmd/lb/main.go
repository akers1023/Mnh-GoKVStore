package main

import (
  "log"
  "net/http"
  "net/url"
  "time"

  "mnhgo/mnh-go-load-balancer/internal/backend"
  "mnhgo/mnh-go-load-balancer/internal/balancer"
  "mnhgo/mnh-go-load-balancer/service"
)

// Các hằng số cấu hình
const (
  LoadBalancerPort    = ":8000"
  HealthCheckInterval = 10 * time.Second
)

func main() {
  // 1. DANH SÁCH CÁC BACKEND SERVER
  // Giả định bạn có 3 server backend đang chạy trên các cổng khác nhau
  backendURLs := []string{
    "http://localhost:8081",
    "http://localhost:8082",
    "http://localhost:8083",
  }

  // 2. KHỞI TẠO SERVER POOL
  pool := backend.NewServerPool()

  for _, rawURL := range backendURLs {
    serverURL, err := url.Parse(rawURL)
    if err != nil {
      log.Fatalf("Invalid backend URL: %s", rawURL)
    }
    pool.AddBackend(serverURL)
  }

  // 3. KHỞI TẠO HEALTH CHECK
  // Bắt đầu Goroutine chạy Health Check định kỳ
  log.Printf("Starting health check with interval %s...", HealthCheckInterval)
  pool.HealthCheck(HealthCheckInterval)

  // 4. KHỞI TẠO BALANCER VÀ DỊCH VỤ
  // Sử dụng thuật toán Round Robin
  roundRobinBalancer := balancer.NewRoundRobin(pool)

  // Tạo dịch vụ Load Balancer (là HTTP Handler chính)
  lbService := service.NewLoadBalancerService(pool, roundRobinBalancer)

  // 5. CHẠY LOAD BALANCER
  server := http.Server{
    Addr:    LoadBalancerPort,
    Handler: lbService, // Sử dụng LoadBalancerService làm Handler
  }

  log.Printf("Load Balancer starting on %s...", LoadBalancerPort)
  if err := server.ListenAndServe(); err != nil {
    log.Fatalf("Load Balancer failed to start: %v", err)
  }
}
