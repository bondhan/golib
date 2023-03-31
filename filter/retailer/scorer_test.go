package retailer

import (
	"context"
	"fmt"
	"testing"

	"github.com/bondhan/golib/domain/marketing"
)

func TestMatchScore(t *testing.T) {
	data := map[string]interface{}{
		"number": "RT0001",
		"user":   "user1",
		"address": map[string]interface{}{
			"province":     "Jawa Barat",
			"city":         "Bandung",
			"district":     "Cibeunying Kaler",
			"district_sub": "Cibeunying Kidul",
			"zip_code":     "40132",
		},

		"area": map[string]interface{}{
			"region":     "Jawa Barat",
			"region_sub": "Bandung",
			"cluster":    "Kota",
			"warehouse":  "Dago",
		},
		"segments": []string{"small", "medium"},
	}

	dataSet := []struct {
		attributes map[string]interface{}
		score      int
	}{
		{
			map[string]interface{}{
				AddressProvinceKey: "Jawa Barat",
				AddressCityKey:     "Bandung",
			},
			4,
		},
		{
			map[string]interface{}{
				AddressProvinceKey: "Jawa Barat",
				AddressCityKey:     "Cimahi",
			},
			-1,
		},
	}

	for _, dt := range dataSet {
		promo := marketing.NewDummyPromo("1234")
		for k, v := range dt.attributes {
			promo.SetAttributes(k, v)
		}
		score, err := Calculate(context.Background(), data, promo)
		if err != nil {
			t.Error(err)
		}

		if score != dt.score {
			t.Error()
		}
	}
}

func TestConfigure(t *testing.T) {
	conf := map[string]int{
		AddressProvinceKey: 10,
		AddressCityKey:     20,
	}

	Configure(conf)
	data := map[string]interface{}{
		"number": "RT0001",
		"user":   "user1",
		"address": map[string]interface{}{
			"province":     "Jawa Barat",
			"city":         "Bandung",
			"district":     "Cibeunying Kaler",
			"district_sub": "Cibeunying Kidul",
			"zip_code":     "40132",
		},

		"area": map[string]interface{}{
			"region":     "Jawa Barat",
			"region_sub": "Bandung",
			"cluster":    "Kota",
			"warehouse":  "Dago",
		},
		"segments": []string{"small", "medium"},
	}

	promo := marketing.NewDummyPromo("1234")
	promo.SetAttributes(AddressProvinceKey, "Jawa barat")
	promo.SetAttributes(AddressCityKey, "bandung")

	score, err := Calculate(context.Background(), data, promo)
	if err != nil {
		t.Error(err)
	}

	if score != 30 {
		t.Error()
	}
}

func TestPreScore(t *testing.T) {
	conf := map[string]int{
		AddressProvinceKey: 10,
		AddressCityKey:     20,
	}

	Configure(conf)

	promo := marketing.NewDummyPromo("1234")
	promo.SetAttributes(AddressProvinceKey, "Jawa barat")
	promo.SetAttributes(AddressCityKey, "bandung")

	score, err := PreScore(context.Background(), promo)
	if err != nil {
		t.Error(err)
	}

	if score != 30 {
		t.Error()
	}
}

func TestMultiScore(t *testing.T) {
	Configure(defaultScore)
	data := map[string]interface{}{
		"number": "RT0001",
		"user":   "user1",
		"address": map[string]interface{}{
			"province":     "Jawa Barat",
			"city":         "Bandung",
			"district":     "Cibeunying Kaler",
			"district_sub": "Cibeunying Kidul",
			"zip_code":     "40132",
		},

		"area": map[string]interface{}{
			"region":     "Jawa Barat",
			"region_sub": "Bandung",
			"cluster":    "Kota",
			"warehouse":  "Dago",
		},
		"segments": []string{"small", "medium"},
	}

	promoSet := []map[string]interface{}{
		{
			AddressProvinceKey: []string{"Jawa Barat"},
		},
		{
			AddressProvinceKey: []string{"Jawa Barat"},
			AddressCityKey:     []string{"Bandung"},
		},
		{
			AddressProvinceKey: []string{"Jawa Barat"},
			AddressCityKey:     []string{"Cimahi"},
		},
		{
			AddressProvinceKey: []string{"Jawa Barat"},
			AddressCityKey:     []string{"Bandung"},
			AreaClusterKey:     []string{"Kota"},
		},
		{
			AddressProvinceKey: []string{"Jawa Barat"},
			AddressCityKey:     []string{"Bandung"},
			AreaRegionKey:      []string{"Jawa Barat"},
		},
		{
			SegmentsKey: []string{"medium", "big"},
		},
		{
			AreaClusterKey: []string{"Kota"},
			SegmentsKey:    []string{"medium"},
		},
	}

	promos := make([]interface{ marketing.Promo }, len(promoSet))

	for i, p := range promoSet {
		pr := marketing.NewDummyPromo(fmt.Sprintf("%v", i))
		for k, v := range p {
			pr.SetAttributes(k, v)
		}
		promos[i] = pr
	}

	pr, err := Top(context.Background(), data, promos)
	if err != nil {
		t.Error(err)
	}

	if pr.GetID() != "6" {
		t.Error("actual: ", pr.GetID())
	}

	out, err := Validates(context.Background(), data, promos, false)
	if err != nil {
		t.Error(err)
	}

	if len(out) != 6 {
		t.Error()
	}

	if out[0].GetID() != "0" {
		t.Error()
	}

	sorted, err := Validates(context.Background(), data, promos, true)
	if err != nil {
		t.Error(err)
	}

	if sorted[0].GetID() != "6" {
		t.Error("actual: ", sorted[0].GetID())
	}

}

func TestPreFilter(t *testing.T) {
	Configure(defaultScore)
	data := map[string]interface{}{
		"number": "RT0001",
		"user":   "user1",
		"address": map[string]interface{}{
			"province":     "Jawa Barat",
			"city":         "Bandung",
			"district":     "Cibeunying Kaler",
			"district_sub": "Cibeunying Kidul",
			"zip_code":     "40132",
		},

		"area": map[string]interface{}{
			"region":     "Jawa Barat",
			"region_sub": "Bandung",
			"cluster":    "Kota",
			"warehouse":  "Dago",
		},
		"segments": []string{"small", "medium"},
	}

	promoSet := []map[string]interface{}{
		{
			AddressProvinceKey: []string{"Jawa Barat"},
			"validSKU":         []string{"A1234", "B1234"},
		},
		{
			AddressProvinceKey: []string{"Jawa Barat"},
			AddressCityKey:     []string{"Bandung"},
			"validSKU":         []string{"A1234", "B1234"},
		},
		{
			AddressProvinceKey: []string{"Jawa Barat"},
			AddressCityKey:     []string{"Cimahi"},
			"validSKU":         []string{"A1234", "B1234"},
		},
		{
			AddressProvinceKey: []string{"Jawa Barat"},
			AddressCityKey:     []string{"Bandung"},
			AreaClusterKey:     []string{"Kota"},
			"validSKU":         []string{"A1234", "B1234"},
		},
		{
			AddressProvinceKey: []string{"Jawa Barat"},
			AddressCityKey:     []string{"Bandung"},
			AreaRegionKey:      []string{"Jawa Barat"},
			"validSKU":         []string{"A1233", "B1234"},
		},
		{
			SegmentsKey: []string{"medium", "big"},
			"validSKU":  []string{"A1234", "B1234"},
		},
		{
			AreaClusterKey: []string{"Kota"},
			SegmentsKey:    []string{"medium"},
			"validSKU":     []string{"A1233", "B1234"},
		},
	}

	promos := make([]interface{ marketing.Promo }, len(promoSet))

	for i, p := range promoSet {
		pr := marketing.NewDummyPromo(fmt.Sprintf("%v", i))
		for k, v := range p {
			pr.SetAttributes(k, v)
		}
		promos[i] = pr
	}

	skuFilter := func(sku string) PreFilter {
		return func(ctx context.Context, attributes map[string]interface{}) bool {
			validSKU := attributes["validSKU"].([]string)
			for _, v := range validSKU {
				if v == sku {
					return true
				}
			}
			return false
		}
	}

	pr, err := Top(context.Background(), data, promos, skuFilter("A1234"))
	if err != nil {
		t.Error(err)
	}

	if pr.GetID() != "5" {
		t.Error("actual: ", pr.GetID())
	}

	out, err := Validates(context.Background(), data, promos, false, skuFilter("A1234"))
	if err != nil {
		t.Error(err)
	}

	if len(out) != 4 {
		t.Error("actual: ", len(out))
	}

}

func BenchmarkMultiScore(t *testing.B) {
	Configure(defaultScore)
	for i := 0; i < t.N; i++ {
		data := map[string]interface{}{
			"number": "RT0001",
			"user":   "user1",
			"address": map[string]interface{}{
				"province":     "Jawa Barat",
				"city":         "Bandung",
				"district":     "Cibeunying Kaler",
				"district_sub": "Cibeunying Kidul",
				"zip_code":     "40132",
			},

			"area": map[string]interface{}{
				"region":     "Jawa Barat",
				"region_sub": "Bandung",
				"cluster":    "Kota",
				"warehouse":  "Dago",
			},
			"segments": []string{"small", "medium"},
		}

		promoSet := []map[string]interface{}{
			{
				AddressProvinceKey: []string{"Jawa Barat"},
			},
			{
				AddressProvinceKey: []string{"Jawa Barat"},
				AddressCityKey:     []string{"Bandung"},
			},
			{
				AddressProvinceKey: []string{"Jawa Barat"},
				AddressCityKey:     []string{"Cimahi"},
			},
			{
				AddressProvinceKey: []string{"Jawa Barat"},
				AddressCityKey:     []string{"Bandung"},
				AreaClusterKey:     []string{"Kota"},
			},
			{
				AddressProvinceKey: []string{"Jawa Barat"},
				AddressCityKey:     []string{"Bandung"},
				AreaRegionKey:      []string{"Jawa Barat"},
			},
			{
				SegmentsKey: []string{"medium", "big"},
			},
			{
				AreaClusterKey: []string{"Kota"},
				SegmentsKey:    []string{"medium"},
			},
		}

		promos := make([]interface{ marketing.Promo }, len(promoSet))

		for i, p := range promoSet {
			pr := marketing.NewDummyPromo(fmt.Sprintf("%v", i))
			for k, v := range p {
				pr.SetAttributes(k, v)
			}
			promos[i] = pr
		}

		pr, err := Top(context.Background(), data, promos)
		if err != nil {
			t.Error(err)
		}

		if pr.GetID() != "6" {
			t.Error("actual: ", pr.GetID())
		}

		out, err := Validates(context.Background(), data, promos, false)
		if err != nil {
			t.Error(err)
		}

		if len(out) != 6 {
			t.Error()
		}

		if out[0].GetID() != "0" {
			t.Error()
		}

		sorted, err := Validates(context.Background(), data, promos, true)
		if err != nil {
			t.Error(err)
		}

		if sorted[0].GetID() != "6" {
			t.Error("actual: ", sorted[0].GetID())
		}
	}
}
