package util

import (
	"fmt"

	"github.com/bondhan/golib/constant"
)

type AssertOpt struct {
	Path  string      `json:"path"`
	Op    string      `json:"op"`
	Value interface{} `json:"value"`
}

type AssertOpts []AssertOpt

func (f AssertOpts) and(ctx interface{}) (bool, error) {
	if len(f) == 0 {
		return true, nil
	}
	for _, opt := range f {
		switch opt.Op {
		case constant.OR:
			val, ok := opt.Value.(AssertOpts)
			if !ok {
				return false, fmt.Errorf("value is not filter opts")
			}
			ok, err := val.or(ctx)
			if err != nil {
				return false, err
			}
			if !ok {
				return false, nil
			}
		default:
			if !Assert(opt.Path, ctx, opt.Value, opt.Op) {
				return false, nil
			}
		}
	}

	return true, nil
}

func (f AssertOpts) or(ctx interface{}) (bool, error) {
	if len(f) == 0 {
		return true, nil
	}
	for _, opt := range f {
		switch opt.Op {
		case constant.AND:
			val, ok := opt.Value.(AssertOpts)
			if !ok {
				return false, fmt.Errorf("value is not filter opts")
			}
			ok, err := val.and(ctx)
			if err != nil {
				return false, err
			}
			if ok {
				return true, nil
			}
		default:
			if Assert(opt.Path, ctx, opt.Value, opt.Op) {
				return true, nil
			}
		}
	}

	return false, nil
}

func (f AssertOpts) Assert(ctx interface{}) (bool, error) {
	if len(f) == 0 {
		return true, nil
	}

	return f.and(ctx)

}
