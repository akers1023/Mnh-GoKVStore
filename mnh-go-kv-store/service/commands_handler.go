package service

import (
  "fmt"
  "strconv"
  "strings"
  "time"

  "mnhgo/mnh-go-kv-store/internal/protocol"
  "mnhgo/mnh-go-kv-store/internal/store"
)

// HandlerFunc định nghĩa chữ ký cho tất cả các hàm xử lý lệnh
type HandlerFunc func(s *store.Store, aof *store.AOF, args []protocol.Value) []byte

// CommandsHandler chứa các tham chiếu đến Store và AOF để thực hiện lệnh
type CommandsHandler struct {
  store    *store.Store
  aof      *store.AOF
  commands map[string]HandlerFunc
}

func NewCommandsHandler(s *store.Store, aof *store.AOF) *CommandsHandler {
  h := &CommandsHandler{
    store: s,
    aof:   aof,
  }
  h.commands = map[string]HandlerFunc{
    "PING":    h.handlePING,
    "SET":     h.handleSET,
    "GET":     h.handleGET,
    "DEL":     h.handleDEL,
    "EXISTS":  h.handleEXISTS,
    "TTL":     h.handleTTL,
    "HSET":    h.handleHSET,
    "HGET":    h.handleHGET,
    "HGETALL": h.handleHGETALL,
    // Thêm các lệnh khác vào đây
  }
  return h
}

func (h *CommandsHandler) ExecuteAOFCommand(cmdValue protocol.Value) {
  if cmdValue.Typ != "array" || len(cmdValue.Array) == 0 {
    return
  }

  commandName := strings.ToUpper(cmdValue.Array[0].Bulk)
  args := cmdValue.Array[1:]

  if handler, ok := h.commands[commandName]; ok {
    handler(h.store, nil, args)
  }
}

// HandleCommand là điểm vào chính để xử lý lệnh từ client
func (h *CommandsHandler) HandleCommand(cmdValue protocol.Value) []byte {
  // Kiểm tra xem lệnh có phải là Array không
  if cmdValue.Typ != "array" || len(cmdValue.Array) == 0 {
    return protocol.Value{Typ: "error", Str: "ERR invalid command format"}.Marshal()
  }

  // Lấy tên lệnh (phần tử đầu tiên) và chuyển thành chữ hoa
  commandName := strings.ToUpper(cmdValue.Array[0].Bulk)

  // Lấy các đối số (phần còn lại của mảng)
  args := cmdValue.Array[1:]

  // Tìm handler
  if handler, ok := h.commands[commandName]; ok {
    return handler(h.store, h.aof, args)
  }

  return protocol.Value{Typ: "error", Str: fmt.Sprintf("ERR unknown command '%s'", commandName)}.Marshal()
}

func (h *CommandsHandler) handlePING(s *store.Store, aof *store.AOF, args []protocol.Value) []byte {
  return protocol.Value{Typ: "string", Str: "PONG"}.Marshal()
}

func (h *CommandsHandler) handleSET(s *store.Store, aof *store.AOF, args []protocol.Value) []byte {
  if len(args) < 2 {
    return protocol.Value{Typ: "error", Str: "ERR wrong number of arguments for 'set' command"}.Marshal()
  }
  // Giả định args[0] là key và args[1] là value (Bulk String)
  key := args[0].Bulk
  value := args[1].Bulk

  // Parse TTL arguments (EX = seconds, PX = milliseconds)
  var ttl time.Duration = 0
  if len(args) >= 4 {
    ttlStr := strings.ToUpper(args[2].Bulk)
    if ttlStr == "EX" || ttlStr == "PX" {
      secs, err := strconv.Atoi(args[3].Bulk)
      if err == nil {
        if ttlStr == "EX" {
          ttl = time.Duration(secs) * time.Second
        } else {
          ttl = time.Duration(secs) * time.Millisecond
        }
      }
    }
  }

  s.SET(key, value, ttl)

  // Ghi lệnh vào AOF
  if aof != nil {
    commandParts := []string{"SET", key, value}
    if ttl > 0 {
      commandParts = append(commandParts, "EX", strconv.Itoa(int(ttl.Seconds())))
    }
    aof.WriteCommand(protocol.MarshalCommand(commandParts))
  }

  return protocol.Value{Typ: "string", Str: "OK"}.Marshal()
}

func (h *CommandsHandler) handleGET(s *store.Store, aof *store.AOF, args []protocol.Value) []byte {
  if len(args) != 1 {
    return protocol.Value{Typ: "error", Str: "ERR wrong number of arguments for 'get' command"}.Marshal()
  }

  value, found := s.GET(args[0].Bulk)

  if !found {
    return protocol.Value{Typ: "null"}.Marshal()
  }
  return protocol.Value{Typ: "bulk", Bulk: value}.Marshal()
}

func (h *CommandsHandler) handleHSET(s *store.Store, aof *store.AOF, args []protocol.Value) []byte {
  if len(args) < 3 || len(args)%2 != 1 {
    return protocol.Value{Typ: "error", Str: "ERR wrong number of arguments for 'hset' command"}.Marshal()
  }

  key := args[0].Bulk
  // HSET có thể có nhiều cặp field-value
  fieldsAdded := 0
  for i := 1; i < len(args); i += 2 {
    field := args[i].Bulk
    value := args[i+1].Bulk
    if s.HSET(key, field, value) {
      fieldsAdded++
    }
  }

  // Ghi lệnh vào AOF
  if aof != nil {
    commandParts := make([]string, len(args)+1)
    commandParts[0] = "HSET"
    commandParts[1] = key
    for i := 1; i < len(args); i++ {
      commandParts[i+1] = args[i].Bulk
    }
    aof.WriteCommand(protocol.MarshalCommand(commandParts))
  }

  return protocol.Value{Typ: "integer", Num: fieldsAdded}.Marshal()
}

func (h *CommandsHandler) handleHGET(s *store.Store, aof *store.AOF, args []protocol.Value) []byte {
  if len(args) != 2 {
    return protocol.Value{Typ: "error", Str: "ERR wrong number of arguments for 'hget' command"}.Marshal()
  }

  key := args[0].Bulk
  field := args[1].Bulk

  value, found := s.HGET(key, field)

  if !found {
    return protocol.Value{Typ: "null"}.Marshal()
  }
  return protocol.Value{Typ: "bulk", Bulk: value}.Marshal()
}

func (h *CommandsHandler) handleDEL(s *store.Store, aof *store.AOF, args []protocol.Value) []byte {
  if len(args) == 0 {
    return protocol.Value{Typ: "error", Str: "ERR wrong number of arguments for 'del' command"}.Marshal()
  }

  // Xóa nhiều key
  count := 0
  keysToDelete := make([]string, 0)
  for _, arg := range args {
    key := arg.Bulk
    // Kiểm tra nếu key tồn tại trước khi xóa (để đếm đúng)
    exists := s.EXISTS(key)
    s.DELETE(key)
    if exists {
      count++
      keysToDelete = append(keysToDelete, key)
    }
  }

  // Ghi lệnh vào AOF
  if aof != nil && len(keysToDelete) > 0 {
    commandParts := append([]string{"DEL"}, keysToDelete...)
    aof.WriteCommand(protocol.MarshalCommand(commandParts))
  }

  return protocol.Value{Typ: "integer", Num: count}.Marshal()
}

func (h *CommandsHandler) handleEXISTS(s *store.Store, aof *store.AOF, args []protocol.Value) []byte {
  if len(args) == 0 {
    return protocol.Value{Typ: "error", Str: "ERR wrong number of arguments for 'exists' command"}.Marshal()
  }

  existsCount := 0
  for _, arg := range args {
    if s.EXISTS(arg.Bulk) {
      existsCount++
    }
  }

  return protocol.Value{Typ: "integer", Num: existsCount}.Marshal()
}

func (h *CommandsHandler) handleTTL(s *store.Store, aof *store.AOF, args []protocol.Value) []byte {
  if len(args) != 1 {
    return protocol.Value{Typ: "error", Str: "ERR wrong number of arguments for 'ttl' command"}.Marshal()
  }

  ttl := s.TTL(args[0].Bulk)
  return protocol.Value{Typ: "integer", Num: int(ttl)}.Marshal()
}

func (h *CommandsHandler) handleHGETALL(s *store.Store, aof *store.AOF, args []protocol.Value) []byte {
  if len(args) != 1 {
    return protocol.Value{Typ: "error", Str: "ERR wrong number of arguments for 'hgetall' command"}.Marshal()
  }

  key := args[0].Bulk
  hash, found := s.HGETALL(key)

  if !found {
    return protocol.Value{Typ: "array", Array: []protocol.Value{}}.Marshal()
  }

  // Return an array alternating field and value
  array := make([]protocol.Value, len(hash)*2)
  idx := 0
  for field, value := range hash {
    array[idx] = protocol.Value{Typ: "bulk", Bulk: field}
    array[idx+1] = protocol.Value{Typ: "bulk", Bulk: value}
    idx += 2
  }

  return protocol.Value{Typ: "array", Array: array}.Marshal()
}

