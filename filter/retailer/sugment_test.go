package retailer

import (
	"context"
	"testing"

	"github.com/bondhan/golib/retailer"
)

func TestSegmentsScorrer(t *testing.T) {
	ret := &retailer.RetailerContext{
		Segments: []string{"segment1", "segment2"},
	}

	dataSet := []struct {
		value interface{}
		score int
	}{
		{
			"segment1",
			10,
		},
		{
			"segment3",
			-1,
		},
		{
			[]string{"segment1", "segment2"},
			10,
		},
		{
			[]string{"segment3"},
			-1,
		},
		{
			nil,
			-1,
		},
	}

	scorer := SegmentScorer(10)

	for _, data := range dataSet {
		if got := scorer(context.Background(), ret, data.value); got != data.score {
			t.Errorf("SegmentScorer() = %v, want %v", got, data.score)
		}
	}

}
