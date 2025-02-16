package jobautomation

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/khaledhikmat/yt-extractor/job"
	"github.com/khaledhikmat/yt-extractor/service/audio"
	"github.com/khaledhikmat/yt-extractor/service/cloudconvert"
	"github.com/khaledhikmat/yt-extractor/service/config"
	"github.com/khaledhikmat/yt-extractor/service/data"
	"github.com/khaledhikmat/yt-extractor/service/lgr"
	"github.com/khaledhikmat/yt-extractor/service/storage"
	"github.com/khaledhikmat/yt-extractor/service/transcription"
	"github.com/khaledhikmat/yt-extractor/service/youtube"

	jobaudio "github.com/khaledhikmat/yt-extractor/job/audio"
	jobextraction "github.com/khaledhikmat/yt-extractor/job/extraction"
	jobtranscription "github.com/khaledhikmat/yt-extractor/job/transcription"
)

// This job processor map is for internal use and contains only the job processors that might be enqueued
// The server api has one similar to this
var jobProcs = map[data.JobType]job.Processor{
	data.JobTypeExtraction:         jobextraction.Processor,
	data.JobTypeExtractionError:    jobextraction.Processor,
	data.JobTypeAudio:              jobaudio.Processor,
	data.JobTypeAudioError:         jobaudio.Processor,
	data.JobTypeTranscription:      jobtranscription.Processor,
	data.JobTypeTranscriptionError: jobtranscription.Processor,
}

func Processor(ctx context.Context,
	channelID string,
	jobID int64,
	pageSize int,
	errorStream chan error,
	cfgsvc config.IService,
	datasvc data.IService,
	ytsvc youtube.IService,
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

	// Construct a series of jobs to process
	// The re-attempts (or error processors) are included before the actual processors
	// This is to ensure that the re-attempts do not re-do the ones that just failed
	jobs := []data.Job{
		{
			ChannelID: channelID,
			Type:      data.JobTypeExtractionError,
			State:     data.JobStateQueued,
		},
		{
			ChannelID: channelID,
			Type:      data.JobTypeExtraction,
			State:     data.JobStateQueued,
		},
		{
			ChannelID: channelID,
			Type:      data.JobTypeAudioError,
			State:     data.JobStateQueued,
		},
		{
			ChannelID: channelID,
			Type:      data.JobTypeAudio,
			State:     data.JobStateQueued,
		},
		{
			ChannelID: channelID,
			Type:      data.JobTypeTranscriptionError,
			State:     data.JobStateQueued,
		},
		{
			ChannelID: channelID,
			Type:      data.JobTypeTranscription,
			State:     data.JobStateQueued,
		},
	}

	// Run each job processor synchronously
	for _, job := range jobs {
		select {
		case <-ctx.Done():
			finalState = data.JobStateCancelled
			return
		default:
		}

		proc, ok := jobProcs[job.Type]
		if !ok {
			errorStream <- fmt.Errorf("job type %s does not have a processor", job.Type)
			errors++
			return
		}

		job.StartedAt = time.Now()
		jobID, err := datasvc.NewJob(job)
		if err != nil {
			errorStream <- fmt.Errorf("new job produced %s", err.Error())
			errors++
			return
		}

		proc(ctx, job.ChannelID, jobID, pageSize, errorStream, cfgsvc, datasvc, ytsvc, audiosvc, storagesvc, cloudconvertsvc, transcriptionsvc)
	}

	lgr.Logger.Debug("jobtranscription.Processor",
		slog.String("event", "done"),
	)
}
