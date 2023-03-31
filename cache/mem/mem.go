package mem

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/mitchellh/mapstructure"

	"github.com/bondhan/golib/cache/driver"
)

const schema = "mem"

var (
	ErrKeyAlreadyExists = errors.New("key already exists")
)

type memObject struct {
	expired time.Time
	value   interface{}
	timer   *time.Timer
	deleted bool
}

// MemoryCache memory cache object
type MemoryCache struct {
	data map[string]*memObject
	mux  *sync.RWMutex
}

func init() {
	driver.Register(schema, NewCache)
}

// NewCache create new memory cache
func NewCache(url *url.URL) (driver.CacheDriver, error) {
	return &MemoryCache{
		data: make(map[string]*memObject),
		mux:  &sync.RWMutex{},
	}, nil
}

// NewMemoryCache new memory instance
func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		data: make(map[string]*memObject),
		mux:  &sync.RWMutex{},
	}
}

func (m *MemoryCache) set(key string, value interface{}, exp int) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	mo := &memObject{}

	switch val := value.(type) {
	case string, bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, []byte:
		mo.value = val
	default:
		b, err := json.Marshal(val)
		if err != nil {
			return err
		}
		mo.value = b
	}

	if exp > 0 {
		mo.expired = time.Now().Add(time.Duration(exp) * time.Second)

		// A timer function is scheduled to run. It runs in a separate goroutine
		// and deletes the cache entry
		mo.timer = time.AfterFunc(time.Duration(exp)*time.Second, func() {
			m.mux.Lock()
			if mo.deleted {
				m.mux.Unlock()
				return
			}
			delete(m.data, key)
			m.mux.Unlock()
		})
	}

	m.data[key] = mo

	return nil
}

func (m *MemoryCache) get(key string) interface{} {
	m.mux.RLock()
	val, ok := m.data[key]
	m.mux.RUnlock()

	if !ok {
		return nil
	}

	return val.value
}

func (m *MemoryCache) del(key string) {
	m.mux.Lock()
	defer m.mux.Unlock()

	v, ok := m.data[key]
	if !ok {
		return
	}

	delete(m.data, key)

	if !v.timer.Stop() {
		v.deleted = true
	}
}

// Set adds an item to the cache. The item is removed from the cache upon timeout
//
// If the Set operation is successful, it returns a nil error. If there is an existing
// entry for the specified key,, the cache entry is not updated and an error (ErrKeyAlreadyExists)
// is returned
func (m *MemoryCache) Set(_ context.Context, key string, value interface{}, expiration int) error {
	return m.set(key, value, expiration)
}

// Get returns the value for the specified key and returns it as:
//
//	[]byte
//
// If the key does not exist, an error is returned (driver.NotFound)
func (m *MemoryCache) Get(_ context.Context, key string) ([]byte, error) {
	val := m.get(key)
	if val == nil {
		return nil, driver.NotFound
	}

	switch v := val.(type) {
	case int, int8, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return []byte(fmt.Sprintf("%v", v)), nil
	case bool:
		if v {
			return []byte("1"), nil
		}
		return []byte("0"), nil
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	default:
		return json.Marshal(v)
	}
}

// GetObject get value in object
func (m *MemoryCache) GetObject(_ context.Context, key string, doc interface{}) error {
	val := m.get(key)
	if val == nil {
		return driver.NotFound
	}

	switch v := val.(type) {
	case []byte:
		return json.Unmarshal(v, doc)
	default:
		return mapstructure.Decode(val, doc)
	}
}

// GetString get string value
func (m *MemoryCache) GetString(_ context.Context, key string) (string, error) {
	val := m.get(key)
	if val == nil {
		return "", driver.NotFound
	}
	return fmt.Sprintf("%v", val), nil
}

// GetInt get int value
func (m *MemoryCache) GetInt(_ context.Context, key string) (int64, error) {
	val := m.get(key)
	if val == nil {
		return 0, driver.NotFound
	}

	vi, err := strconv.Atoi(fmt.Sprintf("%v", val))
	if err != nil {
		return 0, err
	}

	return int64(vi), nil
}

// GetFloat get float value
func (m *MemoryCache) GetFloat(_ context.Context, key string) (float64, error) {
	val := m.get(key)
	if val == nil {
		return 0, driver.NotFound
	}

	f, ok := val.(float64)
	if !ok {
		return 0, errors.New("invalid stored value")
	}

	return f, nil
}

// Exist check if key exist
func (m *MemoryCache) Exist(_ context.Context, key string) bool {
	return m.get(key) != nil
}

// RemainingTime get remaining time
func (m *MemoryCache) RemainingTime(_ context.Context, key string) int {
	m.mux.RLock()
	val, ok := m.data[key]
	m.mux.RUnlock()

	if !ok {
		return -1
	}

	if !val.deleted {
		return int(math.Ceil(time.Until(val.expired).Seconds()))
	}

	return -1
}

// Delete delete record
func (m *MemoryCache) Delete(_ context.Context, key string, opts ...driver.DeleteOptions) error {
	m.del(key)

	return nil
}

func (m *MemoryCache) GetKeys(ctx context.Context, pattern string) []string {
	keys := make([]string, len(m.data))
	i := 0
	for k := range m.data {
		keys[i] = k
		i++
	}
	return keys
}

// Close close cache
func (m *MemoryCache) Close() error {
	m.mux.Lock()
	m.data = make(map[string]*memObject)
	m.mux.Unlock()
	return nil
}

func (m *MemoryCache) As(i interface{}) bool {
	return false
}

func (m *MemoryCache) Flush(ctx context.Context) error {
	return m.Close()
}
