package container

import (
  "fmt"
  "os"
  "os/exec"
  "syscall"
)

// Run là hàm chính để tạo, thiết lập và chạy container
// command: Lệnh sẽ chạy bên trong container (ví dụ: ["/bin/sh"])
// rootFsPath: Đường dẫn đến root filesystem (chứa bin, lib,...)
func Run(command []string, rootFsPath string) error {
  // 1. Chuẩn bị Process
  // Sử dụng os/exec để tạo một command object cho tiến trình con.
  // Mục đích: Tiến trình con sẽ tự gọi lại chính file thực thi này (main.go)
  // với một flag đặc biệt (ví dụ: "child") để biết nó đang ở trong container.
  cmd := exec.Command("/proc/self/exe", append([]string{"child"}, command...)...)

  // 2. Thiết lập Namespaces (Cơ chế cô lập cốt lõi)
  // CLONE_NEWUTS: Cô lập hostname
  // CLONE_NEWPID: Cô lập Process IDs
  // CLONE_NEWNS: Cô lập Mount points (filesystem)
  cmd.SysProcAttr = &syscall.SysProcAttr{
    Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
  }

  // 3. Thiết lập I/O
  // Kết nối stdin, stdout, stderr của container với terminal hiện tại
  cmd.Stdin = os.Stdin
  cmd.Stdout = os.Stdout
  cmd.Stderr = os.Stderr

  // 4. Chạy và Chờ đợi
  if err := cmd.Run(); err != nil {
    return fmt.Errorf("failed to run command in parent process: %w", err)
  }

  return nil
}

// ChildProcess là hàm được tiến trình con gọi sau khi đã ở trong Namespaces mới.
// Nó thực hiện các bước thiết lập cuối cùng bên trong container.
func ChildProcess(command []string, rootFsPath string) error {
  // 1. Thiết lập Hostname (Đã được cô lập bởi NEWUTS)
  if err := syscall.Sethostname([]byte("container-box")); err != nil {
    return fmt.Errorf("failed to set hostname: %w", err)
  }

  // 2. Thiết lập Root Filesystem (Đã được cô lập bởi NEWNS)
  if err := setupRootFS(rootFsPath); err != nil {
    return fmt.Errorf("failed to set root filesystem: %w", err)
  }

  // TODO: Áp dụng Cgroups (giới hạn tài nguyên) ở đây

  // 3. Thực thi Lệnh
  // Sử dụng syscall.Exec để thay thế tiến trình hiện tại bằng lệnh của người dùng
  // Tuyệt đối không gọi return sau khi Exec thành công!
  if err := syscall.Exec(command[0], command, os.Environ()); err != nil {
    return fmt.Errorf("failed to execute command %s: %w", command[0], err)
  }

  // Không thể xảy ra nếu Exec thành công.
  return nil
}

// TODO: Triển khai Cgroup Apply/Cleanup
/*
func ApplyCgroups(...) { ... }
func CleanupCgroups(...) { ... }
*/
