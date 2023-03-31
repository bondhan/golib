package cache

import (
	"context"

	"github.com/bondhan/golib/cache/driver"
)

type Cache struct {
	driver driver.CacheDriver
}

func New(urlStr string) (*Cache, error) {
	drv, err := driver.New(urlStr)
	if err != nil {
		return nil, err
	}
	return &Cache{driver: drv}, nil
}

func (c *Cache) Set(ctx context.Context, key string, value interface{}, expiration int) error {
	return c.driver.Set(ctx, key, value, expiration)
}

func (c *Cache) GetBytes(ctx context.Context, key string) ([]byte, error) {
	return c.driver.Get(ctx, key)
}

func (c *Cache) GetString(ctx context.Context, key string) (string, error) {
	return c.driver.GetString(ctx, key)
}

func (c *Cache) GetInt(ctx context.Context, key string) (int64, error) {
	return c.driver.GetInt(ctx, key)
}

func (c *Cache) GetFloat(ctx context.Context, key string) (float64, error) {
	return c.driver.GetFloat(ctx, key)
}

func (c *Cache) Get(ctx context.Context, key string, out interface{}) error {
	return c.driver.GetObject(ctx, key, out)
}

func (c *Cache) GetKeys(ctx context.Context, pattern string) []string {
	return c.driver.GetKeys(ctx, pattern)
}

func (c *Cache) Exist(ctx context.Context, key string) bool {
	return c.driver.Exist(ctx, key)
}

func (c *Cache) Delete(ctx context.Context, key string, opts ...driver.DeleteOptions) error {
	return c.driver.Delete(ctx, key, opts...)
}

func (c *Cache) RemainingTime(ctx context.Context, key string) int {
	return c.driver.RemainingTime(ctx, key)
}

func (c *Cache) As(i interface{}) bool {
	return c.driver.As(i)
}

func (c *Cache) Flush(ctx context.Context) error {
	return c.driver.Flush(ctx)
}
