package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/adriein/tibia-char/pkg/constants"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/log"
	otellog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

type SeverityProcessor struct {
	otellog.Processor
	Level log.Severity
}

func (p *SeverityProcessor) OnEmit(ctx context.Context, record *otellog.Record) error {
	if record.Severity() < p.Level {
		return nil
	}

	return p.Processor.OnEmit(ctx, record)
}

func (p *SeverityProcessor) Enabled(ctx context.Context, param otellog.EnabledParameters) bool {
	if param.Severity < p.Level {
		return false
	}

	return p.Processor.Enabled(ctx, param)
}

func Create() (*slog.Logger, func(context.Context) error) {
	if os.Getenv(constants.Env) == constants.Dev {
		opts := &slog.HandlerOptions{
			Level: slog.LevelDebug,
			ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
				if attr.Key == slog.TimeKey {
					formatted := attr.Value.Time().Format(time.DateTime)

					return slog.String(slog.TimeKey, formatted)
				}

				return attr
			},
		}

		return slog.New(slog.NewTextHandler(os.Stdout, opts)), func(ctx context.Context) error { return nil }
	}

	ctx := context.Background()

	httpExporter, err := otlploghttp.New(ctx,
		otlploghttp.WithEndpoint("eu.i.posthog.com"),
		otlploghttp.WithURLPath("/i/v1/logs"),
		otlploghttp.WithHeaders(map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", os.Getenv(constants.PosthogSdkApiKey)),
		}),
	)

	if err != nil {
		slog.Error("Error creating OTEL exporter")
		os.Exit(1)
	}

	stdoutExporter, _ := stdoutlog.New()

	res, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("tibia-char"),
			semconv.ServiceVersionKey.String(os.Getenv(constants.ImgVersion)),
			semconv.DeploymentEnvironmentNameKey.String(os.Getenv(constants.Env)),
		),
	)

	levelProcessor := &SeverityProcessor{
		Processor: otellog.NewBatchProcessor(httpExporter),
		Level:     log.SeverityInfo,
	}

	loggerProvider := otellog.NewLoggerProvider(
		otellog.WithProcessor(levelProcessor),
		otellog.WithProcessor(otellog.NewSimpleProcessor(stdoutExporter)),
		otellog.WithResource(res),
	)

	return otelslog.NewLogger("", otelslog.WithLoggerProvider(loggerProvider)), loggerProvider.Shutdown
}
