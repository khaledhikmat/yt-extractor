package server

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/khaledhikmat/yt-extractor/service/audio"
	"github.com/khaledhikmat/yt-extractor/service/cloudconvert"
	"github.com/khaledhikmat/yt-extractor/service/config"
	"github.com/khaledhikmat/yt-extractor/service/data"
	"github.com/khaledhikmat/yt-extractor/service/lgr"
	"github.com/khaledhikmat/yt-extractor/service/storage"
	"github.com/khaledhikmat/yt-extractor/service/transcription"
	"github.com/khaledhikmat/yt-extractor/service/youtube"
	"github.com/mdobak/go-xerrors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

type ginWithContext func(canxCtx context.Context, errorStream chan error) error

var (
	meter = otel.Meter(fmt.Sprintf("yt.extractor.%s.server", os.Getenv("APP_NAME")))

	invocationCounter metric.Int64Counter
)

func init() {
	var err error
	invocationCounter, err = meter.Int64Counter(
		fmt.Sprintf("yt.extractor.%s.server.invocation.counter", os.Getenv("APP_NAME")),
		metric.WithDescription(fmt.Sprintf("The number of %s server invocations", os.Getenv("APP_NAME"))),
		metric.WithUnit("1"),
	)
	if err != nil {
		lgr.Logger.Error(
			"creating counter",
			slog.Any("error", xerrors.New(err.Error())),
		)
	}
}

func Run(canxCtx context.Context,
	errorStream chan error,
	cfgsvc config.IService,
	datasvc data.IService,
	ytsvc youtube.IService,
	audiosvc audio.IService,
	storagesvc storage.IService,
	cloudconvertsvc cloudconvert.IService,
	transcriptionsvc transcription.IService) error {
	// Setup the Gin router
	r := gin.Default()
	cfg := cors.DefaultConfig()
	cfg.AllowAllOrigins = true
	cfg.AllowCredentials = true
	cfg.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "OPTIONS", "DELETE"}
	cfg.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	r.Use(cors.New(cfg))

	// TODO: Add middleware to handle API key authentication

	// Setup home routes
	// TODO: Add routes

	// Setup API routes
	apiRoutes(canxCtx, r, errorStream, cfgsvc, datasvc, ytsvc, audiosvc, storagesvc, cloudconvertsvc, transcriptionsvc)

	fn := getRunWithCanxFn(r, ":"+cfgsvc.GetAPIPort())
	return fn(canxCtx, errorStream)
}

func getRunWithCanxFn(r *gin.Engine, port string) ginWithContext {
	return func(canxCtx context.Context, errorStream chan error) error {
		go func() {
			if err := r.Run(port); err != nil {
				errorStream <- fmt.Errorf("error runing gin: %v", err)
				return
			}
		}()

		// Wait until cancelled
		<-canxCtx.Done()
		return canxCtx.Err()
	}
}
