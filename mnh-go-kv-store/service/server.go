package service

import (
  "fmt"
  "io"
  "log"
  "net"

  "mnhgo/mnh-go-kv-store/internal/protocol"
)

const DefaultPort = ":6379"

// Server chứa các thành phần mạng và logic xử lý lệnh
type Server struct {
  listener net.Listener
  handler  *CommandsHandler // Tham chiếu đến bộ xử lý lệnh
}

func NewServer(handler *CommandsHandler) *Server {
  return &Server{
    handler: handler,
  }
}

// Start khởi động Listener và bắt đầu vòng lặp chấp nhận kết nối
func (s *Server) Start(addr string) error {
  if addr == "" {
    addr = DefaultPort
  }

  var err error
  s.listener, err = net.Listen("tcp", addr)
  if err != nil {
    return fmt.Errorf("failed to listen on %s: %w", addr, err)
  }

  log.Printf("KV Store listening on %s", addr)
  s.acceptLoop()

  return nil
}

// acceptLoop là vòng lặp chính chấp nhận kết nối và khởi tạo Goroutine xử lý
func (s *Server) acceptLoop() {
  for {
    conn, err := s.listener.Accept()
    if err != nil {
      log.Printf("Error accepting connection: %v", err)
      continue
    }
    // Xử lý mỗi kết nối trong một Goroutine riêng biệt
    go s.handleConn(conn)
  }
}

// handleConn xử lý một kết nối client duy nhất
func (s *Server) handleConn(conn net.Conn) {
  defer conn.Close()

  log.Printf("New connection from %s", conn.RemoteAddr())

  resp := protocol.NewResp(conn)

  // Vòng lặp để đọc lệnh liên tục từ client
  for {
    // 1. Đọc lệnh từ client (RESP format)
    cmdValue, _, err := resp.Read()

    if err != nil {
      if err == io.EOF {
        log.Printf("Connection closed by client: %s", conn.RemoteAddr())
        return
      }
      log.Printf("Error reading command from %s: %v", conn.RemoteAddr(), err)

      // Gửi phản hồi lỗi giao thức và đóng kết nối
      conn.Write(protocol.Value{Typ: "error", Str: "ERR protocol error"}.Marshal())
      return
    }

    // 2. Chuyển lệnh đã parse tới CommandsHandler
    response := s.handler.HandleCommand(cmdValue)

    // 3. Gửi phản hồi RESP trở lại client
    _, err = conn.Write(response)
    if err != nil {
      log.Printf("Error writing response to %s: %v", conn.RemoteAddr(), err)
      return
    }
  }
}
