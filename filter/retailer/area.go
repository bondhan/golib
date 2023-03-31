package retailer

import (
	"context"
	"strings"

	"github.com/bondhan/golib/domain/retailer"
)

const (
	AreaRegionKey    = "AreaRegion"
	AreaRegionSubKey = "AreaRegionSub"
	AreaClusterKey   = "AreaCluster"
	AreaWarehouseKey = "AreaWarehouse"
)

func AreaRegionScorer(score int) Scorer {
	return func(ctx context.Context, ret *retailer.RetailerContext, val interface{}) int {
		switch v := val.(type) {
		case string:
			if strings.EqualFold(ret.Area.Region, v) {
				return score
			}
			return -1
		case []string:
			for _, vv := range v {
				if strings.EqualFold(ret.Area.Region, vv) {
					return score
				}
			}
			return -1
		default:
			return -1
		}
	}
}

func AreaRegionSubScorer(score int) Scorer {
	return func(ctx context.Context, ret *retailer.RetailerContext, val interface{}) int {
		switch v := val.(type) {
		case string:
			if strings.EqualFold(ret.Area.RegionSub, v) {
				return score
			}
			return -1
		case []string:
			for _, vv := range v {
				if strings.EqualFold(ret.Area.RegionSub, vv) {
					return score
				}
			}
			return -1
		default:
			return -1
		}
	}
}

func AreaClusterScorer(score int) Scorer {
	return func(ctx context.Context, ret *retailer.RetailerContext, val interface{}) int {
		switch v := val.(type) {
		case string:
			if strings.EqualFold(ret.Area.Cluster, v) {
				return score
			}
			return -1
		case []string:
			for _, vv := range v {
				if strings.EqualFold(ret.Area.Cluster, vv) {
					return score
				}
			}
			return -1
		default:
			return -1
		}
	}
}

func AreaWarehouseScorer(score int) Scorer {
	return func(ctx context.Context, ret *retailer.RetailerContext, val interface{}) int {
		switch v := val.(type) {
		case string:
			if strings.EqualFold(ret.Area.Warehouse, v) {
				return score
			}
			return -1
		case []string:
			for _, vv := range v {
				if strings.EqualFold(ret.Area.Warehouse, vv) {
					return score
				}
			}
			return -1
		default:
			return -1
		}
	}
}
