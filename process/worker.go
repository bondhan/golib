package process

import (
	"context"
	"errors"
	"fmt"
)

const (
	SENDER_MODE = 1
	GETTER_MODE = 2
)

type pack struct {
	name  string
	value interface{}
	err   error
}

type GetterFunc func(ctx context.Context, arg interface{}) (interface{}, error)
type SenderFunc func(ctx context.Context, arg interface{}) error

type Worker struct {
	mode  int
	tasks map[string]interface{}
	args  map[string]interface{}
}

func NewWorker() *Worker {
	return &Worker{
		tasks: make(map[string]interface{}),
		args:  map[string]interface{}{},
	}
}

func (w *Worker) AddTask(id string, fn interface{}, arg interface{}) error {
	switch f := fn.(type) {
	case GetterFunc:
		if w.mode == 0 {
			w.mode = GETTER_MODE
		}
		if w.mode != GETTER_MODE {
			return errors.New("cannot mix the function, it's getter mode")
		}
		w.tasks[id] = f
	case SenderFunc:
		if w.mode == 0 {
			w.mode = SENDER_MODE
		}
		if w.mode != SENDER_MODE {
			return errors.New("cannot mix the function, it's sender mode")
		}
		w.tasks[id] = f
	default:
		return errors.New("unsupported function")
	}
	if arg != nil {
		w.args[id] = arg
	}
	return nil
}

func (w *Worker) Execute(ctx context.Context, args interface{}) error {
	_, err := w.Call(ctx, args)
	return err
}

func (w *Worker) Call(ctx context.Context, args interface{}) (map[string]interface{}, error) {
	if len(w.tasks) == 0 {
		return nil, errors.New("empty task")
	}

	receive := make(chan *pack)
	//Use this to signal that we're ready to receive data
	send := make(chan bool)
	defer close(receive)
	defer close(send)

	for n, t := range w.tasks {
		arg := args
		if a, ok := w.args[n]; ok {
			arg = a
		}
		go func(out chan<- *pack, ready <-chan bool, name string, fn interface{}) {
			for {
				//<-cc
				select {
				case <-ctx.Done():
					return
				case _, ok := <-ready:
					if !ok {
						return
					}
					switch fx := fn.(type) {
					case GetterFunc:
						o, e := fx(ctx, arg)
						out <- &pack{name: name, value: o, err: e}
						return
					case SenderFunc:
						e := fx(ctx, arg)
						out <- &pack{name: name, value: e != nil, err: e}
						return
					}
				}
			}

		}(receive, send, n, t)
	}

	out := make(map[string]interface{})

	for {
		select {
		case <-ctx.Done():
			return nil, errors.New("[worker] cancelled")
		case send <- true:
			pkg := <-receive
			if pkg.err != nil {
				fmt.Println("Error executing function")
				return nil, pkg.err
			}

			out[pkg.name] = pkg.value
			if len(out) == len(w.tasks) {
				return out, nil
			}
		}
	}
}
