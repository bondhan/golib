package principal

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"

	"github.com/bondhan/golib/util"
)

const XPrincipalHeaderKey = "X-Principal"

type ctxKey string

func (c ctxKey) String() string {
	return "principal context " + string(c)
}

const principalCtx = ctxKey("principalContext")

func DecodePrincipal(principal interface{}) (*PrincipalContext, error) {
	switch r := principal.(type) {
	case *PrincipalContext:
		return r, nil
	case PrincipalContext:
		return &r, nil
	default:
		var ret PrincipalContext
		if err := util.DecodeJSON(principal, &ret); err != nil {
			return nil, err
		}
		return &ret, nil
	}
}

func DecodePrincipalFromString(principal string) (*PrincipalContext, error) {
	b, err := base64.StdEncoding.DecodeString(principal)
	if err != nil {
		return nil, err
	}

	var tmp map[string]interface{}
	if err := json.Unmarshal(b, &tmp); err != nil {
		return nil, err
	}

	return DecodePrincipal(tmp)
}

func GetPrincipalFromHeaders(headers map[string]string) (*PrincipalContext, error) {
	principal, ok := headers[XPrincipalHeaderKey]
	if !ok {
		return nil, errors.New("[filter/principal] principal not found in headers")
	}

	return DecodePrincipalFromString(principal)
}

func GetPrincipalFromContext(ctx context.Context) (*PrincipalContext, error) {
	if str, ok := ctx.Value(principalCtx).(string); ok {
		return DecodePrincipalFromString(str)
	}
	return nil, errors.New("[filter/principal] principal not found in context")
}

func GetEncodedPrincipalFromContext(ctx context.Context) string {
	if str, ok := ctx.Value(principalCtx).(string); ok {
		return str
	}

	return ""
}

func GetRawPrincipalFromContext(ctx context.Context) []byte {
	if str, ok := ctx.Value(principalCtx).(string); ok {
		b, err := base64.StdEncoding.DecodeString(str)
		if err != nil {
			return nil
		}
		return b
	}
	return nil
}

func PropagatePrincipal(ctx context.Context, headers map[string]string, key string) (context.Context, error) {
	if key == "" {
		key = XPrincipalHeaderKey
	}

	principal, ok := headers[key]
	if !ok {
		return nil, errors.New("[filter/principal] principal not found in headers")
	}

	return context.WithValue(ctx, principalCtx, principal), nil
}

func WithPrincipalContext(ctx context.Context, ret *PrincipalContext) context.Context {
	str, err := ret.Encode()
	if err != nil {
		return ctx
	}
	return context.WithValue(ctx, principalCtx, str)
}

func WithPrincipalStringContext(ctx context.Context, ret string) context.Context {
	return context.WithValue(ctx, principalCtx, ret)
}

type PrincipalContext struct {
	ID        string `json:"id" firestore:"id"`
	User      string `json:"user,omitempty" firestore:"user"`
	CreatedAt int    `json:"createdAt,omitempty" firestore:"createdAt"`
	UpdatedAt int    `json:"updatedAt,omitempty" firestore:"updatedAt"`
}

func (r *PrincipalContext) Encode() (string, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}
