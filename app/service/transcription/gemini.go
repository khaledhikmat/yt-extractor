package transcription

import (
	"fmt"

	"github.com/khaledhikmat/yt-extractor/service/config"
)

type geminiService struct {
	ConfigSvc config.IService
}

func newGemini(cfgsvc config.IService) IService {
	return &geminiService{
		ConfigSvc: cfgsvc,
	}
}

func (svc *geminiService) TranscribeAudio(_ string) (string, error) {
	return "", fmt.Errorf("Gemini transcription service not implemented")
}
