package ratelimit

import (
	"context"
	_ "embed"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	//go:embed tokenscript.lua
	luaScript   string
	tokenScript = redis.NewScript(luaScript)
)

type Bucket struct {
	client   rediser
	key      string
	capacity int64
	rate     float64
}

type rediser interface {
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd
	EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) *redis.Cmd
	EvalRO(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd
	EvalShaRO(ctx context.Context, sha1 string, keys []string, args ...interface{}) *redis.Cmd
	ScriptExists(ctx context.Context, hashes ...string) *redis.BoolSliceCmd
	ScriptLoad(ctx context.Context, script string) *redis.StringCmd
}

func NewBucket(client rediser, key string, capacity int64, rate float64) *Bucket {
	return &Bucket{
		client:   client,
		key:      key,
		capacity: capacity,
		rate:     rate,
	}
}

func (b *Bucket) Take(count int64) (bool, error) {
	result, err := tokenScript.Run(context.Background(), b.client, []string{b.key}, b.capacity, b.rate, time.Now().UnixMilli(), count).Int64()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}
	return result > 0, nil
}
