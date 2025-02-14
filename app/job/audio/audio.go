package jobaudio

import (
	"context"
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
	_ youtube.IService,
	_ audio.IService,
	_ storage.IService,
	cloudconvertsvc cloudconvert.IService,
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

	// Retrieve unaudioed videos from Youtube
	if job.Type == data.JobTypeAudioError {
		videos, err = datasvc.RetrieveAudioErroredVideos(channelID, pageSize)
	} else {
		videos, err = datasvc.RetrieveUnaudioedVideos(channelID, pageSize)
	}
	if err != nil {
		errorStream <- err
		errors++
	}

	lgr.Logger.Debug("jobaudio.Processor",
		slog.String("event", "receivedVideos"),
		slog.Int("videos", len(videos)),
	)

	// Update each videos with transcribed data
	// WARNING: Any error causes the audio URL to be set to invalid
	// This means that audio will be re-attempted
	for _, video := range videos {
		// If the context is cancelled, exit the loop
		// But execute the defer block first
		select {
		case <-ctx.Done():
			finalState = data.JobStateCancelled
			return
		default:
		}

		var audioURL string

		lgr.Logger.Debug("jobaudio.Processor",
			slog.String("event", "aboutToConvert"),
			slog.String("videoId", video.VideoID),
		)

		// Use CloudConvert service to convert MP4 (stored in S3) to MP3 (stored in S3)
		audioURL, err = cloudconvertsvc.ConvertVideoToAudio(video.ChannelID, video.VideoID)
		if err != nil {
			errorStream <- err
			errors++
			audioURL = service.InvalidURL
			updateDb(datasvc, errorStream, &video, &job, &audioURL)
		}

		lgr.Logger.Debug("jobaudio.Processor",
			slog.String("event", "updatingDb"),
			slog.String("videoId", video.VideoID),
			slog.String("audioUrl", audioURL),
		)

		// Update the video with audio URL
		updateDb(datasvc, errorStream, &video, &job, &audioURL)
	}

	lgr.Logger.Debug("jobaudio.Processor",
		slog.String("event", "done"),
	)
}

func updateDb(datasvc data.IService, errorStream chan error, video *data.Video, job *data.Job, url *string) {
	// Update the video with transcription URL
	now := time.Now()
	video.AudioURL = url
	video.AudioedAt = &now
	err := datasvc.UpdateVideo(video, job.Type)
	if err != nil {
		errorStream <- err
	}
}
