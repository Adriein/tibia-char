package logger

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/adriein/tibia-char/pkg/constants"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	otellog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

func Create() (*slog.Logger, func(context.Context) error) {
	if os.Getenv(constants.Env) == constants.Dev {
		opts := &slog.HandlerOptions{
			Level: slog.LevelDebug,
			ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
				if attr.Key == slog.TimeKey {
					formatted := attr.Value.Time().UTC().Format(time.DateTime)

					return slog.String(slog.TimeKey, formatted)
				}

				return attr
			},
		}

		return slog.New(slog.NewTextHandler(os.Stdout, opts)), func(ctx context.Context) error { return nil }
	}

	ctx := context.Background()

	exporter, err := otlploghttp.New(ctx,
		otlploghttp.WithEndpoint("eu.i.posthog.com"),
		otlploghttp.WithURLPath("/i/v1/logs"),
		otlploghttp.WithHeaders(map[string]string{
			"Authorization": "Bearer phc_vvgsk4o4CHgSnoexamQcDrHWQViTesuKoPLhMNeCqkqk",
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
		),
	)

	loggerProvider := otellog.NewLoggerProvider(
		otellog.WithProcessor(otellog.NewBatchProcessor(exporter)),
		otellog.WithProcessor(otellog.NewSimpleProcessor(stdoutExporter)),
		otellog.WithResource(res),
	)

	return otelslog.NewLogger("", otelslog.WithLoggerProvider(loggerProvider)), loggerProvider.Shutdown
}
