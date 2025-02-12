package transcription

import (
	"testing"
)

// WARNIMNG: Requires that env vars be set:
// LOCAL_TRANSCRIPTION_FOLDER
// OPENAPI_API_KEY

func TestTranscription(t *testing.T) {
	svc := New(nil)
	text, err := svc.TranscribeAudio("../../temp/transcriptions/small-sample.mp3")
	if err != nil {
		t.Error(err)
	}
	t.Log(text)
}
