package retailer

import (
	"context"
	"errors"
	"sort"

	"github.com/bondhan/golib/domain/marketing"
	"github.com/bondhan/golib/domain/retailer"
	"github.com/bondhan/golib/log"
)

type Scorer func(ctx context.Context, retailer *retailer.RetailerContext, val interface{}) int

type ScorerFactory func(score int) Scorer

type PreFilter func(ctx context.Context, attributes map[string]interface{}) bool

type scoreStore struct {
	score int
	index int
}

var scorer *PromoScorer
var scorerFactories = map[string]ScorerFactory{
	AddressProvinceKey:    AddressProvinceScorer,
	AddressCityKey:        AddressCityScorer,
	AddressDistrictKey:    AddressDistrictScorer,
	AddressDistrictSubKey: AddressDistrictSubScorer,
	AddressZipKey:         AddressZipScorer,
	SegmentsKey:           SegmentScorer,
	AreaRegionKey:         AreaRegionScorer,
	AreaRegionSubKey:      AreaRegionSubScorer,
	AreaClusterKey:        AreaClusterScorer,
	AreaWarehouseKey:      AreaWarehouseScorer,
	ScheduleKey:           ScheduleScorer,
	WeightKey:             WeightScorer,
}

var defaultScore = map[string]int{
	AddressProvinceKey:    1,
	AddressCityKey:        3,
	AddressDistrictKey:    5,
	AddressDistrictSubKey: 10,
	AddressZipKey:         6,
	AreaRegionKey:         2,
	AreaRegionSubKey:      10,
	AreaClusterKey:        26,
	AreaWarehouseKey:      65,
	SegmentsKey:           130,
	ScheduleKey:           260,
	WeightKey:             1,
}

func Get() *PromoScorer {
	if scorer == nil {
		scorer = DefaultScorer()
	}
	return scorer
}

func DefaultScorer() *PromoScorer {
	return initScorer(defaultScore)
}

func initScorer(conf map[string]int) *PromoScorer {
	sc := &PromoScorer{
		scorers: make(map[string]Scorer),
		scores:  conf,
	}
	for k, v := range conf {
		if v <= 0 {
			continue
		}
		if fn, ok := scorerFactories[k]; ok {
			sc.scorers[k] = fn(v)
		}
	}
	return sc
}

func Configure(config map[string]int) {
	scorer = initScorer(config)
}

type PromoScorer struct {
	scorers map[string]Scorer
	scores  map[string]int
}

func PreScore(ctx context.Context, promo marketing.Promo) (int, error) {
	return Get().PreScore(ctx, promo.GetFilteredAttributes())
}

func Calculate(ctx context.Context, ret interface{}, promo marketing.Promo) (int, error) {
	logger := log.GetLogger(ctx, "filter/retailer", "Calculate")
	r, err := retailer.DecodeRetailer(ret)
	if err != nil {
		logger.WithError(err).Error("failed to decode retailer")
		return -1, err
	}

	return Get().Calculate(ctx, r, promo.GetFilteredAttributes())
}

func (s PromoScorer) PreScore(ctx context.Context, attributes map[string]interface{}) (int, error) {
	score := 0
	for k := range attributes {
		if i, ok := s.scores[k]; ok {
			score += i
		}
	}
	return score, nil
}

func (s *PromoScorer) Calculate(ctx context.Context, retailer *retailer.RetailerContext, attributes map[string]interface{}) (int, error) {
	score := 0
	for k, v := range attributes {
		if fn, ok := s.scorers[k]; ok {
			i := fn(ctx, retailer, v)
			if i < 0 {
				return i, nil
			}
			score += i
		}
	}
	return score, nil
}

func Top(ctx context.Context, ret interface{}, promos []interface{ marketing.Promo }, fns ...PreFilter) (marketing.Promo, error) {
	logger := log.GetLogger(ctx, "filter/retailer", "Top")
	r, err := retailer.DecodeRetailer(ret)
	if err != nil {
		logger.WithError(err).Error("failed to decode retailer")
		return nil, err
	}

	ps := Get()

	max := -1
	var top marketing.Promo
	for _, promo := range promos {
		attrs := promo.GetFilteredAttributes()
		if !IsMatch(ctx, attrs, fns...) {
			continue
		}
		score, err := ps.Calculate(ctx, r, attrs)
		if err != nil {
			logger.WithError(err).WithField("promo", promo.GetID()).Warn("failed to calculate score")
			return nil, err
		}
		if score > max {
			max = score
			top = promo
		}
	}
	if max < 0 {
		logger.WithError(err).Error("no promo matched")
		return nil, errors.New("[filter/retailer] no promo matched")
	}

	return top, nil
}

func Validates(ctx context.Context, ret interface{}, promos []interface{ marketing.Promo }, isSort bool, fns ...PreFilter) ([]interface{ marketing.Promo }, error) {
	logger := log.GetLogger(ctx, "filter/retailer", "Sort")
	r, err := retailer.DecodeRetailer(ret)
	if err != nil {
		logger.WithError(err).Error("failed to decode retailer")
		return nil, err
	}

	ps := Get()

	scores := make([]scoreStore, 0, len(promos))

	for i, promo := range promos {
		attrs := promo.GetFilteredAttributes()
		if !IsMatch(ctx, attrs, fns...) {
			continue
		}
		score, err := ps.Calculate(ctx, r, attrs)
		if err != nil {
			logger.WithError(err).WithField("promo", promo.GetID()).Warn("failed to calculate score")
			continue
		}
		if score < 0 {
			logger.WithField("promo", promo.GetID()).Warn("promo not matched")
			continue
		}
		scores = append(scores, scoreStore{score, i})
	}
	if isSort {
		sort.Slice(scores, func(i, j int) bool {
			return scores[i].score > scores[j].score
		})
	}

	out := make([]interface{ marketing.Promo }, len(scores))

	for i, score := range scores {
		out[i] = promos[score.index]
	}

	return out, nil
}

func IsMatch(ctx context.Context, attrs map[string]interface{}, fns ...PreFilter) bool {
	for _, fn := range fns {
		if !fn(ctx, attrs) {
			return false
		}
	}
	return true
}
