package transcription

import (
	"fmt"

	"github.com/khaledhikmat/yt-extractor/service/config"
)

var providers map[string]IService

type transcriptionService struct {
	ConfigSvc config.IService
}

func New(cfgsvc config.IService) IService {
	providers = map[string]IService{
		"openai": newOpenai(cfgsvc),
		"gemini": newGemini(cfgsvc),
	}
	return &transcriptionService{
		ConfigSvc: cfgsvc,
	}
}

func (svc *transcriptionService) TranscribeAudio(audioFilePath string) (string, error) {
	r, ok := providers[svc.ConfigSvc.GetTranscriptionProvider()]
	if !ok {
		return "", fmt.Errorf("transcription provider %s not found", svc.ConfigSvc.GetTranscriptionProvider())
	}

	return r.TranscribeAudio(audioFilePath)
}
