package container

import (
  "fmt"
  "syscall"
)

// setupRootFS thay đổi Root Filesystem của container thành rootFsPath
// Đây là bước quan trọng để cô lập hệ thống file.
func setupRootFS(rootFsPath string) error {
  // 1. Thay đổi thư mục gốc bằng chroot (cơ chế đơn giản hơn)
  // chroot chỉ thay đổi thư mục gốc của tiến trình hiện tại,
  // nhưng vẫn còn các điểm mount của hệ thống cũ.
  if err := syscall.Chroot(rootFsPath); err != nil {
    return fmt.Errorf("chroot failed: %w", err)
  }

  // 2. Thay đổi thư mục làm việc hiện tại thành thư mục gốc mới
  if err := syscall.Chdir("/"); err != nil {
    return fmt.Errorf("chdir failed: %w", err)
  }

  // 3. (Nâng cao: Dùng Mounts) Mount /proc
  // Cần mount /proc mới trong container để có thông tin về các PID mới
  // Nếu không, container sẽ thấy PID của host.
  if err := syscall.Mount("proc", "/proc", "proc", 0, ""); err != nil {
    // Bỏ qua lỗi nếu /proc đã mount, nhưng cần thiết cho PID isolation
    fmt.Printf("Warning: Failed to mount /proc: %v\n", err)
  }

  // Redis/Docker thường dùng pivot_root cho cô lập tốt hơn,
  // nhưng chroot đơn giản hơn để bắt đầu.

  return nil
}

/*
// Nếu muốn dùng pivot_root (cách làm của Docker, phức tạp hơn):
func pivotRoot(newRoot string) error {
    // ... logic pivot_root phức tạp hơn chroot ...
}
*/
