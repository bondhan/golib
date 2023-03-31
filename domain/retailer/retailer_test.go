package retailer

import (
	"context"
	"testing"
)

func TestDecode(t *testing.T) {
	data := map[string]interface{}{
		"id":     "1234566",
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

	ret, err := DecodeRetailer(data)
	if err != nil {
		t.Error(err)
	}
	if ret.Number != data["number"] {
		t.Error()
	}

	if ret.Address.City != "Bandung" {
		t.Error()
	}
}

func TestDecodeEncode(t *testing.T) {
	data := map[string]interface{}{
		"id":     "1234566",
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

	ret, err := DecodeRetailer(data)
	if err != nil {
		t.Error(err)
	}

	encoded, err := ret.Encode()
	if err != nil {
		t.Error(err)
	}

	headers := map[string]string{
		RetailerHeaderKey: encoded,
	}

	dret, err := GetRetailerFromHeaders(headers)
	if err != nil {
		t.Error(err)
	}

	if dret.Number != data["number"] {
		t.Error()
	}

	if dret.Address.City != "Bandung" {
		t.Error()
	}
}

func TestContext(t *testing.T) {
	data := map[string]interface{}{
		"id":     "1234566",
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

	ret, err := DecodeRetailer(data)
	if err != nil {
		t.Error(err)
	}

	ctx := context.Background()

	ctx = WithRetailerContext(ctx, ret)

	dret, err := GetRetailerFromContext(ctx)
	if err != nil {
		t.Error(err)
	}

	if dret.Number != data["number"] {
		t.Error()
	}

	if dret.Address.City != "Bandung" {
		t.Error()
	}
}
