package store

import (
  "sync"
  "time"
)

// Entry lưu trữ giá trị và thời gian hết hạn của một key
type Entry struct {
  Value     interface{} // Có thể là string, map[string]string, list, v.v.
  ExpiresAt time.Time   // Thời điểm hết hạn (Zero time.Time nếu không hết hạn)
}

// Store chứa dữ liệu chính và Mutex để quản lý đồng thời
type Store struct {
  data map[string]Entry
  mu   sync.RWMutex // RWMutex cho phép đọc đồng thời, nhưng khóa khi ghi
}

func NewStore() *Store {
  return &Store{
    data: make(map[string]Entry),
  }
}

// SET: Thiết lập giá trị cho một key với thời gian hết hạn tùy chọn
func (s *Store) SET(key string, value string, ttl time.Duration) {
  s.mu.Lock()
  defer s.mu.Unlock()

  entry := Entry{Value: value}
  if ttl > 0 {
    entry.ExpiresAt = time.Now().Add(ttl)
  }

  s.data[key] = entry
}

// GET: Lấy giá trị từ một key
func (s *Store) GET(key string) (string, bool) {
  s.mu.RLock()
  entry, ok := s.data[key]
  s.mu.RUnlock()

  if !ok {
    return "", false
  }

  // Kiểm tra hết hạn (TTL)
  if !entry.ExpiresAt.IsZero() && time.Now().After(entry.ExpiresAt) {
    // Gọi DELETE để xóa entry hết hạn
    s.DELETE(key)
    return "", false
  }

  // Ép kiểu giá trị (giả sử là string cho lệnh GET cơ bản)
  strVal, isString := entry.Value.(string)
  if !isString {
    return "", false // Lỗi nếu key tồn tại nhưng không phải string
  }
  return strVal, true
}

// DELETE: Xóa một key khỏi Store
func (s *Store) DELETE(key string) {
  s.mu.Lock()
  defer s.mu.Unlock()

  delete(s.data, key)
}

// HSET: Thiết lập giá trị cho một trường (field) trong Hash
func (s *Store) HSET(key string, field string, value string) bool {
  s.mu.Lock()
  defer s.mu.Unlock()

  entry, ok := s.data[key]

  if !ok {
    // Key không tồn tại: tạo Entry mới với Hash Map
    hash := make(map[string]string)
    hash[field] = value
    s.data[key] = Entry{Value: hash}
    return true
  }

  hash, isHash := entry.Value.(map[string]string)
  if !isHash {
    // Lỗi: Key tồn tại nhưng không phải là Hash (ví dụ: là String)
    return false
  }

  hash[field] = value
  return true
}

// HGET: Lấy giá trị của một trường (field) trong Hash
func (s *Store) HGET(key string, field string) (string, bool) {
  s.mu.RLock()
  defer s.mu.RUnlock()

  entry, ok := s.data[key]
  if !ok {
    return "", false
  }

  // Kiểm tra và ép kiểu sang Hash Map
  hash, isHash := entry.Value.(map[string]string)
  if !isHash {
    return "", false
  }

  val, fieldFound := hash[field]
  return val, fieldFound
}

// HGETALL: Lấy tất cả field-value trong Hash
func (s *Store) HGETALL(key string) (map[string]string, bool) {
  s.mu.RLock()
  defer s.mu.RUnlock()

  entry, ok := s.data[key]
  if !ok {
    return nil, false
  }

  // Kiểm tra và ép kiểu sang Hash Map
  hash, isHash := entry.Value.(map[string]string)
  if !isHash {
    return nil, false
  }

  // Tạo copy để tránh race condition
  result := make(map[string]string)
  for k, v := range hash {
    result[k] = v
  }

  return result, true
}

// EXISTS: Kiểm tra xem key có tồn tại không
func (s *Store) EXISTS(key string) bool {
  s.mu.RLock()
  entry, ok := s.data[key]
  s.mu.RUnlock()

  if !ok {
    return false
  }

  // Kiểm tra hết hạn
  if !entry.ExpiresAt.IsZero() && time.Now().After(entry.ExpiresAt) {
    s.DELETE(key) // Delete expired entry
    return false
  }

  return true
}

// TTL: Lấy thời gian còn lại (Time To Live) của key, trả về giây
func (s *Store) TTL(key string) int {
  s.mu.RLock()
  entry, ok := s.data[key]
  s.mu.RUnlock()

  if !ok {
    return -2 // Key không tồn tại
  }

  if entry.ExpiresAt.IsZero() {
    return -1 // Key tồn tại nhưng không có TTL
  }

  ttl := entry.ExpiresAt.Sub(time.Now())
  if ttl <= 0 {
    s.DELETE(key) // Key đã hết hạn
    return -2
  }

  return int(ttl.Seconds())
}
