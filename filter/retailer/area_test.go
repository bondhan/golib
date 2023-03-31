package retailer

import (
	"context"
	"testing"

	"github.com/bondhan/golib/retailer"
)

func TestRegionScorrer(t *testing.T) {
	ret := &retailer.RetailerContext{
		Area: &retailer.Area{
			Region: "Jawa Barat",
		},
	}

	dataSet := []struct {
		value interface{}
		score int
	}{
		{
			"Jawa barat",
			1,
		},
		{
			"jateng",
			-1,
		},
		{
			[]string{"Jawa barat", "jateng"},
			1,
		},
		{
			[]string{"jateng"},
			-1,
		},
		{
			nil,
			-1,
		},
	}

	scorer := AreaRegionScorer(1)

	for _, data := range dataSet {
		if got := scorer(context.Background(), ret, data.value); got != data.score {
			t.Errorf("AreaRegionScorer() = %v, want %v", got, data.score)
		}
	}

}

func TestRegionSubScorrer(t *testing.T) {
	ret := &retailer.RetailerContext{
		Area: &retailer.Area{
			RegionSub: "Bandung",
		},
	}

	dataSet := []struct {
		value interface{}
		score int
	}{
		{
			"bandung",
			1,
		},
		{
			"purwakarta",
			-1,
		},
		{
			[]string{"bandung", "Purwakarta"},
			1,
		},
		{
			[]string{"purwakarta"},
			-1,
		},
		{
			nil,
			-1,
		},
	}

	scorer := AreaRegionSubScorer(1)

	for _, data := range dataSet {
		if got := scorer(context.Background(), ret, data.value); got != data.score {
			t.Errorf("AreaRegionSubScorer() = %v, want %v", got, data.score)
		}
	}

}

func TestClusterScorrer(t *testing.T) {
	ret := &retailer.RetailerContext{
		Area: &retailer.Area{
			Cluster: "kota",
		},
	}

	dataSet := []struct {
		value interface{}
		score int
	}{
		{
			"Kota",
			1,
		},
		{
			"cimahi",
			-1,
		},
		{
			[]string{"kota", "cimahi"},
			1,
		},
		{
			[]string{"cimahi"},
			-1,
		},
		{
			nil,
			-1,
		},
	}

	scorer := AreaClusterScorer(1)

	for _, data := range dataSet {
		if got := scorer(context.Background(), ret, data.value); got != data.score {
			t.Errorf("AreaClusterScorer() = %v, want %v", got, data.score)
		}
	}

}

func TestWarehouseScorrer(t *testing.T) {
	ret := &retailer.RetailerContext{
		Area: &retailer.Area{
			Warehouse: "dago",
		},
	}

	dataSet := []struct {
		value interface{}
		score int
	}{
		{
			"Dago",
			1,
		},
		{
			"Kiaracondong",
			-1,
		},
		{
			[]string{"dago", "kiaracondong"},
			1,
		},
		{
			[]string{"kiaracondong"},
			-1,
		},
		{
			nil,
			-1,
		},
	}

	scorer := AreaWarehouseScorer(1)

	for _, data := range dataSet {
		if got := scorer(context.Background(), ret, data.value); got != data.score {
			t.Errorf("AreaWarehouseScorer() = %v, want %v", got, data.score)
		}
	}

}
