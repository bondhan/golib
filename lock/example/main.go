package main

import (
	"context"
	"fmt"
	"time"

	"github.com/bondhan/golib/lock"
	_ "github.com/bondhan/golib/lock/redis"
)

func main() {
	ctx := context.Background()
	url := "redis://localhost:6379,localhost:6380,localhost:6381/test"

	dlock, err := lock.New(url)
	if err != nil {
		panic(err)
	}
	for i := 0; i < 1000; i++ {
		orderID := ""

		for {
			orderID = generate()
			// Try lock and return immediately
			if err := dlock.TryLock(ctx, orderID, 20); err == nil {
				break
			}
			fmt.Println("duplicate")
		}
		fmt.Println(orderID)
		time.Sleep(5 * time.Microsecond)
	}

}

func generate() string {
	now := time.Now().UnixMilli()
	orderID := fmt.Sprintf("GTO-%v", now)
	return orderID
}
