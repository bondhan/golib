package retailer

import (
	"context"
	"strings"

	"github.com/bondhan/golib/domain/retailer"
)

const SegmentsKey = "Segments"

func SegmentScorer(score int) Scorer {
	return func(ctx context.Context, ret *retailer.RetailerContext, val interface{}) int {
		switch v := val.(type) {
		case string:
			for _, vv := range ret.Segments {
				if strings.EqualFold(vv, v) {
					return score
				}
			}
			return -1
		case []string:
			for _, vv := range v {
				for _, vvv := range ret.Segments {
					if strings.EqualFold(vvv, vv) {
						return score
					}
				}
			}
			return -1
		default:
			return -1
		}
	}
}
