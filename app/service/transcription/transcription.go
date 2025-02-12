package transcription

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/khaledhikmat/yt-extractor/service/config"
)

type transcriptionResponse struct {
	Text string `json:"text"`
}

type transcriptionService struct {
	ConfigSvc config.IService
}

func New(cfgsvc config.IService) IService {
	return &transcriptionService{
		ConfigSvc: cfgsvc,
	}
}

func (svc *transcriptionService) TranscribeAudio(audioFilePath string) (string, error) {
	// Open the audio file
	file, err := os.Open(audioFilePath)
	if err != nil {
		return "", fmt.Errorf("could not open audio file: %w", err)
	}
	defer file.Close()

	// Create a buffer to store the multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add the audio file to the form data
	part, err := writer.CreateFormFile("file", audioFilePath)
	if err != nil {
		return "", fmt.Errorf("could not create form file: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return "", fmt.Errorf("could not copy file data: %w", err)
	}

	// Add model field to the form data (we use "whisper-1" here)
	_ = writer.WriteField("model", "whisper-1")

	// Close the multipart writer
	writer.Close()

	// Send the request to OpenAI's Whisper endpoint
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/audio/transcriptions", body)
	if err != nil {
		return "", fmt.Errorf("could not create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+os.Getenv("OPENAI_API_KEY"))
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, respBody)
	}

	// Parse the response
	var transcriptionResponse transcriptionResponse
	if err := json.NewDecoder(resp.Body).Decode(&transcriptionResponse); err != nil {
		return "", fmt.Errorf("could not decode response: %w", err)
	}

	// Close the file before deleting it
	file.Close()

	// Delete the local file
	err = os.Remove(audioFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to delete local audio file: %w", err)
	}

	return transcriptionResponse.Text, nil
}
