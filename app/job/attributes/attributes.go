package jobattributes

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/khaledhikmat/yt-extractor/service/audio"
	"github.com/khaledhikmat/yt-extractor/service/cloudconvert"
	"github.com/khaledhikmat/yt-extractor/service/config"
	"github.com/khaledhikmat/yt-extractor/service/data"
	"github.com/khaledhikmat/yt-extractor/service/lgr"
	"github.com/khaledhikmat/yt-extractor/service/storage"
	"github.com/khaledhikmat/yt-extractor/service/transcription"
	"github.com/khaledhikmat/yt-extractor/service/youtube"
	"github.com/khaledhikmat/yt-extractor/utils"
)

func Processor(ctx context.Context,
	channelID string,
	jobID int64,
	pageSize int,
	errorStream chan error,
	cfgsvc config.IService,
	datasvc data.IService,
	ytsvc youtube.IService,
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
	ytvideos := []youtube.Video{}
	finalState := data.JobStateCompleted

	defer func() {
		// Update job state to completed
		now := time.Now()
		job.State = finalState
		job.Videos = int64(len(ytvideos))
		job.Errors = int64(errors)
		job.CompletedAt = &now
		err = datasvc.UpdateJob(&job)
		if err != nil {
			errorStream <- err
			return
		}
	}()

	// Retrieve videos from Youtube
	ytvideos, err = ytsvc.RetrieveVideos(channelID, pageSize)
	if err != nil {
		errorStream <- err
		errors++
	}

	// Insert/update videos into the database
	insertedIDs := []int64{}
	for _, ytvideo := range ytvideos {

		// If the context is cancelled, exit the loop
		// But execute the defer block first
		select {
		case <-ctx.Done():
			finalState = data.JobStateCancelled
			return
		default:
		}

		video := data.Video{
			ChannelID: channelID,
			VideoID:   ytvideo.ID,
			VideoURL:  ytvideo.URL,
			Title:     ytvideo.Title,
			PublishedAt: func() time.Time {
				parsedTime, _ := time.Parse(time.RFC3339, ytvideo.PublishedAt)
				return parsedTime
			}(),
			Views: func() int64 {
				views, _ := strconv.ParseInt(ytvideo.Views, 10, 64)
				return views
			}(),
			Comments: func() int64 {
				comments, _ := strconv.ParseInt(ytvideo.Comments, 10, 64)
				return comments
			}(),
			Likes: func() int64 {
				likes, _ := strconv.ParseInt(ytvideo.Likes, 10, 64)
				return likes
			}(),
			Duration: utils.ExtractDurationInSecs(ytvideo.Duration),
			Short:    ytvideo.Short,
		}

		// Insert or update the video into the database
		insert, id, err := datasvc.NewVideo(video)
		if err != nil {
			errorStream <- err
			errors++
			continue
		}

		// If the video was inserted, add the ID to the list so we can notify the automation webhook
		if insert {
			insertedIDs = append(insertedIDs, id)
		}
	}

	// Call automation webhook to convey that we have new videos
	if len(insertedIDs) > 0 {
		err = postToAutomationWebhook(insertedIDs, cfgsvc.GetAutomationWebhookURL())
		if err != nil {
			errorStream <- err
		}
	}

	lgr.Logger.Debug("jobattributes.Processor",
		slog.String("event", "done"),
	)
}

func postToAutomationWebhook(insertIDs []int64, url string) error {
	if url == "" {
		return fmt.Errorf("postToAutomationWebhook - automation webhook URL is empty")
	}

	// Send the request to Make webhook endpoint
	strIDs := make([]string, len(insertIDs))
	for i, id := range insertIDs {
		strIDs[i] = strconv.FormatInt(id, 10)
	}

	// Join the strings with commas
	payload := strings.Join(strIDs, ",")
	req, err := http.NewRequest("POST", url, strings.NewReader(payload))
	if err != nil {
		return fmt.Errorf("postToAutomationWebhook - could not create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("postToAutomationWebhook - request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("postToAutomationWebhook - unexpected status code: %d, response: %s", resp.StatusCode, respBody)
	}

	return nil
}
