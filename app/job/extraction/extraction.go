package jobextraction

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/khaledhikmat/yt-extractor/service"
	"github.com/khaledhikmat/yt-extractor/service/audio"
	"github.com/khaledhikmat/yt-extractor/service/cloudconvert"
	"github.com/khaledhikmat/yt-extractor/service/config"
	"github.com/khaledhikmat/yt-extractor/service/data"
	"github.com/khaledhikmat/yt-extractor/service/lgr"
	"github.com/khaledhikmat/yt-extractor/service/storage"
	"github.com/khaledhikmat/yt-extractor/service/transcription"
	"github.com/khaledhikmat/yt-extractor/service/youtube"
)

func Processor(ctx context.Context,
	channelID string,
	jobID int64,
	pageSize int,
	errorStream chan error,
	_ config.IService,
	datasvc data.IService,
	ytsvc youtube.IService,
	_ audio.IService,
	storagesvc storage.IService,
	_ cloudconvert.IService,
	_ transcription.IService) {

	// Update job state to running
	job, err := datasvc.RetrieveJobByID(jobID)
	if err != nil {
		errorStream <- err
		return
	}
	job.State = data.JobStateRunning
	err = datasvc.UpdateJob(&job)
	if err != nil {
		errorStream <- err
		return
	}

	errors := 0
	videos := []data.Video{}
	finalState := data.JobStateCompleted

	defer func() {
		// Update job state to completed
		now := time.Now()
		job.State = finalState
		job.Videos = int64(len(videos))
		job.Errors = int64(errors)
		job.CompletedAt = &now
		err = datasvc.UpdateJob(&job)
		if err != nil {
			errorStream <- err
			return
		}
	}()

	// Retrieve unextracted videos from Youtube
	if job.Type == data.JobTypeExtractionError {
		videos, err = datasvc.RetrieveExtractErroredVideos(channelID, pageSize)
	} else {
		videos, err = datasvc.RetrieveUnextractedVideos(channelID, pageSize)
	}
	if err != nil {
		errorStream <- err
		errors++
	}

	// Accumulate all video URLs
	videoURLs := []string{}
	for _, video := range videos {
		videoURLs = append(videoURLs, video.VideoURL)
	}

	// Extract videos
	results, err := ytsvc.ExtractVideos(ctx, errorStream, videoURLs)
	if err != nil {
		if err.Error() == "context canceled" {
			finalState = data.JobStateCancelled
			return
		}
		errorStream <- err
		errors++
	}

	// Update videos with extracted data
	// WARNING: Any error causes the extraction URL to be set to invalid
	// This means that extraction will re-attempted
	for _, video := range videos {
		// If the context is cancelled, exit the loop
		// But execute the defer block first
		select {
		case <-ctx.Done():
			finalState = data.JobStateCancelled
			return
		default:
		}

		var extractionURL string

		localReference, ok := results[video.VideoURL]
		if !ok {
			errorStream <- fmt.Errorf("video %s not found in extraction results", video.VideoURL)
			errors++
			extractionURL = service.InvalidURL
			updateDb(datasvc, errorStream, &video, &job, &extractionURL)
			continue
		}

		if localReference != service.InvalidURL {
			// Store the local reference video to an external storage
			extractionURL, err = storagesvc.NewFile(ctx, video.ChannelID, localReference, fmt.Sprintf("%s.mp4", video.VideoID))
			if err != nil {
				errorStream <- err
				errors++
				extractionURL = service.InvalidURL
				updateDb(datasvc, errorStream, &video, &job, &extractionURL)
				continue
			}
		} else {
			// If the video is not extracted, use the unextracted URL
			extractionURL = localReference
		}

		// WARNING: We need to store the unextracted URL to prevent cyclic extraction on the same video
		// It might be tempting to not store anything, but we need to prevent cyclic extraction

		// Update the video with the extraction URL if successful
		updateDb(datasvc, errorStream, &video, &job, &extractionURL)
	}

	lgr.Logger.Debug("jobextraction.Processor",
		slog.String("event", "done"),
	)
}

func updateDb(datasvc data.IService, errorStream chan error, video *data.Video, job *data.Job, url *string) {
	// Update the video with transcription URL
	now := time.Now()
	video.ExtractionURL = url
	video.ExtractedAt = &now
	err := datasvc.UpdateVideo(video, job.Type)
	if err != nil {
		errorStream <- err
	}
}
