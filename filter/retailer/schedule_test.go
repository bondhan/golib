package retailer

import (
	"context"
	"testing"
	"time"

	"github.com/bondhan/golib/retailer"
)

func TestSchedule(t *testing.T) {
	ret := &retailer.RetailerContext{
		Address: &retailer.Address{
			District: "Cimahi",
		},
	}

	dataSet := []struct {
		value interface{}
		score int
	}{
		{
			"5m * * * * *",
			1,
		},
		{
			"10s",
			-1,
		},
		{
			"0 0 1 1 *",
			-1,
		},
		{
			nil,
			-1,
		},
	}

	scorer := ScheduleScorer(1)

	for _, data := range dataSet {
		if got := scorer(context.Background(), ret, data.value); got != data.score {
			t.Errorf("ScheduleScorer() = %v, want %v", got, data.score)
		}
	}
}

func TestCustomTime(t *testing.T) {
	ret := &retailer.RetailerContext{
		Address: &retailer.Address{
			District: "Cimahi",
		},
	}

	dataSet := []struct {
		value interface{}
		score int
	}{
		{
			"1h 1 1 1 4 *",
			1,
		},
		{
			"1m 1 1 1 4 *",
			-1,
		},
		{
			"10s",
			-1,
		},
		{
			"0 0 1 1 *",
			-1,
		},
		{
			nil,
			-1,
		},
	}

	scorer := ScheduleScorer(1)
	ctx := WithCustomTime(context.Background(), time.Date(2021, time.April, 1, 1, 5, 0, 0, time.UTC))
	for _, data := range dataSet {
		if got := scorer(ctx, ret, data.value); got != data.score {
			t.Errorf("ScheduleScorer() = %v, want %v", got, data.score)
		}
	}
}

func TestWeightScorer(t *testing.T) {
	scorer := WeightScorer(1)
	ret := &retailer.RetailerContext{
		Address: &retailer.Address{
			District: "Cimahi",
		},
	}

	dataSet := []struct {
		value interface{}
		score int
	}{
		{
			10,
			10,
		},
		{
			int32(2),
			2,
		},
		{
			int64(4),
			4,
		},
		{
			uint(5),
			5,
		},
		{
			uint32(6),
			6,
		},
		{
			uint64(7),
			7,
		},
		{
			float32(8),
			8,
		},
		{
			float64(9),
			9,
		},
		{
			"10",
			10,
		},
		{
			"0.1",
			0,
		},
	}

	ctx := context.Background()
	for _, data := range dataSet {
		if got := scorer(ctx, ret, data.value); got != data.score {
			t.Errorf("ScheduleScorer() = %v, want %v", got, data.score)
		}
	}
}

func TestPopulate(t *testing.T) {
	schedule := "1h 0 * * * *"

	ts := time.Date(2023, 2, 23, 10, 0, 0, 0, time.Local)
	out, err := PopulateActiveTimes(ts, ts.Add(5*time.Hour), schedule)
	if err != nil {
		t.Error(err)
	}

	if len(out) != 5 {
		t.Error(out)
	}
}
