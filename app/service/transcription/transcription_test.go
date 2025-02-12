package transcription

import (
	"testing"

	"github.com/khaledhikmat/yt-extractor/service/config"
)

// WARNIMNG: Requires that env vars be set:
// TRANSCRIPTION_PROVIDER
// LOCAL_TRANSCRIPTION_FOLDER
// OPENAPI_API_KEY

func TestTranscription(t *testing.T) {
	configSvc := config.New()
	svc := New(configSvc)
	text, err := svc.TranscribeAudio("../../temp/transcriptions/small-sample.mp3")
	if err != nil {
		t.Error(err)
	}
	t.Log(text)
}
