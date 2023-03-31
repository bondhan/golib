package process

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWorker(t *testing.T) {
	wrk := NewWorker()
	max := 10

	mfn := func(i int) GetterFunc {
		return func(ctx context.Context, job interface{}) (interface{}, error) {
			fmt.Println(job)
			return i, nil
		}
	}

	for i := 0; i < max; i++ {
		wrk.AddTask(fmt.Sprintf("%v", i), mfn(i), fmt.Sprintf("%v", i))
	}

	out, err := wrk.Call(context.Background(), "one")
	assert.Nil(t, err)
	assert.Equal(t, max, len(out))
	//assert.True(t, false)
}

func TestWorkerSender(t *testing.T) {
	wrk := NewWorker()
	max := 10

	mfn := func(i int) SenderFunc {
		return func(ctx context.Context, job interface{}) error {
			fmt.Println(job)
			return nil
		}
	}

	for i := 0; i < max; i++ {
		wrk.AddTask(fmt.Sprintf("%v", i), mfn(i), fmt.Sprintf("%v", i))
	}

	err := wrk.Execute(context.Background(), "one")
	assert.Nil(t, err)
	//assert.True(t, false)
}

func TestWorkerError(t *testing.T) {
	wrk := NewWorker()
	err := wrk.AddTask("one", GetterFunc(func(ctx context.Context, job interface{}) (interface{}, error) {
		fmt.Println("one")
		return "one", nil
	}), nil)
	assert.Nil(t, err)

	wrk.AddTask("two", GetterFunc(func(ctx context.Context, job interface{}) (interface{}, error) {
		return nil, errors.New("Just error")
	}), nil)

	wrk.AddTask("three", GetterFunc(func(ctx context.Context, job interface{}) (interface{}, error) {
		fmt.Println("three")
		return "three", nil
	}), nil)

	_, err = wrk.Call(context.Background(), "one")
	assert.NotNil(t, err)
	fmt.Println(err)
	//assert.False(t, true)
}

func TestWorkerCancel(t *testing.T) {
	wrk := NewWorker()
	max := 10

	mfn := func(i int) SenderFunc {
		return func(ctx context.Context, job interface{}) error {
			time.Sleep(1 * time.Second)
			fmt.Println(job)
			return nil
		}
	}

	for i := 0; i < max; i++ {
		wrk.AddTask(fmt.Sprintf("%v", i), mfn(i), fmt.Sprintf("%v", i))
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(1 * time.Second)
		cancel()
	}()
	err := wrk.Execute(ctx, "one")
	assert.NotNil(t, err)
	//assert.True(t, false)
}
