package jobattributes

import (
	"context"
	"strconv"
	"time"

	"github.com/khaledhikmat/yt-extractor/service/audio"
	"github.com/khaledhikmat/yt-extractor/service/cloudconvert"
	"github.com/khaledhikmat/yt-extractor/service/config"
	"github.com/khaledhikmat/yt-extractor/service/data"
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
	_ config.IService,
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
		_, err := datasvc.NewVideo(video)
		if err != nil {
			errorStream <- err
			errors++
			continue
		}
	}
}
