package protocol

import (
  "bufio"
  "fmt"
  "io"
  "strconv"
)

//Xử lý Giao tiếp Mạng & Serialization.
const (
  STRING  = '+'
  ERROR   = '-'
  INTEGER = ':'
  BULK    = '$'
  ARRAY   = '*'
)

type Value struct {
  Typ   string
  Str   string
  Num   int
  Bulk  string
  Array []Value
}

type Resp struct {
  reader *bufio.Reader
}

func NewResp(rd io.Reader) *Resp {
  return &Resp{reader: bufio.NewReader(rd)}
}

func (r *Resp) readLine() (line []byte, n int, err error) {
  line, err = r.reader.ReadBytes('\n')
  if err != nil {
    return nil, 0, err
  }

  n = len(line)
  if n > 1 && line[n-2] == '\r' {
    return line[:n-2], n, nil
  }

  return line, n, fmt.Errorf("bad format: expected CRLF ending")
}

func (r *Resp) readInteger() (x int, n int, err error) {
  line, readBytes, err := r.readLine()
  if err != nil {
    return 0, 0, err
  }

  num, err := strconv.Atoi(string(line))
  if err != nil {
    return 0, readBytes, err
  }

  return num, readBytes, nil
}

func (r *Resp) readBulk() (Value, int, error) {
  bulkLen, n, err := r.readInteger()
  if err != nil {
    return Value{}, 0, err
  }

  totalBytes := n

  if bulkLen == -1 {
    return Value{Typ: "null"}, totalBytes, nil
  }

  bulk := make([]byte, bulkLen)
  n, err = r.reader.Read(bulk)
  if err != nil {
    return Value{}, 0, err
  }
  totalBytes += n

  line, n, err := r.readLine()
  if err != nil || len(line) != 0 {
    return Value{}, 0, fmt.Errorf("protocol error: expected CRLF after bulk data")
  }
  totalBytes += n

  return Value{Typ: "bulk", Bulk: string(bulk)}, totalBytes, nil
}

func (r *Resp) readArray() (Value, int, error) {
  arrayLen, n, err := r.readInteger()
  if err != nil {
    return Value{}, 0, err
  }
  totalBytes := n

  if arrayLen == -1 {
    return Value{Typ: "null"}, totalBytes, nil
  }

  array := make([]Value, arrayLen)
  for i := 0; i < arrayLen; i++ {
    val, n, err := r.Read()
    if err != nil {
      return Value{}, 0, err
    }
    array[i] = val
    totalBytes += n
  }

  return Value{Typ: "array", Array: array}, totalBytes, nil
}

func (r *Resp) Read() (Value, int, error) {
  typ, err := r.reader.Peek(1)
  if err != nil {
    return Value{}, 0, err
  }

  r.reader.ReadByte()
  totalBytes := 1

  switch typ[0] {
  case STRING:
    line, n, err := r.readLine()
    return Value{Typ: "string", Str: string(line)}, totalBytes + n, err
  case ERROR:
    line, n, err := r.readLine()
    return Value{Typ: "error", Str: string(line)}, totalBytes + n, err
  case INTEGER:
    num, n, err := r.readInteger()
    return Value{Typ: "integer", Num: num}, totalBytes + n, err
  case BULK:
    val, n, err := r.readBulk()
    return val, totalBytes + n, err
  case ARRAY:
    val, n, err := r.readArray()
    return val, totalBytes + n, err
  default:
    return Value{}, 0, fmt.Errorf("unknown type: %c", typ[0])
  }
}

func (v Value) Marshal() []byte {
  switch v.Typ {
  case "string":
    return []byte(fmt.Sprintf("+%s\r\n", v.Str))
  case "error":
    return []byte(fmt.Sprintf("-%s\r\n", v.Str))
  case "integer":
    return []byte(fmt.Sprintf(":%d\r\n", v.Num))
  case "bulk":
    return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(v.Bulk), v.Bulk))
  case "null":
    return []byte("$-1\r\n")
  case "array":
    result := fmt.Sprintf("*%d\r\n", len(v.Array))
    for _, item := range v.Array {
      result += string(item.Marshal())
    }
    return []byte(result)
  default:
    return []byte("-ERR unknown type\r\n")
  }
}

// MarshalCommand creates a RESP array from command parts (for AOF writing)
func MarshalCommand(parts []string) []byte {
  array := make([]Value, len(parts))
  for i, part := range parts {
    array[i] = Value{Typ: "bulk", Bulk: part}
  }
  return Value{Typ: "array", Array: array}.Marshal()
}
