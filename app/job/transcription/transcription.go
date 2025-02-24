package jobtranscription

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
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
	cfgsvc config.IService,
	datasvc data.IService,
	_ youtube.IService,
	audiosvc audio.IService,
	storagesvc storage.IService,
	_ cloudconvert.IService,
	transcriptionsvc transcription.IService) {
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

	// Retrieve untranscribed videos
	// Retrieve unaudioed videos from Youtube
	if job.Type == data.JobTypeTranscriptionError {
		videos, err = datasvc.RetrieveTranscribeErroredVideos(channelID, pageSize)
	} else {
		videos, err = datasvc.RetrieveUntranscribedVideos(channelID, pageSize)
	}
	if err != nil {
		errorStream <- err
		errors++
	}

	lgr.Logger.Debug("jobtranscription.Processor",
		slog.String("event", "receivedVideos"),
		slog.Int("videos", len(videos)),
	)

	// Update each videos with transcribed data
	// WARNING: Any error causes the transcription URL to be set to invalid
	// This means that transcription will be re-attempted
	for _, video := range videos {
		// If the context is cancelled, exit the loop
		// But execute the defer block first
		select {
		case <-ctx.Done():
			finalState = data.JobStateCancelled
			return
		default:
		}

		// Process a single video for audio transcription
		err := Process(ctx, &video, errorStream, job.Type, cfgsvc, datasvc, audiosvc, storagesvc, transcriptionsvc)
		if err != nil {
			errors++
			continue
		}
	}

	lgr.Logger.Debug("jobtranscription.Processor",
		slog.String("event", "done"),
	)
}

// Exported to allow the server API (i.e. server/api.go) to also call it
// upon webhook invocations
func Process(ctx context.Context,
	video *data.Video,
	errorStream chan error,
	jobType data.JobType,
	cfgsvc config.IService,
	datasvc data.IService,
	audiosvc audio.IService,
	storagesvc storage.IService,
	transcriptionsvc transcription.IService) error {
	var transcriptionURL string

	lgr.Logger.Debug("jobtranscription.Process",
		slog.String("event", "aboutToSplit"),
		slog.String("videoId", video.VideoID),
	)

	// Use the audio service to segment the audio into 10-min audio files
	// using the video audio URL so it can be easily transcribed
	localAudioFiles, err := audiosvc.SplitAudio(*video.AudioURL)
	if err != nil {
		errorStream <- err
		transcriptionURL = service.InvalidURL
		updateDb(datasvc, errorStream, video, jobType, &transcriptionURL)
		return err
	}

	lgr.Logger.Debug("jobtranscription.Process",
		slog.String("event", "aboutToTranscribe"),
		slog.String("videoId", video.VideoID),
		slog.Int("audioFiles", len(localAudioFiles)),
	)

	defer func() {
		// Delete local audio files
		for _, file := range localAudioFiles {
			// Ignore errors because they may have been deleted already
			_ = os.Remove(file)
		}
	}()

	// Use the transcription service to transcribe the segmented audio files
	var transcribedText string
	// For each segmented audio file
	for _, file := range localAudioFiles {
		// Use the transcription service to get a transcribed text
		// and delete local audio file
		txt, err := transcriptionsvc.TranscribeAudio(file)
		if err != nil {
			errorStream <- err
			transcriptionURL = service.InvalidURL
			updateDb(datasvc, errorStream, video, jobType, &transcriptionURL)
			return err
		}

		lgr.Logger.Debug("jobtranscription.Process",
			slog.String("event", "completedIndividualTranscription"),
			slog.String("videoId", video.VideoID),
			slog.String("audioFile", file),
		)

		transcribedText += txt
	}

	lgr.Logger.Debug("jobtranscription.Process",
		slog.String("event", "savingToLocalFile"),
		slog.String("videoId", video.VideoID),
	)

	// Save transcribed text in a file
	localTextFile, err := saveToFile(transcribedText, cfgsvc.GetLocalTranscriptionFolder(), fmt.Sprintf("%s.txt", video.VideoID))
	if err != nil {
		errorStream <- err
		transcriptionURL = service.InvalidURL
		updateDb(datasvc, errorStream, video, jobType, &transcriptionURL)
		return err
	}

	defer func() {
		// Delete local text file
		// Ignore errors because it may have been deleted already
		_ = os.Remove(localTextFile)
	}()

	lgr.Logger.Debug("jobtranscription.Process",
		slog.String("event", "uploadingToS3"),
		slog.String("videoId", video.VideoID),
	)

	// Upload to S3 and delete local text file
	transcriptionURL, err = storagesvc.NewFile(ctx, video.ChannelID, localTextFile, fmt.Sprintf("%s.txt", video.VideoID))
	if err != nil {
		errorStream <- err
		transcriptionURL = service.InvalidURL
		updateDb(datasvc, errorStream, video, jobType, &transcriptionURL)
		return err
	}

	lgr.Logger.Debug("jobtranscription.Process",
		slog.String("event", "updatingDb"),
		slog.String("videoId", video.VideoID),
		slog.String("audioUrl", *video.AudioURL),
		slog.String("transcriptionUrl", transcriptionURL),
	)

	// Update the video with transcription URL
	updateDb(datasvc, errorStream, video, jobType, &transcriptionURL)
	return nil
}

func saveToFile(text, folder, fileName string) (string, error) {
	// Ensure the directory exists
	err := os.MkdirAll(folder, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("failed to create directory: %v", err)
	}

	// Create the file path
	filePath := filepath.Join(folder, fileName)

	// Write the transcribed text to the file
	err = os.WriteFile(filePath, []byte(text), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write to file: %v", err)
	}

	return filePath, nil
}

func updateDb(datasvc data.IService, errorStream chan error, video *data.Video, jobType data.JobType, url *string) {
	// Update the video with transcription URL
	now := time.Now()
	video.TranscriptionURL = url
	video.TranscribedAt = &now
	err := datasvc.UpdateVideo(video, jobType)
	if err != nil {
		errorStream <- err
	}
}
