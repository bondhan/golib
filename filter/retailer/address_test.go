package retailer

import (
	"context"
	"testing"

	"github.com/bondhan/golib/retailer"
)

func TestProvinceScorrer(t *testing.T) {
	ret := &retailer.RetailerContext{
		Address: &retailer.Address{
			Province: "Jawa Barat",
		},
	}

	dataSet := []struct {
		value interface{}
		score int
	}{
		{
			"Jawa Barat",
			1,
		},
		{
			[]string{"Jawa Tengah", "Jawa Barat"},
			1,
		},
		{
			[]string{"Jawa Tengah"},
			-1,
		},
		{
			nil,
			-1,
		},
	}

	scorer := AddressProvinceScorer(1)

	for _, data := range dataSet {
		if got := scorer(context.Background(), ret, data.value); got != data.score {
			t.Errorf("AddressProvinceScorer() = %v, want %v", got, data.score)
		}
	}

}

func TestCityScorrer(t *testing.T) {
	ret := &retailer.RetailerContext{
		Address: &retailer.Address{
			City: "Bandung",
		},
	}

	dataSet := []struct {
		value interface{}
		score int
	}{
		{
			"Bandung",
			1,
		},
		{
			"Bandung Barat",
			-1,
		},
		{
			[]string{"Bandung Barat", "bandung"},
			1,
		},
		{
			[]string{"Bandung Barat"},
			-1,
		},
		{
			nil,
			-1,
		},
	}

	scorer := AddressCityScorer(1)

	for _, data := range dataSet {
		if got := scorer(context.Background(), ret, data.value); got != data.score {
			t.Errorf("AddressCityScorer() = %v, want %v", got, data.score)
		}
	}

}

func TestDistrictScorrer(t *testing.T) {
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
			"cimahi",
			1,
		},
		{
			"kopo",
			-1,
		},
		{
			[]string{"cimahi", "kopo"},
			1,
		},
		{
			[]string{"kopo"},
			-1,
		},
		{
			nil,
			-1,
		},
	}

	scorer := AddressDistrictScorer(1)

	for _, data := range dataSet {
		if got := scorer(context.Background(), ret, data.value); got != data.score {
			t.Errorf("AddressDistrictScorer() = %v, want %v", got, data.score)
		}
	}

}

func TestDistrictSubScorrer(t *testing.T) {
	ret := &retailer.RetailerContext{
		Address: &retailer.Address{
			DistrictSub: "Sangkuriang",
		},
	}

	dataSet := []struct {
		value interface{}
		score int
	}{
		{
			"sangkuriang",
			1,
		},
		{
			"cimbeluit",
			-1,
		},
		{
			[]string{"cimbeluit", "sangkuriang"},
			1,
		},
		{
			[]string{"cimbeluit"},
			-1,
		},
		{
			nil,
			-1,
		},
	}

	scorer := AddressDistrictSubScorer(1)

	for _, data := range dataSet {
		if got := scorer(context.Background(), ret, data.value); got != data.score {
			t.Errorf("AddressDistrictSubScorer() = %v, want %v", got, data.score)
		}
	}

}

func TestZipCodeScorrer(t *testing.T) {
	ret := &retailer.RetailerContext{
		Address: &retailer.Address{
			ZipCode: "55122",
		},
	}

	dataSet := []struct {
		value interface{}
		score int
	}{
		{
			"55122",
			1,
		},
		{
			"551221",
			-1,
		},
		{
			[]string{"55122", "55123"},
			1,
		},
		{
			[]string{"551221"},
			-1,
		},
		{
			nil,
			-1,
		},
	}

	scorer := AddressZipScorer(1)

	for _, data := range dataSet {
		if got := scorer(context.Background(), ret, data.value); got != data.score {
			t.Errorf("AddressZipScorer() = %v, want %v", got, data.score)
		}
	}

}
