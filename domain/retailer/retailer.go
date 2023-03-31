package retailer

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"

	"github.com/bondhan/golib/util"
)

const RetailerHeaderKey = "X-Retailer"

type ctxKey string

func (c ctxKey) String() string {
	return "retailer context " + string(c)
}

const retailerCtx = ctxKey("retailerContext")

func DecodeRetailer(retailer interface{}) (*RetailerContext, error) {
	switch r := retailer.(type) {
	case *RetailerContext:
		return r, nil
	case RetailerContext:
		return &r, nil
	default:
		var ret RetailerContext
		if err := util.DecodeJSON(retailer, &ret); err != nil {
			return nil, err
		}
		return &ret, nil
	}
}

func DecodeRetailerFromString(retailer string) (*RetailerContext, error) {
	b, err := base64.StdEncoding.DecodeString(retailer)
	if err != nil {
		return nil, err
	}

	var tmp map[string]interface{}
	if err := json.Unmarshal(b, &tmp); err != nil {
		return nil, err
	}

	return DecodeRetailer(tmp)
}

func GetRetailerFromHeaders(headers map[string]string) (*RetailerContext, error) {
	retailer, ok := headers[RetailerHeaderKey]
	if !ok {
		return nil, errors.New("[filter/retailer] retailer not found in headers")
	}

	return DecodeRetailerFromString(retailer)
}

func GetRetailerFromContext(ctx context.Context) (*RetailerContext, error) {
	if str, ok := ctx.Value(retailerCtx).(string); ok {
		return DecodeRetailerFromString(str)
	}
	return nil, errors.New("[filter/retailer] retailer not found in context")
}

func GetRawRetailerFromContext(ctx context.Context) []byte {
	if str, ok := ctx.Value(retailerCtx).(string); ok {
		b, err := base64.StdEncoding.DecodeString(str)
		if err != nil {
			return nil
		}
		return b
	}
	return nil
}

func GetEncodedRetailerFromContext(ctx context.Context) string {
	if str, ok := ctx.Value(retailerCtx).(string); ok {
		return str
	}

	return ""
}

func PropagateRetailer(ctx context.Context, headers map[string]string, key string) (context.Context, error) {
	if key == "" {
		key = RetailerHeaderKey
	}

	retailer, ok := headers[key]
	if !ok {
		return nil, errors.New("[filter/retailer] retailer not found in headers")
	}

	return context.WithValue(ctx, retailerCtx, retailer), nil
}

func WithRetailerContext(ctx context.Context, ret *RetailerContext) context.Context {
	str, err := ret.Encode()
	if err != nil {
		return ctx
	}
	return context.WithValue(ctx, retailerCtx, str)
}

func WithRetailerStringContext(ctx context.Context, ret string) context.Context {
	return context.WithValue(ctx, retailerCtx, ret)
}

type RetailerContext struct {
	ID         string            `json:"id"`
	Name       string            `json:"name,omitempty"`
	Area       *Area             `json:"area,omitempty"`
	Address    *Address          `json:"address,omitempty"`
	Segments   []string          `json:"segments,omitempty"`
	Number     string            `json:"number,omitempty"`
	User       string            `json:"user,omitempty"`
	OwnerInfo  *OwnerInfo        `json:"owner,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

type Area struct {
	Region    string `json:"region,omitempty"`
	RegionSub string `json:"regionSub,omitempty"`
	Cluster   string `json:"cluster,omitempty"`
	Warehouse string `json:"warehouse,omitempty"`
}

type Address struct {
	Province       string `json:"province,omitempty"`
	City           string `json:"city,omitempty"`
	District       string `json:"district,omitempty"`
	DistrictSub    string `json:"districtSub,omitempty"`
	ZipCode        string `json:"zipCode,omitempty"`
	Detail         string `json:"detail,omitempty"`
	Location       string `json:"location,omitempty"`
	Latitude       string `json:"latitude,omitempty"`
	Longitude      string `json:"longitude,omitempty"`
	Place          string `json:"place,omitempty"`
	RoadWidth      string `json:"roadWidth,omitempty"`
	StreetCategory string `json:"streetCategory,omitempty"`
}

type OwnerInfo struct {
	Email string `json:"email,omitempty"`
	Phone string `json:"phone,omitempty"`
	Name  string `json:"name,omitempty"`
}

func (r *RetailerContext) Encode() (string, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}
