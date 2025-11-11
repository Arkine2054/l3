package cache

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	rdb *redis.Client
}

func New(addr string) *Client {
	if strings.HasPrefix(addr, "redis://") {
		opt, err := redis.ParseURL(addr)
		if err != nil {
			panic(err)
		}
		return &Client{rdb: redis.NewClient(opt)}
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &Client{rdb: rdb}
}

func (c *Client) Close() error {
	return c.rdb.Close()
}

func (c *Client) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

func (c *Client) SetFull(ctx context.Context, id int64, status string, attempts int, lastErr string) error {
	key := statusKey(id)
	m := map[string]interface{}{
		"status":   status,
		"attempts": attempts,
	}
	if lastErr != "" {
		m["last_error"] = lastErr
	}
	return c.rdb.HSet(ctx, key, m).Err()
}

func (c *Client) SetStatus(ctx context.Context, id int64, status string, ttl time.Duration) error {
	key := statusKey(id)

	if err := c.rdb.HSet(ctx, key, "status", status).Err(); err != nil {
		return fmt.Errorf("redis HSet error: %w", err)
	}

	if ttl > 0 {
		if err := c.rdb.Expire(ctx, key, ttl).Err(); err != nil {
			return fmt.Errorf("redis Expire error: %w", err)
		}
	}

	return nil
}

func (c *Client) Get(ctx context.Context, id int64) (map[string]string, error) {
	res, err := c.rdb.HGetAll(ctx, statusKey(id)).Result()
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, redis.Nil
	}
	return res, nil
}

func statusKey(id int64) string {
	return fmt.Sprintf("notification:%d", id)
}
