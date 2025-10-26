package client

import (
  "fmt"
  "net"
  "time"

  "mnhgo/mnh-go-kv-store/internal/protocol"
)

// Client struct quản lý kết nối TCP và I/O với Server
type Client struct {
  conn net.Conn
  resp *protocol.Resp
}

// NewClient thiết lập kết nối TCP đến địa chỉ server (ví dụ: "localhost:6379")
func NewClient(addr string) (*Client, error) {
  conn, err := net.Dial("tcp", addr)
  if err != nil {
    return nil, fmt.Errorf("failed to connect to %s: %w", addr, err)
  }

  client := &Client{
    conn: conn,
    // Sử dụng protocol.NewResp để đọc phản hồi từ kết nối
    resp: protocol.NewResp(conn),
  }
  return client, nil
}

// Close đóng kết nối client
func (c *Client) Close() error {
  return c.conn.Close()
}

// executeCommand gửi lệnh RESP và chờ phản hồi từ server
// cmds là các thành phần của lệnh (ví dụ: "SET", "key1", "value1")
func (c *Client) executeCommand(cmds ...string) (protocol.Value, error) {
  // 1. Mã hóa lệnh thành định dạng RESP Array
  array := make([]protocol.Value, len(cmds))
  for i, cmd := range cmds {
    // Tạo từng phần tử lệnh dưới dạng Bulk String
    array[i] = protocol.Value{Typ: "bulk", Bulk: cmd}
  }
  command := protocol.Value{Typ: "array", Array: array}

  // 2. Gửi byte stream RESP qua kết nối
  _, err := c.conn.Write(command.Marshal()) // Giả định protocol.Value có Marshal()
  if err != nil {
    return protocol.Value{}, fmt.Errorf("failed to write command: %w", err)
  }

  // 3. Đọc và giải mã phản hồi từ server
  response, _, err := c.resp.Read()
  if err != nil {
    return protocol.Value{}, fmt.Errorf("failed to read response: %w", err)
  }

  // 4. Xử lý lỗi từ server (RESP Error)
  if response.Typ == "error" {
    return protocol.Value{}, fmt.Errorf("server error: %s", response.Str)
  }

  return response, nil
}

// --- Các hàm API cụ thể ---

// SET: Thiết lập Key-Value
// ttl là Duration (ví dụ: 10 * time.Second). 0 có nghĩa là không hết hạn
func (c *Client) SET(key string, value string, ttl time.Duration) (string, error) {
  cmds := []string{"SET", key, value}

  if ttl > 0 {
    // Thêm tham số EX (seconds) cho TTL
    cmds = append(cmds, "EX", fmt.Sprintf("%d", int(ttl.Seconds())))
  }

  response, err := c.executeCommand(cmds...)
  if err != nil {
    return "", err
  }

  // Phản hồi thành công cho SET là Simple String "OK"
  if response.Typ != "string" || response.Str != "OK" {
    return "", fmt.Errorf("unexpected response type for SET: %s", response.Typ)
  }

  return response.Str, nil
}

// GET: Lấy giá trị của Key
func (c *Client) GET(key string) (string, error) {
  response, err := c.executeCommand("GET", key)
  if err != nil {
    return "", err
  }

  // Xử lý trường hợp Key không tồn tại (RESP Null Bulk String)
  if response.Typ == "null" {
    return "", nil
  }

  // Phản hồi thành công là Bulk String
  if response.Typ != "bulk" {
    return "", fmt.Errorf("unexpected response type for GET: %s", response.Typ)
  }

  return response.Bulk, nil
}

// HSET: Thiết lập Field-Value trong Hash
func (c *Client) HSET(key string, field string, value string) (int, error) {
  response, err := c.executeCommand("HSET", key, field, value)
  if err != nil {
    return 0, err
  }

  // HSET trả về Integer (số field được thêm)
  if response.Typ != "integer" {
    return 0, fmt.Errorf("unexpected response type for HSET: %s", response.Typ)
  }

  return response.Num, nil
}

// HGET: Lấy giá trị của Field trong Hash
func (c *Client) HGET(key string, field string) (string, error) {
  response, err := c.executeCommand("HGET", key, field)
  if err != nil {
    return "", err
  }

  if response.Typ == "null" {
    return "", nil
  }

  if response.Typ != "bulk" {
    return "", fmt.Errorf("unexpected response type for HGET: %s", response.Typ)
  }

  return response.Bulk, nil
}
