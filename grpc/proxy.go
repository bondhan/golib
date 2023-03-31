package grpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"
	"time"

	mw "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	prom "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/soheilhy/cmux"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/contrib/propagators/opencensus"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type Proxy struct {
	Server      *grpc.Server
	Mux         *http.ServeMux
	GWMux       *runtime.ServeMux
	httpAddress string
}

type proxyOption struct {
	recordPath        string
	httpAddress       string
	unaryInterceptors []grpc.UnaryServerInterceptor
}

type ProxyOpt func(*proxyOption)

func WithRecorder(path string) ProxyOpt {
	return func(po *proxyOption) {
		po.recordPath = path
	}
}

func WithHttpPort(host string, port int) ProxyOpt {
	if host == "" {
		host = "0.0.0.0"
	}
	return func(po *proxyOption) {
		po.httpAddress = fmt.Sprintf("%s:%d", host, port)
	}
}

func WithMiddleware(mw ...grpc.UnaryServerInterceptor) ProxyOpt {
	return func(po *proxyOption) {
		po.unaryInterceptors = mw
	}
}

func New(opts ...ProxyOpt) *Proxy {
	popt := &proxyOption{}

	for _, o := range opts {
		o(popt)
	}

	reg := prometheus.NewRegistry()

	// Create some standard server metrics.
	grpcMetrics := prom.NewServerMetrics()
	reg.MustRegister(grpcMetrics)

	prop := propagation.NewCompositeTextMapPropagator(b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader)), propagation.TraceContext{}, propagation.Baggage{}, opencensus.Binary{})

	customFunc := func(p interface{}) (err error) {
		log.Print(string(debug.Stack()))
		return status.Errorf(codes.Unknown, "panic triggered: %v", p)
	}
	recOpts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(customFunc),
	}
	uchain := []grpc.UnaryServerInterceptor{
		otelgrpc.UnaryServerInterceptor(otelgrpc.WithPropagators(prop)),
		prom.UnaryServerInterceptor,
		grpc_recovery.UnaryServerInterceptor(recOpts...),
	}

	if popt.recordPath != "" {
		uchain = append(uchain, NewRecorderServerInterceptor(popt.recordPath))
	}

	if len(popt.unaryInterceptors) > 0 {
		uchain = append(uchain, popt.unaryInterceptors...)
	}

	schain := []grpc.StreamServerInterceptor{
		otelgrpc.StreamServerInterceptor(otelgrpc.WithPropagators(prop)),
		prom.StreamServerInterceptor,
		grpc_recovery.StreamServerInterceptor(recOpts...),
	}

	srv := grpc.NewServer(
		grpc.StreamInterceptor(mw.ChainStreamServer(schain...)),
		grpc.UnaryInterceptor(mw.ChainUnaryServer(uchain...)),
	)

	grpcMetrics.InitializeMetrics(srv)
	reflection.Register(srv)

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.Handle("/metrics", promhttp.Handler())
	gwmux := runtime.NewServeMux()
	if popt.httpAddress == "" {
		mux.Handle("/", otelhttp.NewHandler(gwmux, "gRPC"))
	}

	return &Proxy{
		Server:      srv,
		Mux:         mux,
		GWMux:       gwmux,
		httpAddress: popt.httpAddress,
	}
}

func (s *Proxy) Serve(host string, port int) {
	if host == "" {
		host = "0.0.0.0"
	}

	if port == 0 {
		port = 50050
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	listening, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	httpAddr := s.httpAddress

	httpL, err := net.Listen("tcp", s.httpAddress)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	grpcL := listening

	var m cmux.CMux
	if s.httpAddress == "" {
		m = cmux.New(listening)
		// Create a grpc listener first
		grpcL = m.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))
		httpL = m.Match(cmux.Any())
		httpAddr = addr
	}

	srv := &http.Server{
		Addr:    httpAddr,
		Handler: s.Mux,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		<-ctx.Done()
		fmt.Println("shutting down gracefully")

		timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer func() {
			stop()
			cancel()
			wg.Done()
		}()

		if err := srv.Shutdown(timeoutCtx); err != nil {
			fmt.Println(err)
		}
		s.Server.GracefulStop()
		fmt.Println("shutdown completed")
	}()

	if httpAddr != addr {
		fmt.Printf("\nlistening http on %s\n", httpAddr)
	}
	go srv.Serve(httpL)
	go s.Server.Serve(grpcL)

	if m != nil {
		m.Serve()
	}

	wg.Wait()
}
