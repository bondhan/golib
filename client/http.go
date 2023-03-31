package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	nurl "net/url"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
)

type HttpRequest struct {
	Method  string
	URL     string
	Payload interface{}
	Headers map[string]string
}

type HttpResponse struct {
	Status  int
	Body    []byte
	Error   error
	Headers http.Header
}

func (r *HttpResponse) Payload(doc interface{}) error {
	if r.Error != nil {
		return r.Error
	}

	if r.Body == nil {
		return errors.New("[client/http] empty body result")
	}

	return json.Unmarshal(r.Body, doc)
}

func (h *HttpRequest) Get(ctx context.Context) (*HttpResponse, error) {
	return send(ctx, "GET", h.URL, nil, h.Headers)
}

func (h *HttpRequest) Delete(ctx context.Context) (*HttpResponse, error) {
	return send(ctx, "DELETE", h.URL, nil, h.Headers)
}

func (h *HttpRequest) Post(ctx context.Context) (*HttpResponse, error) {
	return send(ctx, "POST", h.URL, h.Payload, h.Headers)
}

func (h *HttpRequest) Put(ctx context.Context) (*HttpResponse, error) {
	return send(ctx, "PUT", h.URL, h.Payload, h.Headers)
}

func (h *HttpRequest) Patch(ctx context.Context) (*HttpResponse, error) {
	return send(ctx, "PATCH", h.URL, h.Payload, h.Headers)
}

func (h *HttpRequest) Submit(ctx context.Context) (*HttpResponse, error) {
	if h.Method == "" {
		return nil, errors.New("[client/http] missing method")
	}
	return send(ctx, strings.ToUpper(h.Method), h.URL, h.Payload, h.Headers)
}

func send(ctx context.Context, method, url string, payload interface{}, headers map[string]string) (*HttpResponse, error) {
	client := &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport, otelhttp.WithPropagators(propagation.TraceContext{}))}

	var b *bytes.Buffer
	if payload != nil {
		var bp []byte
		switch p := payload.(type) {
		case string:
			bp = []byte(p)
		case []byte:
			bp = p
		default:
			o, err := json.Marshal(payload)
			if err != nil {
				return nil, err
			}
			bp = o
		}

		b = bytes.NewBuffer(bp)
	}

	host := "svc"
	uri, err := nurl.Parse(url)
	if err != nil {
		host = uri.Hostname()
	}
	tr := otel.Tracer("http/client")
	ctx, span := tr.Start(ctx, method+" "+url, trace.WithAttributes(semconv.PeerServiceKey.String(host)))
	defer span.End()
	var req *http.Request
	if b == nil {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
		if err != nil {
			return nil, err
		}
	} else {
		req, err = http.NewRequestWithContext(ctx, method, url, b)
		if err != nil {
			return nil, err
		}
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	rsp := &HttpResponse{
		Status:  resp.StatusCode,
		Headers: resp.Header,
	}

	if resp.StatusCode >= 300 {
		err := errors.New(string(data))
		span.RecordError(err)
		rsp.Error = err
		return rsp, err
	}

	rsp.Body = data

	return rsp, nil
}
