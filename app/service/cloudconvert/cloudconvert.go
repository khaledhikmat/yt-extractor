package cloudconvert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
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
				"access_key_id":     os.Getenv("AWS_ACCESS_KEY_ID"),
				"secret_access_key": os.Getenv("AWS_SECRET_ACCESS_KEY"),
				"region":            os.Getenv("STORAGE_REGION"),
				"bucket":            os.Getenv("STORAGE_BUCKET"),
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
				"access_key_id":     os.Getenv("AWS_ACCESS_KEY_ID"),
				"secret_access_key": os.Getenv("AWS_SECRET_ACCESS_KEY"),
				"region":            os.Getenv("STORAGE_REGION"),
				"bucket":            os.Getenv("STORAGE_BUCKET"),
				"key":               fmt.Sprintf("%s/%s.mp3", channelID, videoID),
			},
		},
	}

	jobBody, _ := json.Marshal(jobData)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jobBody))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("CLOUDCONVERT_API_KEY"))
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
	return checkJobStatus(jobID, channelID, videoID)
}

func checkJobStatus(jobID, channelID, videoID string) (string, error) {
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
		req.Header.Set("Authorization", "Bearer "+os.Getenv("CLOUDCONVERT_API_KEY"))
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
			url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", os.Getenv("STORAGE_BUCKET"), os.Getenv("STORAGE_REGION"), fmt.Sprintf("%s/%s.mp3", channelID, videoID))
			return url, nil
		} else if status == "failed" {
			return "", fmt.Errorf("job failed, status: %s", status)
		}

		// Wait for 5 seconds before retrying
		time.Sleep(5 * time.Second)
		idx++
	}
}
