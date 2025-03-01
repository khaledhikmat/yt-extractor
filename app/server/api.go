package server

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/khaledhikmat/yt-extractor/service"
	"github.com/khaledhikmat/yt-extractor/service/audio"
	"github.com/khaledhikmat/yt-extractor/service/cloudconvert"
	"github.com/khaledhikmat/yt-extractor/service/config"
	"github.com/khaledhikmat/yt-extractor/service/data"
	"github.com/khaledhikmat/yt-extractor/service/storage"
	"github.com/khaledhikmat/yt-extractor/service/transcription"
	"github.com/khaledhikmat/yt-extractor/service/youtube"

	"github.com/khaledhikmat/yt-extractor/job"
	jobattributes "github.com/khaledhikmat/yt-extractor/job/attributes"
	jobaudio "github.com/khaledhikmat/yt-extractor/job/audio"
	jobautomation "github.com/khaledhikmat/yt-extractor/job/automation"
	jobextraction "github.com/khaledhikmat/yt-extractor/job/extraction"
	jobtranscription "github.com/khaledhikmat/yt-extractor/job/transcription"
)

const (
	version = "1.0.0"
)

var jobProcs = map[data.JobType]job.Processor{
	data.JobTypeAttributes:         jobattributes.Processor,
	data.JobTypeExtraction:         jobextraction.Processor,
	data.JobTypeExtractionError:    jobextraction.Processor,
	data.JobTypeAudio:              jobaudio.Processor,
	data.JobTypeAudioError:         jobaudio.Processor,
	data.JobTypeTranscription:      jobtranscription.Processor,
	data.JobTypeTranscriptionError: jobtranscription.Processor,
	data.JobTypeAutomation:         jobautomation.Processor,
}

func apiRoutes(ctx context.Context,
	r *gin.Engine,
	errorStream chan error,
	cfgsvc config.IService,
	datasvc data.IService,
	ytsvc youtube.IService,
	audiosvc audio.IService,
	storagesvc storage.IService,
	cloudconvertsvc cloudconvert.IService,
	transcriptionsvc transcription.IService) {

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": fmt.Sprintf("version: %s - env: %s", version, cfgsvc.GetRuntimeEnvironment()),
		})
	})

	r.POST("/admins/reset", func(c *gin.Context) {
		_ = isPermitted(c, datasvc)
		if 1 == 1 {
			c.JSON(403, gin.H{
				"message": "Invalid or missing API key or not alllowed",
			})
			return
		}

		err := datasvc.ResetFactory()
		if err != nil {
			c.JSON(400, gin.H{
				"message": fmt.Sprintf("reset factory produced %s", err.Error()),
			})
			return
		}

		c.JSON(200, gin.H{
			"data": nil,
		})
	})

	r.GET("/videos", func(c *gin.Context) {
		isPermitted := isPermitted(c, datasvc)
		if !isPermitted {
			c.JSON(403, gin.H{
				"message": "Invalid or missing API key",
			})
			return
		}

		channelID := c.Query("c")
		if channelID == "" {
			c.JSON(400, gin.H{
				"message": "channel ID is required",
			})
			return
		}

		page, e := strconv.Atoi(c.Query("p"))
		if e != nil {
			page = 1
		}

		pageSize, e := strconv.Atoi(c.Query("s"))
		if e != nil {
			pageSize = 50
		}

		order := c.Query("o")
		if order == "" {
			order = "published_at"
		}

		dir := c.Query("d")
		if dir == "" {
			dir = "desc"
		}

		//jobType := c.Query("t")
		videos, err := datasvc.RetrieveVideos(channelID, page, pageSize, order, dir)
		if err != nil {
			c.JSON(400, gin.H{
				"message": fmt.Sprintf("retrieve videos produced %s", err.Error()),
			})
			return
		}

		c.JSON(200, gin.H{
			"data": videos,
		})
	})

	r.GET("/videos/unextracted", func(c *gin.Context) {
		isPermitted := isPermitted(c, datasvc)
		if !isPermitted {
			c.JSON(403, gin.H{
				"message": "Invalid or missing API key",
			})
			return
		}

		channelID := c.Query("c")
		if channelID == "" {
			c.JSON(400, gin.H{
				"message": "channel ID is required",
			})
			return
		}

		pageSize, e := strconv.Atoi(c.Query("s"))
		if e != nil {
			pageSize = 50
		}

		videos, err := datasvc.RetrieveUnextractedVideos(channelID, pageSize)
		if err != nil {
			c.JSON(400, gin.H{
				"message": fmt.Sprintf("retrieve unextracted videos produced %s", err.Error()),
			})
			return
		}

		c.JSON(200, gin.H{
			"data": videos,
		})
	})

	r.GET("/videos/extracterrored", func(c *gin.Context) {
		isPermitted := isPermitted(c, datasvc)
		if !isPermitted {
			c.JSON(403, gin.H{
				"message": "Invalid or missing API key",
			})
			return
		}

		channelID := c.Query("c")
		if channelID == "" {
			c.JSON(400, gin.H{
				"message": "channel ID is required",
			})
			return
		}

		pageSize, e := strconv.Atoi(c.Query("s"))
		if e != nil {
			pageSize = 50
		}

		videos, err := datasvc.RetrieveExtractErroredVideos(channelID, pageSize)
		if err != nil {
			c.JSON(400, gin.H{
				"message": fmt.Sprintf("retrieve extract errored videos produced %s", err.Error()),
			})
			return
		}

		c.JSON(200, gin.H{
			"data": videos,
		})
	})

	r.GET("/videos/unexternalized", func(c *gin.Context) {
		isPermitted := isPermitted(c, datasvc)
		if !isPermitted {
			c.JSON(403, gin.H{
				"message": "Invalid or missing API key",
			})
			return
		}

		channelID := c.Query("c")
		if channelID == "" {
			c.JSON(400, gin.H{
				"message": "channel ID is required",
			})
			return
		}

		pageSize, e := strconv.Atoi(c.Query("s"))
		if e != nil {
			pageSize = 50
		}

		videos, err := datasvc.RetrieveUnexternalizedVideos(channelID, pageSize)
		if err != nil {
			c.JSON(400, gin.H{
				"message": fmt.Sprintf("retrieve unexternalized videos produced %s", err.Error()),
			})
			return
		}

		c.JSON(200, gin.H{
			"data": videos,
		})
	})

	r.GET("/videos/unaudioed", func(c *gin.Context) {
		isPermitted := isPermitted(c, datasvc)
		if !isPermitted {
			c.JSON(403, gin.H{
				"message": "Invalid or missing API key",
			})
			return
		}

		channelID := c.Query("c")
		if channelID == "" {
			c.JSON(400, gin.H{
				"message": "channel ID is required",
			})
			return
		}

		pageSize, e := strconv.Atoi(c.Query("s"))
		if e != nil {
			pageSize = 50
		}

		videos, err := datasvc.RetrieveUnaudioedVideos(channelID, pageSize)
		if err != nil {
			c.JSON(400, gin.H{
				"message": fmt.Sprintf("retrieve unaudioed videos produced %s", err.Error()),
			})
			return
		}

		c.JSON(200, gin.H{
			"data": videos,
		})
	})

	r.GET("/videos/audioerrored", func(c *gin.Context) {
		isPermitted := isPermitted(c, datasvc)
		if !isPermitted {
			c.JSON(403, gin.H{
				"message": "Invalid or missing API key",
			})
			return
		}

		channelID := c.Query("c")
		if channelID == "" {
			c.JSON(400, gin.H{
				"message": "channel ID is required",
			})
			return
		}

		pageSize, e := strconv.Atoi(c.Query("s"))
		if e != nil {
			pageSize = 50
		}

		videos, err := datasvc.RetrieveAudioErroredVideos(channelID, pageSize)
		if err != nil {
			c.JSON(400, gin.H{
				"message": fmt.Sprintf("retrieve audio errored videos produced %s", err.Error()),
			})
			return
		}

		c.JSON(200, gin.H{
			"data": videos,
		})
	})

	r.GET("/videos/untranscribed", func(c *gin.Context) {
		isPermitted := isPermitted(c, datasvc)
		if !isPermitted {
			c.JSON(403, gin.H{
				"message": "Invalid or missing API key",
			})
			return
		}

		channelID := c.Query("c")
		if channelID == "" {
			c.JSON(400, gin.H{
				"message": "channel ID is required",
			})
			return
		}

		pageSize, e := strconv.Atoi(c.Query("s"))
		if e != nil {
			pageSize = 50
		}

		videos, err := datasvc.RetrieveUntranscribedVideos(channelID, pageSize)
		if err != nil {
			c.JSON(400, gin.H{
				"message": fmt.Sprintf("retrieve untranscribed videos produced %s", err.Error()),
			})
			return
		}

		c.JSON(200, gin.H{
			"data": videos,
		})
	})

	r.GET("/videos/transcribeerrored", func(c *gin.Context) {
		isPermitted := isPermitted(c, datasvc)
		if !isPermitted {
			c.JSON(403, gin.H{
				"message": "Invalid or missing API key",
			})
			return
		}

		channelID := c.Query("c")
		if channelID == "" {
			c.JSON(400, gin.H{
				"message": "channel ID is required",
			})
			return
		}

		pageSize, e := strconv.Atoi(c.Query("s"))
		if e != nil {
			pageSize = 50
		}

		videos, err := datasvc.RetrieveTranscribeErroredVideos(channelID, pageSize)
		if err != nil {
			c.JSON(400, gin.H{
				"message": fmt.Sprintf("retrieve transcribe errored videos produced %s", err.Error()),
			})
			return
		}

		c.JSON(200, gin.H{
			"data": videos,
		})
	})

	r.GET("/videos/updated", func(c *gin.Context) {
		isPermitted := isPermitted(c, datasvc)
		if !isPermitted {
			c.JSON(403, gin.H{
				"message": "Invalid or missing API key",
			})
			return
		}

		channelID := c.Query("c")
		if channelID == "" {
			c.JSON(400, gin.H{
				"message": "channel ID is required",
			})
			return
		}

		pageSize, e := strconv.Atoi(c.Query("s"))
		if e != nil {
			pageSize = 50
		}

		videos, err := datasvc.RetrieveUpdatedVideos(channelID, pageSize)
		if err != nil {
			c.JSON(400, gin.H{
				"message": fmt.Sprintf("retrieve updated videos produced %s", err.Error()),
			})
			return
		}

		c.JSON(200, gin.H{
			"data": videos,
		})
	})

	r.GET("/jobs", func(c *gin.Context) {
		isPermitted := isPermitted(c, datasvc)
		if !isPermitted {
			c.JSON(403, gin.H{
				"message": "Invalid or missing API key",
			})
			return
		}

		jobID := c.Query("i")
		if jobID == "" {
			c.JSON(400, gin.H{
				"message": "job ID is required",
			})
			return
		}

		id, e := strconv.Atoi(jobID)
		if e != nil {
			c.JSON(400, gin.H{
				"message": "job ID could not be parsed",
			})
			return
		}

		job, err := datasvc.RetrieveJobByID(int64(id))
		if err != nil {
			c.JSON(400, gin.H{
				"message": fmt.Sprintf("retrieve job produced %s", err.Error()),
			})
			return
		}

		c.JSON(200, gin.H{
			"data": job,
		})
	})

	r.POST("/jobs", func(c *gin.Context) {
		isPermitted := isPermitted(c, datasvc)
		if !isPermitted {
			c.JSON(403, gin.H{
				"message": "Invalid or missing API key",
			})
			return
		}

		var job data.Job
		if err := c.ShouldBindJSON(&job); err != nil {
			c.JSON(400, gin.H{
				"message": fmt.Sprintf("invalid job: %s", err.Error()),
			})
			return
		}

		if job.ChannelID == "" {
			c.JSON(400, gin.H{
				"message": "channel ID is required",
			})
			return
		}

		if job.Type == "" {
			c.JSON(400, gin.H{
				"message": "job type is required",
			})
			return
		}

		pageSize, e := strconv.Atoi(c.Query("s"))
		if e != nil {
			pageSize = 50
		}

		id, err := ProcessJob(ctx, job, pageSize, true, errorStream, cfgsvc, datasvc, ytsvc, audiosvc, storagesvc, cloudconvertsvc, transcriptionsvc)
		if err != nil {
			c.JSON(400, gin.H{
				"message": fmt.Sprintf("process job produced %s", err.Error()),
			})
			return
		}

		c.JSON(200, gin.H{
			"data": id, // TODO: Return a status URL
		})
	})

	r.PUT("/videos", func(c *gin.Context) {
		isPermitted := isPermitted(c, datasvc)
		if !isPermitted {
			c.JSON(403, gin.H{
				"message": "Invalid or missing API key",
			})
			return
		}

		id, err := strconv.Atoi(c.Query("i"))
		if err != nil {
			c.JSON(400, gin.H{
				"message": fmt.Sprintf("invalid id: %s", err.Error()),
			})
			return
		}

		// The only thing we allow clients to update is to set the externalization
		video, err := datasvc.RetrieveVideoByID(int64(id))
		if err != nil {
			c.JSON(400, gin.H{
				"message": fmt.Sprintf("Retrieving video id %d caused error: %s", id, err.Error()),
			})
			return
		}

		err = datasvc.UpdateVideo(&video, data.JobTypeExternalization)
		if err != nil {
			c.JSON(400, gin.H{
				"message": fmt.Sprintf("Updating video %d caused error: %s", id, err.Error()),
			})
			return
		}

		c.JSON(200, gin.H{
			"data": nil,
		})
	})

	r.GET("/audio", func(c *gin.Context) {
		isPermitted := isPermitted(c, datasvc)
		if !isPermitted {
			c.JSON(403, gin.H{
				"message": "Invalid or missing API key",
			})
			return
		}

		channelID := c.Query("c")
		if channelID == "" {
			c.JSON(400, gin.H{
				"message": "channel ID is required",
			})
			return
		}

		videoID := c.Query("v")
		if videoID == "" {
			c.JSON(400, gin.H{
				"message": "video ID is required",
			})
			return
		}

		audioURL := c.Query("u")
		if audioURL == "" {
			c.JSON(400, gin.H{
				"message": "audio URL is required",
			})
			return
		}

		// Split the audio URL into multiple files and upload them to the storage
		files, err := audiosvc.SplitAudio(audioURL)
		if err != nil {
			c.JSON(400, gin.H{
				"message": fmt.Sprintf("spliting audio %s produced %s", audioURL, err.Error()),
			})
			return
		}

		// Store the local reference video to an external storage
		urls := []string{}
		idx := 1
		for _, file := range files {
			url, err := storagesvc.NewFile(ctx, channelID, file, fmt.Sprintf("%s_%d.mp3", videoID, idx))
			if err != nil {
				c.JSON(400, gin.H{
					"message": fmt.Sprintf("new error produced %s", err.Error()),
				})
				return
			}
			urls = append(urls, url)
			idx++
		}

		c.JSON(200, gin.H{
			"data": urls,
		})
	})

	r.POST("/errors", func(c *gin.Context) {
		isPermitted := isPermitted(c, datasvc)
		if !isPermitted {
			c.JSON(403, gin.H{
				"message": "Invalid or missing API key",
			})
			return
		}

		var thisError data.Error
		if err := c.ShouldBindJSON(&thisError); err != nil {
			c.JSON(400, gin.H{
				"message": fmt.Sprintf("invalid error: %s", err.Error()),
			})
			return
		}

		// Force an initial state
		err := datasvc.NewError(thisError.Source, thisError.Body)
		if err != nil {
			c.JSON(400, gin.H{
				"message": fmt.Sprintf("new error produced %s", err.Error()),
			})
			return
		}

		c.JSON(200, gin.H{
			"data": nil,
		})
	})

	r.POST("/cloudconvert/webhook", func(c *gin.Context) {
		// isPermitted := isPermitted(c, datasvc)
		// if !isPermitted {
		// 	c.JSON(403, gin.H{
		// 		"message": "Invalid or missing API key",
		// 	})
		// 	return
		// }

		var webhookResponse map[string]interface{}
		if err := c.ShouldBindJSON(&webhookResponse); err != nil {
			errorStream <- fmt.Errorf("cloudconvert webhook - invalid response payload: %s", err.Error())
			c.JSON(400, gin.H{
				"message": fmt.Sprintf("invalid response payload: %s", err.Error()),
			})
			return
		}

		fmt.Println("Webhook Payload:", webhookResponse)

		status := webhookResponse["event"]
		if status == "" {
			// Do not return error to the webhook
			errorStream <- fmt.Errorf("cloudconvert webhook - event status is missing")
			c.JSON(200, gin.H{
				"data": nil,
			})
			return
		}

		if webhookResponse["job"] == nil {
			// Do not return error to the webhook
			errorStream <- fmt.Errorf("cloudconvert webhook - job is missing")
			c.JSON(200, gin.H{
				"data": nil,
			})
			return
		}

		tag := webhookResponse["job"].(map[string]interface{})["tag"].(string)
		if tag == "" {
			// Do not return error to the webhook
			errorStream <- fmt.Errorf("cloudconvert webhook - tag is missing")
			c.JSON(200, gin.H{
				"data": nil,
			})
			return
		}

		// We use the tag as a correlation ID
		tagParts := strings.Split(tag, "|")
		if len(tagParts) != 2 {
			// Do not return error to the webhook
			errorStream <- fmt.Errorf("cloudconvert webhook - tag %s is invalid", tag)
			c.JSON(200, gin.H{
				"data": nil,
			})
			return
		}

		fmt.Printf("Cloudconvert webhook - status: %s, tag: %s\n", status, tag)

		channelID := tagParts[0]
		videoID := tagParts[1]
		url := service.InvalidURL
		if status == "job.finished" {
			// URL must be generated from the bucket name and the folder
			url = constructStorageURL(cfgsvc, channelID, videoID)
			// Guard against multiple webhook posts by making sure that we
			// return an error if the audio URL is not expected
			video, err := updateDb(datasvc, errorStream, channelID, videoID, &url)
			if err == nil {
				// Start an asynchronous transcription processor to process the just-completed audio
				go func() {
					// Ignore the error because it has already been streamed to the error stream
					_ = jobtranscription.Process(ctx, &video, errorStream, data.JobTypeTranscription, cfgsvc, datasvc, audiosvc, storagesvc, transcriptionsvc)
				}()
			}
		} else if status == "job.failed" {
			url = service.InvalidURL
			errorStream <- fmt.Errorf("cloudconvert webhook - job failed for channel %s and video %s", channelID, videoID)
			_, _ = updateDb(datasvc, errorStream, channelID, videoID, &url)
		}

		// If the job status is not finished and is not failed, do not do anything.

		c.JSON(200, gin.H{
			"data": nil,
		})
	})
}

func updateDb(datasvc data.IService, errorStream chan error, channelID, videoID string, URL *string) (data.Video, error) {
	fmt.Printf("Cloudconvert webhook - updateDb - channel ID: %s, video ID: %s - URL: %s\n", channelID, videoID, *URL)
	video, err := datasvc.RetrieveVideoByIDs(channelID, videoID)
	if err != nil {
		errorStream <- err
		return video, err
	}

	if video.AudioURL == nil {
		errorStream <- fmt.Errorf("audio URL should not be nil")
		return video, fmt.Errorf("audio URL should not be nil")
	}

	if *video.AudioURL != service.AcceptedURL {
		errorStream <- fmt.Errorf("audio URL %s is not in waiting state", *video.AudioURL)
		return video, fmt.Errorf("audio URL %s is not in waiting state", *video.AudioURL)
	}

	// Update the video with audio URL
	now := time.Now()
	video.AudioURL = URL
	video.AudioedAt = &now
	err = datasvc.UpdateVideo(&video, data.JobTypeAudio)
	if err != nil {
		errorStream <- err
		return video, err
	}

	fmt.Printf("Cloudconvert webhook - updateDb3\n")
	return video, nil
}

func ProcessJob(ctx context.Context,
	job data.Job,
	pageSize int,
	async bool,
	errorStream chan error,
	cfgsvc config.IService,
	datasvc data.IService,
	ytsvc youtube.IService,
	audiosvc audio.IService,
	storagesvc storage.IService,
	cloudconvertsvc cloudconvert.IService,
	transcriptionsvc transcription.IService) (int64, error) {
	fmt.Printf("Processing job %s for channel %s\n", job.Type, job.ChannelID)
	// Validate there is a processor for the job type
	proc, ok := jobProcs[job.Type]
	if !ok {
		return -1, fmt.Errorf("job type %s does not have a processor", job.Type)
	}

	// Check to make sure there is no existing job for the same type and channel
	isPending, err := datasvc.IsPendingJobsByTypeNChannel(job.ChannelID, job.Type)
	if err != nil {
		return -1, fmt.Errorf("is pending jobs by type and channel produced %s", err.Error())
	}
	if isPending {
		return -1, fmt.Errorf("job type %s for channel %s is already pending", job.Type, job.ChannelID)
	}

	// Force an initial state
	job.State = data.JobStateQueued
	job.StartedAt = time.Now()
	id, err := datasvc.NewJob(job)
	if err != nil {
		return -1, fmt.Errorf("new job produced %s", err.Error())
	}

	if async {
		// Start the job processor asynchronously
		go proc(ctx, job.ChannelID, id, pageSize, errorStream, cfgsvc, datasvc, ytsvc, audiosvc, storagesvc, cloudconvertsvc, transcriptionsvc)
	} else {
		// Start the job processor synchronously
		proc(ctx, job.ChannelID, id, pageSize, errorStream, cfgsvc, datasvc, ytsvc, audiosvc, storagesvc, cloudconvertsvc, transcriptionsvc)
	}

	return id, nil
}

func isPermitted(c *gin.Context, datasvc data.IService) bool {
	apiKey := c.GetHeader("api-key")
	if apiKey == "" {
		return false
	}

	isvalid, err := datasvc.IsAPIKeyValid(apiKey)
	if err != nil {
		return false
	}

	if !isvalid {
		return false
	}

	return true
}

func constructStorageURL(cfgsvc config.IService, channelID, videoID string) string {
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", cfgsvc.GetStorageBucket(), cfgsvc.GetStorageRegion(), fmt.Sprintf("%s/%s.mp3", channelID, videoID))

}
