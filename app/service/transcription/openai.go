package transcription

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/khaledhikmat/yt-extractor/service/config"
	"github.com/khaledhikmat/yt-extractor/service/lgr"
)

type transcriptionResponse struct {
	Text string `json:"text"`
}

type openaiService struct {
	ConfigSvc config.IService
}

func newOpenai(cfgsvc config.IService) IService {
	return &openaiService{
		ConfigSvc: cfgsvc,
	}
}

func (svc *openaiService) TranscribeAudio(audioFilePath string) (string, error) {
	lgr.Logger.Debug("TranscribeAudio",
		slog.String("audioFilePath", audioFilePath),
	)

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

	lgr.Logger.Debug("TranscribeAudio",
		slog.String("event", "uploadedFile"),
	)

	// Send the request to OpenAI's Whisper endpoint
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/audio/transcriptions", body)
	if err != nil {
		return "", fmt.Errorf("could not create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+svc.ConfigSvc.GetOpenAIKey())
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

	lgr.Logger.Debug("TranscribeAudio",
		slog.String("event", "receivedResponse"),
	)

	// Parse the response
	var transcriptionResponse transcriptionResponse
	if err := json.NewDecoder(resp.Body).Decode(&transcriptionResponse); err != nil {
		return "", fmt.Errorf("could not decode response: %w", err)
	}

	// Close the file before deleting it
	file.Close()

	lgr.Logger.Debug("TranscribeAudio",
		slog.String("event", "deletingFile"),
	)

	// Delete the local file
	err = os.Remove(audioFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to delete local audio file: %w", err)
	}

	return transcriptionResponse.Text, nil
}
