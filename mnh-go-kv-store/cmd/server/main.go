package main

import (
  "log"

  "mnhgo/mnh-go-kv-store/internal/store"
  "mnhgo/mnh-go-kv-store/service"
)

func main() {
  // Khởi tạo Store và AOF
  myStore := store.NewStore()
  myAOF, err := store.NewAOF("database.aof")
  if err != nil {
    log.Fatalf("Failed to initialize AOF: %v", err)
  }

  // Khởi tạo CommandsHandler
  handler := service.NewCommandsHandler(myStore, myAOF)

  // Tải lại dữ liệu từ AOF khi khởi động
  if err := myAOF.ReadAndLoad(handler); err != nil {
    log.Printf("Warning: Failed to load AOF data: %v", err)
  }

  // Khởi động Server
  server := service.NewServer(handler)
  if err := server.Start(":6379"); err != nil {
    log.Fatalf("Server failed to start: %v", err)
  }
}
