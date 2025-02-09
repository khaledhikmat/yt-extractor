package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/mdobak/go-xerrors"
	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/contrib/propagators/autoprop"
	"go.opentelemetry.io/otel"
	noopmeter "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
	nooptrace "go.opentelemetry.io/otel/trace/noop"

	"github.com/khaledhikmat/yt-extractor/http"
	"github.com/khaledhikmat/yt-extractor/service/audio"
	"github.com/khaledhikmat/yt-extractor/service/config"
	"github.com/khaledhikmat/yt-extractor/service/data"
	"github.com/khaledhikmat/yt-extractor/service/lgr"
	"github.com/khaledhikmat/yt-extractor/service/storage"
	"github.com/khaledhikmat/yt-extractor/service/youtube"
)

const (
	waitOnShutdown = 4 * time.Second
)

// To determine a Youtube channel ID:
// - Visit the channel's main page.
// - Right-click anywhere and select "View Page Source."
// - Search for "channelId" in the page source (Ctrl+F or Command+F).
// const (
// 	mohmdElhamyChannelID = "UCP-PfkMcOKriSxFMH7pTxfA"
// 	maxVideos            = 10
// )

func main() {
	rootCtx := context.Background()
	canxCtx, canxFn := context.WithCancel(rootCtx)

	// Hook up a signal handler to cancel the context
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		lgr.Logger.Info(
			"received kill signal",
			slog.Any("signal", sig),
		)
		canxFn()
	}()

	// Load env vars if we are in DEV mode
	if os.Getenv("RUN_TIME_ENV") == "dev" ||
		os.Getenv("RUN_TIME_ENV") == "" {
		err := godotenv.Load()
		if err != nil {
			lgr.Logger.Error(
				"loading env vars error",
				slog.Any("error", xerrors.New(err.Error())),
			)
			return
		}
	}

	// Create Services
	configSvc := config.New()
	dataSvc := data.New(configSvc)
	youtubeSvc := youtube.New(configSvc)
	audioSvc := audio.New(configSvc)
	storageSvc := storage.New(configSvc)

	// Setup OpenTelemetry
	shutdown, err := setupOpenTelemetry(rootCtx, configSvc)
	if err != nil {
		lgr.Logger.Error(
			"setting up OpenTelemetry",
			slog.Any("error", xerrors.New(err.Error())),
		)
		return
	}
	defer func() {
		err := shutdown(rootCtx)
		if err != nil {
			lgr.Logger.Error(
				"shutting down OpenTelemetry",
				slog.Any("error", xerrors.New(err.Error())),
			)
		}
	}()

	// Print the extractor version
	_ = youtubeSvc.PrintExtractorVersion()

	// Create an error stream
	errorStream := make(chan error)
	defer close(errorStream)

	// Run the http server
	go func() {
		err = http.Run(canxCtx, errorStream, configSvc, dataSvc, youtubeSvc, audioSvc, storageSvc)
		if err != nil {
			errorStream <- err
		}
	}()

	// Wait for cancellation, completion or error
	for {
		select {
		case <-canxCtx.Done():
			lgr.Logger.Info(
				"main context cancelled",
			)
			goto resume
		case <-time.After(time.Duration(configSvc.GetExtractionPeriod()) * time.Minute):
			// Do periodic extraction requests if configured
			if configSvc.IsPeriodicExtraction() {
				job := data.Job{
					ChannelID: configSvc.GetExtractionChannelID(),
					Type:      data.JobTypeExtraction,
				}
				_, err = http.ProcessJob(canxCtx, job, 10, errorStream, configSvc, dataSvc, youtubeSvc, audioSvc, storageSvc)
				if err != nil {
					errorStream <- err
				}
			}
		case e := <-errorStream:
			// Add error table to the database
			err := dataSvc.NewError("main", e.Error())
			if err != nil {
				lgr.Logger.Error(
					"error saving error to database",
					slog.Any("error", xerrors.New(err.Error())),
				)
			}
		}
	}

	// Wait in a non-blocking way 4 seconds for all the go routines to exit
	// This is needed because the go routines may need to report as they are existing
resume:
	// Cancel the context if not already cancelled
	if canxCtx.Err() == nil {
		// Force cancel the context
		canxFn()
	}

	lgr.Logger.Info(
		"main is waiting for all go routines to exit",
	)

	// The only way to exit the main function is to wait for the shutdown
	// duration
	timer := time.NewTimer(waitOnShutdown)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			// Timer expired, proceed with shutdown
			lgr.Logger.Info(
				"main shutdown waiting period expired. Exiting now",
				slog.Duration("period", waitOnShutdown),
			)

			// Shutdown Services
			configSvc.Finalize()
			dataSvc.Finalize()
			youtubeSvc.Finalize()

			return
		case e := <-errorStream:
			// Handle error received on errorStream
			lgr.Logger.Error(
				"error received on stream",
				slog.Any("error", xerrors.New(e.Error())),
			)
		}
	}
}

// Reference:
// https://cloud.google.com/stackdriver/docs/instrumentation/setup/go
// setupOpenTelemetry sets up the OpenTelemetry SDK and exporters for metrics and
// traces. If it does not return an error, call shutdown for proper cleanup.
func setupOpenTelemetry(ctx context.Context, cfgSvc config.IService) (shutdown func(context.Context) error, err error) {
	if !cfgSvc.IsOpenTelemetry() {
		// Set Noop Tracer Provider
		otel.SetTracerProvider(nooptrace.NewTracerProvider())

		// Set Noop Meter Provider
		otel.SetMeterProvider(noopmeter.NewMeterProvider())

		// Return a no-op shutdown function
		return func(_ context.Context) error {
			return nil
		}, nil
	}

	var shutdownFuncs []func(context.Context) error

	// shutdown combines shutdown functions from multiple OpenTelemetry
	// components into a single function.
	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// Configure Context Propagation to use the default W3C traceparent format
	otel.SetTextMapPropagator(autoprop.NewTextMapPropagator())

	// Configure Trace Export to send spans as OTLP
	texporter, err := autoexport.NewSpanExporter(ctx)
	if err != nil {
		err = errors.Join(err, shutdown(ctx))
		return
	}
	tp := trace.NewTracerProvider(trace.WithBatcher(texporter))
	shutdownFuncs = append(shutdownFuncs, tp.Shutdown)
	otel.SetTracerProvider(tp)

	// Configure Metric Export to send metrics as OTLP
	mreader, err := autoexport.NewMetricReader(ctx)
	if err != nil {
		err = errors.Join(err, shutdown(ctx))
		return
	}
	mp := metric.NewMeterProvider(
		metric.WithReader(mreader),
	)
	shutdownFuncs = append(shutdownFuncs, mp.Shutdown)
	otel.SetMeterProvider(mp)

	return shutdown, nil
}
