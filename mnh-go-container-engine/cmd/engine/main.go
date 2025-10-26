package main

import (
  "fmt"
  "log"
  "os"

  "mnhgo/mnh-go-container-engine/internal/container"
)

// RootFS Path: Giả định thư mục gốc của container nằm ở đây
// Bạn cần chuẩn bị một thư mục có cấu trúc như Linux (bin, lib, usr, proc)
const DefaultRootFS = "/path/to/your/rootfs"

func main() {
  if len(os.Args) < 2 {
    showHelp()
    os.Exit(1)
  }

  // Kiểm tra xem đây là tiến trình Parent hay Child
  switch os.Args[1] {
  case "run":
    // Tiến trình Parent: Tạo môi trường cô lập
    run()
  case "child":
    // Tiến trình Child: Thiết lập môi trường và chạy lệnh người dùng
    child()
  case "help", "--help", "-h":
    showHelp()
    os.Exit(0)
  default:
    log.Printf("Unknown command: %s\n", os.Args[1])
    showHelp()
    os.Exit(1)
  }
}

func showHelp() {
  fmt.Println(`
Mnh-Go-Container-Engine - A minimal container runtime

USAGE:
  go run main.go run <command>

EXAMPLES:
  go run main.go run /bin/sh
  go run main.go run /bin/ls
  go run main.go run /bin/echo "Hello from container!"

IMPORTANT NOTES:
  - This requires Linux kernel features (namespaces, cgroups)
  - On Windows, install WSL2 to run this
  - On Mac, use a Linux VM or Docker Desktop
  - Must run as root/sudo for namespace operations

CURRENT CONFIGURATION:
  RootFS: DefaultRootFS

To change the root filesystem path, edit DefaultRootFS in main.go

For more information, see README.md
`)
}

// Parent Process (Gọi run)
func run() {
  log.Printf("Starting Parent Process: %v", os.Args)

  // Lệnh sẽ chạy bên trong container (ví dụ: /bin/sh)
  // Bắt đầu từ os.Args[2:] để bỏ qua "program_name" và "run"
  command := os.Args[2:]
  
  if len(command) == 0 {
    log.Fatalf("No command specified. Usage: go run main.go run <command>")
  }

  if err := container.Run(command, DefaultRootFS); err != nil {
    log.Fatalf("Run failed: %v\nHint: Make sure rootfs path is correct and you have sudo privileges", err)
  }
}

// Child Process (Gọi child)
func child() {
  log.Printf("Starting Child Process: %v", os.Args)

  // Lệnh sẽ chạy (Bỏ qua "program_name" và "child")
  command := os.Args[2:]

  if err := container.ChildProcess(command, DefaultRootFS); err != nil {
    // Log.Fatal sẽ exit process, điều này tốt vì Exec thất bại là lỗi nghiêm trọng
    log.Fatalf("Child process failed: %v", err)
  }
}
