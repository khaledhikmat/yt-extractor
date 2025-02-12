package jobextraction

import (
	"context"
	"fmt"
	"time"

	"github.com/khaledhikmat/yt-extractor/service"
	"github.com/khaledhikmat/yt-extractor/service/audio"
	"github.com/khaledhikmat/yt-extractor/service/cloudconvert"
	"github.com/khaledhikmat/yt-extractor/service/config"
	"github.com/khaledhikmat/yt-extractor/service/data"
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
	if job.Type == data.JobTypeErroredExtraction {
		videos, err = datasvc.RetrieveErroredVideos(channelID, pageSize)
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
	for _, video := range videos {
		// If the context is cancelled, exit the loop
		// But execute the defer block first
		select {
		case <-ctx.Done():
			finalState = data.JobStateCancelled
			return
		default:
		}

		localReference, ok := results[video.VideoURL]
		if !ok {
			continue
		}

		var url string
		if localReference != service.UnextractedVideoURL {
			// Store the local reference video to an external storage
			url, err = storagesvc.NewFile(ctx, video.ChannelID, localReference, fmt.Sprintf("%s.mp4", video.VideoID))
			if err != nil {
				errorStream <- err
				errors++
				continue
			}
		} else {
			// If the video is not extracted, use the unextracted URL
			url = localReference
		}

		// Update the video with the extraction URL
		now := time.Now()
		video.ExtractionURL = &url
		video.ExtractedAt = &now
		err = datasvc.UpdateVideo(&video, job.Type)
		if err != nil {
			errorStream <- err
			errors++
			continue
		}
	}
}
