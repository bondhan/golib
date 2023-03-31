package grpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/fullstorydev/grpcurl"
	"github.com/jhump/protoreflect/dynamic/grpcdynamic"
	"github.com/jhump/protoreflect/grpcreflect"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/trace"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	reflectpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"google.golang.org/grpc/status"
)

const (
	tracerOtel   = "otel"
	tracerCensus = "opencensus"
	contextKey   = "opencensus-request-span"
)

type ConnectionConfig struct {
	MaxMsgSize int   `json:"max_msg_size" mapstructure:"max_msg_size"`
	KeepAlive  int64 `json:"keep_alive" mapstructure:"keep_alive"`
	Timeout    int64 `json:"timeout" mapstructure:"timeout"`
	Address    string
	TLS        *TLSConfig `json:"tls" mapstructure:"tls"`
	UserAgent  string     `json:"user_agent" mapstructure:"user_agent"`
}

type TLSConfig struct {
	Authority  string `json:"authority" mapstructure:"authority"`
	Servername string `json:"server_name" mapstructure:"server_name"`
	Insecure   bool   `json:"insecure" mapstructure:"insecure"`
	CACert     string `json:"ca_cert" mapstructure:"ca_cert"`
	Cert       string `json:"cert" mapstructure:"cert"`
	Key        string `json:"key" mapstructure:"key"`
}

type DynamicClient struct {
	DescriptorPath string            `json:"descriptor_path" mapstructure:"descriptor_path"`
	DescriptorName string            `json:"descriptor_name" mapstructure:"descriptor_name"`
	Connection     *ConnectionConfig `json:"connection" mapstructure:"connection"`
	VerbosityLevel int               `json:"verbosity_level" mapstructure:"verbosity_level"`
	OmitEmptyField bool              `json:"omit_empty" mapstructure:"omit_empty"`
	Method         string
	Tracer         string
	channel        grpcdynamic.Channel
	descriptor     grpcurl.DescriptorSource
}

type Request struct {
	Payload io.Reader
	Output  io.ReadWriter
	Headers []string
	Status  *status.Status
}

func (c *DynamicClient) Dial(ctx context.Context) error {
	if c.Connection == nil {
		return errors.New("empty connection config")
	}

	reflect := false

	if c.DescriptorName == "" {
		//return errors.New("missing descriptor name")
		reflect = true
	}

	if c.DescriptorPath == "" {
		c.DescriptorPath = "."
	}

	if c.Tracer == "" {
		c.Tracer = tracerOtel
	}

	opts := []grpc.DialOption{
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	}

	if c.Tracer == tracerCensus {
		opts = []grpc.DialOption{
			grpc.WithStatsHandler(new(ocgrpc.ClientHandler)),
		}
	}

	con, err := dial(ctx, c.Connection, opts...)
	if err != nil {
		return err
	}

	if reflect {
		refClient := grpcreflect.NewClient(ctx, reflectpb.NewServerReflectionClient(con))
		c.descriptor = grpcurl.DescriptorSourceFromServer(ctx, refClient)
	} else {
		fileSource, err := grpcurl.DescriptorSourceFromProtoFiles([]string{c.DescriptorPath}, c.DescriptorName)
		if err != nil {
			return err
		}
		c.descriptor = fileSource
	}

	c.channel = con

	return nil
}

func (c *DynamicClient) Invoke(ctx context.Context, req *Request) error {
	status, err := c.invoke(ctx, req.Payload, req.Output, req.Headers)
	if status != nil {
		req.Status = status
	}
	return err
}

func (c *DynamicClient) invoke(ctx context.Context, data io.Reader, out io.ReadWriter, headers []string) (*status.Status, error) {
	switch c.Tracer {
	case tracerCensus:
		span := trace.FromContext(ctx)
		if span == nil {
			span, _ = ctx.Value(contextKey).(*trace.Span)
		}
		defer span.End()
	default:
		tracer := otel.Tracer("grpc/dclient")
		_, span := tracer.Start(ctx, "invoke")
		defer span.End()
	}

	if c.channel == nil || c.descriptor == nil {
		if err := c.Dial(ctx); err != nil {
			return nil, err
		}
	}

	options := grpcurl.FormatOptions{
		EmitJSONDefaultFields: !c.OmitEmptyField,
		IncludeTextSeparator:  false,
		AllowUnknownFields:    true,
	}

	rf, formatter, err := grpcurl.RequestParserAndFormatter(grpcurl.Format("json"), c.descriptor, data, options)
	if err != nil {
		return nil, err
	}

	h := &grpcurl.DefaultEventHandler{
		Out:            out,
		Formatter:      formatter,
		VerbosityLevel: c.VerbosityLevel,
	}

	if err := grpcurl.InvokeRPC(ctx, c.descriptor, c.channel, c.Method, headers, h, rf.Next); err != nil {
		return nil, err
	}

	return h.Status, h.Status.Err()
}

func dial(ctx context.Context, config *ConnectionConfig, dopts ...grpc.DialOption) (*grpc.ClientConn, error) {
	dialTime := 10 * time.Second
	if config.Timeout > 0 {
		dialTime = time.Duration(time.Duration(config.Timeout) * time.Second)
	}
	ctx, cancel := context.WithTimeout(ctx, dialTime)
	defer cancel()
	opts := make([]grpc.DialOption, 0)
	opts = append(opts, dopts...)
	if config.KeepAlive > 0 {
		timeout := time.Duration(time.Duration(config.KeepAlive) * time.Second)
		opts = append(opts, grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:    timeout,
			Timeout: timeout,
		}))
	}
	if config.MaxMsgSize > 0 {
		opts = append(opts, grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(config.MaxMsgSize)))
	}
	var creds credentials.TransportCredentials

	if config.TLS != nil {
		var err error
		cfg, err := grpcurl.ClientTLSConfig(config.TLS.Insecure, config.TLS.CACert, config.TLS.Cert, config.TLS.Key)
		if err != nil {
			return nil, err
		}
		creds = credentials.NewTLS(cfg)

		// can use either -servername or -authority; but not both
		if config.TLS.Servername != "" && config.TLS.Authority != "" {
			if config.TLS.Servername == config.TLS.Authority {
				fmt.Println("[WARN] Both servername and authority are present; prefer only -authority.")
			} else {
				return nil, errors.New("cannot specify different values for servername and authority")
			}
		}
		overrideName := config.TLS.Servername
		if overrideName == "" {
			overrideName = config.TLS.Authority
		}

		if overrideName != "" {
			opts = append(opts, grpc.WithAuthority(overrideName))
		}

		if config.TLS.Authority != "" {
			opts = append(opts, grpc.WithAuthority(config.TLS.Authority))
		}
	}

	grpcurlUA := "grpcurl/krakend"
	if config.UserAgent != "" {
		grpcurlUA = config.UserAgent
	}
	opts = append(opts, grpc.WithUserAgent(grpcurlUA))

	network := "tcp"

	cc, err := grpcurl.BlockingDial(ctx, network, config.Address, creds, opts...)
	if err != nil {
		return nil, err
	}
	return cc, nil
}
