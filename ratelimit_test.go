package ratelimit

import (
	"context"
	"fmt"
	"testing"
	"time"

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
	bucket := NewBucket(rdb, "test", 1, 10)
	if !bucket.Take() {
		t.Errorf("expected: true, got: false")
	}

	bucket2 := NewBucket(rdb, "test2", 1, 0)
	if bucket2.Take() {
		t.Errorf("expected: false, got: true")
	}
}

func TestTakeN(t *testing.T) {
	bucket3 := NewBucket(rdb, "test3", 10, 10)
	if !bucket3.TakeN(5) {
		fmt.Println(time.Now().Unix())
		t.Errorf("expected: true, got: false")
	}

	bucket4 := NewBucket(rdb, "test4", 10, 1)
	if bucket4.TakeN(11) {
		t.Errorf("expected: false, got: true")
	}
}

func TestTakeWithRedisDown(t *testing.T) {
	bucket5 := NewBucket(nil, "test", 1, 10)
	if !bucket5.Take() {
		t.Errorf("expected: true, got: false")
	}

	bucket6 := NewBucket(nil, "test2", 1, 0)
	if bucket6.Take() {
		t.Errorf("expected: false, got: true")
	}
}

func TestTakeNWithRedisDown(t *testing.T) {
	bucket7 := NewBucket(nil, "test3", 10, 10)
	if !bucket7.TakeN(5) {
		t.Errorf("expected: true, got: false")
	}

	bucket8 := NewBucket(nil, "test4", 10, 1)
	if bucket8.TakeN(11) {
		t.Errorf("expected: false, got: true")
	}
}
