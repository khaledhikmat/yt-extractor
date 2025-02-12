package jobtranscription

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

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
	cloudconvertsvc cloudconvert.IService,
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
	videos, err = datasvc.RetrieveUntranscribedVideos(channelID, pageSize)
	if err != nil {
		errorStream <- err
		errors++
	}

	lgr.Logger.Debug("jobtranscription.Processor",
		slog.String("event", "receivedVideos"),
		slog.Int("videos", len(videos)),
	)

	// Update each videos with transcribed data
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
		var transcriptionURL string

		lgr.Logger.Debug("jobtranscription.Processor",
			slog.String("event", "aboutToConvert"),
			slog.String("videoId", video.VideoID),
		)

		// - Use CloudConvert service to convert MP4 (stored in S3) to MP3 (stored in S3)
		audioURL, err = cloudconvertsvc.ConvertVideoToAudio(video.ChannelID, video.VideoID)
		if err != nil {
			errorStream <- err
			errors++
			continue
		}

		lgr.Logger.Debug("jobtranscription.Processor",
			slog.String("event", "aboutToSplit"),
			slog.String("videoId", video.VideoID),
		)

		// Use the audio service to segment the audio into 10-min audio files using URL
		localAudioFiles, err := audiosvc.SplitAudio(audioURL)
		if err != nil {
			errorStream <- err
			errors++
			continue
		}

		lgr.Logger.Debug("jobtranscription.Processor",
			slog.String("event", "aboutToTranscribe"),
			slog.String("videoId", video.VideoID),
			slog.Int("audioFiles", len(localAudioFiles)),
		)

		// Use the transcription service to transcribe the segmented audio files
		var transcribedText string
		// For each segmented audio file
		for _, file := range localAudioFiles {
			// Use the transcription service to get a transcribed text and delete local audio file
			txt, err := transcriptionsvc.TranscribeAudio(file)
			if err != nil {
				errorStream <- err
				errors++
				continue
			}

			lgr.Logger.Debug("jobtranscription.Processor",
				slog.String("event", "completedIndividualTranscription"),
				slog.String("videoId", video.VideoID),
				slog.String("audioFile", file),
			)

			transcribedText += txt
		}

		lgr.Logger.Debug("jobtranscription.Processor",
			slog.String("event", "savingToLocalFile"),
			slog.String("videoId", video.VideoID),
		)

		// Save transcribed text in a file
		localTextFile, err := saveToFile(transcribedText, cfgsvc.GetLocalTranscriptionFolder(), fmt.Sprintf("%s.txt", video.VideoID))
		if err != nil {
			errorStream <- err
			errors++
			continue
		}

		lgr.Logger.Debug("jobtranscription.Processor",
			slog.String("event", "uploadingToS3"),
			slog.String("videoId", video.VideoID),
		)

		// Upload to S3 and delete local text file
		transcriptionURL, err = storagesvc.NewFile(ctx, video.ChannelID, localTextFile, fmt.Sprintf("%s.txt", video.VideoID))
		if err != nil {
			errorStream <- err
			errors++
			continue
		}

		lgr.Logger.Debug("jobtranscription.Processor",
			slog.String("event", "updatingDb"),
			slog.String("videoId", video.VideoID),
			slog.String("audioUrl", audioURL),
			slog.String("transcriptionUrl", transcriptionURL),
		)

		// Update the video with audio and transcription URLs
		now := time.Now()
		video.AudioURL = &audioURL
		video.TranscriptionURL = &transcriptionURL
		video.ProcessedAt = &now
		err = datasvc.UpdateVideo(&video, job.Type)
		if err != nil {
			errorStream <- err
			errors++
			continue
		}
	}

	lgr.Logger.Debug("jobtranscription.Processor",
		slog.String("event", "done"),
	)
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
