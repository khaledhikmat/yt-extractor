package job

import (
	"context"

	"github.com/khaledhikmat/yt-extractor/service/audio"
	"github.com/khaledhikmat/yt-extractor/service/cloudconvert"
	"github.com/khaledhikmat/yt-extractor/service/config"
	"github.com/khaledhikmat/yt-extractor/service/data"
	"github.com/khaledhikmat/yt-extractor/service/storage"
	"github.com/khaledhikmat/yt-extractor/service/transcription"
	"github.com/khaledhikmat/yt-extractor/service/youtube"
)

// Signature of job processors
type Processor func(ctx context.Context,
	channelID string,
	jobId int64,
	pageSize int,
	errorStream chan error,
	cfgsvc config.IService,
	datasvc data.IService,
	ytsvc youtube.IService,
	audioSvc audio.IService,
	storagesvc storage.IService,
	cloudconvertsvc cloudconvert.IService,
	transcriptionsvc transcription.IService)
