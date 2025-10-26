package store

import (
  "bufio"
  "fmt"
  "io"
  "os"
  "sync"

  "mnhgo/mnh-go-kv-store/internal/protocol"
)

// AOFCommandExecutor là giao diện mà CommandsHandler (trong gói service) phải thực hiện
// Nó được sử dụng để tái tạo trạng thái Store khi đọc file AOF.
type AOFCommandExecutor interface {
  ExecuteAOFCommand(cmdValue protocol.Value) // cmdValue là lệnh đã được parse
}

// AOF struct quản lý file và buffer để ghi dữ liệu AOF
type AOF struct {
  file   *os.File
  mu     sync.Mutex
  writer *bufio.Writer
}

// NewAOF khởi tạo hoặc mở file AOF
func NewAOF(path string) (*AOF, error) {
  // os.O_APPEND|os.O_CREATE|os.O_WRONLY: Mở, ghi, tạo nếu không tồn tại, ghi tiếp vào cuối
  f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
  if err != nil {
    return nil, err
  }

  aof := &AOF{
    file:   f,
    writer: bufio.NewWriter(f),
  }
  return aof, nil
}

// WriteCommand ghi một lệnh RESP đã mã hóa vào file AOF.
func (a *AOF) WriteCommand(cmd []byte) error {
  a.mu.Lock()
  defer a.mu.Unlock()

  _, err := a.writer.Write(cmd)
  if err != nil {
    return err
  }

  // Flush dữ liệu từ buffer ra đĩa.
  return a.writer.Flush()
}

// ReadAndLoad đọc file AOF khi khởi động server để tái tạo trạng thái Store.
func (a *AOF) ReadAndLoad(executor AOFCommandExecutor) error {
  // 1. Đóng file hiện tại và mở lại ở chế độ đọc
  if err := a.file.Close(); err != nil {
    return err
  }
  f, err := os.OpenFile(a.file.Name(), os.O_RDONLY, 0644)
  if err != nil {
    return err
  }
  defer f.Close()

  // 2. Khởi tạo RESP Reader từ file để tái sử dụng logic parsing
  respReader := protocol.NewResp(f)

  // 3. Vòng lặp đọc và thực thi lệnh từ file AOF
  for {
    // Dùng Resp.Read() để đọc và parse một lệnh RESP hoàn chỉnh
    cmdValue, _, err := respReader.Read()

    if err != nil {
      if err == io.EOF {
        break // Đã đọc hết file
      }
      // Báo cáo lỗi parsing trong file AOF và dừng tải
      return fmt.Errorf("error parsing AOF command: %v", err)
    }

    executor.ExecuteAOFCommand(cmdValue)
  }

  // 5. Sau khi đọc xong, mở lại file ở chế độ ghi/append
  f, err = os.OpenFile(a.file.Name(), os.O_APPEND|os.O_WRONLY, 0644)
  if err != nil {
    return err
  }
  a.file = f
  a.writer = bufio.NewWriter(f)

  return nil
}
