package ratelimit

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client

func init() {
	rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}
}

func TestTakeSuccess(t *testing.T) {
	bucket := NewBucket(rdb, "test", 10, 1)
	if ok, err := bucket.Take(1); err != nil {
		t.Errorf("error: %v", err)
	} else if !ok {
		t.Errorf("expected: true, got: %v", ok)
	}
}

func TestTakeFail(t *testing.T) {
	bucket := NewBucket(rdb, "test", 2, 2)
	if ok, err := bucket.Take(3); err != nil {
		t.Errorf("error: %v", err)
	} else if ok {
		t.Errorf("expected: false, got: %v", ok)
	}
}
