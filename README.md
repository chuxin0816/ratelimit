# ratelimit
The ratelimit package provides a distributed token bucket rate limiter for Go using Redis.
## Usage

Import the package:

```go
import "github.com/chuxin0816/ratelimit"
```

```bash
go get "github.com/chuxin0816/ratelimit"
```

## Example

```go
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}
    
    // create a bucket with a capacity of 100 tokens and a fill rate of 200 tokens per second
    bucket := NewBucket(rdb, "tokenBucket", 100, 200)

    // consume 1 tokens from the bucket
    ok, err := bucket.Take(1)
    if err != nil {
        panic(err)
    }
    if !ok {
        fmt.Println("rate limited")
    }
```