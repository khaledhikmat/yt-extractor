package cloudconvert

import (
	"testing"

	"github.com/khaledhikmat/yt-extractor/service/config"
)

// WARNIMNG: Requires that env vars be set:
// AWS_ACCESS_KEY_ID
// AWS_SECRET_ACCESS_KEY
// STORAGE_REGION
// STORAGE_BUCKET
// CLOUDCONVERT_API_KEY

func TestConversion(t *testing.T) {
	configSvc := config.New()
	svc := New(configSvc)
	s3URL, err := svc.ConvertVideoToAudio("UCP-PfkMcOKriSxFMH7pTxfA", "02o2vrpOogc")
	if err != nil {
		t.Error(err)
	}
	t.Log(s3URL)
}
