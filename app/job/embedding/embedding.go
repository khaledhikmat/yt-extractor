package jobembedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
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
	_ config.IService,
	datasvc data.IService,
	_ youtube.IService,
	_ audio.IService,
	_ storage.IService,
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

	// Retrieve unembedded videos
	videos, err = datasvc.RetrieveUnembeddedVideos(channelID, pageSize)
	if err != nil {
		errorStream <- err
		errors++
	}

	lgr.Logger.Debug("jobembedding.Processor",
		slog.String("event", "receivedVideos"),
		slog.Int("videos", len(videos)),
	)
	fmt.Printf("jobembedding.Processor - %d videos\n", len(videos))

	// Embed each video in a vector database using Firebase Function endpoint
	for _, video := range videos {
		// If the context is cancelled, exit the loop
		// But execute the defer block first
		select {
		case <-ctx.Done():
			finalState = data.JobStateCancelled
			return
		default:
		}

		lgr.Logger.Debug("jobembedding.Processor",
			slog.String("event", "aboutToEmbed"),
			slog.String("videoId", video.VideoID),
		)
		err = postToEmbeddingEndpoint(video)
		if err != nil {
			lgr.Logger.Debug("jobembedding.Processor",
				slog.String("event", "errorEmbedding"),
				slog.String("videoId", video.VideoID),
				slog.Any("Error", err),
			)
			errorStream <- err
			errors++
			continue
		}

		lgr.Logger.Debug("jobembedding.Processor",
			slog.String("event", "doneEmbedding"),
			slog.String("videoId", video.VideoID),
		)
	}

	lgr.Logger.Debug("jobembedding.Processor",
		slog.String("event", "done"),
	)
	fmt.Printf("jobembedding.Processor - done\n")
}

type embeddingVideo struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	PublishedAt      string `json:"publishedAt"`
	YTVideoURL       string `json:"ytVideoUrl"`
	VideoURL         string `json:"videoUrl"`
	AudioURL         string `json:"audioUrl"`
	TranscriptionURL string `json:"transcriptionUrl"`
}

// WARNING: Needed for the darn Firebase Function endpoint
// The endpoint expects the JSON payload to be wrapped in a "data" key
type embeddingVideoEnvelope struct {
	Data embeddingVideo `json:"data"`
}

func postToEmbeddingEndpoint(video data.Video) error {
	url := "https://storevideoembeddingfn-fhwhrucnhq-uc.a.run.app"

	// Join the strings with commas
	payload := embeddingVideoEnvelope{
		Data: embeddingVideo{
			ID:    video.VideoID,
			Title: video.Title,
			PublishedAt: func(t time.Time) string {
				formattedTime := t.Format("2006/01/02 3:04 PM")
				_, offset := t.Zone()
				offsetHours := offset / 3600
				offsetString := fmt.Sprintf("(GMT%+d)", offsetHours)
				return fmt.Sprintf("%s %s", formattedTime, offsetString)
			}(video.PublishedAt),
			YTVideoURL:       video.VideoURL,
			VideoURL:         *video.ExtractionURL,
			AudioURL:         *video.AudioURL,
			TranscriptionURL: *video.TranscriptionURL,
		},
	}
	lgr.Logger.Debug("jobembedding.Processor",
		slog.String("event", "embeddingPayload"),
		slog.String("videoId", video.VideoID),
		slog.Any("payload", payload),
	)

	// Serialize the payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("postToEmbeddingEndpoint - could not marshal payload: %w", err)
	}

	// Create the HTTP request with the JSON payload
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("postToEmbeddingEndpoint - could not create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("postToEmbeddingEndpoint - request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("postToEmbeddingEndpoint - unexpected status code: %d, response: %s", resp.StatusCode, respBody)
	}

	return nil
}
