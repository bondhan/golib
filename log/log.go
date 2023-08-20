package log

import (
	"context"
	"fmt"

	"path"
	"runtime"
	"strings"
	"time"

	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
	"github.com/snowzach/rotatefilehook"
	"github.com/uptrace/opentelemetry-go-extra/otellogrus"
	"go.opentelemetry.io/otel/trace"
)

var (
	logger *logrus.Logger
)

func init() {
	if logger == nil {
		logger = logrus.New()
	}
}

const Default = "default"

type DefaultFieldHook struct {
	fields map[string]interface{}
}

func (h *DefaultFieldHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *DefaultFieldHook) Fire(e *logrus.Entry) error {
	for i, v := range h.fields {
		e.Data[i] = v
	}
	return nil
}

type LogConfig struct {
	IsProduction bool
	IsJSON       bool
	LogFileName  string
	Fields       map[string]interface{}
}

type LogOption func(*LogConfig)

func LogToFile(fileName string) LogOption {
	return func(o *LogConfig) {
		o.LogFileName = fileName
	}
}

func IsProduction(isProd bool) LogOption {
	return func(o *LogConfig) {
		o.IsProduction = isProd
	}
}

func IsJSONFormatter(enable bool) LogOption {
	return func(o *LogConfig) {
		o.IsJSON = enable
	}
}

func LogAdditionalFields(fields map[string]interface{}) LogOption {
	return func(o *LogConfig) {
		o.Fields = fields
	}
}

// NewLogInstance ...
func NewLogInstance(logOptions ...LogOption) *logrus.Logger {
	var level logrus.Level

	//default configuration
	lc := &LogConfig{}
	lc.LogFileName = Default

	for _, opt := range logOptions {
		opt(lc)
	}

	//if it is production will output warn and error level
	if lc.IsProduction {
		level = logrus.WarnLevel
	} else {
		level = logrus.TraceLevel
	}

	logger.SetLevel(level)
	logger.SetOutput(colorable.NewColorableStdout())

	if lc.IsJSON {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
			PrettyPrint:     true,
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				s := strings.Split(f.Function, ".")
				funcname := s[len(s)-1]
				_, filename := path.Split(f.File)
				return funcname, filename
			},
		})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: time.RFC3339,
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				s := strings.Split(f.Function, ".")
				funcname := s[len(s)-1]
				_, filename := path.Split(f.File)
				return funcname, filename
			},
		})
	}

	enableLogFile := false
	if lc.LogFileName != "" {
		enableLogFile = true
	}

	if enableLogFile {
		dt := time.Now().UTC()
		rotateFileHook, err := rotatefilehook.NewRotateFileHook(rotatefilehook.RotateFileConfig{
			Filename:   "logs/" + dt.Format("20060102") + "_" + lc.LogFileName,
			MaxSize:    50, // megabytes
			MaxBackups: 3,
			MaxAge:     28, //days
			Level:      level,
			Formatter: &logrus.JSONFormatter{
				TimestampFormat: time.RFC3339,
				CallerPrettyfier: func(f *runtime.Frame) (string, string) {
					s := strings.Split(f.Function, ".")
					funcname := s[len(s)-1]
					_, filename := path.Split(f.File)
					return funcname, filename
				},
			},
		})

		if err != nil {
			logger.Fatalf("Failed to initialize file rotate hook: %v", err)
		}

		logger.AddHook(rotateFileHook)
	}

	logger.AddHook(&DefaultFieldHook{lc.Fields})

	return logger
}

func GetLogger(ctx context.Context, pkg, fnName string) *logrus.Entry {

	_, file, _, _ := runtime.Caller(1)
	file = file[strings.LastIndex(file, "/")+1:]

	fields := logrus.Fields{
		"function": fnName,
		"package":  pkg,
		"source":   file,
		"level":    GetLevel(),
	}
	span := trace.SpanFromContext(ctx)
	if span != nil {
		if span.SpanContext().HasSpanID() {
			fields["span_id"] = span.SpanContext().SpanID().String()
		}
		if span.SpanContext().HasTraceID() {
			fields["trace_id"] = span.SpanContext().TraceID().String()
		}
	}

	return WithContext(ctx).WithFields(fields)
}

func SetLevel(level string) {
	switch level {
	case "panic":
		logrus.SetLevel(logrus.PanicLevel)
	case "fatal":
		logrus.SetLevel(logrus.FatalLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	case "warning":
		logrus.SetLevel(logrus.WarnLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}
}

func WithContext(ctx context.Context) *logrus.Entry {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "<???>"
		line = 1
	} else {
		slash := strings.LastIndex(file, "/")
		file = file[slash+1:]
	}
	logger.SetReportCaller(true)
	return logger.WithContext(ctx).WithField("source", fmt.Sprintf("%s:%d", file, line))
}

func GetLevel() string {
	return strings.ToUpper(logger.GetLevel().String())
}

func Configure(format, level string, sensitiveFields ...string) {

	switch strings.ToLower(format) {
	case "json":
		logrus.SetFormatter(&logrus.JSONFormatter{})
	case "safe_json":
		if len(sensitiveFields) == 0 {
			sensitiveFields = []string{"password", "passwd", "pass", "secret", "token"}
		}
		logrus.SetFormatter(&SafeJSONFormatter{sensitiveFields: sensitiveFields})
	default:
		logrus.SetFormatter(&logrus.TextFormatter{})
	}

	lvl := strings.ToLower(level)
	SetLevel(lvl)
	levels := []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
	}

	if lvl == "debug" {
		levels = append(levels, logrus.DebugLevel)
	}

	logrus.AddHook(otellogrus.NewHook(otellogrus.WithLevels(levels...)))
}
