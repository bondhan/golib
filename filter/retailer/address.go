package retailer

import (
	"context"
	"strings"

	"github.com/bondhan/golib/domain/retailer"
)

const (
	AddressProvinceKey    = "AddressProvince"
	AddressCityKey        = "AddressCity"
	AddressDistrictKey    = "AddressDistrict"
	AddressDistrictSubKey = "AddressDistrictSub"
	AddressZipKey         = "AddressZip"
)

func AddressProvinceScorer(score int) Scorer {
	return func(ctx context.Context, ret *retailer.RetailerContext, val interface{}) int {
		switch v := val.(type) {
		case string:
			if strings.EqualFold(ret.Address.Province, v) {
				return score
			}
			return -1
		case []string:
			for _, vv := range v {
				if strings.EqualFold(ret.Address.Province, vv) {
					return score
				}
			}
			return -1
		default:
			return -1
		}
	}
}

func AddressCityScorer(score int) Scorer {
	return func(ctx context.Context, ret *retailer.RetailerContext, val interface{}) int {
		switch v := val.(type) {
		case string:
			if strings.EqualFold(ret.Address.City, v) {
				return score
			}
			return -1
		case []string:
			for _, vv := range v {
				if strings.EqualFold(ret.Address.City, vv) {
					return score
				}
			}
			return -1
		default:
			return -1
		}
	}
}

func AddressDistrictScorer(score int) Scorer {
	return func(ctx context.Context, ret *retailer.RetailerContext, val interface{}) int {
		switch v := val.(type) {
		case string:
			if strings.EqualFold(ret.Address.District, v) {
				return score
			}
			return -1
		case []string:
			for _, vv := range v {
				if strings.EqualFold(ret.Address.District, vv) {
					return score
				}
			}
			return -1
		default:
			return -1
		}
	}
}

func AddressDistrictSubScorer(score int) Scorer {
	return func(ctx context.Context, ret *retailer.RetailerContext, val interface{}) int {
		switch v := val.(type) {
		case string:
			if strings.EqualFold(ret.Address.DistrictSub, v) {
				return score
			}
			return -1
		case []string:
			for _, vv := range v {
				if strings.EqualFold(ret.Address.DistrictSub, vv) {
					return score
				}
			}
			return -1
		default:
			return -1
		}
	}
}

func AddressZipScorer(score int) Scorer {
	return func(ctx context.Context, ret *retailer.RetailerContext, val interface{}) int {
		switch v := val.(type) {
		case string:
			if strings.EqualFold(ret.Address.ZipCode, v) {
				return score
			}
			return -1
		case []string:
			for _, vv := range v {
				if strings.EqualFold(ret.Address.ZipCode, vv) {
					return score
				}
			}
			return -1
		default:
			return -1
		}
	}
}
