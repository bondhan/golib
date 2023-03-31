package tracing

import (
	"context"
	"errors"
	"os"
	"strings"

	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/contrib/propagators/opencensus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	oteljaeger "go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	stdout "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"

	gconf "github.com/bondhan/golib/config"
	"github.com/bondhan/golib/log"
	"github.com/bondhan/golib/util"
)

type TracerConfig struct {
	JaegerAgentHost    string  `json:"jaeger_agent_host"`
	JaegerAgentPort    string  `json:"jaeger_agent_port"`
	JaegerCollectorURL string  `json:"jaeger_url"`
	JaegerMode         string  `json:"jaeger_mode"`
	NewRelicKey        string  `json:"newrelic_apikey"`
	NewRelicURL        string  `json:"newrelic_url"`
	OtelMode           string  `json:"otel_agent_mode"`
	OtelEndpoint       string  `json:"otel_agent_endpoint"`
	SampleRate         float64 `json:"tracer_sample_rate"`
}

func GetProvider(kind, name string, conf interface{}) (trace.TracerProvider, error) {
	logger := log.GetLogger(context.Background(), "tracing", "GetProvider")
	var cfg *TracerConfig

	if conf == nil {
		if err := gconf.EnvToStruct(&cfg); err != nil {
			return nil, err
		}
	}

	switch cf := conf.(type) {
	case TracerConfig:
		cfg = &cf
	case *TracerConfig:
		cfg = cf
	case gconf.Getter:
		cfg = new(TracerConfig)
		if err := cf.Unmarshal(cfg); err != nil {
			return nil, err
		}
	default:
		cfg = new(TracerConfig)
		if err := util.DecodeJSON(conf, cfg); err != nil {
			return nil, err
		}
	}

	if cfg == nil && kind != "no-op" {
		return nil, errors.New("missing configuration")
	}

	var provider trace.TracerProvider
	var exporter sdk.SpanExporter
	switch kind {
	case "newrelic":
		secure := strings.HasPrefix(cfg.NewRelicURL, "https")
		opts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(cfg.NewRelicURL),
			otlptracegrpc.WithHeaders(map[string]string{
				"api-key": cfg.NewRelicKey,
			}),
		}
		if !secure {
			opts = append(opts, otlptracegrpc.WithInsecure())
		}
		driver := otlptracegrpc.NewClient(opts...)
		exp, err := otlptrace.New(context.Background(), driver)
		if err != nil {
			logger.Errorf("creating OTLP trace exporter: %v", err)
			return nil, err
		}
		exporter = exp
		logger.Info("using new relic exporter")
	case "stdout":
		otExporter, err := stdout.New(stdout.WithPrettyPrint())
		if err != nil {
			break
		}
		exporter = otExporter
		logger.Info("using stdout exporter")
	case "jaeger":
		switch cfg.JaegerMode {
		case "collector":
			pr, err := oteljaeger.New(oteljaeger.WithCollectorEndpoint(oteljaeger.WithEndpoint(cfg.JaegerCollectorURL)))
			if err != nil {
				return nil, err
			}
			exporter = pr
			logger.Info("using jaeger collector exporter")
		default:
			opts := []oteljaeger.AgentEndpointOption{
				oteljaeger.WithAgentHost(cfg.JaegerAgentHost),
				oteljaeger.WithAgentPort(cfg.JaegerAgentPort),
			}
			pr, err := oteljaeger.New(oteljaeger.WithAgentEndpoint(opts...))
			if err != nil {
				return nil, err
			}
			exporter = pr
			logger.Info("using jaeger agent exporter")
		}
	case "otel":
		secure := strings.HasPrefix(cfg.OtelEndpoint, "https")
		var driver otlptrace.Client
		switch cfg.OtelMode {
		case "http":
			opts := []otlptracehttp.Option{otlptracehttp.WithEndpoint(cfg.OtelEndpoint)}
			if !secure {
				opts = append(opts, otlptracehttp.WithInsecure())
			}
			driver = otlptracehttp.NewClient(opts...)
			logger.Info("using OTLP http exporter")
		case "grpc":
			opts := []otlptracegrpc.Option{otlptracegrpc.WithEndpoint(cfg.OtelEndpoint)}
			if !secure {
				opts = append(opts, otlptracegrpc.WithInsecure())
			}
			driver = otlptracegrpc.NewClient(opts...)
			logger.Info("using OTLP GRPC exporter")
		default:
			return nil, errors.New("unsupported mode")
		}

		exp, err := otlptrace.New(context.Background(), driver)
		if err != nil {
			logger.Errorf("creating OTLP trace exporter: %v", err)
			return nil, err
		}
		exporter = exp
	default:
		logger.Info("using NooP tracer ")
		provider = trace.NewNoopTracerProvider()
		otel.SetTracerProvider(provider)
		return provider, nil
	}

	rate := float64(1)
	sampler := sdk.AlwaysSample()
	if cfg.SampleRate > 0 {
		rate = cfg.SampleRate
		sampler = sdk.TraceIDRatioBased(rate)
	}

	tp := sdk.NewTracerProvider(
		// Always be sure to batch in production.
		sdk.WithBatcher(exporter),
		// Record information about this application in an Resource.
		sdk.WithSampler(sampler),
		sdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(name),
			attribute.String("environment", os.Getenv("ENV")),
			attribute.Int64("ID", 1),
		)),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader)), propagation.TraceContext{}, propagation.Baggage{}, opencensus.Binary{}))
	return tp, nil
}
