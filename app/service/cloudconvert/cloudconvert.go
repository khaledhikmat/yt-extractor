package cloudconvert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/khaledhikmat/yt-extractor/service/config"
	"github.com/khaledhikmat/yt-extractor/service/lgr"
)

type cloudConvertService struct {
	ConfigSvc config.IService
}

func New(cfgsvc config.IService) IService {
	return &cloudConvertService{
		ConfigSvc: cfgsvc,
	}
}

func (svc *cloudConvertService) ConvertVideoToAudio(channelID, videoID string) (string, error) {
	lgr.Logger.Debug("ConvertFromURL",
		slog.String("channelId", channelID),
		slog.String("videoId", videoID),
	)

	// This assumes that the video is stored in S3

	url := "https://api.cloudconvert.com/v2/jobs"
	jobData := map[string]interface{}{
		"tasks": map[string]interface{}{
			"import-my-file": map[string]interface{}{
				"operation":         "import/s3",
				"access_key_id":     svc.ConfigSvc.GetAWSAccessKeyID(),
				"secret_access_key": svc.ConfigSvc.GetAWSSecretAccessKey(),
				"region":            svc.ConfigSvc.GetAWSRegion(),
				"bucket":            svc.ConfigSvc.GetStorageBucket(),
				"key":               fmt.Sprintf("%s/%s.mp4", channelID, videoID),
			},
			"convert-my-file": map[string]interface{}{
				"operation":     "convert",
				"input":         "import-my-file",
				"input_format":  "mp4",
				"output_format": "mp3",
			},
			"export-my-file": map[string]interface{}{
				"operation":         "export/s3",
				"input":             "convert-my-file",
				"access_key_id":     svc.ConfigSvc.GetAWSAccessKeyID(),
				"secret_access_key": svc.ConfigSvc.GetAWSSecretAccessKey(),
				"region":            svc.ConfigSvc.GetAWSRegion(),
				"bucket":            svc.ConfigSvc.GetStorageBucket(),
				"key":               fmt.Sprintf("%s/%s.mp3", channelID, videoID),
			},
		},
	}

	jobBody, _ := json.Marshal(jobData)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jobBody))
	req.Header.Set("Authorization", "Bearer "+svc.ConfigSvc.GetCloudConvertKey())
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("failed to create job, status: %s", resp.Status)
	}

	var jobResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&jobResp)
	if err != nil {
		return "", fmt.Errorf("failed to decode job response: %v", err)
	}

	jobID := jobResp["data"].(map[string]interface{})["id"].(string)
	if jobID == "" {
		return "", fmt.Errorf("job id is empty")
	}

	// Poll for job status
	return svc.checkJobStatus(jobID, channelID, videoID)
}

func (svc *cloudConvertService) checkJobStatus(jobID, channelID, videoID string) (string, error) {
	lgr.Logger.Debug("checkJobStatus",
		slog.String("jobId", jobID),
	)

	client := &http.Client{}
	jobStatusURL := fmt.Sprintf("https://api.cloudconvert.com/v2/jobs/%s", jobID)

	idx := 1
	for {
		lgr.Logger.Debug("checkJobStatus",
			slog.String("jobId", jobID),
			slog.Int("attempt", idx),
		)

		req, _ := http.NewRequest("GET", jobStatusURL, nil)
		req.Header.Set("Authorization", "Bearer "+svc.ConfigSvc.GetCloudConvertKey())
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("failed to show job, status: %s", resp.Status)
		}

		var jobStatusResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&jobStatusResponse)
		if err != nil {
			return "", fmt.Errorf("failed to decode job status response: %v", err)
		}

		status := jobStatusResponse["data"].(map[string]interface{})["status"].(string)
		if status == "finished" {
			// URL must be generated from the bucket name and the folder
			url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", svc.ConfigSvc.GetStorageBucket(), svc.ConfigSvc.GetStorageRegion(), fmt.Sprintf("%s/%s.mp3", channelID, videoID))
			return url, nil
		} else if status == "failed" {
			return "", fmt.Errorf("cloudconvert job failed, status: %s", status)
		}

		// Wait for 5 seconds before retrying
		time.Sleep(5 * time.Second)
		idx++

		// Timeout after 180 attempts ~= 15 minutes
		if idx > 120 {
			lgr.Logger.Debug("checkJobStatus",
				slog.String("timeout", "15 mins"),
			)
			return "", fmt.Errorf("cloudconvert job timedout after 15 minutes")
		}
	}
}
