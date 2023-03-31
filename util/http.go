package util

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

const (
	authHeader = "Authorization"
	BODY       = "body"
	HEADER     = "header"
	QUERY      = "query"
	PARAM      = "param"
	JWT        = "jwt"
	COOKIE     = "cookie"
)

var defaultTokenCookieKey = "token"

type RequestWrapper struct {
	Method  string
	Headers map[string]string
	Body    []byte
	Params  map[string]string
	Jwt     []byte
	Query   url.Values
	RawURL  string
	Cookies map[string]string
}

type ScopeOption interface {
	Name() string
	GetValue(key string) interface{}
}

func with(scopes []interface{}, scope string) bool {
	if len(scopes) == 0 {
		return true
	}
	for _, s := range scopes {
		switch v := s.(type) {
		case ScopeOption:
			if v.Name() == scope {
				return true
			}
		default:
			if fmt.Sprintf("%v", v) == scope || fmt.Sprintf("%v", v) == "*" {
				return true
			}
		}
	}
	return false
}

func getScopeOption(scopes []interface{}, scope string) ScopeOption {
	for _, s := range scopes {
		switch v := s.(type) {
		case ScopeOption:
			if v.Name() == scope {
				return v
			}
		}
	}
	return nil
}

func CopyRequest(req *http.Request, scopes ...interface{}) (*RequestWrapper, error) {
	rw := &RequestWrapper{
		RawURL: req.URL.String(),
	}

	if with(scopes, BODY) {
		body, _ := ioutil.ReadAll(req.Body)
		if body != nil {
			req.Body = ioutil.NopCloser(bytes.NewReader(body))
		}
		rw.Body = body
	}

	if with(scopes, JWT) {
		token := ""
		th := req.Header[authHeader]
		if len(th) > 0 {
			token = th[0]
		}

		if len(token) > 7 {
			token = token[7:]
		}

		var jwt []byte

		if parts := strings.Split(token, "."); len(parts) == 3 {
			if b, err := base64.RawStdEncoding.DecodeString(parts[1]); err == nil {
				jwt = b
			}
		}
		rw.Jwt = jwt
	}

	if cookiesOpt := getScopeOption(scopes, COOKIE); cookiesOpt != nil {
		rw.Cookies = make(map[string]string)
		for _, c := range req.Cookies() {
			rw.Cookies[c.Name] = c.Value
			if tokenCookieKey := fmt.Sprintf("%v", cookiesOpt.GetValue(defaultTokenCookieKey)); tokenCookieKey != "" && tokenCookieKey == c.Name {
				var jwt []byte
				if parts := strings.Split(c.Value, "."); len(parts) == 3 {
					if b, err := base64.RawStdEncoding.DecodeString(parts[1]); err == nil {
						jwt = b
					}
				}
				rw.Jwt = jwt
			}
		}
	}

	if with(scopes, QUERY) {
		rw.Query = req.URL.Query()
	}

	if with(scopes, HEADER) {
		hs := make(map[string]string)
		for k := range req.Header {
			hs[k] = req.Header.Get(k)
		}
		rw.Headers = hs
	}

	return rw, nil
}

func (r *RequestWrapper) toEvalContext() map[string]interface{} {
	ctx := make(map[string]interface{})
	if r.Headers != nil {
		ctx[HEADER] = r.Headers
	}

	if r.Body != nil {
		var tmp map[string]interface{}
		if err := json.Unmarshal(r.Body, &tmp); err == nil {
			ctx[BODY] = tmp
		}
	}

	if r.Params != nil {
		ctx[PARAM] = r.Params
	}

	if r.Jwt != nil {
		var tmp map[string]interface{}
		if err := json.Unmarshal(r.Jwt, &tmp); err == nil {
			ctx[JWT] = tmp
		}
	}

	if r.Query != nil {
		tmp := make(map[string]string)
		for k, v := range r.Query {
			tmp[k] = v[0]
		}
		ctx[QUERY] = tmp
	}

	if r.Cookies != nil {
		ctx[COOKIE] = r.Cookies
	}
	return ctx
}

func (r *RequestWrapper) eval(expr string) (interface{}, error) {
	ctx := r.toEvalContext()
	eval, err := NewEvaluator(expr)
	if err != nil {
		return nil, err
	}
	return eval.Eval(ctx)
}

func (r *RequestWrapper) GetValue(path string) interface{} {
	if strings.HasPrefix(path, ":") {
		val, err := r.eval(strings.TrimPrefix(path, ":"))
		if err == nil {
			return val
		}
	}
	part := strings.Split(path, ".")[0]
	path = strings.TrimPrefix(path, part+".")

	toString := false
	toNum := false

	if strings.HasPrefix(path, "'") {
		path = strings.TrimPrefix(path, "'")
		toNum = true
	}

	if strings.HasPrefix(path, `"`) {
		path = strings.TrimPrefix(path, `"`)
		toString = true
	}

	val := r.getVal(part, path)
	if val == nil {
		return nil
	}

	if toString {
		return fmt.Sprintf("%v", val)
	}

	if toNum {
		sn := fmt.Sprintf("%v", val)
		if strings.Contains(sn, ".") {
			fl, err := strconv.ParseFloat(sn, 64)
			if err != nil {
				return nil
			}
			return fl
		}

		in, err := strconv.ParseInt(sn, 10, 64)
		if err != nil {
			return nil
		}
		return in
	}

	switch v := val.(type) {
	case float64:
		if v == float64(int64(v)) {
			return int64(v)
		}
		return v
	default:
		return v
	}
}

func (r *RequestWrapper) getVal(part, path string) interface{} {
	switch part {
	case BODY:
		if path == "" {
			out := make(map[string]interface{})
			if err := json.Unmarshal(r.Body, &out); err != nil {
				return nil
			}
			return out
		}
		if v := gjson.GetBytes(r.Body, path); v.Exists() {
			return v.Value()
		}
		return nil
	case JWT:
		if path == "" {
			out := make(map[string]interface{})
			if err := json.Unmarshal(r.Jwt, &out); err != nil {
				return nil
			}
			return out
		}
		if v := gjson.GetBytes(r.Jwt, path); v.Exists() {
			return v.Value()
		}
		return nil
	case PARAM:
		return r.Params[path]
	case HEADER:
		for k, v := range r.Headers {
			if strings.EqualFold(k, path) {
				return v
			}
		}
		return nil
	case QUERY:
		return r.getMultipleValues(path)
	case COOKIE:
		if v, ok := r.Cookies[path]; ok {
			return v
		}
		return nil
	default:
		return nil
	}
}

func (r *RequestWrapper) getMultipleValues(key string) interface{} {
	v := r.Query
	if v == nil {
		return nil
	}
	vs := v[key]
	if len(vs) == 0 {
		return nil
	}
	if len(vs) == 1 {
		return vs[0]
	}
	return vs
}

func (r *RequestWrapper) SetValue(path string, val interface{}) error {
	parts := strings.Split(path, ".")
	if len(parts) < 2 {
		return errors.New("invalid result path")
	}

	switch strings.ToLower(parts[0]) {
	case HEADER:
		if r.Headers == nil {
			r.Headers = make(map[string]string)
		}
		r.Headers[parts[1]] = fmt.Sprintf("%v", val)
		return nil
	case BODY:
		res, err := sjson.SetBytes(r.Body, strings.Join(parts[1:], "."), val)
		if err != nil {
			return err
		}
		r.Body = res
		return nil
	case QUERY:
		if r.Query == nil {
			r.Query = make(url.Values)
		}
		r.Query.Set(parts[1], fmt.Sprintf("%v", val))
		return nil
	case COOKIE:
		if r.Cookies == nil {
			r.Cookies = make(map[string]string)
		}
		r.Cookies[parts[1]] = fmt.Sprintf("%v", val)
		return nil
	default:
		return errors.New("invalid path")
	}
}

func (r *RequestWrapper) InjectRequest(req *http.Request, scopes ...interface{}) error {
	if with(scopes, BODY) {
		req.Body = ioutil.NopCloser(bytes.NewReader(r.Body))
	}

	if with(scopes, HEADER) {
		if req.Header == nil {
			req.Header = make(http.Header)
		}
		for k, v := range r.Headers {
			req.Header.Set(k, v)
		}
	}

	if with(scopes, QUERY) {
		req.URL.RawQuery = r.Query.Encode()
	}

	if with(scopes, COOKIE) {
		for k, v := range r.Cookies {
			req.AddCookie(&http.Cookie{Name: k, Value: v})
		}
	}

	return nil
}

func (r *RequestWrapper) GenerateRequestID(includeHeaderFields, excludeHeaderFields, excludeBodyFields []string) []byte {
	buf := &bytes.Buffer{}
	if len(r.Body) > 0 {
		body := r.Body
		if len(excludeBodyFields) > 0 {
			for _, f := range excludeBodyFields {
				b, err := sjson.DeleteBytes(body, f)
				if err == nil {
					body = b
				}
			}
		}
		buf.Write(body)
	}

	headers := r.Headers
	if len(headers) > 0 {
		for f := range headers {
			if !contains(includeHeaderFields, f) {
				delete(headers, f)
			}
		}

		if len(excludeHeaderFields) > 0 {
			for _, f := range excludeHeaderFields {
				delete(headers, f)
			}
		}

		b, err := json.Marshal(headers)
		if err == nil {
			buf.Write(b)
		}
	}

	if r.RawURL != "" {
		buf.WriteString(r.RawURL)
	}

	return Hash(buf.Bytes())
}

func contains(s []string, e string) bool {
	for _, h := range s {
		if h == e {
			return true
		}
	}
	return false
}

func GenerateRequestID(req *http.Request, excludeHeaderFields, excludeBodyFields []string) []byte {
	if req.Method == http.MethodGet {
		return nil
	}
	var body []byte
	raw, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil
	}
	if len(raw) > 0 {
		body = raw
		req.Body = ioutil.NopCloser(bytes.NewReader(raw))
	}

	head := &bytes.Buffer{}
	req.Header.Clone().Write(head)
	headers := head.Bytes()

	if len(excludeHeaderFields) > 0 {
		for _, f := range excludeHeaderFields {
			headers, err = sjson.DeleteBytes(headers, f)
			if err != nil {
				return nil
			}
		}
	}

	if len(excludeBodyFields) > 0 {
		for _, f := range excludeBodyFields {
			b, err := sjson.DeleteBytes(body, f)
			if err == nil {
				body = b
			}
		}
	}

	buf := &bytes.Buffer{}
	buf.Write(body)
	buf.Write(headers)
	buf.WriteString(req.URL.String())
	return Hash(buf.Bytes())
}

type CookieOpt struct{}

func (o *CookieOpt) Name() string {
	return COOKIE
}

func (o *CookieOpt) GetValue(key string) interface{} {
	return ""
}

type CookieTokenOpt struct {
	key string
}

func (o *CookieTokenOpt) Name() string {
	return COOKIE
}

func (o *CookieTokenOpt) GetValue(key string) interface{} {
	return o.key
}

func WithCookieScopeOpt() ScopeOption {
	return &CookieOpt{}
}

func WithCookieTokenScopeOpt(key string) ScopeOption {
	return &CookieTokenOpt{
		key: key,
	}
}
