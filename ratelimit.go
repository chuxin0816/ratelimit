package ratelimit

import (
	"context"
	_ "embed"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
	xrate "golang.org/x/time/rate"
)

const pingInterval = time.Millisecond * 100

var (
	//go:embed tokenscript.lua
	luaScript   string
	tokenScript = redis.NewScript(luaScript)
)

type Bucket struct {
	rdb           rediser
	key           string
	capacity      int
	rate          int
	redisAlive    uint32
	rescueLock    sync.Mutex
	rescueLimiter *xrate.Limiter
	monitoring    bool
}

type rediser interface {
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd
	EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) *redis.Cmd
	EvalRO(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd
	EvalShaRO(ctx context.Context, sha1 string, keys []string, args ...interface{}) *redis.Cmd
	ScriptExists(ctx context.Context, hashes ...string) *redis.BoolSliceCmd
	ScriptLoad(ctx context.Context, script string) *redis.StringCmd
	Ping(ctx context.Context) *redis.StatusCmd
}

func NewBucket(rdb rediser, key string, rate, capacity int) *Bucket {
	if rate <= 0 {
		panic("rate must be greater than 0")
	}
	if capacity < 0 {
		panic("capacity must be greater than or equal to 0")
	}
	bucket := &Bucket{
		rdb:           rdb,
		key:           key,
		capacity:      capacity,
		rate:          rate,
		redisAlive:    1,
		rescueLimiter: xrate.NewLimiter(xrate.Every(time.Second/time.Duration(rate)), capacity),
	}
	if rdb == nil {
		bucket.redisAlive = 0
	}
	return bucket
}

func (b *Bucket) Take() bool {
	return b.TakeN(1)
}

func (b *Bucket) TakeN(count int) bool {
	if atomic.LoadUint32(&b.redisAlive) == 0 {
		return b.rescueLimiter.AllowN(time.Now(), count)
	}
	result, err := tokenScript.Run(context.Background(), b.rdb, []string{b.key}, b.rate, b.capacity, time.Now().Unix(), count).Int()
	if err != nil {
		if err == redis.Nil || err == context.Canceled || err == context.DeadlineExceeded {
			return false
		}
		b.monitor()
		return b.rescueLimiter.AllowN(time.Now(), count)
	}
	return result > 0
}

func (b *Bucket) monitor() {
	b.rescueLock.Lock()
	defer b.rescueLock.Unlock()

	if b.monitoring {
		return
	}

	b.monitoring = true
	atomic.StoreUint32(&b.redisAlive, 0)

	go b.waitForRedis()
}

func (b *Bucket) waitForRedis() {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		b.rescueLock.Lock()
		b.monitoring = false
		b.rescueLock.Unlock()
	}()

	for range ticker.C {
		if b.rdb.Ping(context.Background()).Err() == nil {
			atomic.StoreUint32(&b.redisAlive, 1)
			return
		}
	}
}
