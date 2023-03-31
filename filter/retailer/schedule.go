package retailer

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/bondhan/golib/retailer"
	"github.com/robfig/cron/v3"
)

const ScheduleKey = "Schedule"
const WeightKey = "Weight"

type ctxKey string

func (c ctxKey) String() string {
	return "retailer filter context key " + string(c)
}

const debugTimeCtx = ctxKey("debugTime")

func WithCustomTime(ctx context.Context, ts time.Time) context.Context {
	return context.WithValue(ctx, debugTimeCtx, ts)
}

func getTime(ctx context.Context) time.Time {
	if ctx == nil {
		return time.Now()
	}

	if ts, ok := ctx.Value(debugTimeCtx).(time.Time); ok {
		return ts
	}

	return time.Now()
}

func ScheduleScorer(score int) Scorer {
	specParser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)

	return func(ctx context.Context, ret *retailer.RetailerContext, val interface{}) int {
		switch v := val.(type) {
		case string:
			if isOnSchedule(ctx, specParser, v) {
				return score
			}

			return -1
		case []string:
			for _, s := range v {
				if isOnSchedule(ctx, specParser, s) {
					return score
				}
			}
			return -1
		default:
			return -1
		}
	}
}

func WeightScorer(score int) Scorer {
	if score < 0 {
		score = 1
	}
	return func(ctx context.Context, ret *retailer.RetailerContext, val interface{}) int {
		switch v := val.(type) {
		case int:
			return v * score
		case int32:
			return int(v) * score
		case int64:
			return int(v) * score
		case uint:
			return int(v) * score
		case uint32:
			return int(v) * score
		case uint64:
			return int(v) * score
		case float32:
			return int(v) * score
		case float64:
			return int(v) * score
		case string:
			if i, err := strconv.Atoi(v); err == nil {
				return i * score
			}
			return 0
		default:
			return 0
		}
	}
}

func isOnSchedule(ctx context.Context, parser cron.Parser, schedule string) bool {
	parts := strings.Split(schedule, " ")
	dur := parts[0]
	exp := strings.Join(parts[1:], " ")
	sched, err := parser.Parse(exp)
	if err != nil {
		return false
	}
	ds, err := time.ParseDuration(dur)
	if err != nil {
		return false
	}

	ts := getTime(ctx)
	next := sched.Next(ts.Add(-ds))
	until := next.Add(ds)
	if ts.After(next) && ts.Before(until) {
		return true
	}
	return false
}

type ActiveTime struct {
	Start time.Time
	End   time.Time
}

func PopulateActiveTimes(start, end time.Time, schedule string) ([]ActiveTime, error) {
	var out []ActiveTime
	if schedule == "" {
		out = append(out, ActiveTime{Start: start, End: end})
		return out, nil
	}
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	parts := strings.Split(schedule, " ")
	dur := parts[0]
	exp := strings.Join(parts[1:], " ")
	sched, err := parser.Parse(exp)
	if err != nil {
		return nil, err
	}
	ds, err := time.ParseDuration(dur)
	if err != nil {
		return nil, err
	}

	begin := start.Add(-1 * time.Second)
	for {
		next := sched.Next(begin)
		until := next.Add(ds)
		out = append(out, ActiveTime{Start: next, End: until})
		if until.Add(1 * time.Second).After(end) {
			break
		}
		begin = until.Add(-1 * time.Second)
	}

	return out, nil
}
